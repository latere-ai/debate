<script setup lang="ts">
import { useT } from '../i18n';
const t = useT();
const year = new Date().getFullYear();
</script>

<template>
  <footer class="site-footer">
    <div class="footer-row">
      <div class="footer-links">
        <a href="https://github.com/latere-ai/debate" target="_blank" rel="noopener">GitHub</a>
        <a href="https://github.com/latere-ai/debate#readme" target="_blank" rel="noopener">{{ t('footer.docs') }}</a>
        <a href="https://github.com/changkun/agents-byzantine-tolerance" target="_blank" rel="noopener">{{ t('footer.research') }}</a>
        <a href="https://latere.ai/" target="_blank" rel="noopener">Latere</a>
      </div>
      <div class="footer-meta">
        <span class="agon-brand">Agon</span> · {{ t('footer.tagline') }} · © {{ year }} Latere
      </div>
    </div>
  </footer>
</template>
