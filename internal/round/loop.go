package round

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"latere.ai/x/debate/internal/agent"
	"latere.ai/x/debate/internal/critic"
	"latere.ai/x/debate/internal/ledger"
	"latere.ai/x/debate/internal/state"
)

// Proposer is the orchestrator's view of the proposer driver.
type Proposer interface {
	FirstRound(ctx context.Context, pointer string) (*agent.ProposerResult, error)
	NextRound(ctx context.Context, forkID, pointer string) (*agent.ProposerResult, error)
}

// CriticFactory produces a Critic for the given fork index.
type CriticFactory func(forkIdx int) agent.Critic

// Engine is the orchestration core; it integrates [09]-[18] and emits
// the *Summary that [22]/[23] consume.
type Engine struct {
	Sess        *state.Session
	Cwd         string
	ForkCount   int
	Proposer    Proposer
	NewCritic   CriticFactory
	MaxTurn     int
	CostCap     int
	HookMode    bool
	TaskContext string
	DiffPatch   string
	// Progress is the writer used for per-fork and per-round progress
	// lines. nil means silent. cmd/debate sets this to os.Stderr in
	// non-hook mode. The Stop-hook path leaves it nil since claude
	// swallows the stderr anyway.
	Progress io.Writer
}

func (e *Engine) progf(format string, args ...any) {
	if e.Progress == nil {
		return
	}
	_, _ = fmt.Fprintf(e.Progress, format+"\n", args...)
}

// Summary is what Run returns on a successful end-to-end run.
type Summary struct {
	Sess        *state.Session
	Termination TerminationReason
	Forks       []ForkOutcome
	TokensUsed  int
	Usage       agent.TokenUsage
	// USD is the run-level total cost reported by the agent CLIs,
	// summed across every fork and round. Zero when the agents do
	// not surface a total_cost_usd field (e.g. codex critic, mocks).
	USD         float64
	WallSeconds int
	Headline    *ledger.Record
	Unresolved  int
}

// ForkOutcome carries the per-fork termination + last round. Topic is
// the lens the critic declared in R1 (the "aspect:" line of its
// markdown output) and is empty if R1 never completed.
type ForkOutcome struct {
	Index       int
	Topic       string
	Rounds      int
	Termination TerminationReason
	Usage       ForkUsage
}

// ForkUsage aggregates one fork's token consumption split by role.
// Critic and Proposer both contain TokenUsage broken down into input,
// output, cache_creation, cache_read; Total is the convenience sum.
// CriticUSD, ProposerUSD and TotalUSD mirror the same split for the
// total_cost_usd field reported by the agent CLIs (zero when the agent
// does not surface it). Stored to <session>/forks/critic-N/stats.json
// and surfaced on the completion progress line.
type ForkUsage struct {
	Critic      agent.TokenUsage `json:"critic"`
	Proposer    agent.TokenUsage `json:"proposer"`
	Total       agent.TokenUsage `json:"total"`
	CriticUSD   float64          `json:"critic_usd"`
	ProposerUSD float64          `json:"proposer_usd"`
	TotalUSD    float64          `json:"total_usd"`
	// Rounds is the per-round breakdown in execution order, useful for
	// spotting the round where the cache went cold.
	Rounds []Usage `json:"rounds"`
}

// Usage records one critic-or-proposer round's token consumption.
// Sits in ForkUsage.Rounds; package-qualified name is round.Usage.
// USD is the agent-reported total_cost_usd for that round (zero when
// the agent does not surface it).
type Usage struct {
	Round int              `json:"round"`
	Role  string           `json:"role"`
	Usage agent.TokenUsage `json:"usage"`
	USD   float64          `json:"usd"`
	MS    int              `json:"ms"`
}

// Typed errors ([20]).
var (
	ErrInterrupted    = errors.New("debate interrupted")
	ErrCostCap        = errors.New("debate cost cap reached")
	ErrMalformedTwice = errors.New("debate malformed output twice")
	ErrAgentFatal     = errors.New("debate agent fatal error")
)

var defenseLineRE = regexp.MustCompile(`(?m)^\s*(concede|rebut|push-back)\s+(c\d+-\d+)\b`)

// Run executes the orchestrator. Forks run serially.
func (e *Engine) Run(ctx context.Context) (*Summary, error) {
	det := &Detector{MaxTurn: e.MaxTurn, CostCap: e.CostCap}
	cost := NewCostMeter(e.CostCap)
	start := time.Now()
	sum := &Summary{Sess: e.Sess, Termination: TermSteadyState}

	for forkIdx := 1; forkIdx <= e.ForkCount; forkIdx++ {
		if ctx.Err() != nil {
			sum.Termination = TermInterrupted
			break
		}
		priorTopics := claimedTopics(sum.Forks)
		outcome, runStop, err := e.runFork(ctx, forkIdx, priorTopics, det, cost)
		sum.Forks = append(sum.Forks, outcome)
		if err != nil {
			return nil, err
		}
		if runStop != "" {
			sum.Termination = runStop
			break
		}
	}

	// Finalize unresolved.
	agg, err := ledger.Aggregate(e.Sess)
	if err != nil {
		return nil, err
	}
	for id, r := range agg {
		if r.Status == ledger.StatusOpen || r.Status == ledger.StatusRebutted {
			r.Status = ledger.StatusUnresolved
			_ = ledger.Append(e.Sess, r)
			agg[id] = r
		}
	}
	for _, r := range agg {
		if r.Status == ledger.StatusUnresolved {
			sum.Unresolved++
		}
	}
	sum.TokensUsed = cost.Used()
	for _, f := range sum.Forks {
		sum.Usage.Add(f.Usage.Total)
		sum.USD += f.Usage.TotalUSD
	}
	sum.WallSeconds = int(time.Since(start).Seconds())
	return sum, nil
}

func (e *Engine) runFork(ctx context.Context, forkIdx int, priorTopics []string, det *Detector, cost *CostMeter) (ForkOutcome, TerminationReason, error) {
	out := ForkOutcome{Index: forkIdx, Termination: TermSteadyState}
	cri := e.NewCritic(forkIdx)
	// R1 starts in auto mode; the critic declares its topic in the
	// reply, the orchestrator captures it after R1 and locks subsequent
	// rounds to that topic.
	a := critic.Auto(forkIdx, e.ForkCount, priorTopics)
	var (
		forkID   string
		hist     []ForkHistory
		runStop  TerminationReason
		priorIDs []string
	)

	e.progf("[debate] fork %d/%d: starting (topic to be declared in R1)", forkIdx, e.ForkCount)

	for round := 1; round <= e.MaxTurn; round++ {
		if ctx.Err() != nil {
			runStop = TermInterrupted
			break
		}
		if cost.ExceedsCap() {
			runStop = TermCostCap
			break
		}
		out.Rounds = round
		roundStart := time.Now()
		if round%2 == 1 {
			// Critic round.
			e.progf("[debate] fork %d/%d %s: R%d critic running...", forkIdx, e.ForkCount, forkLabel(out.Topic), round)
			res, stats, err := e.criticRound(ctx, cri, a, forkIdx, round, priorIDs)
			if err != nil {
				return out, "", fmt.Errorf("%w: critic %d round %d: %v", ErrAgentFatal, forkIdx, round, err)
			}
			// Capture the declared topic the first time we see one and
			// lock subsequent rounds to it. Without this R3+ would still
			// run under the auto skeleton and the critic could drift.
			if out.Topic == "" && res.declaredTopic != "" {
				out.Topic = res.declaredTopic
				a = critic.Locked(forkIdx, e.ForkCount, out.Topic)
			}
			cost.Add(res.tokens)
			out.Usage.Critic.Add(res.usage)
			out.Usage.Total.Add(res.usage)
			out.Usage.CriticUSD += res.usd
			out.Usage.TotalUSD += res.usd
			out.Usage.Rounds = append(out.Usage.Rounds, Usage{
				Round: round, Role: "critic", Usage: res.usage, USD: res.usd,
				MS: int(time.Since(roundStart).Milliseconds()),
			})
			e.progf("[debate] fork %d/%d %s: R%d critic done in %s (new=%d, re-attack=%d, withdraw=%d, dropped=%d) %s",
				forkIdx, e.ForkCount, forkLabel(out.Topic), round, fmtDur(time.Since(roundStart)),
				stats.KeptIntroduce, stats.KeptReAttack, stats.KeptWithdraw,
				stats.DroppedNoReproduce+stats.DroppedStyle+stats.DroppedCrossAspect,
				fmtUsage(res.usage, res.usd))
			hist = append(hist, ForkHistory{
				Round: round, NewAttacks: stats.KeptIntroduce, ReAttacks: stats.KeptReAttack,
				Withdrawn:     stats.KeptWithdraw,
				MalformedFlag: stats.Total > 0 && (stats.KeptIntroduce+stats.KeptReAttack+stats.KeptWithdraw) == 0,
			})
			priorIDs = res.priorIDs
			if det.MalformedTwice(hist) {
				runStop = TermMalformedOutput
				break
			}
			if det.SteadyState(hist) {
				out.Termination = TermSteadyState
				e.progf("[debate] fork %d/%d %s: steady state reached at R%d", forkIdx, e.ForkCount, forkLabel(out.Topic), round)
				break
			}
		} else {
			// Proposer round.
			e.progf("[debate] fork %d/%d %s: R%d proposer running...", forkIdx, e.ForkCount, forkLabel(out.Topic), round)
			pointer := fmt.Sprintf("Some comments at @forks/critic-%d/rounds/r%d-critic.md. Please resolve or respond. If you disagree, please raise it.",
				forkIdx, round-1)
			var pr *agent.ProposerResult
			var err error
			if forkID == "" {
				pr, err = e.Proposer.FirstRound(ctx, pointer)
				if err == nil {
					forkID = pr.ForkID
					_ = state.WriteProposerState(e.Sess, forkIdx, &state.ProposerState{
						Agent: "claude", ForkSessionID: forkID,
					})
				}
			} else {
				pr, err = e.Proposer.NextRound(ctx, forkID, pointer)
			}
			if err != nil {
				return out, "", fmt.Errorf("%w: proposer fork %d round %d: %v", ErrAgentFatal, forkIdx, round, err)
			}
			cost.Add(pr.Tokens)
			out.Usage.Proposer.Add(pr.Usage)
			out.Usage.Total.Add(pr.Usage)
			out.Usage.ProposerUSD += pr.USD
			out.Usage.TotalUSD += pr.USD
			out.Usage.Rounds = append(out.Usage.Rounds, Usage{
				Round: round, Role: "proposer", Usage: pr.Usage, USD: pr.USD,
				MS: int(pr.Duration.Milliseconds()),
			})
			body := pr.Response + "\n\n---\nmodified-files:\n"
			for _, f := range pr.ChangedFiles {
				body += "  - " + f + "\n"
			}
			if err := state.WriteRound(e.Sess, forkIdx, round, state.RoleProposer, []byte(body)); err != nil {
				return out, "", err
			}
			_ = state.AppendTranscript(e.Sess, &state.TranscriptRecord{
				TS: time.Now().UTC(), Fork: forkIdx, Round: round, Role: "proposer",
				Path: filepath.Join("forks", fmt.Sprintf("critic-%d", forkIdx), "rounds", fmt.Sprintf("r%d-proposer.md", round)),
				MS:   int(pr.Duration.Milliseconds()),
			})
			conceded, rebutted := updateLedgerFromDefense(e.Sess, pr.Response, pr.ChangedFiles, round)
			e.progf("[debate] fork %d/%d %s: R%d proposer done in %s (conceded=%d, rebutted=%d, files=%d) %s",
				forkIdx, e.ForkCount, forkLabel(out.Topic), round, fmtDur(time.Since(roundStart)),
				conceded, rebutted, len(pr.ChangedFiles), fmtUsage(pr.Usage, pr.USD))
		}
	}
	if out.Rounds >= e.MaxTurn && runStop == "" {
		out.Termination = TermMaxTurn
	}
	if err := state.WriteForkStats(e.Sess, forkIdx, forkStatsFile{
		Schema:      schemaForkStats,
		ForkIndex:   forkIdx,
		Topic:       out.Topic,
		Rounds:      out.Rounds,
		Termination: ifEmpty(string(runStop), string(out.Termination)),
		Usage:       out.Usage,
	}); err != nil {
		return out, "", err
	}
	u := out.Usage.Total
	e.progf("[debate] fork %d/%d %s: terminated %s after R%d (in=%d out=%d cache_create=%d cache_read=%d total=%d cost=$%.4f)",
		forkIdx, e.ForkCount, forkLabel(out.Topic), ifEmpty(string(runStop), string(out.Termination)),
		out.Rounds, u.Input, u.Output, u.CacheCreate, u.CacheRead, u.Total(), out.Usage.TotalUSD)
	return out, runStop, nil
}

const schemaForkStats = "debate.fork-stats.v0"

type forkStatsFile struct {
	Schema      string    `json:"schema"`
	ForkIndex   int       `json:"fork_index"`
	Topic       string    `json:"topic"`
	Rounds      int       `json:"rounds"`
	Termination string    `json:"termination"`
	Usage       ForkUsage `json:"usage"`
}

// claimedTopics collects every non-empty topic critics in already-run
// forks declared, so the next fork's auto prompt can ask for a
// distinct angle.
func claimedTopics(forks []ForkOutcome) []string {
	out := make([]string, 0, len(forks))
	for _, f := range forks {
		if f.Topic != "" {
			out = append(out, f.Topic)
		}
	}
	return out
}

// forkLabel renders the topic in progress lines, falling back to
// "(topic pending)" before the critic has declared.
func forkLabel(topic string) string {
	if topic == "" {
		return "(topic pending)"
	}
	return topic
}

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// fmtUsage renders a per-round token + cost suffix appended to the
// "done in Xs" progress line. Cost is omitted when zero (codex critic
// and the e2e mock don't surface it). Format is intentionally compact
// so a 2-fork run still fits one line per round.
func fmtUsage(u agent.TokenUsage, usd float64) string {
	if usd > 0 {
		return fmt.Sprintf("[in=%d out=%d cache_create=%d cache_read=%d cost=$%.4f]",
			u.Input, u.Output, u.CacheCreate, u.CacheRead, usd)
	}
	return fmt.Sprintf("[in=%d out=%d cache_create=%d cache_read=%d]",
		u.Input, u.Output, u.CacheCreate, u.CacheRead)
}

func ifEmpty(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

type criticRoundResult struct {
	tokens        int
	usage         agent.TokenUsage
	usd           float64
	priorIDs      []string
	declaredTopic string
}

func (e *Engine) criticRound(ctx context.Context, cri agent.Critic, a critic.Aspect, forkIdx, round int, priorIDs []string) (criticRoundResult, critic.ParseStats, error) {
	in := agent.CriticInput{
		Aspect: a, CriticIndex: forkIdx, Round: round,
		SystemPrompt: critic.Assemble(a, forkIdx, round, ""),
		TaskContext:  e.TaskContext, DiffPatch: e.DiffPatch,
		Cwd: e.Cwd, Deadline: 5 * time.Minute,
	}
	// R3 and later: hand the critic absolute paths to the previous critic
	// round (r-2) and the proposer's defense (r-1). Without these the
	// system prompt promises "the proposer's prior responses (referenced
	// by file)" but no file is referenced, so the agent reads the same
	// diff/task as R1 and follows the "nothing new -> emit empty doc"
	// directive. Spec 19 §"R3..R(max)".
	if round >= 3 {
		priorCritic := round - 2
		priorProposer := round - 1
		forkRel := filepath.Join("forks", fmt.Sprintf("critic-%d", forkIdx), "rounds")
		in.PriorRoundFiles = []agent.RoundFileRef{
			{
				Path:  e.Sess.Path(filepath.Join(forkRel, fmt.Sprintf("r%d-critic.md", priorCritic))),
				Round: priorCritic, Role: "critic",
			},
			{
				Path:  e.Sess.Path(filepath.Join(forkRel, fmt.Sprintf("r%d-proposer.md", priorProposer))),
				Round: priorProposer, Role: "proposer",
			},
		}
	}
	res, err := cri.Round(ctx, in)
	if err != nil {
		return criticRoundResult{}, critic.ParseStats{}, err
	}
	declared := critic.ExtractDeclaredAspect(res.Markdown)
	expected := a.Name
	if expected == "" || expected == "auto" {
		expected = declared
	}
	attacks, stats, err := critic.Parse(res.Markdown, expected, forkIdx, round, priorIDs, critic.ParseOption{})
	if err != nil {
		return criticRoundResult{tokens: res.Tokens, usage: res.Usage, usd: res.USD, declaredTopic: declared}, stats, err
	}
	renderTopic := expected
	if renderTopic == "" {
		renderTopic = "auto"
	}
	rendered := critic.Render(forkIdx, round, renderTopic, attacks)
	if err := state.WriteRound(e.Sess, forkIdx, round, state.RoleCritic, rendered); err != nil {
		return criticRoundResult{tokens: res.Tokens, usage: res.Usage, usd: res.USD, declaredTopic: declared}, stats, err
	}
	for _, at := range attacks {
		st := ledger.StatusOpen
		if at.Disposition == critic.DispWithdraw {
			st = ledger.StatusWithdrawn
		}
		ri := at.RoundIntroduced
		_ = ledger.Append(e.Sess, ledger.Record{
			AttackID: at.AttackID, CriticIndex: forkIdx, Aspect: expected,
			RoundIntroduced:   ifNonZero(ri),
			Location:          at.Location,
			Claim:             at.Claim,
			ExpectedViolation: at.ExpectedViolation,
			Reproduction:      at.Reproduction,
			RoundLastTouched:  round,
			Status:            st,
			ReAttacked:        at.Disposition == critic.DispReAttack,
		})
	}
	_ = state.AppendTranscript(e.Sess, &state.TranscriptRecord{
		TS: time.Now().UTC(), Fork: forkIdx, Round: round, Role: "critic",
		Path: filepath.Join("forks", fmt.Sprintf("critic-%d", forkIdx), "rounds", fmt.Sprintf("r%d-critic.md", round)),
		MS:   int(res.Duration.Milliseconds()),
	})
	tokens := res.Tokens
	if tokens == 0 {
		tokens = EstimateTokens(in.SystemPrompt + res.Markdown)
	}
	usage := res.Usage
	// Compute new priorIDs as the union of priorIDs + new ids (excl. withdrawals).
	idSet := map[string]bool{}
	for _, id := range priorIDs {
		idSet[id] = true
	}
	for _, at := range attacks {
		if at.Disposition != critic.DispWithdraw {
			idSet[at.AttackID] = true
		} else {
			delete(idSet, at.AttackID)
		}
	}
	out := make([]string, 0, len(idSet))
	for id := range idSet {
		out = append(out, id)
	}
	return criticRoundResult{tokens: tokens, usage: usage, usd: res.USD, priorIDs: out, declaredTopic: declared}, stats, nil
}

func ifNonZero(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

func updateLedgerFromDefense(sess *state.Session, response string, changed []string, round int) (conceded, rebutted int) {
	for _, m := range defenseLineRE.FindAllStringSubmatch(response, -1) {
		verb, id := strings.ToLower(m[1]), m[2]
		var st ledger.Status
		switch verb {
		case "concede":
			st = ledger.StatusConceded
			conceded++
		case "rebut":
			st = ledger.StatusRebutted
			rebutted++
		case "push-back":
			// Stays open; orchestrator currently does not track count.
			continue
		}
		rec := ledger.Record{
			AttackID: id, RoundLastTouched: round, Status: st,
		}
		if st == ledger.StatusConceded {
			rec.ConcessionFiles = append([]string(nil), changed...)
		}
		_ = ledger.Append(sess, rec)
	}
	return conceded, rebutted
}
