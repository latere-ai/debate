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
	Aspects     []string
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
	WallSeconds int
	Headline    *ledger.Record
	Unresolved  int
}

// ForkOutcome carries the per-fork termination + last round.
type ForkOutcome struct {
	Index       int
	Aspect      string
	Rounds      int
	Termination TerminationReason
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

	for forkIdx := 1; forkIdx <= len(e.Aspects); forkIdx++ {
		if ctx.Err() != nil {
			sum.Termination = TermInterrupted
			break
		}
		outcome, runStop, err := e.runFork(ctx, forkIdx, e.Aspects[forkIdx-1], det, cost)
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
	sum.WallSeconds = int(time.Since(start).Seconds())
	return sum, nil
}

func (e *Engine) runFork(ctx context.Context, forkIdx int, aspect string, det *Detector, cost *CostMeter) (ForkOutcome, TerminationReason, error) {
	out := ForkOutcome{Index: forkIdx, Aspect: aspect, Termination: TermSteadyState}
	cri := e.NewCritic(forkIdx)
	a := critic.Lookup(aspect)
	var (
		forkID   string
		hist     []ForkHistory
		runStop  TerminationReason
		priorIDs []string
	)

	e.progf("[debate] fork %d/%d %s: starting", forkIdx, len(e.Aspects), aspect)

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
			e.progf("[debate] fork %d/%d %s: R%d critic running...", forkIdx, len(e.Aspects), aspect, round)
			res, stats, err := e.criticRound(ctx, cri, a, forkIdx, round, priorIDs)
			if err != nil {
				return out, "", fmt.Errorf("%w: critic %d round %d: %v", ErrAgentFatal, forkIdx, round, err)
			}
			cost.Add(res.tokens)
			e.progf("[debate] fork %d/%d %s: R%d critic done in %s (new=%d, re-attack=%d, withdraw=%d, dropped=%d)",
				forkIdx, len(e.Aspects), aspect, round, fmtDur(time.Since(roundStart)),
				stats.KeptIntroduce, stats.KeptReAttack, stats.KeptWithdraw,
				stats.DroppedNoReproduce+stats.DroppedStyle+stats.DroppedCrossAspect)
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
				e.progf("[debate] fork %d/%d %s: steady state reached at R%d", forkIdx, len(e.Aspects), aspect, round)
				break
			}
		} else {
			// Proposer round.
			e.progf("[debate] fork %d/%d %s: R%d proposer running...", forkIdx, len(e.Aspects), aspect, round)
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
			e.progf("[debate] fork %d/%d %s: R%d proposer done in %s (conceded=%d, rebutted=%d, files=%d)",
				forkIdx, len(e.Aspects), aspect, round, fmtDur(time.Since(roundStart)),
				conceded, rebutted, len(pr.ChangedFiles))
		}
	}
	if out.Rounds >= e.MaxTurn && runStop == "" {
		out.Termination = TermMaxTurn
	}
	e.progf("[debate] fork %d/%d %s: terminated %s after R%d (tokens used: %d)",
		forkIdx, len(e.Aspects), aspect, ifEmpty(string(runStop), string(out.Termination)),
		out.Rounds, cost.Used())
	return out, runStop, nil
}

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func ifEmpty(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

type criticRoundResult struct {
	tokens   int
	priorIDs []string
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
	attacks, stats, err := critic.Parse(res.Markdown, a.Name, forkIdx, round, priorIDs, critic.ParseOption{})
	if err != nil {
		return criticRoundResult{tokens: res.Tokens}, stats, err
	}
	rendered := critic.Render(forkIdx, round, a.Name, attacks)
	if err := state.WriteRound(e.Sess, forkIdx, round, state.RoleCritic, rendered); err != nil {
		return criticRoundResult{tokens: res.Tokens}, stats, err
	}
	for _, at := range attacks {
		st := ledger.StatusOpen
		if at.Disposition == critic.DispWithdraw {
			st = ledger.StatusWithdrawn
		}
		ri := at.RoundIntroduced
		_ = ledger.Append(e.Sess, ledger.Record{
			AttackID: at.AttackID, CriticIndex: forkIdx, Aspect: a.Name,
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
	return criticRoundResult{tokens: tokens, priorIDs: out}, stats, nil
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
