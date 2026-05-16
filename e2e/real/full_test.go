//go:build real_e2e

// Package real_test runs end-to-end tests against the actual `claude`
// and `codex` binaries on the maintainer's PATH. Gated by both the
// `real_e2e` build tag and the runtime presence of both binaries.
//
// Run locally: RUN_REAL=1 go test -tags real_e2e -timeout 30m ./e2e/real/...
//
// Closes [27-release.md](../../specs/27-release.md) G13 per
// [32-real-e2e-suite.md](../../specs/32-real-e2e-suite.md) Path A.
package real_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRealEndToEnd_OneFork is the only test in this file by design (per
// spec 32). One fork against real claude + real codex; asserts shape,
// not content.
func TestRealEndToEnd_OneFork(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not on PATH; skipping real-e2e")
	}
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not on PATH; skipping real-e2e")
	}

	root := repoRoot(t)
	binDir := t.TempDir()
	agon := build(t, root, binDir)

	repo := fixtureRepoWith47LineDiff(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, agon,
		"--main", "claude",
		"--side", "codex",
		"--max-turn", "4",
		"--aspects", "security",
		"--changed-lines-min", "10",
	)
	cmd.Dir = repo
	cmd.Env = scrubbedEnv()
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	err := cmd.Run()
	wall := time.Since(start)

	t.Logf("wall: %s", wall)
	t.Logf("stdout: %s", stdout.String())
	t.Logf("stderr: %s", stderr.String())

	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("agon run failed (non-exit error): %v", err)
		}
		// Exit codes 0 (clean) and 1 (unresolved leaves) are both fine for
		// a real-e2e shape check; anything else is failure.
		if ee.ExitCode() != 1 {
			t.Fatalf("agon exit code %d, expected 0 or 1", ee.ExitCode())
		}
	}

	// Find the latest session dir.
	sessions := filepath.Join(repo, ".agon", "sessions")
	entries, err := os.ReadDir(sessions)
	if err != nil {
		t.Fatalf("read sessions dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("no session directory under %s", sessions)
	}
	sessionDir := filepath.Join(sessions, entries[len(entries)-1].Name())

	// summary.md exists and is non-empty.
	summary, err := os.ReadFile(filepath.Join(sessionDir, "summary.md"))
	if err != nil {
		t.Fatalf("read summary.md: %v", err)
	}
	if len(summary) == 0 {
		t.Fatalf("summary.md is empty")
	}

	// attacks.jsonl: every line must parse as JSON. Empty file is allowed
	// (a fork that found no attacks is a valid shape).
	atkBytes, err := os.ReadFile(filepath.Join(sessionDir, "attacks.jsonl"))
	if err != nil {
		t.Fatalf("read attacks.jsonl: %v", err)
	}
	for i, line := range splitLines(atkBytes) {
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			t.Fatalf("attacks.jsonl line %d not JSON: %v\n%s", i+1, err, line)
		}
	}

	// At least one fork has an r1-critic.md (the critic ran).
	forks, err := os.ReadDir(filepath.Join(sessionDir, "forks"))
	if err != nil {
		t.Fatalf("read forks dir: %v", err)
	}
	if len(forks) == 0 {
		t.Fatalf("no forks under %s", sessionDir)
	}
	found := false
	for _, f := range forks {
		p := filepath.Join(sessionDir, "forks", f.Name(), "rounds", "r1-critic.md")
		if _, err := os.Stat(p); err == nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no r1-critic.md under any fork; the critic never ran")
	}

	t.Logf("real-e2e PASS: session=%s wall=%s", sessionDir, wall)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func build(t *testing.T, root, binDir string) string {
	t.Helper()
	out := filepath.Join(binDir, "agon")
	cmd := exec.Command("go", "build", "-o", out, "./cmd/agon")
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build agon: %v\n%s", err, b)
	}
	return out
}

func fixtureRepoWith47LineDiff(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init", "-q"},
		{"git", "-c", "user.email=t@e.com", "-c", "user.name=t",
			"commit", "--allow-empty", "-q", "-m", "init"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, b)
		}
	}
	// Realistic-shaped diff with security implications so the security
	// critic has something to grip on.
	body := strings.Repeat("// fixture line for real-e2e\n", 47)
	if err := os.WriteFile(filepath.Join(dir, "search.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// scrubbedEnv returns os.Environ() (ANTHROPIC_API_KEY preserved;
// real-e2e needs it).
func scrubbedEnv() []string {
	return os.Environ()
}

func splitLines(b []byte) []string {
	s := string(b)
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}
