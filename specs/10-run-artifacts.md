# Spec 10 - Run-level artifacts

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) §"Session persistence" → Layout for design intent.

**Depends on:** [04](04-cli-flags.md), [06](06-preflight.md), [07](07-claude-transcript.md), [08](08-diff.md), [09](09-state-dir.md).
**Consumed by:** [19](19-round-loop.md), [21](21-signals.md), [23](23-summary-render.md), [25](25-probes.md), [26](26-tests.md).

## Scope

In: schemas and write-rules for the four run-level files: `start.json`, `end.json`, `transcript.jsonl`, and the cross-session `log.jsonl`.

Out: per-fork files ([11](11-fork-artifacts.md)), `attacks.jsonl` ([12](12-attacks-ledger.md)), `summary.md` rendering ([23](23-summary-render.md)).

## start.json

Written atomically before any agent process spawns. Once written, never modified.

```jsonc
{
  "schema":          "agon.start.v0",
  "session_id":      "20260506T141233Z-q3a9f1",
  "started_at":      "2026-05-06T14:12:33Z",      // RFC3339 UTC

  "proposer": {
    "agent":         "claude",                     // "claude" | "codex"
    "model":         "claude-sonnet-4-6"           // "" if defaults
  },
  "critic": {
    "agent":         "codex",
    "model":         ""
  },

  "task_context":    "<verbatim first user turn>", // see [07]
  "task_source":     "transcript",                 // "transcript" | "session-id-resume" | "flag"

  "diff": {
    "from":          "HEAD",
    "to":            ".",
    "changed_lines": 47,
    "files":         ["src/api.py", "src/util.py"],
    "patch_path":    "diff.patch"                  // relative to session root; written via [11]'s helper
  },

  "config": {
    "max_turn":          6,
    "side_count":        4,
    "aspects":           ["functional-logic", "security", "code-quality", "performance"],
    "cost_cap":          50000,
    "changed_lines_min": 10,
    "format":            "markdown",
    "main_model":        "claude-sonnet-4-6",
    "side_model":        ""
  },

  "root_session": {
    "id":              "<root claude session-id, or null>",
    "transcript_path": "<absolute path or null>",
    "cwd":             "<absolute cwd>"
  },

  "agon_version":  "v0.0.1",
  "go_version":      "go1.26.0"
}
```

Rules:

- The `task_context` field is the verbatim string from [07](07-claude-transcript.md). Do not summarize, truncate, or pretty-print.
- `diff.patch_path` always points at `<session>/diff.patch`, written by [11](11-fork-artifacts.md)'s `WriteRunDiff` (because the per-fork capture reuses the helper) - but the run-level snapshot here is the *initial* one before any fork runs.
- `proposer` and `critic` use the same `agent` discriminator (`claude` | `codex`); v0 only writes `proposer.agent == "claude"` (see [06](06-preflight.md)).
- Schema version is `agon.start.v0`; bumped on any breaking change.

## end.json

Written atomically at termination, success or failure.

```jsonc
{
  "schema":          "agon.end.v0",
  "session_id":      "20260506T141233Z-q3a9f1",
  "ended_at":        "2026-05-06T14:18:02Z",

  "termination": {
    "reason":        "steady-state",     // see [20] enum
    "fork_index":    null,               // when a per-fork condition fired, the 1-based index
    "round":         null                // round number at termination
  },

  "stats": {
    "total_attacks":     15,
    "by_status": {
      "open":            0,
      "conceded":        7,
      "rebutted":        4,
      "withdrawn":       2,
      "unresolved":      2
    },
    "rounds_executed_per_fork": [6, 4, 6, 5],
    "tokens_used":       38421,
    "cost_cap":          50000,
    "wall_seconds":      329
  },

  "headline": {
    "attack_id":     "c1-3",
    "contention":    4
  } | null,

  "exit_code":       1,                  // 1 = unresolved leaves; 0 = clean.
  "summary_path":    "summary.md"
}
```

Rules:

- `termination.reason` is one of: `"steady-state" | "max-turn" | "cost-cap" | "malformed-output" | "interrupted"`. ([20](20-termination.md) owns the detection.)
- `headline` is `null` iff `by_status.unresolved == 0`.
- `exit_code` records the process exit code (0 = clean, 1 = unresolved leaves). [21](21-signals.md)/[23](23-summary-render.md) own how it maps to the process exit.
- Written *before* `log.jsonl` (see lifecycle invariants in [09](09-state-dir.md)).

## transcript.jsonl

Per-session, append-only, one line per round dispatch. Acts as a forward index of round files for human/audit consumption. Critically, this is **not** the root claude transcript - that name was reused once and we now consistently call this `transcript.jsonl` to match [01-overview.md](01-overview.md).

```jsonc
{"ts":"2026-05-06T14:12:34Z","fork":1,"round":1,"role":"critic","path":"forks/critic-1/rounds/r1-critic.md","ms":4823}
{"ts":"2026-05-06T14:12:39Z","fork":1,"round":2,"role":"proposer","path":"forks/critic-1/rounds/r2-proposer.md","ms":12104}
{"ts":"2026-05-06T14:12:51Z","fork":1,"round":3,"role":"critic","path":"forks/critic-1/rounds/r3-critic.md","ms":3902}
...
```

Fields:

- `ts` - RFC3339 UTC, when the round file was persisted.
- `fork` - 1-based critic index.
- `round` - 1-based round number within that fork.
- `role` - `"critic"` (odd rounds) or `"proposer"` (even rounds).
- `path` - relative to the session root.
- `ms` - wall time spent producing this round (subprocess invocation + parse).

Order is time-of-write; under v0's serial fork execution this is also lex-by-fork-then-round.

## log.jsonl

Cross-session, lives at `<state-dir>/log.jsonl`. Append-only. Exactly one line per `agon` invocation (or per skipped trivial diff).

Two record shapes:

```jsonc
// Completed run
{"ts":"2026-05-06T14:18:02Z","kind":"run","session":"20260506T141233Z-q3a9f1","termination":"steady-state","unresolved":2,"tokens":38421,"wall_s":329,"summary":".agon/sessions/20260506T141233Z-q3a9f1/summary.md"}

// Trivial-diff skip (see [08])
{"ts":"2026-05-06T14:09:11Z","kind":"skipped","reason":"trivial-diff","changed_lines":3,"threshold":10}

// Interrupted run (no end.json was written; entry recovered by a future spec or kept absent in v0)
// v0 does NOT write a "kind":"run" line for interrupted runs - the absence of one is the signal.
```

Rules:

- The completed-run line is the very last write of the orchestrator, after `end.json` is durable on disk (see [09](09-state-dir.md) lifecycle invariants).
- `summary` is included only when `unresolved > 0`; on clean runs the user does not need to be redirected to `summary.md`.
- `kind` is the discriminator for downstream aggregators.

## Public Go interfaces

```go
// internal/state/run.go
package state

func WriteStart(sess *Session, s *StartFile) error
func WriteEnd(sess *Session, e *EndFile) error
func AppendTranscript(sess *Session, r *TranscriptRecord) error
func AppendLog(stateDirAbs string, r *LogRecord) error

// (Schemas above expressed as Go structs with `json:"..."` tags;
// types are mechanically derivable from this spec.)
```

Wraps [09](09-state-dir.md)'s `AtomicWrite` and `AppendLine`. No business logic in this package; only marshalling.

## Test contract

- Unit: `WriteStart` then read-back yields a struct equal to the input.
- Unit: missing required field in any of the four schemas fails JSON marshalling validation.
- Golden: a fixture session producing `start.json`, `end.json`, `transcript.jsonl`, `log.jsonl` matches `testdata/golden/run-*.json`.
- Crash test: SIGKILL after `WriteStart` and before `WriteEnd` leaves `start.json` valid JSON and `log.jsonl` without a `kind:"run"` line for that session.

## Acceptance criteria

- [x] All four schemas decode/encode round-trip via `encoding/json`.
- [x] Schema version strings appear at top level of each (`agon.start.v0`, etc.).
- [x] `WriteEnd` is the last fsynced write before `AppendLog`; ordering enforced by an integration test.
- [x] No path in this spec writes the contents of `start.json` more than once per session (immutability check).
