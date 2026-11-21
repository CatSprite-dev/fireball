<script setup lang="ts">
import {computed, reactive, ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Position, PositionTable} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'
import {rub, signedPercent, signedRub, tone} from '../../format'

const props = defineProps<{accountIDs: string[] | null; period: string}>()

type ColumnKey = 'ticker' | 'quantity' | 'value' | 'invested' | 'income' | 'profit' | 'profitPct' | 'share'
type Column = {key: ColumnKey; title: string; tip: string}

const columns: Column[] = [
  {key: 'ticker', title: 'Бумага', tip: 'Тикер и название'},
  {key: 'quantity', title: 'Кол-во', tip: 'Количество на сегодня'},
  {key: 'value', title: 'Стоимость, ₽', tip: 'Рыночная стоимость сегодня'},
  {key: 'invested', title: 'Вложено, ₽', tip: 'Стоимость остатка по средней цене'},
  {key: 'income', title: 'Выплаты, ₽', tip: 'Дивиденды и купоны за период за вычетом налогов на них'},
  {key: 'profit', title: 'Прибыль, ₽', tip: 'Зафиксированная + изменение бумажной + выплаты за период'},
  {key: 'profitPct', title: 'Прибыль, %', tip: 'Пока не считается'},
  {key: 'share', title: 'Доля', tip: 'Доля в стоимости выбранных счетов'},
]

const groups = [
  {kind: 'etf', title: 'Фонды'},
  {kind: 'bond', title: 'Облигации'},
  {kind: 'share', title: 'Акции'},
  {kind: 'futures', title: 'Фьючерсы'},
  {kind: 'option', title: 'Опционы'},
  {kind: 'currency', title: 'Валюта'},
  {kind: 'other', title: 'Прочее'},
]
const open = reactive<Record<string, boolean>>(Object.fromEntries(groups.map(g => [g.kind, true])))

const showClosed = ref(false)
const noGroups = ref(false)

const table = ref<PositionTable | null>(null)
const error = ref('')

// Ответы могут прийти не по порядку, если выбор меняют быстро: показываем только последний запрос.
let request = 0

watch([() => props.accountIDs, () => props.period], async ([ids, period]) => {
  if (ids === null) return
  const current = ++request
  table.value = null
  error.value = ''
  try {
    const result = await PortfolioService.Positions(ids, period)
    if (current === request) table.value = result
  } catch (e) {
    if (current === request) error.value = e instanceof Error ? e.message : String(e)
  }
}, {immediate: true})

// Сортировка: повторный клик по той же колонке меняет направление. Сортируется внутри групп.
const sortKey = ref<ColumnKey>('value')
const sortDesc = ref(true)
function sortBy(key: ColumnKey) {
  if (sortKey.value === key) {
    sortDesc.value = !sortDesc.value
  } else {
    sortKey.value = key
    sortDesc.value = key !== 'ticker'
  }
}
function sortIcon(key: ColumnKey) {
  if (sortKey.value !== key) return ''
  return sortDesc.value ? '↓' : '↑'
}
function ariaSort(key: ColumnKey) {
  if (sortKey.value !== key) return 'none'
  return sortDesc.value ? 'descending' : 'ascending'
}

// Деньги приходят строкой; в число — только для показа и сортировки.
function compare(a: Position, b: Position): number {
  const key = sortKey.value
  const result = key === 'ticker'
    ? a.ticker.localeCompare(b.ticker, 'ru')
    : (a[key] === '' ? -Infinity : Number(a[key])) - (b[key] === '' ? -Infinity : Number(b[key]))
  return sortDesc.value ? -result : result
}

const visible = computed(() =>
  (table.value?.positions ?? []).filter(p => showClosed.value || !p.closed).sort(compare),
)

// Итоги групп и всей таблицы считает Go: окно деньги не складывает.
// Итог группы — по всем её позициям, вместе с закрытыми, даже когда они скрыты.
const grouped = computed(() => {
  if (noGroups.value) return [{kind: 'all', title: '', rows: visible.value, value: '', profit: ''}]
  return groups
    .map(g => {
      const total = table.value?.groups?.find(t => t.kind === g.kind)
      const rows = visible.value.filter(p => p.kind === g.kind)
      return {...g, rows, value: total?.value ?? '', profit: total?.profit ?? ''}
    })
    .filter(g => g.rows.length > 0)
})

const unexplainedHelp =
  'Брокер не отдаёт через API некоторые операции (например, налог при закрытии ИИС). ' +
  'Мы видим разницу в деньгах на счёте и вычитаем её, чтобы цифры сходились с брокером'

const quantity = new Intl.NumberFormat('ru-RU', {maximumFractionDigits: 6})
function count(s: string): string {
  return s === '' ? '—' : quantity.format(Number(s))
}
// Доля: число и полоска; полная полоска — 25% и больше.
function shareWidth(s: string): string {
  return `${Math.min(Number(s) / 25, 1) * 100}%`
}
</script>

<template>
  <section class="positions" aria-label="Бумаги">
    <div class="head">
      <h2 class="label">Бумаги · {{ visible.length }}</h2>
      <label class="check"><input type="checkbox" v-model="showClosed" /> Закрытые позиции</label>
      <label class="check"><input type="checkbox" v-model="noGroups" /> Без групп</label>
    </div>

    <p v-if="error" class="message error">{{ error }}</p>
    <p v-else-if="!table" class="message muted">Загружаю…</p>

    <div v-else class="scroll">
      <table>
        <thead>
          <tr>
            <th v-for="c in columns" :key="c.key" :class="{num: c.key !== 'ticker'}" :aria-sort="ariaSort(c.key)">
              <button type="button" class="sort" :class="{active: sortKey === c.key}" :title="c.tip" @click="sortBy(c.key)">
                {{ c.title }}<span class="arrow">{{ sortIcon(c.key) }}</span>
              </button>
            </th>
          </tr>
        </thead>

        <tbody v-for="g in grouped" :key="g.kind">
          <tr v-if="!noGroups" class="group">
            <th scope="rowgroup">
              <button type="button" class="toggle" :aria-expanded="open[g.kind]" @click="open[g.kind] = !open[g.kind]">
                {{ open[g.kind] ? '▾' : '▸' }} {{ g.title }} <span class="muted">· {{ g.rows.length }}</span>
              </button>
            </th>
            <td></td>
            <td class="num">{{ rub(g.value) }}</td>
            <td></td>
            <td></td>
            <td class="num" :class="tone(g.profit)">{{ signedRub(g.profit) }}</td>
            <td></td>
            <td></td>
          </tr>
          <template v-if="noGroups || open[g.kind]">
            <tr v-for="p in g.rows" :key="p.key" :class="{closed: p.closed}">
              <td class="security">
                <span class="ticker">{{ p.ticker }}</span>
                <span v-if="p.name" class="name">{{ p.name }}</span>
              </td>
              <td class="num muted">{{ count(p.quantity) }}</td>
              <td class="num strong">{{ rub(p.value) }}</td>
              <td class="num muted">{{ rub(p.invested) }}</td>
              <td class="num" :class="{muted: Number(p.income) === 0}">{{ rub(p.income) }}</td>
              <td class="num strong" :class="tone(p.profit)">{{ signedRub(p.profit) }}</td>
              <td class="num" :class="tone(p.profitPct)">{{ p.profitPct ? signedPercent(p.profitPct) : '—' }}</td>
              <td class="num">
                <span class="share">
                  <span>{{ p.share ? `${rub(p.share)}%` : '—' }}</span>
                  <span class="share-bar" aria-hidden="true"><span :style="{width: shareWidth(p.share)}"></span></span>
                </span>
              </td>
            </tr>
          </template>
        </tbody>

        <tbody>
          <tr class="group">
            <th scope="rowgroup" :colspan="columns.length">На уровне счёта</th>
          </tr>
          <tr class="dashed">
            <td colspan="5">Комиссии и налоги без бумаги</td>
            <td class="num muted">{{ signedRub(table.accountLevel) }}</td>
            <td colspan="2"></td>
          </tr>
          <tr class="dashed">
            <td colspan="2">
              Необъяснённое расхождение <HelpMark :text="unexplainedHelp" />
              <span v-if="Number(table.unexplained) === 0" class="badge">Сходится</span>
            </td>
            <td class="num muted">{{ signedRub(table.unexplained) }}</td>
            <td colspan="2"></td>
            <td class="num muted">{{ table.unexplainedProfit ? signedRub(table.unexplainedProfit) : '—' }}</td>
            <td colspan="2"></td>
          </tr>
        </tbody>

        <tfoot>
          <tr>
            <th scope="row">Итого</th>
            <td></td>
            <td class="num">{{ rub(table.totalValue) }}</td>
            <td></td>
            <td></td>
            <td class="num" :class="tone(table.totalProfit)">{{ signedRub(table.totalProfit) }}</td>
            <td colspan="2"></td>
          </tr>
        </tfoot>
      </table>
    </div>
  </section>
</template>

<style scoped>
.positions {
  border: 1px solid var(--line);
  background: var(--panel);
}
.head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 18px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--line);
}
.head h2 {
  margin: 0 auto 0 0;
}
.check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  cursor: pointer;
}
.check input {
  width: 15px;
  height: 15px;
  margin: 0;
  accent-color: var(--acc);
}
.message {
  margin: 0;
  padding: 12px;
}
.error {
  color: var(--neg);
}
.scroll {
  overflow-x: auto;
}
table {
  width: 100%;
  min-width: 900px;
  border-collapse: collapse;
}
th, td {
  padding: 7px 12px;
  text-align: left;
  font-weight: inherit;
  white-space: nowrap;
}
td.num, th.num {
  text-align: right;
}
td {
  font-family: var(--cond);
  font-size: 14px;
  border-bottom: 1px solid var(--dim);
}
thead th {
  padding-top: 0;
  padding-bottom: 0;
  border-bottom: 1px solid var(--line);
}
.sort {
  width: 100%;
  height: 34px;
  padding: 0;
  border: 0;
  background: transparent;
  text-align: inherit;
  color: var(--muted);
  font-family: var(--cond);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .06em;
  text-transform: uppercase;
  cursor: pointer;
}
.sort.active {
  color: var(--fg);
}
.arrow {
  margin-left: 4px;
}
.group > * {
  padding: 6px 12px;
  background: var(--sub);
  border-bottom: 1px solid var(--line);
  font-family: var(--cond);
  font-size: 13px;
  font-weight: 600;
}
.toggle {
  padding: 0;
  border: 0;
  background: transparent;
  font-weight: 600;
  cursor: pointer;
}
.security {
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ticker {
  display: inline-block;
  min-width: 48px;
  margin-right: 10px;
  font-weight: 700;
}
.name {
  font-family: var(--sans);
  font-size: 13px;
  color: var(--muted);
}
.strong {
  font-weight: 600;
}
.closed {
  opacity: 0.6;
}
.share {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.share-bar {
  display: inline-block;
  width: 40px;
  height: 4px;
  background: var(--dim);
}
.share-bar > span {
  display: block;
  height: 100%;
  background: var(--fg);
}
.dashed td {
  border-bottom-style: dashed;
  font-family: var(--sans);
  font-size: 13px;
}
.dashed td.num {
  font-family: var(--cond);
  font-size: 14px;
}
.badge {
  margin-left: 8px;
  padding: 1px 6px;
  border: 1px solid var(--pos);
  color: var(--pos);
  font-family: var(--cond);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .06em;
  text-transform: uppercase;
}
tfoot th, tfoot td {
  padding: 10px 12px;
  border-bottom: 0;
  font-family: var(--cond);
  font-size: 16px;
  font-weight: 700;
}
</style>
