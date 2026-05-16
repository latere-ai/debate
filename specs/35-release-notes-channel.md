# Spec 35 - Release-notes channel for probe & gate outcomes

> **Status: ❌ retracted (2026-05-08)** - The premise of this spec (a separate "release-notes channel" carrying probe/gate evidence) was abandoned in favour of the standard goreleaser flow used by sibling repos (e.g. `latere-cli`): tag → CI → goreleaser publishes binaries + auto-generated changelog from the commit log. No separate release-notes file or release-evidence asset is shipped. Probes under `scripts/probes/` remain as developer tooling; their outcomes are not artifacts. Spec [27](27-release.md) was simplified in the same pass to drop the G1-G18 release-blocker checklist down to a short pre-tag sanity list.
>
> Original content (Option 1 vs Option 2 agon, the 2026-05-08 amendment introducing Option 3, and the file-structure / spec 27 amendments / acceptance criteria sections) is preserved below as historical record. Nothing below this banner is active.

---

## Original (retracted) content

## Amendment (2026-05-08)

Option 1 was implemented but does not match the natural release workflow:

> 1. Tag a new release.
> 2. CI triggers an execution: builds binaries, attaches them to the GitHub release, generates release notes from commits.
> 3. The release page is the artifact users land on.

Option 1 inserts a manual step *before* the tag: a human runs probes locally, pastes the recordings into `release-notes-v0.0.1.md`, commits, then tags. That coupling is what the user pushed back on. Option 2 (mutable GH release body) is closer to the workflow but loses the audit-trail posture spec 35 was protecting.

Neither original option is right. A third option - **release-asset bundle** - was missed.

### Option 3: release-asset bundle

Probes emit their recordings into a build-output directory (e.g. `dist/release-evidence/`); CI (or a human running `make probe` followed by `gh release upload`) attaches that directory as a `.tar.gz` asset on the GitHub release alongside the binaries.

Pros:
- **Matches the user-expected workflow**: tag → CI → release with binaries + evidence bundle. No pre-tag commit dance.
- **Audit-trail preserved**: GitHub release assets are immutable once uploaded (unlike the release body). `gh release download v0.0.1 -p evidence.tar.gz` fetches the exact bytes that shipped.
- **No per-version file proliferation**: nothing accumulates in the repo tree across `v0.0.2`, `v0.0.3`, ... .
- **Cross-spec citations resolve**: follow-up specs cite "release asset `evidence.tar.gz` of `<tag>`" and the bytes are reachable offline via `gh release download`.
- **Goreleaser integration is one-line**: `extra_files` under the release block, or `archives.files` if bundled into the per-arch tarballs (preferred: separate asset, since evidence is platform-independent).

Cons:
- Probes that need a real claude install (G4, G16) still cannot run in CI - they have to be run locally and the bundle uploaded by hand. But this is true under Option 1 too; the only change is the upload mechanism (`gh release upload` instead of `git commit`).
- Slightly less greppable from a clone: `gh release download` is one command but is not `grep -r`. Mitigated because the bundle is small and a maintainer can keep one extracted locally during a release cut.

### Why Option 3 wins

The audit-trail argument that picked Option 1 over Option 2 still stands; Option 3 satisfies it via release-asset immutability rather than git history. The user-workflow argument that broke Option 1 does not break Option 3, because evidence collection is decoupled from the tag/commit flow.

The acceptance criteria below replace the original Option 1 criteria. The original criteria are kept (struck through) for traceability.

**Depends on:** [27](27-release.md).
**Consumed by:** [27](27-release.md), [29](29-probe-signal-latency-outcome.md), [30](30-probe-trivial-diff-perf-outcome.md), [32](32-real-e2e-suite.md), [34](34-real-claude-end-to-end-smoke.md).

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

## File structure (Option 3, active)

Probes write their recording blocks under `dist/release-evidence/` (gitignored). At release time the directory is packaged into a single asset attached to the GitHub release:

```
dist/release-evidence/
├── README.md              # human-readable index with the same section layout as the
│                          # original Option 1 file (probe outcomes, test gates, UX gates)
├── probes/
│   ├── signal-latency.txt        # spec 29
│   └── trivial-diff-perf.txt     # spec 30
├── gates/
│   ├── real-e2e.txt              # spec 32 (or "retracted" marker)
│   └── real-claude-end-to-end.txt # spec 34
└── manifest.json          # {tag, host_os, claude_version, codex_version,
                           #  binary_sha256, generated_at, gates: [...]}
```

Packaged at release time as `release-evidence-<tag>.tar.gz` and uploaded via:

```sh
tar -czf dist/release-evidence-${TAG}.tar.gz -C dist release-evidence
gh release upload "${TAG}" dist/release-evidence-${TAG}.tar.gz
```

Or wired into `.goreleaser.yaml` via `release.extra_files`:

```yaml
release:
  extra_files:
    - glob: ./dist/release-evidence-*.tar.gz
```

`.gitignore` adds `dist/release-evidence/` and `dist/release-evidence-*.tar.gz`; nothing about evidence is committed.

### File structure (Option 1, superseded)

Original Option 1 design - kept here so prior commits citing `release-notes-v0.0.1.md` remain readable:

```markdown
# agon v0.0.1

Release-cut evidence. See specs/27-release.md for the gate process.

## Probe outcomes
### signal-latency (G5)              # spec 29
### trivial-diff-perf (G6)           # spec 30

## Test gates
### real-e2e (G13)                   # spec 32 (or retracted)

## UX gates
### real-claude end-to-end (G16)     # spec 34
```

The committed `release-notes-v0.0.1.md` at repo root is removed as part of the Option 3 transition.

## Spec 27 amendments

Required for Option 3:

- [27-release.md](27-release.md) §"Release notes" replaces the "committed to `release-notes-v<version>.md` alongside the tag" paragraph with:
  > Probe and gate outcomes (G4–G7, G13, G15, G16) ship as a release asset (`release-evidence-<tag>.tar.gz`) attached to the GitHub release. Probes write into `dist/release-evidence/` locally; the directory is packaged and uploaded at release time (via `gh release upload` or goreleaser `release.extra_files`). The auto-generated GitHub release body covers code changes; the asset bundle covers release-cut evidence. Audit-trail posture is preserved by GitHub release-asset immutability.
- [27-release.md](27-release.md) §"Release flow" step 3 ("Run real-e2e + probes") gains a sub-step: package `dist/release-evidence/` into the bundle and upload to the rc release.
- [27-release.md](27-release.md) §"Release artifacts" lists `release-evidence-<version>.tar.gz` alongside the four per-arch archives and `checksums.txt`.
- [27-release.md](27-release.md) §"Test contract" line referencing `release-notes-v0.0.1.md` is rewritten to reference the asset bundle: "`release-evidence-v0.0.1.tar.gz` is attached to the GA release with all probe outcomes inside."
- [27-release.md](27-release.md) §"Acceptance criteria" line about `release-notes-v0.0.1.md` is rewritten in the same way.
- Gate-row descriptions (G4, G13) that read "recorded in `release-notes-v0.0.1.md`" become "recorded in `dist/release-evidence/probes/<name>.txt`, shipped in the release asset bundle".

Follow-up specs [28](28-probe-no-output-stop-hook-outcome.md)–[34](34-real-claude-end-to-end-smoke.md) need a parallel pass: every "appended to `release-notes-v0.0.1.md`" instruction is rewritten to "written to `dist/release-evidence/<area>/<name>.txt`". The recording-block format is unchanged; only the destination path changes.

## Decision

**Option 3: release-asset bundle.** Tag → CI builds binaries + uploads evidence asset → release page is the single landing artifact. The audit-trail posture is preserved through GitHub release-asset immutability rather than git history. Goreleaser owns the auto-generated changelog (code changes between tags); the asset bundle owns release-cut evidence. Both ship together with the tag and neither pre-tag commits nor mutable release-body edits are needed.

## Acceptance criteria

Active (Option 3):

- [ ] Option 3 selected and Option 1 marked superseded (this amendment).
- [ ] [27-release.md](27-release.md) §"Release notes", §"Release flow", §"Release artifacts", §"Test contract", §"Acceptance criteria", and gate-row descriptions for G4/G13 are updated per [§Spec 27 amendments](#spec-27-amendments).
- [ ] Follow-up specs [28](28-probe-no-output-stop-hook-outcome.md)–[34](34-real-claude-end-to-end-smoke.md) are updated to write into `dist/release-evidence/<area>/<name>.txt` instead of appending to `release-notes-v0.0.1.md`.
- [ ] `dist/release-evidence/` and `dist/release-evidence-*.tar.gz` are added to `.gitignore`.
- [ ] `.goreleaser.yaml` `release.extra_files` (or an equivalent `gh release upload` step in `.github/workflows/release.yml`) attaches `release-evidence-<tag>.tar.gz` to the GitHub release.
- [ ] `release-notes-v0.0.1.md` is removed from the repo root; its current contents are migrated into `dist/release-evidence/` for the v0.0.1 release cut so no evidence is lost.

Superseded (Option 1, retained for traceability):

- [x] ~~Option 1 selected (this section).~~
- [x] ~~[27-release.md](27-release.md) §"Release notes" edited to remove the contradiction.~~
- [x] ~~`release-notes-v0.0.1.md` skeleton committed; follow-up specs ([28](28-probe-no-output-stop-hook-outcome.md)–[34](34-real-claude-end-to-end-smoke.md)) write into it.~~
