package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/latere-ai/debate/internal/critic"
)

func TestNewCriticPanicsOnUnknown(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	NewCritic("unknown")
}

func TestNewCriticReturnsCorrectImpl(t *testing.T) {
	if _, ok := NewCritic("codex").(*CodexCritic); !ok {
		t.Error("codex factory: wrong type")
	}
	if _, ok := NewCritic("claude").(*ClaudeCritic); !ok {
		t.Error("claude factory: wrong type")
	}
}

func TestAssemblePrompt(t *testing.T) {
	a := critic.Lookup("security")
	in := CriticInput{
		Aspect: a, CriticIndex: 1, Round: 1, SystemPrompt: "SYS",
		TaskContext: "TASK", DiffPatch: "DIFF",
		PriorRoundFiles: []RoundFileRef{{Path: "rounds/r2-proposer.md", Round: 2, Role: "proposer"}},
	}
	got := AssemblePrompt(in)
	if !strings.Contains(got, "SYS") || !strings.Contains(got, "# Task") || !strings.Contains(got, "DIFF") {
		t.Errorf("missing block; got: %.300s", got)
	}
	if !strings.Contains(got, "@rounds/r2-proposer.md") {
		t.Error("missing prior round reference")
	}
}

// Smoke test: when bin is missing, Exec returns an error and we map it.
func TestCodexMissingBinary(t *testing.T) {
	c := &CodexCritic{Bin: "/no/such/binary-z9z"}
	_, err := c.Round(context.Background(), CriticInput{
		Aspect: critic.Lookup("security"), CriticIndex: 1, Round: 1,
		SystemPrompt: "SYS", TaskContext: "T", DiffPatch: "D",
		Cwd: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
