# Probes

Standalone scripts that verify environment-dependent invariants
against a real claude / codex install. See [specs/25-probes.md].

| Script | Purpose | GA gate? |
|---|---|---|
| `no-output-stop-hook.sh` | Does a no-output Stop hook write `hook_*` attachments to root JSONL? | yes (G4) |
| `signal-latency.sh` | SIGINT-to-exit latency on a stuck child < 5s. | yes (G5) |
| `trivial-diff-perf.sh` | Trivial-diff exit-fast wall < 200ms. | yes (G6) |
| `interactive-stdout.sh` | Does a Stop hook's stdout render in interactive claude? | manual (G7) |

## Run all

```
make probe
```

`make probe` exits 0 iff all non-skipped probes pass. SKIP is fine
for the optional probe.

## Environment

- `KEEP=1` keeps the temp dirs probes create.
- `DEBATE_BIN=/path/to/debate` overrides the binary location for
  `trivial-diff-perf.sh`.
