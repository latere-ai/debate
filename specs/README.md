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

## Related research and future work

`debate` productizes the architecture from
[agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance)
spec 07 ([Adversarial
Debate](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md)).
The same research line has six follow-up specs (08–13) that explore
extensions and limits of the debate protocol. Each one's empirical
result either licenses or constrains the productization direction
here:

- **[BFT 08 - Compute-Asymmetric Debate](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/08-compute-asymmetric-debate.md)**
  - tests Brown-Cohen et al. (2023) doubly-efficient debate at
  {1×, 5×, 10×, 50×} compute asymmetry between honest and Byzantine
  players. **Future work:** if soundness stays flat under compute
  asymmetry, expose per-role compute knobs (proposer vs. critic
  model, max-turn, retry budget) so users can defend against a
  well-resourced attacker without paying matched compute themselves.
- **[BFT 09 - Debate Depth and Recursive Sub-Debate](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/09-debate-depth-recursion.md)**
  - sweeps round budget K ∈ {2, 4, 8, 12} on the flat protocol and
  compares flat-K vs. recursive sub-debate at matched total compute.
  **Future work:** the answer determines whether the v1 productization
  axis is "more rounds" (just raise `--max-turn` defaults) or
  "structured recursion" (a real architectural addition: sub-debates
  spawned per unresolved leaf).
- **[BFT 10 - Stochastic-System Soundness](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/10-stochastic-system-soundness.md)**
  - necessary-condition check that soundness survives the LLM
  temperature range. **Future work:** if soundness drops above some
  T\*, debate must pin temperature on both proposer and critic - which
  means a `--temperature` flag and refusing to run against agent
  configs that override it. Today we trust the agent CLI defaults.
- **[BFT 11 - Verifier as Spot-Check (PCP leaf format)](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/11-pcp-leaf-spot-check.md)**
  - tightens leaf "stakes" from freeform prose to structured tuples
  (`{file_path, line_range, expected_byte_pattern}`) and measures the
  soundness/judge-cost Pareto. **Future work:** narrower stakes
  reduce the human's read time on `summary.md`; spec 13 here already
  picks the stake from the critic's reproduction field, but a
  structured leaf format would let the headline render as a
  click-to-jump reference instead of a paragraph.
- **[BFT 12 - Adversarial Obfuscation / Prover-Estimator Debate](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/12-prover-estimator-obfuscation.md)**
  - identifies the obfuscated-arguments attack class and tests the
  Prover-Estimator fix from Brown-Cohen et al. (2025). **Future
  work:** if obfuscation is a real LLM attack class, the binary
  concede/rebut/withdraw critic in spec 13 here is unsound on
  pathological diffs. The fix is replacing the critic's
  binary-disposition contract with a scalar plausibility estimator.
  This is a v2-class architectural change, not a config tweak.
- **[BFT 13 - Debate Query Complexity (DQC) Scaling](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/13-dqc-scaling.md)**
  - measures whether judge-read tokens scale as O(log *n*) or
  O(*n*^a) as snippet complexity grows. **Future work:** the empirical
  scaling slope bounds the headline-rendering story. If LLM debate
  doesn't realize the doubly-efficient property at scale, the
  contention-scored headline (spec 22 here) becomes the *only*
  doubly-efficient channel and `summary.md` body length needs an
  explicit cap.

The general pattern: the BFT-repo line answers "is debate sound under
condition X?"; this repo asks "given the answer, what does the
production tool look like?" Most of the answers point at v1 work
items already in [01-overview.md §v1](01-overview.md), but BFT specs
12 (Prover-Estimator) and 09 (recursion) describe architectural
changes large enough to be v2 territory.
