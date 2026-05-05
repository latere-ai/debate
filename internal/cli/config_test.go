package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestProjectConfigOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, ".debate.toml")
	if err := os.WriteFile(cfg, []byte(`max_turn = 10
side_count = 2
aspects = ["a", "b"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "debate"}
	f := Bind(cmd)
	f.Config = cfg
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Effective(cmd, f); err != nil {
		t.Fatal(err)
	}
	if f.MaxTurn != 10 {
		t.Errorf("max_turn from file: got %d, want 10", f.MaxTurn)
	}
	if f.SideCount != 2 {
		t.Errorf("side_count from file: got %d, want 2", f.SideCount)
	}
	if len(f.Aspect) != 2 || f.Aspect[0] != "a" {
		t.Errorf("aspects from file: got %v", f.Aspect)
	}
}

func TestCLIFlagWinsOverConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, ".debate.toml")
	if err := os.WriteFile(cfg, []byte(`max_turn = 10`), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "debate"}
	f := Bind(cmd)
	f.Config = cfg
	if err := cmd.ParseFlags([]string{"--max-turn", "3"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Effective(cmd, f); err != nil {
		t.Fatal(err)
	}
	if f.MaxTurn != 3 {
		t.Errorf("CLI flag should win: got %d, want 3", f.MaxTurn)
	}
}

func TestUnknownKeyRejected(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, ".debate.toml")
	if err := os.WriteFile(cfg, []byte(`bogus_key = 42`), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "debate"}
	f := Bind(cmd)
	f.Config = cfg
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Effective(cmd, f); err == nil {
		t.Error("expected error for unknown TOML key")
	}
}
