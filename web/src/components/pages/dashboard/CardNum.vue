<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import type { DoraIconName } from '@/types/app'
import type { DashboardTone } from '@/types/business'
import { zhCN } from '@/locales/zh-CN'

const router = useRouter()
const messages = zhCN.dashboard

const props = defineProps<{
  title: string
  value: string | number
  description?: string
  badgeText?: string
  iconName?: DoraIconName
  routePath?: string
  trendText?: string
  trendType?: 'up' | 'down' | 'flat'
  tone?: DashboardTone
  scopeLabel?: string
  unit?: string
}>()

const toneMap = {
  blue: { card: 'dashboard-metric-blue' },
  green: { card: 'dashboard-metric-green' },
  red: { card: 'dashboard-metric-red' },
  amber: { card: 'dashboard-metric-amber' },
  purple: { card: 'dashboard-metric-purple' },
  slate: { card: 'dashboard-metric-slate' }
} as const

const cardToneClass = computed(() => toneMap[props.tone || 'blue'].card)
const trendClass = computed(() => {
  if (props.trendType === 'up') return 'dashboard-metric-trend-up'
  if (props.trendType === 'down') return 'dashboard-metric-trend-down'
  return 'dashboard-metric-trend-flat'
})
const isInteractive = computed(() => !!props.routePath)

const handleClick = () => {
  if (props.routePath) router.push(props.routePath)
}
</script>

<template>
  <button
    type="button"
    class="dashboard-metric-card w-full transition-all duration-[var(--motion-fast)] text-left"
    :class="[cardToneClass, isInteractive ? 'cursor-pointer' : 'cursor-default']"
    :aria-label="`${title}：${value}${unit || ''}，${description || messages.status}，${scopeLabel || ''} ${trendText || ''}`"
    :disabled="!isInteractive"
    @click="handleClick"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="dashboard-metric-content">
        <p class="dashboard-metric-title">{{ title }}</p>
        <div class="dashboard-metric-value">{{ value }}<span v-if="unit">{{ unit }}</span></div>
        <p class="dashboard-metric-desc">
          {{ description || messages.status }}
          <span v-if="trendText" :class="trendClass">{{ trendText }}</span>
        </p>
        <span v-if="scopeLabel || badgeText" class="dashboard-metric-badge">{{ scopeLabel || badgeText }}</span>
      </div>
      <div class="dashboard-metric-icon">
        <DoraIcon v-if="iconName" :name="iconName" :size="22" />
      </div>
    </div>

  </button>
</template>
