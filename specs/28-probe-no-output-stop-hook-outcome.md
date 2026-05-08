# Spec 28 - Probe G4 outcome: no-output Stop hook

> **Status: ✅ implemented** (G4 PASS against claude 2.1.131. Spec 01 wording tightened to drop the "probe owed" caveats. The release-blocker gate this spec closed was retracted in the 2026-05-08 simplification of [27](27-release.md); the probe and its disposition stand as historical record.)
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"v0 release blockers" and [25-probes.md](25-probes.md) for design intent.

**Depends on:** [25](25-probes.md), [27](27-release.md).
**Consumed by:** [27](27-release.md).

## Scope

In: a single, recorded execution of `scripts/probes/no-output-stop-hook.sh` against the maintainer's local `claude` install, and the disposition of [01-overview.md](01-overview.md) §"Lifecycle invariants" wording based on the result.

Out: changes to the probe script itself ([25](25-probes.md)), changes to the Stop-hook plumbing ([24](24-stop-hook.md)).

## Why this is the GA blocker

[01-overview.md](01-overview.md) commits to one of two wordings for the root-preservation invariant:

- **PASS branch:** "byte-identical root JSONL" - applies to non-Stop-hook modes *and* Option B.
- **FAIL branch:** "no debate-content pollution" - the stricter byte-identical claim is dropped for Option B; a single per-run `hook_success` attachment is the worst-case mutation.

The probe is the only way to discriminate. It must run before tagging.

## Execution

1. Maintainer runs `scripts/probes/no-output-stop-hook.sh` on a machine with `claude` 2.1.x installed.
2. If `claude --print` errors on a fresh project (interactive trust prompt), maintainer runs `claude` once interactively in `$WORKDIR`, accepts trust, then re-runs the probe.
3. Captures full stdout and exit code. Both go into the release-notes channel chosen in [35](35-release-notes-channel.md).

## Recording format

The recorded outcome is a single block:

```
probe: no-output-stop-hook
claude_version: <output of `claude --version`>
host_os: darwin|linux
exit_code: 0 | 1 | 3
verdict: PASS | FAIL | SKIP
hook_attachments: <verbatim from probe stdout, or "(none)">
```

`SKIP` is permitted only when `claude` cannot be exercised non-interactively after step 2 (exit code 3 from the probe). A `SKIP` does **not** unblock GA; the probe must produce PASS or FAIL.

## Disposition

- **PASS:** [01-overview.md](01-overview.md) §"Lifecycle invariants" loses the "(probe owed before v0 GA)" caveat; the wording tightens to byte-identical root JSONL across all modes. Maintainer edits [01-overview.md](01-overview.md) lines that currently hedge ("the byte-identical claim only holds in...") and removes the parenthetical "probe owed" notes in lines 23, 71, 565.
- **FAIL:** [01-overview.md](01-overview.md) §"Lifecycle invariants" stays as-is (it is already correct under this branch); the parenthetical "probe owed" notes are replaced by a one-line probe-confirmed reference to the recording.

In either branch, [27-release.md](27-release.md) G4 flips from `[x]` (aspirational) to a real check, with the recorded outcome cited.

## Acceptance criteria

- [x] Probe ran once on the release-candidate machine; full stdout captured.
- [x] Recording captured (originally in `release-notes-v0.0.1.md`; removed in the 2026-05-08 simplification along with the release-evidence bundle). The probe stdout transcript stays in the maintainer's local notes; outcome is not an artifact per simplified [27](27-release.md).
- [x] [01-overview.md](01-overview.md) wording updated per disposition: lines 23, 71, 565 now read "probe-confirmed 2026-05" instead of "probe owed before v0 GA".
- [x] ~~[27-release.md](27-release.md) G4 outcome line cites the recording by name.~~ *(retracted: G4 no longer exists as a release blocker.)*

## Out-of-scope-but-fixed

The probe had two pre-existing bugs that prevented it from ever
producing a verdict in this environment:

- The path-encoding step used logical `pwd`, which on macOS produces a
  different string from claude's canonical-path encoding (claude resolves
  /tmp to /private/tmp). Fixed by using `pwd -P`.
- `claude --print` triggered a stdin-wait warning on every invocation.
  Fixed by redirecting stdin from /dev/null on both `claude --print`
  calls.
