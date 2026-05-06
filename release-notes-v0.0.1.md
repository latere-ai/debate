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
claude_version: TBD
host_os: TBD
exit_code: TBD
verdict: TBD
hook_attachments: TBD
```

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
host_os: TBD
host_cpu: TBD
binary_sha256: TBD
wall_ms: TBD
verdict: TBD
```

### interactive-stdout (G7) - specs/31

```
probe: interactive-stdout
host_os: TBD
claude_version: TBD
verdict: TBD
captured_line_present: TBD
skip_reason: TBD
```

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
