<script setup lang="ts">
import {ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Performance} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import AccountsFilter from './AccountsFilter.vue'
import PeriodSwitch from './PeriodSwitch.vue'
import SummaryCards from './SummaryCards.vue'
import PerformanceCards from './PerformanceCards.vue'
import PositionsTable from './PositionsTable.vue'

// null — счета ещё не загружены, выбора пока нет.
const accountIDs = ref<string[] | null>(null)
const period = ref('12m')
const index = ref('MCFTR')

const performance = ref<Performance | null>(null)
const performanceError = ref('')

// Ответы могут прийти не по порядку, если выбор меняют быстро: показываем только последний запрос.
let request = 0

watch([accountIDs, period, index], async ([ids, p, i]) => {
  if (ids === null) return
  const current = ++request
  performance.value = null
  performanceError.value = ''
  try {
    const result = await PortfolioService.Performance(ids, p, i)
    if (current === request) performance.value = result
  } catch (e) {
    if (current === request) performanceError.value = e instanceof Error ? e.message : String(e)
  }
})
</script>

<template>
  <div class="page">
    <AccountsFilter @change="ids => (accountIDs = ids)" />
    <PeriodSwitch v-model="period" />
    <SummaryCards :accountIDs="accountIDs" :period="period" />
    <PerformanceCards v-model:index="index" :performance="performance" :error="performanceError" />
    <PositionsTable :accountIDs="accountIDs" :period="period" />
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
</style>
