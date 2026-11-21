<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {use} from 'echarts/core'
import {CanvasRenderer} from 'echarts/renderers'
import {LineChart} from 'echarts/charts'
import {GridComponent, TooltipComponent} from 'echarts/components'
import VChart from 'vue-echarts'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Charts, Performance, Point} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import {day} from '../../format'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent])

const props = defineProps<{accountIDs: string[] | null; period: string; index: string; performance: Performance | null}>()

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

// Цвета — из CSS-переменных, как у остального окна.
const css = getComputedStyle(document.documentElement)
const color = (name: string) => css.getPropertyValue(name).trim()
const theme = {fg: color('--fg'), muted: color('--muted'), dim: color('--dim'), line: color('--line'), panel: color('--panel'), acc: color('--acc')}
const font = 'IBM Plex Sans Condensed'

// Линии выбранного графика: портфель — сплошная, индекс и тень — пунктир цветом индекса.
const lines = computed(() => {
  const c = charts.value
  if (!c) return []
  const shadow = `Тень в ${props.index}`
  switch (selected.value) {
    case 'shadowTrades':
      return [{name: 'Бумаги', points: c.shadowTrades.actual, shadow: false}, {name: shadow, points: c.shadowTrades.shadow, shadow: true}]
    case 'shadowDeposits':
      return [{name: 'Портфель', points: c.shadowDeposits.actual, shadow: false}, {name: shadow, points: c.shadowDeposits.shadow, shadow: true}]
    case 'vsIndex':
      return [{name: 'Портфель, TWR', points: c.twr, shadow: false}, {name: props.index, points: c.index, shadow: true}]
    case 'value':
      return [{name: 'Стоимость', points: c.value, shadow: false}]
  }
})

// Пояснение под графиком.
const note = computed(() => {
  const i = props.index
  switch (selected.value) {
    case 'shadowTrades':
      return `Тень — те же покупки и продажи акций, облигаций и фондов, как если бы деньги шли в ${i}.`
    case 'shadowDeposits':
      return `Тень — те же пополнения и выводы, вложенные в ${i}.`
    case 'vsIndex':
      return props.performance?.compareFrom
        ? `Рост портфеля и индекса с ${day(props.performance.compareFrom)}: раньше индекса нет.`
        : 'Рост портфеля и индекса с начала периода.'
    case 'value':
      return 'Стоимость выбранных счетов на конец каждого дня.'
  }
})

const percent = computed(() => selected.value === 'vsIndex')
const exact = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})
const whole = new Intl.NumberFormat('ru-RU', {maximumFractionDigits: 0})
const compact = new Intl.NumberFormat('ru-RU', {notation: 'compact', maximumFractionDigits: 2})
const minus = (s: string) => s.replace('-', '−')
const axisLabel = (v: number) => minus(percent.value ? `${whole.format(v)}%` : compact.format(v))
const endLabel = (v: number) => minus(percent.value ? `${exact.format(v)}%` : whole.format(v))
const tipValue = (v: number) => minus(`${exact.format(v)} ${percent.value ? '%' : '₽'}`)

type TooltipParam = {seriesName: string; color: string; value: [string, number]}

const option = computed(() => ({
  animation: false,
  textStyle: {fontFamily: font},
  grid: {left: 8, right: 80, top: 12, bottom: 4, containLabel: true},
  xAxis: {
    type: 'time',
    axisLine: {lineStyle: {color: theme.line}},
    axisTick: {show: false},
    axisLabel: {color: theme.muted, fontFamily: font, fontSize: 12},
    splitLine: {show: false},
  },
  yAxis: {
    type: 'value',
    scale: true,
    axisLabel: {color: theme.muted, fontFamily: font, fontSize: 12, formatter: axisLabel},
    splitLine: {lineStyle: {color: theme.dim}},
  },
  tooltip: {
    trigger: 'axis',
    backgroundColor: theme.panel,
    borderColor: theme.fg,
    borderRadius: 0,
    textStyle: {color: theme.fg, fontFamily: font, fontSize: 12},
    axisPointer: {lineStyle: {color: theme.muted}},
    formatter: (params: TooltipParam[]) =>
      [
        day(params[0]?.value[0]),
        ...params.map(p => `<span style="display:inline-block;width:8px;height:8px;margin-right:6px;background:${p.color}"></span>${p.seriesName}: <b>${tipValue(p.value[1])}</b>`),
      ].join('<br>'),
  },
  series: lines.value.map(l => ({
    name: l.name,
    type: 'line',
    showSymbol: false,
    color: l.shadow ? theme.acc : theme.fg,
    lineStyle: {width: l.shadow ? 2 : 2.5, type: l.shadow ? [6, 4] : 'solid'},
    endLabel: {
      show: true,
      color: l.shadow ? theme.acc : theme.fg,
      fontFamily: font,
      fontSize: 13,
      fontWeight: 700,
      formatter: ({value}: {value: [string, number]}) => endLabel(value[1]),
    },
    data: (l.points ?? []).map((p: Point) => [p.date, Number(p.value)]),
  })),
}))
</script>

<template>
  <div class="chart-panel">
    <div class="seg modes" role="group" aria-label="Режим графика">
      <button v-for="t in tabs" :key="t.key" type="button" :aria-pressed="t.key === selected" @click="selected = t.key">{{ t.title }}</button>
    </div>

    <div class="legend">
      <span v-for="l in lines" :key="l.name" class="legend-item">
        <span :class="['swatch', {shadow: l.shadow}]"></span>{{ l.name }}
      </span>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-else-if="!charts" class="loading muted">Загружаю…</p>
    <VChart v-else class="chart" :option="option" :update-options="{notMerge: true}" autoresize />

    <p class="note muted">{{ note }}</p>
  </div>
</template>

<style scoped>
.chart-panel {
  grid-column: span 2;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  background: var(--panel);
}
.modes button {
  font-size: 12px;
  font-weight: 500;
}
.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 18px;
  font-size: 12px;
  color: var(--muted);
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}
.swatch {
  display: inline-block;
  width: 18px;
  border-top: 2px solid var(--fg);
}
.swatch.shadow {
  border-top: 2px dashed var(--acc);
}
.chart, .loading, .error {
  height: 260px;
  margin: 0;
}
.error {
  color: var(--neg);
}
.note {
  margin: 0;
  font-size: 12px;
}
</style>
