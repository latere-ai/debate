#!/usr/bin/env bash
# Probe: does a Stop hook that emits nothing still mutate root JSONL?
# v0 GA blocker. See specs/25-probes.md and specs/01-overview.md
# §"v0 release blockers".

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"

require_bin claude
require_bin jq
ensure_clean_env

WORKDIR=$(mk_tmpdir)
trap '[ "${KEEP:-0}" = "1" ] || rm -rf "$WORKDIR"' EXIT

cd "$WORKDIR"
git init -q
echo "fixture" > readme.txt
git -c user.email=t@e.com -c user.name=t add . && \
  git -c user.email=t@e.com -c user.name=t commit -q -m init

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

claude --print "ack" >/dev/null 2>&1 || {
  echo "claude --print failed (likely needs interactive trust); cannot run probe" >&2
  exit 3
}

PROJECTS_DIR="$HOME/.claude/projects/$(pwd | sed 's|/|-|g')"
ROOT_JSONL=$(ls -t "$PROJECTS_DIR"/*.jsonl 2>/dev/null | head -1 || true)
if [ -z "$ROOT_JSONL" ]; then
  echo "no root JSONL found under $PROJECTS_DIR" >&2
  exit 3
fi

BEFORE_SHA=$(sha256 "$ROOT_JSONL")
BEFORE_SIZE=$(stat -f%z "$ROOT_JSONL" 2>/dev/null || stat -c%s "$ROOT_JSONL")

claude --resume "$(basename "$ROOT_JSONL" .jsonl)" --print "ack again" >/dev/null 2>&1 || true

AFTER_SHA=$(sha256 "$ROOT_JSONL")
AFTER_SIZE=$(stat -f%z "$ROOT_JSONL" 2>/dev/null || stat -c%s "$ROOT_JSONL")

HOOK_ATTACHMENTS=$(jq -c 'select(.type | tostring | test("hook_")) | {type, ts: .timestamp}' "$ROOT_JSONL" 2>/dev/null || true)

echo "before sha: $BEFORE_SHA  size: $BEFORE_SIZE"
echo "after  sha: $AFTER_SHA   size: $AFTER_SIZE"
echo "hook_* attachments after run:"
echo "${HOOK_ATTACHMENTS:-(none)}"

if [ -z "$HOOK_ATTACHMENTS" ]; then
  echo "PASS: no hook_* attachments written by no-output Stop hook"
  exit 0
else
  echo "FAIL: Stop hook produced hook_* attachments in root JSONL even with no output"
  echo "Implication: keep the 'no debate-content pollution' wording in spec 01."
  exit 1
fi
