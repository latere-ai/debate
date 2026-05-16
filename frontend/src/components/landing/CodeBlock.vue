<script setup lang="ts">
import { ref, computed } from 'vue';
import { useContent } from '../../content';

const props = defineProps<{ comment: string; command: string }>();

const content = useContent();
const labels = computed(() => content.value.install);
const copied = ref(false);
const lines = computed(() => props.command.split('\n'));

function copy() {
  if (typeof navigator !== 'undefined' && navigator.clipboard) {
    navigator.clipboard.writeText(props.command);
  }
  copied.value = true;
  setTimeout(() => {
    copied.value = false;
  }, 1400);
}
</script>

<template>
  <div class="code">
    <div class="code-content">
      <span class="c-c"># {{ comment }}</span>
      <div v-for="(line, i) in lines" :key="i"><span class="c-tok">$</span> {{ line }}</div>
    </div>
    <button class="copy-btn" :class="{ copied }" type="button" @click="copy">
      {{ copied ? labels.copied : labels.copy }}
    </button>
  </div>
</template>
