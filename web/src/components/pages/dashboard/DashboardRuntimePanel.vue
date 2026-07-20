<script setup lang="ts">
import DoraIcon from '@/components/ui/DoraIcon.vue'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.dashboard

defineProps<{
  siteTitle: string
  cards: Array<{
    label: string
    value: number
    suffix: string
    tone: string
  }>
}>()
</script>

<template>
  <el-card class="dashboard-panel dashboard-runtime-card">
    <template #header>
      <div class="dashboard-panel-title">{{ messages.runtimeOverview }}</div>
    </template>
    <div class="dashboard-runtime-grid" role="list" :aria-label="messages.runtimeMetrics">
      <div v-for="item in cards" :key="item.label" class="dashboard-runtime-tile" :class="`dashboard-runtime-${item.tone}`" role="listitem">
        <div class="dashboard-runtime-label">{{ item.label }}</div>
        <div class="dashboard-runtime-value">{{ item.value }} <span>{{ item.suffix }}</span></div>
      </div>
    </div>
    <div class="dashboard-runtime-footer">
      <DoraIcon name="activity" :size="16" />
      <span>{{ siteTitle }}{{ messages.runtimeDescriptionSuffix }}</span>
    </div>
  </el-card>
</template>
