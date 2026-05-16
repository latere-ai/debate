<script setup lang="ts">
import { ref, computed } from 'vue';
import { useContent } from '../content';

const content = useContent();
const nav = computed(() => content.value.nav);
const open = ref(false);

function close() {
  open.value = false;
}
</script>

<template>
  <header class="nav">
    <div class="nav-inner">
      <a href="#top" class="nav-brand" @click="close">
        <span class="agon-mark">Agon</span>
        <small>{{ nav.by }}</small>
      </a>
      <button
        class="nav-toggle"
        type="button"
        :aria-expanded="open"
        aria-label="Toggle navigation"
        @click="open = !open">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path v-if="!open" d="M3 6h18M3 12h18M3 18h18" />
          <path v-else d="M6 6l12 12M6 18L18 6" />
        </svg>
      </button>
      <nav class="nav-links" :class="{ 'is-open': open }">
        <a
          v-for="link in nav.links"
          :key="link.href"
          class="nav-link"
          :href="link.href"
          @click="close">{{ link.label }}</a>
        <a
          class="btn btn-ghost"
          href="https://github.com/latere-ai/agon"
          target="_blank"
          rel="noopener"
          style="margin-left: 8px"
          @click="close">{{ nav.github }}</a>
        <a class="btn btn-primary" href="#install" @click="close">{{ nav.install }}</a>
      </nav>
    </div>
  </header>
</template>
