# Spec 08 - Working-tree diff and trivial-diff gate

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) §"Risks" → flow disruption for design intent.

**Depends on:** [02](02-go-module.md), [04](04-cli-flags.md).
**Consumed by:** [11](11-fork-artifacts.md), [17](17-claude-proposer.md), [18](18-critic-drivers.md), [19](19-round-loop.md).

## Scope

In: a `git diff` wrapper that produces the unified diff between two refs (or one ref and the working tree), the changed-line counter used by `--changed-lines-min`, and the trivial-diff exit-fast path.

Out: applying diffs (the proposer-clone does that via its own tools); per-fork diff capture (that's [11](11-fork-artifacts.md), which uses this spec's helpers).

## Public Go interfaces

```go
// internal/input/diff.go
package input

type DiffSpec struct {
    From string  // git ref or "HEAD"
    To   string  // git ref or "." (working tree)
    Cwd  string  // absolute path; runs git from here
}

type Diff struct {
    Patch        string  // unified-diff text (multi-file)
    ChangedLines int     // additions + deletions across all hunks
    Files        []string // paths touched (for audit)
}

// Compute runs `git diff` and returns a *Diff.
//
// Equivalents:
//   - From="HEAD", To="."        ->  git diff HEAD             (working tree vs HEAD)
//   - From="HEAD~1", To="HEAD"   ->  git diff HEAD~1 HEAD     (committed range)
//   - From="HEAD", To="HEAD"     ->  empty patch, 0 changed lines
//
// Includes untracked files (via `git ls-files --others --exclude-standard`
// piped through `git diff --no-index`) only when To == ".".
func Compute(ctx context.Context, s DiffSpec) (*Diff, error)

// Trivial returns true iff d.ChangedLines < threshold.
func Trivial(d *Diff, threshold int) bool
```

Errors:

- `ErrNotGitRepo` when `git rev-parse --is-inside-work-tree` fails.
- `ErrGit` wrapping the underlying exec error with stderr captured.

## Behavior

- `git` is invoked via `os/exec`; absolute path resolved via `exec.LookPath`. No shell.
- Environment: `GIT_OPTIONAL_LOCKS=0`, `LC_ALL=C` to keep output stable; HOME inherited for credentials.
- For `To == "."`:
  - First call `git diff --no-color HEAD` for tracked changes.
  - Then for each untracked file from `git ls-files --others --exclude-standard`, append a synthetic `git diff --no-index /dev/null <file>` block. This matches what a reviewer would see if the proposer had `git add`-ed the new files.
- `ChangedLines` counts lines starting with `+` or `-` in hunk bodies, excluding the `+++ ` / `--- ` headers. (Standard unified-diff convention.)
- `Files` is parsed from `+++ b/<path>` and `--- a/<path>` headers, deduplicated.

## Trivial-diff exit-fast path

The orchestrator calls `Compute` immediately after pre-flight and before opening any session folder. If `Trivial(d, flags.ChangedLinesMin)`:

1. Append a single line to `<state-dir>/log.jsonl` (see [10](10-run-artifacts.md)) of shape:
   ```json
   {"ts":"<RFC3339>","skipped":"trivial-diff","changed_lines":<n>,"threshold":<m>}
   ```
2. Print to stderr (always, not gated on verbose):
   ```
   [agon] skipped: trivial diff (<n> changed lines < <m> threshold)
   ```
3. Exit 0. (Under `--hook-mode` also 0; the gate fires before the unresolved-leaves logic.)

Total wall time on the trivial path must be `< 100ms` excluding `git diff`. Measured in [25](25-probes.md)'s perf probe.

## --diff-from / --diff-to interpretation

| `--diff-from` | `--diff-to` | semantics |
|---|---|---|
| `HEAD` | `.` | tracked changes in working tree + untracked files |
| `HEAD~1` | `HEAD` | the last commit's net change |
| `<branch>` | `HEAD` | changes since branch point |
| `<sha>` | `<sha>` | empty if equal, else the range |

When `--session-id` is set and the user did not pass `--diff-from`, the default `HEAD` is correct (the proposer just wrote code into the working tree). When `--session-id` is empty (manual mode against a committed range), the user is expected to set `--diff-from`/`--diff-to` explicitly.

## Test contract

- Unit: empty diff → `ChangedLines == 0`, trivial gate trips.
- Unit: known fixture diff → `ChangedLines` matches `git diff --shortstat` count.
- Unit: untracked file in `To == "."` mode shows up in the patch.
- Unit: diff with binary file → patch contains the `Binary files differ` line, `ChangedLines == 0` for that file.
- Unit: `ErrNotGitRepo` returned outside a git repo.
- Integration: `make build && cd <fixture-repo> && bin/agon --changed-lines-min 100` exits 0 with the trivial-diff stderr line.

## Acceptance criteria

- [x] `Compute` returns a non-nil `*Diff` for every valid input shape.
- [x] Trivial gate exits in `< 100ms` on a small diff, verified by [25](25-probes.md).
- [x] No git command is spawned with a shell; no `bash -c` anywhere in this package.
- [x] Stderr line format and `log.jsonl` shape match the strings above (golden test).
