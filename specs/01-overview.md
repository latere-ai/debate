# Spec 01 — Design Overview

The full design for `debate`. This document is the authoritative spec; the project README is a usage reference.

## What this is

A tool that productizes the debate architecture from [agents-byzantine-tolerance/specs/07-adversarial-debate.md](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md). The upstream spec 07 measures whether debate works on coding tasks; this repo specifies a tool that bets it does, in a way that fails gracefully if 07's H6 (lazy critic collapses the architecture) bites.

## Goal

When Claude finishes a coding task, run a critic (Codex by default) that produces concrete adversarial comments on the diff. Claude either fixes or defends each comment. Up to N rounds of cross-examination. Unresolved leaves surface to the human at the end as a structured review. The human inspects only what wasn't resolved by debate, not the full output.

## Architecture

- **Proposer (P)** — the Claude Code session that just wrote code. Sees the original task and its own diff.
- **Critic(s) (C₁..Cₖ)** — k independent adversarial agents (Codex by default), each scoped to a different aspect (correctness, security, perf, …). Sees the diff + task context only.
- **Judge** — the human, but only at the end and only on unresolved leaves. Optional LLM-judge mode for triage.
- **Mediator** — a small orchestrator process that routes messages between P and the Cᵢ and tracks per-claim resolution state. This is the actual binary we ship.

### Fork model (load-bearing)

The debate must not pollute the root session. The user's original Claude Code conversation is the **root** — when debate ends, the user resumes from that root, with the working tree updated (any concession-fixes have landed) and a summary on disk. The conversation log of the root session never sees a debate turn.

Each critic gets its own **debate fork**: a clone of the root session paired with one critic process. Forks are isolated:

- **No cross-critic leakage.** Critic A's debate is invisible to critic B. Each fork has its own (proposer-clone, critic) conversation pair. Critics never see each other's attacks or the proposer's responses to other critics.
- **No root pollution.** The proposer in a fork is a clone of the root produced via Claude Code's built-in `--fork-session` flag. The original root session's transcript is unchanged after the debate.
- **Shared working tree.** Concession-fixes land in the real working tree as they happen. Critics in later forks see the updated code, but not the conversation that produced it.
- **Serial execution in v0.** Forks run one at a time to avoid working-tree races. Parallel forks via per-fork git worktrees (`claude --worktree`) are deferred until the basic flow is proven (see Out of scope).

#### Tree shape

```
root session  (user → assistant: "code written")  ←─── user resumes here when done
   │
   ├── fork-1: proposer-clone-1 ↔ critic-1
   │     └── rounds/r{1,2,3,...}-{critic,proposer}.md
   │
   ├── fork-2: proposer-clone-2 ↔ critic-2
   │     └── rounds/r{1,2,3,...}-{critic,proposer}.md
   │
   └── fork-3: proposer-clone-3 ↔ critic-3
         └── rounds/r{1,2,3,...}-{critic,proposer}.md
                                  │
                                  ▼
                            summary.md  (aggregated across forks)
                                  │
                                  ▼
                            stdout to user
```

Branches do not see each other. The user, looking at their Claude Code session, sees only the root. The debate lives entirely on disk and in the orchestrator's stdout. The proposer-clone sessions are artifacts — they remain in the session picker grouped under root (Claude Code groups forks under their root automatically) but the user can ignore them.

### Mechanism — how forks are created and driven

The orchestrator uses three Claude Code primitives, all already documented:

1. **Forking the proposer.** For each critic, the orchestrator runs:

   ```
   claude --resume <root-id> --fork-session --output-format json \
          -p "Some comments at @<path-to-rounds/r1-critic-i.md>. Please resolve or respond. If you disagree, please raise it."
   ```

   `--fork-session` creates a copy of the root conversation and switches into it. The pointer message is the first user turn in the fork. The JSON output includes the new fork session ID, which the orchestrator captures for subsequent rounds in this fork.

2. **Continuing within a fork.** Each subsequent round in the same fork uses:

   ```
   claude --resume <fork-id> -p "Critic re-attacks at @<path>. Please resolve or respond."
   ```

   No new fork — same fork, next user turn.

3. **Driving the critic.** Codex has no fork concept; each `codex exec` is a fresh process. The critic-side state we need (per-round inputs, per-round outputs) lives on disk in the round files; codex reads the previous round's proposer file at start of each round.

The wrap-up step prints to stdout. **Claude Code provides no way to inject an assistant turn into the root session**, and the user explicitly does not want a system-reminder-shaped notification (`additionalContext` from a `SessionStart` hook would technically work but violates the channel constraint — it's a system message, not natural feedback). The summary lives at `summary.md` and on stdout; that's the contract.

### Channel constraint (load-bearing)

**Critic output must reach the proposer as a verbatim user message — no system-prompt wrapping, no skill framing, no slash-command template.** The proposer should react to critic comments the same way it would react to the human pasting a code review: full normal defense behavior, no "I'm in skill mode following steps" override.

This rules out skills, slash commands, and any plugin-layer prompt template as the delivery channel.

### Payload via file reference

The verbatim user message stays short. The critic's full critique is written to disk, and the message just points at it:

```
Some comments at @.debate/sessions/<id>/rounds/r1-critic-1.md. Please resolve or respond. If you disagree, please raise it.
```

This is exactly how a human hands an agent a code review in normal Claude Code usage — `@file, please address`. The proposer's defense behavior is whatever Claude does with that pattern in any other session. No special protocol, no system-prompt-shaped instruction.

Three reasons this matters:

- **Token economy.** Long critiques don't bloat every defense round's context. The file is read on demand.
- **Audit trail.** Each round's critique is a durable artifact under `.debate/sessions/<id>/rounds/`, not a chat-buffer scrap.
- **Naturalism preserved.** The pointer message is short and natural; the critique text is whatever the critic wrote, untouched.

The mechanism end-to-end:

- **Critic → proposer.** Orchestrator captures the critic's output, writes it to `forks/critic-<i>/rounds/r<n>-critic.md`, then dispatches the pointer to the fork's proposer-clone (via `--fork-session` on the first round, `--resume <fork-id>` on subsequent rounds).
- **Proposer → critic.** Symmetric. Orchestrator captures the proposer-clone's chat response (and notes which files were modified — concession-fixes show up as a diff in the working tree), writes them to `forks/critic-<i>/rounds/r<n>-proposer.md`, then dispatches a short pointer to the critic: `"Proposer responses at <path>. Review the defenses; for any unresolved attack, decide whether to re-attack or withdraw."` Codex reads the file via its file-access tool.

The pointer messages are the only orchestrator-authored text the agents see. Everything substantive — critique, defense, re-attack — lives in files and travels by reference.

### Rounds

Per fork (forks run serially in v0):

1. **R0 — Setup.** Orchestrator extracts task context from the root session's transcript, computes the working-tree diff, and forks the proposer via `claude --resume <root-id> --fork-session ...`.
2. **R1 — Attack.** Critic Cᵢ produces a structured attack list against the diff. Each attack names a concrete leaf: `{location, claim, expected violation, reproduction}`. Orchestrator persists output to `forks/critic-<i>/rounds/r1-critic.md`. Attacks lacking a reproduction are dropped at parse time, not by the proposer.
3. **R2 — Defense.** Proposer-clone responds with one of: `concede` (and apply fix), `rebut` (with specific counter-evidence), `push-back` (request clarification — only allowed once per attack). Orchestrator persists the response into `forks/critic-<i>/rounds/r2-proposer.md`; concession-fixes show up as a diff in the real working tree.
4. **R3..R(max_turn) — Cross-examination.** Critic and proposer-clone alternate, each round persisted to its file. Per-claim resolution state advances or stalls.
5. **Fork-wrap.** When this fork's termination condition fires, write its per-fork attack ledger.

After all forks complete: aggregate across forks, write `summary.md` and `end.json`, print to stdout. Root session is untouched.

## Build options

The channel constraint above eliminates most of the design space. Anything that wraps the critic's output in framing (skills, slash commands, plugin command templates) is out. What remains:

### Option A — CLI binary only

A standalone `debate` orchestrator. User invokes from terminal after a coding session.

```
debate --max-turn=10 --main claude --side codex --side-count=3 \
       --aspect correctness,security,perf --session-id <claude-session>
```

The CLI uses `claude --resume <id>` to inject critic comments as user turns into the existing Claude Code session, and `codex exec` to run the critic.

- Pro: clean channel (verbatim user-message injection); composable; works in CI; trivially supports more critics.
- Con: not auto-triggered; user must invoke manually after each session.
- Verdict: **this is the primitive.** Build first.

### Option B — CLI + opt-in Stop hook (recommended)

Same CLI as A. Add a Stop hook in `.claude/settings.json` that captures the just-finished session ID and invokes the CLI against it.

- Pro: automatic triggering for users who want it; manual invocation still works.
- Con: requires Stop hook to be able to capture session ID and the CLI to drive `--resume` against a session that just stopped (verify mechanics in a spike before committing).
- Verdict: **recommended once A works.** Hook is two lines of shell that calls the CLI; no plugin, no skill, no slash command.

### Rejected options

- **Slash command (`/debate`)**: violates the channel constraint — slash commands inject a template into the conversation. Even if the template only said "run the debate process," that's still a system-prompt-shaped artifact the proposer sees before the critic's text.
- **Skill (`debate-review` or `debate-defense`)**: same reason. Skills carry instructions Claude follows. The whole point is that Claude follows its *normal* coding-feedback instincts when responding to the critic, not a skill-specific methodology.
- **Plugin packaging**: premature productization. A two-line hook + a CLI binary doesn't need a plugin manifest. Revisit if a second user shows up.

## CLI surface

```
debate [--main claude] [--side codex] [--side-count 1]
       [--max-turn 6] [--aspect general]
       [--session-id <root-claude-session-id>]
       [--transcript <path-to-jsonl>]
       [--diff-from HEAD] [--diff-to .]
       [--task-context "<original task>"]
       [--judge none|llm|human]
       [--cost-cap 50000]
       [--state-dir .debate]
       [--format markdown|json]
```

Notes:

- `--session-id` is the **root** session ID. The orchestrator forks from it for each critic via `claude --resume <root> --fork-session`. The root session is never modified. Without `--session-id`, the orchestrator falls back to fresh `claude -p` invocations per round (no proposer continuity within a fork — much more expensive).
- `--transcript` is optional but useful: the Stop hook payload includes `transcript_path` pointing at the root session's JSONL. Passing it lets the orchestrator extract task context cheaply (no second `claude` call to inspect the session).
- `--side-count` and `--aspect` interact: if `--aspect a,b,c` is given with `--side-count 3`, each critic gets one aspect. If counts mismatch, error.
- `--max-turn` counts P+C exchanges combined per fork. 6 = 3 attack rounds + 3 defense rounds within one fork. With `side-count=3` and max-turn=6, the worst case is 18 round-exchanges total (serial across forks).
- `--task-context` is mandatory when neither `--session-id` nor `--transcript` is given. Otherwise the orchestrator extracts it from the first user turn in the transcript.
- `--cost-cap` is mandatory and aborts gracefully (surfaces partial review). Multi-critic multi-turn debates blow token budgets fast.
- Exit code 0 if zero unresolved leaves, 1 otherwise. Lets it gate CI.

## Hook surface (optional)

The Stop hook receives a JSON payload on stdin (per Claude Code hook docs) containing `session_id`, `transcript_path`, `cwd`, `stop_reason`, and `output`. The hook script reads `session_id` from that payload and invokes the CLI:

```json
{
  "hooks": {
    "Stop": [{
      "command": "jq -r '.session_id' | xargs -I{} debate --session-id {} --max-turn 6"
    }]
  }
}
```

Or as a script if you want more logic (skip on trivial diffs, etc.):

```bash
#!/usr/bin/env bash
PAYLOAD=$(cat)
SESSION_ID=$(echo "$PAYLOAD" | jq -r '.session_id')
TRANSCRIPT=$(echo "$PAYLOAD" | jq -r '.transcript_path')
exec debate --session-id "$SESSION_ID" --transcript "$TRANSCRIPT" --max-turn 6
```

That's it. No plugin manifest, no slash command, no skill. The hook is one stdin-read because all the work is in the CLI. Default: hook is **not installed**; users opt in by adding the entry. The hook should return exit 0 promptly so Claude Code's UI doesn't appear hung — the orchestrator runs in the foreground and prints to stdout.

## Session persistence

Each debate run is a session with explicit start and end markers and an on-disk record. This is what makes the orchestrator auditable and the "review unresolved later" workflow possible.

### Layout

Each invocation creates a folder under `--state-dir` (default `.debate/`):

```
.debate/
  log.jsonl                          # one line per debate run, appended at end
  sessions/
    <ISO8601>-<short-id>/
      start.json                     # timestamp, root-session-id, task-context, diff snapshot, config
      forks/                         # one folder per critic-fork (mirrors the tree shape)
        critic-1/
          fork-session-id            # the proposer-clone's session ID, captured from --fork-session
          rounds/
            r1-critic.md             # critic 1's R1 attack list (full text)
            r2-proposer.md           # proposer-clone's R2 defense
            r3-critic.md             # critic's response to R2
            ...                      # one file per round, durable as it happens
        critic-2/
          fork-session-id
          rounds/
            ...
      transcript.jsonl               # append-only index across forks: pointers to round files
      attacks.jsonl                  # per-attack records aggregated across forks
      summary.md                     # human-facing summary, written at termination
      end.json                       # termination condition, stats, exit code
```

### Lifecycle invariants the orchestrator must enforce

- `start.json` written atomically before any agent process spawns.
- Per-fork: `fork-session-id` written immediately after `--fork-session` returns, before R1's pointer is dispatched.
- Round files are written before the corresponding pointer message is dispatched. The agent receiving the pointer can always Read the file — no race.
- `transcript.jsonl` and `attacks.jsonl` are append-only — never rewrite, never seek-back. A killed process leaves a valid (truncated) record.
- `summary.md` and `end.json` are written only at termination (clean or interrupted).
- `log.jsonl` is appended last, after `end.json` is durable. A run with `end.json` missing is an interrupted session; user can inspect `forks/<i>/rounds/` directly.
- **Root session is never modified.** No `claude --resume <root-id>` without `--fork-session`. Verifiable invariant: the root session's transcript file timestamp should be unchanged after a debate run completes.

### Surfacing rule

- **Zero unresolved leaves at termination**: orchestrator is silent on stdout except for one line referencing the `log.jsonl` entry. No summary file is opened or surfaced. `summary.md` is still written for audit, but the user is not interrupted.
- **≥ 1 unresolved leaves**: orchestrator prints the path to `summary.md` plus the *headline contradicting signal* (see below) on stdout. Exit code 1.
- **Interrupted (Ctrl-C, cost-cap, malformed-output)**: same as ≥ 1 unresolved — surface what's there.

### Headline contradicting signal

Among unresolved leaves, the headline is the attack with the highest **contention score**:

```
contention(attack) = rounds_survived + (1 if critic re_attacked_after_defense else 0)
```

Tie-break by first appearance. This is deliberately a cheap rule with no LLM scoring — adding semantic-scoring would re-introduce the lazy-judge problem at the headline step. If contention turns out to be a poor proxy in practice, upgrade later by having both sides self-report confidence each round and weighting by `confidence_critic × confidence_proposer`.

### `.gitignore`

`.debate/` should not be committed. Orchestrator checks `.gitignore` on first run and prints a warning (not a hard error) if `.debate/` is missing from it. Doesn't auto-edit the file — that's the user's call.

## Configuration

Project-level `.debate.toml` overrides defaults:

```toml
max_turn = 6
side_count = 2
aspects = ["correctness", "security"]
cost_cap_tokens = 50000
trigger = "manual"           # "manual" | "stop"
allow_style_attacks = false  # default: critic must attack behavior, not style
```

## Termination conditions

The mediator terminates and writes `summary.md` + `end.json` when any of these fire:

1. **Steady state** — no new attacks two rounds running.
2. **Max turn** — hard cap reached.
3. **Cost cap** — token budget hit.
4. **Malformed output** — critic produces ill-formed attacks two rounds running (defensive: model is broken or prompt collapsed).
5. **User interrupt** — Ctrl-C; the orchestrator's signal handler still writes `end.json` before exiting.

The summary header names which condition fired, so the human knows whether to trust "0 unresolved" (steady state) or treat it as truncation (max-turn / cost-cap). Only steady-state termination with zero unresolved is "clean" — every other condition surfaces.

## Output format

`summary.md` structure:

```
# Debate review — terminated: steady-state | max-turn | cost-cap | interrupted

## Headline (most contested unresolved)
- [security/api.py:88] Input sanitization
  - Critic: framework auto-escape doesn't cover `LIKE` patterns.
  - Proposer: parameterized via SQLAlchemy.
  - **Stake**: input `'; DROP TABLE--%` against the search endpoint.
  - Contention: 3 rounds survived, critic re-attacked after defense.

## Other unresolved (2, sorted by contention)
- ...

## Resolved (12)
- [conceded] Off-by-one in pagination loop → fixed at src/api.py:42
- [rebutted] Race condition claim → critic withdrew after seeing lock at api.py:18

## Stats
critic-found-real-issue rate: 7/15 attacks led to a fix
debate cost: 38k tokens, 6 turns, 2 critics
session: .debate/sessions/2026-05-05T14-22-31-a3f9b1/
```

When unresolved count is zero, `summary.md` still has the Resolved + Stats sections (no Headline, no Unresolved) and is written but not surfaced. The user only sees one line on stdout pointing at `.debate/log.jsonl`.

The Headline section is the entire justification for the tool. If it's noise across many sessions, the tool fails — and the cross-session `log.jsonl` makes that measurable rather than vibes-based. The Stats block lets the user spot-check whether the critic is actually working: if `critic-found-real-issue rate` trends near 0, disable the hook.

## Risks

- **Lazy-critic risk (the binding one).** This is upstream spec 07's H6. Codex may rubber-stamp. **Mitigation**: do not ship the tool until upstream 07a (bug detection, deterministic judge) returns critic-found-bug rate ≥ 60% on seeded bugs. If it's below 30%, this whole project is dead on arrival.
- **Cost.** Multi-critic multi-turn debates 5–10x a coding session's token bill. Cost cap is mandatory; default it conservatively (50k tokens).
- **Flow disruption.** Auto-Stop on every completion is annoying for trivial edits. Default trigger is `manual`; auto is opt-in. Consider a `--changed-lines-min N` filter (only debate if diff is non-trivial).
- **Critic context starvation.** Critic only sees diff + task context, not the broader codebase. Produces false-positive attacks ("this function isn't called!" — yes it is, elsewhere). Mitigations: critic prompt requires concrete reproduction, and the proposer is allowed to rebut with `file:line` references the critic is forbidden from re-attacking.
- **Stylistic-gripe drift.** Critic attacks style not behavior. `allow_style_attacks = false` enforces it in the critic prompt; mediator drops style-shaped attacks at parse time.
- **Asymmetric truth.** Proposer has more context than critic; may over-defend when actually wrong. Mitigation: `--judge llm` mode triages unresolved leaves; default `none` just surfaces them and trusts the human.
- **Critic colludes with proposer.** If critic and proposer share a model family, they're more likely to agree on the same wrong answer. **Heterogeneity is structural here, not optional** — Claude proposer + Codex critic is the default for a reason. Don't ship a Claude-vs-Claude default.

## Out of scope

- **Skill or slash-command entry points.** Both wrap critic output in framing that distorts the proposer's response. The channel constraint says verbatim user-message via `claude --resume` only.
- **Plugin packaging.** Two-line hook + CLI binary doesn't need a manifest. Revisit only if multiple unrelated users adopt it.
- **Injecting into the root session.** Claude Code provides no way to add an assistant turn to an existing session, and the natural alternatives (`additionalContext` from `SessionStart`/`UserPromptSubmit` hooks) all produce system-reminder-shaped messages that violate the channel constraint. Wrap-up is stdout-only.
- **Parallel forks via per-fork worktrees.** Tempting (`claude --worktree` plus per-fork git worktrees would isolate the working tree per fork and allow parallel execution), but raises a concession-merge problem when two critics both ask for fixes that conflict. v0 is serial; revisit when the basic flow is proven and the conflict story is worth solving.
- Training a better critic. Use whatever Codex gives us; if it's bad, the tool fails (and that's the right outcome).
- Streaming TUI. v1 is batch.
- Persistent debate state across user sessions. Each invocation is fresh against the current diff.
- Critic tool access beyond reading the diff and the task context. Deliberate — keeps the critic narrow and prevents critic-rabbit-holes.
- Auto-applying critic-suggested fixes. Concession-fixes are written by the proposer-clone within its fork, not by the critic. The critic never edits.

## Relationship to upstream research

This tool is the productization of the debate architecture studied in [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance). Specifically it bets on the architecture from [spec 07](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md) and inherits its risks.

**This tool should not be built before upstream spec 07a returns positive on H1 (debate beats voting at equal compute) and not-bitten on H6 (lazy critic).** Order of work:

1. Run upstream 07a (bug detection, deterministic judge, 2-player honest debate) — measure critic-found-bug rate.
2. If ≥ 60%: build this repo's CLI orchestrator (Option A).
3. If 07a also shows H1 holds: add the Stop hook (Option B).
4. If 07's H6 bites (lazy critic): re-prompt critic, re-measure; if still bad, abandon the tool.

This ordering means the tool is gated on a real measurement, not a hope.
