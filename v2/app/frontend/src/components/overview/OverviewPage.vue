<script setup lang="ts">
import {ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Performance} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import SummaryCards from './SummaryCards.vue'
import PerformanceCards from './PerformanceCards.vue'
import PortfolioCharts from './PortfolioCharts.vue'
import PositionsTable from './PositionsTable.vue'
import MoreMetrics from './MoreMetrics.vue'

// Счета и период выбираются в шапке. null — счета ещё не загружены, выбора пока нет.
const props = defineProps<{accountIDs: string[] | null; period: string}>()
const emit = defineEmits<{asOf: [day: string]}>()

// Сравниваем пока только с MCFTR.
const index = 'MCFTR'

const performance = ref<Performance | null>(null)
const performanceError = ref('')
const more = ref(false)

// Ответы могут прийти не по порядку, если выбор меняют быстро: показываем только последний запрос.
let request = 0

watch([() => props.accountIDs, () => props.period], async ([ids, p]) => {
  if (ids === null) return
  const current = ++request
  performance.value = null
  performanceError.value = ''
  try {
    const result = await PortfolioService.Performance(ids, p, index)
    if (current === request) performance.value = result
  } catch (e) {
    if (current === request) performanceError.value = e instanceof Error ? e.message : String(e)
  }
}, {immediate: true})
</script>

<template>
  <div class="page">
    <SummaryCards :accountIDs="accountIDs" :period="period" @as-of="d => emit('asOf', d)" />

    <section class="chart-block" aria-label="График">
      <PortfolioCharts :accountIDs="accountIDs" :period="period" :index="index" :performance="performance" />
      <PerformanceCards :index="index" :performance="performance" :error="performanceError" @more="more = true" />
    </section>

    <PositionsTable :accountIDs="accountIDs" :period="period" />

    <MoreMetrics v-if="more" :performance="performance" @close="more = false" />
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.chart-block {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
  border: 1px solid var(--line);
}
</style>
