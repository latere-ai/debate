# Spec 26 — Test strategy and mock harnesses

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) for design intent.

**Depends on:** every implementation spec from [04](04-cli-flags.md) through [24](24-stop-hook.md).
**Consumed by:** [27](27-release.md).

## Scope

In: the test layering (unit / integration / e2e), mock claude and codex harnesses, fixture directory layout, golden-file conventions, the contract for what each spec's tests cover.

Out: the GA gate checklist ([27](27-release.md)), probes against real installs ([25](25-probes.md)).

## Layering

| Layer | Where | What | When it runs |
|---|---|---|---|
| Unit | `internal/<pkg>/*_test.go` | Pure functions, parsers, schemas, error mappings. | Every commit; CI required. |
| Integration | `internal/<pkg>/*_integration_test.go` | Multi-package paths with real filesystem, `state.Session`, mock subprocesses. | Every commit; CI required. |
| E2E (mock) | `e2e/mock/*_test.go` | Whole-binary runs with `bin/debate` against the mock harness as `claude`/`codex`. | Every commit; CI required. |
| E2E (real) | `e2e/real/*_test.go` | Whole-binary runs against real `claude` and `codex`; gated `RUN_REAL=1`. | Local + nightly; not on PRs. |
| Probes | `scripts/probes/` | See [25](25-probes.md). | Pre-release. |

Build tags: integration tests use `//go:build integration` plus the file-name convention; CI sets `-tags integration`. Real-e2e uses `//go:build real_e2e`.

## Mock claude harness

`testdata/bin/mock-claude` (a Go binary built into the test cache):

```go
// e2e/mock/claudemock/main.go
package main

// mock-claude reads its CLI args, looks up a scripted response, and emits
// the JSON shape [17] expects.
//
// Behavior driven by a fixture file at $MOCK_CLAUDE_SCRIPT (a JSON list
// of {"args_glob": "...", "stdout": {...}, "stderr": "...", "exit": 0}
// matchers) — first-match wins.
```

Capabilities (because [17](17-claude-proposer.md) and [18](18-critic-drivers.md) drive `claude` in three modes):

- `--resume <id> --fork-session -p <pointer>` → emits a `result` with a fresh `session_id` and a scripted response.
- `--resume <fork-id> -p <pointer>` → emits a `result` with the same `session_id` and a scripted response.
- `-p <prompt>` (no resume) → emits a `result` with a fresh `session_id`.

Scripts live at `testdata/scripts/claude/<scenario>.json`.

## Mock codex harness

`testdata/bin/mock-codex` (also Go):

```go
// e2e/mock/codexmock/main.go
package main

// mock-codex emits the JSON event stream [18] expects:
//   {"type":"thread.started","thread_id":"<uuid>"}
//   {"type":"item.completed","item":{"type":"agent_message","content":"..."}}
//   {"type":"thread.completed"}
//
// Driven by $MOCK_CODEX_SCRIPT.
```

## Fixtures

```
testdata/
  bin/                              # mock binaries built by tests
  scripts/
    claude/
      proposer-r1-r2.json           # one-fork sequence
      proposer-r1-r2-r3-r4.json     # cross-examination
      proposer-401.json             # auth error
      ...
    codex/
      critic-2attacks-clean.json
      critic-style-only.json        # for [14] F2 drop test
      critic-malformed.json
      ...
  transcripts/
    minimal.jsonl                   # first user turn extractable
    multipart.jsonl
    no-user-turn.jsonl
  diffs/
    trivial.patch                   # 3 changed lines
    typical.patch                   # 47 changed lines
  golden/
    summary/
      clean.md                      # 0 unresolved
      one-unresolved.md
      two-unresolved-tied.md        # tie-break check
    prompts/
      functional-logic-r1.txt
      security-r1.txt
      ...
    run-artifacts/
      start.json
      end.json
      ...
```

## Golden-file convention

- Format: exact bytes the renderer produces.
- Update path: failing test prints a unified diff against the expected file. To regenerate: `UPDATE_GOLDEN=1 go test ./...`. CI fails if a regenerate-style update lands in a PR (lint check).
- Per-spec ownership: each spec lists the golden files it covers; the test names match the spec section names.

## Per-spec test coverage map

| Spec | Required test types |
|---|---|
| [02](02-go-module.md) | None beyond `go build`/`go test` toolchain checks. |
| [03](03-ci-lint-release.md) | CI workflow validates by running. |
| [04](04-cli-flags.md) | Unit (each flag round-trip + env override). |
| [05](05-config-file.md) | Unit (layering matrix). |
| [06](06-preflight.md) | Unit (each exit code path) + integration (recursion guard). |
| [07](07-claude-transcript.md) | Unit (path encode/decode, parse, edge cases) + fixture transcripts. |
| [08](08-diff.md) | Unit (computed line counts) + integration (real git fixture). |
| [09](09-state-dir.md) | Unit (atomic-write semantics) + race test (`-race -count=10`). |
| [10](10-run-artifacts.md) | Unit (round-trip schemas) + golden. |
| [11](11-fork-artifacts.md) | Unit (parity check, atomic-write) + integration. |
| [12](12-attacks-ledger.md) | Unit (state machine round-trip, body spill) + truncated-line tolerance test. |
| [13](13-critic-output-format.md) | Fixture (worked example parses without warnings). |
| [14](14-attack-parser.md) | Fuzz + table-driven (each filter rule, each disposition). |
| [15](15-aspect-prompts.md) | Golden (per-aspect) + cross-aspect filter test. |
| [16](16-subprocess-infra.md) | Unit (env scrub, JSON sanitize, cancellation latency). |
| [17](17-claude-proposer.md) | Mock-driven unit + real-e2e gated by `RUN_REAL=1`. |
| [18](18-critic-drivers.md) | Mock-driven unit + recursion-guard probe. |
| [19](19-round-loop.md) | Mock-driven integration (full fork sequences). |
| [20](20-termination.md) | Unit (each detector) + integration (run-level termination). |
| [21](21-signals.md) | Unit (handler) + integration (process-tree teardown). |
| [22](22-contention-headline.md) | Unit + property (determinism). |
| [23](23-summary-render.md) | Golden (per termination shape). |
| [24](24-stop-hook.md) | Unit (settings merge) + integration (script invocation). |
| [25](25-probes.md) | Smoke against mock binaries; real-claude probes manual. |

## CI integration

`.github/workflows/ci.yml` ([03](03-ci-lint-release.md)) runs:

```
go test -race -tags integration ./...
```

E2E mock tests are inside `internal/...` and `e2e/mock/...` — both run by default. Real-e2e (`//go:build real_e2e`) is opted in via a separate workflow gated on a workflow-dispatch event.

## Test contract

- All tests must run in under 60 seconds total on CI hardware (excluding real-e2e).
- No test reaches out to the network (real-e2e is the exception, gated).
- No test depends on environment variables beyond `RUN_REAL`, `UPDATE_GOLDEN`, `KEEP`.
- Race detector is on for the entire suite.

## Acceptance criteria

- [x] All four golden subdirectories present with at least one fixture each.
- [x] Mock harnesses build via `go test`'s build cache; no manual `make`.
- [x] CI runs `-race`; suite green.
- [x] Each spec from 02–25 has a `*_test.go` file in its package implementing its required test types (audit-via-grep).
- [x] Real-e2e workflow exists but is opt-in; no PR is blocked on it.
