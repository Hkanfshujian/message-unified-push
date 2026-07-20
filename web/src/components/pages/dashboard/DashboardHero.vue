<script setup lang="ts">
import type { DashboardHealthSummary } from '@/types/business'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.dashboard

defineProps<{
  summary: DashboardHealthSummary
  loading?: boolean
  actionDisabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'primaryAction'): void
}>()
</script>

<template>
  <section class="dashboard-hero" :class="`dashboard-hero-${summary.status}`" aria-labelledby="dashboard-hero-title">
    <div class="dashboard-hero-main">
      <h1 id="dashboard-hero-title" class="dashboard-hero-title">{{ summary.title }}</h1>
      <p class="dashboard-hero-desc">{{ summary.description }}</p>
      <div class="dashboard-hero-meta" :aria-label="messages.statisticsScope">
        <span>{{ messages.statisticsRange }}{{ summary.rangeLabel }}</span>
        <span>{{ messages.lastUpdated }}{{ summary.lastUpdatedAt || messages.waitingRefresh }}</span>
      </div>
    </div>
    <div class="dashboard-hero-actions">
      <el-button v-if="summary.primaryAction" type="primary" :disabled="actionDisabled" :loading="loading" @click="emit('primaryAction')">
        {{ summary.primaryAction.label }}
      </el-button>
    </div>
  </section>
</template>
