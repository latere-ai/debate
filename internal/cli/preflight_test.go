package cli

import (
	"context"
	"errors"
	"testing"
)

func validFlags() *Flags {
	f := DefaultFlags()
	f.SessionID = "abc-123" // satisfy task-context check
	return f
}

func TestPreflightOK(t *testing.T) {
	t.Setenv("DEBATE_IN_PROGRESS", "")
	plan, err := Preflight(context.Background(), validFlags())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Forks) != 4 {
		t.Errorf("forks: got %d, want 4", len(plan.Forks))
	}
	if plan.Forks[0].Index != 1 || plan.Forks[3].Index != 4 {
		t.Errorf("fork indexing 1-based: got %v", plan.Forks)
	}
}

func TestPreflightRecursionGuard(t *testing.T) {
	t.Setenv("DEBATE_IN_PROGRESS", "1")
	_, err := Preflight(context.Background(), validFlags())
	if !errors.Is(err, ErrRecursionGuard) {
		t.Errorf("got %v, want ErrRecursionGuard", err)
	}
}

func TestPreflightCodexProposerRejected(t *testing.T) {
	f := validFlags()
	f.Main = "codex"
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 102 {
		t.Errorf("expected exit 102, got %v", err)
	}
}

func TestPreflightSameFamilyMissingModels(t *testing.T) {
	f := validFlags()
	f.Side = "claude"
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 110 {
		t.Errorf("expected exit 110, got %v", err)
	}
}

func TestPreflightSameFamilySameModels(t *testing.T) {
	f := validFlags()
	f.Side = "claude"
	f.MainModel = "claude-x"
	f.SideModel = "claude-x"
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 111 {
		t.Errorf("expected exit 111, got %v", err)
	}
}

func TestPreflightArityMismatch(t *testing.T) {
	f := validFlags()
	f.SideCount = 3
	f.Aspect = []string{"a", "b"}
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 120 {
		t.Errorf("expected exit 120, got %v", err)
	}
}

func TestPreflightNoTaskContext(t *testing.T) {
	f := DefaultFlags()
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 130 {
		t.Errorf("expected exit 130, got %v", err)
	}
}
