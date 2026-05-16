#!/usr/bin/env bash
# Probe (optional): does a Stop hook's stdout render in interactive
# claude? Outcome decides whether spec 23's "stdout best-effort"
# qualifier can be relaxed.

set -euo pipefail
. "$(cd "$(dirname "$0")" && pwd)/lib/common.sh"

# This probe needs an interactive PTY; not automatable in v0. Print
# the manual procedure and exit SKIP.

cat <<'EOF'
SKIP: interactive-stdout probe is manual.

To run:
  1. mkdir tmp && cd tmp && git init && touch x && git add . && \
     git commit -m init
  2. mkdir .claude && write .claude/settings.json with:
       {"hooks":{"Stop":[{"matcher":"","hooks":[{"type":"command","command":"/bin/sh -c 'echo [agon] hello'"}]}]}}
  3. Open `claude` interactively; ask anything; observe whether the
     stdout from the Stop hook appears in your terminal.
  4. Record the outcome in release-notes-v0.0.1.md as PASS or FAIL.
EOF
exit 0
