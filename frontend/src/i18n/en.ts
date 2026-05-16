export const en: Record<string, string> = {
  'nav.install': 'Install',

  'hero.eyebrow': 'Adversarial verification gate',
  'hero.title': '<span class="agon-brand">Agon</span>:<br>make AI-produced work defend itself',
  'hero.sub':
    'An independent critic cross-examines what an AI produced. The producer must defend or concede every concrete attack. Only the disputes that survive reach a human.',
  'hero.cta.install': 'Install',
  'hero.cta.repo': 'Read the protocol',

  'what.label': 'What it is',
  'what.title': 'Not a judge. A <span class="agon-brand">protocol</span>.',
  'what.p1':
    '<strong>Agon</strong> sits between an AI agent and whoever, or whatever, consumes its output. It runs an honest, competent critic against the work, forces the producing agent to defend or fix every concrete attack, and surfaces only what stays contested.',
  'what.p2':
    'The artifact need not be code. A diff, a research write-up, an outcome analysis, a plan, a high-stakes decision: the same protocol applies. What resolves becomes a pass signal another agent can consume directly; what stays contested is what a person reviews.',

  'proto.label': 'The protocol',
  'proto.title': 'Propose, attack, defend, until only the unresolved remains',
  'proto.lead':
    'A proposer and one or more independent critics, each pressing a distinct aspect. Cross-examination runs until the dispute reaches steady state or a round cap. Only what stays unresolved surfaces, ranked by a pure contention score.',
  'proto.s1.t': 'Proposal',
  'proto.s1.d': 'The proposer answers the task: a claim, a diff, an argument.',
  'proto.s2.t': 'Attack',
  'proto.s2.d': 'Each critic picks its own aspect (functional logic, security, code quality, performance) and produces concrete attacks: a specific input X yielding output Y that violates Z. Not vibes.',
  'proto.s3.t': 'Defense',
  'proto.s3.d': 'The proposer responds to each attack: concede, and the proposer-clone applies the fix, or rebut with a specific counter-claim.',
  'proto.s4.t': 'Surface',
  'proto.s4.d': 'Every attack carries a stable id in an append-only ledger. Only unresolved disputes surface, ranked by how many rounds each survived. No model scores the outcome.',
  'proto.judge':
    'The surfacing layer is a pure rule: the contention score is rounds survived plus a bit for whether an attack was re-raised. There is no model in that loop, on purpose. The bounded-judge result from the adversarial-verification literature is the intuition behind this, not the runtime; its honest limits are set out under Academic foundations.',

  'why.label': 'Why it is different',
  'why.title': 'Four properties the alternatives do not replicate',
  'why.c1.t': 'One honest player suffices',
  'why.c1.d':
    'A dishonest proposer must hold a consistent lie across every cross-examination round; an honest critic needs one inconsistency. Because each critic owns a distinct aspect, a weak critic on one aspect does not break coverage on the others; weak aspects get dropped from defaults, not the tool.',
  'why.c2.t': 'Vendor-neutral by construction',
  'why.c2.d':
    'The default pairing is cross-family: Claude proposes, Codex critiques. Same model on both sides is the model reviewing itself, and is rejected. No model vendor will ship the neutral layer; the incentive is to sell more of its own tokens.',
  'why.c3.t': 'Channel purity',
  'why.c3.d':
    'Critic output reaches the proposer as a verbatim user message, not a skill or template. The proposer defends the way it would against a person pasting a review. Wrapping it distorts the defense.',
  'why.c4.t': 'Auditable by design',
  'why.c4.d':
    'Stable attack ids, an append-only ledger, contention-scored headlines by a pure rule with no model judging at the surfacing layer. A security team reads a session like a court transcript.',

  'signal.label': 'Resolved vs. contested',
  'signal.title': 'A machine-readable signal, not just a human headline',
  'signal.p1':
    'Inside an agent loop the contention score is a decision gate: attacks the proposer resolves are a <strong>pass</strong> the calling agent proceeds on; the contested tail above the threshold is what escalates to a human.',
  'signal.p2':
    'This is the high-stakes-decision use: agents resolve, proceed; not resolved, a person reviews. It is not a new capability; it is what the existing ledger and contention score already are when the consumer is another agent rather than a person.',

  'found.label': 'Academic foundations',
  'found.title': 'Grounded in the adversarial-verification literature, and honest about the limits',
  'found.p1':
    'Agon productizes the adversarial-verification architecture of Irving, Christiano &amp; Amodei (2018). Its complexity-theoretic intuition, that the adversarial protocol reaches PSPACE under optimal play, strictly above NP, is <em>suggestive</em>, not a claim about LLMs: LLMs are not optimal players.',
  'found.p2':
    'The closer theoretical fit is Brown-Cohen, Irving &amp; Piliouras (2023), which extends the result to stochastic systems and to honest players with polynomial simulation budgets, both required for it to apply to LLMs at all. Their 2025 Prover-Estimator protocol addresses the obfuscated-arguments attack on the plain protocol.',
  'found.research.cite': 'Research home.',
  'found.research.note':
    'the open research suite this productizes, probing adversarial verification across compute, depth, stochasticity, leaf format, obfuscation, and query-complexity scaling.',
  'found.honest':
    'Honest framing: the formal soundness results are about the protocol under stated assumptions, not a guarantee about any particular model. Application to real LLMs is empirically motivated and hypothesis-stage. The gating metric is the per-aspect critic-found-bug rate; if a critic does not actually attack, the protocol collapses to the proposer alone. Agon does not prove your code correct, and does not claim to remove the need for trust.',

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
  'footer.theme': 'Theme',
  'footer.language': 'Language',
  'footer.rights': 'Agon, a Latere product. Adversarial review for any AI-produced artifact.',

  'nf.text': 'That page does not exist.',
  'nf.home': 'Back to home',
};
