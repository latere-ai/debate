# Spec 11 — Per-fork artifacts

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Session persistence" → Layout for design intent.

**Depends on:** [04](04-cli-flags.md), [08](08-diff.md), [09](09-state-dir.md).
**Consumed by:** [17](17-claude-proposer.md), [18](18-critic-drivers.md), [19](19-round-loop.md), [22](22-contention-headline.md), [23](23-summary-render.md).

## Scope

In: schemas and write rules for the per-fork files inside `forks/critic-<i>/`: `proposer-state.json`, `diff.patch`, and the round files `rounds/r<n>-<role>.md`.

Out: the run-level files ([10](10-run-artifacts.md)), `attacks.jsonl` ([12](12-attacks-ledger.md)), the round-loop control flow ([19](19-round-loop.md)).

## Layout (recap)

```
sessions/<id>/forks/critic-<i>/
  proposer-state.json    # written once at fork creation, may be appended via field updates
  diff.patch             # snapshot of the working tree at THIS fork's R1 (refreshed per fork)
  rounds/
    r1-critic.md         # always odd-round files are critic outputs
    r2-proposer.md       # always even-round files are proposer outputs
    r3-critic.md
    ...
```

Files are atomic-write (see [09](09-state-dir.md)); none of them are append-only.

## proposer-state.json

Two shapes, discriminated by `agent`:

```jsonc
// Claude-as-proposer (v0)
{
  "schema":            "debate.proposer-state.v0",
  "agent":             "claude",
  "model":             "claude-sonnet-4-6",     // "" = CLI default
  "fork_session_id":   "5a8c9b1e-...",          // captured from --fork-session JSON, R1
  "root_session_id":   "f3d2a..."               // for audit; root is never resumed
}

// Codex-as-proposer (v1; documented for forward-compat, not written in v0)
{
  "schema":            "debate.proposer-state.v0",
  "agent":             "codex",
  "model":             "gpt-5",
  "round_thread_ids": [
    {"round": 2, "thread_id": "..."},
    {"round": 4, "thread_id": "..."}
  ]
}
```

Write rules:

- Written immediately after the call that produced the field returns ([01-overview.md](01-overview.md) Lifecycle invariants).
- For claude: written *once*, on R1's return; subsequent rounds reuse the captured `fork_session_id` and the file is read-only thereafter.
- For codex (v1): `round_thread_ids` grows by one element per even round; the file is overwritten via atomic write each time. Never append in place.

## diff.patch

The unified-diff text from [08](08-diff.md)'s `Compute`, written verbatim to `<fork>/diff.patch`. Refreshed per fork at fork start (after prior critics' concession-fixes have landed in the working tree); not modified mid-fork.

The orchestrator also writes the *initial* run-level diff to `<session>/diff.patch` once at R0 setup — separate file, same on-disk format.

## Round files

`rounds/r<n>-<role>.md` where `<role> ∈ {critic, proposer}`.

Convention: odd rounds are critic outputs, even rounds are proposer outputs. R1 is always `r1-critic.md`; R2 is always `r2-proposer.md`; etc. (See [19](19-round-loop.md) for round numbering.)

For critic rounds, the file content is the *normalized* output from [14](14-attack-parser.md) — not the raw critic stdout. Normalization (id reassignment, style-drop, reproduction-required filter) runs *before* the file is written, since this file is the one the proposer reads via `@<path>` pointer (see [01-overview.md](01-overview.md)).

For proposer rounds, the file content is the proposer-clone's chat response, captured from claude's JSON `result` field. Plus, on the same `AtomicWrite` call, a trailing block of working-tree-modification metadata appended by the orchestrator:

```markdown
<verbatim proposer chat response>

---
modified-files:
  - src/api.py
  - src/util.py
```

The `modified-files` block is computed by diffing the working tree against its state at R1 within the same fork (see [11.behavior.changed-files] below). It's purely informational; the proposer's actual concession-fix diff lives in the working tree.

### Critic-output post-normalize file format

The file the proposer reads. Markdown with one section per surviving attack:

```markdown
# Critic <i> — round <n> attacks

aspect: <aspect>

## c<i>-<seq> [<location>]

claim: <one-paragraph claim>

expected violation: <what the bug looks like at runtime>

reproduction:
<exact input/command/test that triggers it>

---

## c<i>-<seq+1> [<location>]
...
```

(The exact emitter rules and post-normalize format are owned by [14](14-attack-parser.md). This spec only specifies "what gets written here is the normalized form, not the raw stdout.")

### Critic re-attack rounds (R3, R5, ...)

Same format. Re-attacks reuse the original `attack_id`; new attacks pick up the next sequence in the critic's namespace. Withdrawn attacks are *omitted* from later round files (their absence is the withdrawal signal); status transitions to `withdrawn` in [12](12-attacks-ledger.md).

## Public Go interfaces

```go
// internal/state/fork.go
package state

func WriteProposerState(sess *Session, fork int, ps *ProposerState) error
func WriteForkDiff(sess *Session, fork int, patch string) error
func WriteRunDiff(sess *Session, patch string) error  // <session>/diff.patch
func WriteRound(sess *Session, fork, round int, role Role, body []byte) error

type Role int
const (
    RoleCritic   Role = iota  // odd rounds
    RoleProposer              // even rounds
)

// ChangedFilesAfter returns the file paths modified by the proposer-clone
// during a fork, computed as `git diff --name-only <fork-snapshot>..HEAD-or-WT`.
// Used to populate the `modified-files` block in proposer round files.
func ChangedFilesAfter(forkRoot string, since *Diff) ([]string, error)
```

`Role` constraint: `WriteRound(_, _, n, RoleCritic, _)` requires `n` odd; `RoleProposer` requires `n` even. Violations are programmer errors and panic in development builds; in release builds they return `ErrRoleParity`.

## Behavior

- Round files use `os.O_EXCL` via the underlying `AtomicWrite`; re-writing a round number panics ([19](19-round-loop.md) is responsible for monotonic round counting).
- The pointer message dispatched to the agent always references the *just-written* file path; [09](09-state-dir.md)'s atomic-write completes before pointer dispatch (file-before-pointer invariant).
- Concession-fix detection: the orchestrator captures `git status --porcelain` before each fork's R1, then again after each proposer round; the diff is the `modified-files` set.

## Test contract

- Unit: `WriteRound` rejects role/parity mismatches.
- Unit: `WriteRound` re-write of an existing round number returns the underlying `O_EXCL` error.
- Unit: `ChangedFilesAfter` returns the right file list for a synthetic working-tree mutation.
- Golden: a synthetic fork's per-file outputs match `testdata/golden/fork-*/`.

## Acceptance criteria

- [ ] All per-fork files are atomic-write only (no append-in-place).
- [ ] Round-file naming (`r<n>-<role>.md`) is enforced at the API boundary, not by callers.
- [ ] `proposer-state.json` for claude is written exactly once per fork.
- [ ] Critic round files contain the normalized output from [14](14-attack-parser.md), not raw stdout.
