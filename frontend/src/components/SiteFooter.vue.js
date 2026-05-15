import { useT } from '../i18n';
const t = useT();
const year = new Date().getFullYear();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.footer, __VLS_intrinsicElements.footer)({
    ...{ class: "site-footer" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-links" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://github.com/latere-ai/debate",
    target: "_blank",
    rel: "noopener",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://github.com/latere-ai/debate#readme",
    target: "_blank",
    rel: "noopener",
});
(__VLS_ctx.t('footer.docs'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://github.com/changkun/agents-byzantine-tolerance",
    target: "_blank",
    rel: "noopener",
});
(__VLS_ctx.t('footer.research'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    href: "https://latere.ai/",
    target: "_blank",
    rel: "noopener",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "footer-meta" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "agon-brand" },
});
(__VLS_ctx.t('footer.tagline'));
(__VLS_ctx.year);
/** @type {__VLS_StyleScopedClasses['site-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-row']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-links']} */ ;
/** @type {__VLS_StyleScopedClasses['footer-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['agon-brand']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            year: year,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
