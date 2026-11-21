<script setup lang="ts">
import {ref} from 'vue'
import OverviewPage from './components/overview/OverviewPage.vue'
import AccountsFilter from './components/overview/AccountsFilter.vue'
import PeriodSwitch from './components/overview/PeriodSwitch.vue'
import {day} from './format'

const tabs = ['Обзор', 'Пропорции', 'Как считается'] as const
const tab = ref<(typeof tabs)[number]>('Обзор')

// Счета и период выбираются в шапке и действуют на весь «Обзор».
// null — счета ещё не загружены, выбора пока нет.
const accountIDs = ref<string[] | null>(null)
const period = ref('12m')
// День, на который загружены данные: приходит с карточками «Обзора».
const asOf = ref('')
</script>

<template>
  <!-- Шапка — полоса под кнопками окна macOS: окно тянется за её фон, кнопки и списки таскать нельзя. -->
  <header class="header">
    <div class="cell brand">
      <svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="15" cy="12" r="6.5" fill="var(--me)" />
        <path d="M1.5 9h6M3.5 12h4M1.5 15h6" stroke="var(--me)" stroke-width="2.2" stroke-linecap="square" />
      </svg>
      <span class="name">FIREBALL OUTCOME</span>
    </div>

    <nav class="cell tabs" aria-label="Разделы">
      <button v-for="t in tabs" :key="t" type="button" :aria-pressed="t === tab" @click="tab = t">{{ t }}</button>
    </nav>

    <div class="cell spacer"></div>

    <div class="cell control">
      <AccountsFilter @change="ids => (accountIDs = ids)" />
    </div>

    <div class="cell control">
      <PeriodSwitch v-model="period" />
    </div>

    <div class="cell control">
      <span class="as-of">
        <span class="label">Данные на</span>
        <span class="date">{{ day(asOf) }}</span>
      </span>
      <!-- Обновление данных появится вместе с базой. -->
      <button type="button" class="icon" aria-label="Обновить данные" title="Обновить данные: пока не работает">
        <svg width="14" height="14" viewBox="0 0 16 16" aria-hidden="true">
          <path d="M13.5 8a5.5 5.5 0 1 1-1.6-3.9M13.5 2.5v3h-3" fill="none" stroke="var(--fg)" stroke-width="1.6" stroke-linecap="square" />
        </svg>
      </button>
    </div>
  </header>

  <main class="main">
    <OverviewPage v-if="tab === 'Обзор'" :accountIDs="accountIDs" :period="period" @as-of="d => (asOf = d)" />
    <div v-else class="empty">Раздел «{{ tab }}» пока не готов.</div>
  </main>
</template>

<style scoped>
/* Ячейки шапки разделены линией в 1px: фон шапки — цвет линии, у ячеек — фон блока. */
.header {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: 1px;
  background: var(--line);
  border-bottom: 1px solid var(--line);
  --wails-draggable: drag;
}
.cell {
  display: flex;
  align-items: center;
  min-height: 52px;
  background: var(--panel);
}
.cell button, .control {
  --wails-draggable: no-drag;
}
/* Слева место под кнопки окна macOS. */
.brand {
  gap: 10px;
  padding: 0 18px 0 84px;
}
.name {
  font-family: var(--cond);
  font-weight: 700;
  font-size: 19px;
  letter-spacing: .08em;
  white-space: nowrap;
}
.tabs button {
  height: 100%;
  padding: 0 16px;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--muted);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.tabs button[aria-pressed="true"] {
  border-bottom-color: var(--fg);
  color: var(--fg);
}
.spacer {
  flex: 1 1 0;
}
.control {
  gap: 12px;
  padding: 8px 14px;
}
.as-of {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.date {
  font-family: var(--cond);
  font-size: 15px;
  font-weight: 600;
}
.icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--line);
  background: transparent;
  cursor: pointer;
}
.main {
  --wails-draggable: no-drag;
  padding: 14px;
}
.empty {
  border: 1px dashed var(--line);
  padding: 40px;
  text-align: center;
  color: var(--muted);
}
</style>
