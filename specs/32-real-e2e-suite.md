# Spec 32 - Real-e2e suite (write or retract)

> **Status: ✅ implemented** (Path A; `e2e/real/full_test.go` lives behind `//go:build real_e2e`. The release-blocker gate G13 this spec closed was retracted in the 2026-05-08 simplification of [27](27-release.md); the suite remains as opt-in `workflow_dispatch` coverage.)
> Implementation spec for `agon`. Resolves the gap between [.github/workflows/real-e2e.yml](../.github/workflows/real-e2e.yml), [26-tests.md](26-tests.md), and the actual file tree.

**Depends on:** [16](16-subprocess-infra.md), [17](17-claude-proposer.md), [18](18-critic-drivers.md), [19](19-round-loop.md), [26](26-tests.md).
**Consumed by:** [27](27-release.md), [.github/workflows/real-e2e.yml](../.github/workflows/real-e2e.yml).

## The gap

`.github/workflows/real-e2e.yml` runs `go test -tags real_e2e ./e2e/real/...`. [26-tests.md](26-tests.md) commits to the same path. **`e2e/real/` does not exist** and no test file carries the `real_e2e` build tag. The workflow is unrunnable as written; G13 in [27-release.md](27-release.md) cannot be cleared.

The decision below picks one of two paths. Both close the gap; they differ in confidence and cost.

## Path A: write a minimal real-e2e suite

Scope: one test file, `e2e/real/full_test.go`, build-tagged `real_e2e`, gated on `RUN_REAL=1` env. It exercises one fork end-to-end against real `claude` and real `codex` and asserts the same shape the mock suite asserts.

Test contract:

- `TestRealEndToEnd_OneFork` (only test in the file, deliberately):
  - Initialises a temp git repo with a 47-line diff fixture (reuse `testdata/diffs/typical.patch` if present).
  - Spawns `bin/agon` with `--main claude --side codex --max-turn 4 --aspects security`.
  - Asserts: process exits within 5 minutes; `summary.md` exists and is non-empty; `attacks.jsonl` parses; at least one `forks/critic-*/rounds/r1-critic.md` exists.
  - Skips with `t.Skip` if `claude` or `codex` are missing on `$PATH`, *not* if `RUN_REAL` is unset (the build tag already gates that).

Scope explicitly excludes:

- Multi-fork timing tests - too noisy with real models.
- Output-content assertions - real models drift; mock parity covers shape.
- Cost-cap or termination-condition assertions - covered by mock e2e in `e2e/cli/full_test.go`.

Adds: one file, no new dependencies. The mock suite still owns shape coverage; this test only proves "the wires don't crackle when a real LLM is on the other end."

## Path B: retract the gate

Scope: edit [26-tests.md](26-tests.md) to remove the "E2E (real)" row and the `real_e2e` build tag column; edit [27-release.md](27-release.md) to drop G13; delete `.github/workflows/real-e2e.yml`. The argument: the mock harness is already shape-faithful (mocks are derived from the actual JSON shapes) and the probes (G4–G7) cover the environment-dependent claims.

Adds nothing; removes infrastructure that isn't used.

## Recommendation

Path A. The cost is one Go file (~80 lines) and one workflow run before GA. The benefit is a real "does this break against the live LLM CLI" smoke that no other test gives. Path B trades a real check for documentation cleanup; the spec already paid for the workflow file.

If maintainer time is the binding constraint at GA, Path B is acceptable but requires explicit deletion of the unrunnable workflow - leaving the workflow file in place with no targets is the worst outcome.

## Acceptance criteria

- [x] Maintainer chooses Path A or Path B in this spec's "Decision" section.
- [x] If Path A: `e2e/real/full_test.go` exists with the build tag and skip logic; `RUN_REAL=1 go test -tags real_e2e ./e2e/real/...` exits 0 once on the maintainer's machine.
- [ ] If Path B: [26-tests.md](26-tests.md) and [27-release.md](27-release.md) updated; `.github/workflows/real-e2e.yml` removed. *(N/A: Path A chosen)*
- [x] ~~[27-release.md](27-release.md) G13 either cites the recording (Path A) or is removed (Path B).~~ *(retracted: G13 no longer exists as a release blocker; the test stays as opt-in coverage.)*

## Decision

**Path A**. The real-e2e workflow file already existed and pointed at a missing target; writing one shape-only test to make the workflow runnable costs less than excising spec text in three files. The test skips when claude or codex are missing on `$PATH`, so it is non-blocking for non-maintainer environments. CI does not auto-run this path; only `workflow_dispatch` does, per [`.github/workflows/real-e2e.yml`](../.github/workflows/real-e2e.yml).
