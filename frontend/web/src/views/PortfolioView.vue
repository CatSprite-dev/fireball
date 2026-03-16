<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePortfolioStore } from '../stores/portfolio'
import { useAuthStore } from '../stores/auth'
import MetricsGrid from '../components/MetricsGrid.vue'
import OverviewTab from '../components/OverviewTab.vue'
import HoldingsTab from '../components/HoldingsTab.vue'

const router = useRouter()
const portfolio = usePortfolioStore()
const auth = useAuthStore()

const activeTab = ref<'overview' | 'holdings'>('overview')

function logout() {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  portfolio.load()
})
</script>

<template>
  <div class="page">

    <div class="header">
      <div>
        <h1>Investment Fireball</h1>
        <p>Track, analyze, and grow your investment portfolio</p>
      </div>
      <button class="logout-btn" @click="logout" title="Logout">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"
          fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
          <polyline points="16 17 21 12 16 7"/>
          <line x1="21" y1="12" x2="9" y2="12"/>
        </svg>
      </button>
    </div>

    <div v-if="portfolio.isLoading" class="state-message">
      Loading portfolio...
    </div>

    <div v-else-if="portfolio.error" class="state-message error">
      {{ portfolio.error }}
    </div>

    <template v-else>
      <MetricsGrid :metrics="portfolio.metrics" />

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

        <OverviewTab v-if="activeTab==='overview'" :investments="portfolio.investments"/>
        <HoldingsTab v-if="activeTab==='holdings'" :investments="portfolio.investments"/>
      </div>
    </template>

  </div>
</template>

<style scoped>
.page {
  max-width: 1200px;
  margin: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header h1 {
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 4px;
}

.header p {
  color: var(--muted-foreground);
  margin: 0;
}

.logout-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  color: var(--muted-foreground);
}

.logout-btn:hover {
  background: var(--muted);
}

.state-message {
  text-align: center;
  padding: 48px;
  color: var(--muted-foreground);
}

.state-message.error {
  color: #dc2626;
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
</style>