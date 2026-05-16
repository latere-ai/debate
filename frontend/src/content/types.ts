/* ============================================================
   Landing content schema (Dialectic design).

   Ported from the design bundle's content.js (CONTENT) and the
   STAGE_DATA constant in v2-dialectic.jsx. Fields whose copy
   contains inline markup are HTML strings rendered via v-html;
   the strict typing makes en/zh parity a compile-time error.
   ============================================================ */

export interface NavLink {
  label: string;
  href: string;
}

export interface StageRow {
  /** Round id, e.g. "R1" */
  r: string;
  /** Which column the bubble sits in */
  side: 'p' | 'c';
  /** Round label, e.g. "Proposal" */
  label: string;
  /** Bubble body (HTML) */
  html: string;
}

export interface TranscriptRow {
  n: string;
  actor: 'proposer' | 'critic' | 'judge';
  actorLabel: string;
  /** Message body (HTML) */
  html: string;
  tag?: { label: string; kind: 'attack' | 'resolved' | 'contested' | 'stake' };
  isStake?: boolean;
}

export type CompareCell = 'agon' | 'no' | 'partial';

export interface LandingContent {
  meta: { title: string; description: string };

  nav: {
    by: string;
    links: NavLink[];
    github: string;
    install: string;
  };

  hero: {
    stampProposer: string;
    stampCritic: string;
    /** HTML — includes <em> accents and the ★ */
    title: string;
    /** HTML */
    sub: string;
    ctaPrimary: string;
    ctaSecondary: string;
    worksWith: string;
    moreLabel: string;
  };

  stage: {
    /** HTML — case line */
    head: string;
    proposerCol: string;
    criticCol: string;
    proposerName: string;
    criticName: string;
    rows: StageRow[];
    verdictKey: string;
    verdictText: string;
    verdictRight: string;
  };

  transcript: {
    eyebrow: string;
    /** HTML (title with <em>) */
    title: string;
    lead: string;
    case: string;
    meta: string;
    rows: TranscriptRow[];
    footer: [string, string];
  };

  why: {
    eyebrow: string;
    title: string;
    pillars: { k: string; t: string; d: string }[];
  };

  compare: {
    eyebrow: string;
    title: string;
    /** [Property, Agon, Raw LLM, PR review] */
    headers: [string, string, string, string];
    rows: { p: string; cols: [CompareCell, CompareCell, CompareCell] }[];
  };

  usecases: {
    eyebrow: string;
    title: string;
    lead: string;
    items: { i: string; t: string; d: string }[];
  };

  arch: {
    eyebrow: string;
    title: string;
    lead: string;
    cap: string;
  };

  signal: {
    eyebrow: string;
    title: string;
    lead: string;
    cells: {
      tag: string;
      kind: 'r' | 'c';
      num: string;
      label: string;
      route: string;
      desc: string;
    }[];
  };

  found: {
    eyebrow: string;
    title: string;
    lead: string;
    refs: {
      yr: string;
      cite: string;
      em: string;
      tail: string;
      link: string;
      href: string;
    }[];
    pullquote: { q: string; cite: string };
    honestStrong: string;
    honestRest: string;
  };

  hook: {
    eyebrow: string;
    title: string;
    desc: string;
    cta: string;
    lines: { p: string; cmd?: string; attn?: string; l?: string }[];
  };

  faq: {
    eyebrow: string;
    title: string;
    items: { q: string; a: string }[];
  };

  install: {
    eyebrow: string;
    title: string;
    lead: string;
    a: { c: string; cmd: string };
    b: { c: string; cmd: string };
    copy: string;
    copied: string;
    ctaPrimary: string;
    ctaSecondary: string;
  };

  footer: {
    tagline: string;
    columns: { head: string; links: NavLink[] }[];
    /** HTML — {year} is substituted at render time */
    meta: string;
    thesis: string;
  };
}
