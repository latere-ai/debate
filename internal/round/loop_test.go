package round

import (
	"context"
	"strings"
	"testing"
	"time"

	"latere.ai/x/debate/internal/agent"
	"latere.ai/x/debate/internal/ledger"
	"latere.ai/x/debate/internal/state"
)

type stubProposer struct {
	first, next func(string) (*agent.ProposerResult, error)
}

func (s *stubProposer) FirstRound(_ context.Context, pointer string) (*agent.ProposerResult, error) {
	return s.first(pointer)
}

func (s *stubProposer) NextRound(_ context.Context, _ string, pointer string) (*agent.ProposerResult, error) {
	return s.next(pointer)
}

type stubCritic struct {
	rounds []string
	idx    int
}

func (s *stubCritic) Round(_ context.Context, _ agent.CriticInput) (*agent.CriticResult, error) {
	if s.idx >= len(s.rounds) {
		return &agent.CriticResult{Markdown: "# Critic 1 — round 99 attacks\n\naspect: security\n", Duration: time.Millisecond}, nil
	}
	out := &agent.CriticResult{Markdown: s.rounds[s.idx], Duration: time.Millisecond}
	s.idx++
	return out, nil
}

func TestEngineSingleForkSteadyState(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r1 := "# Critic 1 — round 1 attacks\n\naspect: security\n\n## c1-1 [x.go:1]\n\nclaim: leaks token\n\nexpected violation: panic at runtime\n\nreproduction:\n```\ngo test\n```\n"
	r3 := "# Critic 1 — round 3 attacks\n\naspect: security\n" // empty: no new
	r5 := "# Critic 1 — round 5 attacks\n\naspect: security\n" // empty: steady state at R5

	e := &Engine{
		Sess: sess, Cwd: t.TempDir(),
		Aspects: []string{"security"},
		Proposer: &stubProposer{
			first: func(_ string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "fork-1", Response: "rebut c1-1: framework escapes", Tokens: 10}, nil
			},
			next: func(_ string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "fork-1", Response: "no further action", Tokens: 10}, nil
			},
		},
		NewCritic: func(_ int) agent.Critic { return &stubCritic{rounds: []string{r1, r3, r5}} },
		MaxTurn:   6, CostCap: 100000, TaskContext: "task", DiffPatch: "diff",
	}
	sum, err := e.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sum.Termination != TermSteadyState {
		t.Errorf("termination: got %s, want steady-state", sum.Termination)
	}
	agg, _ := ledger.Aggregate(sess)
	if r := agg["c1-1"]; r.Status != ledger.StatusRebutted && r.Status != ledger.StatusUnresolved {
		t.Errorf("attack status after rebut: got %s", r.Status)
	}
}

func TestEngineCostCap(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r1 := "# Critic 1 — round 1 attacks\n\naspect: security\n\n## c1-1 [x:1]\n\nclaim: leaks\n\nexpected violation: panic\n\nreproduction:\n```\nx\n```\n"
	e := &Engine{
		Sess: sess, Cwd: t.TempDir(),
		Aspects: []string{"security"},
		Proposer: &stubProposer{
			first: func(string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "f", Response: "rebut c1-1: ok", Tokens: 50000}, nil
			},
			next: func(string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "f", Response: "rebut c1-1: ok", Tokens: 50000}, nil
			},
		},
		NewCritic: func(_ int) agent.Critic {
			return &stubCritic{rounds: []string{r1, r1, r1, r1, r1}}
		},
		MaxTurn: 6, CostCap: 10000, TaskContext: "t", DiffPatch: "d",
	}
	sum, err := e.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sum.Termination != TermCostCap {
		t.Errorf("termination: got %s, want cost-cap", sum.Termination)
	}
}

func TestDefenseLineParsing(t *testing.T) {
	out := defenseLineRE.FindAllStringSubmatch("here is text\nconcede c1-1\nrebut c1-2\n", -1)
	if len(out) != 2 {
		t.Errorf("got %d matches, want 2", len(out))
	}
	if out[0][1] != "concede" || out[0][2] != "c1-1" {
		t.Errorf("first: %v", out[0])
	}
	if !strings.HasPrefix(out[1][2], "c1-") {
		t.Errorf("second: %v", out[1])
	}
}
