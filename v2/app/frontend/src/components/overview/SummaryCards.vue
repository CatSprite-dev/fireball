<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Overview} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'

const props = defineProps<{accountIDs: string[] | null; period: string}>()

// «Вложено» и «Прибыль» — за выбранный период.
const during = computed(() => (props.period === 'all' ? 'за всё время' : 'за период'))

const overview = ref<Overview | null>(null)
const error = ref('')

// Деньги приходят строкой; в число — только для показа, считать с ним нельзя.
const rub = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})

// Пока нет данных или строка пустая — прочерк.
function money(s: string | undefined): string {
  return s ? `${rub.format(Number(s))} ₽` : '— ₽'
}

// Ответы могут прийти не по порядку, если галочки щёлкают быстро: показываем только последний запрос.
let request = 0

watch([() => props.accountIDs, () => props.period], async ([ids, period]) => {
  if (ids === null) return
  const current = ++request
  overview.value = null
  error.value = ''
  try {
    const result = await PortfolioService.Overview(ids, period)
    if (current === request) overview.value = result
  } catch (e) {
    if (current === request) error.value = e instanceof Error ? e.message : String(e)
  }
})
</script>

<template>
  <div class="cards">
    <section class="card">
      <h3>Стоимость <HelpMark /></h3>
      <p v-if="error">{{ error }}</p>
      <p v-else-if="!overview">Загружаю…</p>
      <template v-else>
        <p class="value">{{ money(overview.value) }}</p>
        <p>на {{ overview.asOf }}</p>
      </template>
    </section>

    <section class="card">
      <h3>Вложено {{ during }} <HelpMark /></h3>
      <p class="value">{{ money(overview?.invested) }}</p>
    </section>

    <section class="card">
      <h3>Прибыль {{ during }} <HelpMark /></h3>
      <p class="value">{{ money(overview?.profit) }}</p>
      <p>
        сделки: зафикс. {{ money(overview?.realized) }} <HelpMark /> /
        бумажная {{ money(overview?.paper) }} <HelpMark />
      </p>
      <p>выплаты: {{ money(overview?.income) }} <HelpMark /></p>
    </section>
  </div>
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
h3, p {
  margin: 0 0 0.25rem;
}
.value {
  font-size: 1.5rem;
}
</style>
