# Spec 20 - Termination conditions and cost-cap accounting

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) §"Termination conditions" for design intent.

**Depends on:** [04](04-cli-flags.md), [12](12-attacks-ledger.md), [14](14-attack-parser.md), [19](19-round-loop.md).
**Consumed by:** [19](19-round-loop.md), [21](21-signals.md), [23](23-summary-render.md).

## Scope

In: detection of all termination conditions, the cost-cap accumulator (token estimator across calls), the malformed-output two-rounds-running rule, and the per-fork vs run-level distinction.

Out: writing `end.json` ([10](10-run-artifacts.md)), signal handling ([21](21-signals.md)), exit code ([23](23-summary-render.md)).

## Termination reasons

```go
type TerminationReason string
const (
    TermSteadyState     TerminationReason = "steady-state"
    TermMaxTurn         TerminationReason = "max-turn"
    TermCostCap         TerminationReason = "cost-cap"
    TermMalformedOutput TerminationReason = "malformed-output"
    TermInterrupted     TerminationReason = "interrupted"
)
```

`steady-state` and `max-turn` are per-fork (a fork can hit one of these and the next fork still runs). `cost-cap`, `malformed-output`, and `interrupted` are run-level - the orchestrator stops scheduling new forks and rolls remaining attacks into `unresolved`.

## Public Go interfaces

```go
// internal/round/termination.go
package round

type ForkHistory struct {
    Round           int
    NewAttacks      int  // count of Disposition=Introduce in this critic round
    ReAttacks       int
    Withdrawn       int
    ParseErrors     int  // [14] returned a non-nil err
    MalformedFlag   bool // [14] returned a typed parse error or zero attacks AND ParseErrors > 0
}

type Detector struct {
    MaxTurn       int
    CostCap       int
    MaxRounds     int      // alias of MaxTurn for clarity in tests
}

// SteadyState returns true iff the last two critic rounds in this fork
// produced zero new attacks AND zero re-attacks. (Pure new attacks
// dropping to zero isn't enough - re-attacks also count as "still
// disputing.")
func (d *Detector) SteadyState(history []ForkHistory) bool

// MaxTurnReached: round number reached d.MaxTurn.
func (d *Detector) MaxTurnReached(round int) bool

// MalformedTwice: the last two critic rounds in this fork BOTH had
// MalformedFlag = true.
func (d *Detector) MalformedTwice(history []ForkHistory) bool

// CostCapHit returns true iff totalTokens >= d.CostCap.
func (d *Detector) CostCapHit(totalTokens int) bool
```

`SteadyState` requires *at least three* critic rounds executed (R1, R3, R5). It is undefined for fewer; the round loop must check `len(history) >= 3` before calling.

## Cost-cap accumulator

```go
// internal/round/cost.go

type CostMeter struct {
    cap     int
    used    int
    perCall []int
}

func (c *CostMeter) Add(tokens int)
func (c *CostMeter) Used() int
func (c *CostMeter) Remaining() int   // cap - used; can go negative
func (c *CostMeter) ExceedsCap() bool // used >= cap
```

The meter is a single instance per `Engine`. Both proposer and critic calls feed `Add(result.Tokens)` after each invocation completes ([19](19-round-loop.md)).

### Token estimation when the agent doesn't report

claude `--output-format json` populates `usage.input_tokens` + `usage.output_tokens` (newer versions). When absent, the orchestrator estimates:

```
estimated_tokens = ceil(len(prompt+response) / 4)
```

(Heuristic: ~4 chars/token; well known floor.) The estimator is documented in code as best-effort; v0 does not require an exact accounting.

For codex, the JSON event stream sometimes carries a `usage` event. When absent, same 4-chars/token estimate.

### Cost-cap behavior

When `CostCapHit()` becomes true *after* a round completes:

1. The current round's outputs are persisted normally (round file, ledger entries).
2. The orchestrator stops scheduling further rounds and stops launching new forks.
3. Pending attacks in the active fork transition to `unresolved` ([12](12-attacks-ledger.md)).
4. `Engine.Run` returns with `Termination = TermCostCap`.

The cap is a *post-call* check; it never aborts a call mid-stream.

## Malformed-output rule

`MalformedFlag` is true when [14](14-attack-parser.md) returns:

- `len(attacks) == 0` AND `stats.Total > 0` (parser saw sections but dropped all).
- OR a typed parse error (`ErrMalformedHeader`, `ErrUnclosedFence`, etc.).

Two rounds in a row with `MalformedFlag == true` triggers `TermMalformedOutput`. Defensive: the model is broken or the prompt collapsed; further rounds are unlikely to recover.

## Steady-state worked example

```
Fork 1 history:
  Round 1 (critic): NewAttacks=4, ReAttacks=0
  Round 2 (proposer)         (not in history; only critic rounds counted)
  Round 3 (critic): NewAttacks=1, ReAttacks=2
  Round 4 (proposer)
  Round 5 (critic): NewAttacks=0, ReAttacks=0   ← steady-state begins
  Round 6 (proposer)
  Round 7 (critic): NewAttacks=0, ReAttacks=0   ← two rounds running

→ SteadyState returns true. Fork wraps with TermSteadyState; no R8 proposer round runs.
```

(Implementation note: only critic rounds are appended to `history`; proposer rounds are tracked separately but not used by the detector.)

## Run-level vs per-fork interaction

Per-fork termination (`steady-state`, `max-turn`) closes that fork only; the loop moves to the next fork.

Run-level termination (`cost-cap`, `malformed-output`, `interrupted`) breaks the outer loop. Forks not yet started never run. The active fork's already-persisted state stays; pending attacks in that fork become `unresolved`.

The summary's `Termination` field is the *run-level* reason if any was hit; otherwise it's `steady-state` (every fork either hit steady-state or max-turn naturally). Per-fork outcomes are recorded in `Summary.Forks[i].Termination` (`steady-state` vs `max-turn` only - run-level termination is reflected at the summary level, not per-fork).

## Test contract

- Unit: each detector method has a positive and negative case.
- Unit: `SteadyState` returns false when only one critic round has run.
- Unit: `MalformedTwice` requires consecutive malformed; one malformed + one good resets.
- Unit: cost-cap accumulator with a sequence of (1k, 5k, 10k, 40k) tokens trips the 50k cap on the fourth call.
- Integration: a synthetic run with three forks and `--max-turn 4` produces three `TermMaxTurn` per-fork outcomes and an overall `TermSteadyState` (because no run-level condition fired).

## Acceptance criteria

- [x] All five `TerminationReason` values reachable in the test suite.
- [x] Cost meter tracks both proposer and critic tokens.
- [x] Steady-state detection requires `len(history) >= 3`; below that returns false.
- [x] Malformed-twice resets on a single good round.
- [x] Detector is a pure value type; no global state.
