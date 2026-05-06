package agent

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/latere-ai/debate/internal/critic"
)

func buildHelper(t *testing.T, name, src string) string {
	t.Helper()
	dir := t.TempDir()
	gosrc := filepath.Join(dir, "main.go")
	if err := os.WriteFile(gosrc, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, name)
	c := exec.Command("go", "build", "-o", out, gosrc)
	if b, err := c.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, b)
	}
	return out
}

func TestCodexCriticEmptyContent(t *testing.T) {
	bin := buildHelper(t, "empty-codex", `package main
import "fmt"
func main() {
	fmt.Println(`+"`"+`{"type":"thread.started","thread_id":"x"}`+"`"+`)
	fmt.Println(`+"`"+`{"type":"thread.completed"}`+"`"+`)
}
`)
	c := &CodexCritic{Bin: bin}
	_, err := c.Round(context.Background(), CriticInput{
		Aspect: critic.Lookup("security"), CriticIndex: 1, Round: 1,
		SystemPrompt: "x", TaskContext: "t", DiffPatch: "d",
		Cwd: t.TempDir(), Deadline: 5 * time.Second,
	})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestCodexCriticStderrMappings(t *testing.T) {
	bin := buildHelper(t, "rl-codex", `package main
import (
	"fmt"
	"os"
)
func main() {
	fmt.Fprintln(os.Stderr, "error: rate limit exceeded")
	os.Exit(1)
}
`)
	c := &CodexCritic{Bin: bin}
	_, err := c.Round(context.Background(), CriticInput{
		Aspect: critic.Lookup("security"), CriticIndex: 1, Round: 1,
		Cwd: t.TempDir(), Deadline: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("got %v", err)
	}
}

func TestClaudeProposerCwdMismatch(t *testing.T) {
	bin := buildHelper(t, "noconv-claude", `package main
import (
	"fmt"
	"os"
)
func main() {
	fmt.Fprintln(os.Stderr, "Error: No conversation found with session ID abc")
	os.Exit(1)
}
`)
	p := &ClaudeProposer{Bin: bin, Cwd: t.TempDir(), RootID: "abc"}
	_, err := p.FirstRound(context.Background(), "ping")
	if err == nil || !strings.Contains(err.Error(), "cwd mismatch") {
		t.Errorf("expected ErrCwdMismatch, got %v", err)
	}
}

func TestClaudeProposerAuthError(t *testing.T) {
	bin := buildHelper(t, "auth-claude", `package main
import (
	"fmt"
	"os"
)
func main() {
	fmt.Fprintln(os.Stderr, "Authentication error: 401 Unauthorized")
	os.Exit(1)
}
`)
	p := &ClaudeProposer{Bin: bin, Cwd: t.TempDir(), RootID: "abc"}
	_, err := p.FirstRound(context.Background(), "ping")
	if err == nil || !strings.Contains(err.Error(), "auth") {
		t.Errorf("expected ErrAuth, got %v", err)
	}
}
