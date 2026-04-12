import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router/index'
import App from './App.vue'
import { useAuthStore } from './stores/auth'
import VueApexCharts from 'vue3-apexcharts'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(VueApexCharts)

app.mount('#app')