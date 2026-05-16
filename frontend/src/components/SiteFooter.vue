<script setup lang="ts">
import { computed } from 'vue';
import { storeToRefs } from 'pinia';
import { usePrefsStore, type Theme, type Locale } from '../stores/prefs';
import { useContent } from '../content';

const content = useContent();
const footer = computed(() => content.value.footer);
const metaHtml = computed(() =>
  footer.value.meta.replace('{year}', String(new Date().getFullYear())),
);

const prefs = usePrefsStore();
const { theme, locale } = storeToRefs(prefs);

const themes: { v: Theme; label: string }[] = [
  { v: 'light', label: '☀' },
  { v: 'dark', label: '☾' },
  { v: 'auto', label: '◐' },
];
const locales: { v: Locale; label: string }[] = [
  { v: 'en', label: 'EN' },
  { v: 'zh', label: '中' },
];

function isExternal(href: string) {
  return href.startsWith('http');
}
</script>

<template>
  <footer class="agon-footer">
    <div class="container">
      <div class="footer-grid">
        <div>
          <div class="footer-brand-row">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M4 16 C 4 8, 11 4, 14 8 C 17 12, 10 16, 10 16 C 10 16, 17 20, 20 16 C 23 12, 16 8, 16 8" />
            </svg>
            <span class="footer-brand-text">Agon</span>
          </div>
          <p class="footer-tagline">{{ footer.tagline }}</p>
        </div>

        <div v-for="col in footer.columns" :key="col.head">
          <div class="footer-col-head">{{ col.head }}</div>
          <div class="footer-col">
            <a
              v-for="link in col.links"
              :key="link.href"
              :href="link.href"
              :target="isExternal(link.href) ? '_blank' : undefined"
              :rel="isExternal(link.href) ? 'noopener' : undefined">{{ link.label }}</a>
          </div>
        </div>
      </div>

      <div class="footer-base">
        <span class="footer-meta" v-html="metaHtml"></span>
        <div class="footer-base-right">
          <div class="footer-prefs">
            <div class="footer-seg" role="group" aria-label="Theme">
              <button
                v-for="opt in themes"
                :key="opt.v"
                type="button"
                class="footer-seg-btn"
                :class="{ 'is-active': theme === opt.v }"
                @click="prefs.setTheme(opt.v)">{{ opt.label }}</button>
            </div>
            <div class="footer-seg" role="group" aria-label="Language">
              <button
                v-for="opt in locales"
                :key="opt.v"
                type="button"
                class="footer-seg-btn"
                :class="{ 'is-active': locale === opt.v }"
                @click="prefs.setLocale(opt.v)">{{ opt.label }}</button>
            </div>
          </div>
          <span class="footer-thesis">{{ footer.thesis }}</span>
        </div>
      </div>
    </div>
  </footer>
</template>
