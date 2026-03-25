import { defineStore } from 'pinia'
import { tokenToString } from 'typescript'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
    const token = ref(localStorage.getItem('token') ?? '')

    const isLoggedIn = computed(() => !!token.value)

    function setToken(newToken: string) {
        token.value = newToken
        localStorage.setItem('token', newToken)
    }

    function logout() {
        token.value = ''
        localStorage.removeItem('token')
    }

    return { token, isLoggedIn, setToken, logout }
})