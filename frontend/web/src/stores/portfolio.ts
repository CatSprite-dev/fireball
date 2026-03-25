import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { fetchPortfolio } from '../api/auth'
import type { UserFullPortfolio, Investment, Metrics, ChartData } from '../types'
import { Y } from 'vue-router/dist/index-BzEKChPW.js'

function parseMoney(m: { units: string; nano: number } | undefined): number {
    if (!m) return 0
    const units = parseFloat(m.units || '0') || 0
    const nanos = (m.nano || 0) / 1e9
    return units + nanos
}

export const usePortfolioStore = defineStore('portfolio', () => {
    const portfolio = ref<UserFullPortfolio | null>(null)
    const chartData = ref<ChartData | null>(null)
    const isLoading = ref(false)
    const error = ref('')

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

    const investments = computed<Investment[]>(() => {
        if (!portfolio.value) return []

        return portfolio.value.positions
            .filter(pos => pos.instrumentType?.toLowerCase() !== 'currency')
            .map(pos => ({
                id: pos.positionUid,
                name: pos.name,
                ticker: pos.ticker,
                type: pos.instrumentType?.toLowerCase() ?? 'other',
                quantity: parseMoney(pos.quantity),
                purchasePrice: parseMoney(pos.averagePositionPrice),
                currentPrice: parseMoney(pos.currentPrice),
                dividends: parseMoney(pos.dividends),
            }))
    })
    
    const metrics = computed<Metrics>(() => {
        if (!portfolio.value) return {
            totalInvested: 0,
            currentValue: 0,
            totalGain: 0,
            totalGainPercent: 0,
            dailyYield: 0,
            dailyYieldRelative: 0,
            portfolioSize: 0,
        }

        const totalInvested = investments.value.reduce(
            (sum, inv) => sum + inv.quantity * inv.purchasePrice, 0
        )
        const currentValue = parseMoney(portfolio.value.totalAmountPortfolio)
        const totalGain = parseMoney(portfolio.value.expectedYield)
        const totalGainPercent = totalInvested > 0 ? (totalGain / totalInvested) * 100 : 0
        const dailyYield = parseMoney(portfolio.value.dailyYield)
        const dailyYieldRelative = parseMoney(portfolio.value.dailyYieldRelative)

        return {
            totalInvested,
            currentValue,
            totalGain,
            totalGainPercent,
            dailyYield,
            dailyYieldRelative,
            portfolioSize: investments.value.length,
        }
    })

    async function load() {
        const auth = useAuthStore()
        if (!auth.token) return

        isLoading.value = true
        error.value = ''

        try {
            const result = await fetchPortfolio(auth.token)
            portfolio.value = result.user_portfolio
            chartData.value = result.chart_data
        } catch (e) {
            if (e instanceof Error && e.message === 'UNAUTHORIZED') {
                auth.logout()
                error.value = 'Session expired, please log in again'
            } else {
                error.value = 'Failed to load portfolio'
            }
        } finally {
            isLoading.value = false
        }
    }

    return { raw: portfolio, chartSeries, investments, metrics, isLoading, error, load }
})