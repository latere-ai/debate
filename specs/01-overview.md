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

### When the proposer is Codex instead of Claude

Codex (0.128+) has session persistence and supports interactive `/fork` and `/resume` slash commands inside the TUI, plus `codex exec resume <id> "<prompt>"` for non-interactive resume. **But it has no non-interactive fork.** Verified by probe (2026-05, codex 0.128.0):

- `codex fork <id> "<prompt>"` errors with `"Error: stdin is not a terminal"` when invoked from a script — TUI-only entry point.
- `codex exec resume <id> "<prompt>"` works non-interactively but **modifies the resumed session in place** (the session file grows; `thread_id` does not change).
- `--ephemeral` does not save us: it suppresses persistence for *fresh* sessions only, not resumed ones. Resume-with-ephemeral still pollutes the resumed file.

|                                       | Claude (2.1.x)                                  | Codex (0.128+)                                 |
|---------------------------------------|-------------------------------------------------|------------------------------------------------|
| Non-interactive fork                  | `claude --resume <id> --fork-session -p ...` ✅ | None ❌ (`codex fork` requires TTY)            |
| Non-interactive resume                | `claude --resume <id> -p ...` ✅                | `codex exec resume <id> "<prompt>"` ✅         |
| Resume modifies resumed session?      | Only the fork (root preserved)                  | **Yes** — session file grows in place          |
| `--ephemeral` prevents that?          | N/A                                             | No (applies only to fresh sessions)            |
| Auto-trigger via Stop hook            | Yes                                             | No equivalent                                  |
| Sandbox flag on resume                | `--permission-mode` accepted                    | `--sandbox` rejected on `exec resume` (inherits from parent) |

**Implication.** When `--main codex`, the orchestrator cannot create a non-mutating fork off the user's codex session. Instead, each defense round is a **stateless `codex exec`** with the full context re-supplied as the prompt:

```
codex exec --skip-git-repo-check --sandbox <mode> --json \
  "<task + current diff + all prior rounds in this fork>"
```

- Each round produces a fresh `thread_id` (no continuation of the prior round). The orchestrator carries fork state on disk in the round files and re-feeds it each round.
- No cache amortization across rounds — input tokens paid in full each time. This is the dominant cost difference vs. claude mode.
- The user's existing codex session (if any) is not touched: the orchestrator's `codex exec` calls don't resume any prior session.
- Auto-trigger is not available; only manual CLI invocation.
- Capture `thread_id` from the first JSON event on stdout (`{"type":"thread.started","thread_id":"<uuid>"}`); the final response is in the `item.completed` event with `type: "agent_message"`.

The CLI's `--main` and `--side` flags remain symmetric — any pairing works. Architectural contracts (channel constraint, file-pointer payload, contention scoring, headline output) are identical in both modes; only the proposer-driving mechanism differs.

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

### Verified primitives (2026-05, claude 2.1.129)

The mechanism above was probed against a real Claude Code installation before this spec was finalized. Findings:

- **`claude --resume <root> --fork-session -p "..." --output-format json` works.** Returns a JSON result with a new `session_id` (different from root). Root JSONL is bit-identical preserved (mtime and size unchanged) across fork creation. ✅
- **`claude --resume <fork-id> -p "..."` continues the fork in place.** Same session_id across rounds (no fork-of-fork); fork JSONL grows; root JSONL still bit-identical. ✅
- **The fork inherits root's full conversation history.** When the fork is asked "what have you said so far," the response includes both the root's prior turns and the new fork turns. The proposer-clone has the original task + code in context, exactly as required. ✅
- **The fork JSONL is stored alongside root** under `~/.claude/projects/<encoded-cwd>/<fork-id>.jsonl`. ✅

#### Constraints uncovered by the probe (must inform implementation)

- **`--resume` is cwd-scoped.** Running `claude --resume <id>` from any cwd other than the one the session was created in returns *No conversation found with session ID*. The orchestrator must `cd` to the cwd captured in the Stop hook payload before any `claude --resume` call. This is not optional.
- **`ANTHROPIC_API_KEY` env var pollution.** If set (even to a stale/invalid value), `claude -p` uses it instead of the OAuth keychain and fails with 401. The hook script must either `unset ANTHROPIC_API_KEY` before invoking the orchestrator, or document the precondition. (This bites in subprocesses inheriting env from a parent shell with a stale key.)
- **JSON output may contain control characters in `result`** that break naive parsers. The orchestrator must use a proper JSON library (Python `json`, Go `encoding/json`), not `jq` with raw shell pipes.
- **First-call cost.** First `claude -p` invocation in a fresh window primes a large system-prompt cache (~32k tokens, ~$0.20). Subsequent calls within the 5-minute cache window are cheap. Run all rounds of a debate in close succession to amortize.

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

## Critic specialization

Multi-critic isn't a frill — it's the structural mitigation for lazy-critic risk. The debate-theoretic property (Irving 2018, building on PCP intuitions): **soundness needs one competent honest player, not all players honest**. Translated to multi-aspect critics: bugs in aspect *i* need only the critic on aspect *i* to be honest and competent. A lazy critic on aspect *j* doesn't break aspect *i* coverage.

This pushes the binding empirical question from "is THE critic competent?" (whole-tool risk) to "is the critic on aspect *X* competent?" (per-aspect risk). The latter is much more tractable: a weak aspect gets dropped from defaults without killing the tool.

### Default aspect set

The default `--aspect` list is a production-quality bar, not a pure bug-detection set:

- **functional-logic** — does it do what the task asked? Off-by-ones, missing branches, silent-failure paths, edge cases the spec implies but the code missed.
- **security** — input validation, authn/authz, injection (SQL, shell, template), data exposure, secrets in logs, unsafe deserialization.
- **code-quality** — maintainability red flags: long functions, swallowed exceptions, unclear naming *where it bites readability*, dead branches. **Not** style preferences.
- **performance** — algorithmic complexity, N+1 IO, unnecessary allocations or copies, blocking calls in hot paths.

Each critic is prompted with exactly one aspect, and its system prompt explicitly forbids cross-aspect attacks. The mediator drops attacks tagged with the wrong aspect at parse time. This narrows each critic's mandate, which is where lazy rubber-stamping hides — a generalist critic asked "review this" is much more likely to say "looks fine" than a specialist critic asked "find concrete security flaws in this diff."

### Aspect ≠ style

`code-quality` is not a license for stylistic gripes:

- **Code quality**: "This function silently swallows the exception, breaking the calling contract." (real correctness/maintainability impact)
- **Style**: "Use single quotes instead of double quotes." (no impact)

The default critic prompt requires every attack to name a concrete behavior or maintainability impact, not a preference. Style-shaped attacks are dropped at parse time.

### Aspect list is open

The four defaults cover typical backend coding work. Other teams might add `concurrency-safety` for systems code, `accessibility` for frontend, `api-compatibility` for libraries, `migration-safety` for schema/state changes. Aspect names are free-form and become part of the critic's system prompt; the mediator just uses them as routing labels and headline tags.

## Heterogeneity (proposer vs. critic)

The default pairing is cross-family: claude proposer, codex critic. Family asymmetry is the strongest form of independence — different training corpora, different RLHF objectives, different priors. It's the cleanest way to keep proposer and critic from rubber-stamping each other.

But sometimes only one agent is available, or a user wants same-family debate (claude/claude, codex/codex). The tool supports this with one hard constraint:

**When `--main` and `--side` are the same agent family, `--main-model` and `--side-model` must both be set and must differ.** Same model on both sides collapses to "the model debating itself" — same priors, same blind spots, no heterogeneity. The CLI errors out if either is unset or if they match.

Recommended same-family pairings:

- **claude/claude**: e.g. `--main-model claude-sonnet-4-6 --side-model claude-opus-4-7`. Different capability tiers or training generations.
- **codex/codex**: e.g. `--main-model gpt-5 --side-model o3`. Different model families inside codex.

Cross-family pairings don't require explicit model flags — the family difference already provides independence. Model flags are optional in those cases and default to each agent's CLI default.

For multi-critic (`--side-count > 1`), `--side-model` applies to all critics; aspect specialization provides the per-critic diversity. Per-critic model config is out of scope for v0.

## Build options

The channel constraint above eliminates most of the design space. Anything that wraps the critic's output in framing (skills, slash commands, plugin command templates) is out. What remains:

### Option A — CLI binary (the primitive)

A standalone `debate` orchestrator. Always built first; everything else layers on it.

```
debate --max-turn=10 --main claude --side codex --side-count=3 \
       --aspect correctness,security,perf --session-id <claude-session>
```

Uses `claude --resume <id>` to drive forks (claude-as-proposer mode) or fresh `codex exec` per round (codex-as-proposer mode). Required as-is for: codex-as-proposer mode, CI gating, scripted batch runs, and as the backend for Option B.

### Option B — CLI + Stop hook (default UX for claude-as-proposer)

Hook fires when claude finishes responding, invokes the CLI synchronously, prints progress and the final summary in the user's terminal, exits. This is the optimal UX: the user codes normally, debate fires automatically, summary appears in the same session.

- Pro: zero workflow friction. User doesn't have to remember to run `debate` after every session.
- Con: every claude stop triggers the orchestrator unless gated; the gate (`--changed-lines-min`) is essential to avoid debating trivial completions.
- Verdict: **default for claude-as-proposer.** The hook is two lines of shell that invokes the CLI.

### Rejected options

- **Slash command (`/debate`)**: violates the channel constraint — slash commands inject a template into the conversation. Even if the template only said "run the debate process," that's still a system-prompt-shaped artifact the proposer sees before the critic's text.
- **Skill (`debate-review` or `debate-defense`)**: same reason. Skills carry instructions Claude follows. The whole point is that Claude follows its *normal* coding-feedback instincts when responding to the critic, not a skill-specific methodology.
- **Plugin packaging**: premature productization. A two-line hook + a CLI binary doesn't need a plugin manifest. Revisit if a second user shows up.

## CLI surface

```
debate [--main claude] [--side codex] [--side-count 1]
       [--main-model <model>] [--side-model <model>]
       [--max-turn 6] [--aspect general]
       [--session-id <root-claude-session-id>]
       [--transcript <path-to-jsonl>]
       [--diff-from HEAD] [--diff-to .]
       [--task-context "<original task>"]
       [--judge none|llm|human]
       [--cost-cap 50000]
       [--changed-lines-min 10]
       [--state-dir .debate]
       [--format markdown|json]
```

Notes:

- `--session-id` is the **root** session ID (Claude-as-proposer mode only). The orchestrator forks from it for each critic via `claude --resume <root> --fork-session`. The root session is never modified. Without `--session-id`, the orchestrator falls back to fresh `claude -p` invocations per round (no proposer continuity within a fork — much more expensive). When `--main codex`, this flag is ignored (codex has no non-mutating fork; see codex section).
- `--main-model` and `--side-model` are optional when `--main` and `--side` are different agent families (cross-family asymmetry suffices). When the families match, both flags are required and must differ — see Heterogeneity section. CLI errors out otherwise.
- The orchestrator must be invoked from the cwd that owns the root session — `claude --resume <id>` is cwd-scoped. The hook-supplied `cwd` field is authoritative. The CLI errors out if invoked from a different cwd.
- `--transcript` is optional but useful: the Stop hook payload includes `transcript_path` pointing at the root session's JSONL. Passing it lets the orchestrator extract task context cheaply (no second `claude` call to inspect the session).
- `--side-count` and `--aspect` interact: if `--aspect a,b,c` is given with `--side-count 3`, each critic gets one aspect. If counts mismatch, error.
- `--max-turn` counts P+C exchanges combined per fork. 6 = 3 attack rounds + 3 defense rounds within one fork. With `side-count=3` and max-turn=6, the worst case is 18 round-exchanges total (serial across forks).
- `--task-context` is mandatory when neither `--session-id` nor `--transcript` is given. Otherwise the orchestrator extracts it from the first user turn in the transcript.
- `--cost-cap` is mandatory and aborts gracefully (surfaces partial review). Multi-critic multi-turn debates blow token budgets fast.
- `--changed-lines-min` is the trivial-diff gate. Below the threshold, the orchestrator prints one status line (`[debate] skipped: trivial diff`) and exits fast. Critical when the Stop hook is auto-triggering on every claude session-stop.
- Exit code 0 if zero unresolved leaves, 1 otherwise. Lets it gate CI.

## Trigger via Stop hook (default for claude-as-proposer)

The Stop hook is the **default install path**, not optional. It's how the optimal UX is delivered: user opens claude interactively, codes normally, and when claude finishes responding the debate orchestrator runs in the same terminal, prints progress, prints the final summary, and returns control.

### Hook configuration

The Stop hook entry in `.claude/settings.json` must use the **verbose format** (the simpler `{"command": "..."}` style is silently dropped from the registry — verified against claude 2.1.129):

```json
{
  "hooks": {
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/debate-stop-hook.sh"
          }
        ]
      }
    ]
  }
}
```

The hook script:

```bash
#!/usr/bin/env bash
PAYLOAD=$(cat)
SESSION_ID=$(echo "$PAYLOAD" | jq -r '.session_id')
TRANSCRIPT=$(echo "$PAYLOAD" | jq -r '.transcript_path')
CWD=$(echo "$PAYLOAD" | jq -r '.cwd')

# Stale ANTHROPIC_API_KEY in env causes 401 in claude -p subprocesses
unset ANTHROPIC_API_KEY

# --resume requires running from the cwd that owns the project's session dir
cd "$CWD" || exit 1

# Run the orchestrator in foreground; on exit, emit a single
# systemMessage with the summary path so claude's next response
# (or session) can reference it.
SUMMARY=$(debate --session-id "$SESSION_ID" --transcript "$TRANSCRIPT" --max-turn 6 --format markdown)
SUMMARY_PATH=$(echo "$SUMMARY" | grep -oE '\.debate/sessions/[^ ]+/summary\.md' | head -1)

# Stop event's JSON output schema accepts: continue, suppressOutput,
# stopReason, decision, reason, systemMessage. NOT additionalContext.
jq -n --arg msg "Debate review at @$SUMMARY_PATH" '{"systemMessage":$msg}'
exit 0
```

The hook payload contains `session_id`, `transcript_path`, `cwd`, `stop_reason`, and `output` as JSON on stdin. No plugin manifest, no slash command, no skill — all the work lives in the `debate` CLI; the hook just routes the payload.

**Project-level vs user-level settings.** Project `.claude/settings.json` works (claude reads it), but my probe of `claude -p` showed hooks defined there can be filtered if the project isn't trusted. The cleanest install path is *user-level* `~/.claude/settings.json` for tools meant to apply across projects, or `.claude/settings.json` accepted via a one-time interactive trust prompt for project-specific config.

### What the user sees during debate

**Probed against `claude -p` (claude 2.1.129):** hooks fire correctly, but **none of the obvious channels surface to user-facing output streams in `-p` mode**:

| Channel from hook | Visible to user in `-p`? |
|---|---|
| Plain stderr | ❌ No |
| Plain stdout | ❌ No |
| JSON `systemMessage` (valid for Stop) | ❌ Not in captured stdout/stderr |
| JSON `additionalContext` (for Stop) | ❌ Schema rejects it for Stop event |

**Implication.** Live per-fork progress in the user's claude terminal — the original UX I sketched — is **probably not deliverable** via standard hook channels. Interactive mode rendering (TUI) is *unverified* (I cannot drive an interactive claude from inside this agent), so it's possible interactive does surface hook output where `-p` doesn't. But the conservative assumption is that it doesn't, and the spec must work in that case.

The realistic UX is two-phase:

1. **At Stop**: hook fires, runs orchestrator synchronously to completion. Claude is "stopping" and the prompt is unavailable until orchestrator exits. **No mid-flight progress is shown to the user** — they see nothing, then the prompt returns. Typical run: 30s–3min.
2. **After hook returns**: hook emits a `systemMessage` JSON with the summary path. Claude renders this as a system reminder on the next user-visible context. The user opens `summary.md` to see results.

If interactive mode turns out to render hook output cleanly (worth verifying — see below), upgrade to live progress: the orchestrator writes per-fork status lines to stderr, and the user watches them scroll. But this should be treated as an *enhancement* gated on verification, not the design baseline.

UX properties (conservative baseline):

- **Synchronous.** Prompt is unavailable until hook exits.
- **Cancellable.** Ctrl-C in the user's terminal sends SIGINT through the process tree to the orchestrator. Orchestrator catches it, writes `end.json` with `terminated: interrupted`, exits non-zero. Claude finishes stopping normally.
- **Trivial-diff gate.** `--changed-lines-min 10` filters out single-line edits etc. Below threshold the hook returns in <100ms with a `systemMessage` saying "debate skipped: trivial diff."
- **Surfacing**: only via `systemMessage` rendered as a system reminder; the user's recourse is `summary.md`.

### Interactive verification (a 30-second test you should run before v0)

Live in-session progress depends on claude's interactive TUI rendering at least *one* hook channel. Easiest probe:

```bash
mkdir -p /tmp/dbg-probe/.claude
cat > /tmp/dbg-probe/.claude/settings.json <<'EOF'
{
  "hooks": {
    "Stop": [{
      "matcher": "",
      "hooks": [{
        "type": "command",
        "command": "echo PROBE_STDERR >&2; echo '{\"systemMessage\":\"PROBE_SYSMSG\"}'; exit 0"
      }]
    }]
  }
}
EOF
cd /tmp/dbg-probe
claude   # interactive. Type any prompt, observe what (if anything) appears around the response end.
```

What to look for:
- `PROBE_STDERR` appearing in the terminal → hook stderr surfaces; live-progress UX is viable.
- `PROBE_SYSMSG` appearing as a system note in the chat → systemMessage renders in interactive (and we should use it for the summary surface).
- Neither → conservative two-phase UX is the only path; v0 should not promise live progress.

(The first run may need to accept a project-trust prompt; that's expected.)

### Manual invocation (fallback)

For codex-as-proposer (no Stop hook equivalent), CI gating, or running debate against a saved session out-of-band, invoke the CLI directly:

```
debate --session-id <root-claude-session-id> --max-turn 6
```

Or with codex as proposer:

```
debate --main codex --side claude --main-model gpt-5 --side-model claude-sonnet-4-6 \
       --task-context "$(< task.md)" --diff-from HEAD~1
```

Manual invocation is the only path for codex-as-proposer; for claude-as-proposer it's a fallback when the Stop hook isn't appropriate (CI, batch runs, debugging).

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

The general rule: any file referenced by an `@<path>` pointer must exist on disk before the agent receives the pointer.

- `start.json` written atomically before any agent process spawns.
- **R1 ordering is the unique case.** `forks/critic-<i>/rounds/r1-critic.md` is written first (orchestrator runs the critic in a fresh `codex exec`/equivalent and persists its output). Only then does the orchestrator run `claude --resume <root> --fork-session -p "<pointer to r1-critic.md>" --output-format json`, which creates the fork *and* delivers R1 in one shot. The fork's session ID is captured from that call's JSON output (the `session_id` field) and written to `forks/critic-<i>/fork-session-id` *immediately on return* — there is no earlier moment to write it, since the ID does not exist before the call.
- The proposer-clone's R2 defense text is in the same call's JSON `result` field; it's persisted to `r2-proposer.md` after the call returns.
- For R3 onward (within the same fork): each round file is written before the orchestrator dispatches the corresponding pointer via `claude --resume <fork-id> -p "<pointer>"`. By that point the fork already exists, so file-then-pointer ordering is straightforward.
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
side_count = 4
aspects = ["functional-logic", "security", "code-quality", "performance"]
cost_cap_tokens = 50000
trigger = "stop"             # "stop" (default for claude-as-proposer) | "manual"
allow_style_attacks = false  # default: code-quality critic attacks impact, not preference

# Models. Optional when main and side are different agent families.
# Required (and must differ) when main and side are the same family.
# main = "claude"
# side = "codex"
# main_model = "claude-sonnet-4-6"  # only used when main == side family
# side_model = "claude-opus-4-7"
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

- **Per-aspect lazy-critic risk** (was previously framed as "the binding one"; that overgeneralized the single-critic case). A lazy critic produces no value in its aspect. With multi-critic across distinct aspects, no single lazy critic collapses the whole tool — only its aspect goes uncovered. The binding question becomes **per-aspect**: for each aspect we ship as default, is the critic prompt + model competent enough to find real instances? Mitigation: measure critic-found-bug rate *per aspect* in upstream 07a; aspects below threshold get dropped from defaults rather than the tool being abandoned. The debate-theoretic intuition (one competent honest player suffices for soundness) is what makes this work.
- **Cost.** Multi-critic multi-turn debates 5–10x a coding session's token bill. Cost cap is mandatory; default it conservatively (50k tokens).
- **Flow disruption.** Auto-Stop on every claude completion would fire on trivial edits (typo fixes, single-line changes), which is the dominant failure mode of "always-on review tools." Mitigation is structural, not opt-in: `--changed-lines-min` (default 10) gates debate at the orchestrator entry point, so the hook returns in milliseconds for trivial diffs. The user sees one status line confirming the gate fired, not a debate run.
- **Critic context starvation.** Critic only sees diff + task context, not the broader codebase. Produces false-positive attacks ("this function isn't called!" — yes it is, elsewhere). Mitigations: critic prompt requires concrete reproduction, and the proposer is allowed to rebut with `file:line` references the critic is forbidden from re-attacking.
- **Stylistic-gripe drift (especially in `code-quality` aspect).** Critic drifts from real maintainability impact into formatting/naming preferences. `code-quality` is the most exposed aspect because the line between "real quality issue" and "preference" is fuzzier than for `security` or `performance`. Mitigation: critic prompt requires every attack to name a concrete behavior or maintainability impact; mediator drops style-shaped attacks at parse time (heuristic: attack contains "should be" + naming/formatting language without a concrete behavior claim).
- **Asymmetric truth.** Proposer has more context than critic; may over-defend when actually wrong. Mitigation: `--judge llm` mode triages unresolved leaves; default `none` just surfaces them and trusts the human.
- **Critic colludes with proposer (same family + same model).** Same agent family + same model = the model debating itself, with the same priors and same blind spots. The CLI rejects same-model configurations when families match (`--main-model` and `--side-model` must differ). Same-family/different-model is allowed but weaker than cross-family — use it only when cross-family isn't available (single-agent install). The default `--main claude --side codex` provides cross-family heterogeneity automatically.

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

**Spec 08 should not be built before upstream spec 07a returns positive results, measured per aspect.** What "positive" means with aspect-specialized critics:

1. Run upstream 07a per aspect (functional-logic, security, code-quality, performance) — aspect-specialized critics, not one generalist.
2. For each aspect, measure critic-found-bug rate on seeded bugs of that aspect.
3. **Default aspects = aspects where the rate is ≥ 60%.**
4. **Acceptable thresholds**: at least two aspects pass; otherwise the tool is hollow. If only one aspect passes, it's a single-aspect linter, not a debate tool.
5. If H1 also holds (debate beats voting at equal compute, on per-aspect tasks): add Option B's Stop hook.

The previous order-of-work conflated all critics into one threshold. With aspect specialization, the failure of one aspect doesn't kill the tool — it just narrows the default set. Upstream 07's H6 ("lazy critic collapses the architecture") is real but only at the per-aspect level.
