<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {use} from 'echarts/core'
import {CanvasRenderer} from 'echarts/renderers'
import {LineChart} from 'echarts/charts'
import {GridComponent, LegendComponent, TooltipComponent} from 'echarts/components'
import VChart from 'vue-echarts'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Charts, Point} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'

use([CanvasRenderer, LineChart, GridComponent, LegendComponent, TooltipComponent])

const props = defineProps<{accountIDs: string[] | null; period: string; index: string}>()

type ChartKey = 'shadowTrades' | 'shadowDeposits' | 'vsIndex' | 'value'
const tabs: {key: ChartKey; title: string}[] = [
  {key: 'shadowTrades', title: 'Тень по бумагам'},
  {key: 'shadowDeposits', title: 'Тень по пополнениям'},
  {key: 'vsIndex', title: 'Портфель против индекса'},
  {key: 'value', title: 'Стоимость портфеля'},
]
const selected = ref<ChartKey>('shadowTrades')

const charts = ref<Charts | null>(null)
const error = ref('')

// Ответы могут прийти не по порядку, если выбор меняют быстро: показываем только последний запрос.
let request = 0

watch([() => props.accountIDs, () => props.period, () => props.index], async ([ids, period, index]) => {
  if (ids === null) return
  const current = ++request
  charts.value = null
  error.value = ''
  try {
    const result = await PortfolioService.Charts(ids, period, index)
    if (current === request) charts.value = result
  } catch (e) {
    if (current === request) error.value = e instanceof Error ? e.message : String(e)
  }
}, {immediate: true})

// Линии выбранного графика. Строки в числа — только для рисования.
const lines = computed(() => {
  const c = charts.value
  if (!c) return []
  switch (selected.value) {
    case 'shadowTrades':
      return [{name: 'Бумаги', points: c.shadowTrades.actual}, {name: `Тень в ${props.index}`, points: c.shadowTrades.shadow}]
    case 'shadowDeposits':
      return [{name: 'Портфель', points: c.shadowDeposits.actual}, {name: `Тень в ${props.index}`, points: c.shadowDeposits.shadow}]
    case 'vsIndex':
      return [{name: 'Портфель (TWR)', points: c.twr}, {name: props.index, points: c.index}]
    case 'value':
      return [{name: 'Стоимость', points: c.value}]
  }
})

const unit = computed(() => (selected.value === 'vsIndex' ? '%' : '₽'))
const number = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})
const axisNumber = new Intl.NumberFormat('ru-RU', {maximumFractionDigits: 0})

type TooltipParam = {seriesName: string; marker: string; value: [string, number]}

const option = computed(() => ({
  animation: false,
  legend: {},
  grid: {left: 16, right: 16, top: 40, bottom: 8, containLabel: true},
  xAxis: {type: 'time'},
  yAxis: {type: 'value', scale: true, axisLabel: {formatter: (v: number) => `${axisNumber.format(v)} ${unit.value}`}},
  tooltip: {
    trigger: 'axis',
    formatter: (params: TooltipParam[]) =>
      [params[0]?.value[0] ?? '', ...params.map(p => `${p.marker}${p.seriesName}: ${number.format(p.value[1])} ${unit.value}`)].join('<br>'),
  },
  series: lines.value.map(l => ({
    name: l.name,
    type: 'line',
    showSymbol: false,
    data: (l.points ?? []).map((p: Point) => [p.date, Number(p.value)]),
  })),
}))
</script>

<template>
  <section>
    <div class="tabs">
      <label v-for="t in tabs" :key="t.key">
        <input type="radio" v-model="selected" :value="t.key" /> {{ t.title }}
      </label>
    </div>

    <p v-if="error">{{ error }}</p>
    <p v-else-if="!charts">Загружаю…</p>
    <VChart v-else class="chart" :option="option" :update-options="{notMerge: true}" autoresize />
  </section>
</template>

<style scoped>
.tabs {
  display: flex;
  gap: 1rem;
}
.chart {
  height: 320px;
}
</style>
