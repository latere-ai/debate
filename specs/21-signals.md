# Spec 21 — Signal handling and graceful exit

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Cancellable" and §"Termination conditions" → User interrupt.

**Depends on:** [09](09-state-dir.md), [10](10-run-artifacts.md), [16](16-subprocess-infra.md), [19](19-round-loop.md), [20](20-termination.md).
**Consumed by:** [23](23-summary-render.md), [27](27-release.md).

## Scope

In: SIGINT/SIGTERM handlers wired into the orchestrator's root `context.Context`, child-process tear-down, durable `end.json` write before exit, exit-code translation through `--hook-mode`.

Out: cost-cap detection ([20](20-termination.md)), summary rendering ([23](23-summary-render.md)).

## Public Go interfaces

```go
// internal/round/signals.go
package round

// InstallHandler returns a derived context that cancels on the named
// signals, plus a finalize function the caller defers.
//
// On signal:
//   1. ctx is cancelled. [16]'s subprocess teardown propagates SIGINT
//      to all spawned process groups.
//   2. The Engine's main loop observes ctx.Err() at the next round
//      boundary and breaks out with TerminationReason = "interrupted".
//   3. Engine writes end.json + appends a NON-"kind:run" line is NOT
//      written — log.jsonl absence is the interrupt signal (see [10]).
//   4. The finalize function the caller deferred restores the default
//      handlers; a second SIGINT during finalize triggers os.Exit(130)
//      to honor the user's "I really mean it" cadence.
func InstallHandler(ctx context.Context, sigs ...os.Signal) (context.Context, func())
```

Default `sigs`: `os.Interrupt`, `syscall.SIGTERM`. SIGHUP not handled in v0 (orphaning to a parent shell that closes is the user's call).

## Sequence on first SIGINT

```
T+0ms     SIGINT received
T+0ms     ctx.cancel() fires
T+1ms     [16].Exec sends SIGINT to active child process group(s)
T+0..2s   Children handle SIGINT and exit (claude/codex have their own
          handlers; codex prints "Interrupted." to stderr)
T+2s      [16] sends SIGKILL to any still-running child
T+~       Engine.Run's outer loop returns ErrInterrupted
T+~       buildSummary(...) runs; Termination = "interrupted"
T+~       state.WriteEnd(sess, e) — atomic, fsynced
T+~       Engine.Run returns
T+~       finalize() restores default handlers
T+~       cmd/debate/main.go exits with the right code (see [23])
```

Total latency on a "stuck" subprocess: 2s SIGINT grace + immediate kill = ~2s worst case. Verified by [25](25-probes.md).

## Sequence on second SIGINT

A second SIGINT *during* finalize (i.e., before `Engine.Run` returns) bypasses the graceful path:

1. The finalize function's secondary handler calls `os.Exit(130)`.
2. `end.json` may be missing (it would have been written but the process is dying).
3. `log.jsonl` will not have a `kind:run` line for this session (see [10](10-run-artifacts.md)).

Detection by the user: a session folder under `sessions/` with `end.json` missing AND no log entry is "doubly-interrupted." v0 documents this state as recoverable for inspection but not for resume (resume is v1).

## SIGTERM behavior

Same path as SIGINT. CI runners and PID-1 init systems use SIGTERM for shutdown; honoring it is required for clean container exit.

## Exit-code translation

After `Engine.Run` returns, `cmd/debate/main.go` translates the result into a process exit code:

```
intrinsic := summaryExitCode(summary)   // 0, 1, or pre-flight code from [06]

if hookMode {
    os.Exit(0)
} else {
    os.Exit(intrinsic)
}
```

Where `summaryExitCode`:

| Termination | Unresolved count | exit code |
|---|---|---|
| steady-state | 0 | 0 |
| steady-state | ≥ 1 | 1 |
| max-turn | any | 1 |
| cost-cap | any | 1 |
| malformed-output | any | 1 |
| interrupted | any | 130 (manual) / 0 (`--hook-mode`) |

(`--hook-mode` overrides intrinsic to 0 except in pre-flight failures, which always propagate. See [06](06-preflight.md), [23](23-summary-render.md).)

## Process-tree teardown

Owned by [16](16-subprocess-infra.md) at the per-call level. This spec only requires that `Engine.Run` *not* swallow `ctx.Err()`; instead it returns it (or wraps as `ErrInterrupted`) so `cmd/debate/main.go` knows to exit non-zero.

## Test contract

- Unit: `InstallHandler`'s context cancels within 1ms of `signal.SendSignal(os.Getpid(), os.Interrupt)`.
- Integration: `Engine.Run` cancelled mid-fork writes `end.json` with `Termination = "interrupted"` before returning.
- Integration: a child process spawned by [16](16-subprocess-infra.md) actually receives SIGINT (process-group test).
- Probe: ([25](25-probes.md)) measures wall time from SIGINT to `Engine.Run` return; must be < 5s on a stuck mock subprocess.

## Acceptance criteria

- [ ] Single SIGINT writes `end.json` with `Termination = "interrupted"`.
- [ ] Double SIGINT bypasses cleanup and exits 130 within 100ms.
- [ ] Subprocesses are killed (verified by `pgrep -f` after the test).
- [ ] Exit-code matrix above is covered by table-driven tests.
- [ ] Pre-flight failure exit codes propagate even under `--hook-mode`.
