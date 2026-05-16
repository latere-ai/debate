<script setup lang="ts">
defineProps<{
  id: string;
  eyebrow: string;
  /** HTML — may contain <em> accents */
  title: string;
  lead?: string;
}>();
</script>

<template>
  <section :id="id" class="section dark">
    <div class="section-grid">
      <aside class="sg-side reveal">
        <span class="eyebrow">{{ eyebrow }}</span>
        <h2 class="section-title" v-html="title"></h2>
        <p v-if="lead" class="section-lead">{{ lead }}</p>
      </aside>
      <div class="reveal">
        <slot />
      </div>
    </div>
  </section>
</template>
