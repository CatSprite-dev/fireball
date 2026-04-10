import { defineStore } from 'pinia'
import { tokenToString } from 'typescript'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
    const isLoggedIn = ref(false)
    const isReady = ref(false)

    async function checkAuth() {
        const response = await fetch('/api/ping', { method: 'GET' })
        isLoggedIn.value = response.ok
        isReady.value = true
    }

    async function logout() {
        const response = await fetch('/api/logout', { method: 'POST' })
        isLoggedIn.value = false
    }

    return { isLoggedIn, isReady, checkAuth, logout }
})