// Package hook_test runs the bash Stop-hook script against a fake
// agon binary so we can assert env scrubbing, recursion-guard
// behavior, cwd resolution, and the --hook-mode pass-through.
package hook_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const fakeAgon = `#!/usr/bin/env bash
# Records args + selected env vars to $RECORD_FILE, then exits 0.
{
  echo "ARGS:" "$@"
  echo "AGON_IN_PROGRESS=$AGON_IN_PROGRESS"
  echo "ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY-<unset>}"
  echo "PWD=$(pwd)"
} > "$RECORD_FILE"
exit 0
`

func writeFakeAgon(t *testing.T, dir string) string {
	t.Helper()
	// Binary name is external contract, not prose: it must match the
	// `exec agon` in scripts/agon-stop-hook.sh. Phase 2 flips both
	// to `agon` together (+ a `agon` shim); change them in lockstep.
	p := filepath.Join(dir, "agon")
	if err := os.WriteFile(p, []byte(fakeAgon), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func repoRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func TestHookForwardsArgsAndEnv(t *testing.T) {
	root := repoRoot(t)
	hookScript := filepath.Join(root, "scripts", "agon-stop-hook.sh")

	binDir := t.TempDir()
	writeFakeAgon(t, binDir)

	recDir := t.TempDir()
	recordFile := filepath.Join(recDir, "record.txt")
	cwdDir := t.TempDir()

	payload := fmt.Sprintf(`{"session_id":"abc-123","transcript_path":"/tmp/t.jsonl","cwd":"%s"}`, cwdDir)

	cmd := exec.Command("bash", hookScript)
	cmd.Stdin = strings.NewReader(payload)
	cmd.Env = []string{
		"PATH=" + binDir + ":" + os.Getenv("PATH"),
		"RECORD_FILE=" + recordFile,
		"ANTHROPIC_API_KEY=stale", // hook script must scrub this
		// no AGON_IN_PROGRESS - should be set by the hook before exec
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook exec: %v", err)
	}

	body, err := os.ReadFile(recordFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	for _, want := range []string{
		"--hook-mode",
		"--session-id abc-123",
		"--transcript /tmp/t.jsonl",
		"--max-turn 6",
		"AGON_IN_PROGRESS=1",
		"ANTHROPIC_API_KEY=<unset>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in hook record:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "PWD="+cwdDir) {
		t.Errorf("hook should cd to %q before exec; record:\n%s", cwdDir, got)
	}
}

func TestHookRecursionGuardShortCircuit(t *testing.T) {
	root := repoRoot(t)
	hookScript := filepath.Join(root, "scripts", "agon-stop-hook.sh")

	binDir := t.TempDir()
	writeFakeAgon(t, binDir)

	recDir := t.TempDir()
	recordFile := filepath.Join(recDir, "record.txt")

	cmd := exec.Command("bash", hookScript)
	cmd.Stdin = strings.NewReader(`{"session_id":"abc","transcript_path":"","cwd":""}`)
	cmd.Env = []string{
		"PATH=" + binDir + ":" + os.Getenv("PATH"),
		"RECORD_FILE=" + recordFile,
		"AGON_IN_PROGRESS=1", // guard already set
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook should exit 0 under recursion guard: %v", err)
	}
	if _, err := os.Stat(recordFile); !os.IsNotExist(err) {
		t.Errorf("recursion guard should NOT exec agon; record file exists: %v", err)
	}
}

func TestHookHandlesEmptyPayload(t *testing.T) {
	root := repoRoot(t)
	hookScript := filepath.Join(root, "scripts", "agon-stop-hook.sh")

	binDir := t.TempDir()
	writeFakeAgon(t, binDir)

	recDir := t.TempDir()
	recordFile := filepath.Join(recDir, "record.txt")

	cmd := exec.Command("bash", hookScript)
	cmd.Stdin = strings.NewReader(`{}`)
	cmd.Env = []string{
		"PATH=" + binDir + ":" + os.Getenv("PATH"),
		"RECORD_FILE=" + recordFile,
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("hook on empty payload: %v", err)
	}
	body, _ := os.ReadFile(recordFile)
	got := string(body)
	if !strings.Contains(got, "--hook-mode") {
		t.Errorf("hook with empty payload should still pass --hook-mode:\n%s", got)
	}
}
