// Package cli_test runs end-to-end tests against the actual `bin/debate`
// binary, with claude and codex resolved via PATH to mock harnesses.
//
// Coverage: the full v0 happy path — preflight, transcript locate +
// task context extraction, working-tree diff, per-fork serial loop,
// critic → proposer round trips through the real subprocess layer,
// ledger writes, summary.md render, log.jsonl append, exit code
// translation.
//
// Run locally: `make e2e` or `go test ./e2e/cli/...`. No real
// claude/codex install required.
package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// repoRoot returns the absolute path to the debate repo (where this
// test file lives is e2e/cli/, two levels deep).
func repoRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// build compiles the named main package into binDir/<basename>.
func build(t *testing.T, root, mainPath, binDir string) string {
	t.Helper()
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(binDir, filepath.Base(mainPath))
	if filepath.Base(mainPath) == "debate" {
		out = filepath.Join(binDir, "debate")
	}
	cmd := exec.Command("go", "build", "-o", out, mainPath)
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", mainPath, err, b)
	}
	return out
}

// mockClaudeRules is the matcher script the mock claude binary will
// consume via $MOCK_CLAUDE_SCRIPT.
type matcher struct {
	ArgsContains string         `json:"args_contains"`
	Stdout       map[string]any `json:"stdout"`
	Stderr       string         `json:"stderr"`
	Exit         int            `json:"exit"`
}

func writeClaudeScript(t *testing.T, dir string, rules []matcher) string {
	t.Helper()
	p := filepath.Join(dir, "mock-claude-script.json")
	b, err := json.Marshal(rules)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// fixtureRepo creates a tiny git repo with a real working-tree diff.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init", "-q"},
		{"git", "-c", "user.email=t@e.com", "-c", "user.name=t", "commit", "--allow-empty", "-q", "-m", "init"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, b)
		}
	}
	// Add a non-trivial diff.
	if err := os.WriteFile(filepath.Join(dir, "search.go"),
		[]byte(strings.Repeat("// fixture line\n", 30)), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runDebate invokes the binary with the given args and captures
// stdout/stderr/exit code.
type runResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runDebate(t *testing.T, debate string, env []string, cwd string, args ...string) runResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, debate, args...)
	cmd.Dir = cwd
	cmd.Env = env
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := runResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		var ee *exec.ExitError
		if asExit(err, &ee) {
			res.ExitCode = ee.ExitCode()
		} else {
			t.Logf("run error (not ExitError): %v", err)
			res.ExitCode = -1
		}
	}
	return res
}

func asExit(err error, target **exec.ExitError) bool {
	for e := err; e != nil; {
		if x, ok := e.(*exec.ExitError); ok {
			*target = x
			return true
		}
		type unwrap interface{ Unwrap() error }
		if u, ok := e.(unwrap); ok {
			e = u.Unwrap()
			continue
		}
		break
	}
	return false
}

// stripParentPath returns the shell PATH with the binDir prepended so
// the mock binaries shadow any real claude/codex.
func patchedPATH(t *testing.T, mockBinDir string) []string {
	t.Helper()
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			continue
		}
		if strings.HasPrefix(kv, "ANTHROPIC_API_KEY=") {
			continue
		}
		if strings.HasPrefix(kv, "DEBATE_IN_PROGRESS=") {
			continue
		}
		out = append(out, kv)
	}
	out = append(out, fmt.Sprintf("PATH=%s:%s", mockBinDir, os.Getenv("PATH")))
	return out
}

func TestFullE2E_HappyPath(t *testing.T) {
	root := repoRoot(t)
	binDir := t.TempDir()

	// Build debate + mocks.
	debate := build(t, root, "./cmd/debate", binDir)
	build(t, root, "./e2e/mock/claudemock", binDir)
	build(t, root, "./e2e/mock/codexmock", binDir)
	// PATH lookup expects the binaries to be named "claude" / "codex".
	if err := os.Rename(filepath.Join(binDir, "claudemock"), filepath.Join(binDir, "claude")); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(binDir, "codexmock"), filepath.Join(binDir, "codex")); err != nil {
		t.Fatal(err)
	}

	// Fixture repo with a non-trivial diff.
	repo := fixtureRepo(t)

	// Mock claude: every call returns a result containing "concede c1-1".
	scriptPath := writeClaudeScript(t, binDir, []matcher{
		{
			ArgsContains: "--fork-session",
			Stdout: map[string]any{
				"type": "result", "subtype": "success",
				"session_id": "fork-1",
				"result":     "concede c1-1: I'll add a parameterized query.",
				"is_error":   false,
				"usage":      map[string]any{"input_tokens": 100, "output_tokens": 50},
			},
		},
		{
			Stdout: map[string]any{
				"type": "result", "subtype": "success",
				"session_id": "fork-1",
				"result":     "Nothing further on this round.",
				"is_error":   false,
				"usage":      map[string]any{"input_tokens": 50, "output_tokens": 25},
			},
		},
	})

	// Mock codex: emit a 1-attack critic markdown that the proposer
	// (mock claude above) will concede.
	criticContent := "# Critic 1 — round 1 attacks\n\n" +
		"aspect: security\n\n" +
		"## c1-1 [search.go:1]\n\n" +
		"claim: SQL injection via unparameterized LIKE.\n\n" +
		"expected violation: panic-free injection of boolean logic.\n\n" +
		"reproduction:\n```\ncurl 'http://localhost/search?q=%25%27 OR 1=1--'\n```\n"

	stateDir := filepath.Join(repo, ".debate")
	env := patchedPATH(t, binDir)
	env = append(env,
		"MOCK_CLAUDE_SCRIPT="+scriptPath,
		"MOCK_CODEX_CONTENT="+criticContent,
		"MOCK_CODEX_THREAD_ID=mock-thread",
		"HOME="+t.TempDir(),
	)

	res := runDebate(t, debate, env, repo,
		"--task-context", "search handler",
		"--main", "claude",
		"--side", "codex",
		"--side-count", "1",
		"--aspect", "security",
		"--max-turn", "4",
		"--changed-lines-min", "5",
		"--state-dir", stateDir,
	)

	if res.ExitCode != 0 && res.ExitCode != 1 {
		t.Fatalf("unexpected exit %d\nstdout: %s\nstderr: %s", res.ExitCode, res.Stdout, res.Stderr)
	}

	// Verify session folder exists.
	sessions, err := os.ReadDir(filepath.Join(stateDir, "sessions"))
	if err != nil {
		t.Fatalf("sessions dir missing: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session folder, got %d", len(sessions))
	}
	sess := filepath.Join(stateDir, "sessions", sessions[0].Name())

	// Verify start.json schema.
	startBody, err := os.ReadFile(filepath.Join(sess, "start.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(startBody), `"schema": "debate.start.v0"`) {
		t.Errorf("start.json missing schema; body=%s", startBody)
	}

	// Verify summary.md exists with expected sections.
	sumBody, err := os.ReadFile(filepath.Join(sess, "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	sum := string(sumBody)
	if !strings.Contains(sum, "## Stats") {
		t.Errorf("summary missing Stats section: %s", sum)
	}
	if !strings.Contains(sum, "critic-found-bug rate") {
		t.Errorf("summary missing rate stat: %s", sum)
	}

	// Verify end.json carries termination + exit_code.
	endBody, err := os.ReadFile(filepath.Join(sess, "end.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(endBody), `"termination":`) {
		t.Errorf("end.json missing termination: %s", endBody)
	}

	// Verify attacks.jsonl has at least one record.
	atkBody, err := os.ReadFile(filepath.Join(sess, "attacks.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(atkBody), `"attack_id":"c1-1"`) {
		t.Errorf("attacks.jsonl missing c1-1: %s", atkBody)
	}

	// Verify cross-session log has the run line.
	logBody, err := os.ReadFile(filepath.Join(stateDir, "log.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logBody), `"kind":"run"`) {
		t.Errorf("log.jsonl missing run record: %s", logBody)
	}
}

func TestFullE2E_TrivialDiffSkip(t *testing.T) {
	root := repoRoot(t)
	binDir := t.TempDir()
	debate := build(t, root, "./cmd/debate", binDir)

	// Empty repo (no diff).
	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q"},
		{"git", "-c", "user.email=t@e.com", "-c", "user.name=t", "commit", "--allow-empty", "-q", "-m", "init"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, b)
		}
	}

	stateDir := filepath.Join(repo, ".debate")
	env := patchedPATH(t, binDir)

	res := runDebate(t, debate, env, repo,
		"--task-context", "no changes",
		"--side-count", "1", "--aspect", "security",
		"--max-turn", "4", "--changed-lines-min", "10",
		"--state-dir", stateDir,
	)

	if res.ExitCode != 0 {
		t.Fatalf("trivial diff should exit 0, got %d\nstderr: %s", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stderr, "[debate] skipped") {
		t.Errorf("expected skip line on stderr; got: %s", res.Stderr)
	}
	// log.jsonl should have a kind=skipped record.
	logBody, err := os.ReadFile(filepath.Join(stateDir, "log.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logBody), `"kind":"skipped"`) {
		t.Errorf("log.jsonl missing skipped record: %s", logBody)
	}
	// No session folder should be created.
	if _, err := os.Stat(filepath.Join(stateDir, "sessions")); !os.IsNotExist(err) {
		t.Errorf("trivial path should not create sessions/: %v", err)
	}
}

func TestFullE2E_RecursionGuard(t *testing.T) {
	root := repoRoot(t)
	binDir := t.TempDir()
	debate := build(t, root, "./cmd/debate", binDir)

	repo := fixtureRepo(t)
	env := patchedPATH(t, binDir)
	env = append(env, "DEBATE_IN_PROGRESS=1")

	res := runDebate(t, debate, env, repo,
		"--task-context", "anything",
		"--side-count", "1", "--aspect", "security",
		"--state-dir", filepath.Join(repo, ".debate"),
	)
	if res.ExitCode != 0 {
		t.Fatalf("recursion guard should exit 0, got %d\nstderr: %s", res.ExitCode, res.Stderr)
	}
	// Should have produced no session.
	if _, err := os.Stat(filepath.Join(repo, ".debate", "sessions")); !os.IsNotExist(err) {
		t.Error("recursion-guard short-circuit should not create sessions/")
	}
}

func TestFullE2E_PreflightExitCode(t *testing.T) {
	root := repoRoot(t)
	binDir := t.TempDir()
	debate := build(t, root, "./cmd/debate", binDir)

	repo := fixtureRepo(t)
	env := patchedPATH(t, binDir)

	// --side-count != len(--aspect) → exit 120
	res := runDebate(t, debate, env, repo,
		"--task-context", "x",
		"--side-count", "3",
		"--aspect", "security,functional-logic",
		"--state-dir", filepath.Join(repo, ".debate"),
	)
	if res.ExitCode != 120 {
		t.Errorf("expected exit 120, got %d\nstderr: %s", res.ExitCode, res.Stderr)
	}
}
