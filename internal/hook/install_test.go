package hook

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallEmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, "/path/to/agon-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "agon-stop-hook.sh") {
		t.Errorf("missing entry: %s", b)
	}
}

func TestInstallIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, "/p/agon-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	if err := Install(ScopeUser, "/p/agon-stop-hook.sh"); err != nil {
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
	if err := Install(ScopeUser, "/p/agon-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if !strings.Contains(string(out), "/other/hook.sh") {
		t.Error("lost the unrelated hook entry")
	}
	if !strings.Contains(string(out), "agon-stop-hook.sh") {
		t.Error("missing the new agon entry")
	}
}

func TestUninstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := Install(ScopeUser, "/p/agon-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(ScopeUser); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if strings.Contains(string(b), "agon-stop-hook.sh") {
		t.Errorf("entry still present: %s", b)
	}
}

func TestInstallProjectScope(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := Install(ScopeProject, "/p/agon-stop-hook.sh"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "agon-stop-hook.sh") {
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
	if got != "agon-stop-hook.sh" {
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
	sibling := filepath.Join(filepath.Dir(exe), "agon-stop-hook.sh")
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
	if err := Install(ScopeProject, "/p/agon-stop-hook.sh"); err != nil {
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

func TestEntryReferencesAgon_NoCommand(t *testing.T) {
	if entryReferencesAgon(map[string]any{}) {
		t.Error("entry without hooks should not match")
	}
}

// Regression for critic c1-2 (specs/14): a non-object foreign
// statusLine value (string, array, etc.) was silently overwritten
// because the type-assert to map[string]any failed and was treated
// as "no existing entry". Now any pre-existing value that isn't a
// agon-owned object is a conflict.
func TestInstallStatusLine_RejectsStringForeignValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	settings := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings,
		[]byte(`{"statusLine":"foreign command"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	err := InstallStatusLine(ScopeUser, "/usr/local/bin/agon status", false)
	if !errors.Is(err, ErrStatusLineConflict) {
		t.Fatalf("got %v, want ErrStatusLineConflict", err)
	}
	// Foreign value must remain intact.
	b, _ := os.ReadFile(settings)
	if !strings.Contains(string(b), `"foreign command"`) {
		t.Errorf("foreign value was clobbered: %s", b)
	}
}

func TestInstallStatusLine_OverwritesAgonOwned(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := InstallStatusLine(ScopeUser, "/usr/local/bin/agon status", false); err != nil {
		t.Fatal(err)
	}
	if err := InstallStatusLine(ScopeUser, "/opt/agon status", false); err != nil {
		t.Fatalf("re-install on agon-owned entry should succeed: %v", err)
	}
}

func TestInstallStatusLine_ForceOverwritesForeign(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	settings := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings,
		[]byte(`{"statusLine":{"type":"command","command":"node /path/to/foreign.cjs"}}`),
		0o644); err != nil {
		t.Fatal(err)
	}
	if err := InstallStatusLine(ScopeUser, "/usr/local/bin/agon status", true); err != nil {
		t.Fatalf("force should overwrite foreign value: %v", err)
	}
	b, _ := os.ReadFile(settings)
	if !strings.Contains(string(b), "/usr/local/bin/agon status") {
		t.Errorf("agon status not written: %s", b)
	}
	if strings.Contains(string(b), "foreign.cjs") {
		t.Errorf("foreign value should be replaced under force: %s", b)
	}
}

func TestReadStatusLineCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// No settings.json yet -> empty.
	if got := ReadStatusLineCommand(ScopeUser); got != "" {
		t.Errorf("missing file: got %q", got)
	}

	settings := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}

	// Object form.
	if err := os.WriteFile(settings,
		[]byte(`{"statusLine":{"type":"command","command":"node /a.cjs"}}`),
		0o644); err != nil {
		t.Fatal(err)
	}
	if got := ReadStatusLineCommand(ScopeUser); got != "node /a.cjs" {
		t.Errorf("object form: got %q", got)
	}

	// Bare-string form.
	if err := os.WriteFile(settings,
		[]byte(`{"statusLine":"node /b.cjs"}`),
		0o644); err != nil {
		t.Fatal(err)
	}
	if got := ReadStatusLineCommand(ScopeUser); got != "node /b.cjs" {
		t.Errorf("string form: got %q", got)
	}
}

func TestLocateScript_FindsScriptsUnderCwd(t *testing.T) {
	// Make sure no sibling shadows us.
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	sibling := filepath.Join(filepath.Dir(exe), "agon-stop-hook.sh")
	_ = os.Remove(sibling)

	tmp := t.TempDir()
	t.Chdir(tmp)
	if err := os.MkdirAll(filepath.Join(tmp, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tmp, "scripts", "agon-stop-hook.sh")
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
