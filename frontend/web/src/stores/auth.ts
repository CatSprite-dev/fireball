import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
    const isLoggedIn = ref(false)
    const isReady = ref(false)
    const error = ref('')

    async function checkAuth() {
        const response = await fetch('/api/ping', { method: 'GET' })
        isLoggedIn.value = response.ok
        isReady.value = true
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