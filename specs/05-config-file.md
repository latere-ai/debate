# Spec 05 — Config file (`.debate.toml`)

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §Configuration for design intent.

**Depends on:** [02](02-go-module.md), [04](04-cli-flags.md).
**Consumed by:** [06](06-preflight.md).

## Scope

In: TOML schema, search-path resolution, layering rules between config file, CLI flags, env vars, and built-in defaults.

Out: validating the resulting effective config (see [06](06-preflight.md)).

## Library

`github.com/BurntSushi/toml` for parsing. Pinned in `go.mod`.

## Schema

`.debate.toml`:

```toml
# Optional. Each field maps 1:1 to the CLI flag of the same name with
# `_` instead of `-`. Unknown keys are an error (strict decoding).

main = "claude"
side = "codex"
side_count = 4
main_model = ""
side_model = ""
max_turn = 6
aspects = ["functional-logic", "security", "code-quality", "performance"]
cost_cap_tokens = 50000          # NB: TOML name; CLI is --cost-cap
changed_lines_min = 10
state_dir = ".debate"
format = "markdown"
judge = "none"
trigger = "stop"                  # "stop" | "manual"; informational, drives install instructions only
allow_style_attacks = false       # if true, [14] keeps style-shaped attacks instead of dropping
```

Naming rule: TOML keys are `snake_case`; CLI flags are `--kebab-case`. Mapping is mechanical. The two TOML-only keys (`trigger`, `allow_style_attacks`) have no CLI counterpart in v0.

## Search path

When `--config` is empty, `debate` looks for `.debate.toml` in this order and uses the first hit:

1. Current working directory (`./.debate.toml`).
2. The git repo root, if cwd is inside a repo (`git rev-parse --show-toplevel`).
3. `$XDG_CONFIG_HOME/debate/config.toml`, falling back to `$HOME/.config/debate/config.toml`.
4. None — built-in defaults only.

When `--config <path>` is set, that path is used; missing file is an error.

## Layering

Effective config is computed in this order, last writer wins per field:

1. Built-in defaults (from [04](04-cli-flags.md) flag table).
2. User config (`$XDG_CONFIG_HOME/...`), if present.
3. Project config (`./.debate.toml` or repo root), if present.
4. Env vars (`DEBATE_*`).
5. CLI flags.

Sentinel: an absent field at level N never overwrites a present field at level N-1. `cobra` distinguishes "user supplied" via `flag.Changed`; the loader uses that flag.

## Public Go interfaces

```go
// internal/cli/config.go
package cli

// Effective returns the post-layering config. The returned *Flags is the
// same struct produced by Bind() with config-file fields merged in
// according to the layering rules.
//
// Precondition: cmd.ParseFlags has run; raw env values are already
// reflected in the bound *Flags.
func Effective(cmd *cobra.Command, f *Flags) (*Flags, error)
```

Errors:

- `ErrConfigNotFound` (only when `--config` was explicit).
- `ErrConfigDecode` wrapping the underlying `toml.Decoder` error with file path and key.

## Behavior

- Unknown TOML keys are rejected (strict decoding) with the failing key in the error.
- A field absent from TOML keeps its prior-level value.
- `aspects` and `--aspect` collapse onto the same `Flags.Aspect` slice; CLI replaces, not appends.
- `cost_cap_tokens` populates the `--cost-cap`-bound field; the dual name is preserved for back-compat with the design doc.

## Test contract

- Unit: built-in defaults survive when no file and no env are set.
- Unit: project config overrides user config.
- Unit: env var overrides both config files.
- Unit: CLI flag wins over all of the above.
- Unit: unknown TOML key errors with the key name.
- Unit: `--config <path>` with missing file returns `ErrConfigNotFound`.

## Acceptance criteria

- [ ] All nine TOML keys decode into `Flags`.
- [ ] Search path matches the order above; covered by table-driven tests.
- [ ] Layering precedence demonstrable end-to-end: same field set at every level resolves to the CLI value.
- [ ] Strict decoding catches typos before [06](06-preflight.md) runs.
