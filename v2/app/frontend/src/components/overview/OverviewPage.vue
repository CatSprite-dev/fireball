<script setup lang="ts">
import {ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Charts, Overview, Performance} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import SummaryCards from './SummaryCards.vue'
import PerformanceCards from './PerformanceCards.vue'
import PortfolioCharts from './PortfolioCharts.vue'
import type {ChartMode} from './chartMode'
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

const overview = ref<Overview | null>(null)
const overviewError = ref('')

// Режим графика: от него зависит, что показывает панель справа.
const mode = ref<ChartMode>('shadowTrades')
const charts = ref<Charts | null>(null)
const chartsError = ref('')

// Ответы могут прийти не по порядку, если выбор меняют быстро: показываем только последний запрос.
let request = 0
let chartsRequest = 0
let overviewRequest = 0

watch([() => props.accountIDs, () => props.period], async ([ids, p]) => {
  if (ids === null) return
  const current = ++overviewRequest
  overview.value = null
  overviewError.value = ''
  try {
    const result = await PortfolioService.Overview(ids, p)
    if (current === overviewRequest) {
      overview.value = result
      emit('asOf', result.asOf)
    }
  } catch (e) {
    if (current === overviewRequest) overviewError.value = e instanceof Error ? e.message : String(e)
  }
}, {immediate: true})

watch([() => props.accountIDs, () => props.period], async ([ids, p]) => {
  if (ids === null) return
  const current = ++chartsRequest
  charts.value = null
  chartsError.value = ''
  try {
    const result = await PortfolioService.Charts(ids, p, index)
    if (current === chartsRequest) charts.value = result
  } catch (e) {
    if (current === chartsRequest) chartsError.value = e instanceof Error ? e.message : String(e)
  }
}, {immediate: true})

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
    <SummaryCards :overview="overview" :error="overviewError" :period="period" />

    <section class="chart-block" aria-label="График">
      <PortfolioCharts v-model:mode="mode" :charts="charts" :error="chartsError" :index="index" :performance="performance" />
      <PerformanceCards
        :mode="mode"
        :charts="charts"
        :overview="overview"
        :index="index"
        :performance="performance"
        :error="performanceError"
        @more="more = true"
      />
    </section>

    <PositionsTable :accountIDs="accountIDs" :period="period" />

    <MoreMetrics v-if="more" :performance="performance" @close="more = false" />
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.chart-block {
  display: grid;
  grid-template-columns: minmax(0, 2.05fr) minmax(0, 1fr);
  gap: 1px;
  background: var(--line);
  border: 1px solid var(--line);
}
</style>
