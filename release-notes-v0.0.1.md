# debate v0.0.1

First v0 release. Productizes the adversarial-debate architecture from
spec 07 of agents-byzantine-tolerance.

## Probe outcomes

To be filled before tagging the GA release. Run `make probe` against a
real claude/codex install and record outcomes here:

- no-output-stop-hook: <PASS|FAIL|N/A>
- signal-latency: <PASS|FAIL>
- trivial-diff-perf: <PASS|FAIL>
- interactive-stdout: <PASS|FAIL|SKIP>

## Default aspects (this release)

Aspects whose 07a per-aspect critic-found-bug rate is ≥ 60%:

- functional-logic
- security
- code-quality
- performance

(Update before GA based on the actual 07a measurement.)

## Known limitations

- Codex-as-proposer is v1 (architecture documented in spec 01).
- No live in-session UI; summary on disk; stdout best-effort under the
  Stop hook.
- Critics are best-effort isolated (aspect prompt + codex --sandbox
  read-only); strict per-fork sandbox is v1.
- `--judge llm/human` not in v0.

## Install

```
curl -L https://github.com/latere-ai/debate/releases/download/v0.0.1/debate_v0.0.1_<os>_<arch>.tar.gz | tar xz
./debate install-hook --scope user
```

## Verify

```
debate --version
debate --help
```
