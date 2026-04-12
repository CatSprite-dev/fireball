import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { fetchPortfolio } from '../api/fetch_portfolio'
import { parseMoney} from '../composables/useFormatters'
import type { UserFullPortfolio, Investment, Metrics } from '../types'

export const usePortfolioStore = defineStore('portfolio', () => {
    const portfolio = ref<UserFullPortfolio | null>(null)
    const isLoading = ref(false)
    const error = ref('')

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
                totalYield: parseMoney(pos.totalYield),
                totalYieldRelative: parseMoney(pos.totalYieldRelative),
            }))
    })

    const metrics = computed<Metrics>(() => {
        if (!portfolio.value) return {
            totalInvestedOfHoldings: 0,
            totalInvested: 0,
            currentValue: 0,
            totalGain: 0,
            totalGainPercent: 0,
            dailyYield: 0,
            dailyYieldRelative: 0,
            portfolioSize: 0,
        }

        const totalInvestedOfHoldings = investments.value.reduce(
            (sum, inv) => sum + inv.quantity * inv.purchasePrice, 0
        )
        const totalInvested = parseMoney(portfolio.value.totalInvested)
        const currentValue = parseMoney(portfolio.value.totalAmountPortfolio)
        const totalGain = parseMoney(portfolio.value.totalReturn)
        const totalGainPercent = parseMoney(portfolio.value.totalReturnRelative)
        const dailyYield = parseMoney(portfolio.value.dailyYield)
        const dailyYieldRelative = parseMoney(portfolio.value.dailyYieldRelative)

        return {
            totalInvestedOfHoldings,
            totalInvested,
            currentValue,
            totalGain,
            totalGainPercent: totalInvested > 0 ? (totalGain / totalInvested) * 100 : 0,
            dailyYield: parseMoney(portfolio.value.dailyYield),
            dailyYieldRelative: parseMoney(portfolio.value.dailyYieldRelative),
            portfolioSize: investments.value.length,
        }
    })

    async function load() {
        const auth = useAuthStore()
        if (!auth.token) return
        isLoading.value = true
        error.value = ''
        try {
            portfolio.value = await fetchPortfolio(auth.token)
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

    return { raw: portfolio, investments, metrics, isLoading, error, load }
})