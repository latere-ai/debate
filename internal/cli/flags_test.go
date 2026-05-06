package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func newCmd() (*cobra.Command, *Flags) {
	cmd := &cobra.Command{Use: "debate"}
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
	t.Setenv("DEBATE_SIDE_COUNT", "7")
	t.Setenv("DEBATE_HOOK_MODE", "true")
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
	t.Setenv("DEBATE_SIDE_COUNT", "7")
	cmd, f := newCmd()
	if err := cmd.ParseFlags([]string{"--side-count", "2"}); err != nil {
		t.Fatal(err)
	}
	ApplyEnv(cmd, f)
	if f.SideCount != 2 {
		t.Errorf("CLI flag should win: got %d, want 2", f.SideCount)
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

// TestMaxRoundsFor pins the doubling rule cmd/debate/main.go relies
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
