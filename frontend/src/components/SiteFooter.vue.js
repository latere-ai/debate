import { computed } from 'vue';
import { storeToRefs } from 'pinia';
import { usePrefsStore } from '../stores/prefs';
import { useContent } from '../content';
const content = useContent();
const footer = computed(() => content.value.footer);
const metaHtml = computed(() => footer.value.meta.replace('{year}', String(new Date().getFullYear())));
const prefs = usePrefsStore();
const { theme, locale } = storeToRefs(prefs);
const themes = [
    { v: 'light', label: '☀' },
    { v: 'dark', label: '☾' },
    { v: 'auto', label: '◐' },
];
const locales = [
    { v: 'en', label: 'EN' },
    { v: 'zh', label: '中' },
];
function isExternal(href) {
    return href.startsWith('http');
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.footer, __VLS_intrinsicElements.footer)({
    ...{ class: "agon-footer" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-grid" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-brand-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    width: "22",
    height: "22",
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "1.6",
    'stroke-linecap': "round",
    'stroke-linejoin': "round",
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M4 16 C 4 8, 11 4, 14 8 C 17 12, 10 16, 10 16 C 10 16, 17 20, 20 16 C 23 12, 16 8, 16 8",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "footer-brand-text" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "footer-tagline" },
});
(__VLS_ctx.footer.tagline);
for (const [col] of __VLS_getVForSourceType((__VLS_ctx.footer.columns))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (col.head),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "footer-col-head" },
    });
    (col.head);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "footer-col" },
    });
    for (const [link] of __VLS_getVForSourceType((col.links))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
            key: (link.href),
            href: (link.href),
            target: (__VLS_ctx.isExternal(link.href) ? '_blank' : undefined),
            rel: (__VLS_ctx.isExternal(link.href) ? 'noopener' : undefined),
        });
        (link.label);
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-base" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "footer-meta" },
});
__VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.metaHtml) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-base-right" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-prefs" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-seg" },
    role: "group",
    'aria-label': "Theme",
});
for (const [opt] of __VLS_getVForSourceType((__VLS_ctx.themes))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.prefs.setTheme(opt.v);
            } },
        key: (opt.v),
        type: "button",
        ...{ class: "footer-seg-btn" },
        ...{ class: ({ 'is-active': __VLS_ctx.theme === opt.v }) },
    });
    (opt.label);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-seg" },
    role: "group",
    'aria-label': "Language",
});
for (const [opt] of __VLS_getVForSourceType((__VLS_ctx.locales))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.prefs.setLocale(opt.v);
            } },
        key: (opt.v),
        type: "button",
        ...{ class: "footer-seg-btn" },
        ...{ class: ({ 'is-active': __VLS_ctx.locale === opt.v }) },
    });
    (opt.label);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "footer-thesis" },
});
(__VLS_ctx.footer.thesis);
/** @type {__VLS_StyleScopedClasses['agon-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['container']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-brand-row']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-brand-text']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-tagline']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-col-head']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-col']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-base']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-base-right']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-prefs']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-thesis']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            footer: footer,
            metaHtml: metaHtml,
            prefs: prefs,
            theme: theme,
            locale: locale,
            themes: themes,
            locales: locales,
            isExternal: isExternal,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
