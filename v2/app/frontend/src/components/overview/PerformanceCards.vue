<script setup lang="ts">
import {ref} from 'vue'
import HelpMark from '../HelpMark.vue'

const indexes = ['MCFTR', 'MCFTRR']
const index = ref('MCFTR')
</script>

<template>
  <div class="cards">
    <section class="card">
      <h3>TWR <HelpMark /></h3>
      <p class="value">—%</p>
    </section>

    <section class="card">
      <h3>MWR, годовых <HelpMark /></h3>
      <p class="value">—%</p>
    </section>

    <section class="card">
      <h3>Индекс <HelpMark /></h3>
      <select v-model="index">
        <option v-for="i in indexes" :key="i" :value="i">{{ i }}</option>
      </select>
      <p class="value">—%</p>
    </section>

    <section class="card">
      <h3>Против индекса <HelpMark /></h3>
      <p class="value">— п.п.</p>
    </section>
  </div>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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
