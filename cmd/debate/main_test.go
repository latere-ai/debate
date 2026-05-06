package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"latere.ai/x/debate/internal/cli"
)

func TestExitCodeFor(t *testing.T) {
	if got := exitCodeFor(nil); got != 1 {
		t.Errorf("nil err: got %d, want 1", got)
	}
	if got := exitCodeFor(errors.New("boom")); got != 1 {
		t.Errorf("plain err: got %d", got)
	}
	pe := &cli.PreflightError{Code: 42}
	if got := exitCodeFor(pe); got != 42 {
		t.Errorf("preflight: got %d", got)
	}
	// Wrapped preflight error still surfaces the code.
	wrapped := fmt.Errorf("wrapped: %w", &cli.PreflightError{Code: 7})
	if got := exitCodeFor(wrapped); got != 7 {
		t.Errorf("wrapped preflight: got %d, want 7", got)
	}
}

func TestTaskSource(t *testing.T) {
	cases := []struct {
		name string
		in   *cli.Flags
		want string
	}{
		{"flag", &cli.Flags{TaskContext: "x"}, "flag"},
		{"transcript", &cli.Flags{Transcript: "/p"}, "transcript"},
		{"session-id", &cli.Flags{SessionID: "abc"}, "session-id-resume"},
		{"unknown", &cli.Flags{}, "unknown"},
		// Precedence: TaskContext wins over Transcript / SessionID.
		{"flag-wins", &cli.Flags{TaskContext: "x", Transcript: "/p", SessionID: "y"}, "flag"},
		// Transcript wins over SessionID.
		{"transcript-wins", &cli.Flags{Transcript: "/p", SessionID: "y"}, "transcript"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := taskSource(c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestInstallHookCmd_DefaultsToUserScope(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cmd := installHookCmd()
	cmd.SetArgs([]string{"--script-path", "/tmp/fake-debate-stop-hook.sh"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	settings := filepath.Join(dir, ".claude", "settings.json")
	if _, err := os.Stat(settings); err != nil {
		t.Fatalf("settings missing: %v", err)
	}
	b, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(b), "fake-debate-stop-hook.sh") {
		t.Errorf("missing entry: %s", b)
	}
}

func TestInstallHookCmd_ProjectScope(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cmd := installHookCmd()
	cmd.SetArgs([]string{"--scope", "project", "--script-path", "/tmp/fake.sh"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); err != nil {
		t.Fatalf("project settings missing: %v", err)
	}
}

func TestInstallHookCmd_DefaultsLocateScript(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cmd := installHookCmd()
	cmd.SetArgs(nil) // no --script-path; falls back to LocateScript()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if !contains(string(b), "debate-stop-hook.sh") {
		t.Errorf("install-hook with no --script-path did not record any script: %s", b)
	}
}

func TestUninstallHookCmd_DefaultsToUserScope(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Pre-install something to remove.
	inst := installHookCmd()
	inst.SetArgs([]string{"--script-path", "/tmp/fake-debate-stop-hook.sh"})
	if err := inst.Execute(); err != nil {
		t.Fatal(err)
	}

	un := uninstallHookCmd()
	un.SetArgs(nil)
	if err := un.Execute(); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if contains(string(b), "fake-debate-stop-hook.sh") {
		t.Errorf("entry not removed: %s", b)
	}
}

func TestUninstallHookCmd_ProjectScope(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	inst := installHookCmd()
	inst.SetArgs([]string{"--scope", "project", "--script-path", "/tmp/fake.sh"})
	if err := inst.Execute(); err != nil {
		t.Fatal(err)
	}

	un := uninstallHookCmd()
	un.SetArgs([]string{"--scope", "project"})
	if err := un.Execute(); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func TestRealMain_Version(t *testing.T) {
	var buf strings.Builder
	code := realMain([]string{"--version"}, &buf, &buf)
	if code != 0 {
		t.Errorf("--version exit code: got %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "debate") {
		t.Errorf("--version output should mention debate; got %q", buf.String())
	}
}

func TestRealMain_Help(t *testing.T) {
	var buf strings.Builder
	code := realMain([]string{"--help"}, &buf, &buf)
	if code != 0 {
		t.Errorf("--help exit code: got %d, want 0", code)
	}
}

func TestRealMain_BareShowsHelp(t *testing.T) {
	// Bare invocation with no env triggers the help fast-path.
	t.Setenv("DEBATE_TASK_CONTEXT", "")
	t.Setenv("DEBATE_SESSION_ID", "")
	t.Setenv("DEBATE_TRANSCRIPT", "")

	var buf strings.Builder
	code := realMain(nil, &buf, &buf)
	if code != 0 {
		t.Errorf("bare exit code: got %d, want 0", code)
	}
}

func TestRealMain_PreflightExitCode(t *testing.T) {
	// --judge llm is rejected by preflight (v0 only supports 'none').
	var buf strings.Builder
	code := realMain([]string{
		"--task-context", "x",
		"--judge", "llm",
	}, &buf, &buf)
	if code == 0 {
		t.Errorf("expected non-zero exit for --judge llm; got %d", code)
	}
	if !strings.Contains(buf.String(), "debate:") {
		t.Errorf("error line should be prefixed with 'debate:'; got %q", buf.String())
	}
}

func TestRealMain_RecursionGuardSilent(t *testing.T) {
	t.Setenv("DEBATE_IN_PROGRESS", "1")
	var buf strings.Builder
	code := realMain([]string{"--task-context", "x"}, &buf, &buf)
	if code != 0 {
		t.Errorf("recursion guard should exit 0 silently; got code=%d, stderr=%q",
			code, buf.String())
	}
}

func TestRealMain_InstallHookSubcommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	var buf strings.Builder
	code := realMain([]string{"install-hook", "--script-path", "/tmp/x.sh"}, &buf, &buf)
	if code != 0 {
		t.Errorf("install-hook exit: got %d, want 0; stderr=%q", code, buf.String())
	}
	settings := filepath.Join(dir, ".claude", "settings.json")
	if _, err := os.Stat(settings); err != nil {
		t.Errorf("settings missing: %v", err)
	}
}
