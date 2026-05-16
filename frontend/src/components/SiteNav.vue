<script setup lang="ts">
import { useT } from '../i18n';

const t = useT();
</script>

<template>
  <header class="site-header">
    <nav class="nav-container">
      <a href="/" class="logo-link" aria-label="Agon home">
        <span class="agon-brand">Agon</span>
        <span class="logo-sub">by Latere</span>
      </a>
      <div class="nav-actions">
        <a class="nav-link" href="https://github.com/latere-ai/debate" target="_blank" rel="noopener">GitHub</a>
        <a class="nav-link" href="https://latere.ai/" target="_blank" rel="noopener">Latere</a>
        <a class="nav-cta" href="#install">{{ t('nav.install') }}</a>
      </div>
    </nav>
  </header>
</template>
