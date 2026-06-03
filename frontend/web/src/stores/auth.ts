import { defineStore } from 'pinia'
import { ref } from 'vue'
import { usePortfolioStore } from './portfolio'
import { useChartStore } from './chart'

export const useAuthStore = defineStore('auth', () => {
    const isLoggedIn = ref(false)
    const isReady = ref(false)
    const error = ref('')
    const isDemo = ref(false)

    let authPromise: Promise<void> | null = null

    async function checkAuth() {
        if (authPromise) return authPromise
        authPromise = fetch('/api/ping', { method: 'GET' })
            .then(response => {
                isLoggedIn.value = response.ok
                isReady.value = true
            })
            .finally(() => {
                authPromise = null
            })
    }

    async function logout() {
        if (!isDemo.value) {
            const response = await fetch('/api/logout', { method: 'POST' })
            if (!response.ok) {
                console.warn('Logout request failed, clearing local state anyway')
            }
        }
        isLoggedIn.value = false
        isDemo.value = false

        usePortfolioStore().reset()
        useChartStore().reset()
    }

    return { isLoggedIn, isReady, isDemo, checkAuth, logout, error }
})