# Spec 33 - install-hook smoke verification

> **Status: ❌ RETIRED (2026-05-16).** `install-hook` was removed
> with the Stop hook; this smoke no longer applies. Kept as
> historical record. Rationale: [36](36-probe-userpromptsubmit-manual-trigger.md).

**Depends on:** [24](24-stop-hook.md).
**Consumed by:** [27](27-release.md).

## What we're proving

[27-release.md](27-release.md) G15: `agon install-hook --scope user` against a fresh `~/.claude/settings.json` produces a valid verbose-format Stop hook entry. This must be exercised on the release-candidate binary, not the working-tree binary, and not on a settings file that already has hooks (since merge behaviour is the more interesting case but is **not** what G15 asks for).

A second, narrower check is also worth running and recording: install-hook is idempotent on a settings file that already contains the same hook entry (per [24-stop-hook.md](24-stop-hook.md)). This is recorded but does not block GA.

## Execution

1. `make build`.
2. Use a throwaway `$HOME` directory to avoid touching the real settings:
   ```
   FAKE_HOME=$(mktemp -d)
   HOME=$FAKE_HOME ./bin/agon install-hook --scope user
   cat $FAKE_HOME/.claude/settings.json
   ```
3. Inspect the generated `settings.json`:
   - Top-level `hooks.Stop` array exists.
   - Exactly one entry under `hooks.Stop` whose `command` references `agon-stop-hook.sh` (or `bin/agon run --hook-mode`, depending on what install-hook writes).
   - JSON is valid (`jq . settings.json` exits 0).
4. Idempotency check (informational, not blocking):
   ```
   HOME=$FAKE_HOME ./bin/agon install-hook --scope user   # second run
   jq '.hooks.Stop | length' $FAKE_HOME/.claude/settings.json   # should still be 1
   ```

## Recording format

```
gate: install-hook-smoke
host_os: darwin|linux
binary_sha256: <sha of bin/agon>
fresh_install_valid: yes|no
generated_command: <the exact `command` string in the hook entry>
idempotent_second_run: yes|no|not_tested
```

## Disposition

- **fresh_install_valid: yes** - G15 cleared, recording cited in [27-release.md](27-release.md).
- **fresh_install_valid: no** - GA blocked. Bug is in [24-stop-hook.md](24-stop-hook.md) implementation; fix and re-run.
- **idempotent_second_run: no** - does not block GA but must be filed as a v0.0.x patch issue with the recording attached.

## Out of scope

- Settings-file *merge* behaviour against an existing user settings.json with unrelated hooks. That is more interesting than G15, and worth covering, but it belongs to an integration test, not to this gate.
- `--scope project` install-hook verification. Same shape; would just be a second smoke. Not what [27-release.md](27-release.md) asks for.

## Acceptance criteria

- [x] Smoke ran against the release-candidate binary in a throwaway `$HOME`.
- [x] `fresh_install_valid: yes` and the generated command string is recorded verbatim.
- [x] Idempotent second-run check: second invocation does not duplicate the Stop entry (length stays at 1).
- [x] ~~[27-release.md](27-release.md) G15 cites the recording.~~ *(retracted: G15 no longer exists as a release blocker.)*
