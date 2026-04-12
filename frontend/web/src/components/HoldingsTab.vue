<script setup lang="ts">
import type { Investment } from '../types'
import { computed } from 'vue'
import { useFormatters } from '../composables/useFormatters'

const { formatCurrency } = useFormatters()

const props = defineProps<{
    investments: Investment[]
}>()

const rows = computed(() => 
    props.investments.map(inv => {
        const invested = inv.quantity * inv.purchasePrice
        const current = inv.quantity * inv.currentPrice
        const gain = current - invested
        const gainPercent = invested > 0 ? (gain / invested) * 100 : 0
        const dividendsPercent = invested > 0 ? (inv.dividends / invested) * 100 : 0
        return { ...inv, invested, current, gain, gainPercent, dividendsPercent }
    })
)
</script>

<template>
    <div class="card">
        <div class="card-header">
            <h2>Your holdings</h2>
            <p>All positions from your T-Invest portfolio</p>
        </div>

        <div class="table-wrapper">
            <table>
                <thead>
                    <tr>
                        <th class="left">Name</th>
                        <th class="left">Ticker</th>
                        <th class="left">Type</th>
                        <th class="right">Qty</th>
                        <th class="right">Avg Price</th>
                        <th class="right">Current</th>
                        <th class="right">Value</th>
                        <th class="right">Gain/Loss</th>
                        <th class="center">Dividends</th>
                        <th class="center">Total</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="rows.length === 0">
                        <td colspan="9" class="empty">No positions found</td>
                    </tr>
                    <tr v-for="row in rows" :key="row.id">
                        <td class="left name">{{ row.name }}</td>
                        <td class="left ticker">{{ row.ticker }}</td>
                        <td class="left type">{{ row.type }}</td>
                        <td class="right">{{ row.quantity }}</td>
                        <td class="right">{{ formatCurrency(row.purchasePrice) }}</td>
                        <td class="right">{{ formatCurrency(row.currentPrice) }}</td>
                        <td class="right">{{ formatCurrency(row.current) }}</td>
                        <td class="right" :class="row.gain >= 0 ? 'positive' : 'negative'">
                            {{ row.gain >= 0 ? '+' : '' }}{{ formatCurrency(row.gain) }}
                            <span class="percent">
                                ({{ row.gainPercent >= 0 ? '+' : '' }}{{ row.gainPercent.toFixed(2) }}%)
                            </span>
                        </td>
                        <td class="center" :class="row.dividends >= 0 ? 'positive' : ''">
                            <template v-if="row.dividends === 0">-</template>
                            <template v-else>
                                {{ formatCurrency(row.dividends) }}
                                <span class="percent">
                                    (+{{ row.dividendsPercent.toFixed(2) }}%)
                                </span>
                            </template>
                        </td>
                        <td class="center" :class="row.totalYield >= 0 ? 'positive' : 'negative'">
                            <template v-if="row.totalYield === 0">-</template>
                            <template v-else>
                                {{ formatCurrency(row.totalYield) }}
                                <span class="percent">
                                ({{ row.totalYieldRelative >= 0 ? '+' : '' }}{{ row.totalYieldRelative.toFixed(2) }}%)
                            </span>
                            </template>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<style scoped>
.card {
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--card);
    color: var(--card-foreground);
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

.table-wrapper {
    padding: 16px 0 0;
    overflow-x: auto;
}

table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
}

thead tr {
    border-bottom: 1 px solid var(--border);
}

th {
    padding: 10px 16px;
    font-weight: 500;
    color: var(--muted-foreground);
    white-space: nowrap;
}

td {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
}

tr:last-child td {
    border-bottom: none;
}

tr:hover td {
    background: var(--muted);
}

.left { text-align: left; }
.right { text-align: right; }
.center { text-align: center; }

.name {
    font-weight: 500;
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
}

.ticker {
    font-family: monospace;
    color: var(--muted-foreground);
}

.type {
    text-transform: capitalize;
    color: var(--muted-foreground);
}

.percent {
    font-size: 0.75rem;
    opacity: 0.8;
}

.empty {
    text-align: center;
    padding: 48px;
    color: var(--muted-foreground);
}

.positive { color: var(--green); }
.negative { color: var(--red); }
</style>