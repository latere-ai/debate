# Spec 19 — Round loop and per-fork orchestration

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Rounds" and "Lifecycle invariants" for design intent.

**Depends on:** [06](06-preflight.md), [07](07-claude-transcript.md), [08](08-diff.md), [09](09-state-dir.md), [10](10-run-artifacts.md), [11](11-fork-artifacts.md), [12](12-attacks-ledger.md), [14](14-attack-parser.md), [15](15-aspect-prompts.md), [17](17-claude-proposer.md), [18](18-critic-drivers.md).
**Consumed by:** [20](20-termination.md), [21](21-signals.md), [22](22-contention-headline.md), [23](23-summary-render.md).

## Scope

In: the central state machine. R0 setup → for each fork (serial in v0): R1 attack → R2 defense → R3..R(max) cross-examination, with file-before-pointer ordering, per-fork diff capture, and ledger updates after every round.

Out: the *detection* of termination conditions ([20](20-termination.md)), signal handling ([21](21-signals.md)), summary rendering ([23](23-summary-render.md)).

## Public Go interfaces

```go
// internal/round/loop.go
package round

type Engine struct {
    Sess        *state.Session
    Plan        *cli.Plan
    Proposer    agent.Proposer        // currently agent.ClaudeProposer
    NewCritic   func(forkIdx int) agent.Critic  // factory; constructs codex or claude critic with the right aspect
    CostCap     int
    HookMode    bool
}

// Run executes the orchestration. Returns a Summary the renderer ([23])
// consumes. Errors are typed (ErrInterrupted, ErrCostCap,
// ErrMalformedTwice, ErrAgentFatal); see [20] for detection rules.
func (e *Engine) Run(ctx context.Context) (*Summary, error)

type Summary struct {
    Sess         *state.Session
    Termination  TerminationReason     // see [20]
    Forks        []ForkOutcome         // 1-based indexed; len = side-count
    TokensUsed   int
    WallSeconds  int
    Headline     *ledger.Record         // nil iff zero unresolved
    Unresolved   int
}

type ForkOutcome struct {
    Index       int
    Aspect      string
    Rounds      int                    // last round actually executed
    Termination TerminationReason      // per-fork (the steady-state vs max-turn distinction)
}
```

The `agent.Proposer` interface is satisfied by [17](17-claude-proposer.md)'s `ClaudeProposer`; v1 codex-as-proposer adds another implementation.

## Round numbering

```
R0 = orchestrator setup (no agent call)
R1 = critic attack (odd)
R2 = proposer defense (even)
R3 = critic cross-examination (odd)
R4 = proposer defense   (even)
...
R(max_turn) = whichever role's parity that round number maps to
```

`max_turn` from [04](04-cli-flags.md) is the upper bound; actual rounds may be fewer (steady-state termination).

## Sequence per fork

(Pseudo-code; the real implementation is in `internal/round/`.)

```
forkIdx := 1..plan.SideCount
for each forkIdx:
    # R0
    forkDiff := input.Compute(...)                                # [08]
    state.WriteForkDiff(sess, forkIdx, forkDiff.Patch)             # [11]
    aspect := critic.Lookup(plan.Forks[forkIdx-1].Aspect)         # [15]
    cri := e.NewCritic(forkIdx)
    var forkID string                                             # claude fork session id

    # R1 — attack
    raw1, _ := cri.Round(ctx, CriticInput{
        Aspect: aspect, CriticIndex: forkIdx, Round: 1,
        SystemPrompt: critic.Assemble(aspect, forkIdx, 1, ""),
        TaskContext: start.TaskContext,
        DiffPatch: forkDiff.Patch,
        Cwd: plan.Cwd,
    })                                                             # [18]
    attacks1, stats1, _ := critic.Parse(raw1.Markdown, aspect.Name, forkIdx, 1, /*priorIDs*/ nil)  # [14]
    state.WriteRound(sess, forkIdx, 1, RoleCritic, critic.Render(...))                              # [11]
    for a in attacks1: ledger.Append(sess, recordFromAttack(a))    # [12]
    state.AppendTranscript(sess, forkIdx, 1, "critic", ...)        # [10]

    # R2 — defense
    pointer := buildPointer(sess, forkIdx, 1)
    if no fork yet:
        res2, _ := proposer.FirstRound(ctx, pointer)               # [17]
        forkID = res2.ForkID
        state.WriteProposerState(sess, forkIdx, &ProposerState{Agent: "claude", ForkSessionID: forkID, RootSessionID: plan.SessionID})
    else:
        res2, _ := proposer.NextRound(ctx, forkID, pointer)
    body := res2.Response + modifiedFilesBlock(res2.ChangedFiles)
    state.WriteRound(sess, forkIdx, 2, RoleProposer, body)
    updateLedgerFromDefense(res2.Response)                          # parse "concede c1-1", "rebut c1-2", etc.
    state.AppendTranscript(sess, forkIdx, 2, "proposer", ...)
    incrementCost(res2.Tokens)

    # R3..R(max)
    for r := 3 .. plan.MaxTurn:
        if r is odd:                                                # critic round
            priorIDs := pendingIDsForFork(sess, forkIdx)             # [12]
            rawN, _ := cri.Round(ctx, ... PriorRoundFiles: refs(forkIdx, r-2, r-1) ...)
            attacksN, statsN, _ := critic.Parse(rawN.Markdown, aspect.Name, forkIdx, r, priorIDs)
            state.WriteRound(sess, forkIdx, r, RoleCritic, critic.Render(...))
            for a in attacksN: ledger.Append(sess, recordFromAttack(a))   # transitions: introduce | re-attack | withdraw
            state.AppendTranscript(...)
            incrementCost(rawN.Tokens)

            if termination.SteadyState(stats0=statsN, prev=statsPrev) ||
               termination.MalformedTwice(history) ||
               termination.CostCap(...):
                set forkOutcome.Termination
                break
        else:                                                       # proposer round
            pointer := buildPointer(sess, forkIdx, r-1)
            resN, _ := proposer.NextRound(ctx, forkID, pointer)
            state.WriteRound(sess, forkIdx, r, RoleProposer, resN.Response + modifiedFilesBlock(...))
            updateLedgerFromDefense(resN.Response)
            state.AppendTranscript(...)
            incrementCost(resN.Tokens)

    # Fork wrap
    forkOutcome.Rounds = r
    if reached MaxTurn: forkOutcome.Termination = TermMaxTurn

# All forks done
finalizeUnresolved(sess)                                            # any open|rebutted -> unresolved
summary := buildSummary(...)                                        # [22], [23]
return summary, nil
```

## File-before-pointer (invariant)

Every pointer message dispatched to either agent references a file already on disk:

| Round | Pointer points at | File written when |
|---|---|---|
| R1's call | (none — system prompt embeds task + diff) | n/a |
| R2's call | `r1-critic.md` | written by [11](11-fork-artifacts.md) before `proposer.FirstRound` |
| R3's call | `r2-proposer.md` and `r1-critic.md` | both already on disk |
| R4's call | `r3-critic.md` | written before `proposer.NextRound` |
| R(n) | `r(n-1)-<role>.md` and earlier | all on disk before dispatch |

Violations are programmer errors. The driver's `WriteRound` returns immediately on `O_EXCL` collision; file-before-pointer is a contract, not a runtime check.

## Pointer message format

```
Some comments at @forks/critic-<i>/rounds/r<n>-<role>.md. Please resolve or respond. If you disagree, please raise it.
```

(Verbatim from [01-overview.md](01-overview.md) §"Payload via file reference". Path is relative to `plan.Cwd`.)

For the critic's pointer (proposer responses):

```
Proposer responses at @forks/critic-<i>/rounds/r<n>-proposer.md. Review the defenses; for any unresolved attack, decide whether to re-attack or withdraw.
```

These are the *only* orchestrator-authored agent-visible texts.

## Defense-response parsing

The proposer is asked to emit per-attack-id dispositions inline. The orchestrator's lightweight parser scans the response for lines matching:

```
^\s*(concede|rebut|push-back)\s+(c\d+-\d+)\b
```

For each match, the orchestrator updates `attacks.jsonl` accordingly:

- `concede` → `Status = conceded`; populate `concession_files` from `res.ChangedFiles`.
- `rebut` → `Status = rebutted`.
- `push-back` → keep `Status = open` and increment a per-attack `push_back_count`; ≥ 2 push-backs is treated as `rebutted` (one push-back per attack-id is the spec rule, see [01-overview.md](01-overview.md) §"R2 Defense").

Lines that don't match are ignored (the proposer may also produce free-form prose). The full response is still persisted to `r<n>-proposer.md`.

## Per-fork diff refresh

Before each fork's R1, [08](08-diff.md)'s `Compute` runs against the current working tree (which may include prior forks' concessions in v0's serial mode). The result lands in `forks/critic-<i>/diff.patch` ([11](11-fork-artifacts.md)) and is the diff text fed to the critic prompt.

## Test contract

- Unit (mocks): a fork with R1=2 attacks, R2=1 conceded + 1 rebutted, R3=1 re-attack, R4=2 conceded → ledger ends with 4 records, all conceded except 1 unresolved at termination.
- Unit: file-before-pointer violation (write skipped) returns the `O_EXCL` error.
- Integration (mock claude+codex, real `state` + `ledger`): three forks serialize correctly, second fork sees first fork's concession in its diff.
- Race: `Run` cancelled mid-fork via context returns `ErrInterrupted` and leaves `attacks.jsonl` valid (truncated-line tolerant).

## Acceptance criteria

- [ ] All forks run serially; concurrent fork execution is rejected at construction.
- [ ] Each round's file is durable on disk before the next agent call.
- [ ] Ledger is updated within the same round-loop iteration that produced the round file.
- [ ] Per-fork `diff.patch` is captured before R1's critic call.
- [ ] Summary returned by `Run` is fully populated; no field defaulted to a panic-style sentinel.
