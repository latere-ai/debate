#!/usr/bin/env bash
# Probe: trivial-diff exit-fast must take < 200 ms median wall (per
# spec 30; reconciles the as-shipped 200ms threshold with spec 25's
# stale 100ms claim, see specs/30-probe-trivial-diff-perf-outcome.md
# for the rationale). Runs bin/agon against a 1-changed-line repo
# with --changed-lines-min 100 so the trivial-diff gate fires
# immediately.
#
# Three runs; reports min/median/max. Verdict is PASS iff median <
# 200 ms.

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"
ensure_clean_env

REPO=$(cd "$(dirname "$0")/../.." && pwd)
BIN="${AGON_BIN:-$REPO/bin/agon}"
if [ ! -x "$BIN" ]; then
  echo "missing: $BIN (run 'make build' first or set AGON_BIN)" >&2
  exit 2
fi

WORKDIR=$(mk_tmpdir)
trap 'rm -rf "$WORKDIR"' EXIT

cd "$WORKDIR"
git init -q
echo a > a.txt
git -c user.email=t@e.com -c user.name=t add . >/dev/null
git -c user.email=t@e.com -c user.name=t commit -q -m init
echo b >> a.txt

now_ms() { perl -MTime::HiRes=time -e 'printf("%.0f", time*1000)'; }

samples=()
for i in 1 2 3; do
  s=$(now_ms)
  "$BIN" --task-context "noop" --changed-lines-min 100 --side-count 1 \
    >/dev/null 2>&1 || true
  e=$(now_ms)
  samples+=( $(( e - s )) )
done

# Sort ascending and compute min/median/max.
sorted=($(printf '%s\n' "${samples[@]}" | sort -n))
min=${sorted[0]}
median=${sorted[1]}
max=${sorted[2]}

echo "trivial-gate wall (3 runs): min=${min}ms median=${median}ms max=${max}ms"
if [ "$median" -le 200 ]; then
  echo "PASS"
  exit 0
fi
echo "FAIL (median ${median}ms > 200ms budget)"
exit 1
