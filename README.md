# debate

Adversarial review for Claude Code coding sessions.

After Claude finishes a coding task, `debate` forks the session for
one or more critic agents (Codex by default), runs a multi-round
cross-examination per critic, applies any concessions the proposer
makes, and surfaces only the unresolved disputes for human attention.
No debate content ever lands in the root Claude session - debate
happens in branched forks off the root.

> Status: v0 implementation complete. Design lives in
> [specs/01-overview.md](specs/01-overview.md); per-component
> contracts under [specs/](specs/). v0 GA is gated on upstream
> [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance)
> 07a per-aspect rates and the no-output Stop-hook probe - see
> [specs/27-release.md](specs/27-release.md).

## Installation

```sh
go install latere.ai/x/debate/cmd/debate@latest
debate install-hook --scope user
```

Or grab a pre-built tarball from a tagged release:

```sh
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64
curl -L "https://github.com/latere-ai/debate/releases/download/v0.0.1/debate_v0.0.1_${OS}_${ARCH}.tar.gz" | tar xz
./debate install-hook --scope user
```

`install-hook` merges the verbose-format Stop hook entry into
`~/.claude/settings.json` (or `./.claude/settings.json` with `--scope
project`). `uninstall-hook` removes it.

## Example usage

With the Stop hook installed, debate runs automatically when claude
finishes responding. You see one stdout line; the summary lives on
disk:

```text
$ # ...claude does its thing...
[debate] 2 unresolved; see /repo/.debate/sessions/20260506T140905Z-q3a9f1/summary.md

$ cat .debate/sessions/*/summary.md
# Debate review - terminated: steady-state

## Headline (most contested unresolved)
- [security/api.go:88] SQL injection via unparameterized LIKE
  - Critic: framework auto-escape doesn't cover LIKE patterns
  - Proposer: parameterized via SQLAlchemy
  - **Stake**: GET /search?q=%' OR 1=1--
  - Contention: 3 (re-attacked: true)

## Resolved (5)
- [conceded] Off-by-one in pagination → fixed at api.go:42
...

## Stats
critic-found-bug rate: 5/8 attacks led to a fix
debate cost: 38k tokens, 6 rounds, 4 critics
```

Trivial diffs (under `--changed-lines-min`, default 10) short-circuit
in milliseconds. No session folder, just one `kind:"skipped"` line in
`.debate/log.jsonl`.

For CI gating, scripted batch runs, or out-of-band review, invoke
manually:

```sh
debate \
  --session-id <root-claude-session-id> \
  --side-count 4 \
  --aspect functional-logic,security,code-quality,performance \
  --max-turn 6
```

`debate --help` lists every flag. Exit codes: 0 clean, 1 unresolved
leaves, 130 interrupted, 100s pre-flight failure.

## Design architecture

```
                  user's claude session (the "root")
                            │
                            │  --fork-session (per critic)
                ┌───────────┼───────────┐
                ▼           ▼           ▼
          fork-1            fork-2 …    fork-N
          ┌──────┐          ┌──────┐    ┌──────┐
          │ pro- │ <══════> │ ...  │    │ ...  │
          │poser │  rounds  │      │    │      │
          │clone │          │      │    │      │
          └──┬───┘          └──────┘    └──────┘
             │
             ▼ writes round files to disk
   .debate/sessions/<id>/forks/critic-i/rounds/r{1,2,3,…}-{critic,proposer}.md
                                            │
                                            ▼
                                    summary.md  (contention-scored headline + leaves)
```

Five load-bearing pieces (full design in
[spec 01](specs/01-overview.md)):

- **Forked debate, no debate content in root.** Each critic gets its
  own claude fork via `--fork-session`. The user's root transcript
  never sees a debate turn. (The Stop-hook path may add a single
  hook-status attachment per run - probe owed before v0 GA.)
- **Verbatim channel.** Critic output reaches the proposer-clone as a
  plain user turn pointing at a file: `Some comments at @<path>.
  Please resolve or respond.` No skill, slash-command, or
  plugin-template wrapping that would distort the proposer's normal
  defense behavior.
- **Aspect-specialized critics.** Default coverage splits across
  `functional-logic`, `security`, `code-quality`, `performance`. The
  debate-theoretic property - one competent honest player suffices
  for soundness - means a lazy critic on one aspect doesn't break the
  others.
- **Persisted ledger.** Every attack carries a stable id (`c<critic>-<seq>`),
  every transition is appended to `attacks.jsonl`. Headlines are
  picked by a pure contention score (`rounds_survived + (1 if
  re-attacked)`) - no LLM judging at this layer.
- **Best-effort critic isolation.** v0 enforces "diff + task only" by
  aspect prompt and `codex --sandbox read-only`, not OS isolation;
  strict per-fork sandbox dirs are v1.

## Related work

- [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance)
  - research repo studying multi-agent Byzantine fault tolerance,
  including
  [spec 07 / Adversarial Debate](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md),
  the architecture this tool productizes.
- Irving, Christiano & Amodei,
  [*AI Safety via Debate*](https://arxiv.org/abs/1805.00899) (2018) -
  one agent proposes, another finds flaws, a judge inspects only the
  single disputed claim that decides the debate. The
  complexity-theoretic intuition (debate ≈ PSPACE under optimal
  play) motivates the architecture.
