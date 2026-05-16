package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func newCmd() (*cobra.Command, *Flags) {
	cmd := &cobra.Command{Use: "agon"}
	return cmd, Bind(cmd)
}

func TestDefaults(t *testing.T) {
	_, f := newCmd()
	if f.Main != "claude" || f.Side != "codex" {
		t.Errorf("default agents: got main=%q side=%q", f.Main, f.Side)
	}
	if f.SideCount != 4 {
		t.Errorf("default side-count: got %d, want 4", f.SideCount)
	}
	if f.CostCap != 50000 {
		t.Errorf("default cost-cap: got %d, want 50000", f.CostCap)
	}
}

func TestFlagParsing(t *testing.T) {
	cmd, f := newCmd()
	if err := cmd.ParseFlags([]string{"--main-model", "claude-opus", "--side-count", "2", "--hook-mode"}); err != nil {
		t.Fatal(err)
	}
	if f.MainModel != "claude-opus" {
		t.Errorf("main-model: got %q", f.MainModel)
	}
	if f.SideCount != 2 {
		t.Errorf("side-count: got %d", f.SideCount)
	}
	if !f.HookMode {
		t.Error("hook-mode not set")
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("AGON_SIDE_COUNT", "7")
	t.Setenv("AGON_HOOK_MODE", "true")
	cmd, f := newCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	ApplyEnv(cmd, f)
	if f.SideCount != 7 {
		t.Errorf("env side-count: got %d", f.SideCount)
	}
	if !f.HookMode {
		t.Error("env hook-mode not applied")
	}
}

func TestEnvDoesNotOverrideExplicitFlag(t *testing.T) {
	t.Setenv("AGON_SIDE_COUNT", "7")
	cmd, f := newCmd()
	if err := cmd.ParseFlags([]string{"--side-count", "2"}); err != nil {
		t.Fatal(err)
	}
	ApplyEnv(cmd, f)
	if f.SideCount != 2 {
		t.Errorf("CLI flag should win: got %d, want 2", f.SideCount)
	}
}

func TestEnvBindings_AllKeysApplied(t *testing.T) {
	t.Setenv("AGON_MAIN", "claude")
	t.Setenv("AGON_SIDE", "codex")
	t.Setenv("AGON_SIDE_COUNT", "9")
	t.Setenv("AGON_MAIN_MODEL", "m1")
	t.Setenv("AGON_SIDE_MODEL", "m2")
	t.Setenv("AGON_MAX_TURN", "12")
	t.Setenv("AGON_SESSION_ID", "sid-1")
	t.Setenv("AGON_TRANSCRIPT", "/t/p")
	t.Setenv("AGON_DIFF_FROM", "HEAD~3")
	t.Setenv("AGON_DIFF_TO", "HEAD")
	t.Setenv("AGON_TASK_CONTEXT", "build a thing")
	t.Setenv("AGON_JUDGE", "none")
	t.Setenv("AGON_COST_CAP", "12345")
	t.Setenv("AGON_CHANGED_LINES_MIN", "20")
	t.Setenv("AGON_STATE_DIR", "/tmp/x")
	t.Setenv("AGON_FORMAT", "markdown")
	t.Setenv("AGON_HOOK_MODE", "1")
	t.Setenv("AGON_CONFIG", "/c.toml")
	t.Setenv("AGON_VERBOSE", "2")

	cmd, f := newCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	ApplyEnv(cmd, f)

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"Main", f.Main, "claude"},
		{"Side", f.Side, "codex"},
		{"SideCount", f.SideCount, 9},
		{"MainModel", f.MainModel, "m1"},
		{"SideModel", f.SideModel, "m2"},
		{"MaxTurn", f.MaxTurn, 12},
		{"SessionID", f.SessionID, "sid-1"},
		{"Transcript", f.Transcript, "/t/p"},
		{"DiffFrom", f.DiffFrom, "HEAD~3"},
		{"DiffTo", f.DiffTo, "HEAD"},
		{"TaskContext", f.TaskContext, "build a thing"},
		{"Judge", f.Judge, "none"},
		{"CostCap", f.CostCap, 12345},
		{"ChangedLinesMin", f.ChangedLinesMin, 20},
		{"StateDir", f.StateDir, "/tmp/x"},
		{"Format", f.Format, "markdown"},
		{"HookMode", f.HookMode, true},
		{"Config", f.Config, "/c.toml"},
		{"Verbose", f.Verbose, 2},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestEnvBindings_HookModeFalsey(t *testing.T) {
	t.Setenv("AGON_HOOK_MODE", "0")
	cmd, f := newCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	ApplyEnv(cmd, f)
	if f.HookMode {
		t.Errorf("HOOK_MODE=0 should not enable HookMode")
	}
}

func TestEnvBindings_HookModeTrue(t *testing.T) {
	t.Setenv("AGON_HOOK_MODE", "TRUE")
	cmd, f := newCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	ApplyEnv(cmd, f)
	if !f.HookMode {
		t.Errorf("HOOK_MODE=TRUE (case-insensitive) should enable HookMode")
	}
}

func TestEnvBindings_BadIntsIgnored(t *testing.T) {
	t.Setenv("AGON_SIDE_COUNT", "not-a-number")
	t.Setenv("AGON_MAX_TURN", "")
	t.Setenv("AGON_COST_CAP", "weird")
	t.Setenv("AGON_CHANGED_LINES_MIN", "x")
	t.Setenv("AGON_VERBOSE", "x")
	cmd, f := newCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	before := *f
	ApplyEnv(cmd, f)
	// All malformed ints leave the prior value untouched.
	if f.SideCount != before.SideCount {
		t.Errorf("bad SideCount silently mutated to %d", f.SideCount)
	}
}

// TestShouldShowHelp covers the bare-invocation UX: `agon` with no
// args and no env-supplied task source must redirect to help instead
// of running preflight and failing on the cryptic "cannot determine
// task context" error.
func TestShouldShowHelp(t *testing.T) {
	cases := []struct {
		name string
		argc int
		f    *Flags
		want bool
	}{
		{"bare no env", 1, &Flags{}, true},
		{"bare with session-id env", 1, &Flags{SessionID: "abc"}, false},
		{"bare with transcript env", 1, &Flags{Transcript: "/p"}, false},
		{"bare with task-context env", 1, &Flags{TaskContext: "x"}, false},
		{"any flag passed", 2, &Flags{}, false},
		{"flag + value", 3, &Flags{}, false},
		// argc==0 (impossible in practice but defensive) still
		// triggers help, since there's nothing to run on.
		{"argc 0 defensive", 0, &Flags{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ShouldShowHelp(c.argc, c.f); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

// TestIsValidLogMode pins the accepted set. Adding or removing a
// value here is a UX change and must not be silent.
func TestIsValidLogMode(t *testing.T) {
	for _, ok := range []string{LogModeSilent, LogModeConcise, LogModeVerbose} {
		if !IsValidLogMode(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "loud", "thinking", "VERBOSE"} {
		if IsValidLogMode(bad) {
			t.Errorf("%q should be invalid", bad)
		}
	}
}

// TestDefaultLogMode locks the default in. Hook-mode and CI users
// should never have to pass --log-mode just to keep current
// behaviour.
func TestDefaultLogMode(t *testing.T) {
	if got := DefaultFlags().LogMode; got != LogModeConcise {
		t.Errorf("default LogMode: got %q, want %q", got, LogModeConcise)
	}
}

// TestMaxRoundsFor pins the doubling rule cmd/agon/main.go relies
// on: --max-turn N pairs translates to 2N internal rounds. A
// regression here would silently halve or double how long the engine
// runs.
func TestMaxRoundsFor(t *testing.T) {
	cases := []struct{ turn, rounds int }{
		{1, 2}, {2, 4}, {3, 6}, {10, 20},
	}
	for _, c := range cases {
		if got := MaxRoundsFor(c.turn); got != c.rounds {
			t.Errorf("MaxRoundsFor(%d): got %d, want %d", c.turn, got, c.rounds)
		}
	}
}
