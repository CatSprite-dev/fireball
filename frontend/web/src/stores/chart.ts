import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { usePortfolioStore } from './portfolio'
import { fetchChart } from '../api/fetch_chart'
import type { ChartData } from '../types'
import { parseMoney } from '../composables/useFormatters' // Reuse the money parsing function


export const useChartStore = defineStore('chart', () => {
    const chartData = ref<ChartData | null>(null)
    const isLoading = ref(false)
    const error = ref('')

    async function load(period: string = '6m', index: string = 'IMOEX') {
        const auth = useAuthStore()
        const portfolioStore = usePortfolioStore()
        if (!auth.isLoggedIn || !portfolioStore.raw) return

        isLoading.value = true
        error.value = ''
        try {
            chartData.value = await fetchChart(period, index)
        } catch (e) {
            if (e instanceof Error && e.message === 'UNAUTHORIZED') {
                auth.logout()
                error.value = 'Session expired, please log in again'
            } else {
                error.value = 'Failed to load chart'
            }
        } finally {
            isLoading.value = false
        }
    }

    const chartSeries = computed(() => {
    if (!chartData.value) return []

    return [    
        { name: 'Portfolio', 
            data: chartData.value.times
            .map((time, i) => ({
                x: time,
                y: parseMoney(chartData.value?.portfolio[i])
                    }))},
        { name: 'Index',   
            data: chartData.value.times
            .map((time, i) => ({
                x: time,
                y: parseMoney(chartData.value?.index[i])
                    }))},
        ]
    })

    return { chartSeries, isLoading, error, load }
})