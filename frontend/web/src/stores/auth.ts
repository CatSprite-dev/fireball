import { defineStore } from 'pinia'
import { tokenToString } from 'typescript'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
    const isLoggedIn = ref(false)

    async function checkAuth() {
        const response = await fetch('/ping')
        isLoggedIn.value = response.ok
    }

    async function logout() {
        const response = await fetch('/logout', { method: 'POST' })
        isLoggedIn.value = false
    }

    return { isLoggedIn, checkAuth, logout }
})