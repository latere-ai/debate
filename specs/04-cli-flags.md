# Spec 04 — CLI flags

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"CLI surface" for design intent.

**Depends on:** [02](02-go-module.md), [03](03-ci-lint-release.md).
**Consumed by:** [05](05-config-file.md), [06](06-preflight.md), [17](17-claude-proposer.md), [18](18-critic-drivers.md), [19](19-round-loop.md), [20](20-termination.md), [23](23-summary-render.md), [24](24-stop-hook.md).

## Scope

In: every flag's exact name, type, default, env-var override, help text, and the `cobra`-style command tree. Flag parsing only — no validation logic, no behavior side-effects.

Out: validation and pre-flight checks ([06](06-preflight.md)), config-file layering ([05](05-config-file.md)), the actual subprocess invocations.

## Library

`github.com/spf13/cobra` for command tree and flag parsing. `github.com/spf13/pflag` is the underlying flag layer (POSIX/GNU-style `--flag=value` and `--flag value`). Both pinned in `go.mod` at the latest patch.

Single root command (`debate`); no subcommands in v0. (`debate resume <id>` is v1, see [01-overview.md](01-overview.md) §Versioning.)

## Flag set

All flags exposed on the root command:

| Flag | Type | Default | Env override | Notes |
|---|---|---|---|---|
| `--main` | string | `claude` | `DEBATE_MAIN` | proposer agent: `claude` or `codex`. v0 errors on `codex` (see [06](06-preflight.md)). |
| `--side` | string | `codex` | `DEBATE_SIDE` | critic agent: `claude` or `codex`. |
| `--side-count` | int | `4` | `DEBATE_SIDE_COUNT` | number of critic forks. Must equal `len(--aspect)`. |
| `--main-model` | string | `""` | `DEBATE_MAIN_MODEL` | optional cross-family; required + must differ from `--side-model` same-family. |
| `--side-model` | string | `""` | `DEBATE_SIDE_MODEL` | same rule. |
| `--max-turn` | int | `6` | `DEBATE_MAX_TURN` | per-fork cap (P+C exchanges). |
| `--aspect` | stringSlice | `functional-logic,security,code-quality,performance` | `DEBATE_ASPECT` | comma-separated. Names are free-form (see [15](15-aspect-prompts.md)). |
| `--session-id` | string | `""` | `DEBATE_SESSION_ID` | claude root session id. Required for claude-as-proposer auto-trigger. |
| `--transcript` | string | `""` | `DEBATE_TRANSCRIPT` | path to root JSONL. Optional; preferred when present. |
| `--diff-from` | string | `HEAD` | `DEBATE_DIFF_FROM` | git ref for diff base. |
| `--diff-to` | string | `.` | `DEBATE_DIFF_TO` | `.` means working tree; otherwise a git ref. |
| `--task-context` | string | `""` | `DEBATE_TASK_CONTEXT` | mandatory iff neither `--session-id` nor `--transcript` is set. |
| `--judge` | string | `none` | `DEBATE_JUDGE` | one of `none`, `llm`, `human`. v0 only accepts `none`. |
| `--cost-cap` | int | `50000` | `DEBATE_COST_CAP` | total token budget across all forks. |
| `--changed-lines-min` | int | `10` | `DEBATE_CHANGED_LINES_MIN` | trivial-diff gate. |
| `--state-dir` | string | `.debate` | `DEBATE_STATE_DIR` | where session folders go. |
| `--format` | string | `markdown` | `DEBATE_FORMAT` | `markdown` or `json` for `summary.md`/`summary.json`. |
| `--hook-mode` | bool | `false` | `DEBATE_HOOK_MODE` | force exit 0; see [23](23-summary-render.md). |
| `--config` | string | `""` | `DEBATE_CONFIG` | explicit `.debate.toml` path; overrides search ([05](05-config-file.md)). |
| `--verbose` | count | `0` | `DEBATE_VERBOSE` | `-v`, `-vv` for log levels. |
| `--version` | bool | — | — | prints version (see [03](03-ci-lint-release.md)) and exits 0. |
| `--help`, `-h` | bool | — | — | cobra default. |

## Env-var rule

For every flag with an env override: precedence is **CLI flag > env var > config file > built-in default**. Layering belongs to [05](05-config-file.md); this spec only declares what env var binds to which flag.

`stringSlice` env vars use comma as the separator; trailing/leading whitespace per element trimmed.

## Public Go interfaces

```go
// internal/cli/flags.go
package cli

type Flags struct {
    Main             string
    Side             string
    SideCount        int
    MainModel        string
    SideModel        string
    MaxTurn          int
    Aspect           []string
    SessionID        string
    Transcript       string
    DiffFrom         string
    DiffTo           string
    TaskContext      string
    Judge            string
    CostCap          int
    ChangedLinesMin  int
    StateDir         string
    Format           string
    HookMode         bool
    Config           string
    Verbose          int
}

// Bind registers all flags onto the supplied cobra command and returns
// a *Flags whose fields are populated when the command runs.
func Bind(cmd *cobra.Command) *Flags
```

`Bind` only registers and binds. It does not validate, normalize, or read the config file.

## Help text

`cmd/debate/main.go` sets:

```
Use:   "debate"
Short: "Adversarial review for Claude Code coding sessions."
Long:  <one paragraph from README §1>
```

Per-flag help strings copied verbatim from the table above's "Notes" column, prefixed with the type and default in cobra's standard form.

## Test contract

- Unit: each flag round-trips from `os.Args` to `Flags` field with the right type.
- Unit: env-var override precedence verified against an explicit table.
- Unit: `stringSlice` parses both `--aspect a,b,c` and `--aspect=a --aspect=b --aspect=c`.
- Unit: `--version` exits 0 without invoking any other code path.

## Acceptance criteria

- [x] `debate --help` lists all flags with the table's defaults.
- [x] `debate --version` prints version/commit/date.
- [x] Every flag has a `DEBATE_*` env var that, when set, populates the corresponding `Flags` field if no CLI flag is given.
- [x] CLI flag wins over env var when both set.
- [x] No business logic runs in `Bind`; calling it without invoking the cobra command leaves the program in a no-op state.
