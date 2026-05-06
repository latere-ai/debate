package cli

import (
	"context"
	"errors"
	"testing"
)

func TestPreflightCostCapMin(t *testing.T) {
	f := DefaultFlags()
	f.SessionID = "x"
	f.CostCap = 0
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 123 {
		t.Errorf("expected exit 123, got %v", err)
	}
}

func TestPreflightChangedLinesMinNegative(t *testing.T) {
	f := DefaultFlags()
	f.SessionID = "x"
	f.ChangedLinesMin = -1
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 124 {
		t.Errorf("expected exit 124, got %v", err)
	}
}

func TestPreflightSideCountMin(t *testing.T) {
	f := DefaultFlags()
	f.SessionID = "x"
	f.SideCount = 0
	_, err := Preflight(context.Background(), f)
	var pe *PreflightError
	if !errors.As(err, &pe) || pe.Code != 121 {
		t.Errorf("expected exit 121, got %v", err)
	}
}

func TestPreflightCrossFamilyOK(t *testing.T) {
	f := DefaultFlags()
	f.SessionID = "x"
	// Default already cross-family (claude vs codex).
	if _, err := Preflight(context.Background(), f); err != nil {
		t.Errorf("cross-family should pass: %v", err)
	}
}

func TestPreflightSameFamilyDifferentModelsOK(t *testing.T) {
	f := DefaultFlags()
	f.SessionID = "x"
	f.Side = "claude"
	f.MainModel = "claude-sonnet"
	f.SideModel = "claude-opus"
	if _, err := Preflight(context.Background(), f); err != nil {
		t.Errorf("same-family different-model should pass: %v", err)
	}
}

func TestPreflightErrorWraps(t *testing.T) {
	pe := &PreflightError{Code: 42, Msg: "boom", Wrap: errors.New("inner")}
	if got := pe.Error(); got != "boom: inner" {
		t.Errorf("got %q", got)
	}
	if pe.Unwrap() == nil {
		t.Error("Unwrap should return inner")
	}
}

func TestPreflightErrorNoWrap(t *testing.T) {
	pe := &PreflightError{Code: 1, Msg: "boom"}
	if got := pe.Error(); got != "boom" {
		t.Errorf("Error() without Wrap: got %q, want \"boom\"", got)
	}
	if pe.Unwrap() != nil {
		t.Error("Unwrap should be nil when Wrap unset")
	}
}

func TestDecodeCwdFromTranscript(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no-projects-segment", "/some/random/file.jsonl", ""},
		{"projects-segment-decoded", "/Users/x/.claude/projects/-tmp-abc/sess.jsonl", "/tmp/abc"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := decodeCwdFromTranscript(c.in)
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestAgentFamily(t *testing.T) {
	cases := map[string]string{
		"claude":  "claude",
		"codex":   "codex",
		"unknown": "unknown", // unknown family passes through
		"foo":     "foo",
	}
	for in, want := range cases {
		if got := agentFamily(in); got != want {
			t.Errorf("agentFamily(%q) = %q, want %q", in, got, want)
		}
	}
}
