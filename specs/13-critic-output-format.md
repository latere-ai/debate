# Spec 13 — Critic output format (the contract)

> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Rounds" → R1 Attack for design intent.

**Depends on:** none (a pure protocol spec).
**Consumed by:** [14](14-attack-parser.md), [15](15-aspect-prompts.md), [18](18-critic-drivers.md).

## Scope

In: the *exact* markdown shape the critic must emit, the schema of an attack record inside the markdown, and the rules a parser is allowed to assume. This is the contract the aspect prompts in [15](15-aspect-prompts.md) ask the critic to follow, and the surface [14](14-attack-parser.md) reads.

Out: how the critic is invoked ([18](18-critic-drivers.md)), what the parser does after reading ([14](14-attack-parser.md)).

## Why a contract spec separately

Two readers ([14](14-attack-parser.md) the parser, [15](15-aspect-prompts.md) the prompt author) need to agree on the same wire format. Splitting it out keeps either side free to evolve without re-relitigating the format.

## Format

The critic emits a single markdown document. Whitespace outside section bodies is irrelevant. Headers must use `#`-style (not setext). Field names are lowercase.

```markdown
# Critic <critic-index> — round <round-number> attacks

aspect: <aspect-name>

## c<critic-index>-<seq> [<location>]

claim: <one-paragraph claim, single line>

expected violation: <one-paragraph, may include code-fenced examples>

reproduction:
```
<exact runnable input / command / test / minimal repro>
```

---

## c<critic-index>-<seq+1> [<location>]
...
```

## Required fields per attack

| Field | Rule |
|---|---|
| Section header | Exactly `## c<i>-<seq> [<location>]`. `<i>` is the critic's 1-based index from `--side-count`. `<seq>` is monotonically increasing within a round. `<location>` is `path:line` or `path:start-end`. |
| `claim:` | One paragraph, one logical assertion. No bullets, no nested headers. |
| `expected violation:` | One paragraph, may include fenced code blocks for clarity. |
| `reproduction:` (fenced block) | A standalone runnable artifact. Examples: a shell command, a Python doctest, a SQL statement, a curl invocation. **No fenced block → attack is dropped at parse time.** |

Trailing `---` between attacks is a separator hint; the parser tolerates its presence or absence.

## Optional fields

The critic *may* include a leading paragraph after the document header (before the first `##`) summarizing its review approach. Parsers ignore it.

## What the critic MUST NOT emit

- Nested headers under an attack (`### ...`). The parser stops the section at the next `## `.
- Multiple `claim:` / `expected violation:` / `reproduction:` blocks per attack.
- Style-only attacks (no concrete behavior or maintainability impact). [14](14-attack-parser.md)'s heuristic drops these.
- Cross-aspect attacks (a security critic emitting a perf attack). [14](14-attack-parser.md) drops these by `aspect: <name>` mismatch + heuristics on the claim text.
- Attacks lacking a `reproduction:` block (see [14](14-attack-parser.md)).

## Status tags in re-attack rounds

For odd rounds R3, R5, … the critic also names the disposition for each prior attack-id it is re-engaging with:

```markdown
## c1-3 [src/api.py:88] (re-attack)

claim: <restated, possibly tightened>

expected violation: <may incorporate proposer's R2 to narrow the bug>

reproduction:
```
<possibly tighter than R1>
```
```

Or for withdrawal:

```markdown
## c1-3 [src/api.py:88] (withdraw)

reason: <one-paragraph explanation>
```

The disposition tag goes after the location bracket, in parentheses. `re-attack` and `withdraw` are the only two recognized; anything else is treated as a fresh attack with the parent id reused (see [14](14-attack-parser.md) for the validation).

## Worked example

```markdown
# Critic 2 — round 1 attacks

aspect: security

## c2-1 [src/api.py:88]

claim: The search handler concatenates user-supplied input directly into a SQL `LIKE` pattern without escaping `%` or `_`.

expected violation: An attacker can probe the table by submitting `q=%' OR 1=1--`, which terminates the LIKE pattern and injects boolean logic. Even with a parameterized query, framework auto-escape does not cover LIKE-pattern metacharacters in MySQL/PostgreSQL.

reproduction:
```
curl 'http://localhost:8000/search?q=%25%27%20OR%201%3D1--'
```

---

## c2-2 [src/auth.py:42]

claim: The login endpoint logs the full `Authorization` header on auth failure, leaking bearer tokens to the application log.

expected violation: A failed login with a valid-shaped bearer token writes that token to stdout and any structured-log sink the app forwards to.

reproduction:
```
curl -i -H 'Authorization: Bearer test-token-9f1' http://localhost:8000/auth/wrong
# then: grep -F 'Bearer test-token-9f1' app.log
```
```

This document parses to two attacks under [14](14-attack-parser.md), both surviving the reproduction-required filter.

## Test contract

- Fixture document with two well-formed attacks → parser emits two attack records.
- Fixture with missing `reproduction:` → that attack dropped, others retained.
- Fixture with a duplicate `## c2-1` header → second occurrence renamed by the normalizer ([14](14-attack-parser.md)).
- Fixture with a `(re-attack)` tag → parser sets `re_attacked = true` on that record.
- Fixture with a `(withdraw)` tag → parser emits a `Status = withdrawn` record without re-introducing the attack.

## Acceptance criteria

- [ ] The exact format above is what [15](15-aspect-prompts.md)'s prompts require the critic to produce.
- [ ] [14](14-attack-parser.md) parses the worked-example fixture without warnings.
- [ ] Style-shaped attacks and reproduction-less attacks are detectable from the document alone (no agent re-prompting needed).
