<script setup lang="ts">
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {PortfolioService} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import type {Account} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'

// Выбранные счета живут в шапке: фильтр только сообщает о каждом изменении.
const emit = defineEmits<{change: [accountIDs: string[]]}>()

const accounts = ref<Account[]>([])
const selected = ref<string[]>([])
const error = ref('')
const open = ref(false)
const root = ref<HTMLElement | null>(null)

const opened = computed(() => accounts.value.filter(a => !a.closed))
const closed = computed(() => accounts.value.filter(a => a.closed))

// Закрытые счета по умолчанию не отмечены.
function allOpen() {
  selected.value = opened.value.map(a => a.id)
}

onMounted(async () => {
  document.addEventListener('mousedown', closeOutside)
  try {
    accounts.value = (await PortfolioService.Accounts()) ?? []
    allOpen()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
})
onBeforeUnmount(() => document.removeEventListener('mousedown', closeOutside))

function closeOutside(e: MouseEvent) {
  if (open.value && root.value && !root.value.contains(e.target as Node)) open.value = false
}

watch(selected, ids => emit('change', [...ids]))

// «3 счёта», «5 счетов», «21 счёт».
function plural(n: number): string {
  const tens = n % 100, ones = n % 10
  if (tens >= 11 && tens <= 14) return `${n} счетов`
  if (ones === 1) return `${n} счёт`
  if (ones >= 2 && ones <= 4) return `${n} счёта`
  return `${n} счетов`
}

const summary = computed(() => {
  if (error.value) return 'Ошибка'
  if (accounts.value.length === 0) return 'Загружаю…'
  if (selected.value.length === 0) return 'Не выбраны'
  if (selected.value.length === 1) return accounts.value.find(a => a.id === selected.value[0])?.name ?? ''
  return plural(selected.value.length)
})
</script>

<template>
  <div ref="root" class="filter" @keydown.esc="open = false">
    <button type="button" class="toggle" aria-haspopup="true" :aria-expanded="open" :title="error || undefined" @click="open = !open">
      <span class="label">Счета</span>
      <span class="summary">{{ summary }}</span>
      <span v-if="accounts.length" class="count">из {{ accounts.length }}</span>
      <svg width="10" height="10" viewBox="0 0 12 12" aria-hidden="true">
        <path d="M2.5 4.5l3.5 3.5 3.5-3.5" fill="none" stroke="var(--fg)" stroke-width="1.6" />
      </svg>
    </button>

    <div v-if="open" class="popup">
      <div class="head">
        <span class="label">Открытые</span>
        <span class="actions">
          <button type="button" @click="allOpen">Все открытые</button>
          <button type="button" @click="selected = []">Сбросить</button>
        </span>
      </div>
      <label v-for="a in opened" :key="a.id" class="item">
        <input type="checkbox" :value="a.id" v-model="selected" />
        <span>{{ a.name }}</span>
      </label>

      <template v-if="closed.length">
        <div class="label section">Закрытые · не учитываются, пока не отмечены</div>
        <label v-for="a in closed" :key="a.id" class="item muted">
          <input type="checkbox" :value="a.id" v-model="selected" />
          <span>{{ a.name }}</span>
        </label>
      </template>
      <div class="bottom"></div>
    </div>
  </div>
</template>

<style scoped>
.filter {
  position: relative;
}
.toggle {
  height: 36px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  border: 1px solid var(--line);
  background: transparent;
  cursor: pointer;
}
.summary {
  font-weight: 600;
  white-space: nowrap;
}
.count {
  font-family: var(--cond);
  font-size: 12px;
  color: var(--muted);
}
.popup {
  position: absolute;
  top: 40px;
  right: 0;
  z-index: 40;
  width: 320px;
  background: var(--panel);
  border: 1px solid var(--fg);
  box-shadow: 0 16px 40px rgba(0, 0, 0, .35);
}
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--line);
}
.actions {
  display: flex;
  gap: 6px;
}
.actions button {
  height: 26px;
  padding: 0 8px;
  border: 1px solid var(--line);
  background: transparent;
  font-size: 12px;
  cursor: pointer;
}
.item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 12px;
  cursor: pointer;
}
.item input {
  width: 15px;
  height: 15px;
  margin: 0;
  accent-color: var(--fg);
}
.section {
  padding: 8px 12px 4px;
  border-top: 1px solid var(--line);
}
.bottom {
  height: 6px;
}
</style>
