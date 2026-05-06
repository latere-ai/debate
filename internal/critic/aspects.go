// Package critic owns the critic-side protocol: aspect prompts,
// markdown attack format, parser, normalizer.
package critic

import (
	"fmt"
	"strings"
)

// Aspect is a critic's specialization: prompt + cross-aspect filter.
type Aspect struct {
	Name              string
	SystemPrompt      string
	ForbiddenKeywords []string
}

const skeletonHeader = `You are an adversarial code reviewer focused on **%s**. You are
critic %d of %d, reviewing a code diff produced by a Claude Code session
against a task description supplied below.

Your job is to find concrete %s bugs in the diff. You are not reviewing
taste, style, or aspects other than %s. The mediator will drop attacks
that wander outside this aspect.

Hard rules:
1. Each attack MUST name a concrete behavior or maintainability impact
   in its %s domain. No taste comments. No "consider renaming."
2. Each attack MUST include a runnable reproduction. Attacks without
   one are dropped at parse time.
%s

Output format:

# Critic <i> - round <n> attacks

aspect: %s

## c<i>-<seq> [path:line]

claim: <one paragraph>

expected violation: <one paragraph; may include fenced examples>

reproduction:
` + "```" + `
<exact runnable input/command/test/repro>
` + "```" + `

Sources you have:
- The original task description (verbatim, below).
- The unified diff (verbatim, below).
- For round >= 3: the proposer's prior responses (referenced by file).

Sources you do NOT have and must not invent:
- The full source tree. You may only attack code in the diff or files
  the diff directly references. Do not invent file paths.
- Any external system you cannot reach via the reproduction.
`

// Builtin returns the four default aspects in canonical order.
func Builtin() map[string]Aspect {
	return map[string]Aspect{
		"functional-logic": {
			Name: "functional-logic",
			SystemPrompt: aspectPrompt(
				"functional-logic",
				"3. Focus on what the diff is supposed to compute. Off-by-ones, missing\n   branches, silent-failure paths, edge cases the task implies but the\n   code missed, incorrect default values.\n4. Boundary inputs are fair game: empty collections, nil/None, negative\n   numbers, zero, max/min ints, leap years, time-zone transitions,\n   unicode at byte boundaries.",
			),
			ForbiddenKeywords: []string{"sql injection", "race condition", "deadlock", "auth", "rbac", "csrf", "xss", "n+1", "allocations", "blocking call", "hot path"},
		},
		"security": {
			Name: "security",
			SystemPrompt: aspectPrompt(
				"security",
				"3. Focus on input validation, authn/authz, injection (SQL, shell,\n   template, deserialization), data exposure, secrets in logs, unsafe\n   deserialization, missing CSRF/HMAC checks, broken access control.\n4. Reproductions should be minimal exploit-shaped curls, payloads, or\n   test inputs. Theoretical attacks (\"if the attacker had the secret\n   key\") are dropped - name a concrete reachable path.",
			),
			ForbiddenKeywords: []string{"off-by-one", "missing branch", "n+1", "allocations", "blocking call", "long function", "unclear naming", "swallowed exception"},
		},
		"code-quality": {
			Name: "code-quality",
			SystemPrompt: aspectPrompt(
				"code-quality",
				"3. Focus on real maintainability impact: long functions that hide\n   bugs, swallowed exceptions that erase signal, dead branches, unclear\n   naming where it bites readability of THIS diff (not \"I'd prefer x\").\n   Functions that lie about their behavior in their name.\n4. NOT in scope: formatting, single/double quote choices, indent width,\n   comment style, \"I would have written it this way.\" Those are\n   dropped at parse time as style.",
			),
			ForbiddenKeywords: []string{"sql injection", "auth", "race condition", "deadlock", "off-by-one"},
		},
		"performance": {
			Name: "performance",
			SystemPrompt: aspectPrompt(
				"performance",
				"3. Focus on algorithmic complexity, N+1 IO patterns, unnecessary\n   allocations or copies, blocking calls in hot paths, unbounded\n   work-per-request.\n4. The reproduction must demonstrate the cost concretely: a benchmark\n   sketch, a load-test invocation, a calculation showing the\n   complexity blow-up. Vague \"this might be slow\" is dropped.",
			),
			ForbiddenKeywords: []string{"sql injection", "missing branch", "auth", "swallowed exception", "unclear naming"},
		},
	}
}

// Lookup returns the named aspect, falling back to a generic prompt
// for unknown names.
func Lookup(name string) Aspect {
	if a, ok := Builtin()[name]; ok {
		return a
	}
	return Aspect{
		Name: name,
		SystemPrompt: aspectPrompt(
			name,
			fmt.Sprintf("3. Focus on the %s aspect of this code. Define what counts as a\n   bug in this aspect at the start of each attack's `claim` line.\n4. As above: concrete behavior or maintainability impact, runnable\n   reproduction.", name),
		),
	}
}

// Assemble produces the full system prompt for one critic round.
func Assemble(a Aspect, criticIndex, round int, priorRoundsNote string) string {
	var b strings.Builder
	b.WriteString(a.SystemPrompt)
	b.WriteString("\n\nRound: ")
	fmt.Fprintf(&b, "%d (critic-%d)", round, criticIndex)
	if priorRoundsNote != "" {
		b.WriteString("\n\n")
		b.WriteString(priorRoundsNote)
	}
	return b.String()
}

func aspectPrompt(name, rules string) string {
	return fmt.Sprintf(skeletonHeader, name, 0, 0, name, name, name, rules, name)
}
