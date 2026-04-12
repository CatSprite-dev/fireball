import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/:pathMatch(.*)*',
            redirect: '/',
        },
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

router.beforeEach(async (to) => {
    const auth = useAuthStore()

    if (!auth.isReady) {
        await auth.checkAuth()
    }

    if (to.meta.requiresAuth && !auth.isLoggedIn) {
        return '/login'
    }

    if (to.path === '/login' && auth.isLoggedIn) {
        return '/'
    }
})

export default router