# debate

Adversarial review for Claude Code coding sessions.

After Claude finishes a coding task, `debate` forks the session for one
or more critic agents (Codex by default), runs a multi-round
cross-examination per critic, applies any concessions the proposer
makes, and surfaces only the unresolved disputes for human attention.
No debate content ever lands in the root Claude session — debate
happens in branched forks off the root.

**Status: v0 implementation complete.** Build green, full test suite
(unit + e2e + hook) passes locally and in CI. v0 GA is gated on the
upstream `agents-byzantine-tolerance` per-aspect critic-found-bug
rates and the no-output Stop-hook probe — see
[specs/27-release.md](specs/27-release.md) for the full G1–G18 gate
checklist. The design lives in
[specs/01-overview.md](specs/01-overview.md); per-component
implementation contracts are under [specs/](specs/).

## Install

```sh
go install latere.ai/x/debate/cmd/debate@latest
debate install-hook --scope user
```

Or grab a pre-built tarball once a release is tagged:

```sh
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64
curl -L "https://github.com/latere-ai/debate/releases/download/v0.0.1/debate_v0.0.1_${OS}_${ARCH}.tar.gz" | tar xz
./debate install-hook --scope user
```

From source:

```sh
git clone https://github.com/latere-ai/debate.git
cd debate
make build
./bin/debate install-hook --scope user
```

`debate install-hook` merges the verbose-format Stop hook entry into
`~/.claude/settings.json` (or `./.claude/settings.json` with `--scope
project`). `debate uninstall-hook` removes it.

## Usage

### Auto-trigger (default)

With the Stop hook installed, `debate` runs automatically when claude
finishes responding. Control returns to you once the run completes.

- Summary lives at `.debate/sessions/<id>/summary.md`.
- Cross-session log at `.debate/log.jsonl` records every run (one line
  per run; last line is the most recent).
- Stdout rendering during the Stop hook is best-effort — the contract
  is "summary on disk; stdout may also show it." See
  [§Trigger via Stop hook](specs/01-overview.md#trigger-via-stop-hook-default-for-claude-as-proposer)
  in spec 01.

Trivial diffs (under `--changed-lines-min`, default 10) short-circuit
in milliseconds and only append a `kind:"skipped"` entry to
`log.jsonl`; no session folder is created.

### Manual invocation

For CI gating, scripted batch runs, or out-of-band review:

```sh
debate \
  --session-id <root-claude-session-id> \
  --side-count 4 \
  --aspect functional-logic,security,code-quality,performance \
  --max-turn 6
```

`debate --help` lists every flag; `debate --version` prints the
build's version, commit, and date.

Exit codes: 0 = clean run / `--hook-mode` / `--changed-lines-min`
short-circuit / recursion guard; 1 = at least one unresolved leaf;
130 = interrupted; 100s = pre-flight failure (see
[specs/06-preflight.md](specs/06-preflight.md) for the full table).

### What you see during a run

```text
$ debate --session-id … --max-turn 6
[debate] 2 unresolved; see /your/repo/.debate/sessions/20260506T140905Z-q3a9f1/summary.md
$ cat .debate/sessions/*/summary.md
# Debate review — terminated: steady-state

## Headline (most contested unresolved)
- [security/api.go:88] SQL injection via unparameterized LIKE
  - Critic: framework auto-escape doesn't cover LIKE patterns
  - Proposer: parameterized via SQLAlchemy
  - **Stake**: GET /search?q=%' OR 1=1--
  - Contention: 3 (re-attacked: true)
…
```

## Design

The full design is in [specs/01-overview.md](specs/01-overview.md).
Key bets:

- **Forked debate, no debate content in root.** Each critic gets its
  own fork (`claude --fork-session`). No debate turn, no debate text,
  no proposer-clone reply ever lands in the root session's transcript;
  the user resumes from where they left off when debate ends. Fork
  mechanism verified against claude 2.1.129; the Stop-hook path may
  add a single hook-status attachment per run (probe owed before v0
  GA).
- **Verbatim channel.** Critic output reaches the proposer-clone as a
  plain user turn pointing to a file: `Some comments at @<path>.
  Please resolve or respond. If you disagree, please raise it.` No
  skill, slash-command, or plugin-template wrapping.
- **Persisted sessions.** Each run creates `.debate/sessions/<id>/`
  with per-fork round files, an attacks ledger, and a summary. Clean
  runs are silent on stdout; unresolved runs surface a
  contention-scored headline.
- **Aspect-specialized critics.** Default critics split coverage
  across `functional-logic`, `security`, `code-quality`,
  `performance`. The debate-theoretic property (one competent honest
  player suffices for soundness) means a lazy critic on one aspect
  doesn't break the others.
- **Best-effort critic isolation.** v0 enforces "diff + task only" by
  aspect prompt and `codex --sandbox read-only`, not by OS isolation.
  Strict per-fork sandbox dirs are deferred to v1.
- **Gated on per-aspect upstream research.**
  [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance)
  spec 07a measures critic-found-bug rate per aspect; aspects below
  threshold get dropped from defaults rather than the tool being
  abandoned. Ships when at least two aspects pass.

## Configuration

Optional `.debate.toml` at the project root or
`$XDG_CONFIG_HOME/debate/config.toml`:

```toml
max_turn = 6
side_count = 4
aspects = ["functional-logic", "security", "code-quality", "performance"]
cost_cap_tokens = 50000
changed_lines_min = 10
trigger = "stop"             # "stop" (default) | "manual"
allow_style_attacks = false
```

CLI flags > env vars (`DEBATE_*`) > project config > user config >
defaults. See [specs/05-config-file.md](specs/05-config-file.md).

## Develop

```sh
make all       # lint (golangci-lint v2) + vet + test (race) + build
make e2e       # full e2e suite: CLI flow + Stop hook script
make coverage  # writes coverage.html, prints total
make probe     # run pre-GA probes against a real claude/codex install
```

CI runs `make all` on every push and PR (Linux + macOS). Tagging `v*`
runs the same gate plus `goreleaser` to publish the release. Release
notes are auto-generated from commit messages, grouped by Go-style
prefix (`cli:`, `state:`, `agent:`, …).

## Related

- [agents-byzantine-tolerance](https://github.com/changkun/agents-byzantine-tolerance)
  — research repo studying multi-agent Byzantine fault tolerance,
  including
  [spec 07](https://github.com/changkun/agents-byzantine-tolerance/blob/main/specs/07-adversarial-debate.md),
  the architecture this tool productizes.
