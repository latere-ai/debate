<script setup lang="ts">
import { usePrefsStore } from '../stores/prefs';
import { useT } from '../i18n';

const prefs = usePrefsStore();
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
        <div class="seg" role="group" aria-label="Language">
          <button class="seg-btn" :class="{ 'is-active': prefs.locale === 'en' }" @click="prefs.setLocale('en')">EN</button>
          <button class="seg-btn" :class="{ 'is-active': prefs.locale === 'zh' }" @click="prefs.setLocale('zh')">中</button>
        </div>
        <button class="seg-btn" :aria-label="t('nav.theme')" style="border:1px solid var(--border);border-radius:8px;" @click="prefs.toggleTheme()">{{ prefs.themeIcon }}</button>
        <a class="nav-cta" href="#install">{{ t('nav.install') }}</a>
      </div>
    </nav>
  </header>
</template>
