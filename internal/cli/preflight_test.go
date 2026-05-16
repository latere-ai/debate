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
	t.Setenv("AGON_IN_PROGRESS", "")
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
	t.Setenv("AGON_IN_PROGRESS", "1")
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

func TestPreflightSideCountInvalid(t *testing.T) {
	f := validFlags()
	f.SideCount = 0
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 121 {
		t.Errorf("expected exit 121, got %v", err)
	}
}

// TestPreflightMaxTurnZeroRejected pins the new lower bound: with
// --max-turn re-interpreted as critic↔proposer pairs, the minimum
// meaningful value is 1 (one full exchange). Zero is rejected with
// exit 122.
func TestPreflightMaxTurnZeroRejected(t *testing.T) {
	f := validFlags()
	f.MaxTurn = 0
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 122 {
		t.Errorf("expected exit 122, got %v", err)
	}
}

// TestPreflightMaxTurnOneAccepted: a single exchange is now the
// minimum, where pre-rename it would have been rejected ("must be
// >= 2" was rounds, not turns).
func TestPreflightMaxTurnOneAccepted(t *testing.T) {
	f := validFlags()
	f.MaxTurn = 1
	if _, err := Preflight(context.Background(), f); err != nil {
		t.Errorf("--max-turn 1 should be valid in pair semantics; got %v", err)
	}
}

// TestDefaultFlagsMaxTurn locks in the default change from 6 (rounds)
// to 3 (pairs). Same engine behaviour, less ambiguous unit.
func TestDefaultFlagsMaxTurn(t *testing.T) {
	if got := DefaultFlags().MaxTurn; got != 3 {
		t.Errorf("default --max-turn: got %d, want 3 (pairs)", got)
	}
}

// TestPreflightLogModeUnknownRejected asserts that an unknown
// --log-mode value is caught at preflight, not silently accepted
// (which would leave Engine.HeartbeatInterval at 0 with the user
// thinking they had asked for verbose).
func TestPreflightLogModeUnknownRejected(t *testing.T) {
	f := validFlags()
	f.LogMode = "thinking"
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 125 {
		t.Errorf("expected exit 125 for unknown log-mode, got %v", err)
	}
}

func TestPreflightLogModeAcceptsAll(t *testing.T) {
	for _, m := range ValidLogModes {
		f := validFlags()
		f.LogMode = m
		if _, err := Preflight(context.Background(), f); err != nil {
			t.Errorf("--log-mode %q should be valid; got %v", m, err)
		}
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
