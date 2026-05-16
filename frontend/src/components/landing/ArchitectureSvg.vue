<script setup lang="ts">
// Static architecture diagram. Labels stay in English — they read as
// code/typographic plates inside the diagram, not flowing prose. The
// hardcoded ink/paper colours are repainted for dark mode via the
// attribute selectors in dialectic.css.
const ledger = [60, 82, 104, 126, 148, 170];
</script>

<template>
  <svg class="arch-svg" viewBox="0 0 800 360" xmlns="http://www.w3.org/2000/svg">
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
    <g transform="translate(48,164)">
      <rect width="130" height="48" rx="8" fill="none" stroke="#998a82" stroke-width="1.4" stroke-dasharray="4 4" />
      <text x="65" y="22" text-anchor="middle" font-family="ui-monospace, monospace" font-size="11" fill="#998a82" letter-spacing="0.14em">TASK</text>
      <text x="65" y="38" text-anchor="middle" font-family="Inter, sans-serif" font-size="11" fill="#564843">diff · plan · claim</text>
    </g>

    <!-- Proposer — outlined ink -->
    <g transform="translate(240,72)">
      <rect width="170" height="72" rx="10" fill="#fbf8f3" stroke="#1c130f" stroke-width="1.6" />
      <text x="85" y="28" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#1c130f" letter-spacing="0.16em">PROPOSER</text>
      <text x="85" y="52" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="22" fill="#1c130f">agent α</text>
    </g>
    <!-- Critic — filled ink -->
    <g transform="translate(240,216)">
      <rect width="170" height="72" rx="10" fill="#1c130f" stroke="#1c130f" stroke-width="1.6" />
      <text x="85" y="28" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#c25a5b" letter-spacing="0.16em">CRITIC</text>
      <text x="85" y="52" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="22" fill="#fbf8f3">agent β</text>
    </g>
    <!-- Debate -->
    <path d="M 330 144 C 330 165, 330 195, 330 216" fill="none" stroke="#1c130f" stroke-width="1.6" marker-end="url(#d-arr-i)" />
    <path d="M 308 216 C 308 195, 308 165, 308 144" fill="none" stroke="#1c130f" stroke-width="1.6" marker-end="url(#d-arr-i)" />
    <text x="362" y="186" font-family="ui-monospace, monospace" font-size="9" fill="#998a82" letter-spacing="0.12em">R1·R3</text>
    <text x="262" y="186" font-family="ui-monospace, monospace" font-size="9" fill="#998a82" letter-spacing="0.12em" text-anchor="end">R2·R4</text>

    <!-- Task → proposer/critic -->
    <path d="M 180 178 L 238 122" fill="none" stroke="#998a82" stroke-width="1.4" stroke-dasharray="4 4" marker-end="url(#d-arr-m)" />
    <path d="M 180 198 L 238 240" fill="none" stroke="#998a82" stroke-width="1.4" stroke-dasharray="4 4" marker-end="url(#d-arr-m)" />

    <!-- Ledger -->
    <g transform="translate(448,72)">
      <rect width="160" height="216" rx="6" fill="#fbf8f3" stroke="#1c130f" stroke-width="1.4" />
      <text x="80" y="26" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#1c130f" letter-spacing="0.18em">LEDGER · append-only</text>
      <line x1="16" y1="42" x2="144" y2="42" stroke="#d9cbb7" />
      <g v-for="(y, i) in ledger" :key="i">
        <text x="20" :y="y" font-family="ui-monospace, monospace" font-size="9" fill="#998a82">R{{ Math.floor(i / 2) + 1 }}·0{{ i + 1 }}</text>
        <rect x="62" :y="y - 7" width="36" height="3" :fill="i === 5 ? '#8a2a2b' : '#564843'" :opacity="i === 5 ? 1 : 0.5" />
        <text x="106" :y="y" font-family="ui-monospace, monospace" font-size="8" :fill="i === 5 ? '#8a2a2b' : '#998a82'" :font-weight="i === 5 ? 700 : 400">{{ i === 5 ? 'STAKE' : 'rsv' }}</text>
      </g>
      <line x1="16" y1="184" x2="144" y2="184" stroke="#d9cbb7" />
      <text x="80" y="202" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="13" fill="#8a2a2b">★ ATK-1 staked</text>
    </g>
    <path d="M 410 108 L 446 108" fill="none" stroke="#998a82" stroke-width="1.4" marker-end="url(#d-arr-m)" />
    <path d="M 410 250 L 446 250" fill="none" stroke="#998a82" stroke-width="1.4" marker-end="url(#d-arr-m)" />

    <!-- Judge -->
    <g transform="translate(642,150)">
      <rect width="110" height="56" rx="8" fill="#8a2a2b" stroke="#8a2a2b" stroke-width="1.6" />
      <text x="55" y="24" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#fbf8f3" letter-spacing="0.16em">JUDGE</text>
      <text x="55" y="42" text-anchor="middle" font-family="Inter, sans-serif" font-size="10" fill="rgba(251,248,243,0.75)">staked leaf only</text>
    </g>
    <path d="M 610 178 L 640 178" fill="none" stroke="#8a2a2b" stroke-width="1.6" marker-end="url(#d-arr-a)" />
    <text x="625" y="172" text-anchor="middle" font-family="ui-monospace, monospace" font-size="9" fill="#8a2a2b" letter-spacing="0.12em" font-weight="700">★</text>

    <!-- Human -->
    <g transform="translate(642,238)">
      <rect width="110" height="56" rx="8" fill="none" stroke="#564843" stroke-width="1.4" />
      <text x="55" y="24" text-anchor="middle" font-family="ui-monospace, monospace" font-size="10" fill="#564843" letter-spacing="0.16em">HUMAN</text>
      <text x="55" y="42" text-anchor="middle" font-family="Instrument Serif, serif" font-style="italic" font-size="13" fill="#1c130f">decides</text>
    </g>
    <path d="M 696 206 L 696 236" fill="none" stroke="#564843" stroke-width="1.4" stroke-dasharray="4 4" marker-end="url(#d-arr-i)" />
  </svg>
</template>
