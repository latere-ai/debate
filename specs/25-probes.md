# Spec 25 - Probes

> **Status: ✅ implemented**, pruned 2026-05-16: the no-output
> **Stop-hook** probe (the old v0 GA blocker) and the
> **interactive-stdout** probe were retired with the Stop hook. The
> surviving probes are `signal-latency` and `trivial-diff-perf`; the
> one-off `no-output-userpromptsubmit` probe (evidence for *why* the
> hook went) is owned by [36](36-probe-userpromptsubmit-manual-trigger.md).
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) for design intent.

**Depends on:** [16](16-subprocess-infra.md), [21](21-signals.md).
**Consumed by:** [27](27-release.md).

## Scope

In: standalone probe scripts that the user runs against their actual claude/codex installation to verify environment-dependent invariants. The current probes are signal latency and trivial-diff perf. (`no-output-userpromptsubmit.sh` is a one-off recorded in [36](36-probe-userpromptsubmit-manual-trigger.md), not part of the routine set.)

Out: integration tests against mock claude/codex (those live in [26](26-tests.md)).

## Probe harness

`scripts/probes/`:

```
scripts/probes/
  README.md
  signal-latency.sh             # smoke
  trivial-diff-perf.sh          # smoke
  no-output-userpromptsubmit.sh # one-off; see spec 36
  lib/
    common.sh                   # shared helpers
```

Common helpers (`lib/common.sh`):

```bash
require_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 2; }
}
ensure_clean_env() {
  unset ANTHROPIC_API_KEY
}
mk_tmpdir() { mktemp -d "${TMPDIR:-/tmp}/agon-probe-XXXXXX"; }
sha256() { shasum -a 256 "$1" | cut -d' ' -f1; }
```

All probes:

1. Print a one-line description and bail-out hint on failure.
2. Set up an isolated temp dir.
3. Snapshot the relevant root-session state hash before.
4. Run the probe.
5. Snapshot after, compute the diff, print PASS/FAIL with reasoning.
6. Clean up unless `KEEP=1` is set.

## Probe: signal-latency

`scripts/probes/signal-latency.sh`:

```bash
# Spawn `bin/agon` against a mock subprocess, send SIGINT, time how
# long until the process exits.
# Asserts < 5s on a stuck child (per [21]).
```

Uses a mock `claude` shim that sleeps; not gated on a real claude install.

## Probe: trivial-diff-perf

`scripts/probes/trivial-diff-perf.sh`:

```bash
# Run `bin/agon --changed-lines-min 100` against a 5-line-diff repo.
# Asserts median wall time < 200ms across 3 runs.
# Verifies [08]'s exit-fast path performance claim.
# Note: spec 01 §UX still names <100ms as the aspirational target; 200ms
# is the realistic probe budget given four short-lived git subprocess
# calls plus cobra startup. See [30](30-probe-trivial-diff-perf-outcome.md).
```

## Probe runner

`make probe` runs all probes in `scripts/probes/` and prints a summary table:

```
PROBE                          STATUS  DURATION
signal-latency                 PASS    0.4s
trivial-diff-perf              PASS    0.08s
```

Exit code 0 iff all non-skipped probes pass. `make probe` is *not* part of CI (probes need real claude/codex installs); it's intended as a release-cut precondition driven by [27](27-release.md).

## Test contract

- The harness itself has unit tests for `lib/common.sh` (shellcheck pass + bats-style assertions for `mk_tmpdir`, `sha256`).
- The probes are smoke-tested in CI against mock binaries to ensure they don't crash on basic inputs; the *interesting* assertions only run with real installs.

## Acceptance criteria

- [x] `make probe` runs the probe scripts and prints the summary table.
- [x] All probe scripts pass `shellcheck -e SC1091` (sourcing common.sh is allowed).
- [x] No probe modifies state outside its temp dir (audit by `find` after run).
- [x] [27](27-release.md)'s GA checklist references the probe outputs by file path.
