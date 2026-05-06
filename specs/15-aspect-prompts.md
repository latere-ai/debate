# Spec 15 - Aspect prompts

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Critic specialization" for design intent.

**Depends on:** [13](13-critic-output-format.md).
**Consumed by:** [14](14-attack-parser.md) (cross-aspect filter), [18](18-critic-drivers.md) (prompt assembly).

## Scope

In: prompt template structure, the four default aspect prompts verbatim, the cross-aspect keyword sets [14](14-attack-parser.md)'s F3 filter consumes, and the open-aspect extension model.

Out: how the prompt is delivered to the agent ([18](18-critic-drivers.md)), how attacks are parsed ([14](14-attack-parser.md)).

## Public Go interfaces

```go
// internal/critic/aspects.go
package critic

type Aspect struct {
    Name              string
    SystemPrompt      string   // the full prompt string
    ForbiddenKeywords []string // for [14] F3 cross-aspect filter
}

// Builtin returns the four default aspects.
func Builtin() map[string]Aspect

// Lookup returns an Aspect for a name; for unknown names, returns a
// generic Aspect with a default prompt template (see "Open extension"
// below) and an empty ForbiddenKeywords set.
func Lookup(name string) Aspect

// Assemble produces the full system prompt for one critic round:
//   <Aspect.SystemPrompt>
//   <output-format reminder from [13]>
//   <round-specific addenda: round-number, prior-rounds note, etc.>
func Assemble(a Aspect, criticIndex, round int, priorRoundsNote string) string
```

## Template structure (every aspect)

Every aspect's `SystemPrompt` follows the same skeleton:

```
You are an adversarial code reviewer focused on **<aspect>**. You are
critic <i> of <total>, reviewing a code diff produced by a Claude
Code session against a task description supplied below.

Your job is to find concrete <aspect> bugs in the diff. You are not
reviewing taste, style, or aspects other than <aspect>. The mediator
will drop attacks that wander outside this aspect.

Hard rules:
1. Each attack MUST name a concrete behavior or maintainability impact
   in its <aspect> domain. No taste comments. No "consider renaming."
2. Each attack MUST include a runnable reproduction (input, command,
   test, or minimal repro). Attacks without one are dropped at parse
   time. If you cannot construct a reproduction, do not file the attack.
3. <aspect-specific rule>
4. <aspect-specific rule>

Output format:
<inline reminder of [13]'s schema, with worked example>

Sources you have:
- The original task description (verbatim, below).
- The unified diff (verbatim, below).
- For round >= 3: the proposer's prior responses (referenced by file).

Sources you do NOT have and must not invent:
- The full source tree. You may only attack code in the diff or files
  the diff directly references. Do not invent file paths.
- Any external system you cannot reach via the reproduction.
```

The `Assemble` function fills the placeholders and appends a "Round <n>" footer.

## Aspect: functional-logic

```
<skeleton above with "functional-logic">

Aspect-specific rules:
3. Focus on what the diff is supposed to compute. Off-by-ones, missing
   branches, silent-failure paths, edge cases the task implies but the
   code missed, incorrect default values.
4. Boundary inputs are fair game: empty collections, nil/None,
   negative numbers, zero, max/min ints, leap years, time-zone
   transitions, unicode at byte boundaries.
```

ForbiddenKeywords (for cross-aspect filter): `["sql injection", "race condition", "deadlock", "auth", "rbac", "csrf", "xss", "n+1", "allocations", "blocking call", "hot path"]`.

## Aspect: security

```
<skeleton above with "security">

Aspect-specific rules:
3. Focus on input validation, authn/authz, injection (SQL, shell,
   template, deserialization), data exposure, secrets in logs, unsafe
   deserialization, missing CSRF/HMAC checks, broken access control.
4. Reproductions should be minimal exploit-shaped curls, payloads, or
   test inputs. Theoretical attacks ("if the attacker had the secret
   key") are dropped - name a concrete reachable path.
```

ForbiddenKeywords: `["off-by-one", "missing branch", "n+1", "allocations", "blocking call", "long function", "unclear naming", "swallowed exception"]`.

## Aspect: code-quality

```
<skeleton above with "code-quality">

Aspect-specific rules:
3. Focus on real maintainability impact: long functions that hide
   bugs, swallowed exceptions that erase signal, dead branches,
   unclear naming where it bites readability of THIS diff (not "I'd
   prefer x"). Functions that lie about their behavior in their name.
4. NOT in scope: formatting, single/double quote choices, indent
   width, comment style, "I would have written it this way." Those
   are dropped at parse time as style.
```

ForbiddenKeywords: `["sql injection", "auth", "race condition", "deadlock", "off-by-one"]`.

## Aspect: performance

```
<skeleton above with "performance">

Aspect-specific rules:
3. Focus on algorithmic complexity, N+1 IO patterns, unnecessary
   allocations or copies, blocking calls in hot paths, unbounded
   work-per-request.
4. The reproduction must demonstrate the cost concretely: a benchmark
   sketch, a load-test invocation, a calculation showing the
   complexity blow-up. Vague "this might be slow" is dropped.
```

ForbiddenKeywords: `["sql injection", "missing branch", "auth", "swallowed exception", "unclear naming"]`.

## Open extension

Aspect names are free-form strings (see [05](05-config-file.md)). Unknown names get a generic prompt:

```
<skeleton with the user-supplied name>

Aspect-specific rules:
3. Focus on the <name> aspect of this code. Define what counts as a
   bug in this aspect at the start of each attack's `claim` line.
4. As above: concrete behavior or maintainability impact, runnable
   reproduction.
```

ForbiddenKeywords for unknown aspects is empty (the F3 filter is a no-op).

## Cross-aspect filter integration

`Builtin()` returns a map with `ForbiddenKeywords` set per aspect. [14](14-attack-parser.md)'s F3 reads this map to drop attacks that match another aspect's keywords without matching any of the current aspect's expected vocabulary.

To avoid maintaining yet another keyword list, F3's "expected vocabulary for current aspect" is derived from the *forbidden* sets of *all other* aspects: an attack survives iff its claim text doesn't match any forbidden keyword from any other aspect. Equivalent to: "the attack's vocabulary belongs to this aspect."

## Test contract

- Unit: `Builtin()` returns exactly four entries with the four canonical names.
- Unit: `Lookup("functional-logic")` returns the right prompt.
- Unit: `Lookup("not-a-real-aspect")` returns the generic prompt with no forbidden keywords.
- Golden: `Assemble(a, 1, 1, "")` for each aspect produces a known-good prompt string (`testdata/golden/prompts/<aspect>-r1.txt`).
- Cross-aspect: an attack `claim` containing "SQL injection" tagged as `performance` triggers F3 drop in [14](14-attack-parser.md)'s test.

## Acceptance criteria

- [x] All four prompts present verbatim in `internal/critic/aspects.go` (or a `_data` go:embed file).
- [x] Each prompt's structure matches the skeleton (audit-via-test).
- [x] `Assemble` is a pure function (no I/O, no clock).
- [x] Cross-aspect keyword lists do not overlap (a keyword forbidden by aspect A is not in aspect A's "expected vocabulary").
