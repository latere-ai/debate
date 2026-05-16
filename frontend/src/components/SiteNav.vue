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
      <a href="#top" class="nav-brand" aria-label="Agon" @click="close">
        <svg class="nav-glyph" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M4 6 L11 12 L4 18 Z" />
          <path d="M20 6 L13 12 L20 18 Z" />
          <path d="M12 3.5 V20.5" opacity="0.7" />
        </svg>
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
        <a class="btn btn-primary nav-install" href="#install" @click="close">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
          </svg>
          {{ nav.install }}
        </a>
      </nav>
    </div>
  </header>
</template>
