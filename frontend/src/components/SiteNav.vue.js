import { usePrefsStore } from '../stores/prefs';
import { useT } from '../i18n';
const prefs = usePrefsStore();
const t = useT();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.header, __VLS_intrinsicElements.header)({
    ...{ class: "site-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.nav, __VLS_intrinsicElements.nav)({
    ...{ class: "nav-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "/",
    ...{ class: "logo-link" },
    'aria-label': "Agon home",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "agon-brand" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "logo-sub" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "nav-actions" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "nav-link" },
    href: "https://github.com/latere-ai/debate",
    target: "_blank",
    rel: "noopener",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "nav-link" },
    href: "https://latere.ai/",
    target: "_blank",
    rel: "noopener",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "seg" },
    role: "group",
    'aria-label': "Language",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.prefs.setLocale('en');
        } },
    ...{ class: "seg-btn" },
    ...{ class: ({ 'is-active': __VLS_ctx.prefs.locale === 'en' }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.prefs.setLocale('zh');
        } },
    ...{ class: "seg-btn" },
    ...{ class: ({ 'is-active': __VLS_ctx.prefs.locale === 'zh' }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.prefs.toggleTheme();
        } },
    ...{ class: "seg-btn" },
    'aria-label': (__VLS_ctx.t('nav.theme')),
    ...{ style: {} },
});
(__VLS_ctx.prefs.themeIcon);
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ class: "nav-cta" },
    href: "#install",
});
(__VLS_ctx.t('nav.install'));
/** @type {__VLS_StyleScopedClasses['site-header']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-container']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-link']} */ ;
/** @type {__VLS_StyleScopedClasses['agon-brand']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-sub']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-link']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-link']} */ ;
/** @type {__VLS_StyleScopedClasses['seg']} */ ;
/** @type {__VLS_StyleScopedClasses['seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['seg-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-cta']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            prefs: prefs,
            t: t,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
