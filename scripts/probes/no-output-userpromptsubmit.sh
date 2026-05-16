#!/usr/bin/env bash
# Probe: does a no-stdout UserPromptSubmit hook that exits 2 (block +
# "erase the prompt") leave the root JSONL byte-identical, or does the
# blocked sentinel prompt still land as a user entry?
#
# This is the load-bearing fact for the manual-trigger design: a
# sentinel like `/agon-attack` typed into Claude Code would fire a
# UserPromptSubmit hook that setsid-detaches the orchestrator, writes
# nothing to stdout, and exits 2 so the sentinel never becomes a real
# turn. That only satisfies the "no agon-content pollution of root"
# invariant if the blocked prompt does not get recorded in root JSONL.
# Mirrors scripts/probes/no-output-stop-hook.sh. See specs/25-probes.md
# and specs/36-probe-userpromptsubmit-manual-trigger.md.

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"

require_bin claude
require_bin jq
ensure_clean_env

WORKDIR=$(mk_tmpdir)
MARKER="$WORKDIR.ups_fired"
trap '[ "${KEEP:-0}" = "1" ] || rm -rf "$WORKDIR" "$MARKER"' EXIT

SENTINEL="agon-attack-probe-$$"

cd "$WORKDIR"
git init -q
echo "fixture" > readme.txt
git -c user.email=t@e.com -c user.name=t add . && \
  git -c user.email=t@e.com -c user.name=t commit -q -m init

mkdir -p .claude

# Step 1: create the session WITHOUT any hook, so a clean baseline
# root JSONL exists to diff against.
cat > .claude/settings.json <<'EOF'
{}
EOF

claude --print "ack" </dev/null >/dev/null 2>&1 || {
  echo "claude --print failed (likely needs interactive trust); cannot run probe" >&2
  exit 3
}

# claude resolves cwd through realpath (macOS: /tmp -> /private/tmp).
# pwd -P returns the canonical path so the encoded projects dir matches
# what claude actually wrote (spec 28 out-of-scope fix).
CANONICAL=$(pwd -P)
PROJECTS_DIR="$HOME/.claude/projects/$(echo "$CANONICAL" | sed 's|/|-|g')"
ROOT_JSONL=$(ls -t "$PROJECTS_DIR"/*.jsonl 2>/dev/null | head -1 || true)
if [ -z "$ROOT_JSONL" ]; then
  echo "no root JSONL found under $PROJECTS_DIR" >&2
  echo "(claude likely failed to authenticate or did not persist a session;" >&2
  echo " external API key required for --print mode, see spec 28)" >&2
  exit 3
fi
SESSION_ID=$(basename "$ROOT_JSONL" .jsonl)

BEFORE_SHA=$(sha256 "$ROOT_JSONL")
BEFORE_SIZE=$(stat -f%z "$ROOT_JSONL" 2>/dev/null || stat -c%s "$ROOT_JSONL")
BEFORE_LINES=$(wc -l < "$ROOT_JSONL" | tr -d ' ')

# Step 2: install a no-stdout UserPromptSubmit hook that exits 2. The
# marker file proves the hook actually executed (written to a path
# OUTSIDE the project dir so it cannot affect the JSONL/sha). Nothing
# is written to stdout - UserPromptSubmit stdout would itself pollute
# context (unlike Stop-hook stdout), so the design forbids it.
cat > .claude/settings.json <<EOF
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "/bin/sh -c 'date > \"$MARKER\"; exit 2'" }
        ]
      }
    ]
  }
}
EOF

# Step 3: submit the sentinel on the existing session. The hook fires,
# writes nothing to stdout, exits 2 -> Claude blocks/erases the prompt.
# claude exits non-zero on a blocked prompt; that is expected.
claude --resume "$SESSION_ID" --print "$SENTINEL" </dev/null >/dev/null 2>&1 || true

AFTER_SHA=$(sha256 "$ROOT_JSONL")
AFTER_SIZE=$(stat -f%z "$ROOT_JSONL" 2>/dev/null || stat -c%s "$ROOT_JSONL")
AFTER_LINES=$(wc -l < "$ROOT_JSONL" | tr -d ' ')

HOOK_ATTACHMENTS=$(jq -c 'select(.type | tostring | test("hook_")) | {type, ts: .timestamp}' "$ROOT_JSONL" 2>/dev/null || true)
SENTINEL_HITS=$(grep -c -- "$SENTINEL" "$ROOT_JSONL" 2>/dev/null || true)

echo "probe:           no-output-userpromptsubmit (exit 2, no stdout)"
echo "sentinel:        $SENTINEL"
echo "hook fired:      $([ -f "$MARKER" ] && echo yes || echo NO)"
echo "before sha:      $BEFORE_SHA  size: $BEFORE_SIZE  lines: $BEFORE_LINES"
echo "after  sha:      $AFTER_SHA   size: $AFTER_SIZE  lines: $AFTER_LINES"
echo "sentinel in JSONL lines: ${SENTINEL_HITS:-0}"
echo "hook_* attachments after run:"
echo "${HOOK_ATTACHMENTS:-(none)}"

if [ ! -f "$MARKER" ]; then
  echo "INCONCLUSIVE: UserPromptSubmit hook never fired under claude --resume --print;"
  echo "the byte-diff is meaningless. Re-run interactively or with a different driver."
  exit 3
fi

if [ "$BEFORE_SHA" = "$AFTER_SHA" ] && [ "${SENTINEL_HITS:-0}" = "0" ] && [ -z "$HOOK_ATTACHMENTS" ]; then
  echo "PASS: no-stdout UserPromptSubmit exit-2 left root JSONL byte-identical;"
  echo "the blocked sentinel produced no user entry and no hook_* attachment."
  echo "Implication: a UserPromptSubmit sentinel is a viable zero-pollution manual trigger."
  exit 0
else
  echo "FAIL: blocked sentinel prompt mutated root JSONL"
  echo "(sentinel entry and/or hook_* attachment and/or sha drift)."
  echo "Implication: manual trigger cannot claim byte-identical root;"
  echo "fall back to Stop-hook auto + CLI-only manual, documented honestly."
  exit 1
fi
