package cli

import (
	"reflect"
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
	if want := []string{"functional-logic", "security", "code-quality", "performance"}; !reflect.DeepEqual(f.Aspect, want) {
		t.Errorf("default aspects: got %v, want %v", f.Aspect, want)
	}
	if f.CostCap != 50000 {
		t.Errorf("default cost-cap: got %d, want 50000", f.CostCap)
	}
}

func TestFlagParsing(t *testing.T) {
	cmd, f := newCmd()
	cmd.SetArgs([]string{"--main-model", "claude-opus", "--side-count", "2", "--aspect", "x,y", "--hook-mode"})
	if err := cmd.ParseFlags([]string{"--main-model", "claude-opus", "--side-count", "2", "--aspect", "x,y", "--hook-mode"}); err != nil {
		t.Fatal(err)
	}
	if f.MainModel != "claude-opus" {
		t.Errorf("main-model: got %q", f.MainModel)
	}
	if f.SideCount != 2 {
		t.Errorf("side-count: got %d", f.SideCount)
	}
	if !reflect.DeepEqual(f.Aspect, []string{"x", "y"}) {
		t.Errorf("aspect: got %v", f.Aspect)
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
