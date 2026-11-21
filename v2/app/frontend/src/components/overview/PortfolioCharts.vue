<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {registerLocale, use} from 'echarts/core'
import langRU from 'echarts/i18n/langRU-obj.js'
import {CanvasRenderer} from 'echarts/renderers'
import {LineChart} from 'echarts/charts'
import {GraphicComponent, GridComponent, MarkLineComponent, TooltipComponent} from 'echarts/components'
import VChart from 'vue-echarts'
import type {Charts, Performance, Point} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import {day} from '../../format'
import type {ChartMode} from './chartMode'

use([CanvasRenderer, LineChart, GraphicComponent, GridComponent, MarkLineComponent, TooltipComponent])
// Месяцы на оси — по-русски.
registerLocale('RU', langRU)

// Ряды загружает страница: по ним же панель справа показывает итоги выбранного графика.
const props = defineProps<{charts: Charts | null; error: string; index: string; performance: Performance | null}>()
const selected = defineModel<ChartMode>('mode', {required: true})

const tabs: {key: ChartMode; title: string}[] = [
  {key: 'shadowTrades', title: 'Призрак сделок'},
  {key: 'shadowDeposits', title: 'Призрак пополнений'},
  {key: 'vsIndex', title: 'Портфель против индекса'},
  {key: 'value', title: 'Стоимость портфеля'},
]

// Цвета — из CSS-переменных, как у остального окна.
const css = getComputedStyle(document.documentElement)
const color = (name: string) => css.getPropertyValue(name).trim()
const theme = {fg: color('--fg'), muted: color('--muted'), dim: color('--dim'), line: color('--line'), panel: color('--panel'), me: color('--me'), bench: color('--bench')}
const font = 'IBM Plex Sans Condensed'

// Линии выбранного графика: портфель — сплошная оранжевая, индекс и призрак — голубой пунктир.
// ghost — линия призрака: она полупрозрачная, чтобы не спорить с настоящим портфелем.
const lines = computed(() => {
  const c = props.charts
  if (!c) return []
  const i = props.index
  switch (selected.value) {
    case 'shadowTrades':
      return [{name: 'Бумаги', points: c.shadowTrades.actual, shadow: false, ghost: false}, {name: `Призрак в ${i}`, points: c.shadowTrades.shadow, shadow: true, ghost: true}]
    case 'shadowDeposits':
      return [{name: 'Портфель', points: c.shadowDeposits.actual, shadow: false, ghost: false}, {name: `Призрак пополнений в ${i}`, points: c.shadowDeposits.shadow, shadow: true, ghost: true}]
    case 'vsIndex':
      return [{name: 'Портфель, TWR', points: c.twr, shadow: false, ghost: false}, {name: i, points: c.index, shadow: true, ghost: false}]
    case 'value':
      return [{name: 'Стоимость', points: c.value, shadow: false, ghost: false}]
  }
})

// Пояснение под графиком. Призраки, как и TWR, начинаются с первого дня индекса, если он короче периода.
const indexFrom = computed(() =>
  props.performance?.compareFrom ? ` Индекс есть с ${day(props.performance.compareFrom)}, сравнение с этой даты.` : '',
)
const note = computed(() => {
  const i = props.index
  switch (selected.value) {
    case 'shadowTrades':
      return `Призрак: те же покупки и продажи акций, облигаций и фондов, как если бы деньги шли в ${i}.${indexFrom.value}`
    case 'shadowDeposits':
      return `Призрак: те же пополнения и выводы, вложенные в ${i}.${indexFrom.value}`
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
// Подписи на концах линий — с копейками: дельта призраков в панели равна разнице этих подписей.
const compact = new Intl.NumberFormat('ru-RU', {notation: 'compact', maximumFractionDigits: 2})
const minus = (s: string) => s.replace('-', '−')
const axisLabel = (v: number) => minus(percent.value ? `${whole.format(v)}%` : compact.format(v))
const endLabel = (v: number) => minus(percent.value ? `${exact.format(v)}%` : exact.format(v))
// Дата в подсказке: «21 ноября 2026».
const longDay = new Intl.DateTimeFormat('ru-RU', {day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC'})
const tipDay = (s: string | undefined) => (s ? longDay.format(new Date(`${s}T00:00:00Z`)).replace(' г.', '') : '')
// Подписи оси дат: год — жирным, месяцы — со строчной, внутри месяца — день и месяц.
const months = ['янв', 'фев', 'мар', 'апр', 'май', 'июн', 'июл', 'авг', 'сен', 'окт', 'ноя', 'дек']
function axisDay(value: number): string {
  const d = new Date(value)
  if (d.getDate() !== 1) return `${d.getDate()} ${months[d.getMonth()]}`
  if (d.getMonth() === 0) return `{year|${d.getFullYear()}}`
  return months[d.getMonth()]
}
const tipValue = (v: number) => minus(`${exact.format(v)} ${percent.value ? '%' : '₽'}`)

type TooltipParam = {seriesName: string; color: string; value: [string, number]}

// Подписи значений на концах линий рисуем сами, поверх графика: endLabel ECharts ставит
// в позицию сам, и близкие подписи не раздвигаются. После каждой отрисовки берём пиксели
// последних точек и раздвигаем подписи, если они ближе LABEL_GAP.
const LABEL_GAP = 16
type EndLabel = {text: string; color: string; x: number; y: number}
const chart = ref<InstanceType<typeof VChart> | null>(null)
const endLabels = ref<EndLabel[]>([])
watch(lines, () => (endLabels.value = []))

function placeEndLabels() {
  const c = chart.value
  if (!c) return
  const placed: EndLabel[] = []
  lines.value.forEach((l, i) => {
    const p = l.points?.[l.points.length - 1]
    if (!p) return
    const px = c.convertToPixel({seriesIndex: i}, [p.date, Number(p.value)] as never) as unknown as number[] | undefined
    if (!px) return
    placed.push({text: endLabel(Number(p.value)), color: l.shadow ? theme.bench : theme.me, x: Math.round(px[0] + 8), y: px[1]})
  })
  // Сверху вниз: каждая подпись не ближе LABEL_GAP к предыдущей, потом все вместе — обратно к середине.
  placed.sort((a, b) => a.y - b.y)
  const before = placed.reduce((sum, l) => sum + l.y, 0)
  for (let k = 1; k < placed.length; k++) placed[k].y = Math.max(placed[k].y, placed[k - 1].y + LABEL_GAP)
  const drift = placed.length ? (placed.reduce((sum, l) => sum + l.y, 0) - before) / placed.length : 0
  placed.forEach(l => (l.y = Math.round(l.y - drift)))
  // Перерисовка после смены подписей снова вызовет placeEndLabels: обновляем, только если что-то сдвинулось.
  if (JSON.stringify(placed) !== JSON.stringify(endLabels.value)) endLabels.value = placed
}

const option = computed(() => ({
  animation: false,
  textStyle: {fontFamily: font},
  grid: {left: 8, right: 100, top: 12, bottom: 4, containLabel: true},
  xAxis: {
    type: 'time',
    axisLine: {lineStyle: {color: theme.line}},
    axisTick: {show: false},
    axisLabel: {
      color: theme.muted,
      fontFamily: font,
      fontSize: 13,
      formatter: axisDay,
      rich: {year: {color: theme.fg, fontFamily: font, fontSize: 13, fontWeight: 700}},
    },
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
        tipDay(params[0]?.value[0]),
        ...params.map(p => `<span style="display:inline-block;width:8px;height:8px;margin-right:6px;background:${p.color}"></span>${p.seriesName}: <b>${tipValue(p.value[1])}</b>`),
      ].join('<br>'),
  },
  series: lines.value.map(l => ({
    name: l.name,
    type: 'line',
    showSymbol: false,
    color: l.shadow ? theme.bench : theme.me,
    lineStyle: {width: l.shadow ? 2 : 2.5, type: l.shadow ? [5, 4] : 'solid', opacity: l.ghost ? 0.6 : 1},
    // В процентах ноль — заметнее остальной сетки.
    markLine: percent.value && !l.shadow
      ? {silent: true, symbol: 'none', label: {show: false}, lineStyle: {color: theme.muted, type: 'solid', width: 1}, data: [{yAxis: 0}]}
      : undefined,
    data: (l.points ?? []).map((p: Point) => [p.date, Number(p.value)]),
  })),
  graphic: endLabels.value.map(l => ({
    type: 'text',
    x: l.x,
    y: l.y,
    silent: true,
    style: {text: l.text, fill: l.color, font: `700 14px "${font}"`, verticalAlign: 'middle'},
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
        <span :class="['swatch', {shadow: l.shadow, ghost: l.ghost}]"></span>{{ l.name }}
      </span>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-else-if="!charts" class="loading muted">Загружаю…</p>
    <VChart
      v-else
      ref="chart"
      class="chart"
      :init-options="{locale: 'RU'}"
      :option="option"
      :update-options="{notMerge: true}"
      autoresize
      @finished="placeEndLabels"
    />

    <p class="note muted">{{ note }}</p>
  </div>
</template>

<style scoped>
.chart-panel {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 18px;
  background: var(--panel);
}
.modes button {
  padding: 0 14px;
  font-size: 13px;
  font-weight: 500;
}
.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 20px;
  font-size: 13px;
  color: var(--muted);
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.swatch {
  display: inline-block;
  width: 20px;
  border-top: 3px solid var(--me);
}
.swatch.ghost {
  opacity: .6;
}
.swatch.shadow {
  border-top: 3px dashed var(--bench);
}
.chart, .loading, .error {
  height: 300px;
  margin: 0;
}
.error {
  color: var(--neg);
}
.note {
  margin: 4px 0 0;
}
</style>
