# Spec 17 — Claude proposer driver

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Mechanism" → "Forking the proposer" / "Continuing within a fork" for design intent.

**Depends on:** [04](04-cli-flags.md), [07](07-claude-transcript.md), [11](11-fork-artifacts.md), [16](16-subprocess-infra.md).
**Consumed by:** [19](19-round-loop.md).

## Scope

In: driving the proposer-clone via `claude --resume`. R1 forks off the root with `--fork-session` and delivers the first critic-pointer message in one shot; R3, R5, … resume the fork. Captures session id, response text, working-tree changes. Maps every documented failure mode (`401 Unauthorized`, "no conversation found", control-char output, etc.) onto typed errors.

Out: codex- or claude-as-critic invocation ([18](18-critic-drivers.md)), prompt assembly for the *critic* ([15](15-aspect-prompts.md)).

## Public Go interfaces

```go
// internal/agent/claude_proposer.go
package agent

type ClaudeProposer struct {
    Bin       string             // resolved claude path; "" = LookPath("claude")
    Cwd       string              // must equal Plan.Cwd; --resume is cwd-scoped
    RootID    string              // root claude session id
    Model     string              // "" = CLI default
    Deadline  time.Duration       // per-call default 5m
}

type ProposerResult struct {
    ForkID         string         // captured on R1 from --fork-session JSON
    Response       string         // claude's text reply (defense)
    Tokens         int            // sum of input + output tokens (if reported)
    USD            float64        // best-effort cost figure (0 if unavailable)
    Stdout         []byte         // raw, retained for debugging at --verbose >= 2
    ChangedFiles   []string       // working-tree files touched during this round
    Duration       time.Duration
}

// FirstRound creates a fork off the root session and processes R1 in
// one shot. Equivalent of:
//   claude --resume <root> --fork-session --output-format json -p "<pointer>"
//
// Returns ProposerResult with ForkID populated.
func (p *ClaudeProposer) FirstRound(ctx context.Context, pointer string) (*ProposerResult, error)

// NextRound continues an existing fork:
//   claude --resume <fork-id> --output-format json -p "<pointer>"
//
// ForkID in the result equals forkID (passed through unchanged).
func (p *ClaudeProposer) NextRound(ctx context.Context, forkID, pointer string) (*ProposerResult, error)

// ApplyDefaultModel returns a copy with Model defaulted from the agent's
// own CLI default when caller passed "".
func (p *ClaudeProposer) ApplyDefaultModel() *ClaudeProposer
```

Errors:

- `ErrCwdMismatch` — `claude --resume` returned "no conversation found"; caller likely violated the cwd-scope rule.
- `ErrAuth` — exit code from claude indicates 401; `ANTHROPIC_API_KEY` was set despite [16](16-subprocess-infra.md)'s `CleanEnv` (defensive).
- `ErrTimeout` — context deadline exceeded.
- `ErrJSON` — output was not parseable JSON even after sanitization.
- `ErrEmptyResult` — JSON parsed but `result` field was empty.

## Command construction

`FirstRound`:

```
claude --resume <RootID> --fork-session --output-format json --print <pointer>
       [--permission-mode <mode>]            # only if explicitly requested via env
       [--model <Model>]                      # if non-empty
```

`--print` is the same as `-p`; using the long form for grep-ability in shell scripts.

`NextRound`:

```
claude --resume <forkID> --output-format json --print <pointer>
       [--model <Model>]
```

The pointer is the verbatim user-message text (file-pointer channel from [01-overview.md](01-overview.md)). Caller assembles this string ([19](19-round-loop.md)).

Cwd: `p.Cwd` is set in `Run.Cwd`. `LC_ALL=C` is set by [16](16-subprocess-infra.md)'s `CleanEnv`.

## JSON output shape

claude's `--output-format json` returns a single JSON document on stdout:

```jsonc
{
  "type":         "result",
  "subtype":      "success",
  "session_id":   "5a8c9b1e-...",       // for FirstRound this is the NEW fork id
  "result":       "<claude's chat response, including any code blocks>",
  "is_error":     false,
  "total_cost_usd": 0.0123,             // present in newer versions; tolerate absence
  "usage": {
    "input_tokens": 32011,
    "output_tokens": 412
  }
}
```

`FirstRound` reads `session_id` into `ForkID`. `NextRound` checks `session_id` matches the input `forkID`; mismatch is `ErrUnexpectedFork`.

`Tokens = usage.input_tokens + usage.output_tokens` (if both present; 0 otherwise).

`USD = total_cost_usd` if present; 0 otherwise.

## Working-tree change detection

Before each call, the orchestrator captures `git status --porcelain` ([11](11-fork-artifacts.md)'s helper). After the call returns, it captures again and diffs the two lists to populate `ChangedFiles`. Files only touched (mtime change but no content change) are excluded.

## Failure mode mapping

| claude exit / output | Mapped error |
|---|---|
| `Error: No conversation found with session ID` on stderr | `ErrCwdMismatch` |
| Exit 1 + `Authentication error: 401` on stderr | `ErrAuth` |
| Context cancelled / deadline | `ErrTimeout` (with `Result.Killed`) |
| Stdout not valid JSON after sanitization | `ErrJSON` |
| `is_error: true` in JSON | `ErrAgentError(subtype, result)` |
| Empty `result` field | `ErrEmptyResult` |

Each typed error wraps the raw `agent.Result` for debugging.

## Behavior

- `Bin` resolves once at construction (or first call); reuse the cached path.
- The `--permission-mode` flag is only added when env `DEBATE_PERMISSION_MODE` is set; v0 default omits it.
- ANSI escapes in `result` (e.g., from claude's tool-output rendering) are not stripped — the proposer's chat reply is captured verbatim and persisted to `r<n>-proposer.md` ([11](11-fork-artifacts.md)).
- The driver does *not* parse the response (no attack-shape parsing, no JSON-inside-result extraction). It returns raw text; downstream readers (`r<n>-proposer.md` + the next critic round) consume it as-is.

## Test contract

- Unit (with mock `claude`): `FirstRound` extracts `ForkID` from a fixture JSON.
- Unit: `NextRound` round-trips the same fork id.
- Unit: a fixture with `is_error: true` returns `ErrAgentError` with the subtype.
- Unit: stdout containing control-char `result` field decodes via [16](16-subprocess-infra.md)'s sanitizer.
- Integration: against real `claude` (gated by `RUN_AGENT_TESTS=1`), a known-good prompt returns a non-empty result.

## Acceptance criteria

- [x] All five typed errors reachable.
- [x] `--fork-session` is *always* in `FirstRound`'s args; never in `NextRound`'s.
- [x] `claude --resume <root>` (without `--fork-session`) is absent from this package's call sites; lint test grep verifies.
- [x] `--cwd` set on `Run.Cwd`; integration test asserts cwd-scoped failures map to `ErrCwdMismatch`.
- [x] `Tokens` accumulator usable by [20](20-termination.md)'s cost-cap gate.
