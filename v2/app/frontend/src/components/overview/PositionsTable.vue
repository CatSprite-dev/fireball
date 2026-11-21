<script setup lang="ts">
import {computed, reactive, ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Position, PositionTable} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'

const props = defineProps<{accountIDs: string[] | null; period: string}>()

type ColumnKey = 'ticker' | 'quantity' | 'value' | 'invested' | 'income' | 'profit' | 'profitPct' | 'share'
type Column = {key: ColumnKey; title: string}

const columns: Column[] = [
  {key: 'ticker', title: 'Бумага'},
  {key: 'quantity', title: 'Кол-во'},
  {key: 'value', title: 'Стоимость'},
  {key: 'invested', title: 'Вложено'},
  {key: 'income', title: 'Выплаты'},
  {key: 'profit', title: 'Прибыль ₽'},
  {key: 'profitPct', title: 'Прибыль %'},
  {key: 'share', title: 'Доля'},
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
    sortDesc.value = true
  }
}
function sortIcon(key: ColumnKey) {
  if (sortKey.value !== key) return '↕'
  return sortDesc.value ? '↓' : '↑'
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
const grouped = computed(() =>
  groups
    .map(g => {
      const total = table.value?.groups?.find(t => t.kind === g.kind)
      const rows = visible.value.filter(p => p.kind === g.kind)
      return {...g, rows, value: total?.value ?? '', profit: total?.profit ?? ''}
    })
    .filter(g => g.rows.length > 0),
)

const unexplainedHelp =
  'Брокер не отдаёт через API некоторые операции (например, налог при закрытии ИИС). ' +
  'Мы видим разницу в деньгах на счёте и вычитаем её, чтобы цифры сходились с брокером'

const money = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})
const quantity = new Intl.NumberFormat('ru-RU', {maximumFractionDigits: 6})
function show(s: string, format = money): string {
  return s === '' ? '—' : format.format(Number(s))
}
</script>

<template>
  <section>
    <div class="head">
      <h2>Бумаги</h2>
      <label><input type="checkbox" v-model="showClosed" /> закрытые</label>
      <label><input type="checkbox" v-model="noGroups" /> без групп</label>
    </div>

    <p v-if="error">{{ error }}</p>
    <p v-else-if="!table">Загружаю…</p>

    <table v-else>
      <thead>
        <tr>
          <th v-for="c in columns" :key="c.key" :class="{num: c.key !== 'ticker'}">
            <button class="sort" @click="sortBy(c.key)">{{ c.title }} {{ sortIcon(c.key) }}</button>
            <HelpMark v-if="c.key !== 'ticker'" />
          </th>
        </tr>
      </thead>

      <template v-for="g in noGroups ? [{kind: 'all', title: '', rows: visible, value: '', profit: ''}] : grouped" :key="g.kind">
        <tbody>
          <tr v-if="!noGroups" class="group">
            <th>
              <button class="toggle" @click="open[g.kind] = !open[g.kind]">{{ open[g.kind] ? '▾' : '▸' }} {{ g.title }}</button>
            </th>
            <td></td>
            <td class="num">{{ show(g.value) }}</td>
            <td></td>
            <td></td>
            <td class="num">{{ show(g.profit) }}</td>
            <td></td>
            <td></td>
          </tr>
          <template v-if="noGroups || open[g.kind]">
            <tr v-for="p in g.rows" :key="p.key" :class="{closed: p.closed}">
              <td>
                {{ p.ticker }}
                <small v-if="p.name"> {{ p.name }}</small>
              </td>
              <td class="num">{{ show(p.quantity, quantity) }}</td>
              <td class="num">{{ show(p.value) }}</td>
              <td class="num">{{ show(p.invested) }}</td>
              <td class="num">{{ show(p.income) }}</td>
              <td class="num">{{ show(p.profit) }}</td>
              <td class="num">{{ show(p.profitPct) }}</td>
              <td class="num">{{ show(p.share) }}</td>
            </tr>
          </template>
        </tbody>
      </template>

      <tbody>
        <tr class="group">
          <th :colspan="columns.length">На уровне счёта</th>
        </tr>
        <tr>
          <td :colspan="5">Комиссии и налоги без бумаги</td>
          <td class="num">{{ show(table.accountLevel) }}</td>
          <td colspan="2"></td>
        </tr>
        <tr>
          <td colspan="2">Необъяснённое расхождение <HelpMark :text="unexplainedHelp" /></td>
          <td class="num">{{ show(table.unexplained) }}</td>
          <td colspan="2"></td>
          <td class="num">{{ show(table.unexplainedProfit) }}</td>
          <td colspan="2"></td>
        </tr>
      </tbody>

      <tfoot>
        <tr>
          <th>Итого</th>
          <td></td>
          <td class="num">{{ show(table.totalValue) }}</td>
          <td></td>
          <td></td>
          <td class="num">{{ show(table.totalProfit) }}</td>
          <td colspan="2"></td>
        </tr>
      </tfoot>
    </table>
  </section>
</template>

<style scoped>
.head {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.head h2 {
  margin: 0 auto 0 0;
}
table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 0.5rem;
}
th, td {
  padding: 0.25rem 0.5rem;
  text-align: left;
}
.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}
thead th, tfoot th, tfoot td {
  border-bottom: 1px solid;
}
tfoot th, tfoot td {
  border-top: 1px solid;
}
.group > * {
  padding-top: 0.75rem;
}
.closed {
  opacity: 0.6;
}
.sort, .toggle {
  all: unset;
  cursor: pointer;
}
</style>
