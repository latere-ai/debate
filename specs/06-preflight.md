# Spec 06 — Pre-flight checks

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"CLI surface" notes for design intent.

**Depends on:** [04](04-cli-flags.md), [05](05-config-file.md).
**Consumed by:** every spec downstream of CLI parsing — pre-flight is the gate before any agent process spawns or any state-dir file is written.

## Scope

In: every check that must pass after flags+config merge and before [09](09-state-dir.md) creates a session folder. Cwd, env hygiene, recursion guard, family/model/arity validation, judge mode gate, codex-as-proposer rejection. Every error code and exit message.

Out: actual session-folder creation ([09](09-state-dir.md)), subprocess invocation, prompt assembly.

## Public Go interfaces

```go
// internal/cli/preflight.go
package cli

// Preflight runs every pre-flight check against the effective Flags.
// On success it returns a *Plan summarizing what the run will do
// (number of forks, aspect-to-critic mapping, resolved cwd, etc).
// On failure it returns a *PreflightError that carries the exit code
// the CLI should propagate.
func Preflight(ctx context.Context, f *Flags) (*Plan, error)

type Plan struct {
    Cwd          string                // resolved absolute path
    Forks        []ForkPlan            // len == SideCount, indexed 1..N
    HookMode     bool
    StateDirAbs  string                // absolute --state-dir
    SessionID    string                // root claude session-id, or "" when manual
    TranscriptPath string
}

type ForkPlan struct {
    Index  int     // 1-based
    Aspect string  // one of f.Aspect
}

type PreflightError struct {
    Code int       // exit code
    Msg  string    // human-readable; goes to stderr
    Wrap error     // optional underlying error
}
```

## Checks (ordered)

Each check runs in this order; the first failure short-circuits with the listed exit code and message.

### 1. Recursion guard

If `os.Getenv("DEBATE_IN_PROGRESS") != ""`, the orchestrator was re-entered through the Stop hook firing on its own subprocesses. **Exit 0** with no message; this is the fast path that protects against infinite-fork recursion (see [01-overview.md](01-overview.md) §"Recursion guard contract" and [24](24-stop-hook.md)).

The hook script itself also checks this env; the duplicate check inside the orchestrator covers the case where someone invokes `debate` manually inside a debate-spawned subshell.

### 2. cwd resolution

`Plan.Cwd = filepath.Abs(getwd)`. If the `Flags.Transcript` argument resolves to a path *not* under (or equal to) the cwd that owns the session (claude `--resume` is cwd-scoped, see [01-overview.md](01-overview.md)), `debate` exits **101** with:

```
debate: --transcript points at a session whose cwd does not match the current directory; cd to <expected> and retry
```

Determining the session cwd: walk up the transcript path against `~/.claude/projects/<encoded-cwd>/...`; the encoded segment between `projects/` and the next `/` decodes to the cwd ([07](07-claude-transcript.md) owns the encoding).

### 3. Env hygiene

If `os.Getenv("ANTHROPIC_API_KEY") != ""`, `os.Unsetenv` it and emit a single line on stderr at `--verbose >= 1`:

```
debate: unset stale ANTHROPIC_API_KEY for this run (claude OAuth keychain will be used)
```

Not an error. (See [01-overview.md](01-overview.md) §"Constraints uncovered by the probe".)

### 4. v0 mode gates

| Condition | Exit | Message |
|---|---|---|
| `Main == "codex"` | 102 | `--main codex is v1; v0 supports --main claude only` |
| `Judge != "none"` | 103 | `--judge llm/human is v1; v0 supports --judge none only` |

### 5. Family/model rule

(See [01-overview.md](01-overview.md) §Heterogeneity.)

Same family means `agentFamily(Main) == agentFamily(Side)` where `agentFamily("claude") = "claude"` and `agentFamily("codex") = "codex"`.

| Condition | Exit | Message |
|---|---|---|
| Same family AND (`MainModel == ""` OR `SideModel == ""`) | 110 | `--main and --side are the same family; both --main-model and --side-model are required and must differ` |
| Same family AND `MainModel == SideModel` (after normalization) | 111 | `--main-model and --side-model must differ when --main and --side are the same family` |

Cross-family: `MainModel`/`SideModel` are optional and ignored if blank (each agent's CLI default applies).

### 6. Side-count vs aspect arity

| Condition | Exit | Message |
|---|---|---|
| `SideCount != len(Aspect)` | 120 | `--side-count (N) must equal len(--aspect) (M)` |
| `SideCount < 1` | 121 | `--side-count must be ≥ 1` |
| `MaxTurn < 2` | 122 | `--max-turn must be ≥ 2 (one attack + one defense minimum)` |
| `CostCap < 1` | 123 | `--cost-cap must be ≥ 1` |
| `ChangedLinesMin < 0` | 124 | `--changed-lines-min must be ≥ 0` |

`Plan.Forks[i] = ForkPlan{Index: i+1, Aspect: Aspect[i]}`. Order is preserved; aspect names are not deduplicated (caller's responsibility).

### 7. Task-context source

If neither `SessionID` nor `Transcript` nor `TaskContext` is set, exit **130** with:

```
debate: cannot determine task context: pass --session-id, --transcript, or --task-context
```

When `SessionID` is set but `Transcript` is empty, [07](07-claude-transcript.md) computes the transcript path; pre-flight does not check file existence here (that's [07](07-claude-transcript.md)'s concern).

### 8. State dir writability

`Plan.StateDirAbs = filepath.Abs(StateDir)`. If the parent dir is not writable, exit **140**:

```
debate: cannot write under <parent>: <os error>
```

Pre-flight does not create the dir (see [09](09-state-dir.md)); it only stat-checks the parent.

### 9. .gitignore advisory

If cwd is inside a git repo and the repo's `.gitignore` does not list `.debate/` (or a parent thereof), emit a warning to stderr (always, regardless of `--verbose`):

```
debate: warning: .debate/ is not in .gitignore — consider adding it before committing
```

Not an error. Does not auto-edit ([01-overview.md](01-overview.md) §`.gitignore`).

## Exit codes (full table)

| Code | Meaning |
|---|---|
| 0 | success (or recursion-guard short-circuit, or `--help`/`--version`) |
| 1 | `≥ 1` unresolved leaves (see [23](23-summary-render.md)); not a pre-flight code, listed for completeness |
| 101 | cwd mismatch |
| 102 | codex-as-proposer not in v0 |
| 103 | judge mode not in v0 |
| 110 | missing same-family models |
| 111 | identical same-family models |
| 120 | side-count vs aspect arity |
| 121 | side-count < 1 |
| 122 | max-turn < 2 |
| 123 | cost-cap < 1 |
| 124 | changed-lines-min < 0 |
| 130 | no task-context source |
| 140 | state-dir parent not writable |

`--hook-mode` overrides exits 1 and ≥101 to 0 only after pre-flight passes; pre-flight failures *always* propagate the real exit code (the hook script's `exec` will surface the failure visibly to the user via stderr).

## Test contract

- Table-driven test per check, asserting (effective Flags) → (exit code, error substring).
- Fuzz-light test: every flag default plus a single field mutation per row, ensuring no spurious exit.
- Recursion-guard test: with `DEBATE_IN_PROGRESS=1` set, `Preflight` returns nil error and the caller exits 0 without further work.

## Acceptance criteria

- [x] All exit codes above are reachable in tests.
- [x] Each error message is verbatim as listed (lint test compares against a golden file).
- [x] No filesystem mutation during pre-flight other than the state-dir-parent stat (no `Mkdir`, no file create).
- [x] Returning `*Plan` with `Forks` indexed and aspect-mapped is the only success path.
