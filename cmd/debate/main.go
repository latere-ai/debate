// Command debate is the adversarial review CLI for Claude Code coding
// sessions. The full design lives in specs/01-overview.md.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"latere.ai/x/debate/internal/agent"
	"latere.ai/x/debate/internal/cli"
	"latere.ai/x/debate/internal/hook"
	"latere.ai/x/debate/internal/input"
	"latere.ai/x/debate/internal/ledger"
	"latere.ai/x/debate/internal/round"
	"latere.ai/x/debate/internal/state"
	"latere.ai/x/debate/internal/summary"
)

// Set via -ldflags by goreleaser / Makefile.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:           "debate",
		Short:         "Adversarial review for Claude Code coding sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}
	root.SetVersionTemplate("debate {{.Version}}\n")

	flags := cli.Bind(root)
	root.RunE = func(_ *cobra.Command, _ []string) error {
		_, err := cli.Effective(root, flags)
		if err != nil {
			return err
		}
		plan, err := cli.Preflight(root.Context(), flags)
		if err != nil {
			if errors.Is(err, cli.ErrRecursionGuard) {
				return nil
			}
			return err
		}
		return Run(root.Context(), flags, plan)
	}

	root.AddCommand(installHookCmd(), uninstallHookCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "debate:", err)
		os.Exit(exitCodeFor(err))
	}
}

// Run is the end-to-end orchestrator entry point. Exposed so e2e tests
// in this package can drive it directly without re-parsing flags.
func Run(ctx context.Context, flags *cli.Flags, plan *cli.Plan) error {
	// Compute initial diff and gate.
	diff, err := input.Compute(ctx, input.DiffSpec{
		From: flags.DiffFrom, To: flags.DiffTo, Cwd: plan.Cwd,
	})
	if err != nil {
		return err
	}
	// Auto-fallback: when the user didn't override the diff range and
	// the working tree is clean (claude already committed its changes),
	// debate the last commit instead of an empty diff. Manual invocation
	// after `git commit` is the common case where this matters.
	if diff.ChangedLines == 0 && flags.DiffFrom == "HEAD" && flags.DiffTo == "." {
		if fb, fbErr := input.Compute(ctx, input.DiffSpec{
			From: "HEAD~1", To: "HEAD", Cwd: plan.Cwd,
		}); fbErr == nil && fb.ChangedLines > 0 {
			fmt.Fprintln(os.Stderr, "[debate] working tree clean; falling back to HEAD~1..HEAD")
			diff = fb
			flags.DiffFrom, flags.DiffTo = "HEAD~1", "HEAD"
		}
	}
	if input.Trivial(diff, flags.ChangedLinesMin) {
		_ = state.AppendLog(plan.StateDirAbs, &state.LogRecord{
			TS: time.Now().UTC(), Kind: "skipped", Reason: "trivial-diff",
			ChangedLines: diff.ChangedLines, Threshold: flags.ChangedLinesMin,
		})
		fmt.Fprintf(os.Stderr, "[debate] skipped: trivial diff (%d changed lines < %d threshold)\n",
			diff.ChangedLines, flags.ChangedLinesMin)
		return nil
	}

	// Resolve task context.
	taskCtx := flags.TaskContext
	if taskCtx == "" && (plan.TranscriptPath != "" || flags.SessionID != "") {
		if home, herr := os.UserHomeDir(); herr == nil {
			if tp, lerr := input.LocateTranscript(home, plan.Cwd, flags.SessionID, plan.TranscriptPath); lerr == nil {
				if t, rerr := input.ReadTranscript(tp); rerr == nil {
					taskCtx = t.FirstUser
				}
			}
		}
	}
	if taskCtx == "" {
		taskCtx = "(task context unavailable)"
	}

	// Open the session.
	sess, err := state.NewSession(plan.StateDirAbs, len(plan.Forks), time.Now())
	if err != nil {
		return err
	}
	if err := state.WriteStart(sess, &state.StartFile{
		SessionID:   sess.ID,
		StartedAt:   sess.StartedAt,
		Proposer:    state.AgentRef{Agent: flags.Main, Model: flags.MainModel},
		Critic:      state.AgentRef{Agent: flags.Side, Model: flags.SideModel},
		TaskContext: taskCtx,
		TaskSource:  taskSource(flags),
		Diff: state.DiffSnap{
			From: flags.DiffFrom, To: flags.DiffTo,
			ChangedLines: diff.ChangedLines, Files: diff.Files,
			PatchPath: "diff.patch",
		},
		Config: state.ConfigSnap{
			MaxTurn: flags.MaxTurn, SideCount: flags.SideCount,
			CostCap: flags.CostCap, ChangedLinesMin: flags.ChangedLinesMin,
			HookMode: flags.HookMode, Format: flags.Format,
			MainModel: flags.MainModel, SideModel: flags.SideModel,
		},
		RootSession: state.RootSession{
			ID: flags.SessionID, TranscriptPath: plan.TranscriptPath, Cwd: plan.Cwd,
		},
		DebateVersion: version,
		GoVersion:     runtime.Version(),
	}); err != nil {
		return err
	}
	if err := state.WriteRunDiff(sess, diff.Patch); err != nil {
		return err
	}

	// Wire and run the engine.
	proposer := &agent.ClaudeProposer{
		Cwd:      plan.Cwd,
		RootID:   flags.SessionID,
		Model:    flags.MainModel,
		Deadline: 5 * time.Minute,
	}
	criticFactory := func(_ int) agent.Critic { return agent.NewCritic(flags.Side) }
	// Progress lines: Stop-hook path swallows stderr so leave it nil
	// there; manual invocation gets per-fork/per-round status on stderr
	// unless --log-mode silent was passed.
	var progress io.Writer
	if !flags.HookMode && flags.LogMode != cli.LogModeSilent {
		progress = os.Stderr
	}
	heartbeat := round.DefaultHeartbeatInterval
	if flags.LogMode == cli.LogModeVerbose {
		heartbeat = 5 * time.Second
	}
	eng := &round.Engine{
		Sess: sess, Cwd: plan.Cwd, ForkCount: len(plan.Forks),
		Proposer:  proposer,
		NewCritic: criticFactory,
		MaxRounds: cli.MaxRoundsFor(flags.MaxTurn), CostCap: flags.CostCap, HookMode: flags.HookMode,
		TaskContext: taskCtx, DiffPatch: diff.Patch,
		Progress:          progress,
		HeartbeatInterval: heartbeat,
	}
	sumRes, err := eng.Run(ctx)
	if err != nil {
		return err
	}

	// Render summary + persist end.json.
	agg, _ := ledger.Aggregate(sess)
	decide := summary.Decide(sumRes)
	if err := summary.Persist(sumRes, agg, decide.ExitCode); err != nil {
		return err
	}
	_ = state.AppendLog(plan.StateDirAbs, &state.LogRecord{
		TS: time.Now().UTC(), Kind: "run", Session: sess.ID,
		Termination: string(sumRes.Termination), Unresolved: sumRes.Unresolved,
		Tokens: sumRes.TokensUsed, WallSeconds: sumRes.WallSeconds,
		Summary: sess.Path("summary.md"),
	})
	// Print the rendered summary to stdout (TTY-styled when
	// interactive, plain markdown when piped). The hook-mode path
	// stays silent: claude swallows stdout there.
	if !flags.HookMode {
		if body, readErr := os.ReadFile(sess.Path("summary.md")); readErr == nil {
			_, _ = summary.PrintRendered(os.Stdout, body, summary.IsTerminal(os.Stdout))
			fmt.Println()
		}
	}
	if decide.StdoutLine != "" {
		fmt.Println(decide.StdoutLine)
	}

	// Translate exit code, applying --hook-mode.
	if flags.HookMode {
		return nil
	}
	if decide.ExitCode != 0 {
		os.Exit(decide.ExitCode)
	}
	return nil
}

func taskSource(f *cli.Flags) string {
	switch {
	case f.TaskContext != "":
		return "flag"
	case f.Transcript != "":
		return "transcript"
	case f.SessionID != "":
		return "session-id-resume"
	default:
		return "unknown"
	}
}

func exitCodeFor(err error) int {
	var pe *cli.PreflightError
	if errors.As(err, &pe) && pe != nil {
		return pe.Code
	}
	return 1
}

func installHookCmd() *cobra.Command {
	var scope, scriptPath string
	cmd := &cobra.Command{
		Use:   "install-hook",
		Short: "Install the Stop hook into ~/.claude/settings.json (or project)",
		RunE: func(_ *cobra.Command, _ []string) error {
			s := hook.ScopeUser
			if scope == "project" {
				s = hook.ScopeProject
			}
			if scriptPath == "" {
				scriptPath = hook.LocateScript()
			}
			return hook.Install(s, scriptPath)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "user", "user | project")
	cmd.Flags().StringVar(&scriptPath, "script-path", "", "explicit path to debate-stop-hook.sh")
	return cmd
}

func uninstallHookCmd() *cobra.Command {
	var scope string
	cmd := &cobra.Command{
		Use:   "uninstall-hook",
		Short: "Remove the Stop hook from ~/.claude/settings.json (or project)",
		RunE: func(_ *cobra.Command, _ []string) error {
			s := hook.ScopeUser
			if scope == "project" {
				s = hook.ScopeProject
			}
			return hook.Uninstall(s)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "user", "user | project")
	return cmd
}
