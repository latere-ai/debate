# Spec 12 — Attacks ledger (`attacks.jsonl`)

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"`attacks.jsonl` schema" for design intent.

**Depends on:** [09](09-state-dir.md).
**Consumed by:** [14](14-attack-parser.md), [19](19-round-loop.md), [20](20-termination.md), [22](22-contention-headline.md), [23](23-summary-render.md).

## Scope

In: the on-disk schema, append-only state-transition rules, the forward-read aggregator that computes per-attack final state, and the recovery contract for interrupted sessions.

Out: parsing the critic's raw markdown output (that's [14](14-attack-parser.md)); contention scoring (that's [22](22-contention-headline.md)).

## Schema

`attacks.jsonl`, append-only, one JSON object per line. Every line is a *transition record* for an `attack_id`; the final state of an attack is computed by reading forward and keeping the latest non-empty value for each field.

```jsonc
{
  "schema":             "debate.attack.v0",
  "ts":                 "2026-05-06T14:12:34Z",
  "attack_id":          "c1-3",                // critic-index + sequence; stable across rounds
  "critic_index":       1,
  "aspect":             "security",

  // First-introduction fields. Set on the record that introduces the attack.
  // null on later transition records (means "unchanged").
  "round_introduced":   1,
  "location":           "src/api.py:88",
  "claim":              "input not sanitized before LIKE pattern",
  "expected_violation": "SQL injection via search parameter",
  "reproduction":       "GET /search?q=%' OR 1=1--",

  // Live state.
  "round_last_touched": 3,
  "status":             "rebutted",            // open | conceded | rebutted | withdrawn | unresolved
  "rounds_survived":    2,
  "re_attacked":        true,
  "concession_files":   ["src/api.py"],        // populated when status -> conceded

  // Optional. Pointers into round files for audit.
  "introduced_in":      "forks/critic-1/rounds/r1-critic.md",
  "last_touched_in":    "forks/critic-1/rounds/r3-critic.md",

  // Body spill (see [09]). When the inline fields exceed 64KB, the
  // record is written without claim/expected_violation/reproduction
  // and body_path points at a sidecar markdown file.
  "body_path":          ""                     // "" = inline; otherwise relative path
}
```

## State machine

```
                 introduce (R1, by critic)
                          │
                          ▼
                       OPEN ───── proposer concedes ─────► CONCEDED
                       │ ▲                                     │
                       │ └── proposer rebuts ─► REBUTTED       (terminal)
                       │                       │
                       │                       │ critic re-attacks (next critic round)
                       │                       ▼
                       └─────── (back to OPEN, re_attacked = true)

   At any open state at termination → status flips to UNRESOLVED.
   Critic withdraws at any time → WITHDRAWN (terminal).
```

Transition rules (each emits one append-only record):

| From | Event | To | Required record fields beyond identity |
|---|---|---|---|
| ∅ | introduce | open | location, claim, expected_violation, reproduction, round_introduced, introduced_in |
| open | concede | conceded | concession_files, round_last_touched, last_touched_in |
| open | rebut | rebutted | round_last_touched, last_touched_in |
| rebutted | re-attack | open | round_last_touched (incremented), last_touched_in, re_attacked=true |
| open | withdraw | withdrawn | round_last_touched, last_touched_in |
| any | terminate-as-unresolved | unresolved | round_last_touched (= max round seen) |

Re-attacks reuse the original `attack_id` and *do not* introduce a new attack; new attacks discovered by the critic at later rounds get fresh ids.

## Public Go interfaces

```go
// internal/ledger/attacks.go
package ledger

type Status string
const (
    StatusOpen       Status = "open"
    StatusConceded   Status = "conceded"
    StatusRebutted   Status = "rebutted"
    StatusWithdrawn  Status = "withdrawn"
    StatusUnresolved Status = "unresolved"
)

type Record struct {
    Schema             string    `json:"schema"`
    TS                 time.Time `json:"ts"`
    AttackID           string    `json:"attack_id"`
    CriticIndex        int       `json:"critic_index"`
    Aspect             string    `json:"aspect"`
    RoundIntroduced    *int      `json:"round_introduced,omitempty"`
    Location           string    `json:"location,omitempty"`
    Claim              string    `json:"claim,omitempty"`
    ExpectedViolation  string    `json:"expected_violation,omitempty"`
    Reproduction       string    `json:"reproduction,omitempty"`
    RoundLastTouched   int       `json:"round_last_touched"`
    Status             Status    `json:"status"`
    RoundsSurvived     int       `json:"rounds_survived"`
    ReAttacked         bool      `json:"re_attacked"`
    ConcessionFiles    []string  `json:"concession_files,omitempty"`
    IntroducedIn       string    `json:"introduced_in,omitempty"`
    LastTouchedIn      string    `json:"last_touched_in,omitempty"`
    BodyPath           string    `json:"body_path,omitempty"`
}

// Append writes one transition record.
func Append(sess *state.Session, r Record) error

// Aggregate reads attacks.jsonl forward and returns the final state per attack-id.
// Records with the same attack_id are folded: later non-zero/non-empty fields
// supersede earlier ones; the last status wins.
func Aggregate(sess *state.Session) (map[string]Record, error)

// Pending returns the subset of Aggregate whose status is open or rebutted —
// i.e., attacks still alive and eligible for further rounds. Used by [20]'s
// steady-state detector.
func Pending(agg map[string]Record) []Record
```

## Body spill

When `len(claim) + len(expected_violation) + len(reproduction)` would push the JSONL line over 64KB, the orchestrator writes the bodies to:

```
forks/critic-<i>/attacks/<attack_id>.md
```

…with sections `## Claim`, `## Expected violation`, `## Reproduction`, and sets `body_path` to that relative path. Inline fields are omitted from the JSONL record.

`Aggregate` resolves `body_path` lazily: callers that need the body call `LoadBody(sess, r) (Record, error)` which reads and inlines the markdown sections.

## Recovery from interrupted runs

`Aggregate` is robust against a partial trailing line: the reader stops at the last newline before EOF, parses everything up to that point. The trailing partial line is silently discarded.

A partial transition (e.g., the orchestrator crashed *between* writing the proposer-defense round file and appending the corresponding ledger record) is tolerated: the ledger may be missing the most recent transition. The summary aggregator ([23](23-summary-render.md)) treats it as if the transition never happened, which is conservative and correct (the attack stays in its prior state).

## Test contract

- Unit: every transition rule round-trips through `Append` → `Aggregate`.
- Unit: re-attack carries `re_attacked = true` and increments `rounds_survived`.
- Unit: 64KB body triggers spill to `attacks/<id>.md` and `LoadBody` reconstructs.
- Unit: a truncated final line is discarded by `Aggregate`.
- Golden: a fixture sequence of records produces a known `map[string]Record`.

## Acceptance criteria

- [ ] All five `Status` values reachable end-to-end.
- [ ] No code path under `internal/ledger/` opens `attacks.jsonl` for writing without `O_APPEND`.
- [ ] Body spill triggers at exactly 64KB (test with 65535-byte and 65537-byte fixtures).
- [ ] `Aggregate`'s output is deterministic for a given input (no map-iteration order leaks).
