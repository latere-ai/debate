<script setup lang="ts">
import { useHead } from '@unhead/vue';
import DefaultLayout from '../layouts/DefaultLayout.vue';
import { useT } from '../i18n';
const t = useT();
useHead({ title: '404 — Agon' });
</script>

<template>
  <DefaultLayout>
    <div class="nf">
      <h1 class="agon-brand">404</h1>
      <p class="prose">{{ t('nf.text') }}</p>
      <p><a class="btn btn-ghost" href="/">{{ t('nf.home') }}</a></p>
    </div>
  </DefaultLayout>
</template>
