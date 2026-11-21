<script setup lang="ts">
import type {Performance} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'

// Выбранный индекс живёт на странице: v-model. Цифры страница передаёт готовыми.
const index = defineModel<string>('index', {required: true})
defineProps<{performance: Performance | null; error: string}>()

const indexes = ['MCFTR']

const number = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})

// Пока нет данных или строка пустая — прочерк.
function percent(s: string | undefined, unit = '%'): string {
  return s ? `${number.format(Number(s))}${unit}` : `—${unit}`
}
</script>

<template>
  <p v-if="error" class="error">{{ error }}</p>
  <div class="cards">
    <section class="card">
      <h3>TWR <HelpMark /></h3>
      <p class="value">{{ percent(performance?.twr) }}</p>
      <p v-if="performance">с {{ performance.from }}</p>
    </section>

    <section class="card">
      <h3>Индекс <HelpMark /></h3>
      <select v-model="index">
        <option v-for="i in indexes" :key="i" :value="i">{{ i }}</option>
      </select>
      <p class="value">{{ percent(performance?.index) }}</p>
    </section>

    <section class="card">
      <h3>Против индекса <HelpMark /></h3>
      <p class="value">{{ percent(performance?.vsIndex, ' п.п.') }}</p>
      <p v-if="performance?.compareFrom">индекс есть с {{ performance.compareFrom }}, сравнение с этого дня</p>
    </section>
  </div>

  <details class="more">
    <summary>Подробнее</summary>
    <section class="card">
      <h3>MWR, годовых <HelpMark /></h3>
      <p class="value">{{ percent(performance?.mwr) }}</p>
    </section>
  </details>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}
.card {
  border: 1px solid;
  padding: 0.75rem;
}
.more .card {
  margin-top: 0.5rem;
  max-width: 20rem;
}
h3, p {
  margin: 0 0 0.25rem;
}
.value {
  font-size: 1.5rem;
}
</style>
