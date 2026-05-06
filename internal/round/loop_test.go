package round

import (
	"context"
	"encoding/json"
	"os"
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
		ForkCount: 1,
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
		ForkCount: 1,
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
		ForkCount: 1,
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

func TestEnginePerForkUsageStats(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r1 := "# Critic 1 - round 1 attacks\n\naspect: security\n\n## c1-1 [x.go:1]\n\nclaim: leaks token\n\nexpected violation: panic at runtime\n\nreproduction:\n```\ngo test\n```\n"
	r3 := "# Critic 1 - round 3 attacks\n\naspect: security\n"
	r5 := "# Critic 1 - round 5 attacks\n\naspect: security\n"
	cu := agent.TokenUsage{Input: 1000, Output: 200, CacheCreate: 800, CacheRead: 4000}
	pu := agent.TokenUsage{Input: 500, Output: 150, CacheCreate: 0, CacheRead: 3000}
	const criticUSD = 0.0125
	const proposerUSD = 0.0080
	sc := &usageCritic{
		rounds: []string{r1, r3, r5},
		usage:  cu,
		usd:    criticUSD,
	}
	e := &Engine{
		Sess: sess, Cwd: t.TempDir(),
		ForkCount: 1,
		Proposer: &stubProposer{
			first: func(string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "fork-1", Response: "rebut c1-1: ok", Tokens: 10, Usage: pu, USD: proposerUSD}, nil
			},
			next: func(string) (*agent.ProposerResult, error) {
				return &agent.ProposerResult{ForkID: "fork-1", Response: "no further action", Tokens: 10, Usage: pu, USD: proposerUSD}, nil
			},
		},
		NewCritic: func(_ int) agent.Critic { return sc },
		MaxTurn:   6, CostCap: 1_000_000, TaskContext: "task", DiffPatch: "diff",
	}
	sum, err := e.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sum.Forks) != 1 {
		t.Fatalf("want 1 fork, got %d", len(sum.Forks))
	}
	got := sum.Forks[0].Usage
	criticRounds := sc.idx
	wantCritic := agent.TokenUsage{
		Input:       cu.Input * criticRounds,
		Output:      cu.Output * criticRounds,
		CacheCreate: cu.CacheCreate * criticRounds,
		CacheRead:   cu.CacheRead * criticRounds,
	}
	if got.Critic != wantCritic {
		t.Errorf("critic usage: got %+v, want %+v", got.Critic, wantCritic)
	}
	if got.Total.Total() != got.Critic.Total()+got.Proposer.Total() {
		t.Errorf("total mismatch: got %+v", got.Total)
	}
	if sum.Usage != got.Total {
		t.Errorf("run-level usage: got %+v, want %+v", sum.Usage, got.Total)
	}
	proposerRounds := sum.Forks[0].Rounds - criticRounds
	wantCriticUSD := criticUSD * float64(criticRounds)
	wantProposerUSD := proposerUSD * float64(proposerRounds)
	wantTotalUSD := wantCriticUSD + wantProposerUSD
	if !floatEq(got.CriticUSD, wantCriticUSD) {
		t.Errorf("critic USD: got %v, want %v", got.CriticUSD, wantCriticUSD)
	}
	if !floatEq(got.ProposerUSD, wantProposerUSD) {
		t.Errorf("proposer USD: got %v, want %v", got.ProposerUSD, wantProposerUSD)
	}
	if !floatEq(got.TotalUSD, wantTotalUSD) {
		t.Errorf("total USD: got %v, want %v", got.TotalUSD, wantTotalUSD)
	}
	if !floatEq(sum.USD, wantTotalUSD) {
		t.Errorf("run-level USD: got %v, want %v", sum.USD, wantTotalUSD)
	}
	for _, r := range got.Rounds {
		want := criticUSD
		if r.Role == "proposer" {
			want = proposerUSD
		}
		if !floatEq(r.USD, want) {
			t.Errorf("round %d (%s) USD: got %v, want %v", r.Round, r.Role, r.USD, want)
		}
	}
	statsBytes, err := os.ReadFile(filepath.Join(sess.Root, "forks", "critic-1", "stats.json"))
	if err != nil {
		t.Fatal(err)
	}
	var on map[string]any
	if err := json.Unmarshal(statsBytes, &on); err != nil {
		t.Fatalf("stats.json invalid JSON: %v", err)
	}
	if on["schema"] != "debate.fork-stats.v0" {
		t.Errorf("stats.json schema: %v", on["schema"])
	}
	if on["topic"] != "security" {
		t.Errorf("stats.json topic: %v", on["topic"])
	}
	usageJSON, ok := on["usage"].(map[string]any)
	if !ok {
		t.Fatalf("stats.json usage block missing or wrong shape: %v", on["usage"])
	}
	for _, k := range []string{"critic_usd", "proposer_usd", "total_usd"} {
		if _, ok := usageJSON[k]; !ok {
			t.Errorf("stats.json usage missing %q", k)
		}
	}
	if v, _ := usageJSON["total_usd"].(float64); !floatEq(v, wantTotalUSD) {
		t.Errorf("stats.json total_usd: got %v, want %v", v, wantTotalUSD)
	}
}

func floatEq(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-9
}

type usageCritic struct {
	rounds []string
	idx    int
	usage  agent.TokenUsage
	usd    float64
}

func (s *usageCritic) Round(_ context.Context, _ agent.CriticInput) (*agent.CriticResult, error) {
	if s.idx >= len(s.rounds) {
		return &agent.CriticResult{
			Markdown: "# Critic 1 - round 99 attacks\n\naspect: security\n",
			Duration: time.Millisecond, Usage: s.usage, USD: s.usd, Tokens: s.usage.Input + s.usage.Output,
		}, nil
	}
	out := &agent.CriticResult{
		Markdown: s.rounds[s.idx], Duration: time.Millisecond,
		Usage: s.usage, USD: s.usd, Tokens: s.usage.Input + s.usage.Output,
	}
	s.idx++
	return out, nil
}

// TestEnginePromptCachingPerFork verifies that each fork's recorded
// usage actually shows prompt caching at work: R1 of every fork creates
// cache (CacheCreate>0, CacheRead=0) and later rounds read from cache
// (CacheRead>0). The orchestrator does not control the cache (the agent
// CLI does), so the assertion here is that the fork-level accounting
// preserves the per-call cache_creation_input_tokens /
// cache_read_input_tokens that the agents reported - and does so
// independently for every fork.
func TestEnginePromptCachingPerFork(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 2, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r1 := "# Critic 1 - round 1 attacks\n\naspect: security\n\n## c1-1 [x.go:1]\n\nclaim: leaks token\n\nexpected violation: panic at runtime\n\nreproduction:\n```\ngo test\n```\n"
	r3 := "# Critic 1 - round 3 attacks\n\naspect: security\n"
	r5 := "# Critic 1 - round 5 attacks\n\naspect: security\n"

	// Per-fork sequence the stubs replay. R1 critic primes the cache;
	// R3/R5 critic and R4 proposer hit it. R2 proposer (FirstRound)
	// primes the proposer-side cache.
	criticUsages := []agent.TokenUsage{
		{Input: 200, Output: 100, CacheCreate: 5000, CacheRead: 0}, // R1
		{Input: 50, Output: 80, CacheCreate: 100, CacheRead: 5000}, // R3
		{Input: 30, Output: 50, CacheCreate: 0, CacheRead: 5100},   // R5
	}
	proposerUsages := []agent.TokenUsage{
		{Input: 100, Output: 200, CacheCreate: 4000, CacheRead: 0},  // R2
		{Input: 50, Output: 100, CacheCreate: 200, CacheRead: 4000}, // R4
	}

	prop := &cachingProposer{usages: proposerUsages}
	e := &Engine{
		Sess: sess, Cwd: t.TempDir(),
		ForkCount: 2,
		Proposer:  prop,
		NewCritic: func(_ int) agent.Critic {
			return &cachingCritic{rounds: []string{r1, r3, r5}, usages: criticUsages}
		},
		MaxTurn: 6, CostCap: 1_000_000, TaskContext: "task", DiffPatch: "diff",
	}

	sum, err := e.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sum.Forks) != 2 {
		t.Fatalf("forks: got %d, want 2", len(sum.Forks))
	}

	for _, f := range sum.Forks {
		fu := f.Usage
		// Critic side: at least one round read from cache.
		if fu.Critic.CacheRead == 0 {
			t.Errorf("fork %d: critic CacheRead = 0; prompt caching not observed on critic side", f.Index)
		}
		if fu.Critic.CacheCreate == 0 {
			t.Errorf("fork %d: critic CacheCreate = 0; cache never primed", f.Index)
		}
		// Proposer side: --resume reuses the conversation, so R4 must
		// see cache reads even though R2 only created cache.
		if fu.Proposer.CacheRead == 0 {
			t.Errorf("fork %d: proposer CacheRead = 0; --resume cache not seen", f.Index)
		}
		if fu.Proposer.CacheCreate == 0 {
			t.Errorf("fork %d: proposer CacheCreate = 0; --resume never primed cache", f.Index)
		}
		// A working cache should be paying for itself: cache reads
		// should outweigh fresh inputs across the fork.
		if fu.Total.CacheRead <= fu.Total.Input {
			t.Errorf("fork %d: cache_read=%d <= input=%d; cache not amortising input cost",
				f.Index, fu.Total.CacheRead, fu.Total.Input)
		}
		// Per-round assertions: R1 must NOT yet show cache reads (the
		// cache it creates is what later rounds consume). At least one
		// critic and one proposer round after R1 must show CacheRead>0.
		var sawR1 bool
		var sawCriticHit, sawProposerHit bool
		for _, r := range fu.Rounds {
			if r.Round == 1 {
				sawR1 = true
				if r.Usage.CacheRead != 0 {
					t.Errorf("fork %d R1: CacheRead=%d, want 0 (R1 should only create cache)",
						f.Index, r.Usage.CacheRead)
				}
				if r.Usage.CacheCreate == 0 {
					t.Errorf("fork %d R1: CacheCreate=0, want >0 (R1 should prime cache)", f.Index)
				}
				continue
			}
			if r.Usage.CacheRead == 0 {
				continue
			}
			switch r.Role {
			case "critic":
				sawCriticHit = true
			case "proposer":
				sawProposerHit = true
			}
		}
		if !sawR1 {
			t.Errorf("fork %d: missing R1 in per-round breakdown", f.Index)
		}
		if !sawCriticHit {
			t.Errorf("fork %d: no critic round after R1 shows CacheRead>0", f.Index)
		}
		if !sawProposerHit {
			t.Errorf("fork %d: no proposer round after R1 shows CacheRead>0", f.Index)
		}
	}

	// Cross-fork independence: forks must not share state, so each fork
	// reports its own cache primer in R1 (one CacheCreate event per
	// fork on the critic side, not one for the whole run).
	if sum.Forks[0].Usage.Critic.CacheCreate != sum.Forks[1].Usage.Critic.CacheCreate {
		t.Errorf("per-fork critic CacheCreate diverged: f1=%d f2=%d (forks should be independent)",
			sum.Forks[0].Usage.Critic.CacheCreate, sum.Forks[1].Usage.Critic.CacheCreate)
	}
}

// cachingCritic replays a per-round sequence of TokenUsage values so a
// test can simulate the cache-create / cache-read transition that a
// real claude critic produces across rounds.
type cachingCritic struct {
	rounds []string
	usages []agent.TokenUsage
	idx    int
}

func (s *cachingCritic) Round(_ context.Context, _ agent.CriticInput) (*agent.CriticResult, error) {
	var u agent.TokenUsage
	if s.idx < len(s.usages) {
		u = s.usages[s.idx]
	}
	md := "# Critic 1 - round 99 attacks\n\naspect: security\n"
	if s.idx < len(s.rounds) {
		md = s.rounds[s.idx]
	}
	s.idx++
	return &agent.CriticResult{
		Markdown: md, Duration: time.Millisecond,
		Usage: u, Tokens: u.Input + u.Output,
	}, nil
}

// cachingProposer replays per-fork proposer usage. FirstRound resets
// the per-fork counter so each fork sees the same R2/R4 sequence.
type cachingProposer struct {
	usages []agent.TokenUsage
	idx    int
}

func (s *cachingProposer) FirstRound(_ context.Context, _ string) (*agent.ProposerResult, error) {
	s.idx = 0
	return s.next()
}

func (s *cachingProposer) NextRound(_ context.Context, _ string, _ string) (*agent.ProposerResult, error) {
	return s.next()
}

func (s *cachingProposer) next() (*agent.ProposerResult, error) {
	var u agent.TokenUsage
	if s.idx < len(s.usages) {
		u = s.usages[s.idx]
	}
	s.idx++
	return &agent.ProposerResult{
		ForkID: "fork", Response: "rebut c1-1: ok",
		Tokens: u.Input + u.Output, Usage: u,
	}, nil
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
