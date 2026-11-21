<script setup lang="ts">
import {ref} from 'vue'
import OverviewPage from './components/overview/OverviewPage.vue'

const tabs = ['Обзор', 'Пропорции', 'Как мы считаем'] as const
const tab = ref<(typeof tabs)[number]>('Обзор')
</script>

<template>
  <div class="drag" aria-hidden="true"></div>
  <div class="app">
    <nav class="tabs">
      <button
        v-for="t in tabs"
        :key="t"
        :aria-pressed="t === tab"
        :class="{selected: t === tab}"
        @click="tab = t"
      >
        {{ t }}
      </button>
    </nav>

    <OverviewPage v-if="tab === 'Обзор'" />
    <section v-else-if="tab === 'Пропорции'">
      <h1>Пропорции</h1>
    </section>
    <section v-else>
      <h1>Как мы считаем</h1>
    </section>
  </div>
</template>

<style scoped>
/* Полоса сверху под кнопками окна macOS: высота как InvisibleTitleBarHeight в app/main.go.
   Окно тянется за неё; остальное — кнопки и галочки, их таскать нельзя. */
.drag {
  height: 50px;
  --wails-draggable: drag;
}
.app {
  --wails-draggable: no-drag;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0 1.25rem 1.25rem;
}
.tabs {
  display: flex;
  gap: 0.5rem;
}
.selected {
  font-weight: bold;
  text-decoration: underline;
}
</style>
