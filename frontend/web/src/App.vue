<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useTheme } from '../src/composables/useTheme'

const { isDark, toggleTheme } = useTheme()
</script>

<template>
    <RouterView />

    <button class="theme-btn" @click="toggleTheme" title="Toggle theme">
      <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="18" height="18"
        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="4"/>
        <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/>
      </svg>
      <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18"
        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>
      </svg>
    </button>
</template>

<style scoped>
.theme-btn {
  position: fixed;
  bottom: 16px;
  right: 16px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: var(--primary);
  color: var(--primary-foreground);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgb(0, 0, 0 / 0.2);
  transition: opacity 0.15s;
}

.theme-btn:hover {
  opacity: 0.9;
}
</style>