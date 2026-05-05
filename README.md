# debate

Adversarial review for Claude Code coding sessions.

After Claude finishes a coding task, `debate` forks the session for one or more critic agents (Codex by default), runs a multi-round cross-examination per critic, applies any concessions the proposer makes, and surfaces only the unresolved disputes for human attention. The original Claude session is never modified — debate happens in branched forks off the root.

**Status: design only, no implementation yet.** See [specs/01-overview.md](specs/01-overview.md) for the full design.

## Install

Coming once there's code.

## Usage

Once built:

```
debate --session-id <root-claude-session-id> --side-count 3 \
       --aspect correctness,security,perf --max-turn 6
```

Or auto-trigger via a Claude Code Stop hook — see [Hook surface](specs/01-overview.md#hook-surface-optional) in the spec.

## Design

The full design is in [specs/01-overview.md](specs/01-overview.md). Key bets:

- **Forked debate, untouched root.** Each critic gets its own fork (`claude --fork-session`). The root session's transcript is never modified; the user resumes from where they left off when debate ends. (Mechanism verified against claude 2.1.129 — see [Verified primitives](specs/01-overview.md#verified-primitives-2026-05-claude-21129) in the spec.)
- **Verbatim channel.** Critic output reaches the proposer-clone as a plain user turn pointing to a file: `Some comments at @<path>. Please resolve or respond. If you disagree, please raise it.` No skill, slash-command, or plugin-template wrapping.
- **Persisted sessions.** Each run creates `.debate/sessions/<id>/` with per-fork round files, an attacks ledger, and a summary. Clean runs are silent on stdout; unresolved runs surface a contention-scored headline.
- **Aspect-specialized critics.** Default critics split coverage across `functional-logic`, `security`, `code-quality`, `performance`. The debate-theoretic property (one competent honest player suffices for soundness) means a lazy critic on one aspect doesn't break the others.
- **Gated on per-aspect upstream research.** [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance) spec 07a measures critic-found-bug rate per aspect; aspects below threshold get dropped from defaults rather than the tool being abandoned. Ships when at least two aspects pass.

## Related

- [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance) — research repo studying multi-agent Byzantine fault tolerance, including [spec 07](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md), the architecture this tool productizes.
