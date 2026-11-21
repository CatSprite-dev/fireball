<script setup lang="ts">
// Значок подсказки у цифры; текст показывается при наведении.
defineProps<{text?: string}>()
</script>

<template>
  <abbr class="help" :title="text">?</abbr>
</template>

<style scoped>
.help {
  margin-left: 4px;
  padding: 0 4px;
  border: 1px solid var(--line);
  text-decoration: none;
  letter-spacing: 0;
  cursor: help;
}
</style>
