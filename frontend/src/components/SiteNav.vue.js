import { ref, computed } from 'vue';
import { useContent } from '../content';
const content = useContent();
const nav = computed(() => content.value.nav);
const open = ref(false);
function close() {
    open.value = false;
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.header, __VLS_intrinsicElements.header)({
    ...{ class: "nav" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "nav-inner" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ onClick: (__VLS_ctx.close) },
    href: "#top",
    ...{ class: "nav-brand" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "agon-mark" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
(__VLS_ctx.nav.by);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.open = !__VLS_ctx.open;
        } },
    ...{ class: "nav-toggle" },
    type: "button",
    'aria-expanded': (__VLS_ctx.open),
    'aria-label': "Toggle navigation",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    width: "18",
    height: "18",
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "2",
    'stroke-linecap': "round",
    'aria-hidden': "true",
});
if (!__VLS_ctx.open) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
        d: "M3 6h18M3 12h18M3 18h18",
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
        d: "M6 6l12 12M6 18L18 6",
    });
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.nav, __VLS_intrinsicElements.nav)({
    ...{ class: "nav-links" },
    ...{ class: ({ 'is-open': __VLS_ctx.open }) },
});
for (const [link] of __VLS_getVForSourceType((__VLS_ctx.nav.links))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
        ...{ onClick: (__VLS_ctx.close) },
        key: (link.href),
        ...{ class: "nav-link" },
        href: (link.href),
    });
    (link.label);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ onClick: (__VLS_ctx.close) },
    ...{ class: "btn btn-ghost" },
    href: "https://github.com/latere-ai/agon",
    target: "_blank",
    rel: "noopener",
    ...{ style: {} },
});
(__VLS_ctx.nav.github);
__VLS_asFunctionalElement(__VLS_intrinsicElements.a, __VLS_intrinsicElements.a)({
    ...{ onClick: (__VLS_ctx.close) },
    ...{ class: "btn btn-primary" },
    href: "#install",
});
(__VLS_ctx.nav.install);
/** @type {__VLS_StyleScopedClasses['nav']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-inner']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-brand']} */ ;
/** @type {__VLS_StyleScopedClasses['agon-mark']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-toggle']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-links']} */ ;
/** @type {__VLS_StyleScopedClasses['nav-link']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-ghost']} */ ;
/** @type {__VLS_StyleScopedClasses['btn']} */ ;
/** @type {__VLS_StyleScopedClasses['btn-primary']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            nav: nav,
            open: open,
            close: close,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
