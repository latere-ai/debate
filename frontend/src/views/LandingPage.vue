<script setup lang="ts">
import { onMounted, computed } from 'vue';
import { useHead } from '@unhead/vue';
import DefaultLayout from '../layouts/DefaultLayout.vue';
import { useContent } from '../content';
import SectionShell from '../components/landing/SectionShell.vue';
import HeroStage from '../components/landing/HeroStage.vue';
import ArchitectureSvg from '../components/landing/ArchitectureSvg.vue';
import CodeBlock from '../components/landing/CodeBlock.vue';

const content = useContent();
const c = content; // alias for template brevity

useHead({
  title: computed(() => content.value.meta.title),
  meta: [{ name: 'description', content: computed(() => content.value.meta.description) }],
});

onMounted(() => {
  const els = document.querySelectorAll('.reveal:not(.in)');
  if (!('IntersectionObserver' in window)) {
    els.forEach(e => e.classList.add('in'));
    return;
  }
  const io = new IntersectionObserver(
    entries => {
      entries.forEach(en => {
        if (en.isIntersecting) {
          en.target.classList.add('in');
          io.unobserve(en.target);
        }
      });
    },
    { threshold: 0.08, rootMargin: '0px 0px -8% 0px' },
  );
  els.forEach(e => io.observe(e));
});
</script>

<template>
  <DefaultLayout>
    <!-- Hero -->
    <section id="top" class="hero">
      <div class="hero-bg"></div>
      <div class="hero-grid"></div>
      <div class="container">
        <div class="hero-inner">
          <span class="hero-stamp">
            <span class="stamp-dot-p"></span>
            <span>{{ c.hero.stampProposer }}</span>
            <span>×</span>
            <span>{{ c.hero.stampCritic }}</span>
            <span class="stamp-dot-c"></span>
          </span>
          <h1 class="hero-title" v-html="c.hero.title"></h1>
          <p class="hero-sub" v-html="c.hero.sub"></p>
          <div class="hero-actions">
            <a class="btn btn-primary" href="#install">{{ c.hero.ctaPrimary }}</a>
            <a class="btn btn-ghost" href="#transcript">{{ c.hero.ctaSecondary }}</a>
          </div>

          <div class="harnesses">
            <span class="harnesses-label">{{ c.hero.worksWith }}</span>
            <div class="harnesses-row">
              <a class="harness" href="https://topos.latere.ai" title="Topos — agent platform by Latere">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <circle cx="6" cy="7" r="2.2" />
                  <circle cx="17" cy="6" r="2.2" />
                  <circle cx="18" cy="17" r="2.2" />
                  <circle cx="7" cy="18" r="2.2" />
                  <path d="M8.1 7.4c2.3 1.3 4.8 1.1 6.9-.5M16.5 8.1c1.4 2 1.8 4.3 1.5 6.7M15.9 17.4c-2.1.9-4.4 1.1-6.7.6M6.8 15.8c-.7-2.2-.8-4.4-.2-6.6M9 9.1l6 6" />
                </svg>
                <span><em>Topos</em></span>
              </a>
              <a class="harness" href="https://claude.com/product/claude-code" title="Claude Code by Anthropic">
                <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path d="M12 2.2 L13.1 9.2 L19.6 6.6 L15 11.7 L21.6 14.3 L14.6 14.9 L17 21.6 L12 16.4 L7 21.6 L9.4 14.9 L2.4 14.3 L9 11.7 L4.4 6.6 L10.9 9.2 Z" />
                </svg>
                <span>Claude Code</span>
              </a>
              <a class="harness" href="https://openai.com/codex" title="Codex by OpenAI">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                  <circle cx="12" cy="5.6" r="2.6" />
                  <circle cx="6.4" cy="8.8" r="2.6" />
                  <circle cx="6.4" cy="15.2" r="2.6" />
                  <circle cx="12" cy="18.4" r="2.6" />
                  <circle cx="17.6" cy="15.2" r="2.6" />
                  <circle cx="17.6" cy="8.8" r="2.6" />
                </svg>
                <span>Codex</span>
              </a>
              <a class="harness" href="https://github.com/features/actions" title="GitHub Actions — automated PR review comments">
                <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
                </svg>
                <span>GitHub Action</span>
              </a>
            </div>
          </div>
        </div>

        <HeroStage />
      </div>
    </section>

    <!-- Transcript -->
    <SectionShell
      id="transcript"
      :eyebrow="c.transcript.eyebrow"
      :title="c.transcript.title"
      :lead="c.transcript.lead">
      <div class="tx">
        <div class="tx-head">
          <div class="lights"><span></span><span></span><span></span></div>
          <div class="tx-case">{{ c.transcript.case }}</div>
          <div class="tx-meta">{{ c.transcript.meta }}</div>
        </div>
        <div class="tx-body">
          <div
            v-for="(r, i) in c.transcript.rows"
            :key="i"
            class="tx-msg"
            :class="{ staked: r.isStake }">
            <div class="tx-n">{{ r.n }}</div>
            <div>
              <div class="tx-author" :class="r.actor">{{ r.actorLabel }}</div>
              <div class="tx-content" v-html="r.html"></div>
            </div>
            <div v-if="r.tag" class="tx-pill" :class="r.tag.kind">{{ r.tag.label }}</div>
            <div v-else></div>
          </div>
        </div>
        <div class="tx-foot">
          <span>{{ c.transcript.footer[0] }}</span>
          <span class="tx-v">{{ c.transcript.footer[1] }}</span>
        </div>
      </div>
    </SectionShell>

    <!-- Why -->
    <SectionShell id="why" :eyebrow="c.why.eyebrow" :title="c.why.title">
      <div class="pillars">
        <div v-for="(pl, i) in c.why.pillars" :key="i" class="pillar">
          <div class="pillar-k">{{ pl.k }} · property</div>
          <div class="pillar-t">{{ pl.t }}</div>
          <div class="pillar-d">{{ pl.d }}</div>
        </div>
      </div>
    </SectionShell>

    <!-- Compare -->
    <SectionShell id="compare" :eyebrow="c.compare.eyebrow" :title="c.compare.title">
      <div class="compare">
        <div class="compare-head">
          <div>{{ c.compare.headers[0] }}</div>
          <div class="ch-agon">{{ c.compare.headers[1] }}</div>
          <div>{{ c.compare.headers[2] }}</div>
          <div>{{ c.compare.headers[3] }}</div>
        </div>
        <div v-for="(r, i) in c.compare.rows" :key="i" class="compare-row">
          <div class="cr-prop" :data-h="c.compare.headers[0]">{{ r.p }}</div>
          <div
            v-for="(cell, j) in r.cols"
            :key="j"
            :data-h="c.compare.headers[j + 1]">
            <span v-if="cell === 'agon'"><span class="cmp-tick yes">✓</span><span class="cr-agon">yes</span></span>
            <span v-else-if="cell === 'partial'"><span class="cmp-tick partial">~</span>partial</span>
            <span v-else><span class="cmp-tick no">·</span>no</span>
          </div>
        </div>
      </div>
    </SectionShell>

    <!-- Use cases -->
    <SectionShell
      id="usecases"
      :eyebrow="c.usecases.eyebrow"
      :title="c.usecases.title"
      :lead="c.usecases.lead">
      <div class="usecases">
        <div v-for="(it, i) in c.usecases.items" :key="i" class="usecase">
          <div class="uc-head">
            <div class="uc-icon">{{ it.i }}</div>
            <div class="uc-t">{{ it.t }}</div>
          </div>
          <div class="uc-d">{{ it.d }}</div>
        </div>
      </div>
    </SectionShell>

    <!-- Architecture -->
    <SectionShell
      id="architecture"
      :eyebrow="c.arch.eyebrow"
      :title="c.arch.title"
      :lead="c.arch.lead">
      <div class="arch arch-dark">
        <ArchitectureSvg />
        <div class="arch-cap">{{ c.arch.cap }}</div>
      </div>
    </SectionShell>

    <!-- Signal -->
    <SectionShell
      id="signal"
      :eyebrow="c.signal.eyebrow"
      :title="c.signal.title"
      :lead="c.signal.lead">
      <div class="signal-grid">
        <div
          v-for="(cell, i) in c.signal.cells"
          :key="i"
          class="signal-cell"
          :class="cell.kind === 'r' ? 'resolved' : 'contested'">
          <span class="signal-tag" :class="cell.kind">{{ cell.tag }}</span>
          <div class="signal-num">{{ cell.num }} <small>{{ cell.label }}</small></div>
          <div class="signal-route"><strong>{{ cell.route }}</strong></div>
          <div class="signal-d">{{ cell.desc }}</div>
        </div>
      </div>
    </SectionShell>

    <!-- Foundations -->
    <SectionShell
      id="foundations"
      :eyebrow="c.found.eyebrow"
      :title="c.found.title"
      :lead="c.found.lead">
      <div class="refs">
        <div v-for="(r, i) in c.found.refs" :key="i" class="ref">
          <div class="ref-yr">{{ r.yr }}</div>
          <div class="ref-body">
            <span class="cite">{{ r.cite }}.</span> <em>{{ r.em }}</em>{{ r.tail }}
          </div>
          <a class="ref-link" :href="r.href" target="_blank" rel="noopener">{{ r.link }}</a>
        </div>
      </div>
      <div class="pullquote">
        <q>{{ c.found.pullquote.q }}</q>
        <div class="cite">— {{ c.found.pullquote.cite }}</div>
      </div>
      <div class="honest">
        <strong>{{ c.found.honestStrong }}</strong>{{ c.found.honestRest }}
      </div>
    </SectionShell>

    <!-- Stop-hook -->
    <SectionShell id="hook" :eyebrow="c.hook.eyebrow" :title="c.hook.title">
      <div class="hook">
        <div class="hook-body">
          <div class="hook-d">{{ c.hook.desc }}</div>
          <a class="btn btn-primary" href="#install" style="align-self: flex-start">{{ c.hook.cta }}</a>
        </div>
        <div class="hook-diagram">
          <div v-for="(l, i) in c.hook.lines" :key="i">
            <span class="hd-l">{{ l.p }}{{ l.p ? ' ' : '' }}</span>
            <span v-if="l.cmd" class="hd-cmd">{{ l.cmd }}</span>
            <span v-if="l.attn" class="hd-attn">{{ l.attn }}</span>
            <span v-if="l.l" class="hd-l">{{ l.l }}</span>
          </div>
        </div>
      </div>
    </SectionShell>

    <!-- FAQ -->
    <SectionShell id="faq" :eyebrow="c.faq.eyebrow" :title="c.faq.title">
      <div class="faq">
        <details
          v-for="(it, i) in c.faq.items"
          :key="i"
          class="faq-item"
          :open="i === 0">
          <summary>{{ it.q }}</summary>
          <div class="faq-body">{{ it.a }}</div>
        </details>
      </div>
    </SectionShell>

    <!-- Install -->
    <div id="install" class="install-wrap">
      <div class="install">
        <span class="eyebrow">{{ c.install.eyebrow }}</span>
        <h2 v-html="c.install.title"></h2>
        <p>{{ c.install.lead }}</p>
        <CodeBlock :comment="c.install.a.c" :command="c.install.a.cmd" />
        <CodeBlock :comment="c.install.b.c" :command="c.install.b.cmd" />
        <div class="install-cta">
          <a class="btn btn-primary" href="https://github.com/latere-ai/agon" target="_blank" rel="noopener">{{ c.install.ctaPrimary }}</a>
          <a class="btn btn-ghost" href="https://github.com/latere-ai/agon#readme" target="_blank" rel="noopener">{{ c.install.ctaSecondary }}</a>
        </div>
      </div>
    </div>
  </DefaultLayout>
</template>
