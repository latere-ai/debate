<script setup lang="ts">
// Static architecture diagram. Labels stay in English — they read as
// code/typographic plates inside the diagram, not flowing prose. The
// hardcoded ink/paper colours are repainted for dark mode via the
// attribute selectors in dialectic.css, so keep these literals exact:
//   #fbf8f3 #1c130f #998a82 #564843 #8a2a2b #c25a5b #d9cbb7
//   rgba(251,248,243,0.75)
// viewBox 0 0 880 360 with ≥40px clear margin on every side; no label
// sits on an arrow and no text overflows its plate.
const ledger = [
  { y: 60, id: 'R1·01', stake: false },
  { y: 86, id: 'R2·02', stake: false },
  { y: 112, id: 'R2·03', stake: false },
  { y: 138, id: 'R3·04', stake: false },
  { y: 164, id: 'R3·05', stake: false },
  { y: 190, id: 'R4·06', stake: true },
];
</script>

<template>
  <svg class="arch-svg" viewBox="0 0 880 360" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <marker id="d-arr-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
        <path d="M0,0 L10,5 L0,10 z" fill="#8a2a2b" />
      </marker>
      <marker id="d-arr-i" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
        <path d="M0,0 L10,5 L0,10 z" fill="#1c130f" />
      </marker>
      <marker id="d-arr-m" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
        <path d="M0,0 L10,5 L0,10 z" fill="#998a82" />
      </marker>
    </defs>

    <!-- Task -->
    <g transform="translate(40,154)">
      <rect width="120" height="52" rx="8" fill="none" stroke="#998a82" stroke-width="1.4" stroke-dasharray="4 4" />
      <text x="60" y="22" text-anchor="middle" font-family="ui-monospace, monospace" font-size="11" fill="#998a82" letter-spacing="0.12em">TASK</text>
      <text x="60" y="39" text-anchor="middle" font-family="Inter, sans-serif" font-size="9.5" fill="#564843">diff · plan · claim</text>
    </g>

    <!-- Task → proposer / critic -->
    <path d="M 160 172 L 206 112" fill="none" stroke="#998a82" stroke-width="1.4" stroke-dasharray="4 4" marker-end="url(#d-arr-m)" />
    <path d="M 160 188 L 206 248" fill="none" stroke="#998a82" stroke-width="1.4" stroke-dasharray="4 4" marker-end="url(#d-arr-m)" />

    <!-- Proposer — outlined ink -->
    <g transform="translate(212,66)">
      <rect width="176" height="72" rx="10" fill="#fbf8f3" stroke="#1c130f" stroke-width="1.6" />
      <text x="88" y="28" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#1c130f" letter-spacing="0.14em">PROPOSER</text>
      <text x="88" y="53" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="21" fill="#1c130f">agent α</text>
    </g>
    <!-- Critic — filled ink -->
    <g transform="translate(212,222)">
      <rect width="176" height="72" rx="10" fill="#1c130f" stroke="#1c130f" stroke-width="1.6" />
      <text x="88" y="28" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#c25a5b" letter-spacing="0.14em">CRITIC</text>
      <text x="88" y="53" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="21" fill="#fbf8f3">agent β</text>
    </g>

    <!-- Debate (proposer ⇄ critic), labels offset clear of the arrows -->
    <path d="M 288 140 C 288 165, 288 197, 288 220" fill="none" stroke="#1c130f" stroke-width="1.6" marker-end="url(#d-arr-i)" />
    <path d="M 312 220 C 312 197, 312 165, 312 140" fill="none" stroke="#1c130f" stroke-width="1.6" marker-end="url(#d-arr-i)" />
    <text x="338" y="172" font-family="ui-monospace, monospace" font-size="9" fill="#998a82" letter-spacing="0.1em">R1·R3</text>
    <text x="338" y="196" font-family="ui-monospace, monospace" font-size="9" fill="#998a82" letter-spacing="0.1em">R2·R4</text>

    <!-- Proposer / critic → ledger -->
    <path d="M 388 112 L 448 124" fill="none" stroke="#998a82" stroke-width="1.4" marker-end="url(#d-arr-m)" />
    <path d="M 388 248 L 448 236" fill="none" stroke="#998a82" stroke-width="1.4" marker-end="url(#d-arr-m)" />

    <!-- Ledger -->
    <g transform="translate(452,60)">
      <rect width="200" height="240" rx="6" fill="#fbf8f3" stroke="#1c130f" stroke-width="1.4" />
      <text x="100" y="26" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#1c130f" letter-spacing="0.08em">LEDGER · append-only</text>
      <line x1="18" y1="40" x2="182" y2="40" stroke="#d9cbb7" />
      <g v-for="(r, i) in ledger" :key="i">
        <text x="20" :y="r.y" font-family="ui-monospace, monospace" font-size="9" fill="#998a82">{{ r.id }}</text>
        <rect x="78" :y="r.y - 7" width="60" height="3" :fill="r.stake ? '#8a2a2b' : '#564843'" :opacity="r.stake ? 1 : 0.5" />
        <text x="150" :y="r.y" font-family="ui-monospace, monospace" font-size="8" :fill="r.stake ? '#8a2a2b' : '#998a82'" :font-weight="r.stake ? 700 : 400">{{ r.stake ? 'STAKE' : 'rsv' }}</text>
      </g>
      <line x1="18" y1="208" x2="182" y2="208" stroke="#d9cbb7" />
      <text x="100" y="228" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="13" fill="#8a2a2b">★ ATK-1 staked</text>
    </g>

    <!-- Ledger → judge -->
    <path d="M 652 150 L 716 150" fill="none" stroke="#8a2a2b" stroke-width="1.6" marker-end="url(#d-arr-a)" />
    <text x="684" y="140" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#8a2a2b" letter-spacing="0.1em" font-weight="700">★</text>

    <!-- Judge -->
    <g transform="translate(720,120)">
      <rect width="120" height="60" rx="8" fill="#8a2a2b" stroke="#8a2a2b" stroke-width="1.6" />
      <text x="60" y="26" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#fbf8f3" letter-spacing="0.14em">JUDGE</text>
      <text x="60" y="44" text-anchor="middle" font-family="Inter, sans-serif" font-size="9.5" fill="rgba(251,248,243,0.75)">staked leaf only</text>
    </g>

    <!-- Judge → human -->
    <path d="M 780 182 L 780 228" fill="none" stroke="#564843" stroke-width="1.4" stroke-dasharray="4 4" marker-end="url(#d-arr-i)" />

    <!-- Human -->
    <g transform="translate(720,232)">
      <rect width="120" height="60" rx="8" fill="none" stroke="#564843" stroke-width="1.4" />
      <text x="60" y="26" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#564843" letter-spacing="0.14em">HUMAN</text>
      <text x="60" y="44" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="13" fill="#1c130f">decides</text>
    </g>
  </svg>
</template>
