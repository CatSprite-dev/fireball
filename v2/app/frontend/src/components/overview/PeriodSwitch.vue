<script setup lang="ts">
// Выбранный период живёт на странице: v-model с кодом периода для сервиса.
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
  <div class="row">
    <span>Период:</span>
    <button
      v-for="p in periods"
      :key="p.code"
      :aria-pressed="p.code === selected"
      :class="{selected: p.code === selected}"
      @click="selected = p.code"
    >
      {{ p.label }}
    </button>
  </div>
</template>

<style scoped>
.row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}
.selected {
  font-weight: bold;
  text-decoration: underline;
}
</style>
