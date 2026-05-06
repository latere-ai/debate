#!/usr/bin/env bash
# Probe: SIGINT -> exit latency on a debate run with a stuck child should
# be < 5s, per specs/21-signals.md.
#
# This probe spawns bin/debate against shell-shim "claude" and "codex"
# binaries that sleep forever, sends SIGINT to debate, and measures
# wall-clock time to exit. The previous version of this probe used a
# self-contained bash sleeper, which (a) did not exercise bin/debate at
# all and (b) had a trap-vs-sleep race that made it hang for the full
# sleep duration. See specs/29-probe-signal-latency-outcome.md.

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"
ensure_clean_env

REPO=$(cd "$(dirname "$0")/../.." && pwd)
BIN="$REPO/bin/debate"
if [ ! -x "$BIN" ]; then
  echo "missing: $BIN (run 'make build' first)" >&2
  exit 2
fi

WORKDIR=$(mk_tmpdir)
trap 'rm -rf "$WORKDIR"' EXIT

# Shim claude + codex: each sleeps long enough that the orchestrator's
# round loop is blocked on the subprocess. The sleep duration is
# deliberately larger than the 5s budget so we know any quick exit is
# the orchestrator's signal handler, not the child terminating on its
# own.
mkdir -p "$WORKDIR/bin"
cat > "$WORKDIR/bin/claude" <<'EOF'
#!/usr/bin/env bash
# Stuck child: never returns until killed.
exec sleep 60
EOF
cp "$WORKDIR/bin/claude" "$WORKDIR/bin/codex"
chmod +x "$WORKDIR/bin/claude" "$WORKDIR/bin/codex"

# Tiny git repo with a non-trivial diff so debate doesn't short-circuit
# at the changed-lines-min gate.
REPO_DIR="$WORKDIR/repo"
mkdir -p "$REPO_DIR"
cd "$REPO_DIR"
git init -q
git -c user.email=t@e.com -c user.name=t commit --allow-empty -q -m init
printf '// fixture line\n%.0s' {1..30} > search.go

# Spawn debate in background; capture PID. Strip DEBATE_IN_PROGRESS so
# the recursion guard does not short-circuit.
unset DEBATE_IN_PROGRESS
PATH="$WORKDIR/bin:$PATH" "$BIN" \
  --main claude --side codex \
  --max-turn 2 --side-count 1 \
  --task-context 'signal-latency probe fixture' \
  --changed-lines-min 10 \
  >/dev/null 2>&1 &
pid=$!

# Wait for debate to actually start spawning subprocesses. 1s is enough
# for the preflight phase on any reasonable host.
sleep 1

start_ts=$(python3 -c 'import time; print(time.time())')
kill -INT "$pid"
wait "$pid" 2>/dev/null || true
end_ts=$(python3 -c 'import time; print(time.time())')
elapsed=$(python3 -c "print(f'{$end_ts - $start_ts:.3f}')")

echo "signal-to-exit: ${elapsed}s"
# Compare as floats; bash arithmetic only handles integers.
if python3 -c "import sys; sys.exit(0 if $elapsed < 5 else 1)"; then
  echo "PASS"
  exit 0
fi
echo "FAIL"
exit 1
