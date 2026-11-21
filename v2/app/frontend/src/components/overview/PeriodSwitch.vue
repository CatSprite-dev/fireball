<script setup lang="ts">
import {ref} from 'vue'

const periods = ['1М', '3М', '6М', '12М', '5 лет', 'Всё время']
const selected = ref('12М')
</script>

<template>
  <div class="row">
    <span>Период:</span>
    <button
      v-for="p in periods"
      :key="p"
      :aria-pressed="p === selected"
      :class="{selected: p === selected}"
      @click="selected = p"
    >
      {{ p }}
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
