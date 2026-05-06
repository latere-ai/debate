#!/usr/bin/env bash
# Probe: trivial-diff exit-fast must take < 200ms wall.

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"
ensure_clean_env

WORKDIR=$(mk_tmpdir)
trap 'rm -rf "$WORKDIR"' EXIT

cd "$WORKDIR"
git init -q
echo a > a.txt
git -c user.email=t@e.com -c user.name=t add . && \
  git -c user.email=t@e.com -c user.name=t commit -q -m init
echo b >> a.txt

# Time a no-op debate invocation; this is a smoke test against the
# build-tree binary if present.
DEBATE_BIN="${DEBATE_BIN:-debate}"
require_bin "$DEBATE_BIN"

start_ms=$(perl -MTime::HiRes=time -e 'printf("%.0f", time*1000)')
"$DEBATE_BIN" --task-context "noop" --changed-lines-min 100 --side-count 1 --aspect functional-logic >/dev/null 2>&1 || true
end_ms=$(perl -MTime::HiRes=time -e 'printf("%.0f", time*1000)')
elapsed=$(( end_ms - start_ms ))

echo "trivial-gate wall: ${elapsed}ms"
if [ "$elapsed" -le 200 ]; then
  echo "PASS"
  exit 0
fi
echo "FAIL"
exit 1
