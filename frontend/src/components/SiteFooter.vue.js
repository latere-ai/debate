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
    return href.startsWith('http') || href.startsWith('mailto');
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.footer, __VLS_intrinsicElements.footer)({
    ...{ class: "site-footer" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-brand" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "#top",
    ...{ class: "logo-link" },
    'aria-label': "Agon",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    ...{ class: "logo-icon" },
    width: "24",
    height: "24",
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "1.8",
    'stroke-linecap': "round",
    'stroke-linejoin': "round",
    'aria-hidden': "true",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M4 6 L11 12 L4 18 Z",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M20 6 L13 12 L20 18 Z",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M12 3.5 V20.5",
    opacity: "0.7",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "logo-text" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "footer-tagline" },
});
(__VLS_ctx.footer.tagline);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-cols" },
});
for (const [col] of __VLS_getVForSourceType((__VLS_ctx.footer.columns))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (col.head),
        ...{ class: "footer-col" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h4, __VLS_intrinsicElements.h4)({
        ...{ class: "footer-col-title" },
    });
    (col.head);
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
    ...{ class: "footer-extra" },
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-social" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://www.linkedin.com/company/latere-ai/about/",
    target: "_blank",
    rel: "noopener",
    title: "LinkedIn",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    width: "16",
    height: "16",
    viewBox: "0 0 24 24",
    fill: "currentColor",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://x.com/LatereAI",
    target: "_blank",
    rel: "noopener",
    title: "X",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    width: "16",
    height: "16",
    viewBox: "0 0 24 24",
    fill: "currentColor",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://github.com/latere-ai/agon",
    target: "_blank",
    rel: "noopener",
    title: "GitHub",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    width: "16",
    height: "16",
    viewBox: "0 0 24 24",
    fill: "currentColor",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-bottom" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
__VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.metaHtml) }, null, null);
/** @type {__VLS_StyleScopedClasses['site-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-container']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-brand']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-link']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-text']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-tagline']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-cols']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-col']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-col-title']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-extra']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-prefs']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-social']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-bottom']} */ ;
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
