import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useContent } from '../../content';
const content = useContent();
const stage = computed(() => content.value.stage);
const lastIndex = computed(() => stage.value.rows.length - 1);
// Start on the final state so SSR renders something sensible and a
// reduced-motion visit gets the resolved/staked frame.
const active = ref(stage.value.rows.length - 1);
let timer;
onMounted(() => {
    const reduce = typeof window !== 'undefined' &&
        window.matchMedia &&
        window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (reduce)
        return;
    active.value = 0;
    timer = setInterval(() => {
        active.value = (active.value + 1) % stage.value.rows.length;
    }, 2400);
});
onUnmounted(() => {
    if (timer)
        clearInterval(timer);
});
function bubbleStyle(i) {
    return {
        opacity: active.value >= i ? 1 : 0.3,
        transform: active.value === i ? 'scale(1.02)' : 'scale(1)',
    };
}
function spineClass(i, side) {
    const cls = ['spine-stop'];
    if (active.value >= i)
        cls.push('is-' + side);
    if (i === lastIndex.value && active.value === i)
        cls.push('is-stake');
    return cls;
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage" },
    ...{ style: {} },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-head" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sh-title" },
});
__VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.stage.head) }, null, null);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "sh-actors" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "a-p" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "dot" },
});
(__VLS_ctx.stage.proposerName);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "a-c" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "dot" },
});
(__VLS_ctx.stage.criticName);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-columns" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-col left" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "col-head" },
});
(__VLS_ctx.stage.proposerCol);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "col-name" },
});
(__VLS_ctx.stage.proposerName);
for (const [d, i] of __VLS_getVForSourceType((__VLS_ctx.stage.rows))) {
    ('p' + i);
    if (d.side === 'p') {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "msg m-p" },
            ...{ style: (__VLS_ctx.bubbleStyle(i)) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "msg-label" },
        });
        (d.r);
        (d.label);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        __VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (d.html) }, null, null);
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-spine" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "spine-line" },
    ...{ style: {} },
});
for (const [d, i] of __VLS_getVForSourceType((__VLS_ctx.stage.rows))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: ('s' + i),
        ...{ class: (__VLS_ctx.spineClass(i, d.side)) },
    });
    (d.r);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
    (d.side === 'p' ? 'PROP' : 'CRIT');
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-col right" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "col-head" },
});
(__VLS_ctx.stage.criticCol);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "col-name" },
});
(__VLS_ctx.stage.criticName);
for (const [d, i] of __VLS_getVForSourceType((__VLS_ctx.stage.rows))) {
    ('c' + i);
    if (d.side === 'c') {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "msg m-c" },
            ...{ style: (__VLS_ctx.bubbleStyle(i)) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "msg-label" },
        });
        (d.r);
        (d.label);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        __VLS_asFunctionalDirective(__VLS_directives.vHtml)(null, { ...__VLS_directiveBindingRestFields, value: (d.html) }, null, null);
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ style: {} },
        });
    }
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stage-verdict" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "vk" },
});
(__VLS_ctx.stage.verdictKey);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "vt" },
});
(__VLS_ctx.stage.verdictText);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "vr" },
});
(__VLS_ctx.stage.verdictRight);
/** @type {__VLS_StyleScopedClasses['stage']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-head']} */ ;
/** @type {__VLS_StyleScopedClasses['sh-title']} */ ;
/** @type {__VLS_StyleScopedClasses['sh-actors']} */ ;
/** @type {__VLS_StyleScopedClasses['a-p']} */ ;
/** @type {__VLS_StyleScopedClasses['dot']} */ ;
/** @type {__VLS_StyleScopedClasses['a-c']} */ ;
/** @type {__VLS_StyleScopedClasses['dot']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-columns']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-col']} */ ;
/** @type {__VLS_StyleScopedClasses['left']} */ ;
/** @type {__VLS_StyleScopedClasses['col-head']} */ ;
/** @type {__VLS_StyleScopedClasses['col-name']} */ ;
/** @type {__VLS_StyleScopedClasses['msg']} */ ;
/** @type {__VLS_StyleScopedClasses['m-p']} */ ;
/** @type {__VLS_StyleScopedClasses['msg-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-spine']} */ ;
/** @type {__VLS_StyleScopedClasses['spine-line']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-col']} */ ;
/** @type {__VLS_StyleScopedClasses['right']} */ ;
/** @type {__VLS_StyleScopedClasses['col-head']} */ ;
/** @type {__VLS_StyleScopedClasses['col-name']} */ ;
/** @type {__VLS_StyleScopedClasses['msg']} */ ;
/** @type {__VLS_StyleScopedClasses['m-c']} */ ;
/** @type {__VLS_StyleScopedClasses['msg-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stage-verdict']} */ ;
/** @type {__VLS_StyleScopedClasses['vk']} */ ;
/** @type {__VLS_StyleScopedClasses['vt']} */ ;
/** @type {__VLS_StyleScopedClasses['vr']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            stage: stage,
            bubbleStyle: bubbleStyle,
            spineClass: spineClass,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
