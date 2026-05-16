#!/usr/bin/env bash
# debate-stop-hook.sh - Stop hook for the `debate` CLI.
# See specs/24-stop-hook.md and specs/01-overview.md for design.

set -e

# Recursion guard. The orchestrator spawns `claude --resume <fork-id> -p ...`
# (and `claude -p` for claude-as-critic) subprocesses; those subprocesses
# also fire the Stop hook when they finish. Without this guard the hook
# would re-enter the orchestrator on every round and fork infinitely.
if [ -n "$AGON_IN_PROGRESS" ]; then
  exit 0
fi
export AGON_IN_PROGRESS=1

PAYLOAD=$(cat)
SESSION_ID=$(printf '%s' "$PAYLOAD" | jq -r '.session_id // empty')
TRANSCRIPT=$(printf '%s' "$PAYLOAD" | jq -r '.transcript_path // empty')
CWD=$(printf '%s' "$PAYLOAD" | jq -r '.cwd // empty')

# Stale ANTHROPIC_API_KEY in env causes 401 in claude -p subprocesses.
unset ANTHROPIC_API_KEY

# --resume requires running from the cwd that owns the project's session dir.
if [ -n "$CWD" ]; then
  cd "$CWD" || exit 1
fi

# Hand off to the orchestrator. exec lets stdout/stderr flow through to
# the surrounding shell.
#
# `--hook-mode` forces exit 0 even on unresolved leaves. Without it the
# CLI's intrinsic exit 1 would propagate, and Claude would treat the
# Stop hook as having failed for every interesting review.
exec debate \
  --hook-mode \
  --session-id "$SESSION_ID" \
  --transcript "$TRANSCRIPT" \
  --max-turn 6
