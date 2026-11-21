<script setup lang="ts">
// Выбранный период живёт в шапке: v-model с кодом периода для сервиса.
const selected = defineModel<string>({required: true})

const periods = [
  {code: '1m', label: '1М'},
  {code: '3m', label: '3М'},
  {code: '6m', label: '6М'},
  {code: '12m', label: '12М'},
  {code: '5y', label: '5 лет'},
  {code: 'all', label: 'Всё время'},
]
</script>

<template>
  <div class="seg periods" role="group" aria-label="Период">
    <button
      v-for="p in periods"
      :key="p.code"
      type="button"
      :aria-pressed="p.code === selected"
      @click="selected = p.code"
    >
      {{ p.label }}
    </button>
  </div>
</template>

<style scoped>
.periods button {
  height: 36px;
  min-width: 42px;
  padding: 0 10px;
  font-size: 14px;
  font-family: var(--cond);
  font-weight: 600;
}
</style>
