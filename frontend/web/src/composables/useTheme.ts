import { ref, watch, onMounted } from 'vue'

const isDark = ref(false)

export function useTheme() {
    function applyTheme(dark: boolean) {
        document.documentElement.classList.toggle('dark', dark)
        localStorage.setItem('theme', dark ? 'dark': 'light')
    }

    function toggleTheme() {
        isDark.value = !isDark.value
    }

    function initTheme() {
        const saved = localStorage.getItem('theme')
        const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches

        isDark.value = saved ? saved === 'dark' : systemDark
    }

    watch(isDark, (dark) => {
        applyTheme(dark)
    })

    onMounted(() => {
        initTheme()

        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem('theme')) {
                isDark.value = e.matches
            }
        })
    })

    return { isDark, toggleTheme }
}