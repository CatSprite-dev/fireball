<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { usePortfolioStore } from '../stores/portfolio'

const router = useRouter()
const auth = useAuthStore()

const token = ref('')
const error = ref('')
const isLoading = ref(false)

async function submit() {
    if (!token.value.trim()) {
        error.value = 'Token is required'
        return
    }

    isLoading.value = true
    error.value = ''

    try {
        const response = await fetch('/portfolio', {
            method: 'POST',
            headers: {
                'T-Token': token.value.trim(),
                'Content-Type': 'application/json',
            },
        })

        if (response.status === 401) {
            error.value = 'Token is invalid or expired'
            return
        }

        if (!response.ok) {
            error.value = 'Something went wrong, try again'
            return
        }

        const portfolio = usePortfolioStore()
        auth.setToken(token.value.trim())
        await portfolio.load()
        router.push('/')
    } finally {
        isLoading.value = false
    }
}
</script>

<template>
    <div class="login-page">
        <div class="login-card">
            <div class="login-header">
                <h1>Welcome to Investment Fireball</h1>
                <p>Enter your API token to access your portfolio</p>
            </div>
            
            <div class="login-body">
                <label>Tinkoff API Token</label>
                <input
                    v-model="token"
                    type="password"
                    placeholder="Enter your token"
                    :disabled="isLoading"
                    @keyup.enter="submit"
                />
                <span class="hint">
                    Get your token from
                    <a href="https://www.tinkoff.ru/invest/settings/" target="_blank">
                        Tinkoff Invest
                    </a>
                </span>

                <p v-if="error" class="error">{{ error }}</p>

                <button :disabled="isLoading" @click="submit">
                    {{ isLoading ? 'Connecting...' : 'Access Portfolio' }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.login-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--background);
}

.login-card {
    width: 100%;
    max-width: 400px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--card);
    color: var(--card-background);
    box-shadow: 0 1px 3px rgb(0 0 0 / 0.1);
}

.login-header {
    padding: 24px;
    padding-bottom: 12px;
    border-bottom: 1px solid var (--border);

}

.login-header h1 {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0 0 4px;
}

.login-header p {
    font-size: 0.875rem;
    color: var(--muted-foreground);
    margin: 0;
}

.login-body {
    padding: 24px;
    padding-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

label {
    font-size: 0.875rem;
    font-weight: 500;
}

input {
    height: 40px;
    width: 100%;
    border: 1 px solid var(--input);
    border-radius: 6px;
    background: var(--input);
    padding: 0 12px;
    font-size: 0.875rem;
    color: var(--foreground);
    box-sizing: border-box;
}

input:focus {
    outline: none;
    box-shadow: 0 0 0 2px var(--ring);
}

input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.hint {
    font-size: 0.75rem;
    color: var(--muted-foreground);
}

.hint a {
    color: var(--primary);
    text-decoration: none;
}

.hint a:hover {
    text-decoration: underline;
}


.error {
    font-size: 0.875rem;
    color: #dc2626;
    margin: 0;
}

button {
    height: 40px;
    border: none;
    border-radius: 6px;
    background: var(--primary);
    color: var(--primary-foreground);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    margin-top: 8px;
}

button:hover:not(:disabled) {
    opacity: 0.9;
}

button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
</style>