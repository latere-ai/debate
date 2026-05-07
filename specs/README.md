# debate specs

The implementation specs for `debate`. Each one is the authoritative
contract for a slice of the system; the code is meant to follow what
the spec says, not the other way around. Read in order; each builds on
the ones before it.

## Design

- [01 Overview](01-overview.md) - architecture, fork model, v0/v1 split, lifecycle invariants

## Foundations

- [02 Go module + layout](02-go-module.md)
- [03 CI, lint, release pipeline](03-ci-lint-release.md)
- [04 CLI flags](04-cli-flags.md)
- [05 `.debate.toml` config](05-config-file.md)
- [06 Preflight](06-preflight.md)

## Inputs

- [07 Claude transcript ingest](07-claude-transcript.md)
- [08 Diff capture](08-diff.md)

## On-disk layout

- [09 State directory](09-state-dir.md)
- [10 Run-level artifacts](10-run-artifacts.md)
- [11 Fork-level artifacts](11-fork-artifacts.md)
- [12 Attacks ledger](12-attacks-ledger.md)

## Critic protocol

- [13 Critic output format](13-critic-output-format.md)
- [14 Attack parser](14-attack-parser.md)
- [15 Aspect / topic prompts](15-aspect-prompts.md)

## Subprocess + agents

- [16 Subprocess infra](16-subprocess-infra.md)
- [17 Claude proposer](17-claude-proposer.md)
- [18 Critic drivers](18-critic-drivers.md)

## Orchestration

- [19 Round loop](19-round-loop.md)
- [20 Termination](20-termination.md)
- [21 Signals](21-signals.md)

## Output

- [22 Contention headline](22-contention-headline.md)
- [23 Summary render](23-summary-render.md)

## Triggering + verification

- [24 Stop hook](24-stop-hook.md)
- [25 Probes](25-probes.md)
- [26 Tests](26-tests.md)
- [27 Release process and GA gates](27-release.md)

## Release-cut follow-ups (v0.0.1)

These specs were added during the v0.0.1 release cut. Each one closes
a numbered GA gate from spec 27 and writes its outcome into
[`release-notes-v0.0.1.md`](../release-notes-v0.0.1.md).

- [28 G4 probe: no-output Stop hook](28-probe-no-output-stop-hook-outcome.md) - byte-identical root invariant
- [29 G5 probe: signal latency](29-probe-signal-latency-outcome.md) - SIGINT to exit < 5s
- [30 G6 probe: trivial-diff fast path](30-probe-trivial-diff-perf-outcome.md) - hook returns < 200ms median
- [31 G7 probe: interactive stdout](31-probe-interactive-stdout-outcome.md) - non-blocking, manual
- [32 G13 real-e2e suite](32-real-e2e-suite.md) - e2e/real test behind real_e2e build tag
- [33 G15 install-hook smoke](33-install-hook-smoke.md) - settings.json shape + idempotency
- [34 G16 real-claude end-to-end](34-real-claude-end-to-end-smoke.md) - debate runs with real claude+codex
- [35 Release-notes channel](35-release-notes-channel.md) - decides where probe/gate outcomes live

## Status

All specs are marked ✅ implemented. Spec status lines individually
record what's verified, what's deferred, and what was changed during
the release cut.
