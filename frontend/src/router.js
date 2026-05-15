export const routes = [
    { path: '/', component: () => import('./views/LandingPage.vue') },
    { path: '/:pathMatch(.*)*', component: () => import('./views/NotFoundPage.vue') },
];
