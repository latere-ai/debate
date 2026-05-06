# Spec 24 - Stop hook script and install

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Trigger via Stop hook" for design intent.

**Depends on:** [04](04-cli-flags.md), [16](16-subprocess-infra.md), [23](23-summary-render.md).
**Consumed by:** [25](25-probes.md), [27](27-release.md).

## Scope

In: the `debate-stop-hook.sh` bash script (verbose hook format, recursion guard, env hygiene, `--hook-mode` wiring), the install path (user-level vs project-level `settings.json`), an `install`/`uninstall` subcommand on the `debate` CLI for one-step setup, and the documented manual install path.

Out: the no-output Stop-hook probe ([25](25-probes.md)).

## Hook script

`scripts/debate-stop-hook.sh`:

```bash
#!/usr/bin/env bash
# debate-stop-hook.sh - Stop hook for the `debate` CLI.
# See specs/24-stop-hook.md and specs/01-overview.md for design.

set -e

# Recursion guard. The orchestrator spawns `claude --resume <fork-id> -p ...`
# (and `claude -p` for claude-as-critic) subprocesses; those subprocesses
# also fire the Stop hook when they finish. Without this guard the hook
# would re-enter the orchestrator on every round and fork infinitely.
if [ -n "$DEBATE_IN_PROGRESS" ]; then
  exit 0
fi
export DEBATE_IN_PROGRESS=1

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

# Hand off. exec lets the orchestrator's stdout/stderr flow through to
# the surrounding shell so a manual `tail -f` is not required (though
# stdout rendering is best-effort under hooks; see 23).
#
# `--hook-mode` forces exit 0 even on unresolved leaves. Without it the
# CLI's intrinsic exit 1 would propagate, and Claude would treat the
# Stop hook as having failed for every interesting review.
exec debate \
  --hook-mode \
  --session-id "$SESSION_ID" \
  --transcript "$TRANSCRIPT" \
  --max-turn 6
```

Properties (verified by [25](25-probes.md)):

- Three positional args extracted from the JSON payload: `session_id`, `transcript_path`, `cwd`.
- Recursion guard checks `DEBATE_IN_PROGRESS` first thing.
- `unset ANTHROPIC_API_KEY` runs before `exec`.
- `cd "$CWD"` runs before `exec`.
- The script writes nothing to stdout (no JSON `systemMessage` etc. - pollution per [01-overview.md](01-overview.md) §"Constraints uncovered").
- Uses `exec` so the orchestrator inherits stdout/stderr.

## Install paths

### User-level (recommended)

`~/.claude/settings.json`:

```jsonc
{
  "hooks": {
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          { "type": "command", "command": "/path/to/debate-stop-hook.sh" }
        ]
      }
    ]
  }
}
```

The verbose form (with `matcher` and a `hooks` array) is required; the simpler `{"command":"..."}` form is silently dropped from claude's hook registry (see [01-overview.md](01-overview.md) §"Stop hook output channels").

### Project-level

`./.claude/settings.json` with the same shape. claude requires a one-time interactive trust prompt for project-level hooks. Suitable for repo-pinned configs; requires the `debate` binary on `$PATH` for everyone who clones the repo.

## CLI subcommand for install

`debate install-hook` (added to [04](04-cli-flags.md)'s command tree as a separate cobra subcommand - but flag-only no-config dependency, so cleanly bolts on):

```
debate install-hook [--scope user|project] [--script-path <path>]
debate uninstall-hook [--scope user|project]
```

Behavior:

- `install-hook --scope user` (default) merges the verbose-format Stop hook entry into `$HOME/.claude/settings.json`, creating the file if absent. If a Stop hook entry already exists with a different command, prints a diff and asks for `--force`.
- `install-hook --scope project` does the same for `./.claude/settings.json`.
- `--script-path` defaults to the absolute path of `debate-stop-hook.sh` next to the `debate` binary; falls back to `$(which debate-stop-hook.sh)`.
- `uninstall-hook` removes any Stop entry whose `command` ends in `debate-stop-hook.sh`.

These subcommands edit user-owned config files and therefore use `state.AtomicWrite` ([09](09-state-dir.md)) semantics (temp file + rename + parent fsync) to avoid corrupting `settings.json`.

`debate install-hook` is *not* available in the auto-trigger path; users invoke it explicitly.

## stdout / log.jsonl

The hook script never writes to stdout. The orchestrator's stdout is the *only* user-facing channel (per [23](23-summary-render.md)'s `SurfacingDecision`). Because stdout rendering during a Stop hook is best-effort, the canonical pointer is `.debate/log.jsonl`'s last `kind:"run"` line.

## Settings.json merge rules

JSON-aware merge (not text-level append). The installer parses `settings.json`, walks `hooks.Stop`, ensures exactly one entry whose nested `hooks[*].command` ends with `debate-stop-hook.sh`, and writes the result back atomically.

Idempotent: running `install-hook` twice leaves the file in the same state as running once.

## Test contract

- Unit: hook script's recursion guard exits 0 when `DEBATE_IN_PROGRESS=1`.
- Unit: `install-hook` against an empty file produces a valid verbose-format hook entry.
- Unit: `install-hook` against a file with an unrelated existing Stop hook merges without overwriting.
- Unit: `install-hook` is idempotent.
- Integration ([25](25-probes.md)): a real `claude -p` invocation that triggers the installed hook with `DEBATE_IN_PROGRESS` set returns within 100ms and emits no debate-content into root JSONL.

## Acceptance criteria

- [x] `debate-stop-hook.sh` ships in the release tarball ([03](03-ci-lint-release.md)).
- [x] `debate install-hook --scope user` installs and `debate uninstall-hook --scope user` cleanly removes.
- [x] Recursion guard works (the canonical "if X then exit 0" sequence is the *first* statement after `set -e`).
- [x] No stdout from the hook script during a normal run.
- [x] settings.json edits are atomic (test crash between write and rename leaves the prior file intact).
