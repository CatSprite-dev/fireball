<script setup lang="ts">
import type { Investment, InvestmentWithGain } from '../types'
import { computed } from 'vue'
import { useFormatters } from '../composables/useFormatters';

const { formatCurrency } = useFormatters()

const props = defineProps<{
    investments: Investment[]
}>()

const typeLabels: Record<string, string> = {
    share: 'Stocks',
    etf: 'ETFs',
    bond: 'Bonds',
    futures: 'Futures',
    crypto: 'Crypto',
    currency: 'Currency',
}

function labelForType(type: string): string {
    return typeLabels[type] ?? type.charAt(0).toUpperCase() + type.slice(1)
}

const allocation = computed(() => {
    const totals: Record<string, number> = {}
    let grand = 0

    for (const inv of props.investments) {
        const value = inv.quantity * inv.currentPrice
        totals[inv.type] = (totals[inv.type] ?? 0) + value
        grand += value
    }

    if (grand === 0) return []

    return Object.entries(totals)
        .sort((a,b) => b[1] - a[1])
        .map(([type, value]) => ({
            type,
            label: labelForType(type),
            percent: (value / grand) * 100
    }))
})

const topPerformers = computed<InvestmentWithGain[]>(() => {
    return props.investments
        .map(inv => {
            const invested = inv.quantity * inv.purchasePrice
            const current = inv.quantity * inv.currentPrice
            const gain = current - invested
            const gainPercent = invested > 0 ? (gain / invested) * 100 : 0
            return { ...inv, gain, gainPercent }
        })
        .sort((a,b) => b.gainPercent - a.gainPercent)
        .slice(0,5)
})
</script>

<template>
    <div class="overview">

        <div class="card">
            <div class="card-header">
                <h2>Asset Allocation</h2>
                <p>By investment type</p>
            </div>
            <div class="card-body">
                <div v-if="allocation.length === 0" class="empty">
                    No data
                </div>
                <div v-else class="allocation-list">
                    <div v-for="item in allocation" :key="item.type" class="allocation-row">
                        <div class="allocation-label">
                            <span>{{ item.label }}</span>
                            <span class="percent-text">{{ item.percent.toFixed(1) }}%</span>
                        </div>
                        <div class="bar-track">
                            <div class="bar-fill" :style="{ width: item.percent + '%' }"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="card">
            <div class="card-header">
                <h2>Top Performers</h2>
                <p>Best returns</p>
            </div>
            <div class="card-body">
                <div v-if="topPerformers.length === 0" class="empty">
                    No data
                </div>
                <div v-else class="performers-list">
                    <div v-for="inv in topPerformers" :key="inv.id" class="performer-row">
                        <div class="performer-name">
                            <span class="name">{{ inv.name }}</span>
                            <span class="ticker">{{ inv.ticker }}</span>
                        </div>
                        <div class="performer-gain">
                            <span :class="inv.gainPercent >= 0 ? 'positive' : 'negative'">
                                {{ inv.gainPercent >= 0 ? '+' : '' }}{{ inv.gainPercent.toFixed(2) }}
                            </span>
                            <span class="gain-abs" :class="inv.gain >= 0 ? 'positive' : 'negative'">
                                {{ inv.gain >= 0 ? '+' : ''}}{{ formatCurrency(inv.gain) }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

    </div>
</template>

<style scoped>
.overview {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
}

@media (max-width: 700px) {
    .overview {
        grid-template-columns: 1fr;
    }
}

.card {
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--card);
    color: var(--card-foreground)
}

.card-header {
    padding: 20px 20px 0;
}

.card-header h2 {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0 0 2px;
}

.card-header p {
    font-size: 0.875rem;
    color: var(--muted-foreground);
    margin: 0;
}

.card-body {
    padding: 16px 20px 20px;
}

.empty {
    font-size: 0.875rem;
    color: var(--muted-foreground);
}

.allocation-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.allocation-label {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
    font-weight: 500;
    margin-bottom: 4px;
}

.percent-text {
    color: var(--muted-foreground);
}

.bar-track {
    width: 100%;
    height: 8px;
    background: var(--muted);
    border-radius: 999px;
}

.bar-fill {
    height: 100%;
    background: var(--primary);
    border-radius: 999px;
    transition: width 0.3s ease;
}

.performers-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.performer-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.performer-name {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.name {
    font-size: 0.875rem;
    font-weight: 500;
}

.ticker {
    font-size: 0.75rem;
    columns: var(--muted-foreground);
}

.performer-gain {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 2px;
}

.gain-abs {
    font-size: 0.75rem;
}

.positive { color: #059669; }
.negative { color: #dc2626; }
</style>