import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/login',
            component: () => import('../views/LoginView.vue'),
        },
        {
            path: '/',
            component: () => import('../views/PortfolioView.vue'),
            meta: { requiresAuth: true },
        },
    ],
})

router.beforeEach((to) => {
    const auth = useAuthStore()

    if (to.meta.requiresAuth && !auth.isLoggedIn) {
        return '/login'
    }

    if (to.path === '/login' && auth.isLoggedIn) {
        return '/portfolio'
    }
})

export default router