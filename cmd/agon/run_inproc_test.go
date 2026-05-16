// In-process Run() integration test. Mirrors the shape of
// e2e/cli/full_test.go but invokes Run() directly so coverage
// instrumentation propagates to cmd/debate/main.go's Run + helpers.

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"latere.ai/x/agon/internal/cli"
)

func TestRun_InProcess_HappyPath(t *testing.T) {
	root := repoRootHere(t)
	binDir := t.TempDir()

	// Build mock claude + codex into binDir.
	buildMock(t, root, "./e2e/mock/claudemock", filepath.Join(binDir, "claude"))
	buildMock(t, root, "./e2e/mock/codexmock", filepath.Join(binDir, "codex"))

	// Mock claude script: every call returns a "concede c1-1" result.
	scriptPath := filepath.Join(binDir, "claude-script.json")
	rules := []map[string]any{
		{
			"args_contains": "--fork-session",
			"stdout": map[string]any{
				"type": "result", "subtype": "success",
				"session_id": "fork-1",
				"result":     "concede c1-1: I'll add a parameterized query.",
				"is_error":   false,
				"usage":      map[string]any{"input_tokens": 100, "output_tokens": 50},
			},
		},
		{
			"stdout": map[string]any{
				"type": "result", "subtype": "success",
				"session_id": "fork-1",
				"result":     "Nothing further on this round.",
				"is_error":   false,
				"usage":      map[string]any{"input_tokens": 50, "output_tokens": 25},
			},
		},
	}
	if b, err := json.Marshal(rules); err != nil {
		t.Fatal(err)
	} else if err := os.WriteFile(scriptPath, b, 0o644); err != nil {
		t.Fatal(err)
	}

	criticContent := "# Critic 1 - round 1 attacks\n\n" +
		"aspect: security\n\n" +
		"## c1-1 [search.go:1]\n\n" +
		"claim: SQL injection via unparameterized LIKE.\n\n" +
		"expected violation: panic-free injection of boolean logic.\n\n" +
		"reproduction:\n```\ncurl 'http://localhost/search?q=%25%27 OR 1=1--'\n```\n"

	repo := makeFixtureRepo(t)
	stateDir := filepath.Join(repo, ".debate")

	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOCK_CLAUDE_SCRIPT", scriptPath)
	t.Setenv("MOCK_CODEX_CONTENT", criticContent)
	t.Setenv("MOCK_CODEX_THREAD_ID", "mock-thread")
	t.Setenv("DEBATE_IN_PROGRESS", "")
	if err := os.Unsetenv("DEBATE_IN_PROGRESS"); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	flags := &cli.Flags{
		Main:            "claude",
		Side:            "codex",
		MaxTurn:         4,
		SideCount:       1,
		ChangedLinesMin: 5,
		CostCap:         50000,
		StateDir:        stateDir,
		TaskContext:     "search handler",
		DiffFrom:        "HEAD",
		DiffTo:          ".",
		Format:          "markdown",
		Judge:           "none",
		LogMode:         "concise",
	}
	plan := &cli.Plan{
		Cwd:         repo,
		Forks:       []cli.ForkPlan{{Index: 1}},
		StateDirAbs: stateDir,
	}

	if code, err := Run(context.Background(), flags, plan); err != nil {
		t.Logf("Run returned code=%d err=%v", code, err)
	} else {
		t.Logf("Run returned code=%d", code)
	}

	// Verify session folder + artifacts exist.
	sessions, err := os.ReadDir(filepath.Join(stateDir, "sessions"))
	if err != nil {
		t.Fatalf("sessions dir missing: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session folder, got %d", len(sessions))
	}
	sess := filepath.Join(stateDir, "sessions", sessions[0].Name())

	for _, name := range []string{"start.json", "summary.md", "end.json"} {
		p := filepath.Join(sess, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s missing: %v", name, err)
		}
	}
}

func TestRun_InProcess_HookMode(t *testing.T) {
	root := repoRootHere(t)
	binDir := t.TempDir()
	buildMock(t, root, "./e2e/mock/claudemock", filepath.Join(binDir, "claude"))
	buildMock(t, root, "./e2e/mock/codexmock", filepath.Join(binDir, "codex"))

	// Mock claude: always responds.
	scriptPath := filepath.Join(binDir, "claude-script.json")
	rules := []map[string]any{
		{"stdout": map[string]any{
			"type": "result", "subtype": "success",
			"session_id": "fork-1",
			"result":     "ok",
			"is_error":   false,
			"usage":      map[string]any{"input_tokens": 10, "output_tokens": 5},
		}},
	}
	b, _ := json.Marshal(rules)
	_ = os.WriteFile(scriptPath, b, 0o644)

	criticContent := "# Critic 1 - round 1 attacks\n\naspect: security\n"
	repo := makeFixtureRepo(t)
	stateDir := filepath.Join(repo, ".debate")

	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOCK_CLAUDE_SCRIPT", scriptPath)
	t.Setenv("MOCK_CODEX_CONTENT", criticContent)
	if err := os.Unsetenv("DEBATE_IN_PROGRESS"); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	flags := &cli.Flags{
		Main:            "claude",
		Side:            "codex",
		MaxTurn:         2,
		SideCount:       1,
		ChangedLinesMin: 5,
		CostCap:         50000,
		StateDir:        stateDir,
		TaskContext:     "hook-mode",
		DiffFrom:        "HEAD",
		DiffTo:          ".",
		Format:          "markdown",
		Judge:           "none",
		LogMode:         "silent",
		HookMode:        true,
	}
	plan := &cli.Plan{
		Cwd:         repo,
		Forks:       []cli.ForkPlan{{Index: 1}},
		StateDirAbs: stateDir,
		HookMode:    true,
	}
	code, err := Run(context.Background(), flags, plan)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if code != 0 {
		t.Errorf("hook-mode should exit 0 always; got %d", code)
	}
}

func TestRun_InProcess_VerboseMode(t *testing.T) {
	root := repoRootHere(t)
	binDir := t.TempDir()
	buildMock(t, root, "./e2e/mock/claudemock", filepath.Join(binDir, "claude"))
	buildMock(t, root, "./e2e/mock/codexmock", filepath.Join(binDir, "codex"))

	scriptPath := filepath.Join(binDir, "claude-script.json")
	rules := []map[string]any{
		{"stdout": map[string]any{
			"type": "result", "subtype": "success",
			"session_id": "fork-1",
			"result":     "ok",
			"is_error":   false,
			"usage":      map[string]any{"input_tokens": 10, "output_tokens": 5},
		}},
	}
	b, _ := json.Marshal(rules)
	_ = os.WriteFile(scriptPath, b, 0o644)

	repo := makeFixtureRepo(t)
	stateDir := filepath.Join(repo, ".debate")

	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOCK_CLAUDE_SCRIPT", scriptPath)
	t.Setenv("MOCK_CODEX_CONTENT", "# Critic 1 - round 1 attacks\n\naspect: security\n")
	if err := os.Unsetenv("DEBATE_IN_PROGRESS"); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	flags := &cli.Flags{
		Main:            "claude",
		Side:            "codex",
		MaxTurn:         2,
		SideCount:       1,
		ChangedLinesMin: 5,
		CostCap:         50000,
		StateDir:        stateDir,
		TaskContext:     "verbose-mode",
		DiffFrom:        "HEAD",
		DiffTo:          ".",
		Format:          "markdown",
		Judge:           "none",
		LogMode:         "verbose",
	}
	plan := &cli.Plan{
		Cwd:         repo,
		Forks:       []cli.ForkPlan{{Index: 1}},
		StateDirAbs: stateDir,
	}
	if _, err := Run(context.Background(), flags, plan); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_InProcess_WorkingTreeCleanFallback(t *testing.T) {
	root := repoRootHere(t)
	binDir := t.TempDir()
	buildMock(t, root, "./e2e/mock/claudemock", filepath.Join(binDir, "claude"))
	buildMock(t, root, "./e2e/mock/codexmock", filepath.Join(binDir, "codex"))

	scriptPath := filepath.Join(binDir, "claude-script.json")
	rules := []map[string]any{
		{"stdout": map[string]any{
			"type": "result", "subtype": "success",
			"session_id": "fork-1", "result": "ok",
			"is_error": false,
			"usage":    map[string]any{"input_tokens": 10, "output_tokens": 5},
		}},
	}
	b, _ := json.Marshal(rules)
	_ = os.WriteFile(scriptPath, b, 0o644)

	// Make a repo where the working tree is CLEAN but HEAD~1..HEAD has a
	// real diff. Run() should auto-fall-back to HEAD~1..HEAD instead of
	// short-circuiting on an empty diff.
	dir := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q"},
		{
			"git", "-c", "user.email=t@e.com", "-c", "user.name=t",
			"commit", "--allow-empty", "-q", "-m", "init",
		},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, b)
		}
	}
	body := strings.Repeat("// content\n", 30)
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][]string{
		{"git", "add", "."},
		{
			"git", "-c", "user.email=t@e.com", "-c", "user.name=t",
			"commit", "-q", "-m", "add x.go",
		},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, b)
		}
	}
	// Now working tree is clean; HEAD~1..HEAD has 30 lines of new content.

	stateDir := filepath.Join(dir, ".debate")
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOCK_CLAUDE_SCRIPT", scriptPath)
	t.Setenv("MOCK_CODEX_CONTENT", "# Critic 1 - round 1 attacks\n\naspect: security\n")
	if err := os.Unsetenv("DEBATE_IN_PROGRESS"); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	flags := &cli.Flags{
		Main: "claude", Side: "codex",
		MaxTurn: 2, SideCount: 1,
		ChangedLinesMin: 5, CostCap: 50000,
		StateDir:    stateDir,
		TaskContext: "fallback test",
		DiffFrom:    "HEAD", DiffTo: ".",
		Format: "markdown", Judge: "none", LogMode: "silent",
	}
	plan := &cli.Plan{
		Cwd:         dir,
		Forks:       []cli.ForkPlan{{Index: 1}},
		StateDirAbs: stateDir,
	}
	if _, err := Run(context.Background(), flags, plan); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// The auto-fallback message should have set DiffFrom/DiffTo to
	// HEAD~1..HEAD and produced a session.
	if flags.DiffFrom != "HEAD~1" || flags.DiffTo != "HEAD" {
		t.Errorf("auto-fallback did not adjust diff range: from=%q to=%q",
			flags.DiffFrom, flags.DiffTo)
	}
	sessions, _ := os.ReadDir(filepath.Join(stateDir, "sessions"))
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
}

func TestRun_InProcess_TrivialDiffShortCircuit(t *testing.T) {
	repo := makeFixtureRepo(t)
	stateDir := filepath.Join(repo, ".debate")
	t.Chdir(repo)

	flags := &cli.Flags{
		Main:            "claude",
		Side:            "codex",
		MaxTurn:         2,
		SideCount:       1,
		ChangedLinesMin: 9999, // anything we computed is below this; gate fires
		CostCap:         50000,
		StateDir:        stateDir,
		TaskContext:     "trivial",
		DiffFrom:        "HEAD",
		DiffTo:          ".",
		Format:          "markdown",
		Judge:           "none",
		LogMode:         "silent",
	}
	plan := &cli.Plan{
		Cwd:         repo,
		Forks:       []cli.ForkPlan{{Index: 1}},
		StateDirAbs: stateDir,
	}

	if code, err := Run(context.Background(), flags, plan); err != nil {
		t.Errorf("trivial-diff path errored: code=%d err=%v", code, err)
	}

	// Trivial path: no session dir, only log.jsonl.
	if _, err := os.Stat(filepath.Join(stateDir, "log.jsonl")); err != nil {
		t.Errorf("log.jsonl missing on trivial path: %v", err)
	}
	sessions, err := os.ReadDir(filepath.Join(stateDir, "sessions"))
	if err == nil && len(sessions) != 0 {
		t.Errorf("trivial path created session dir: %v", sessions)
	}
}

// repoRootHere is the cmd/debate-relative helper.
func repoRootHere(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func buildMock(t *testing.T, root, mainPath, outPath string) {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", outPath, mainPath)
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", mainPath, err, b)
	}
}

func makeFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q"},
		{
			"git", "-c", "user.email=t@e.com", "-c", "user.name=t",
			"commit", "--allow-empty", "-q", "-m", "init",
		},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, b)
		}
	}
	body := strings.Repeat("// fixture line\n", 30)
	if err := os.WriteFile(filepath.Join(dir, "search.go"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
