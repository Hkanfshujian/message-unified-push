<script setup lang="ts">
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.dashboard

defineProps<{
  loading: boolean
  error?: string
  isEmpty: boolean
  rangeLabel: string
  insightLabel: string
}>()
</script>

<template>
  <el-card class="dashboard-panel dashboard-trend-card" body-class="!pt-2">
    <template #header>
      <div class="dashboard-panel-header">
        <div>
          <div class="dashboard-panel-title">{{ messages.sendTrend }}</div>
        </div>
      </div>
    </template>
    <div class="dashboard-chart-panel relative" role="img" :aria-label="`${rangeLabel}${messages.trendChartSuffix}${insightLabel}`">
      <slot />
      <AppEmptyState v-if="error && !loading" class="absolute inset-0 bg-card/80" :title="messages.trendLoadFailed" :description="error" />
      <AppEmptyState v-else-if="isEmpty && !loading" class="absolute inset-0 bg-card/80" :description="messages.noTrendData" />
    </div>
  </el-card>
</template>
