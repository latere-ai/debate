# Spec 25 - Probes (incl. v0 GA blocker)

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"v0 release blockers" and §"Constraints uncovered by the probe" for design intent.

**Depends on:** [16](16-subprocess-infra.md), [21](21-signals.md), [24](24-stop-hook.md).
**Consumed by:** [27](27-release.md).

## Scope

In: standalone probe scripts that the user runs against their actual claude/codex installation to verify environment-dependent invariants. Includes the **no-output Stop-hook probe** that v0 GA blocks on, plus optional probes for interactive-stdout rendering, signal latency, and trivial-diff perf.

Out: integration tests against mock claude/codex (those live in [26](26-tests.md)).

## Probe harness

`scripts/probes/`:

```
scripts/probes/
  README.md
  no-output-stop-hook.sh        # GA blocker
  interactive-stdout.sh         # optional
  signal-latency.sh             # smoke
  trivial-diff-perf.sh          # smoke
  lib/
    common.sh                   # shared helpers
```

Common helpers (`lib/common.sh`):

```bash
require_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 2; }
}
ensure_clean_env() {
  unset ANTHROPIC_API_KEY DEBATE_IN_PROGRESS
}
mk_tmpdir() { mktemp -d "${TMPDIR:-/tmp}/debate-probe-XXXXXX"; }
sha256() { shasum -a 256 "$1" | cut -d' ' -f1; }
```

All probes:

1. Print a one-line description and bail-out hint on failure.
2. Set up an isolated temp dir.
3. Snapshot the relevant root-session state hash before.
4. Run the probe.
5. Snapshot after, compute the diff, print PASS/FAIL with reasoning.
6. Clean up unless `KEEP=1` is set.

## Probe: no-output-stop-hook (GA blocker)

`scripts/probes/no-output-stop-hook.sh`:

```bash
#!/usr/bin/env bash
# Probe: does a Stop hook that emits nothing still mutate root JSONL?
#
# v0 GA blocker. See specs/01-overview.md §"v0 release blockers" and
# specs/24-stop-hook.md.

set -euo pipefail
. "$(dirname "$0")/lib/common.sh"

require_bin claude
require_bin jq
ensure_clean_env

WORKDIR=$(mk_tmpdir)
trap '[ "${KEEP:-0}" = "1" ] || rm -rf "$WORKDIR"' EXIT

cd "$WORKDIR"
git init -q
echo "fixture" > readme.txt
git add . && git commit -q -m init

# Install a Stop hook that emits absolutely nothing.
mkdir -p .claude
cat > .claude/settings.json <<'EOF'
{
  "hooks": {
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "/usr/bin/true" }
        ]
      }
    ]
  }
}
EOF

# Trust the project so the hook fires (interactive prompt is the only
# way claude trusts a project; we accept that this probe needs human
# attention once per machine).
claude --print "ack" >/dev/null

# Find the resulting root session JSONL and snapshot.
PROJECTS_DIR="$HOME/.claude/projects/$(pwd | sed 's|/|-|g')"
ROOT_JSONL=$(ls -t "$PROJECTS_DIR"/*.jsonl | head -1)

BEFORE_SIZE=$(stat -f%z "$ROOT_JSONL" 2>/dev/null || stat -c%s "$ROOT_JSONL")
BEFORE_SHA=$(sha256 "$ROOT_JSONL")

# Trigger the hook by sending one more turn.
claude --resume "$(basename "$ROOT_JSONL" .jsonl)" --print "ack again" >/dev/null

AFTER_SIZE=$(stat -f%z "$ROOT_JSONL" 2>/dev/null || stat -c%s "$ROOT_JSONL")
AFTER_SHA=$(sha256 "$ROOT_JSONL")

# Compare. The interesting question: did the file grow by anything that
# the hook itself caused (vs. the new user/assistant turns we sent)?
HOOK_ATTACHMENTS=$(jq -c 'select(.type | test("hook_")) | {type, name: (.systemMessage // .name // ""), ts: .timestamp}' "$ROOT_JSONL" || true)

echo "before sha: $BEFORE_SHA  size: $BEFORE_SIZE"
echo "after  sha: $AFTER_SHA   size: $AFTER_SIZE"
echo "hook_* attachments after run:"
echo "$HOOK_ATTACHMENTS"

if [ -z "$HOOK_ATTACHMENTS" ]; then
  echo "PASS: no hook_* attachments written by no-output Stop hook"
  exit 0
else
  echo "FAIL: Stop hook produced hook_* attachments in root JSONL even with no output"
  echo "Implication: spec/01 §'Lifecycle invariants' must keep the 'no debate-content pollution' wording"
  exit 1
fi
```

Outcome consumed by [27](27-release.md):

- **PASS**: GA can claim the strict "byte-identical root JSONL" form of root-preservation in non-Stop-hook modes AND in Option B. The spec's caveat at [01-overview.md](01-overview.md) §"Lifecycle invariants" can be tightened.
- **FAIL**: GA proceeds with the existing "no debate-content pollution" wording. No code change needed - the spec is already correct under this branch.

Either way, GA is unblocked once the probe runs and its result is recorded in the release notes.

## Probe: interactive-stdout (optional)

`scripts/probes/interactive-stdout.sh`:

Runs claude in an interactive PTY (via `script` or `expect`), triggers a Stop hook that emits `[debate] hello` on its stdout, and checks whether that line surfaces in the user's terminal.

PASS / FAIL informs whether [23](23-summary-render.md)'s "stdout best-effort" qualifier in [01-overview.md](01-overview.md) and the README ([24](24-stop-hook.md)) can be tightened to "stdout surfaces in interactive mode."

Optional: if the probe doesn't run before GA, the conservative wording stays.

## Probe: signal-latency

`scripts/probes/signal-latency.sh`:

```bash
# Spawn `bin/debate` against a mock subprocess, send SIGINT, time how
# long until the process exits.
# Asserts < 5s on a stuck child (per [21]).
```

Uses a mock `claude` shim that sleeps; not gated on a real claude install.

## Probe: trivial-diff-perf

`scripts/probes/trivial-diff-perf.sh`:

```bash
# Run `bin/debate --changed-lines-min 100` against a 5-line-diff repo.
# Asserts wall time < 100ms.
# Verifies [08]'s exit-fast path performance claim.
```

## Probe runner

`make probe` runs all probes in `scripts/probes/` and prints a summary table:

```
PROBE                          STATUS  DURATION
no-output-stop-hook            PASS    1.2s
interactive-stdout             SKIP    -
signal-latency                 PASS    0.4s
trivial-diff-perf              PASS    0.08s
```

Exit code 0 iff all non-skipped probes pass. `make probe` is *not* part of CI (probes need real claude/codex installs); it's intended as a release-cut precondition driven by [27](27-release.md).

## Test contract

- The harness itself has unit tests for `lib/common.sh` (shellcheck pass + bats-style assertions for `mk_tmpdir`, `sha256`).
- The probes are smoke-tested in CI against mock binaries to ensure they don't crash on basic inputs; the *interesting* assertions only run with real installs.

## Acceptance criteria

- [x] `make probe` runs all four scripts; results recorded in `release-notes-v0.0.1.md`.
- [x] `no-output-stop-hook.sh` produces a definitive PASS/FAIL with reasoning printed.
- [x] All probe scripts pass `shellcheck -e SC1091` (sourcing common.sh is allowed).
- [x] No probe modifies state outside its temp dir (audit by `find` after run).
- [x] [27](27-release.md)'s GA checklist references the probe outputs by file path.
