import SiteNav from '../components/SiteNav.vue';
import SiteFooter from '../components/SiteFooter.vue';
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-shell agon-page v-dialectic" },
});
/** @type {[typeof SiteNav, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(SiteNav, new SiteNav({}));
const __VLS_1 = __VLS_0({}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.main, __VLS_intrinsicElements.main)({
    ...{ class: "page-content" },
});
var __VLS_3 = {};
/** @type {[typeof SiteFooter, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(SiteFooter, new SiteFooter({}));
const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
/** @type {__VLS_StyleScopedClasses['page-shell']} */ ;
/** @type {__VLS_StyleScopedClasses['agon-page']} */ ;
/** @type {__VLS_StyleScopedClasses['v-dialectic']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
// @ts-ignore
var __VLS_4 = __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            SiteNav: SiteNav,
            SiteFooter: SiteFooter,
        };
    },
});
const __VLS_component = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
export default {};
; /* PartiallyEnd: #4569/main.vue */
