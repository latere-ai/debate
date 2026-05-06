# Spec 35 - Release-notes channel for probe & gate outcomes

> **Status: ✅ implemented** (Option 1 chosen; spec 27 amended; `release-notes-v0.0.1.md` skeleton committed).
> Implementation spec for `debate`. Resolves an internal contradiction in [27-release.md](27-release.md). Consumed by every other follow-up that records a recording block (28, 29, 30, 31, 32, 33, 34).

**Depends on:** [27](27-release.md).
**Consumed by:** [27](27-release.md), [28](28-probe-no-output-stop-hook-outcome.md), [29](29-probe-signal-latency-outcome.md), [30](30-probe-trivial-diff-perf-outcome.md), [31](31-probe-interactive-stdout-outcome.md), [32](32-real-e2e-suite.md), [33](33-install-hook-smoke.md), [34](34-real-claude-end-to-end-smoke.md).

## The contradiction

[27-release.md](27-release.md) §"Release notes" says:

> There is **no** static `release-notes-vX.md` file in the repo: that would duplicate state and diverge.
> Probe outcomes (see [25](25-probes.md)) are recorded by appending a short note to the GitHub release body via the GH UI or `gh release edit` after `make probe` finishes - they are not committed.

[27-release.md](27-release.md) §"Acceptance criteria" says:

> - [x] `release-notes-v0.0.1.md` is committed before the GA tag and references the probe outcomes by name.

These are mutually exclusive. The G4 gate ("recorded in `release-notes-v0.0.1.md`") and similar-shaped wording in this spec series ([28](28-probe-no-output-stop-hook-outcome.md)–[34](34-real-claude-end-to-end-smoke.md)) need a single source of truth.

## Decision

Pick **one** channel and use it consistently:

### Option 1: in-repo file (`release-notes-v0.0.1.md`)

Pros:
- Survives a release page edit; lives under git history.
- Searchable from a clone; no GH login needed.
- All cross-spec citations (`see release-notes-v0.0.1.md → no-output-stop-hook`) resolve via `grep`.

Cons:
- Per-release file proliferates. v0.0.2, v0.0.3, ... each get one.
- "Duplicates state" critique from spec 27 body: goreleaser already produces a changelog; adding a parallel notes file means two places to keep in sync.

### Option 2: GitHub release body (via `gh release edit`)

Pros:
- Single source of truth lives next to the artifacts users download.
- No file duplication; `release.header`/`release.footer` in `.goreleaser.yaml` already structure it.

Cons:
- Mutable; nothing in git history captures the appended outcomes.
- Cross-spec citations cannot resolve from a clone offline.
- A future maintainer who edits the release body has no review trail.

### Recommendation

**Option 1: in-repo file.** Probe outcomes are part of the audit trail this project is built around (see [01-overview.md](01-overview.md) §"Lifecycle invariants" - the whole point is on-disk evidence). Putting them in a mutable GH release body contradicts that posture. The "duplicates state" objection is weaker than the audit-trail need: goreleaser's auto-generated changelog covers code changes, while `release-notes-v0.0.1.md` covers release-cut evidence (probe outcomes, real-claude smoke, signal latency wall time). They cover different things.

If the maintainer prefers Option 2 anyway, that is acceptable provided every cross-spec citation in [28](28-probe-no-output-stop-hook-outcome.md)–[34](34-real-claude-end-to-end-smoke.md) is updated to read "GH release body of v0.0.1" instead of "release-notes-v0.0.1.md".

## File structure (Option 1)

`release-notes-v0.0.1.md` at repo root:

```markdown
# debate v0.0.1

Release-cut evidence. See specs/27-release.md for the gate process.

## Probe outcomes

### no-output-stop-hook (G4)
<verbatim recording block from spec 28>

### signal-latency (G5)
<verbatim recording block from spec 29>

### trivial-diff-perf (G6)
<verbatim recording block from spec 30>

### interactive-stdout (G7)
<verbatim recording block from spec 31>

## Test gates

### real-e2e (G13)
<recording from spec 32; or "retracted - see specs/32-real-e2e-suite.md">

## UX gates

### install-hook smoke (G15)
<verbatim recording block from spec 33>

### real-claude end-to-end (G16)
<verbatim recording block from spec 34>
```

`.gitignore` does not need updating; the file is intentionally committed.

## Spec 27 amendments

After this spec lands:

- [27-release.md](27-release.md) §"Release notes" body keeps the goreleaser-generated changelog paragraph but the contradicting "no static file" sentence is replaced by:
  > Probe and gate outcomes (G4–G7, G13, G15, G16) are committed to `release-notes-v0.0.1.md` alongside the tag. The auto-generated GitHub release body covers code changes; the in-repo notes file covers release-cut evidence.
- [27-release.md](27-release.md) §"Test contract" line about `release-notes-v0.0.1.md` stays as-is.
- [27-release.md](27-release.md) §"Acceptance criteria" line stays as-is.

(Or, if Option 2 wins: the spec 27 body keeps its current "no static file" wording and the acceptance line is rewritten to point at the GH release body, *and* every recording-citing follow-up spec is mass-edited.)

## Decision

**Option 1: in-repo file `release-notes-v0.0.1.md`.** Audit-trail posture trumps file-count cleanliness. Goreleaser still owns the auto-generated changelog (code changes between tags); the in-repo notes file owns release-cut evidence. Both can coexist - they cover different things.

## Acceptance criteria

- [x] Option 1 selected (this section).
- [x] [27-release.md](27-release.md) §"Release notes" edited to remove the contradiction.
- [x] `release-notes-v0.0.1.md` skeleton committed; follow-up specs ([28](28-probe-no-output-stop-hook-outcome.md)–[34](34-real-claude-end-to-end-smoke.md)) write into it.
