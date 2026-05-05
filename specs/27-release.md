# Spec 27 — v0 release process and GA gates

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"v0 release blockers" for design intent.

**Depends on:** [02](02-go-module.md), [03](03-ci-lint-release.md), [21](21-signals.md), [23](23-summary-render.md), [24](24-stop-hook.md), [25](25-probes.md), [26](26-tests.md).
**Consumed by:** users.

## Scope

In: the v0 release-cut checklist (every gate that must be cleared before tagging `v0.0.1`), the tag/build/sign/publish flow, install instructions for end users, and the post-release rollback path.

Out: feature work; v1 planning.

## v0 GA gate checklist

Each item must be checked off before tagging `v0.0.1` (i.e. the first non-`-rcN` tag).

### Upstream research gates ([01-overview.md](01-overview.md) §"Relationship to upstream research")

- [ ] **G1.** Upstream `agents-byzantine-tolerance` spec 07a has run per-aspect for at least four candidate aspects.
- [ ] **G2.** Critic-found-bug rate ≥ 60% on at least two aspects. Record which two.
- [ ] **G3.** Aspects below threshold are removed from the [05](05-config-file.md) defaults *before* tagging. Every default aspect in `[05]` must satisfy G2.

### Probe gates ([25](25-probes.md))

- [ ] **G4.** `scripts/probes/no-output-stop-hook.sh` has run against the current local claude install. Outcome (PASS or FAIL) recorded in `release-notes-v0.0.1.md`. The spec wording at [01-overview.md](01-overview.md) §"Lifecycle invariants" stays as-is in either branch.
- [ ] **G5.** `scripts/probes/signal-latency.sh` PASS (signal → exit < 5s).
- [ ] **G6.** `scripts/probes/trivial-diff-perf.sh` PASS (trivial gate < 100ms).
- [ ] **G7.** `scripts/probes/interactive-stdout.sh` outcome recorded (PASS, FAIL, or SKIP). Determines whether the README's "stdout best-effort" caveat can be relaxed.

### Code & test gates

- [ ] **G8.** All CI green on `main` for the most recent five commits.
- [ ] **G9.** `make all` green on macOS and Linux from a fresh clone.
- [ ] **G10.** `go test -race ./...` green.
- [ ] **G11.** `golangci-lint run` clean.
- [ ] **G12.** `goreleaser check` clean.
- [ ] **G13.** Real-e2e workflow run at least once on the release-candidate tag with `RUN_REAL=1`. Record outcomes in `release-notes-v0.0.1.md`.

### UX gates

- [ ] **G14.** `debate --version`, `debate --help` print sane output from the release tarball binary (Linux + Darwin × amd64 + arm64).
- [ ] **G15.** `debate install-hook --scope user` against a fresh `~/.claude/settings.json` produces a valid verbose-format Stop hook entry. Verified manually.
- [ ] **G16.** A real claude session followed by a 47-line diff triggers the hook, runs through to summary on disk, and exits cleanly. Wall time ≤ 5 minutes per fork on default `--max-turn 6`.

### Documentation gates

- [ ] **G17.** README's install + usage section reflects the binary layout and `install-hook` subcommand.
- [ ] **G18.** Every implementation spec from [02](02-go-module.md) through [26](26-tests.md) has Acceptance criteria all checked off in the spec itself.

## Release flow

```
1. Cut release branch
   git checkout -b release/v0.0.1

2. Tag a candidate
   git tag v0.0.1-rc1
   git push origin v0.0.1-rc1
   # release.yml triggers; goreleaser builds the candidate as a draft.

3. Run real-e2e + probes against the rc binary
   make probe                            # G4-G7
   RUN_REAL=1 go test -tags real_e2e ./e2e/real/...
   # Iterate as needed: rc2, rc3, ...

4. Once all gates green, tag GA
   git checkout main
   git tag v0.0.1
   git push origin v0.0.1
   # release.yml builds GA, attaches release notes.

5. Promote draft release to public
   gh release edit v0.0.1 --draft=false
```

## Release artifacts

Per [03](03-ci-lint-release.md):

- `debate_<version>_linux_amd64.tar.gz`
- `debate_<version>_linux_arm64.tar.gz`
- `debate_<version>_darwin_amd64.tar.gz`
- `debate_<version>_darwin_arm64.tar.gz`
- `checksums.txt` (sha256 of all archives)

Each archive contains:

```
debate                       # the binary
LICENSE
README.md
debate-stop-hook.sh
```

`README.md` in the archive is the same as the repo README; install instructions reference the `install-hook` subcommand.

## Release notes template

`release-notes-v0.0.1.md`:

```markdown
# debate v0.0.1

First v0 release. Productizes the adversarial-debate architecture from
spec 07 of agents-byzantine-tolerance.

## Probe outcomes

- no-output-stop-hook: <PASS/FAIL>
- signal-latency: <PASS/FAIL>
- trivial-diff-perf: <PASS/FAIL>
- interactive-stdout: <PASS/FAIL/SKIP>

## Default aspects (this release)

<list aspects whose 07a per-aspect rate is ≥ 60%>

## Known limitations

- Codex-as-proposer is v1 (architecture documented in spec 01).
- No live in-session UI; summary on disk; stdout best-effort under the Stop hook.
- Critics are best-effort isolated (aspect prompt + codex --sandbox read-only); strict per-fork sandbox is v1.
- `--judge llm/human` not in v0.

## Upgrade / install

curl -L https://github.com/latere-ai/debate/releases/download/v0.0.1/debate_v0.0.1_<os>_<arch>.tar.gz | tar xz
./debate install-hook --scope user
```

## Rollback

If post-release a critical bug surfaces:

1. Mark the GitHub release as `Pre-release` with a deprecation note.
2. Push a `v0.0.1.x` patch tag with the fix.
3. The verbose-hook format means rolling back the Stop hook is `debate uninstall-hook --scope user` (or manual edit of `settings.json`); no claude restart needed.

`debate` writes nothing to claude's session files in normal operation, so a rollback does not corrupt prior `.debate/sessions/` data.

## Test contract

- A dry-run of the release script against a `v0.0.1-rc-test` tag produces the four archives.
- `goreleaser release --snapshot --clean` succeeds locally.
- `release-notes-v0.0.1.md` is the literal file in the repo for the GA tag, with all probe outcomes filled in.

## Acceptance criteria

- [ ] All G1–G18 gates have a recorded outcome at GA time.
- [ ] Release archives ship with `debate-stop-hook.sh`.
- [ ] `release-notes-v0.0.1.md` is committed before the GA tag and references the probe outcomes by name.
- [ ] Rollback path documented and tested at least once on an `-rc` tag.
