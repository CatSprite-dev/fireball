<script setup lang="ts">
import {ref, onMounted, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Account} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'

// Выбранные счета живут на странице: фильтр только сообщает о каждом изменении.
const emit = defineEmits<{change: [accountIDs: string[]]}>()

const accounts = ref<Account[]>([])
const selected = ref<string[]>([])
const error = ref('')

onMounted(async () => {
  try {
    accounts.value = (await PortfolioService.Accounts()) ?? []
    selected.value = accounts.value.map(a => a.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
})

watch(selected, ids => emit('change', [...ids]))
</script>

<template>
  <div class="row">
    <span>Счета:</span>
    <span v-if="error">{{ error }}</span>
    <span v-else-if="accounts.length === 0">Загружаю…</span>
    <label v-for="a in accounts" :key="a.id">
      <input type="checkbox" :value="a.id" v-model="selected" />
      {{ a.name }}<span v-if="a.closed"> (закрыт)</span>
    </label>
  </div>
</template>

<style scoped>
.row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 1rem;
}
</style>
