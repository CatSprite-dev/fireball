import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
    const isLoggedIn = ref(false)
    const isReady = ref(false)
    const error = ref('')

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
        const response = await fetch('/api/logout', { method: 'POST' })
        if (!response.ok) {
            console.warn('Logout request failed, clearing local state anyway')
        }
        isLoggedIn.value = false
    }

    return { isLoggedIn, isReady, checkAuth, logout, error }
})