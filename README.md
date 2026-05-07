# debate

Adversarial review for Claude Code coding sessions.

After Claude finishes a task, `debate` forks the session for one or
more critic agents (Codex by default), runs a multi-round
cross-examination per critic, applies any concessions the proposer
makes, and surfaces only the unresolved disputes for human attention.
Each critic picks its own attack topic in round 1 (security, perf,
internal-consistency, evidence-gap, ...); later critics are told
which topics are taken and pick something else. No debate content
ever lands in the root Claude session - debate happens in branched
forks off the root.

Design and per-component contracts live under [specs/](specs/) -
start at [specs/README.md](specs/README.md) for the index.
Release-cut evidence (probe outcomes, smoke recordings) is committed
to [release-notes-v0.0.1.md](release-notes-v0.0.1.md).

## Installation

```sh
go install latere.ai/x/debate/cmd/debate@latest && debate install-hook --scope user
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
  --max-turn 6
```

Each of the four critics picks its own topic in R1; the orchestrator
passes prior critics' topics to each later critic as anti-duplication
signal. `debate --help` lists every flag. Exit codes: 0 clean,
1 unresolved leaves, 130 interrupted, 100s pre-flight failure.

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
  never sees a debate turn. Probe-confirmed against claude 2.1.131:
  a no-output Stop hook produces zero `hook_*` attachments, so the
  invariant holds across modes including the Stop-hook path. See
  [specs/28](specs/28-probe-no-output-stop-hook-outcome.md).
- **Verbatim channel.** Critic output reaches the proposer-clone as a
  plain user turn pointing at a file: `Some comments at @<path>.
  Please resolve or respond.` No skill, slash-command, or
  plugin-template wrapping that would distort the proposer's normal
  defense behavior.
- **Self-declared topics.** Each critic chooses its own attack topic
  in R1 (security, perf, internal-consistency, evidence-gap, ...) and
  later critics are told which topics are already claimed so they
  pick something else. No fixed catalog. The debate-theoretic
  property - one competent honest player suffices for soundness -
  means a lazy critic on one topic doesn't break the others.
- **Persisted ledger.** Every attack carries a stable id (`c<critic>-<seq>`),
  every transition is appended to `attacks.jsonl`. Headlines are
  picked by a pure contention score (`rounds_survived + (1 if
  re-attacked)`) - no LLM judging at this layer.
- **Best-effort critic isolation.** v0 enforces "artifact + task
  only" by critic system prompt and `codex --sandbox read-only`, not
  OS isolation; strict per-fork sandbox dirs are v1.

## Related work

- [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance)
  - research repo studying multi-agent Byzantine fault tolerance,
  including
  [spec 07 / Adversarial Debate](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md),
  the architecture this tool productizes. Specs 08–13 in the same
  repo extend the line: compute-asymmetric debate, recursive
  sub-debate, stochastic-system soundness, PCP-style leaf format,
  Prover-Estimator obfuscation defense, and DQC scaling. Each one's
  empirical result either licenses or constrains a future-work
  direction here - see
  [specs/README.md §Related research and future work](specs/README.md#related-research-and-future-work)
  for the per-spec mapping.
- Irving, Christiano & Amodei,
  [*AI Safety via Debate*](https://arxiv.org/abs/1805.00899) (2018) -
  one agent proposes, another finds flaws, a judge inspects only the
  single disputed claim that decides the debate. The
  complexity-theoretic intuition (debate ≈ PSPACE under optimal
  play) motivates the architecture.
- Brown-Cohen, Irving & Piliouras,
  [*Scalable AI Safety via Doubly-Efficient Debate*](https://arxiv.org/abs/2311.14125)
  (2023) - extends 2018 to stochastic systems and proves soundness
  under unbounded compute asymmetry between the players. The formal
  license for applying debate to LLMs at all.
