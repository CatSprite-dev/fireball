<script setup lang="ts">
import {computed} from 'vue'
import type {Performance} from '../../../bindings/github.com/CatSprite-dev/fireball/v2/internal/ui'
import HelpMark from '../HelpMark.vue'
import {day, signedPercent, tone} from '../../format'

// Цифры страница передаёт готовыми.
const props = defineProps<{index: string; performance: Performance | null; error: string}>()
defineEmits<{more: []}>()

// Полоса «против индекса»: от середины вправо — лучше индекса, влево — хуже.
// Полная половина — 20 п.п. и больше; это только картинка, цифра рядом точная.
const vsBar = computed(() => {
  const n = Number(props.performance?.vsIndex ?? 0)
  const width = Math.min(Math.abs(n) / 20, 1) * 50
  return {left: n >= 0 ? '50%' : `${50 - width}%`, width: `${width}%`, tone: tone(props.performance?.vsIndex)}
})
</script>

<template>
  <div class="column">
    <div class="row">
      <span class="label">Сравнивать с</span>
      <span class="chip" title="Индекс МосБиржи полной доходности брутто">{{ index }}</span>
    </div>

    <p v-if="error" class="row error">{{ error }}</p>

    <div class="row">
      <span class="caption">
        <span class="label">Портфель, TWR <HelpMark text="Доходность портфеля без влияния пополнений и выводов" /></span>
        <span class="note muted">с {{ day(performance?.from) }}</span>
      </span>
      <span class="number" :class="tone(performance?.twr)">{{ signedPercent(performance?.twr) }}</span>
    </div>

    <div class="row">
      <span class="caption">
        <span class="label">Индекс {{ index }}</span>
        <span v-if="performance?.compareFrom" class="note muted">с {{ day(performance.compareFrom) }}</span>
      </span>
      <span class="number index">{{ signedPercent(performance?.index) }}</span>
    </div>

    <div class="row stacked">
      <div class="line">
        <span class="caption">
          <span class="label">Против индекса <HelpMark text="TWR минус рост индекса за тот же отрезок" /></span>
          <span v-if="performance?.compareFrom" class="note muted">индекс есть с {{ day(performance.compareFrom) }}, сравнение с этого дня</span>
        </span>
        <span class="number strong" :class="vsBar.tone">{{ signedPercent(performance?.vsIndex, ' п.п.') }}</span>
      </div>
      <div class="vs" aria-hidden="true">
        <div class="vs-fill" :class="vsBar.tone" :style="{left: vsBar.left, width: vsBar.width}"></div>
        <div class="vs-mid"></div>
      </div>
    </div>

    <p class="row note muted">Индекс полной доходности брутто — без налогов, обычно на ~1 п.п. в год выше реально достижимого.</p>

    <button type="button" class="more" @click="$emit('more')">
      <span class="label">Другие показатели</span>
      <svg width="12" height="12" viewBox="0 0 12 12" aria-hidden="true">
        <path d="M4.5 2.5l3.5 3.5-3.5 3.5" fill="none" stroke="var(--fg)" stroke-width="1.6" />
      </svg>
    </button>
  </div>
</template>

<style scoped>
.column {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: var(--line);
}
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 0;
  padding: 10px 16px;
  background: var(--panel);
}
.stacked {
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
}
.line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.row:not(.note):not(.error):not(:first-child) {
  flex: 1;
}
.caption {
  display: flex;
  flex-direction: column;
}
.note {
  font-size: 12px;
}
.error {
  color: var(--neg);
}
.chip {
  height: 26px;
  display: flex;
  align-items: center;
  padding: 0 7px;
  border: 1px solid var(--acc);
  background: var(--acc);
  color: var(--acc-fg);
  font-family: var(--cond);
  font-size: 12px;
  font-weight: 600;
}
.number {
  font-family: var(--cond);
  font-size: 30px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}
.index {
  color: var(--acc);
}
.strong {
  font-weight: 700;
}
.vs {
  position: relative;
  height: 6px;
  background: var(--dim);
}
.vs-fill {
  position: absolute;
  top: 0;
  bottom: 0;
  background: var(--muted);
}
.vs-fill.pos {
  background: var(--pos);
}
.vs-fill.neg {
  background: var(--neg);
}
.vs-mid {
  position: absolute;
  left: 50%;
  top: -4px;
  bottom: -4px;
  width: 1px;
  background: var(--fg);
}
.more {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 0 16px;
  border: 0;
  background: var(--panel);
  cursor: pointer;
  text-align: left;
}
.more:hover .label {
  color: var(--fg);
}
</style>
