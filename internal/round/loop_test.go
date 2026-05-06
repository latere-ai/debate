package round

import (
	"context"
	"path/filepath"
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
	inputs []agent.CriticInput
}

func (s *stubCritic) Round(_ context.Context, in agent.CriticInput) (*agent.CriticResult, error) {
	s.inputs = append(s.inputs, in)
	if s.idx >= len(s.rounds) {
		return &agent.CriticResult{Markdown: "# Critic 1 - round 99 attacks\n\naspect: security\n", Duration: time.Millisecond}, nil
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
	r1 := "# Critic 1 - round 1 attacks\n\naspect: security\n\n## c1-1 [x.go:1]\n\nclaim: leaks token\n\nexpected violation: panic at runtime\n\nreproduction:\n```\ngo test\n```\n"
	r3 := "# Critic 1 - round 3 attacks\n\naspect: security\n" // empty: no new
	r5 := "# Critic 1 - round 5 attacks\n\naspect: security\n" // empty: steady state at R5

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
	r1 := "# Critic 1 - round 1 attacks\n\naspect: security\n\n## c1-1 [x:1]\n\nclaim: leaks\n\nexpected violation: panic\n\nreproduction:\n```\nx\n```\n"
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

// TestCriticRoundPriorFiles ensures R3 onward receives pointers to the
// previous critic and proposer round files. Without these the critic
// agent has no way to react to the proposer's defense and tends to emit
// an empty document (the spec-mandated "nothing to attack" shape).
func TestCriticRoundPriorFiles(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r1 := "# Critic 1 - round 1 attacks\n\naspect: security\n\n## c1-1 [x.go:1]\n\nclaim: leaks token\n\nexpected violation: panic at runtime\n\nreproduction:\n```\ngo test\n```\n"
	r3 := "# Critic 1 - round 3 attacks\n\naspect: security\n\n## c1-1 [x.go:1] (re-attack)\n\nclaim: still leaks\n\nexpected violation: panic\n\nreproduction:\n```\ngo test\n```\n"
	r5 := "# Critic 1 - round 5 attacks\n\naspect: security\n"

	cri := &stubCritic{rounds: []string{r1, r3, r5}}
	e := &Engine{
		Sess: sess, Cwd: t.TempDir(),
		Aspects: []string{"security"},
		Proposer: &stubProposer{
			first: func(string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "fork-1", Response: "rebut c1-1: ok", Tokens: 10}, nil
			},
			next: func(string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "fork-1", Response: "rebut c1-1: ok", Tokens: 10}, nil
			},
		},
		NewCritic: func(_ int) agent.Critic { return cri },
		MaxTurn:   6, CostCap: 100000, TaskContext: "task", DiffPatch: "diff",
	}
	if _, err := e.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(cri.inputs) < 2 {
		t.Fatalf("got %d critic inputs, want >= 2", len(cri.inputs))
	}
	r1in := cri.inputs[0]
	if len(r1in.PriorRoundFiles) != 0 {
		t.Errorf("R1 PriorRoundFiles: got %d, want 0", len(r1in.PriorRoundFiles))
	}
	r3in := cri.inputs[1]
	if len(r3in.PriorRoundFiles) != 2 {
		t.Fatalf("R3 PriorRoundFiles: got %d, want 2", len(r3in.PriorRoundFiles))
	}
	wantCritic := sess.Path(filepath.Join("forks", "critic-1", "rounds", "r1-critic.md"))
	wantProposer := sess.Path(filepath.Join("forks", "critic-1", "rounds", "r2-proposer.md"))
	if r3in.PriorRoundFiles[0].Path != wantCritic || r3in.PriorRoundFiles[0].Role != "critic" || r3in.PriorRoundFiles[0].Round != 1 {
		t.Errorf("R3 PriorRoundFiles[0]: got %+v, want path=%s round=1 role=critic", r3in.PriorRoundFiles[0], wantCritic)
	}
	if r3in.PriorRoundFiles[1].Path != wantProposer || r3in.PriorRoundFiles[1].Role != "proposer" || r3in.PriorRoundFiles[1].Round != 2 {
		t.Errorf("R3 PriorRoundFiles[1]: got %+v, want path=%s round=2 role=proposer", r3in.PriorRoundFiles[1], wantProposer)
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
