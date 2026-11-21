<script setup lang="ts">
import {onBeforeUnmount, onMounted, ref} from 'vue'
import type {Performance} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'
import {signedPercent, tone} from '../../format'

// Панель справа поверх окна: показатели, которым не место на главном экране.
defineProps<{performance: Performance | null}>()
const emit = defineEmits<{close: []}>()

const closeButton = ref<HTMLButtonElement | null>(null)

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => {
  document.addEventListener('keydown', onKey)
  closeButton.value?.focus()
})
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="backdrop" @click="emit('close')"></div>
  <aside class="panel" role="dialog" aria-modal="true" aria-labelledby="more-title">
    <div class="head">
      <h2 id="more-title" class="label">Другие показатели</h2>
      <button ref="closeButton" type="button" class="close" aria-label="Закрыть" @click="emit('close')">
        <svg width="12" height="12" viewBox="0 0 12 12" aria-hidden="true">
          <path d="M2 2l8 8M10 2l-8 8" stroke="var(--fg)" stroke-width="1.6" />
        </svg>
      </button>
    </div>

    <div class="metric">
      <span class="label">MWR, годовых <HelpMark text="Доходность на ваши реальные деньги с учётом сроков пополнений и выводов, в пересчёте на год" /></span>
      <span class="number" :class="tone(performance?.mwr)">{{ signedPercent(performance?.mwr) }}</span>
      <span class="note muted">на ваши деньги</span>
    </div>
  </aside>
</template>

<style scoped>
.backdrop {
  position: fixed;
  inset: 0;
  z-index: 50;
  background: rgba(0, 0, 0, .45);
}
.panel {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 51;
  width: min(360px, 100%);
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: var(--line);
  border-left: 1px solid var(--line);
  box-shadow: -16px 0 40px rgba(0, 0, 0, .35);
  --wails-draggable: no-drag;
}
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 50px;
  padding: 0 12px 0 16px;
  background: var(--panel);
}
h2 {
  margin: 0;
}
.close {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--line);
  background: transparent;
  cursor: pointer;
}
.metric {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 16px;
  background: var(--panel);
}
.metric:last-child {
  flex: 1;
}
.number {
  font-family: var(--cond);
  font-size: 28px;
  font-weight: 600;
  line-height: 1;
}
.note {
  font-size: 12px;
}
</style>
