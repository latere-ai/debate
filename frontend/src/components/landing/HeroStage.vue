<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useContent } from '../../content';

const content = useContent();
const stage = computed(() => content.value.stage);
const lastIndex = computed(() => stage.value.rows.length - 1);

// Start on the final state so SSR renders something sensible and a
// reduced-motion visit gets the resolved/staked frame.
const active = ref(stage.value.rows.length - 1);

let timer: ReturnType<typeof setInterval> | undefined;

onMounted(() => {
  const reduce =
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (reduce) return;
  active.value = 0;
  timer = setInterval(() => {
    active.value = (active.value + 1) % stage.value.rows.length;
  }, 2400);
});

onUnmounted(() => {
  if (timer) clearInterval(timer);
});

function bubbleStyle(i: number) {
  return {
    opacity: active.value >= i ? 1 : 0.3,
    transform: active.value === i ? 'scale(1.02)' : 'scale(1)',
  };
}

function spineClass(i: number, side: 'p' | 'c') {
  const cls = ['spine-stop'];
  if (active.value >= i) cls.push('is-' + side);
  if (i === lastIndex.value && active.value === i) cls.push('is-stake');
  return cls;
}
</script>

<template>
  <div class="stage" style="margin-top: 56px">
    <div class="stage-head">
      <div class="sh-title" v-html="stage.head"></div>
      <div class="sh-actors">
        <div class="a-p"><span class="dot"></span>{{ stage.proposerName }}</div>
        <div class="a-c"><span class="dot"></span>{{ stage.criticName }}</div>
      </div>
    </div>

    <div class="stage-columns">
      <div class="stage-col left">
        <div class="col-head">{{ stage.proposerCol }}</div>
        <div class="col-name">{{ stage.proposerName }}</div>
        <template v-for="(d, i) in stage.rows" :key="'p' + i">
          <div v-if="d.side === 'p'" class="msg m-p" :style="bubbleStyle(i)">
            <span class="msg-label">{{ d.r }} · {{ d.label }}</span>
            <span v-html="d.html"></span>
          </div>
          <div v-else style="min-height: 60px"></div>
        </template>
      </div>

      <div class="stage-spine">
        <div class="spine-line" style="left: 50%"></div>
        <div v-for="(d, i) in stage.rows" :key="'s' + i" :class="spineClass(i, d.side)">
          {{ d.r }}<small>{{ d.side === 'p' ? 'PROP' : 'CRIT' }}</small>
        </div>
      </div>

      <div class="stage-col right">
        <div class="col-head">{{ stage.criticCol }}</div>
        <div class="col-name">{{ stage.criticName }}</div>
        <template v-for="(d, i) in stage.rows" :key="'c' + i">
          <div v-if="d.side === 'c'" class="msg m-c" :style="bubbleStyle(i)">
            <span class="msg-label">{{ d.r }} · {{ d.label }}</span>
            <span v-html="d.html"></span>
          </div>
          <div v-else style="min-height: 60px"></div>
        </template>
      </div>
    </div>

    <div class="stage-verdict">
      <span class="vk">{{ stage.verdictKey }}</span>
      <div class="vt">{{ stage.verdictText }}</div>
      <span class="vr">{{ stage.verdictRight }}</span>
    </div>
  </div>
</template>
