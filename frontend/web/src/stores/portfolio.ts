import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { useAuthStore } from "./auth";
import { fetchPortfolio } from "../api/auth";
import type { UserFullPortfolio, Investment, Metrics } from "../types";

function parseMoney(m: { units: string; nano: number } | undefined): number {
    if (!m) return 0;
    return parseFloat(m.units || '0') + (m.nano || 0) / 1e9;
}

export const usePortfolioStore = defineStore('portfolio', () => {
    const raw = ref<UserFullPortfolio | null>(null)
    const isLoading = ref(false)
    const error = ref('')

    const investments = computed<Investment[]>(() => {
        if (!raw.value) return []

        return raw.value.positions
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
        if (!raw.value) return {
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
        );
        const currentValue = parseMoney(raw.value.totalAmountPortfolio)
        const totalGain = parseMoney(raw.value.expectedYield)
        const totalGainPercent = totalInvested > 0 ? (totalGain / totalInvested) * 100 : 0
        const dailyYield = parseMoney(raw.value.dailyYield)
        const dailyYieldRelative = parseMoney(raw.value.dailyYieldRelative)

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
            raw.value = await fetchPortfolio(auth.token)
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

    return { raw, investments, metrics, isLoading, error, load }
})