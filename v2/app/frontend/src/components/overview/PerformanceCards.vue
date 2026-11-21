<script setup lang="ts">
import {computed} from 'vue'
import type {Charts, Overview, Performance, Point} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'
import {day, rub, signedPercent, signedRub, tone} from '../../format'
import type {ChartMode} from './chartMode'

// Панель справа от графика: три строки итогов выбранного графика — «моё», «эталон», «дельта».
// Цифры страница передаёт готовыми; окно их только показывает.
const props = defineProps<{
  mode: ChartMode
  charts: Charts | null
  overview: Overview | null
  index: string
  performance: Performance | null
  error: string
}>()
defineEmits<{more: []}>()

const first = (points: Point[] | null | undefined) => points?.[0]
const last = (points: Point[] | null | undefined) => points?.[points.length - 1]

// color — цвет числа: me — как линия портфеля, bench — как линия индекса и призрака, fg — нейтральный,
// pos/neg/muted — по знаку разницы. square — квадратик перед подписью: цвет линии, пустой или пунктирный.
type Row = {label: string; sub: string; value: string; color: string; square: 'me' | 'bench' | 'empty' | 'dashed'; help: string}
type Block = {from: string | undefined; unit: string; rows: Row[]}

const block = computed<Block>(() => {
  const c = props.charts
  const i = props.index
  switch (props.mode) {
    case 'vsIndex': {
      // Обе линии графика начинаются с одного дня: с начала периода или с первого дня индекса.
      const from = first(c?.twr)?.date ?? (props.performance?.compareFrom || props.performance?.from)
      return {
        from,
        unit: 'в %',
        rows: [
          {label: 'Портфель, TWR', sub: `с ${day(from)}`, value: signedPercent(last(c?.twr)?.value), color: 'me', square: 'me', help: 'Рост портфеля без влияния пополнений и выводов, с начала графика'},
          {label: `Индекс ${i}`, sub: `с ${day(from)}`, value: signedPercent(last(c?.index)?.value), color: 'bench', square: 'bench', help: `Рост индекса ${i} за тот же отрезок`},
          {label: 'Против индекса', sub: 'портфель − индекс', value: signedPercent(props.performance?.vsIndex, ' п.п.'), color: tone(props.performance?.vsIndex), square: 'empty', help: 'TWR минус рост индекса за тот же отрезок, процентные пункты'},
        ],
      }
    }
    case 'shadowTrades': {
      const diff = c?.shadowTrades.diff
      const ahead = Number(diff) > 0 ? 'бумаги обогнали индекс' : Number(diff) < 0 ? 'бумаги отстали от индекса' : 'бумаги − призрак'
      return {
        from: first(c?.shadowTrades.actual)?.date,
        unit: 'в рублях',
        rows: [
          {label: 'Бумаги', sub: `стоимость на ${day(last(c?.shadowTrades.actual)?.date)}`, value: rub(last(c?.shadowTrades.actual)?.value), color: 'me', square: 'me', help: 'Стоимость акций, облигаций и фондов на последнюю дату'},
          {label: `Призрак в ${i}`, sub: 'те же сделки в индексе', value: rub(last(c?.shadowTrades.shadow)?.value), color: 'bench', square: 'bench', help: `Те же покупки и продажи, как если бы деньги шли в ${i}`},
          {label: 'Дельта', sub: ahead, value: signedRub(diff), color: tone(diff), square: 'empty', help: 'Бумаги минус призрак на последнюю дату'},
        ],
      }
    }
    case 'shadowDeposits':
      return {
        from: first(c?.shadowDeposits.actual)?.date,
        unit: 'в рублях',
        rows: [
          {label: 'Портфель', sub: `стоимость на ${day(last(c?.shadowDeposits.actual)?.date)}`, value: rub(last(c?.shadowDeposits.actual)?.value), color: 'me', square: 'me', help: 'Стоимость выбранных счетов на последнюю дату'},
          {label: 'Призрак пополнений', sub: 'все пополнения сразу в индекс', value: rub(last(c?.shadowDeposits.shadow)?.value), color: 'bench', square: 'bench', help: `Те же пополнения и выводы, вложенные в ${i}`},
          {label: 'Дельта', sub: 'портфель − призрак', value: signedRub(c?.shadowDeposits.diff), color: tone(c?.shadowDeposits.diff), square: 'empty', help: 'Портфель минус призрак на последнюю дату'},
        ],
      }
    case 'value':
      return {
        from: first(c?.value)?.date,
        unit: 'в рублях',
        rows: [
          {label: 'На начало периода', sub: `на ${day(first(c?.value)?.date)}`, value: rub(first(c?.value)?.value), color: 'fg', square: 'me', help: 'Стоимость на конец первого дня графика'},
          {label: 'Пополнения − выводы', sub: 'за период', value: rub(props.overview?.invested), color: 'fg', square: 'dashed', help: 'За период, то же, что «Вложено» в сводке'},
          {label: 'На конец периода', sub: `на ${day(last(c?.value)?.date)}`, value: rub(last(c?.value)?.value), color: 'fg', square: 'me', help: 'Стоимость на последнюю дату'},
        ],
      }
  }
  return {from: undefined, unit: '', rows: []}
})
</script>

<template>
  <div class="column">
    <div class="row">
      <span class="caption">
        <span class="label">Сравнивать с</span>
        <span class="muted">С {{ day(block.from) }} · {{ block.unit }}</span>
      </span>
      <span class="chip" title="Индекс МосБиржи полной доходности брутто">{{ index }}</span>
    </div>

    <p v-if="error" class="row error">{{ error }}</p>

    <div v-for="(r, n) in block.rows" :key="n" class="row value-row">
      <span class="caption">
        <span class="label title">
          <span :class="['square', r.square]" aria-hidden="true"></span>{{ r.label }}<HelpMark :text="r.help" />
        </span>
        <span class="sub muted">{{ r.sub }}</span>
      </span>
      <span :class="['number', r.color]">{{ r.value }}</span>
    </div>

    <p class="row muted">Индекс полной доходности брутто: без налогов, обычно на ~1 п.п. в год выше реально достижимого.</p>

    <button type="button" class="more label" @click="$emit('more')">
      Другие показатели<span aria-hidden="true">›</span>
    </button>
  </div>
</template>

<style scoped>
.column {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: var(--line);
}
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 0;
  padding: 12px 18px;
  background: var(--panel);
}
.value-row {
  flex: 1;
  min-height: 96px;
}
.caption {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.title {
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}
.sub {
  font-size: 12px;
}
.square {
  flex: none;
  width: 10px;
  height: 10px;
  border: 1px solid var(--muted);
}
.square.me {
  background: var(--me);
  border-color: var(--me);
}
.square.bench {
  background: var(--bench);
  border-color: var(--bench);
}
.square.dashed {
  border-style: dashed;
}
.error {
  color: var(--neg);
}
/* Выбранный индекс — инверсия белым, как выбранный период и режим. */
.chip {
  height: 32px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  border: 1px solid var(--fg);
  background: var(--fg);
  color: var(--inv);
  font-family: var(--cond);
  font-size: 14px;
  font-weight: 700;
  letter-spacing: .03em;
}
/* Один размер во всех режимах и строках: «−1 234 567,89» влезает рядом с подписью в окне шириной 1000. */
.number {
  font-family: var(--cond);
  font-size: 26px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}
.me {
  color: var(--me);
}
.bench {
  color: var(--bench);
}
.fg {
  color: var(--fg);
}
.more {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;
  padding: 0 18px;
  border: 0;
  background: var(--panel);
  color: var(--fg);
  cursor: pointer;
  text-align: left;
}
</style>
