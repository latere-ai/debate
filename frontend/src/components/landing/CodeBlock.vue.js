import { ref, computed } from 'vue';
import { useContent } from '../../content';
const props = defineProps();
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
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "code" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "code-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "c-c" },
});
(__VLS_ctx.comment);
for (const [line, i] of __VLS_getVForSourceType((__VLS_ctx.lines))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (i),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "c-tok" },
    });
    (line);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.copy) },
    ...{ class: "copy-btn" },
    ...{ class: ({ copied: __VLS_ctx.copied }) },
    type: "button",
});
(__VLS_ctx.copied ? __VLS_ctx.labels.copied : __VLS_ctx.labels.copy);
/** @type {__VLS_StyleScopedClasses['code']} */ ;
/** @type {__VLS_StyleScopedClasses['code-content']} */ ;
/** @type {__VLS_StyleScopedClasses['c-c']} */ ;
/** @type {__VLS_StyleScopedClasses['c-tok']} */ ;
/** @type {__VLS_StyleScopedClasses['copy-btn']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            labels: labels,
            copied: copied,
            lines: lines,
            copy: copy,
        };
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
