package mock_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"latere.ai/x/agon/internal/agent"
	"latere.ai/x/agon/internal/critic"
)

func buildMock(t *testing.T, name string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", out, "./e2e/mock/"+name)
	cmd.Dir = repoRoot(t)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v: %s", name, err, b)
	}
	return out
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

func TestCodexMockEndToEnd(t *testing.T) {
	bin := buildMock(t, "codexmock")
	t.Setenv("MOCK_CODEX_CONTENT", "# Critic 1 - round 1 attacks\n\naspect: security\n\n## c1-1 [x.go:1]\n\nclaim: leak\n\nexpected violation: panic\n\nreproduction:\n```\ngo test\n```\n")

	c := &agent.CodexCritic{Bin: bin}
	res, err := c.Round(context.Background(), agent.CriticInput{
		Aspect: critic.Lookup("security"), CriticIndex: 1, Round: 1,
		SystemPrompt: "x", TaskContext: "t", DiffPatch: "d",
		Cwd: t.TempDir(), Deadline: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Markdown, "Critic 1") {
		t.Errorf("got: %q", res.Markdown)
	}
}

func TestClaudeMockProposerR1(t *testing.T) {
	bin := buildMock(t, "claudemock")
	p := &agent.ClaudeProposer{
		Bin: bin, Cwd: t.TempDir(), RootID: "abc",
	}
	res, err := p.FirstRound(context.Background(), "ping")
	if err != nil {
		t.Fatal(err)
	}
	if res.ForkID != "mock-session" {
		t.Errorf("ForkID: got %q, want mock-session", res.ForkID)
	}
	if res.Response != "ok" {
		t.Errorf("Response: got %q", res.Response)
	}
}
