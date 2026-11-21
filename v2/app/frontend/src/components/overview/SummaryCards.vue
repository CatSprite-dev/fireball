<script setup lang="ts">
import {computed} from 'vue'
import type {Overview} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'
import {day, rub, signedRub, tone} from '../../format'

// Сводку загружает страница: «Вложено» нужно и панели справа.
const props = defineProps<{overview: Overview | null; error: string; period: string}>()

// «Вложено» и «Прибыль» — за выбранный период.
const during = computed(() => (props.period === 'all' ? 'за всё время' : 'за период'))

// Рубли и копейки отдельно: копейки в стоимости приглушены.
const value = computed(() => {
  const [rubles, kopecks] = rub(props.overview?.value).split(',')
  return {rubles, kopecks: kopecks === undefined ? '' : `,${kopecks}`}
})

// Части прибыли, цвет — по знаку. Окно ничего не складывает: каждая часть пришла из Go отдельно.
const parts = computed(() => {
  const o = props.overview
  if (!o) return []
  return [
    {key: 'realized', title: 'Зафиксированная', value: o.realized, help: 'Прибыль по закрытым сделкам с комиссиями'},
    {key: 'paper', title: 'Бумажная', value: o.paper, help: 'Стоимость позиций минус их стоимость по средней цене'},
    {key: 'income', title: 'Выплаты', value: o.income, help: 'Дивиденды и купоны за вычетом налогов на них'},
  ].map(p => ({...p, tone: tone(p.value)}))
})
</script>

<template>
  <p v-if="error" class="error">{{ error }}</p>
  <section class="summary" aria-label="Сводка">
    <div class="card">
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

    <div class="card">
      <span class="label">Прибыль {{ during }}, ₽ <HelpMark text="Стоимость сегодня минус стоимость накануне периода минус вложено за период" /></span>
      <span class="mid" :class="tone(overview?.profit)">{{ signedRub(overview?.profit) }}</span>
      <div class="parts">
        <span v-for="p in parts" :key="p.key" class="part">
          <span class="muted">{{ p.title }} <HelpMark :text="p.help" /></span>
          <span class="part-value" :class="p.tone">{{ signedRub(p.value) }}</span>
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
  grid-template-columns: 2fr 1fr 1.9fr;
  gap: 1px;
  background: var(--line);
  border: 1px solid var(--line);
}
.card {
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
  padding: 14px 20px;
  background: var(--panel);
}
.big, .mid, .part-value {
  font-family: var(--cond);
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}
.big {
  font-size: 52px;
  line-height: 1.05;
}
.currency {
  margin-left: 6px;
  font-size: 26px;
}
.mid {
  font-size: 36px;
}
.note {
  font-size: 13px;
}
/* Части прибыли разделены линией в 1px, как блоки окна. */
.parts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
  border-top: 1px solid var(--line);
}
.part {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 0 0 12px;
  background: var(--panel);
}
.part:first-child {
  padding-left: 0;
}
.part-value {
  font-size: 18px;
}
</style>
