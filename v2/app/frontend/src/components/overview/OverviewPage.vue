<script setup lang="ts">
import {ref} from 'vue'
import AccountsFilter from './AccountsFilter.vue'
import PeriodSwitch from './PeriodSwitch.vue'
import SummaryCards from './SummaryCards.vue'
import PerformanceCards from './PerformanceCards.vue'
import PositionsTable from './PositionsTable.vue'

// null — счета ещё не загружены, выбора пока нет.
const accountIDs = ref<string[] | null>(null)
</script>

<template>
  <div class="page">
    <AccountsFilter @change="ids => (accountIDs = ids)" />
    <PeriodSwitch />
    <SummaryCards :accountIDs="accountIDs" />
    <PerformanceCards />
    <PositionsTable />
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
</style>
