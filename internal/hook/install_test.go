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

func TestInstallProjectScope(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := Install(ScopeProject, "/p/debate-stop-hook.sh"); err != nil {
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

func TestInstallEmptyScriptPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, ""); err == nil {
		t.Error("expected error for empty scriptPath")
	}
}

func TestInstallRequiresScript(t *testing.T) {
	if err := Install(ScopeUser, ""); err == nil {
		t.Fatal("expected error for empty script path")
	}
}

func TestSettingsPathUnknownScope(t *testing.T) {
	if _, err := SettingsPath(Scope(99)); err == nil {
		t.Error("expected error for unknown scope")
	}
}

func TestSettingsPathProject(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	p, err := SettingsPath(ScopeProject)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(p, filepath.Join(".claude", "settings.json")) {
		t.Errorf("got %q, want suffix .claude/settings.json", p)
	}
}

func TestUninstallNoExistingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Uninstall against a settings.json that doesn't exist creates an
	// empty one with no hooks. That's fine; just must not error.
	if err := Uninstall(ScopeUser); err != nil {
		t.Errorf("Uninstall on missing file: %v", err)
	}
}

func TestReadSettingsBadJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(p, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readSettings(p); err == nil {
		t.Error("expected parse error")
	}
}

func TestReadSettingsEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(p, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := readSettings(p)
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Error("expected empty map, not nil")
	}
}

func TestReadSettingsExplicitNullJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(p, []byte("null"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := readSettings(p)
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Error("null JSON should be coerced to empty map")
	}
}

func TestLocateScript_FallsBackToBareName(t *testing.T) {
	// Move into a directory that does not contain scripts/.
	tmp := t.TempDir()
	t.Chdir(tmp)
	got := LocateScript()
	if got != "debate-stop-hook.sh" {
		// Could legitimately resolve to the binary's sibling if go test
		// installed the test binary somewhere with one alongside; just
		// require that what we got is non-empty.
		if got == "" {
			t.Error("LocateScript returned empty")
		}
	}
}

func TestLocateScript_FindsSiblingNextToBinary(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	sibling := filepath.Join(filepath.Dir(exe), "debate-stop-hook.sh")
	if err := os.WriteFile(sibling, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Skipf("cannot write sibling: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(sibling) })

	got := LocateScript()
	if got != sibling {
		t.Errorf("LocateScript = %q, want %q", got, sibling)
	}
}

func TestUninstallProjectScope(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := Install(ScopeProject, "/p/debate-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(ScopeProject); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallNonexistentScopeError(t *testing.T) {
	if err := Uninstall(Scope(99)); err == nil {
		t.Error("expected error for unknown scope")
	}
}

func TestInstallNonexistentScopeError(t *testing.T) {
	if err := Install(Scope(99), "/p/x.sh"); err == nil {
		t.Error("expected error for unknown scope")
	}
}

func TestEntryReferencesDebate_NoCommand(t *testing.T) {
	if entryReferencesDebate(map[string]any{}) {
		t.Error("entry without hooks should not match")
	}
}

func TestLocateScript_FindsScriptsUnderCwd(t *testing.T) {
	// Make sure no sibling shadows us.
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	sibling := filepath.Join(filepath.Dir(exe), "debate-stop-hook.sh")
	_ = os.Remove(sibling)

	tmp := t.TempDir()
	t.Chdir(tmp)
	if err := os.MkdirAll(filepath.Join(tmp, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tmp, "scripts", "debate-stop-hook.sh")
	if err := os.WriteFile(want, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := LocateScript()
	// Resolve symlinks to handle macOS /tmp -> /private/tmp.
	gotResolved, _ := filepath.EvalSymlinks(got)
	wantResolved, _ := filepath.EvalSymlinks(want)
	if gotResolved != wantResolved {
		t.Errorf("LocateScript = %q (resolved %q), want %q (resolved %q)",
			got, gotResolved, want, wantResolved)
	}
}
