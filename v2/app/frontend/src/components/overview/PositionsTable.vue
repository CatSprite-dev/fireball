<script setup lang="ts">
import {ref, reactive} from 'vue'
import HelpMark from '../HelpMark.vue'

type Column = {key: string; title: string}

const columns: Column[] = [
  {key: 'name', title: 'Бумага'},
  {key: 'quantity', title: 'Кол-во'},
  {key: 'value', title: 'Стоимость'},
  {key: 'invested', title: 'Вложено'},
  {key: 'payouts', title: 'Выплаты'},
  {key: 'profit', title: 'Прибыль ₽'},
  {key: 'profitPercent', title: 'Прибыль %'},
  {key: 'share', title: 'Доля'},
]
const numberColumns = columns.slice(1)

const groups = ['Фонды', 'Облигации', 'Акции', 'Фьючерсы', 'Опционы', 'Валюта']
const open = reactive<Record<string, boolean>>(Object.fromEntries(groups.map(g => [g, true])))

const showClosed = ref(false)
const noGroups = ref(false)

// Сортировка: повторный клик по той же колонке меняет направление.
const sortKey = ref('value')
const sortDesc = ref(true)
function sortBy(key: string) {
  if (sortKey.value === key) {
    sortDesc.value = !sortDesc.value
  } else {
    sortKey.value = key
    sortDesc.value = true
  }
}
function sortIcon(key: string) {
  if (sortKey.value !== key) return '↕'
  return sortDesc.value ? '↓' : '↑'
}

const accountLevel = ['Налоги без бумаги', 'Комиссии без бумаги', 'Необъяснённое расхождение']
</script>

<template>
  <section>
    <div class="head">
      <h2>Бумаги</h2>
      <label><input type="checkbox" v-model="showClosed" /> закрытые</label>
      <label><input type="checkbox" v-model="noGroups" /> без групп</label>
    </div>

    <table>
      <thead>
        <tr>
          <th v-for="c in columns" :key="c.key" :class="{num: c.key !== 'name'}">
            <button class="sort" @click="sortBy(c.key)">{{ c.title }} {{ sortIcon(c.key) }}</button>
            <HelpMark v-if="c.key !== 'name'" />
          </th>
        </tr>
      </thead>

      <tbody v-if="noGroups">
        <tr>
          <td>…</td>
          <td v-for="c in numberColumns" :key="c.key" class="num">—</td>
        </tr>
      </tbody>
      <template v-else>
        <tbody v-for="g in groups" :key="g">
          <tr>
            <th :colspan="columns.length" class="group">
              <button class="toggle" @click="open[g] = !open[g]">{{ open[g] ? '▾' : '▸' }} {{ g }}</button>
            </th>
          </tr>
          <tr v-if="open[g]">
            <td>…</td>
            <td v-for="c in numberColumns" :key="c.key" class="num">—</td>
          </tr>
        </tbody>
      </template>

      <tbody>
        <tr>
          <th :colspan="columns.length" class="group">На уровне счёта</th>
        </tr>
        <tr v-for="name in accountLevel" :key="name">
          <td :colspan="columns.length - 1">{{ name }}</td>
          <td class="num">—</td>
        </tr>
      </tbody>

      <tfoot>
        <tr>
          <th :colspan="columns.length - 1" class="group">Итого</th>
          <td class="num">—</td>
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
}
thead th, tfoot th, tfoot td {
  border-bottom: 1px solid;
}
tfoot th, tfoot td {
  border-top: 1px solid;
}
.group {
  padding-top: 0.75rem;
}
.sort, .toggle {
  all: unset;
  cursor: pointer;
}
</style>
