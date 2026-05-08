# Spec 29 - Probe G5 outcome: signal latency

> **Status: ✅ implemented** (G5 PASS at sub-50 ms after fixing a wired-but-unused signal handler in `cmd/debate/main.go`. The release-blocker gate this spec closed was retracted in the 2026-05-08 simplification of [27](27-release.md); the signal-handler fix stands and is regression-pinned by `e2e/cli/signal_test.go`.)
> Implementation spec for `debate`. See [21-signals.md](21-signals.md) for the underlying invariant and [25-probes.md](25-probes.md) for the probe.

**Depends on:** [21](21-signals.md), [25](25-probes.md).
**Consumed by:** [27](27-release.md).

## Scope

In: a recorded execution of `scripts/probes/signal-latency.sh` on the release-candidate host, and the disposition for the release-cut gate.

Out: changes to the signal-handling code ([21](21-signals.md)) or to the probe itself ([25](25-probes.md)).

## What we're proving

[21-signals.md](21-signals.md) commits to "SIGINT to a stuck child causes the orchestrator to exit in under 5 seconds." The probe spawns `bin/debate` against a sleep-only mock and times the SIGINT-to-exit interval. The release-cut gate is binary: PASS if the measured interval is <5s; FAIL otherwise.

## Execution

1. Run `make build` first - the probe drives the release-candidate binary, not the working-tree binary.
2. Run `scripts/probes/signal-latency.sh`.
3. Capture `wall_seconds`, exit code, and the probe's PASS/FAIL line.

## Recording format

```
probe: signal-latency
host_os: darwin|linux
binary_sha256: <sha of bin/debate at run time>
wall_seconds: <float>
verdict: PASS | FAIL
```

## Disposition

- **PASS:** [27-release.md](27-release.md) G5 outcome line cites the recording.
- **FAIL:** GA blocked. Investigate per [21-signals.md](21-signals.md) §"Process-tree teardown"; the regression is in code, not in the probe. Re-run after fix.

A flaky FAIL (one failure, one PASS) does not unblock GA - the probe must pass on three consecutive runs to count as PASS, recorded in the same block (`runs: 3/3`). This catches a slow-CI-host false negative before it ships.

## Acceptance criteria

- [x] Probe ran on the release-candidate host; outcome recorded.
- [x] If FAIL, root-cause and fix landed; probe re-run to PASS. (See "Out-of-scope-but-fixed" below.)
- [x] Regression test added (`e2e/cli/signal_test.go`) so the wiring cannot silently regress again.
- [x] ~~[27-release.md](27-release.md) G5 cites the recording.~~ *(retracted: G5 no longer exists as a release blocker; regression covered by `e2e/cli/signal_test.go`.)*

## Out-of-scope-but-fixed

The original probe never exercised `bin/debate` and contained a bash
trap-vs-sleep race that made it always FAIL. Rewriting it to actually
spawn `bin/debate` against a sleep-forever shim surfaced a real spec-21
violation: `cmd/debate/main.go` was using cobra's default
`context.Background()`, so the signal handler in
`internal/round/signals.go` was dead code. Fixing both required:

- `cmd/debate/main.go` calls `round.InstallHandler` and passes the
  signal-aware context via `root.ExecuteContext(ctx)`.
- `scripts/probes/signal-latency.sh` is rewritten to spawn `bin/debate`
  with shell-shim claude/codex that sleep, then signal and time exit.
- `e2e/cli/signal_test.go` is the regression-pinning unit at the e2e
  layer: it asserts SIGINT-to-exit < 2 s on a stuck-child run.

The "Out" section of the spec said "no signal-handling code changes"
- this turned out to be wrong because the missing wiring was the root
cause, not a probe bug. Treat that earlier scope-out as superseded by
this section.
