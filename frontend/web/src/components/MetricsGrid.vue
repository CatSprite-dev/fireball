<script setup lang="ts">
import type { Metrics } from '../types'
import { useFormatters } from '../composables/useFormatters';

const { formatCurrency } = useFormatters()

defineProps<{
    metrics: Metrics
}>()
</script>

<template>
    <div class="metrics-grid">

        <div class="card">
            <div class="card-header">
                <span>Total Invested</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1"/>
                    <path d="M3 5v14a2 2 0 0 0 2 2h15a1 1 0 0 0 1-1v-4"/>
                </svg>
            </div>
            <div class="card-value">{{ formatCurrency(metrics.totalInvested) }}</div>
            <div class="card-sub">
                Across {{ metrics.portfolioSize }}
                {{ metrics.portfolioSize === 1 ? 'position' : 'positions' }}
            </div>
        </div>

        <div class="card">
            <div class="card-header">
                <span>Current Value</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="12" x2="12" y1="2" y2="22"/>
                    <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
                </svg>
            </div>
            <div class="card-value">{{ formatCurrency(metrics.currentValue) }}</div>
            <div class="card-sub">Market value of holdings</div>
        </div>

        <div class="card">
            <div class="card-header">
                <span>Total Gain/Loss</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="22 7 13.5 15.5 8.5 10.5 2 17"/>
                    <polyline points="16 7 22 7 22 13"/>
                </svg>
            </div>
            <div class="card-value" :class="metrics.totalGain >= 0 ? 'positive' : 'negative'">
                {{ metrics.totalGain >= 0 ? '+' : '' }}{{ formatCurrency(metrics.totalGain) }}
            </div>
            <div class="card-sub" :class="metrics.totalGainPercent >= 0 ? 'positive' : 'negative'">
                {{ metrics.totalGainPercent >= 0 ? '+' : '' }}{{  metrics.totalGainPercent.toFixed(2) }}% return
            </div>
        </div>

        <div class="card">
            <div class="card-header">
                <span>Daily Yield</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21.21 15.89A10 10 0 1 1 8 2.83"/>
                    <path d="M22 12A10 10 0 0 0 12 2v10z"/>
                </svg>
            </div>
            <div class="card-value" :class="metrics.dailyYield >= 0 ? 'positive' : 'negative'">
                {{ metrics.dailyYield >= 0 ? '+' : '' }}{{ formatCurrency(metrics.dailyYield) }}
            </div>
            <div class="card-sub" :class="metrics.dailyYieldRelative >= 0 ? 'positive' : 'negative'">
                {{ metrics.dailyYieldRelative >= 0 ? '+' : '' }}{{  metrics.dailyYieldRelative.toFixed(2) }}% return
            </div>
        </div>

    </div>
</template>

<style scoped>
.metrics-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
}

@media (max-width: 900px) {
    .metrics-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 500px) {
    .metrics-grid {
        grid-template-columns: 1fr;
    }
}

.card {
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--card);
    color: var(--card-foreground);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--muted-foreground);
    margin-bottom: 4px;
}

.card-header svg {
    color: var(--muted-foreground);
}

.card-value {
    font-size: 1.5rem;
    font-weight: 700;
}

.card-sub {
    font-size: 0.75rem;
    color: var(--muted-foreground);
}

.positive {
    color: #059669;
}

.negative {
    color: #dc2626;
}
</style>