import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router/index'
import App from './App.vue'
import { useAuthStore } from './stores/auth'

const app = createApp(App)
app.use(createPinia())

const auth = useAuthStore()
await auth.checkAuth()

app.use(router)
app.mount('#app')