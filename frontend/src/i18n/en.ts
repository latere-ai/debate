export const en: Record<string, string> = {
  'nav.theme': 'Toggle theme',
  'nav.install': 'Install',

  'hero.eyebrow': 'Adversarial verification gate',
  'hero.title': 'Make AI-produced work<br>defend itself — <span class="agon-brand">Agon</span>',
  'hero.sub':
    'An independent critic cross-examines what an AI produced. The producer must defend or concede every concrete attack. Only the disputes that survive reach a human.',
  'hero.cta.install': 'Install',
  'hero.cta.repo': 'Read the protocol',

  'what.label': 'What it is',
  'what.title': 'Not a judge. A <span class="agon-brand">protocol</span>.',
  'what.p1':
    '<strong>Agon</strong> sits between an AI agent and whoever — or whatever — consumes its output. It runs an honest, competent critic against the work, forces the producing agent to defend or fix every concrete attack, and surfaces only what stays contested.',
  'what.p2':
    'The artifact need not be code. A diff, a research write-up, an outcome analysis, a plan, a high-stakes decision — the same protocol applies. Agon is the internal codename <code>debate</code>; the binary, repository, and Stop-hook contract ship under that name.',

  'proto.label': 'The protocol',
  'proto.title': 'Four rounds, an externally-judged stake',
  'proto.lead':
    'A proposer and a critic, with a judge that inspects only the single leaf the debate ends on — never the full transcript.',
  'proto.s1.t': 'Proposal',
  'proto.s1.d': 'The proposer answers the task — a claim, a diff, an argument.',
  'proto.s2.t': 'Attack',
  'proto.s2.d': 'The critic produces concrete attacks: a specific input X yielding output Y that violates Z. Not vibes.',
  'proto.s3.t': 'Defense',
  'proto.s3.d': 'The proposer responds to each attack: concede, or rebut with a specific counter-claim.',
  'proto.s4.t': 'Stake',
  'proto.s4.d': 'The critic stakes one unresolved attack as the decisive leaf. Concede everything and the proposer wins by default.',
  'proto.judge':
    'The judge evaluates only the staked leaf. Soundness no longer needs an honest majority — it needs one honest player and a calibrated judge. This lowers the trust budget; it does not eliminate trust.',

  'why.label': 'Why it is different',
  'why.title': 'Four properties the alternatives do not replicate',
  'why.c1.t': 'One honest player suffices',
  'why.c1.d':
    'A Byzantine proposer must hold a consistent lie across every cross-examination round; an honest critic needs to find one inconsistency. The failure mode becomes per-aspect, not whole-tool.',
  'why.c2.t': 'Vendor-neutral by construction',
  'why.c2.d':
    'The default pairing is cross-family — Claude proposes, Codex critiques. Same-model-both-sides is the model debating itself and is rejected. No model vendor will ship the neutral layer.',
  'why.c3.t': 'Channel purity',
  'why.c3.d':
    'Critic output reaches the proposer as a verbatim user message, not a skill or template. The proposer defends the way it would against a human pasting a review. Wrapping it distorts the defense.',
  'why.c4.t': 'Auditable by design',
  'why.c4.d':
    'Stable attack ids, an append-only ledger, contention-scored headlines by a pure rule — no LLM judging at the surfacing layer. A security team reads a session like a court transcript.',

  'signal.label': 'Resolved vs. contested',
  'signal.title': 'A machine-readable signal, not just a human headline',
  'signal.p1':
    'Inside an agent loop the contention score is a decision gate: attacks the proposer resolves are a <strong>pass</strong> the calling agent proceeds on; the contested tail above the threshold is what escalates to a human.',
  'signal.p2':
    'This is the high-stakes-decision use — agents resolve ⇒ proceed; not resolved ⇒ human review. It is not a new capability; it is what the existing ledger and contention score already are when the consumer is another agent rather than a person.',

  'found.label': 'Academic foundations',
  'found.title': 'Grounded in the debate literature — and honest about the limits',
  'found.p1':
    'Agon productizes the adversarial-debate architecture of Irving, Christiano &amp; Amodei (2018). The complexity-theoretic intuition — debate ≈ PSPACE under optimal play, strictly above NP — is <em>suggestive</em>, not a claim about LLMs: LLMs are not optimal players.',
  'found.p2':
    'The closer theoretical fit is Brown-Cohen, Irving &amp; Piliouras (2023), which extends the result to stochastic systems and to honest debaters with polynomial simulation budgets — both required for it to apply to LLMs at all. Their 2025 Prover-Estimator protocol addresses the obfuscated-arguments attack on plain debate.',
  'found.research.cite': 'Research home.',
  'found.research.note':
    'the experiment suite this productizes (spec 07, Adversarial Debate; specs 08–13 extend it along compute, depth, stochasticity, leaf format, obfuscation, and query-complexity scaling).',
  'found.honest':
    'Honest framing: the formal soundness results are about the protocol under stated assumptions, not a guarantee about any particular model. Application to real LLMs is empirically motivated and hypothesis-stage — the gating metric is the per-aspect critic-found-bug rate; if a critic does not actually attack, debate collapses to the proposer alone. Agon does not prove your code correct, and does not claim to remove the need for trust.',

  'install.label': 'Install',
  'install.title': 'One binary, an optional Stop hook',
  'install.lead':
    'Vendor-neutral and local-first. Bring your own Claude / Codex; Agon runs the protocol and writes an auditable session to disk.',
  'install.c1': 'one-liner (detects OS/arch, verifies checksum, installs the Stop hook)',
  'install.c2': 'from source (Go 1.26+)',
  'install.repo': 'View on GitHub',

  'footer.docs': 'Docs',
  'footer.research': 'Research',
  'footer.tagline': 'adversarial verification for AI-produced work',

  'nf.text': 'That page does not exist.',
  'nf.home': 'Back to home',
};
