# Spec 16 - Subprocess infrastructure

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) §"Constraints uncovered by the probe" for design intent.

**Depends on:** [02](02-go-module.md), [04](04-cli-flags.md), [06](06-preflight.md).
**Consumed by:** [17](17-claude-proposer.md), [18](18-critic-drivers.md), [21](21-signals.md).

## Scope

In: shared subprocess primitives - environment scrubbing, JSON output decoding tolerant to control characters, deadlines, child-tree teardown on signal, recursion-guard env propagation, structured stderr capture.

Out: any specific agent's invocation arguments ([17](17-claude-proposer.md), [18](18-critic-drivers.md)).

## Public Go interfaces

```go
// internal/agent/exec.go
package agent

type Run struct {
    Bin        string            // absolute path; resolved via exec.LookPath
    Args       []string          // does NOT include Bin
    Cwd        string            // absolute; required
    Stdin      []byte            // optional
    Env        []string          // base env; helpers below modify a copy
    Deadline   time.Duration     // 0 = no deadline
}

type Result struct {
    Stdout    []byte
    Stderr    []byte
    ExitCode  int
    Duration  time.Duration
    Killed    bool              // true if context-cancelled / deadline hit
}

// Exec runs the command. Cancellation propagates to the entire process
// group (setpgid on Linux, posix_spawn-style on Darwin). On context
// cancellation it sends SIGINT, waits 2s, then SIGKILL.
func Exec(ctx context.Context, r Run) (Result, error)

// CleanEnv returns a copy of os.Environ with the named keys removed
// and AGON_IN_PROGRESS=1 added.
func CleanEnv(remove ...string) []string

// DecodeJSONLine decodes one JSON object out of a byte slice that may
// contain unescaped control characters in string fields. Falls back to
// json.Decoder with NewDecoderUseNumber when the standard parse fails.
func DecodeJSONLine(line []byte, dst any) error

// StreamJSON reads stdout line-by-line and emits one decoded value per
// line via the visit function. Used by codex's JSON event stream.
func StreamJSON(stdout io.Reader, visit func(json.RawMessage) error) error
```

## Env scrubbing rules

`CleanEnv` always:

- Removes `ANTHROPIC_API_KEY` (see [01-overview.md](01-overview.md) §"Constraints uncovered").
- Sets `AGON_IN_PROGRESS=1` (recursion guard contract - see [06](06-preflight.md), [24](24-stop-hook.md)).
- Sets `LC_ALL=C` for stable command output where it matters ([08](08-diff.md) git wrapper).
- Preserves `HOME`, `PATH`, `XDG_*`, `CLAUDE_*` (other than the API key), `CODEX_*`.

Optional removals via the `remove ...string` parameter for caller-specific cleanup.

## Cancellation and process group

Linux:

```go
cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
```

On context cancel: `syscall.Kill(-pgid, syscall.SIGINT)` then a 2-second timer; on timeout `syscall.Kill(-pgid, syscall.SIGKILL)`.

Darwin: same `Setpgid` field is supported; the kill loop is identical.

Windows: not supported in v0 (the Stop hook is a bash script anyway). The build still targets `windows/amd64` for the binary so `agon --version` works in CI; subprocess-cancellation tests skip on Windows.

`Result.Killed = true` when the process exited because of context cancellation (not when it self-exited with non-zero).

## JSON decoding tolerance

`DecodeJSONLine` strategy:

1. Try `json.Unmarshal(line, dst)`. If it succeeds, return.
2. On error, sanitize: replace bytes `< 0x20` (except `\t`, `\n`, `\r`) inside JSON string contexts with `\u00XX` escapes. Implementation: a single-pass bytes-level rewriter.
3. Retry `json.Unmarshal` on the sanitized buffer.
4. If still failing, return the original error wrapped with a hex dump of the first 256 bytes.

`StreamJSON` strategy:

- `bufio.Scanner` with `Split(bufio.ScanLines)` and a 1MB buffer (max codex event size, conservative).
- Each line passed through `DecodeJSONLine`.
- Empty lines and lines that fail decode after sanitization are skipped at `--verbose >= 2` and counted; the visit function is not called for them.

## Deadlines and stalls

Each `Run.Deadline` defaults are set by callers ([17](17-claude-proposer.md): 5min per claude call; [18](18-critic-drivers.md): 5min per critic call). `0` = inherit ctx only.

If a subprocess produces no stdout for 60 seconds, the orchestrator emits a stderr line at `--verbose >= 1`:

```
agon: <agent>: no output for 60s, still waiting (pid <n>)
```

This is a heartbeat, not a kill - only context cancellation kills.

## Stderr handling

Subprocess stderr is captured into `Result.Stderr` and *also* mirrored to the orchestrator's stderr at `--verbose >= 2`. At `--verbose < 2` it's silent (claude/codex stderr is noisy by default).

## Recursion-guard propagation

Every subprocess inherits `AGON_IN_PROGRESS=1` (set by `CleanEnv`). The Stop hook script ([24](24-stop-hook.md)) checks this env on entry and exits 0 if set, breaking the recursion that would otherwise occur when the proposer-clone's claude subprocess fires the Stop hook on its own completion.

## Test contract

- Unit: `CleanEnv` removes `ANTHROPIC_API_KEY` and sets `AGON_IN_PROGRESS=1`.
- Unit: `Exec` against `sleep 5` cancels via context within 100ms of cancellation; `Result.Killed == true`.
- Unit: `Exec` against a child that spawns its own child propagates SIGINT to both (process-group test).
- Unit: `DecodeJSONLine` succeeds against a payload with `\x00` and `\x01` inside a string field.
- Unit: `StreamJSON` against a fixture with mixed valid+invalid lines emits only valid ones.

## Acceptance criteria

- [x] No subprocess invocation in [17](17-claude-proposer.md) or [18](18-critic-drivers.md) calls `os/exec` directly; all go through `agent.Exec`.
- [x] Process-group teardown verified by an integration test that spawns a child-of-child and asserts both die on cancellation.
- [x] JSON sanitizer's pass-through rate ≥ 99.9% on a 10k-line replay of real claude output (perf bound, not correctness).
- [x] `AGON_IN_PROGRESS=1` is observable in every spawned process's `/proc/self/environ` (Linux integration test).
