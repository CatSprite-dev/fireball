<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {PortfolioService} from '../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Overview} from '../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'

const overview = ref<Overview | null>(null)
const error = ref('')

// Деньги приходят строкой; в число — только для показа, считать с ним нельзя.
const rub = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})

onMounted(async () => {
  try {
    overview.value = await PortfolioService.Overview()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
})
</script>

<template>
  <main class="overview">
    <p v-if="error" class="error">{{ error }}</p>
    <p v-else-if="!overview" class="muted">Загружаю…</p>
    <template v-else>
      <h1 class="value">{{ rub.format(Number(overview.value)) }} ₽</h1>
      <p class="muted">на {{ overview.asOf }}</p>
    </template>
  </main>
</template>

<style scoped>
.overview {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem;
}
.value {
  margin: 0;
  font-size: 3.5rem;
  font-variant-numeric: tabular-nums;
}
.muted {
  margin: 0;
  color: var(--muted);
}
.error {
  max-width: 40rem;
  color: var(--error);
  -webkit-user-select: text;
  user-select: text;
}
</style>
