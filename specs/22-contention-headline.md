# Spec 22 - Contention scoring and headline selection

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Headline contradicting signal" for design intent.

**Depends on:** [12](12-attacks-ledger.md).
**Consumed by:** [23](23-summary-render.md).

## Scope

In: pure-function scoring of an attack, tie-break rule, the cross-fork aggregation that picks the single headline. No I/O, no LLM, no semantic ranking.

Out: rendering the headline into `summary.md` ([23](23-summary-render.md)).

## Scoring rule

From [01-overview.md](01-overview.md):

```
contention(attack) = rounds_survived + (1 if critic re_attacked_after_defense else 0)
```

Where:

- `rounds_survived = round_last_touched - round_introduced`
- `re_attacked_after_defense = ledger.Record.ReAttacked == true`

`rounds_survived` is the number of round-pairs the attack lived through; an attack introduced in R1 and last touched in R3 has `rounds_survived = 2`. Re-attack adds 1.

Maximum theoretical score with `--max-turn 6`: 5 (introduced R1, last touched R5, re-attacked = 4 + 1 = 5).

## Public Go interfaces

```go
// internal/summary/headline.go
package summary

// Score returns the contention score for one attack record. Pure function.
func Score(r ledger.Record) int

// PickHeadline returns the single headline attack across the entire run,
// or nil if there are zero unresolved attacks.
//
// Selection rule:
//   1. Filter to records with status = "unresolved".
//   2. Sort by Score(r) DESC, then by RoundIntroduced ASC, then by
//      AttackID ASC (lexicographic).
//   3. Return the first.
func PickHeadline(records []ledger.Record) *ledger.Record

// SortByContention returns records sorted using the same rule as
// PickHeadline (same total order). Used by [23] for the "Other unresolved"
// section.
func SortByContention(records []ledger.Record) []ledger.Record
```

## Tie-break

Three keys, in priority order:

1. **Score DESC.** Highest contention first.
2. **`RoundIntroduced` ASC.** Earlier discovery wins; the bug that survived the most rounds since R1 is more contested than one introduced at R5.
3. **`AttackID` ASC** (lexicographic). Final disambiguator; deterministic.

The third key is required so the sort is total - without it, ties would surface as nondeterministic across runs (Go's `sort.Slice` is not stable; even `sort.SliceStable` doesn't help when input order varies).

## Cross-fork rule

Records are aggregated across all forks before scoring; `attack_id` carries the critic index, so cross-fork ids never collide.

The headline is *one* attack across the run, not one per fork. (Per-fork headlines were considered and rejected for v0: they'd dilute the surfacing value.)

## Worked example

```
agg := []ledger.Record{
    {AttackID: "c1-1", RoundIntroduced: 1, RoundLastTouched: 5, ReAttacked: true,  Status: "unresolved"},  // score=5
    {AttackID: "c1-3", RoundIntroduced: 1, RoundLastTouched: 5, ReAttacked: false, Status: "unresolved"},  // score=4
    {AttackID: "c2-2", RoundIntroduced: 3, RoundLastTouched: 5, ReAttacked: true,  Status: "unresolved"},  // score=3
    {AttackID: "c1-7", RoundIntroduced: 1, RoundLastTouched: 1, ReAttacked: false, Status: "conceded"},   // skipped (not unresolved)
}

PickHeadline(agg) -> c1-1
```

## Future-proofing note

The doc says ([01-overview.md](01-overview.md)): "If contention turns out to be a poor proxy in practice, upgrade later by having both sides self-report confidence each round and weighting by `confidence_critic × confidence_proposer`." That's v1+; v0 ships the cheap rule.

This spec deliberately keeps `Score` as a pure function so swapping its implementation later doesn't ripple through the surface.

## Test contract

- Unit: `Score` for the four shapes (low/high `rounds_survived` × `ReAttacked` true/false).
- Unit: `PickHeadline` returns nil for zero unresolved.
- Unit: tie-break in all three keys covered.
- Property: `PickHeadline(shuffle(agg)) == PickHeadline(agg)` (determinism).

## Acceptance criteria

- [x] `Score` is a pure function (no I/O, no clock, no rand).
- [x] `PickHeadline` returns nil iff no record has `Status == "unresolved"`.
- [x] Determinism property holds across 100 random shuffles in the test suite.
- [x] No semantic ranking, no LLM call, no embeddings - the rule is arithmetic only.
