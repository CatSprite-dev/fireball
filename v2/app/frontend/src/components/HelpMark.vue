<script setup lang="ts">
// Значок подсказки у цифры; текст показывается при наведении. У большинства цифр текста пока нет.
defineProps<{text?: string}>()
</script>

<template>
  <span class="help" aria-label="подсказка" :title="text">?</span>
</template>

<style scoped>
.help {
  margin-left: 0.25rem;
  cursor: help;
}
</style>
