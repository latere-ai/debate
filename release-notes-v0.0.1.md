# debate v0.0.1

Release-cut evidence for `v0.0.1`. See `specs/27-release.md` for the gate
process. The auto-generated GitHub release body covers code changes; this
file covers release-cut evidence (probe outcomes, smoke recordings) per
`specs/35-release-notes-channel.md`.

Each block below is appended verbatim by the corresponding follow-up
spec when its gate runs. Empty blocks are placeholders to be filled in
before the GA tag is pushed.

## Probe outcomes

### no-output-stop-hook (G4) - specs/28

```
probe: no-output-stop-hook
claude_version: 2.1.131 (Claude Code)
host_os: darwin (Darwin 25.4.0, arm64)
exit_code: 0
verdict: PASS
hook_attachments: (none)
before_sha: 4335582e7211214047966a9332ba6cbb28ce970bce36b7cfb0f5bc6186beedd2
after_sha:  e6bfaa0437e5b00b4499ab59f1e3bff0a1f3177d5b9bbe554caefe0c48a32746
size_delta: 20913 -> 23070   # growth is from new user/assistant turns, not hook attachments
```

Notes: PASS branch wins. A no-output Stop hook does not write a
`hook_*` attachment into root JSONL. Spec 01 §"v0 release blockers"
and §"Lifecycle invariants" tightened accordingly: the byte-identical
"no debate-content pollution" invariant holds across all modes
including Option B (Stop hook). Probe required two side-fixes:
- Use `pwd -P` to get the canonical path on macOS (where /tmp is a
  symlink into /private/tmp); the original probe used logical pwd and
  could never find claude's session JSONL.
- Redirect stdin from /dev/null on `claude --print` invocations to
  suppress the "no stdin data received in 3s" warning.

### signal-latency (G5) - specs/29

```
probe: signal-latency
host_os: darwin (Darwin 25.4.0, arm64)
runs: 3/3
wall_seconds: {min: 0.031, median: 0.036, max: 0.048}
budget: 5.000
verdict: PASS
```

Notes: the probe was originally broken in two ways - it never invoked
`bin/debate` (used a self-contained bash sleeper) and had a trap-vs-sleep
race that always took 30s. Rewriting the probe to actually exercise
`bin/debate` against a sleep-forever shim surfaced the underlying bug:
`cmd/debate/main.go` used cobra's default `context.Background()`, so the
`round.InstallHandler` signal handler in `internal/round/signals.go` was
dead code and SIGINT did not propagate. Fix: wire `round.InstallHandler`
into main and pass the resulting context via `root.ExecuteContext(ctx)`.
A new e2e regression test (`e2e/cli/signal_test.go`) builds the binary,
spawns it against a stuck shim, sends SIGINT to its process group, and
asserts exit within 2s.

### trivial-diff-perf (G6) - specs/30

```
probe: trivial-diff-perf
host_os: darwin (Darwin 25.4.0)
host_cpu: arm64
binary_sha256: 058ed50b44ebecae6457819c41535d35443278e74ec6aa18a8087835e9c5ba7a
wall_ms: {min: 84, median: 99, max: 136}    # over 3 batches × 3 runs each
budget_ms: 200
verdict: PASS
```

Notes: original probe had a 200 ms threshold; spec 25 line 168 had a
stale 100 ms claim. Reconciled to 200 ms since four short-lived git
subprocess calls plus cobra startup are inherently 80-150 ms on the
fastest hardware. Spec 01 §UX retains <100 ms as the aspirational
user-facing wording; the probe asserts the realistic ceiling.

### interactive-stdout (G7) - specs/31

```
probe: interactive-stdout
host_os: darwin (Darwin 25.4.0, arm64)
claude_version: 2.1.131 (Claude Code)
verdict: SKIP
captured_line_present: n/a
skip_reason: probe is documented as manual-only; automating the claude
             TUI under script(1)/expect would couple the probe to
             claude's interactive-prompt format, which is unstable
             across versions. Non-blocking gate per spec 31.
```

Notes: spec 31 marks this gate as non-blocking. The README and spec 01
keep the conservative "stdout best-effort" wording. Re-run before the
next release; if interactive rendering becomes a contractual
expectation users rely on, automate with a stable claude harness.

## Test gates

### real-e2e (G13) - specs/32

```
gate: real-e2e
host_os: TBD
claude_version: TBD
codex_version: TBD
verdict: TBD
```

## UX gates

### install-hook smoke (G15) - specs/33

```
gate: install-hook-smoke
host_os: darwin (Darwin 25.4.0, arm64)
binary_sha256: 6426377b35078e14815c27172481392bc6e8a9d60db5609500f12c5fa0ab4e72
fresh_install_valid: yes
generated_command: /Users/changkun/dev/changkun.de/debate/scripts/debate-stop-hook.sh
idempotent_second_run: yes
verdict: PASS
```

Notes: command path resolves at install time to the in-repo
`scripts/debate-stop-hook.sh`. In a release tarball install, this would
resolve to the script shipped alongside the binary instead. The
verbose-format Stop hook entry has `matcher: ""` and a single
command-type hook, matching [24-stop-hook.md](specs/24-stop-hook.md).

### real-claude end-to-end (G16) - specs/34

```
gate: real-claude-end-to-end
host_os: TBD
claude_version: TBD
codex_version: TBD
session_dir: TBD
termination: TBD
forks: TBD
max_per_fork_wall: TBD
verdict: TBD
```
