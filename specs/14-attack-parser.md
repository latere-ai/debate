# Spec 14 - Attack parser, normalizer, filters

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Rounds" → R1 Attack for design intent.

**Depends on:** [12](12-attacks-ledger.md), [13](13-critic-output-format.md).
**Consumed by:** [11](11-fork-artifacts.md), [19](19-round-loop.md), [20](20-termination.md).

## Scope

In: parsing the markdown defined in [13](13-critic-output-format.md), deterministic id normalization, the style-drop heuristic, the reproduction-required filter, and the parse-normalize-persist ordering invariant.

Out: invoking the critic ([18](18-critic-drivers.md)), persisting the result ([11](11-fork-artifacts.md), [12](12-attacks-ledger.md)).

## Public Go interfaces

```go
// internal/critic/parser.go
package critic

type Attack struct {
    AttackID          string
    CriticIndex       int
    Aspect            string
    RoundIntroduced   int
    Round             int                 // current round number
    Disposition       Disposition         // Introduce | ReAttack | Withdraw
    Location          string
    Claim             string
    ExpectedViolation string
    Reproduction      string
    WithdrawReason    string              // only when Disposition == Withdraw
}

type Disposition int
const (
    DispIntroduce Disposition = iota
    DispReAttack
    DispWithdraw
)

type ParseStats struct {
    Total                int
    KeptIntroduce        int
    KeptReAttack         int
    KeptWithdraw         int
    DroppedNoReproduce   int
    DroppedStyle         int
    DroppedCrossAspect   int
    Renamed              int    // ids reassigned by the normalizer
}

// Parse reads the critic's raw markdown output and returns a list of
// surviving attacks plus diagnostic stats.
//
// Steps in order:
//   1. Tokenize markdown into sections.
//   2. Validate top header + aspect line.
//   3. For each `##` section: extract id, location, disposition, body fields.
//   4. Filter: drop style-shaped, drop reproduction-less, drop cross-aspect.
//   5. Normalize ids: detect collisions and gaps, reassign deterministically.
//
// `expectedAspect` is the aspect this critic was prompted with. `criticIndex`
// is the 1-based fork index. `round` is the current round number.
// `priorAttackIDs` is the set of ids alive in attacks.jsonl entering this round
// (used to validate re-attack/withdraw references and to seed the next id).
func Parse(raw string, expectedAspect string, criticIndex, round int, priorAttackIDs []string) ([]Attack, ParseStats, error)
```

## Tokenization rules

Mechanical, no LLM:

- Header line at column 0 starting with `#` followed by space.
- Section bodies span from one `##` header to the next `##` header (or EOF).
- Field labels recognized inside a section: `claim:`, `expected violation:`, `reproduction:`. Match case-insensitively, label must be at start of a line.
- The `reproduction:` label is followed by a fenced block (` ``` ` or `~~~`); the block content is captured verbatim.

## Filters

### F1 - reproduction-required

If a section has no fenced block immediately following the `reproduction:` label, drop it. Counter: `DroppedNoReproduce`.

### F2 - style-shape heuristic

Drop a section whose `claim:` matches *any* of the following and whose `expected violation:` does **not** name a concrete behavior:

```
re.MatchString(`(?i)should be (named|called|written as|shorter|more idiomatic|simpler)`)
re.MatchString(`(?i)(naming|formatting|style) (preference|convention)`)
re.MatchString(`(?i)consider (renaming|reformatting|restyling)`)
```

"Concrete behavior" detector: `expected_violation` contains *any* of `panic`, `crash`, `leak`, `inject`, `bypass`, `corrupt`, `race`, `deadlock`, `OOM`, `hang`, `timeout`, `incorrect output`, `wrong result`, `silently swallowed`, `breaks contract`, OR a fenced code block. Counter: `DroppedStyle`.

(`allow_style_attacks = true` in [05](05-config-file.md) skips this filter.)

### F3 - cross-aspect

Each aspect carries a forbidden-keyword set in [15](15-aspect-prompts.md). If the section's `claim` matches a keyword from another aspect *and* none from its own, drop it. Counter: `DroppedCrossAspect`.

## Id normalizer

Deterministic, table-driven. Inputs: emitted ids in document order, expected `c<critic-index>-<n>` shape.

Rules in order:

1. **Shape check.** An id not matching `^c<critic-index>-(\d+)$` is renamed to the next sequence number after the highest valid one in this document. Counter: `Renamed`.
2. **Collision check.** A duplicate id (same string already seen in this document) gets the next sequence number. Counter: `Renamed`.
3. **Continuity.** After step 2, the surviving ids are *renumbered* to be contiguous starting at the next available sequence after `priorAttackIDs`'s max for this critic. (E.g., if R1 had `c1-1, c1-2, c1-3`, R3 starts at `c1-4` for new attacks; re-attacks keep their old id.)

Example:

```
emitted:    c1-1, c1-3, c1-3, c1-bogus
priorMax:   none (R1)
result:     c1-1, c1-2, c1-3, c1-4
            (c1-3 dup -> c1-4? no: continuity collapses it)
            Final: c1-1, c1-2, c1-3, c1-4 (contiguous)
```

For `(re-attack)` and `(withdraw)` dispositions, the id must already exist in `priorAttackIDs`. If it doesn't, the section is treated as `Introduce` with a fresh id; counter: `Renamed`.

## Parse-normalize-persist ordering (invariant)

The orchestrator must:

1. Capture the critic's raw stdout.
2. Call `Parse(...)`. Discard the raw stdout from the on-disk record.
3. Render the survivors back to markdown via `Render(attacks)` (the inverse of `Parse`).
4. Atomic-write the rendered markdown to `forks/critic-<i>/rounds/r<n>-critic.md`.
5. Append one `attacks.jsonl` record per surviving attack ([12](12-attacks-ledger.md)).
6. Only after steps 4 and 5 are durable, dispatch the pointer to the proposer ([19](19-round-loop.md)).

```go
func Render(top RoundHeader, attacks []Attack) []byte
```

The renderer is the canonical writer; what the proposer reads via `@<path>` always matches what's in `attacks.jsonl`.

## Test contract

- Fixture: each disposition (introduce, re-attack, withdraw) round-trips through `Parse` → `Render` losslessly.
- Fixture: each filter rule has a positive (drop) and a negative (keep) test case.
- Fixture: id normalizer's worked example above produces the listed result.
- Fuzz: random valid documents up to 32 attacks parse without panic.
- Property: `len(Parse(Render(x))) == len(x)` for any well-formed `x`.

## Acceptance criteria

- [x] All four counters (`DroppedNoReproduce`, `DroppedStyle`, `DroppedCrossAspect`, `Renamed`) reachable in tests.
- [x] Parse failure on a malformed top header (`# wrong`) returns a typed error with line number, not a panic.
- [x] `Render(Parse(x))` is byte-identical to `x` modulo the trailing-`---` separator (the renderer always emits one).
- [x] No regex compiles on every call; all patterns are package-level `var`s.
