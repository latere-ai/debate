# Spec 09 - State directory and atomic writes

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Session persistence" for design intent.

**Depends on:** [02](02-go-module.md), [04](04-cli-flags.md), [06](06-preflight.md).
**Consumed by:** [10](10-run-artifacts.md), [11](11-fork-artifacts.md), [12](12-attacks-ledger.md), [19](19-round-loop.md), [21](21-signals.md), [23](23-summary-render.md).

## Scope

In: per-run session folder creation, atomic-write primitives, append-only-write primitives, fsync policy, the lifecycle invariants every later persistence spec depends on.

Out: any specific file format (those live in [10](10-run-artifacts.md), [11](11-fork-artifacts.md), [12](12-attacks-ledger.md)).

## Layout (recap)

```
<state-dir>/                                    # default ".debate"
  log.jsonl                                     # cross-session, append-only
  sessions/
    <session-id>/                               # one per run
      start.json                                # written before any agent process spawns
      end.json                                  # written at termination
      summary.md                                # written at termination
      transcript.jsonl                          # append-only index of round files
      attacks.jsonl                             # append-only attack ledger
      forks/
        critic-1/
          proposer-state.json
          diff.patch
          rounds/
            r1-critic.md
            r2-proposer.md
            ...
        critic-2/...
```

Schemas for the individual files live in [10](10-run-artifacts.md) (run-level), [11](11-fork-artifacts.md) (per-fork), [12](12-attacks-ledger.md) (attacks).

## Session id format

`<ISO8601>-<rand6>` where `ISO8601` is `YYYYMMDDTHHMMSSZ` (UTC, compact form, no separators) and `rand6` is six lowercase base32 chars from `crypto/rand`. Example: `20260506T141233Z-q3a9f1`.

Reasons: lexicographically sortable (chronological `ls` order); short enough to type; UTC removes timezone ambiguity.

## Public Go interfaces

```go
// internal/state/dir.go
package state

type Session struct {
    Root          string  // absolute <state-dir>/sessions/<id>
    ID            string
    StateDirAbs   string  // absolute <state-dir>
    StartedAt     time.Time
}

// NewSession creates the session folder skeleton:
//   - sessions/<id>/
//   - sessions/<id>/forks/critic-N/  for N in 1..forkCount
//   - sessions/<id>/forks/critic-N/rounds/
//
// Returns the populated *Session. Fails if the folder already exists.
func NewSession(stateDirAbs string, forkCount int, now time.Time) (*Session, error)

// AtomicWrite writes data to <relative-path-under-Root> via temp file + rename.
// Sets perm 0644. Fsyncs the file before rename and the parent dir after.
func (s *Session) AtomicWrite(rel string, data []byte) error

// AppendLine appends a single line (data + "\n") to <relative-path>.
// Opens with O_APPEND|O_CREATE|O_WRONLY, perm 0644. Does NOT fsync (see fsync policy).
func (s *Session) AppendLine(rel string, data []byte) error

// Path returns the absolute path for a relative one (joinclean).
func (s *Session) Path(rel string) string
```

`AtomicWrite` is for files written exactly once per session (`start.json`, `end.json`, `summary.md`, `proposer-state.json`, `diff.patch`, individual round files). `AppendLine` is for the three append-only files (`transcript.jsonl`, `attacks.jsonl`, the cross-session `log.jsonl`).

`log.jsonl` lives at `<state-dir>/log.jsonl`, not under the session folder. It is shared across sessions.

## Atomic-write algorithm

1. `tmp = rel + ".tmp." + rand6`.
2. Write to `tmp` with `os.O_WRONLY|O_CREATE|O_EXCL`, perm 0644.
3. `f.Sync()`, then `f.Close()`.
4. `os.Rename(tmp, rel)`.
5. `parentDir.Sync()`.

Never `Truncate` an existing file in place; always rename in. Reason: a partial write through a power-loss / SIGKILL must leave the prior version intact (or absent) so [21](21-signals.md)'s recovery sees a coherent state.

## Append-only invariants

- Every line is a single JSON object terminated by exactly one `\n`.
- The orchestrator never seeks back, never rewrites a previous record.
- Lines are at most 64KB; longer entries spill to disk and the JSONL line carries a `body_path` reference (see [12](12-attacks-ledger.md)).
- A killed process leaves a valid prefix: every complete line is recoverable, the (possibly partial) trailing line is discarded by readers.

## fsync policy

| Operation | fsync? |
|---|---|
| `start.json` (atomic write) | yes (data + parent dir) |
| `end.json` (atomic write at termination) | yes |
| `summary.md` (atomic write at termination) | yes |
| Per-fork files (`proposer-state.json`, `diff.patch`, round files) | yes |
| Append-only files (`attacks.jsonl`, `transcript.jsonl`, `log.jsonl`) | no per-line; fsync once at termination via `state.Close()` |

Rationale: per-line fsync on append-only files would dominate latency for high-frequency rounds; correctness is bounded by the "interrupt loses at most the trailing line" invariant.

## Recovery on restart

v0 does not auto-recover (resume is v1, see [01-overview.md](01-overview.md) §Versioning). What the orchestrator *does* guarantee:

- A session whose `end.json` exists is "complete."
- A session whose `end.json` is missing is "interrupted." The user can `cat <session>/forks/*/rounds/*.md` and `<session>/attacks.jsonl` directly.
- `log.jsonl`'s line for an interrupted run is *missing* (it's appended last); detection rule: any session-id under `sessions/` without a corresponding final-status line in `log.jsonl` is interrupted.

## Concurrency

A single `debate` process owns its session folder. Two concurrent `debate` invocations under the same `--state-dir` get distinct session ids (different `rand6`) and never share a folder. The cross-session `log.jsonl` is the only shared file; it uses `O_APPEND` + a single `write()` per line, which POSIX guarantees atomic for `< PIPE_BUF` (4096 bytes), well above any record this spec emits.

## Test contract

- Unit: `NewSession` creates the full skeleton and fails on a pre-existing folder.
- Unit: `AtomicWrite` survives a SIGKILL between step 2 and step 4 with the prior contents intact (simulated).
- Unit: `AppendLine` × 1000 in a tight loop produces a file with exactly 1000 lines.
- Unit: race test: two goroutines `AppendLine` to the same file → both lines present, no interleaving (`race -count=10`).
- Integration: a session folder created by [19](19-round-loop.md), interrupted at random points, leaves a valid prefix readable by [12](12-attacks-ledger.md)'s reader.

## Acceptance criteria

- [x] Session id format matches the regex `^[0-9]{8}T[0-9]{6}Z-[a-z0-9]{6}$`.
- [x] `AtomicWrite` is `O_EXCL`-protected against the temp-file collision case.
- [x] No code path under `internal/state/` calls `os.Truncate` or seeks within an existing file.
- [x] All three append-only files are JSONL with one record per line; checked by a `golden test` that diffs a fixture run.
