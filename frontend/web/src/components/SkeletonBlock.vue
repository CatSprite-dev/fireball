<script setup lang="ts">
defineProps<{
    width?: string
    height?: string
    radius?: string
}>()
</script>

<template>
    <div
        class="skeleton"
        :style="{
            width: width ?? '100px',
            height: height ?? '16px',
            borderRadius: radius ?? '6px',
        }">
    </div>
</template>

<style scoped>
.skeleton {
    background: var(--muted);
    position: relative;
    overflow: hidden;
}

.skeleton::after {
    content: '';
    position: absolute;
    inset: 0;
    transform: translateX(-100%);
    background: linear-gradient(
        90deg,
        transparent,
        rgb(255 255 255 / 0.4),
        transparent
    );
    animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
    100% { transform: translateX(100%); }
}
</style>