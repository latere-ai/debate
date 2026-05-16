# Spec 18 - Critic drivers (codex + claude)

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) §"Mechanism" → "Driving the critic" for design intent.

**Depends on:** [13](13-critic-output-format.md), [15](15-aspect-prompts.md), [16](16-subprocess-infra.md).
**Consumed by:** [19](19-round-loop.md).

## Scope

In: stateless-per-round critic invocation for codex and claude. Prompt assembly (system + task + diff + prior round files), JSON event parsing for codex, plain JSON parsing for claude.

Out: parsing the critic's output ([14](14-attack-parser.md)), choosing the aspect prompt ([15](15-aspect-prompts.md)).

## Public Go interfaces

```go
// internal/agent/critic.go
package agent

type Critic interface {
    // Round runs one critic round and returns the raw markdown output.
    // The orchestrator passes this to [14] for parse+normalize+persist.
    Round(ctx context.Context, in CriticInput) (*CriticResult, error)
}

type CriticInput struct {
    Aspect            critic.Aspect       // from [15]
    CriticIndex       int
    Round             int                 // odd: 1, 3, 5, ...
    SystemPrompt      string              // pre-assembled by [15]'s Assemble
    TaskContext       string              // verbatim first user turn
    DiffPatch         string              // unified diff (per-fork snapshot from [11])
    PriorRoundFiles   []RoundFileRef      // R3+: pointers to prior-round files
    Cwd               string              // absolute; sandbox dir or repo cwd
    Deadline          time.Duration
}

type RoundFileRef struct {
    Path  string  // absolute path or relative to Cwd
    Round int
    Role  string  // "critic" | "proposer"
}

type CriticResult struct {
    Markdown   string         // raw stdout - the critic's emitted document
    ThreadID   string         // codex: from thread.started; claude: session_id
    Tokens     int
    USD        float64
    Stdout     []byte         // raw, retained at --verbose >= 2
    Duration   time.Duration
}
```

Two implementations:

- `CodexCritic` (the cross-family default).
- `ClaudeCritic` (used in same-family `claude/claude` mode).

Both satisfy `Critic`. The orchestrator selects via `agent.NewCritic(family string, ...)`.

## CodexCritic

```
codex exec --skip-git-repo-check --sandbox read-only --json
```

Args appended:

- `--cd <CriticInput.Cwd>` if codex supports; else `Run.Cwd` is used.
- `--model <name>` if `--side-model` is set.

Stdin: not used. Prompt is the single positional argument.

Prompt assembly (single string):

```
<CriticInput.SystemPrompt>

<format reminder from [13]>

# Task
<TaskContext>

# Diff
```diff
<DiffPatch>
```

# Prior rounds (round 3+)
- @<rel-path-of-r2-proposer.md> - proposer's R2 defense
- @<rel-path-of-r3-critic.md>   - your R3 attack list (your own prior round)
...
```

Codex's file-access tool resolves `@<path>` references; the orchestrator uses absolute paths or paths relative to `Cwd`.

Output parsing: codex emits a JSON event stream on stdout, one event per line. Relevant events:

```jsonc
{"type":"thread.started","thread_id":"<uuid>"}
{"type":"item.completed","item":{"id":"...","type":"agent_message","content":"<the markdown>"}}
{"type":"thread.completed", ...}
```

`Round` consumes the stream via [16](16-subprocess-infra.md)'s `StreamJSON`:

1. Capture `thread_id` from the first `thread.started`.
2. Concatenate `content` from every `item.completed` with `type == "agent_message"`.
3. On `thread.completed`, finalize.

`Markdown = <concatenated content>`. Cost/usage best-effort from a `usage` field if present.

## ClaudeCritic

```
claude --output-format json --print <full prompt>
       [--model <side-model>]
       [--permission-mode <mode>]   # only if env requests it
```

**No `--resume`, no `--fork-session`.** Each round is a fresh session. Reasoning: same as in [01-overview.md](01-overview.md) - freshness blocks the critic from inheriting any other session's conversation.

Prompt: same string as `CodexCritic`. Claude reads `@<path>` references via its built-in file tool; paths are absolute or relative to `Run.Cwd`.

Output: same JSON shape as [17](17-claude-proposer.md):

```jsonc
{"type":"result","session_id":"...","result":"<the markdown>","usage":{...}}
```

`Markdown = result`. `ThreadID = session_id` (per-round, not stable).

## Cwd policy

In v0, both critic drivers run with `Cwd` set to the repo cwd (the same as the proposer's). This is "best-effort isolation" per [01-overview.md](01-overview.md) §"Critic isolation".

Hook for v1: when [01-overview.md](01-overview.md)'s strict-isolation v1 lands, `CriticInput.Cwd` will point at a per-fork sandbox dir; this driver's behavior is unchanged (the prompt's `@<path>` references will become relative to the sandbox).

## Failure mode mapping

| Symptom | Error |
|---|---|
| codex exit + stderr "rate limit" | `ErrRateLimit` |
| codex exit + stderr "Error: stdin is not a terminal" | `ErrTTYRequired` (defensive; `exec` mode shouldn't hit this) |
| Empty `agent_message` content (codex) or empty `result` (claude) | `ErrEmptyResult` |
| Context cancelled / deadline | `ErrTimeout` |
| Stdout/stream not parseable after sanitization | `ErrJSON` |

## Behavior

- The drivers do *not* parse the markdown - they return raw stdout. Parsing is [14](14-attack-parser.md).
- `Tokens` accumulator usable by [20](20-termination.md)'s cost-cap gate; for codex the field comes from `usage` events when present, else 0 (best-effort).
- For codex, `--sandbox read-only` is *always* set; lint test verifies.

## Test contract

- Unit (mock codex): JSON event stream with two `agent_message` events concatenates into a single markdown block.
- Unit (mock claude): JSON `result` extracted into `Markdown`.
- Unit: empty `agent_message` returns `ErrEmptyResult`.
- Unit: rate-limit substring in codex stderr returns `ErrRateLimit`.
- Integration (gated `RUN_AGENT_TESTS=1`): real codex against a 5-line diff produces non-empty markdown.

## Acceptance criteria

- [x] Both `CodexCritic` and `ClaudeCritic` satisfy the `Critic` interface.
- [x] `--sandbox read-only` is always present on codex args (lint test grep).
- [x] `--resume` and `--fork-session` are absent from `ClaudeCritic` args.
- [x] Prompt assembly is deterministic (no clock, no rand) given the same `CriticInput`.
- [x] `agent.NewCritic("codex", ...)` returns `*CodexCritic`; `"claude"` returns `*ClaudeCritic`; other family panics.
