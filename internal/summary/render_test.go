package summary

import (
	"strings"
	"testing"

	"latere.ai/x/debate/internal/ledger"
	"latere.ai/x/debate/internal/round"
)

func TestDecideClean(t *testing.T) {
	d := Decide(&round.Summary{Termination: round.TermSteadyState, Unresolved: 0})
	if d.Surface || d.ExitCode != 0 {
		t.Errorf("clean: %+v", d)
	}
}

func TestDecideUnresolved(t *testing.T) {
	d := Decide(&round.Summary{Termination: round.TermSteadyState, Unresolved: 2})
	if !d.Surface || d.ExitCode != 1 {
		t.Errorf("unresolved: %+v", d)
	}
}

func TestDecideInterrupted(t *testing.T) {
	d := Decide(&round.Summary{Termination: round.TermInterrupted})
	if !d.Surface || d.ExitCode != 130 {
		t.Errorf("interrupted: %+v", d)
	}
}

func TestRenderHasHeadlineAndStats(t *testing.T) {
	r := &Render{Format: "markdown"}
	agg := map[string]ledger.Record{
		"c1-1": {
			AttackID: "c1-1", Aspect: "security", Location: "x.go:1",
			Claim: "leak", ExpectedViolation: "panic", Reproduction: "go run", Status: ledger.StatusUnresolved,
			RoundIntroduced: ptr(1), RoundLastTouched: 3, ReAttacked: true,
		},
		"c1-2": {
			AttackID: "c1-2", Aspect: "security", Status: ledger.StatusConceded,
			Claim: "off by one", ConcessionFiles: []string{"x.go"},
		},
	}
	b, _ := r.Bytes(&round.Summary{Termination: round.TermSteadyState, Unresolved: 1}, agg)
	s := string(b)
	if !strings.Contains(s, "Headline") {
		t.Error("missing Headline section")
	}
	if !strings.Contains(s, "## Stats") {
		t.Error("missing Stats section")
	}
	if !strings.Contains(s, "critic-found-bug rate: 1/2") {
		t.Errorf("stat line incorrect; body:\n%s", s)
	}
}
