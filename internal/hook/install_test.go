package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallEmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, "/path/to/debate-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "debate-stop-hook.sh") {
		t.Errorf("missing entry: %s", b)
	}
}

func TestInstallIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, "/p/debate-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	if err := Install(ScopeUser, "/p/debate-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	var settings map[string]any
	_ = json.Unmarshal(b, &settings)
	stop, _ := settings["hooks"].(map[string]any)["Stop"].([]any)
	if len(stop) != 1 {
		t.Errorf("expected 1 Stop entry after two installs, got %d", len(stop))
	}
}

func TestInstallPreservesUnrelatedHooks(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	settings := map[string]any{
		"hooks": map[string]any{
			"Stop": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{"type": "command", "command": "/other/hook.sh"},
					},
				},
			},
		},
	}
	b, _ := json.Marshal(settings)
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Install(ScopeUser, "/p/debate-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if !strings.Contains(string(out), "/other/hook.sh") {
		t.Error("lost the unrelated hook entry")
	}
	if !strings.Contains(string(out), "debate-stop-hook.sh") {
		t.Error("missing the new debate entry")
	}
}

func TestUninstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, "/p/debate-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(ScopeUser); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if strings.Contains(string(b), "debate-stop-hook.sh") {
		t.Errorf("entry still present: %s", b)
	}
}
