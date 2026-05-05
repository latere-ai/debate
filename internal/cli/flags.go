// Package cli wires flags, config, and pre-flight validation.
package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Flags is the post-parse, post-env, post-config effective configuration
// for one debate run. See specs/04-cli-flags.md for field semantics.
type Flags struct {
	Main            string
	Side            string
	SideCount       int
	MainModel       string
	SideModel       string
	MaxTurn         int
	Aspect          []string
	SessionID       string
	Transcript      string
	DiffFrom        string
	DiffTo          string
	TaskContext     string
	Judge           string
	CostCap         int
	ChangedLinesMin int
	StateDir        string
	Format          string
	HookMode        bool
	Config          string
	Verbose         int
}

// DefaultFlags returns the built-in defaults.
func DefaultFlags() *Flags {
	return &Flags{
		Main:            "claude",
		Side:            "codex",
		SideCount:       4,
		MaxTurn:         6,
		Aspect:          []string{"functional-logic", "security", "code-quality", "performance"},
		DiffFrom:        "HEAD",
		DiffTo:          ".",
		Judge:           "none",
		CostCap:         50000,
		ChangedLinesMin: 10,
		StateDir:        ".debate",
		Format:          "markdown",
	}
}

// Bind registers all flags onto cmd and returns a *Flags whose fields
// are populated when cmd runs. Bind does not validate.
func Bind(cmd *cobra.Command) *Flags {
	f := DefaultFlags()

	cmd.Flags().StringVar(&f.Main, "main", f.Main, "proposer agent: claude or codex (codex is v1)")
	cmd.Flags().StringVar(&f.Side, "side", f.Side, "critic agent: claude or codex")
	cmd.Flags().IntVar(&f.SideCount, "side-count", f.SideCount, "number of critic forks; must equal len(--aspect)")
	cmd.Flags().StringVar(&f.MainModel, "main-model", f.MainModel, "proposer model; required (and must differ from --side-model) when same family")
	cmd.Flags().StringVar(&f.SideModel, "side-model", f.SideModel, "critic model; required (and must differ from --main-model) when same family")
	cmd.Flags().IntVar(&f.MaxTurn, "max-turn", f.MaxTurn, "per-fork cap on P+C exchanges")
	cmd.Flags().StringSliceVar(&f.Aspect, "aspect", f.Aspect, "comma-separated list of aspect names")
	cmd.Flags().StringVar(&f.SessionID, "session-id", f.SessionID, "claude root session id (claude-as-proposer auto-trigger)")
	cmd.Flags().StringVar(&f.Transcript, "transcript", f.Transcript, "path to root claude JSONL transcript")
	cmd.Flags().StringVar(&f.DiffFrom, "diff-from", f.DiffFrom, "git ref for diff base")
	cmd.Flags().StringVar(&f.DiffTo, "diff-to", f.DiffTo, "git ref for diff head; '.' means working tree")
	cmd.Flags().StringVar(&f.TaskContext, "task-context", f.TaskContext, "explicit task description; required iff neither --session-id nor --transcript is set")
	cmd.Flags().StringVar(&f.Judge, "judge", f.Judge, "none | llm | human (only 'none' is supported in v0)")
	cmd.Flags().IntVar(&f.CostCap, "cost-cap", f.CostCap, "total token budget across all forks")
	cmd.Flags().IntVar(&f.ChangedLinesMin, "changed-lines-min", f.ChangedLinesMin, "trivial-diff gate threshold")
	cmd.Flags().StringVar(&f.StateDir, "state-dir", f.StateDir, "where session folders live")
	cmd.Flags().StringVar(&f.Format, "format", f.Format, "summary format: markdown or json")
	cmd.Flags().BoolVar(&f.HookMode, "hook-mode", f.HookMode, "force exit 0; used by the default Stop hook")
	cmd.Flags().StringVar(&f.Config, "config", f.Config, "explicit .debate.toml path; empty = search")
	cmd.Flags().CountVarP(&f.Verbose, "verbose", "v", "verbose: -v, -vv")

	return f
}

// ApplyEnv overrides any flag not explicitly set on cmd from the
// matching DEBATE_* env var. Only invoked for flags whose Changed bit
// is false.
func ApplyEnv(cmd *cobra.Command, f *Flags) {
	for _, b := range envBindings(f) {
		if cmd.Flags().Changed(b.flag) {
			continue
		}
		v := os.Getenv(b.env)
		if v == "" {
			continue
		}
		b.set(v)
	}
}

type envBinding struct {
	flag string
	env  string
	set  func(string)
}

func envBindings(f *Flags) []envBinding {
	return []envBinding{
		{"main", "DEBATE_MAIN", func(v string) { f.Main = v }},
		{"side", "DEBATE_SIDE", func(v string) { f.Side = v }},
		{"side-count", "DEBATE_SIDE_COUNT", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				f.SideCount = n
			}
		}},
		{"main-model", "DEBATE_MAIN_MODEL", func(v string) { f.MainModel = v }},
		{"side-model", "DEBATE_SIDE_MODEL", func(v string) { f.SideModel = v }},
		{"max-turn", "DEBATE_MAX_TURN", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				f.MaxTurn = n
			}
		}},
		{"aspect", "DEBATE_ASPECT", func(v string) {
			parts := strings.Split(v, ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				if p = strings.TrimSpace(p); p != "" {
					out = append(out, p)
				}
			}
			f.Aspect = out
		}},
		{"session-id", "DEBATE_SESSION_ID", func(v string) { f.SessionID = v }},
		{"transcript", "DEBATE_TRANSCRIPT", func(v string) { f.Transcript = v }},
		{"diff-from", "DEBATE_DIFF_FROM", func(v string) { f.DiffFrom = v }},
		{"diff-to", "DEBATE_DIFF_TO", func(v string) { f.DiffTo = v }},
		{"task-context", "DEBATE_TASK_CONTEXT", func(v string) { f.TaskContext = v }},
		{"judge", "DEBATE_JUDGE", func(v string) { f.Judge = v }},
		{"cost-cap", "DEBATE_COST_CAP", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				f.CostCap = n
			}
		}},
		{"changed-lines-min", "DEBATE_CHANGED_LINES_MIN", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				f.ChangedLinesMin = n
			}
		}},
		{"state-dir", "DEBATE_STATE_DIR", func(v string) { f.StateDir = v }},
		{"format", "DEBATE_FORMAT", func(v string) { f.Format = v }},
		{"hook-mode", "DEBATE_HOOK_MODE", func(v string) {
			if v == "1" || strings.EqualFold(v, "true") {
				f.HookMode = true
			}
		}},
		{"config", "DEBATE_CONFIG", func(v string) { f.Config = v }},
		{"verbose", "DEBATE_VERBOSE", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				f.Verbose = n
			}
		}},
	}
}
