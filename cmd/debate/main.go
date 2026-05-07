// Command debate is the adversarial review CLI for Claude Code coding
// sessions. The full design lives in specs/01-overview.md.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"syscall"
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

// init populates version/commit/date from runtime build info when
// -ldflags didn't set them. This is the `go install` path: the
// toolchain stamps module version + vcs.revision into the binary
// even though -ldflags is empty. Without this, `go install ...@v0.0.1`
// would print "debate dev (none, unknown)".
func init() {
	if version != "dev" {
		return // ldflags wins; goreleaser / Makefile path
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		version = v
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			commit = s.Value
			if len(commit) > 7 {
				commit = commit[:7]
			}
		case "vcs.time":
			date = s.Value
		}
	}
}

func main() {
	os.Exit(realMain(os.Args[1:], os.Stdout, os.Stderr))
}

// realMain is the testable entry point: returns the process exit code
// instead of calling os.Exit, and accepts argv + stdout/stderr as
// parameters so tests can drive it without process spawning.
func realMain(args []string, stdout, stderr io.Writer) int {
	root := &cobra.Command{
		Use:           "debate",
		Short:         "Adversarial review for Claude Code coding sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}
	root.SetVersionTemplate("debate {{.Version}}\n")
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	flags := cli.Bind(root)
	var exitCode int
	root.RunE = func(cmd *cobra.Command, _ []string) error {
		_, err := cli.Effective(root, flags)
		if err != nil {
			return err
		}
		// Bare `debate` with no args and no env-supplied task source:
		// show help instead of failing preflight with "cannot
		// determine task context". A user who types just the binary
		// name expects orientation, not a cryptic error.
		if cli.ShouldShowHelp(len(args)+1, flags) {
			return cmd.Help()
		}
		plan, err := cli.Preflight(root.Context(), flags)
		if err != nil {
			if errors.Is(err, cli.ErrRecursionGuard) {
				return nil
			}
			return err
		}
		code, runErr := Run(root.Context(), flags, plan)
		exitCode = code
		return runErr
	}

	root.AddCommand(installHookCmd(), uninstallHookCmd(), hookCmd())

	// Install the signal-aware context so SIGINT / SIGTERM cancel the
	// root context and propagate down through agent.Exec's process-group
	// teardown. Required by spec 21.
	ctx, cancel := round.InstallHandler(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := root.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(stderr, "debate:", err)
		return exitCodeFor(err)
	}
	return exitCode
}

// Run is the end-to-end orchestrator entry point. Exposed so e2e tests
// in this package can drive it directly without re-parsing flags.
// Returns the intended process exit code alongside any error: callers
// must propagate both so the binary's exit status matches the
// summary's verdict (0 clean, 1 unresolved leaves, etc.).
func Run(ctx context.Context, flags *cli.Flags, plan *cli.Plan) (int, error) {
	// Compute initial diff and gate.
	diff, err := input.Compute(ctx, input.DiffSpec{
		From: flags.DiffFrom, To: flags.DiffTo, Cwd: plan.Cwd,
	})
	if err != nil {
		return 1, err
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
		return 0, nil
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
		return 1, err
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
		return 1, err
	}
	if err := state.WriteRunDiff(sess, diff.Patch); err != nil {
		return 1, err
	}

	// Wire and run the engine. In verbose mode each agent driver
	// emits live tool / thinking events to stderr while a call is in
	// flight, so the operator sees what the agent is doing instead of
	// only "still running".
	verbose := flags.LogMode == cli.LogModeVerbose
	var eventOut io.Writer
	if verbose && !flags.HookMode {
		eventOut = os.Stderr
	}
	proposer := &agent.ClaudeProposer{
		Cwd:      plan.Cwd,
		RootID:   flags.SessionID,
		Model:    flags.MainModel,
		Deadline: 5 * time.Minute,
		Verbose:  verbose,
		EventOut: eventOut,
	}
	criticFactory := func(_ int) agent.Critic {
		switch flags.Side {
		case "codex":
			return &agent.CodexCritic{Verbose: verbose, EventOut: eventOut}
		case "claude":
			return &agent.ClaudeCritic{Verbose: verbose, EventOut: eventOut}
		default:
			return agent.NewCritic(flags.Side)
		}
	}
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
		Styled:            progress != nil && summary.IsTerminal(os.Stderr),
	}
	sumRes, err := eng.Run(ctx)
	if err != nil {
		return 1, err
	}

	// Render summary + persist end.json.
	agg, _ := ledger.Aggregate(sess)
	decide := summary.Decide(sumRes)
	if err := summary.Persist(sumRes, agg, decide.ExitCode); err != nil {
		return 1, err
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
		return 0, nil
	}
	return decide.ExitCode, nil
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
	var scope, command string
	cmd := &cobra.Command{
		Use:   "install-hook",
		Short: "Install the Stop hook into ~/.claude/settings.json (or project)",
		RunE: func(_ *cobra.Command, _ []string) error {
			s := hook.ScopeUser
			if scope == "project" {
				s = hook.ScopeProject
			}
			// Default: `<absolute-path-to-this-binary> hook`. Self-
			// contained: no shell script on PATH, no `jq` dependency,
			// works after a bare `go install`.
			if command == "" {
				exe, err := os.Executable()
				if err != nil {
					return fmt.Errorf("locate self for hook command: %w", err)
				}
				command = exe + " hook"
			}
			return hook.Install(s, command)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "user", "user | project")
	cmd.Flags().StringVar(&command, "command", "",
		"explicit hook command string (default: \"<this binary> hook\")")
	// Back-compat: old name. Same effect.
	cmd.Flags().StringVar(&command, "script-path", command,
		"deprecated alias for --command; pass a script path or any shell command")
	_ = cmd.Flags().MarkDeprecated("script-path", "use --command instead")
	return cmd
}

// hookCmd is the Stop-hook entry point. claude invokes it with the
// hook payload on stdin; we parse session_id / transcript_path / cwd
// out of it and run the orchestrator in --hook-mode. This replaces
// the v0 shell-script trampoline (scripts/debate-stop-hook.sh) so a
// bare `go install` is enough to make the hook work; no separate
// script on PATH, no `jq` dependency.
func hookCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "hook",
		Short:  "Stop-hook entry point. Reads claude's payload from stdin.",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runHook(cmd.Context())
		},
	}
}

func runHook(ctx context.Context) error {
	// Recursion guard. The orchestrator spawns claude / codex
	// subprocesses; those also fire the Stop hook. Without this guard
	// the hook would re-enter the orchestrator on every round.
	if os.Getenv("DEBATE_IN_PROGRESS") != "" {
		return nil
	}

	// Parse claude's hook payload from stdin. Best-effort: an empty
	// payload still falls through to the orchestrator's preflight.
	payload, _ := io.ReadAll(os.Stdin)
	var p struct {
		SessionID      string `json:"session_id"`
		TranscriptPath string `json:"transcript_path"`
		Cwd            string `json:"cwd"`
	}
	_ = json.Unmarshal(payload, &p)

	// Stale ANTHROPIC_API_KEY in env causes 401 in claude --print
	// subprocesses on OAuth-only accounts. Strip it; the orchestrator
	// also strips it from agent subprocess envs via CleanEnv (spec 16).
	_ = os.Unsetenv("ANTHROPIC_API_KEY")

	// claude --resume is cwd-scoped; cd to where the user's session
	// lives so the resume call hits the right project dir.
	if p.Cwd != "" {
		if err := os.Chdir(p.Cwd); err != nil {
			return fmt.Errorf("cd to hook cwd %q: %w", p.Cwd, err)
		}
	}

	// Build flags as if `--hook-mode --session-id <id> --transcript
	// <path> --max-turn 6` were passed, then drive the same Effective
	// → Preflight → Run pipeline as the root command.
	sub := &cobra.Command{Use: "debate"}
	flags := cli.Bind(sub)
	flags.HookMode = true
	flags.MaxTurn = 6
	flags.SessionID = p.SessionID
	flags.Transcript = p.TranscriptPath
	if _, err := cli.Effective(sub, flags); err != nil {
		return err
	}
	plan, err := cli.Preflight(ctx, flags)
	if err != nil {
		if errors.Is(err, cli.ErrRecursionGuard) {
			return nil
		}
		return err
	}
	code, runErr := Run(ctx, flags, plan)
	if runErr == nil && code != 0 {
		// hook-mode forces exit 0 inside Run on the happy path; a
		// non-zero code here means a preflight-style failure that
		// should propagate so claude marks the hook as failed.
		os.Exit(code)
	}
	return runErr
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
