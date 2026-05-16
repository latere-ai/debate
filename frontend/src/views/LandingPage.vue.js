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
    const io = new IntersectionObserver(entries => {
        entries.forEach(en => {
            if (en.isIntersecting) {
                en.target.classList.add('in');
                io.unobserve(en.target);
            }
        });
    }, { threshold: 0.08, rootMargin: '0px 0px -8% 0px' });
    els.forEach(e => io.observe(e));
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {[typeof DefaultLayout, typeof DefaultLayout, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(DefaultLayout, new DefaultLayout({}));
const __VLS_1 = __VLS_0({}, ...__VLS_functionalComponentArgsRest(__VLS_0));
var __VLS_3 = {};
__VLS_2.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    id: "top",
    ...{ class: "hero" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-bg" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-inner" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "hero-stamp" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "stamp-dot-p" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.c.hero.stampProposer);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.c.hero.stampCritic);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "stamp-dot-c" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({
    ...{ class: "hero-title" },
});
__VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.c.hero.title) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "hero-sub" },
});
__VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.c.hero.sub) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hero-actions" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "btn btn-primary" },
    href: "#install",
});
(__VLS_ctx.c.hero.ctaPrimary);
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "btn btn-ghost" },
    href: "#transcript",
});
(__VLS_ctx.c.hero.ctaSecondary);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "harnesses" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "harnesses-label" },
});
(__VLS_ctx.c.hero.worksWith);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "harnesses-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "harness" },
    href: "https://topos.latere.ai",
    title: "Topos — agent platform by Latere",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "1.7",
    'stroke-linecap': "round",
    'stroke-linejoin': "round",
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "6",
    cy: "7",
    r: "2.2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "17",
    cy: "6",
    r: "2.2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "18",
    cy: "17",
    r: "2.2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "7",
    cy: "18",
    r: "2.2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M8.1 7.4c2.3 1.3 4.8 1.1 6.9-.5M16.5 8.1c1.4 2 1.8 4.3 1.5 6.7M15.9 17.4c-2.1.9-4.4 1.1-6.7.6M6.8 15.8c-.7-2.2-.8-4.4-.2-6.6M9 9.1l6 6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.em, __VLS_intrinsicElements.em)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "harness" },
    href: "https://claude.com/product/claude-code",
    title: "Claude Code by Anthropic",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "currentColor",
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M12 2.2 L13.1 9.2 L19.6 6.6 L15 11.7 L21.6 14.3 L14.6 14.9 L17 21.6 L12 16.4 L7 21.6 L9.4 14.9 L2.4 14.3 L9 11.7 L4.4 6.6 L10.9 9.2 Z",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "harness" },
    href: "https://openai.com/codex",
    title: "Codex by OpenAI",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "1.5",
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "12",
    cy: "5.6",
    r: "2.6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "6.4",
    cy: "8.8",
    r: "2.6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "6.4",
    cy: "15.2",
    r: "2.6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "12",
    cy: "18.4",
    r: "2.6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "17.6",
    cy: "15.2",
    r: "2.6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "17.6",
    cy: "8.8",
    r: "2.6",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "harness" },
    href: "https://github.com/features/actions",
    title: "GitHub Actions — automated PR review comments",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "currentColor",
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
/** @type {[typeof HeroStage, ]} */ ;
// @ts-ignore
const __VLS_4 = __VLS_asFunctionalComponent(HeroStage, new HeroStage({}));
const __VLS_5 = __VLS_4({}, ...__VLS_functionalComponentArgsRest(__VLS_4));
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_7 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "transcript",
    eyebrow: (__VLS_ctx.c.transcript.eyebrow),
    title: (__VLS_ctx.c.transcript.title),
    lead: (__VLS_ctx.c.transcript.lead),
}));
const __VLS_8 = __VLS_7({
    id: "transcript",
    eyebrow: (__VLS_ctx.c.transcript.eyebrow),
    title: (__VLS_ctx.c.transcript.title),
    lead: (__VLS_ctx.c.transcript.lead),
}, ...__VLS_functionalComponentArgsRest(__VLS_7));
__VLS_9.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tx" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tx-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "lights" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tx-case" },
});
(__VLS_ctx.c.transcript.case);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tx-meta" },
});
(__VLS_ctx.c.transcript.meta);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tx-body" },
});
for (const [r, i] of __VLS_getVForSourceType((__VLS_ctx.c.transcript.rows))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
        ...{ class: "tx-msg" },
        ...{ class: ({ staked: r.isStake }) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "tx-n" },
    });
    (r.n);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "tx-author" },
        ...{ class: (r.actor) },
    });
    (r.actorLabel);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "tx-content" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (r.html) }, null, null);
    if (r.tag) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "tx-pill" },
            ...{ class: (r.tag.kind) },
        });
        (r.tag.label);
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tx-foot" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.c.transcript.footer[0]);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "tx-v" },
});
(__VLS_ctx.c.transcript.footer[1]);
var __VLS_9;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_10 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "why",
    eyebrow: (__VLS_ctx.c.why.eyebrow),
    title: (__VLS_ctx.c.why.title),
}));
const __VLS_11 = __VLS_10({
    id: "why",
    eyebrow: (__VLS_ctx.c.why.eyebrow),
    title: (__VLS_ctx.c.why.title),
}, ...__VLS_functionalComponentArgsRest(__VLS_10));
__VLS_12.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "pillars" },
});
for (const [pl, i] of __VLS_getVForSourceType((__VLS_ctx.c.why.pillars))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
        ...{ class: "pillar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "pillar-k" },
    });
    (pl.k);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "pillar-t" },
    });
    (pl.t);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "pillar-d" },
    });
    (pl.d);
}
var __VLS_12;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "compare",
    eyebrow: (__VLS_ctx.c.compare.eyebrow),
    title: (__VLS_ctx.c.compare.title),
}));
const __VLS_14 = __VLS_13({
    id: "compare",
    eyebrow: (__VLS_ctx.c.compare.eyebrow),
    title: (__VLS_ctx.c.compare.title),
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "compare" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "compare-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
(__VLS_ctx.c.compare.headers[0]);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ch-agon" },
});
(__VLS_ctx.c.compare.headers[1]);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
(__VLS_ctx.c.compare.headers[2]);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
(__VLS_ctx.c.compare.headers[3]);
for (const [r, i] of __VLS_getVForSourceType((__VLS_ctx.c.compare.rows))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
        ...{ class: "compare-row" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "cr-prop" },
        'data-h': (__VLS_ctx.c.compare.headers[0]),
    });
    (r.p);
    for (const [cell, j] of __VLS_getVForSourceType((r.cols))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            key: (j),
            'data-h': (__VLS_ctx.c.compare.headers[j + 1]),
        });
        if (cell === 'agon') {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "cmp-tick yes" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "cr-agon" },
            });
        }
        else if (cell === 'partial') {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "cmp-tick partial" },
            });
        }
        else {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "cmp-tick no" },
            });
        }
    }
}
var __VLS_15;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_16 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "usecases",
    eyebrow: (__VLS_ctx.c.usecases.eyebrow),
    title: (__VLS_ctx.c.usecases.title),
    lead: (__VLS_ctx.c.usecases.lead),
}));
const __VLS_17 = __VLS_16({
    id: "usecases",
    eyebrow: (__VLS_ctx.c.usecases.eyebrow),
    title: (__VLS_ctx.c.usecases.title),
    lead: (__VLS_ctx.c.usecases.lead),
}, ...__VLS_functionalComponentArgsRest(__VLS_16));
__VLS_18.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "usecases" },
});
for (const [it, i] of __VLS_getVForSourceType((__VLS_ctx.c.usecases.items))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
        ...{ class: "usecase" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "uc-head" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "uc-icon" },
    });
    (it.i);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "uc-t" },
    });
    (it.t);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "uc-d" },
    });
    (it.d);
}
var __VLS_18;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "architecture",
    eyebrow: (__VLS_ctx.c.arch.eyebrow),
    title: (__VLS_ctx.c.arch.title),
    lead: (__VLS_ctx.c.arch.lead),
}));
const __VLS_20 = __VLS_19({
    id: "architecture",
    eyebrow: (__VLS_ctx.c.arch.eyebrow),
    title: (__VLS_ctx.c.arch.title),
    lead: (__VLS_ctx.c.arch.lead),
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
__VLS_21.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "arch arch-dark" },
});
/** @type {[typeof ArchitectureSvg, ]} */ ;
// @ts-ignore
const __VLS_22 = __VLS_asFunctionalComponent(ArchitectureSvg, new ArchitectureSvg({}));
const __VLS_23 = __VLS_22({}, ...__VLS_functionalComponentArgsRest(__VLS_22));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "arch-cap" },
});
(__VLS_ctx.c.arch.cap);
var __VLS_21;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "signal",
    eyebrow: (__VLS_ctx.c.signal.eyebrow),
    title: (__VLS_ctx.c.signal.title),
    lead: (__VLS_ctx.c.signal.lead),
}));
const __VLS_26 = __VLS_25({
    id: "signal",
    eyebrow: (__VLS_ctx.c.signal.eyebrow),
    title: (__VLS_ctx.c.signal.title),
    lead: (__VLS_ctx.c.signal.lead),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "signal-grid" },
});
for (const [cell, i] of __VLS_getVForSourceType((__VLS_ctx.c.signal.cells))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
        ...{ class: "signal-cell" },
        ...{ class: (cell.kind === 'r' ? 'resolved' : 'contested') },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "signal-tag" },
        ...{ class: (cell.kind) },
    });
    (cell.tag);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "signal-num" },
    });
    (cell.num);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
    (cell.label);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "signal-route" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (cell.route);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "signal-d" },
    });
    (cell.desc);
}
var __VLS_27;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_28 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "foundations",
    eyebrow: (__VLS_ctx.c.found.eyebrow),
    title: (__VLS_ctx.c.found.title),
    lead: (__VLS_ctx.c.found.lead),
}));
const __VLS_29 = __VLS_28({
    id: "foundations",
    eyebrow: (__VLS_ctx.c.found.eyebrow),
    title: (__VLS_ctx.c.found.title),
    lead: (__VLS_ctx.c.found.lead),
}, ...__VLS_functionalComponentArgsRest(__VLS_28));
__VLS_30.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "refs" },
});
for (const [r, i] of __VLS_getVForSourceType((__VLS_ctx.c.found.refs))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
        ...{ class: "ref" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "ref-yr" },
    });
    (r.yr);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "ref-body" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "cite" },
    });
    (r.cite);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.em, __VLS_intrinsicElements.em)({});
    (r.em);
    (r.tail);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
        ...{ class: "ref-link" },
        href: (r.href),
        target: "_blank",
        rel: "noopener",
    });
    (r.link);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "pullquote" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.q, __VLS_intrinsicElements.q)({});
(__VLS_ctx.c.found.pullquote.q);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "cite" },
});
(__VLS_ctx.c.found.pullquote.cite);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "honest" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.c.found.honestStrong);
(__VLS_ctx.c.found.honestRest);
var __VLS_30;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "hook",
    eyebrow: (__VLS_ctx.c.hook.eyebrow),
    title: (__VLS_ctx.c.hook.title),
}));
const __VLS_32 = __VLS_31({
    id: "hook",
    eyebrow: (__VLS_ctx.c.hook.eyebrow),
    title: (__VLS_ctx.c.hook.title),
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
__VLS_33.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hook" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hook-body" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hook-d" },
});
(__VLS_ctx.c.hook.desc);
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "btn btn-primary" },
    href: "#install",
    ...{ style: {} },
});
(__VLS_ctx.c.hook.cta);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "hook-diagram" },
});
for (const [l, i] of __VLS_getVForSourceType((__VLS_ctx.c.hook.lines))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "hd-l" },
    });
    (l.p);
    (l.p ? ' ' : '');
    if (l.cmd) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "hd-cmd" },
        });
        (l.cmd);
    }
    if (l.attn) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "hd-attn" },
        });
        (l.attn);
    }
    if (l.l) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "hd-l" },
        });
        (l.l);
    }
}
var __VLS_33;
/** @type {[typeof SectionShell, typeof SectionShell, ]} */ ;
// @ts-ignore
const __VLS_34 = __VLS_asFunctionalComponent(SectionShell, new SectionShell({
    id: "faq",
    eyebrow: (__VLS_ctx.c.faq.eyebrow),
    title: (__VLS_ctx.c.faq.title),
}));
const __VLS_35 = __VLS_34({
    id: "faq",
    eyebrow: (__VLS_ctx.c.faq.eyebrow),
    title: (__VLS_ctx.c.faq.title),
}, ...__VLS_functionalComponentArgsRest(__VLS_34));
__VLS_36.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "faq" },
});
for (const [it, i] of __VLS_getVForSourceType((__VLS_ctx.c.faq.items))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.details, __VLS_intrinsicElements.details)({
        key: (i),
        ...{ class: "faq-item" },
        open: (i === 0),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.summary, __VLS_intrinsicElements.summary)({});
    (it.q);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "faq-body" },
    });
    (it.a);
}
var __VLS_36;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    id: "install",
    ...{ class: "install-wrap" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "install" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "eyebrow" },
});
(__VLS_ctx.c.install.eyebrow);
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.c.install.title) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.c.install.lead);
/** @type {[typeof CodeBlock, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(CodeBlock, new CodeBlock({
    comment: (__VLS_ctx.c.install.a.c),
    command: (__VLS_ctx.c.install.a.cmd),
}));
const __VLS_38 = __VLS_37({
    comment: (__VLS_ctx.c.install.a.c),
    command: (__VLS_ctx.c.install.a.cmd),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
/** @type {[typeof CodeBlock, ]} */ ;
// @ts-ignore
const __VLS_40 = __VLS_asFunctionalComponent(CodeBlock, new CodeBlock({
    comment: (__VLS_ctx.c.install.b.c),
    command: (__VLS_ctx.c.install.b.cmd),
}));
const __VLS_41 = __VLS_40({
    comment: (__VLS_ctx.c.install.b.c),
    command: (__VLS_ctx.c.install.b.cmd),
}, ...__VLS_functionalComponentArgsRest(__VLS_40));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "install-cta" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "btn btn-primary" },
    href: "https://github.com/latere-ai/agon",
    target: "_blank",
    rel: "noopener",
});
(__VLS_ctx.c.install.ctaPrimary);
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "btn btn-ghost" },
    href: "https://github.com/latere-ai/agon#readme",
    target: "_blank",
    rel: "noopener",
});
(__VLS_ctx.c.install.ctaSecondary);
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['hero']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-bg']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['container']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-inner']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-stamp']} */ ;
/** @type {__VLS_StyleScopedClasses['stamp-dot-p']} */ ;
/** @type {__VLS_StyleScopedClasses['stamp-dot-c']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-title']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['hero-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-primary']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-ghost']} */ ;
/** @type {__VLS_StyleScopedClasses['harnesses']} */ ;
/** @type {__VLS_StyleScopedClasses['harnesses-label']} */ ;
/** @type {__VLS_StyleScopedClasses['harnesses-row']} */ ;
/** @type {__VLS_StyleScopedClasses['harness']} */ ;
/** @type {__VLS_StyleScopedClasses['harness']} */ ;
/** @type {__VLS_StyleScopedClasses['harness']} */ ;
/** @type {__VLS_StyleScopedClasses['harness']} */ ;
/** @type {__VLS_StyleScopedClasses['tx']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-head']} */ ;
/** @type {__VLS_StyleScopedClasses['lights']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-case']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-body']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-msg']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-n']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-author']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-content']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-pill']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-foot']} */ ;
/** @type {__VLS_StyleScopedClasses['tx-v']} */ ;
/** @type {__VLS_StyleScopedClasses['pillars']} */ ;
/** @type {__VLS_StyleScopedClasses['pillar']} */ ;
/** @type {__VLS_StyleScopedClasses['pillar-k']} */ ;
/** @type {__VLS_StyleScopedClasses['pillar-t']} */ ;
/** @type {__VLS_StyleScopedClasses['pillar-d']} */ ;
/** @type {__VLS_StyleScopedClasses['compare']} */ ;
/** @type {__VLS_StyleScopedClasses['compare-head']} */ ;
/** @type {__VLS_StyleScopedClasses['ch-agon']} */ ;
/** @type {__VLS_StyleScopedClasses['compare-row']} */ ;
/** @type {__VLS_StyleScopedClasses['cr-prop']} */ ;
/** @type {__VLS_StyleScopedClasses['cmp-tick']} */ ;
/** @type {__VLS_StyleScopedClasses['yes']} */ ;
/** @type {__VLS_StyleScopedClasses['cr-agon']} */ ;
/** @type {__VLS_StyleScopedClasses['cmp-tick']} */ ;
/** @type {__VLS_StyleScopedClasses['partial']} */ ;
/** @type {__VLS_StyleScopedClasses['cmp-tick']} */ ;
/** @type {__VLS_StyleScopedClasses['no']} */ ;
/** @type {__VLS_StyleScopedClasses['usecases']} */ ;
/** @type {__VLS_StyleScopedClasses['usecase']} */ ;
/** @type {__VLS_StyleScopedClasses['uc-head']} */ ;
/** @type {__VLS_StyleScopedClasses['uc-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['uc-t']} */ ;
/** @type {__VLS_StyleScopedClasses['uc-d']} */ ;
/** @type {__VLS_StyleScopedClasses['arch']} */ ;
/** @type {__VLS_StyleScopedClasses['arch-dark']} */ ;
/** @type {__VLS_StyleScopedClasses['arch-cap']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-cell']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-num']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-route']} */ ;
/** @type {__VLS_StyleScopedClasses['signal-d']} */ ;
/** @type {__VLS_StyleScopedClasses['refs']} */ ;
/** @type {__VLS_StyleScopedClasses['ref']} */ ;
/** @type {__VLS_StyleScopedClasses['ref-yr']} */ ;
/** @type {__VLS_StyleScopedClasses['ref-body']} */ ;
/** @type {__VLS_StyleScopedClasses['cite']} */ ;
/** @type {__VLS_StyleScopedClasses['ref-link']} */ ;
/** @type {__VLS_StyleScopedClasses['pullquote']} */ ;
/** @type {__VLS_StyleScopedClasses['cite']} */ ;
/** @type {__VLS_StyleScopedClasses['honest']} */ ;
/** @type {__VLS_StyleScopedClasses['hook']} */ ;
/** @type {__VLS_StyleScopedClasses['hook-body']} */ ;
/** @type {__VLS_StyleScopedClasses['hook-d']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-primary']} */ ;
/** @type {__VLS_StyleScopedClasses['hook-diagram']} */ ;
/** @type {__VLS_StyleScopedClasses['hd-l']} */ ;
/** @type {__VLS_StyleScopedClasses['hd-cmd']} */ ;
/** @type {__VLS_StyleScopedClasses['hd-attn']} */ ;
/** @type {__VLS_StyleScopedClasses['hd-l']} */ ;
/** @type {__VLS_StyleScopedClasses['faq']} */ ;
/** @type {__VLS_StyleScopedClasses['faq-item']} */ ;
/** @type {__VLS_StyleScopedClasses['faq-body']} */ ;
/** @type {__VLS_StyleScopedClasses['install-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['install']} */ ;
/** @type {__VLS_StyleScopedClasses['eyebrow']} */ ;
/** @type {__VLS_StyleScopedClasses['install-cta']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-primary']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-ghost']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            DefaultLayout: DefaultLayout,
            SectionShell: SectionShell,
            HeroStage: HeroStage,
            ArchitectureSvg: ArchitectureSvg,
            CodeBlock: CodeBlock,
            c: c,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
