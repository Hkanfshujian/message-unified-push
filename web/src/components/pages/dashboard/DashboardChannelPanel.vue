<script setup lang="ts">
import { computed, reactive, type CSSProperties } from 'vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import type { ChannelInsight } from '@/types/business'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.dashboard

const props = defineProps<{
  error?: string
  loading: boolean
  insight: ChannelInsight
  canViewLogs?: boolean
  canViewChannels?: boolean
}>()

const emit = defineEmits<{
  (e: 'openChannel', path: string): void
  (e: 'viewAll', path: string): void
}>()

const donutSegments = computed(() => {
  const circumference = 2 * Math.PI * 47
  const activeItems = props.insight.rankedChannels.filter(item => item.count > 0)
  if (!activeItems.length) return []
  let offset = 0
  return activeItems.map((item, index) => {
    const gap = activeItems.length === 1 ? 0 : 6
    const length = activeItems.length === 1 ? circumference : Math.max((item.count / props.insight.total) * circumference - gap, 2)
    const angle = ((offset + length / 2) / circumference) * 360
    const segment = {
      ...item,
      dasharray: `${length.toFixed(2)} ${(circumference - length).toFixed(2)}`,
      dashoffset: (-offset).toFixed(2),
      angle: `${angle.toFixed(2)}deg`,
      tooltip: `${item.name}：${item.count} 条，占比 ${item.percent}%`
    }
    offset += length + (index === activeItems.length - 1 ? 0 : gap)
    return segment
  })
})

const viewAllPath = computed(() => props.insight.primaryChannel?.action?.path || '/sendways')

const tooltipState = reactive({
  visible: false,
  x: 0,
  y: 0,
  color: '',
  title: '',
  meta: ''
})

const tooltipStyle = computed(() => ({
  left: `${tooltipState.x + 14}px`,
  top: `${tooltipState.y + 14}px`,
  '--tooltip-color': tooltipState.color
}) as CSSProperties)

const showSegmentTooltip = (item: { name: string, count: number, percent: number, color: string }, event: MouseEvent | FocusEvent) => {
  const rect = (event.currentTarget as Element).getBoundingClientRect()
  const mouseEvent = event instanceof MouseEvent ? event : null
  tooltipState.visible = true
  tooltipState.x = mouseEvent?.clientX || rect.left + rect.width / 2
  tooltipState.y = mouseEvent?.clientY || rect.top + rect.height / 2
  tooltipState.color = item.color
  tooltipState.title = item.name
  tooltipState.meta = `${item.count} 条 · 占比 ${item.percent}%`
}

const moveSegmentTooltip = (event: MouseEvent) => {
  if (!tooltipState.visible) return
  tooltipState.x = event.clientX
  tooltipState.y = event.clientY
}

const hideSegmentTooltip = () => {
  tooltipState.visible = false
}
</script>

<template>
  <el-card class="dashboard-panel dashboard-channel-card" body-class="!pt-2">
    <template #header>
      <div class="dashboard-panel-header">
        <div>
          <div class="dashboard-panel-title">{{ messages.channelShare }}</div>
        </div>
      </div>
    </template>
    <div class="dashboard-channel-layout">
      <div class="dashboard-channel-visual">
        <div class="dashboard-chart-panel dashboard-channel-chart relative" role="img" :aria-label="insight.empty ? messages.noChannelStatsAria : `${messages.channelShareChartPrefix}${insight.primaryChannel?.name || messages.unknown}，${insight.primaryChannel?.percent || 0}%`">
          <svg v-if="!insight.empty" class="dashboard-channel-donut" viewBox="0 0 120 120" aria-hidden="true">
            <circle class="dashboard-channel-donut-track" cx="60" cy="60" r="47" />
            <circle
              v-for="item in donutSegments"
              :key="item.name"
              class="dashboard-channel-donut-segment"
              cx="60"
              cy="60"
              r="47"
              :stroke="item.color"
              :stroke-dasharray="item.dasharray"
              :stroke-dashoffset="item.dashoffset"
              tabindex="0"
              @mouseenter="showSegmentTooltip(item, $event)"
              @mousemove="moveSegmentTooltip"
              @mouseleave="hideSegmentTooltip"
              @focus="showSegmentTooltip(item, $event)"
              @blur="hideSegmentTooltip"
            />
          </svg>
          <div class="dashboard-channel-tooltip" :class="{ 'is-visible': tooltipState.visible }" :style="tooltipStyle" aria-hidden="true">
            <div class="dashboard-channel-tooltip-title">
              <span class="dashboard-channel-tooltip-dot" />
              {{ tooltipState.title }}
            </div>
            <div class="dashboard-channel-tooltip-meta">{{ tooltipState.meta }}</div>
          </div>
          <div v-if="!insight.empty" class="dashboard-channel-center">
            <span>{{ messages.highestShare }}</span>
            <strong>{{ insight.primaryChannel?.percent || 0 }}%</strong>
            <span>{{ insight.primaryChannel?.name || messages.unknownChannel }}</span>
          </div>
          <AppEmptyState v-if="error && !loading" class="absolute inset-0 bg-card/80" :title="messages.channelLoadFailed" :description="error" />
          <AppEmptyState v-else-if="insight.empty && !loading" class="absolute inset-0 bg-card/80" :description="messages.noChannelStats" />
        </div>
      </div>
      <div class="dashboard-channel-summary" :aria-label="messages.allChannelRanking">
        <div class="dashboard-channel-rank-head">
          <span>{{ messages.sendChannels }}</span>
          <span>{{ messages.totalPrefix }}{{ insight.total }}{{ messages.itemUnit }}</span>
        </div>
        <div class="dashboard-channel-rank" role="list">
          <button v-for="item in insight.rankedChannels" :key="item.name" type="button" class="dashboard-channel-row" role="listitem" :disabled="!canViewLogs || !item.action" :aria-label="`${item.name}${messages.sentCountPrefix}${item.count}${messages.itemUnit}${messages.sharePrefix}${item.percent}%`" @click="item.action && emit('openChannel', item.action.path)">
            <span class="dashboard-channel-icon" :style="{ backgroundColor: item.color }" aria-hidden="true">{{ item.name.slice(0, 1) }}</span>
            <span class="dashboard-channel-name">{{ item.name }}</span>
            <span class="dashboard-channel-value">{{ item.count }} <em>({{ item.percent }}%)</em></span>
            <span class="dashboard-channel-progress" aria-hidden="true"><span :style="{ width: `${item.percent}%`, backgroundColor: item.color }" /></span>
          </button>
        </div>
      </div>
      <div class="dashboard-channel-footer">
        <button v-if="canViewChannels" type="button" class="dashboard-channel-more" @click="emit('viewAll', viewAllPath)">{{ messages.viewAllChannels }}</button>
        <span v-else class="dashboard-channel-empty">{{ messages.noChannelPermission }}</span>
      </div>
    </div>
  </el-card>
</template>
