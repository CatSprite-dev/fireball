<script setup lang="ts">
import { ref } from 'vue'
import SkeletonBlock from './SkeletonBlock.vue'

const activeTab = ref<'overview' | 'holdings'>('overview')
</script>

<template>
    <div class="skeleton-page">

        <div class="metrics-grid">
            <div v-for="i in 4" :key="i" class="card">
                <div class="card-header">
                    <SkeletonBlock width="100px" height="14px"/>
                    <SkeletonBlock width="16px" height="16px"/>
                </div>
                <SkeletonBlock width="140px" height="32px"/>
                <SkeletonBlock width="80px" height="12px"/>
            </div>
        </div>

        <div class="tabs">
            <div class="tab-list">
                <button
                    class="tab-btn"
                    :class="{ active: activeTab === 'overview'}"
                    @click="activeTab = 'overview'"  
                >
                    Overview
                </button>
                <button
                    class="tab-btn"
                    :class="{ active: activeTab === 'holdings' }"
                    @click="activeTab = 'holdings'"
                >
                    Holdings
                </button>
            </div>

            <div v-if="activeTab === 'overview'" class="overview-skeleton">
                <div class="card" v-for="i in 2" :key="i">
                    <div class="card-header">
                        <SkeletonBlock width="140px" height="18px"/>
                        <SkeletonBlock width="80px" height="12px"/>
                    </div>
                    <div class="rows">
                        <div v-for="j in 4" :key="j" class="row">
                            <SkeletonBlock width="100px" height="14px"/>
                            <SkeletonBlock width="60px" height="8px" radius="999px"/>
                        </div>
                    </div>
                </div>
            </div>

            <div v-if="activeTab === 'holdings'" class="holdings-skeleton">
                <div class="card">
                    <div class="card-header">
                        <SkeletonBlock width="140px" height="18px"/>
                    </div>
                    <div class="rows">
                        <div v-for="i in 8">
                            <SkeletonBlock width="120px" height="14px"/>
                            <SkeletonBlock width="60px" height="14px"/>
                            <SkeletonBlock width="50px" height="14px"/>
                            <SkeletonBlock width="80px" height="14px"/>
                            <SkeletonBlock width="80px" height="14px"/>
                        </div>
                    </div>
                </div>
            </div>

        </div>
    </div>
</template>

<style scoped>
.skeleton-page {
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.metrics-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
}

@media (max-width: 900px) {
    .metrics-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 500px) {
    .metrics-grid {
        grid-template-columns: 1fr;
    }
}

.card {
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--card);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
}

.tabs {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.tab-list {
  display: inline-flex;
  background: var(--muted);
  border-radius: 8px;
  padding: 4px;
  gap: 2px;
}

.tab-btn {
  background: none;
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  color: var(--muted-foreground);
  transition: all 0.15s;
}

.tab-btn.active {
  background: var(--background);
  color: var(--foreground);
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.1);
}

.overview-skeleton {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
}

@media (max-width: 700px) {
    .overview-skeleton {
        grid-template-columns: 1fr;
    }
}

.rows {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 8px;
}

.row {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.holdings-skeleton .card {
    padding-bottom: 0;
}

.table-row {
    display: flex;
    gap: 24px;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid var(--border);
}

.table-row:last-child {
    border-bottom: none;
}
</style>