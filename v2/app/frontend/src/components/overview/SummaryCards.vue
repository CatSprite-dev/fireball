<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Overview} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'
import {day, rub, signedRub, tone} from '../../format'

const props = defineProps<{accountIDs: string[] | null; period: string}>()
const emit = defineEmits<{asOf: [day: string]}>()

// «Вложено» и «Прибыль» — за выбранный период.
const during = computed(() => (props.period === 'all' ? 'за всё время' : 'за период'))

const overview = ref<Overview | null>(null)
const error = ref('')

// Ответы могут прийти не по порядку, если галочки щёлкают быстро: показываем только последний запрос.
let request = 0

watch([() => props.accountIDs, () => props.period], async ([ids, period]) => {
  if (ids === null) return
  const current = ++request
  overview.value = null
  error.value = ''
  try {
    const result = await PortfolioService.Overview(ids, period)
    if (current === request) {
      overview.value = result
      emit('asOf', result.asOf)
    }
  } catch (e) {
    if (current === request) error.value = e instanceof Error ? e.message : String(e)
  }
}, {immediate: true})

// Рубли и копейки отдельно: копейки в стоимости приглушены.
const value = computed(() => {
  const [rubles, kopecks] = rub(overview.value?.value).split(',')
  return {rubles, kopecks: kopecks === undefined ? '' : `,${kopecks}`}
})

// Полоса из частей прибыли: ширина — по модулю части, цвет — по знаку.
// Окно ничего не складывает: каждая часть пришла из Go отдельно.
const parts = computed(() => {
  const o = overview.value
  if (!o) return []
  return [
    {key: 'realized', title: 'Зафиксированная', value: o.realized, help: 'Прибыль по закрытым сделкам с комиссиями'},
    {key: 'paper', title: 'Бумажная', value: o.paper, help: 'Стоимость позиций минус их стоимость по средней цене'},
    {key: 'income', title: 'Выплаты', value: o.income, help: 'Дивиденды и купоны за вычетом налогов на них'},
  ].map(p => ({...p, weight: Math.abs(Number(p.value)), tone: tone(p.value)}))
})
</script>

<template>
  <p v-if="error" class="error">{{ error }}</p>
  <section class="summary" aria-label="Сводка">
    <div class="card wide">
      <span class="label">Стоимость <HelpMark text="Рыночная стоимость бумаг и денег на выбранных счетах; совпадает с брокером" /></span>
      <span class="big">
        <template v-if="overview">{{ value.rubles }}<span class="muted">{{ value.kopecks }}</span></template>
        <template v-else>—</template>
        <span class="currency muted">₽</span>
      </span>
      <span class="note muted">на {{ day(overview?.asOf) }}</span>
    </div>

    <div class="card">
      <span class="label">Вложено <HelpMark :text="`Пополнения минус выводы ${during}`" /></span>
      <span class="mid">{{ rub(overview?.invested) }}</span>
      <span class="note muted">{{ during }}, ₽</span>
    </div>

    <div class="card wide">
      <span class="label">Прибыль {{ during }}, ₽ <HelpMark text="Стоимость сегодня минус стоимость накануне периода минус вложено за период" /></span>
      <span class="mid" :class="tone(overview?.profit)">{{ signedRub(overview?.profit) }}</span>
      <div class="bar" aria-hidden="true">
        <span v-for="p in parts" :key="p.key" :class="['segment', p.key, p.tone]" :style="{flexGrow: p.weight}"></span>
      </div>
      <div class="parts">
        <span v-for="p in parts" :key="p.key" class="part">
          <span class="muted"><span :class="['swatch', p.key, p.tone]"></span>{{ p.title }} <HelpMark :text="p.help" /></span>
          <span class="part-value">{{ signedRub(p.value) }}</span>
        </span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.error {
  margin: 0;
  color: var(--neg);
}
.summary {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
  border: 1px solid var(--line);
}
.card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 6px;
  padding: 12px 16px;
  background: var(--panel);
}
.wide {
  grid-column: span 2;
}
.big, .mid, .part-value {
  font-family: var(--cond);
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}
.big {
  font-size: 44px;
  letter-spacing: -0.01em;
}
.currency {
  margin-left: 6px;
  font-size: 22px;
}
.mid {
  font-size: 28px;
}
.note {
  font-size: 12px;
}
.bar {
  display: flex;
  height: 6px;
  gap: 2px;
}
.segment, .swatch {
  background: var(--muted);
}
.segment.pos, .swatch.pos {
  background: var(--pos);
}
.segment.neg, .swatch.neg {
  background: var(--neg);
}
.segment.paper, .swatch.paper {
  opacity: .45;
}
.segment.income.pos, .swatch.income.pos {
  background: var(--acc);
}
.parts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  font-size: 12px;
}
.part {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.swatch {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 5px;
}
.part-value {
  font-size: 15px;
}
</style>
