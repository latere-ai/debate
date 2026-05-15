import type { RouteRecordRaw } from 'vue-router';

export const routes: RouteRecordRaw[] = [
  { path: '/', component: () => import('./views/LandingPage.vue') },
  { path: '/:pathMatch(.*)*', component: () => import('./views/NotFoundPage.vue') },
];
