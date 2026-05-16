<script setup lang="ts">
import { onMounted } from 'vue';
import { useHead } from '@unhead/vue';
import DefaultLayout from '../layouts/DefaultLayout.vue';
import { useT } from '../i18n';

const t = useT();

useHead({
  title: 'Agon: adversarial verification for AI-produced work',
  meta: [
    { name: 'description', content: 'An independent critic cross-examines AI-produced work; the producer defends or concedes; only surviving disputes reach a human. Vendor-neutral, auditable, formally grounded.' },
  ],
});

onMounted(() => {
  const els = document.querySelectorAll('.reveal');
  if (!('IntersectionObserver' in window)) {
    els.forEach(e => e.classList.add('in'));
    return;
  }
  const io = new IntersectionObserver((entries) => {
    entries.forEach(en => { if (en.isIntersecting) { en.target.classList.add('in'); io.unobserve(en.target); } });
  }, { threshold: 0.12 });
  els.forEach(e => io.observe(e));
});
</script>

<template>
  <DefaultLayout>
    <!-- Hero -->
    <section class="hero">
      <div class="container hero-inner">
        <span class="hero-eyebrow">{{ t('hero.eyebrow') }}</span>
        <h1 class="hero-title" v-html="t('hero.title')"></h1>
        <p class="hero-sub">{{ t('hero.sub') }}</p>
        <div class="hero-actions">
          <a class="btn btn-primary" href="#install">{{ t('hero.cta.install') }}</a>
          <a class="btn btn-ghost" href="https://github.com/latere-ai/debate" target="_blank" rel="noopener">{{ t('hero.cta.repo') }}</a>
        </div>
      </div>
    </section>

    <!-- What it is -->
    <section class="section">
      <div class="container reveal">
        <span class="section-label">{{ t('what.label') }}</span>
        <h2 class="section-title" v-html="t('what.title')"></h2>
        <div class="prose">
          <p v-html="t('what.p1')"></p>
          <p v-html="t('what.p2')"></p>
        </div>
      </div>
    </section>

    <!-- The protocol -->
    <section class="section">
      <div class="container reveal">
        <span class="section-label">{{ t('proto.label') }}</span>
        <h2 class="section-title">{{ t('proto.title') }}</h2>
        <p class="section-lead">{{ t('proto.lead') }}</p>
        <div class="steps">
          <div class="step" v-for="n in 4" :key="n">
            <span class="step-k">R{{ n }}</span>
            <div class="step-t">{{ t('proto.s' + n + '.t') }}</div>
            <p class="step-d">{{ t('proto.s' + n + '.d') }}</p>
          </div>
        </div>
        <p class="note">{{ t('proto.judge') }}</p>
      </div>
    </section>

    <!-- Why different -->
    <section class="section">
      <div class="container reveal">
        <span class="section-label">{{ t('why.label') }}</span>
        <h2 class="section-title">{{ t('why.title') }}</h2>
        <div class="grid grid-2">
          <div class="card">
            <h3>{{ t('why.c1.t') }}</h3>
            <p v-html="t('why.c1.d')"></p>
          </div>
          <div class="card">
            <h3>{{ t('why.c2.t') }}</h3>
            <p>{{ t('why.c2.d') }}</p>
          </div>
          <div class="card">
            <h3>{{ t('why.c3.t') }}</h3>
            <p>{{ t('why.c3.d') }}</p>
          </div>
          <div class="card">
            <h3>{{ t('why.c4.t') }}</h3>
            <p>{{ t('why.c4.d') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Resolved vs contested -->
    <section class="section">
      <div class="container reveal">
        <span class="section-label">{{ t('signal.label') }}</span>
        <h2 class="section-title">{{ t('signal.title') }}</h2>
        <div class="prose">
          <p v-html="t('signal.p1')"></p>
          <p v-html="t('signal.p2')"></p>
        </div>
      </div>
    </section>

    <!-- Academic foundations -->
    <section class="section">
      <div class="container reveal">
        <span class="section-label">{{ t('found.label') }}</span>
        <h2 class="section-title">{{ t('found.title') }}</h2>
        <div class="prose">
          <p v-html="t('found.p1')"></p>
          <p v-html="t('found.p2')"></p>
        </div>
        <div class="refs">
          <p class="ref">
            <span class="ref-cite">Irving, Christiano &amp; Amodei (2018).</span>
            <em>AI Safety via Debate.</em>
            <a href="https://arxiv.org/abs/1805.00899" target="_blank" rel="noopener">arXiv:1805.00899</a>
          </p>
          <p class="ref">
            <span class="ref-cite">Brown-Cohen, Irving &amp; Piliouras (2023).</span>
            <em>Scalable AI Safety via Doubly-Efficient Debate.</em>
            <a href="https://arxiv.org/abs/2311.14125" target="_blank" rel="noopener">arXiv:2311.14125</a>
          </p>
          <p class="ref">
            <span class="ref-cite">Brown-Cohen, Irving &amp; Piliouras (2025).</span>
            <em>Avoiding Obfuscation with Prover-Estimator Debate.</em>
            <a href="https://arxiv.org/abs/2506.13609" target="_blank" rel="noopener">arXiv:2506.13609</a>
          </p>
          <p class="ref">
            <span class="ref-cite">{{ t('found.research.cite') }}</span>
            <em>agents-byzantine-tolerance</em>: {{ t('found.research.note') }}
            <a href="https://github.com/changkun/agents-byzantine-tolerance" target="_blank" rel="noopener">github.com/changkun/agents-byzantine-tolerance</a>
          </p>
        </div>
        <p class="note">{{ t('found.honest') }}</p>
      </div>
    </section>

    <!-- Install -->
    <section class="section" id="install">
      <div class="container reveal">
        <span class="section-label">{{ t('install.label') }}</span>
        <h2 class="section-title">{{ t('install.title') }}</h2>
        <p class="section-lead">{{ t('install.lead') }}</p>
        <pre class="code"><span class="tok-c"># {{ t('install.c1') }}</span>
curl -fsSL https://raw.githubusercontent.com/latere-ai/debate/main/install.sh | sh</pre>
        <pre class="code"><span class="tok-c"># {{ t('install.c2') }}</span>
go install latere.ai/x/debate/cmd/debate@latest
debate install-hook --scope user</pre>
        <p class="prose" style="margin-top:18px;">
          <a class="btn btn-primary" href="https://github.com/latere-ai/debate" target="_blank" rel="noopener">{{ t('install.repo') }}</a>
        </p>
      </div>
    </section>
  </DefaultLayout>
</template>
