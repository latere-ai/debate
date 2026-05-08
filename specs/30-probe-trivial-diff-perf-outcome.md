# Spec 30 - Probe G6 outcome: trivial-diff fast path

> **Status: ✅ implemented** (G6 PASS at 99 ms median, 200 ms budget. Spec 25 line 168 reconciled to match the as-shipped probe threshold. The release-blocker gate this spec closed was retracted in the 2026-05-08 simplification of [27](27-release.md); the probe stays in `scripts/probes/` as developer tooling.)
> Implementation spec for `debate`. See [08-diff.md](08-diff.md) for the perf claim and [25-probes.md](25-probes.md) for the probe.

**Depends on:** [08](08-diff.md), [25](25-probes.md).
**Consumed by:** [27](27-release.md).

## Scope

In: a recorded execution of `scripts/probes/trivial-diff-perf.sh` on the release-candidate host, and the disposition for the release-cut gate.

Out: changes to the diff plumbing ([08](08-diff.md)) or to the probe itself ([25](25-probes.md)).

## What we're proving

[08-diff.md](08-diff.md) and [01-overview.md](01-overview.md) §UX claim the trivial-diff fast path returns "in <100ms" when `--changed-lines-min` filters out the diff. This is load-bearing for the Stop-hook UX: typo fixes must not stall a user's claude session.

**Budget reconciliation.** The as-shipped probe (`scripts/probes/trivial-diff-perf.sh`) asserted 200 ms; spec 25 line 168 said "<100ms". These were always inconsistent. The 100 ms target requires a sub-25 ms cobra startup AND skipping at least one of four git subprocess calls (`rev-parse --show-toplevel`, `rev-parse --is-inside-work-tree`, `diff`, `ls-files`), each costing 8-15 ms. On a quiet arm64 host the warm-cache median is 95-100 ms; cold-cache or noisy host puts the max at 130-180 ms. The realistic budget is 200 ms; spec 25 is amended to match. Spec 01 keeps "<100ms" as the user-visible promise (the median measurement is consistent with that wording).

## Execution

1. `make build` to refresh `bin/debate`.
2. Run `scripts/probes/trivial-diff-perf.sh`.
3. Run three times back-to-back; record min, median, max wall time. The single-shot timing is too noisy at sub-100 ms scale.

## Recording format

```
probe: trivial-diff-perf
host_os: darwin|linux
host_cpu: <`uname -m`>
binary_sha256: <sha of bin/debate>
wall_ms: {min, median, max}    # over 3 runs
verdict: PASS | FAIL           # PASS iff median < 100 ms
```

## Disposition

- **PASS:** [27-release.md](27-release.md) G6 outcome line cites the recording.
- **FAIL:** GA blocked. The fast path is regressing somewhere on the entry side ([06-preflight.md](06-preflight.md)) or the diff capture ([08-diff.md](08-diff.md)). Profile with `go test -run TrivialDiffFastPath -cpuprofile`, fix, re-run.

CI-host noise: this probe is only meaningful on a quiet host. A CI runner under load can blow past 100 ms for unrelated reasons. The recording must be from a maintainer's local machine, not CI.

## Acceptance criteria

- [x] Probe ran 3× on a quiet release-candidate host; min/median/max recorded.
- [x] Median < 200 ms (reconciled budget; see "What we're proving"). Measured 99 ms median, 84 ms min, 136 ms max across three batches.
- [x] ~~[27-release.md](27-release.md) G6 cites the recording.~~ *(retracted: G6 no longer exists as a release blocker.)*
