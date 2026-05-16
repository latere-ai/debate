# Spec 23 - `summary.md` rendering, surfacing rule, exit codes

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) §"Output format" and §"Surfacing rule" for design intent.

**Depends on:** [10](10-run-artifacts.md), [11](11-fork-artifacts.md), [12](12-attacks-ledger.md), [19](19-round-loop.md), [20](20-termination.md), [21](21-signals.md), [22](22-contention-headline.md).
**Consumed by:** [25](25-probes.md), [27](27-release.md).

## Scope

In: the `summary.md` template + renderer, the Stats block, the surfacing rule (silent vs surface on stdout), and the exit-code matrix.

Out: contention scoring ([22](22-contention-headline.md)), the `--format json` alternate output (deferred to v1; placeholder behavior described).

## summary.md template (v0)

```markdown
# Agon review - terminated: <termination>

## Headline (most contested unresolved)
- [<aspect>/<location>] <one-line restatement of claim>
  - Critic: <one-paragraph last-known critic stance>
  - Proposer: <one-paragraph last-known proposer rebuttal>
  - **Stake**: <reproduction>
  - Contention: <rounds_survived> rounds survived, <re-attacked / not re-attacked> after defense.

## Other unresolved (<N>, sorted by contention)
- [<aspect>/<location>] <claim>
  - Stake: <reproduction>
  - Contention: <score>
- ...

## Resolved (<M>)
- [conceded] <claim summary> → fixed at <comma-separated concession_files>
- [rebutted] <claim summary> → critic withdrew after <reason from withdraw record, if available>
- [withdrawn] <claim summary> → critic withdrew at round <round>

## Stats
critic-found-bug rate: <conceded_count>/<total_attacks> attacks led to a fix
agon cost: <tokens> tokens, <total_rounds> rounds, <fork_count> critics
session: <relative path to session folder>
```

Sections are omitted when empty:

- No `## Headline` and no `## Other unresolved` if `unresolved_count == 0`.
- No `## Resolved` if `resolved_count == 0` (rare but possible if every attack ended unresolved).
- `## Stats` always present.

## Public Go interfaces

```go
// internal/summary/render.go
package summary

type Render struct {
    Format string  // "markdown" | "json"
}

// Render returns the bytes to write to summary.md.
func (r *Render) Render(s *round.Summary, agg map[string]ledger.Record) ([]byte, error)

// SurfacingDecision returns whether the orchestrator should print the
// summary path to stdout, and what stdout line to emit.
type SurfacingDecision struct {
    Surface     bool   // true: print line; false: silent (one log.jsonl pointer only)
    StdoutLine  string // exact line to print, no trailing newline
    ExitCode    int    // process exit code (0 clean, 1 unresolved, 130 interrupted)
}

func Decide(s *round.Summary) SurfacingDecision
```

## Renderer rules

- Fields with embedded newlines (e.g., `claim` paragraphs) are indented by two spaces under their bullet so the markdown stays well-formed when rendered.
- Code-fenced blocks inside `reproduction` are escaped via the markdown renderer's "raw" mode - the renderer emits them verbatim, not re-fenced.
- Aspect names are passed through unchanged (free-form per [15](15-aspect-prompts.md)).
- `<comma-separated concession_files>` joins with `, ` (comma + space). Lists ≥ 5 truncate to "first three, ..., last" (`api.py, util.py, ..., db.py`).
- "rounds survived" displays `r.RoundsSurvived` (numeric); the human-readable phrase is the same as the [22](22-contention-headline.md) score.

## Surfacing rule

```
Decide(s):
    unresolved := count(agg, Status=="unresolved")

    switch s.Termination {
    case TermSteadyState:
        if unresolved == 0:
            return Surface=false,
                   StdoutLine="[agon] clean run; see .agon/log.jsonl",
                   ExitCode=0
        else:
            return Surface=true,
                   StdoutLine=fmt("[agon] %d unresolved; see %s", unresolved, summaryPath),
                   ExitCode=1
    case TermMaxTurn, TermCostCap, TermMalformedOutput:
        return Surface=true,
               StdoutLine=fmt("[agon] terminated %s (%d unresolved); see %s",
                              s.Termination, unresolved, summaryPath),
               ExitCode=1
    case TermInterrupted:
        return Surface=true,
               StdoutLine=fmt("[agon] interrupted (%d known unresolved); partial review at %s",
                              unresolved, summaryPath),
               ExitCode=130
    }
```

The stdout line is *exactly one line*, no leading whitespace. It goes to the orchestrator's stdout.

## Exit-code matrix (recap from [21](21-signals.md))

| `Surface` | `Termination` | `unresolved` | exit code |
|---|---|---|---|
| false | steady-state | 0 | 0 |
| true  | steady-state | ≥ 1 | 1 |
| true  | max-turn | any | 1 |
| true  | cost-cap | any | 1 |
| true  | malformed-output | any | 1 |
| true  | interrupted | any | 130 |

Pre-flight failures (codes 100+) always exit with their intrinsic code ([06](06-preflight.md)).

## --format json (v1)

When `--format json`, the renderer emits `summary.json` instead of `summary.md` with the same information as a JSON object:

```jsonc
{
  "schema": "agon.summary.v0",
  "termination": "...",
  "headline": {...} | null,
  "unresolved": [...],
  "resolved": [...],
  "stats": {...}
}
```

v0 does not implement this; the flag is accepted but stored alongside `Format = "markdown"` for forward-compat. The renderer panics with "json format not implemented in v0" if `Format == "json"`. CI test asserts the panic to keep us honest.

## Test contract

- Golden: a fixture `Summary` + ledger renders to a known `summary.md` byte-for-byte (`testdata/golden/summary/*.md`).
- Unit: `Decide` returns the right SurfacingDecision for each termination shape.
- Unit: clean run emits the silent stdout line.
- Unit: empty `Other unresolved` and `Resolved` sections are omitted from rendering.

## Acceptance criteria

- [x] Renderer is deterministic (no map-order leaks; sort-stable inputs).
- [x] All sections optional except `## Stats` (audit test).
- [x] Stdout lines match the `Decide` table exactly (golden-string test).
- [x] Exit-code mapping applied at `cmd/agon/main.go`, not in this package; renderer is exit-code-agnostic.
- [x] `--format json` is rejected with a clear panic in v0; the placeholder is not silently ignored.
