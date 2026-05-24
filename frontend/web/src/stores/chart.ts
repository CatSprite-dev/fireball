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
    const ctrl = ref<AbortController | null>(null)

    async function load(force: boolean = false, period: string = '1y', index: string = 'IMOEX') {
        console.log('in load chart')
        const auth = useAuthStore()
        if (!auth.isLoggedIn) return

        isLoading.value = true
        error.value = ''

        ctrl.value?.abort()
        ctrl.value = new AbortController()

        try {
            console.log('fetching chart data')
            chartData.value = await fetchChart(force, ctrl.value, period, index)
        } catch (e) {
            if (e instanceof DOMException && e.name === 'AbortError') {
                return
            } else if (e instanceof Error && e.message === 'UNAUTHORIZED') {
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
                y: parseMoney(chartData.value?.benchmark[i])
                    }))},
        ]
    })

    function abort() {
        ctrl.value?.abort()
    }

    return { chartSeries, isLoading, error, load, abort }
})