import { createApp } from 'vue';
import {createRouter, createWebHistory} from 'vue-router';
import {createPinia} from 'pinia';
import App from './App.vue';
import { i18n } from './i18n.config.js';

const pinia = createPinia();

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'Home',
            component: () => import(/* webpackChunkName: "Home" */ "./pages/Index.vue"),
        },{
            path: '/:name',
            name: 'Constructor',
            component: () => import(/* webpackChunkName: "Home" */ "./pages/constructor/index.vue"),
        },
    ]
});

createApp(App)
    .use(pinia)
    .use(router)
    .use(i18n)
    .mount('#app')
