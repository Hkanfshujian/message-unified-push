<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import StatCard from '@/components/pages/dashboard/CardNum.vue'
import DashboardChannelPanel from '@/components/pages/dashboard/DashboardChannelPanel.vue'
import DashboardTrendPanel from '@/components/pages/dashboard/DashboardTrendPanel.vue'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import { useECharts } from '@/composables/useECharts'
import { consumeLogsApi, sendLogsApi } from '@/api/logs'
import { statisticsApi } from '@/api/statistics'
import { useRbacStore } from '@/store/rbac'
import { pickDateRangeQuery } from '@/util/routeQuery'
import { notifyError } from '@/util/uiFeedback'
import { analyzeTrendSeries, buildTrendChartOption, createChartThemeTokens, hasTrendChartData, normalizeTrendSeries, toIsoDate, type ChannelStatPoint, type SendTrendPoint } from '@/util/charts'
import { createLastUpdatedLabel, deriveChannelInsight, deriveCoreMetrics, deriveDashboardHealthSummary, normalizeDashboardFilterContext } from '@/util/dashboardControlCenter'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.dashboard

interface DashboardBasicData {
  message_total_num: number
  today_total_num: number
  today_succ_num: number
  today_failed_num: number
}

interface DashboardSendLogItem {
  id: number
  name: string
  log: string
  status: number
  created_on: string
}

interface DashboardConsumeLogItem {
  id: number | string
  subscription_name: string
  raw_message: string
  matched: boolean | string | number
  send_status: string | number
  send_error?: string
  created_on: string
}

type DashboardLogTab = 'send' | 'consume'

const route = useRoute()
const router = useRouter()
const rbacStore = useRbacStore()
const trendChartRef = ref<HTMLElement | null>(null)
const trendChart = useECharts(trendChartRef)
const selectedDateRange = ref<[string, string]>(['', ''])
const logDetailVisible = ref(false)
const selectedLogType = ref<DashboardLogTab>('send')
const selectedSendLog = ref<DashboardSendLogItem | null>(null)
const selectedConsumeLog = ref<DashboardConsumeLogItem | null>(null)
const datePickerPopperOptions = {
  modifiers: [
    {
      name: 'preventOverflow',
      options: {
        boundary: 'viewport',
        padding: 24,
        altAxis: true
      }
    },
    {
      name: 'flip',
      options: {
        padding: 24,
        fallbackPlacements: ['bottom-end', 'bottom', 'bottom-start', 'top-end']
      }
    }
  ]
}

const state = reactive({
  trendDays: 15,
  basicData: {
    message_total_num: 0,
    today_total_num: 0,
    today_succ_num: 0,
    today_failed_num: 0
  } as DashboardBasicData,
  trendData: [] as SendTrendPoint[],
  channelData: [] as ChannelStatPoint[],
  sendLogs: [] as DashboardSendLogItem[],
  consumeLogs: [] as DashboardConsumeLogItem[],
  sendLogPage: 1,
  consumeLogPage: 1,
  activeLogTab: 'send' as DashboardLogTab,
  lastUpdatedAt: '',
  loading: {
    basic: false,
    trend: false,
    channels: false,
    logs: false
  },
  error: {
    trend: '',
    channels: '',
    logs: ''
  },
  logsStale: false
})

let suppressNextRouteReload = false
let themeClassObserver: MutationObserver | null = null

const normalizedTrendData = computed(() => normalizeTrendSeries(state.trendData, getTrendDateRange()))
const isTrendEmpty = computed(() => !hasTrendChartData(normalizedTrendData.value))
const trendDateRange = computed(() => getTrendDateRange())
const dashboardContext = computed(() => normalizeDashboardFilterContext(trendDateRange.value.start, trendDateRange.value.end, state.trendDays))
const trendInsight = computed(() => analyzeTrendSeries(normalizedTrendData.value))
const trendInsightLabel = computed(() => trendInsight.value.empty ? '暂无发送记录' : `成功率 ${trendInsight.value.successRate}% · 失败 ${trendInsight.value.failed} 条`)
const channelInsight = computed(() => deriveChannelInsight(state.channelData, dashboardContext.value))
const healthSummary = computed(() => deriveDashboardHealthSummary(normalizedTrendData.value, dashboardContext.value, state.lastUpdatedAt))
const metricCards = computed(() => deriveCoreMetrics(state.basicData, normalizedTrendData.value, state.channelData, dashboardContext.value))
const canViewTaskLogs = computed(() => rbacStore.hasAnyPermission(['message:sendlogs:view']))
const canViewChannels = computed(() => rbacStore.hasAnyPermission(['message:sendways:view']))
const isRefreshing = computed(() => state.loading.basic || state.loading.trend || state.loading.channels || state.loading.logs)
const activeLogPage = computed(() => state.activeLogTab === 'send' ? state.sendLogPage : state.consumeLogPage)
const activeLogTotal = computed(() => state.activeLogTab === 'send' ? state.sendLogs.length : state.consumeLogs.length)
const activeLogPageCount = computed(() => Math.max(Math.ceil(activeLogTotal.value / 5), 1))
const visibleLogPageItems = computed(() => {
  const total = activeLogPageCount.value
  const current = activeLogPage.value
  if (total <= 5) return Array.from({ length: total }, (_, index) => index + 1)

  if (current <= 2) return [1, 2, 3, 'ellipsis-start', total - 1, total]
  if (current === 3) return [1, 2, 3, 4, 'ellipsis-middle', total]

  const start = Math.min(current, total - 2)
  return [1, `ellipsis-1-${start}`, start, start + 1, start + 2]
})
const displayedSendLogs = computed(() => state.sendLogs.slice((state.sendLogPage - 1) * 5, state.sendLogPage * 5))
const displayedConsumeLogs = computed(() => state.consumeLogs.slice((state.consumeLogPage - 1) * 5, state.consumeLogPage * 5))
const activeLogPaginationText = computed(() => {
  if (!activeLogTotal.value) return '共 0 条'
  return `共 ${activeLogTotal.value} 条`
})
const systemMaintenanceSendLogNames = new Set(['任务日志定时清除', '消费日志定时清除', '登录日志定时清除'])
const toolbarStatusText = computed(() => {
  if (healthSummary.value.status === 'healthy') return '运行正常'
  if (healthSummary.value.status === 'warning') return '需要关注'
  if (healthSummary.value.status === 'critical') return '异常告警'
  return '暂无数据'
})

const getTodayDate = () => toIsoDate(new Date())

const getTrendDateRange = () => {
  const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
  const end = endTime ? endTime.split('T')[0] : getTodayDate()
  const endDate = new Date(end)
  if (Number.isNaN(endDate.getTime())) {
    const now = new Date()
    const startDate = new Date(now)
    startDate.setDate(startDate.getDate() - Math.max(state.trendDays - 1, 0))
    return { start: toIsoDate(startDate), end: toIsoDate(now) }
  }
  const start = startTime ? startTime.split('T')[0] : ''
  if (start) return { start, end }
  const startDate = new Date(endDate)
  startDate.setDate(startDate.getDate() - Math.max(state.trendDays - 1, 0))
  return { start: toIsoDate(startDate), end }
}

const syncSelectedDateRange = () => {
  const range = getTrendDateRange()
  selectedDateRange.value = [range.start, range.end]
}

const getTrendDays = () => {
  const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
  const start = startTime ? startTime.split('T')[0] : ''
  const end = endTime ? endTime.split('T')[0] : ''
  if (start && end) {
    const startMs = new Date(start).getTime()
    const endMs = new Date(end).getTime()
    if (!Number.isNaN(startMs) && !Number.isNaN(endMs) && endMs >= startMs) {
      const diffDays = Math.floor((endMs - startMs) / (24 * 60 * 60 * 1000)) + 1
      return Math.min(Math.max(diffDays, 1), 90)
    }
  }
  return 15
}

const shouldSilenceAuthError = (payload: unknown) => {
  const error = payload as { status?: number; response?: { status?: number; data?: { code?: number } }; data?: { code?: number } }
  if (error?.status === 401 || error?.response?.status === 401) return true
  const code = error?.data?.code || error?.response?.data?.code
  return [20001, 20002, 20003, 20004, 20005].includes(Number(code))
}

const getChartThemeTokens = () => {
  const root = document.documentElement
  const styles = getComputedStyle(root)
  const tokens = createChartThemeTokens(root.classList.contains('dark') ? 'dark' : 'light')
  return {
    ...tokens,
    tooltipBackground: styles.getPropertyValue('--app-overlay-surface').trim() || tokens.tooltipBackground,
    tooltipBorderColor: styles.getPropertyValue('--app-overlay-border').trim() || tokens.tooltipBorderColor
  }
}

const renderCharts = async () => {
  await nextTick()
  const tokens = getChartThemeTokens()
  await trendChart.setOption(buildTrendChartOption(normalizedTrendData.value, tokens), isTrendEmpty.value)
}

const markUpdated = () => {
  state.lastUpdatedAt = createLastUpdatedLabel()
}

const getBasicStatisticData = async () => {
  state.loading.basic = true
  try {
    const rsp = await statisticsApi.basic()
    if (rsp?.data?.code === 200) {
      state.basicData = rsp.data.data || state.basicData
      markUpdated()
    } else if (!shouldSilenceAuthError(rsp)) {
      notifyError(rsp?.data?.msg || '获取基础统计数据失败')
    }
  } catch (error) {
    if (!shouldSilenceAuthError(error)) notifyError('获取基础统计数据时发生错误')
  } finally {
    state.loading.basic = false
  }
}

const getLatestLogsData = async () => {
  state.loading.logs = true
  state.error.logs = ''
  try {
    const [sendRsp, consumeRsp] = await Promise.all([
      sendLogsApi.list({ page: 1, size: 50 }),
      consumeLogsApi.list({ page: 1, size: 5 })
    ])
    if (sendRsp?.data?.code !== 200 || consumeRsp?.data?.code !== 200) throw new Error('日志接口返回失败')
    const businessSendLogs = (sendRsp.data.data?.lists || []).filter((log: DashboardSendLogItem) => !systemMaintenanceSendLogNames.has(log.name || ''))
    state.sendLogs = businessSendLogs
    state.sendLogPage = 1
    state.consumeLogs = consumeRsp.data.data?.list || []
    state.consumeLogPage = 1
    state.logsStale = false
  } catch (error) {
    state.error.logs = '最新日志加载失败，请检查网络后重试'
    state.logsStale = state.sendLogs.length > 0 || state.consumeLogs.length > 0
    if (!shouldSilenceAuthError(error)) notifyError(state.error.logs)
  } finally {
    state.loading.logs = false
  }
}

const getTrendStatisticData = async () => {
  state.loading.trend = true
  state.error.trend = ''
  trendChart.setLoading(true)
  try {
    const days = getTrendDays()
    state.trendDays = days
    const params = new URLSearchParams({
      type: 'trend',
      days: String(days)
    })
    if (dashboardContext.value.routeQuery.start_time) params.set('start_time', dashboardContext.value.routeQuery.start_time)
    if (dashboardContext.value.routeQuery.end_time) params.set('end_time', dashboardContext.value.routeQuery.end_time)
    const rsp = await statisticsApi.query(Object.fromEntries(params.entries()))
    if (rsp?.data?.code === 200) {
      state.trendData = rsp.data.data?.latest_send_data || []
      await renderCharts()
      markUpdated()
    } else if (!shouldSilenceAuthError(rsp)) {
      state.error.trend = rsp?.data?.msg || '获取趋势统计数据失败'
      trendChart.setError(state.error.trend)
      notifyError(state.error.trend)
    }
  } catch (error) {
    if (!shouldSilenceAuthError(error)) {
      state.error.trend = '获取趋势统计数据时发生错误'
      trendChart.setError(state.error.trend)
      notifyError(state.error.trend)
    }
  } finally {
    state.loading.trend = false
    trendChart.setLoading(false)
  }
}

const getChannelStatisticData = async () => {
  state.loading.channels = true
  state.error.channels = ''
  try {
    const rsp = await statisticsApi.channels()
    if (rsp?.data?.code === 200) {
      state.channelData = rsp.data.data?.way_cate_data || []
      markUpdated()
    } else if (!shouldSilenceAuthError(rsp)) {
      state.error.channels = rsp?.data?.msg || '获取渠道统计数据失败'
      notifyError(state.error.channels)
    }
  } catch (error) {
    if (!shouldSilenceAuthError(error)) {
      state.error.channels = '获取渠道统计数据时发生错误'
      notifyError(state.error.channels)
    }
  } finally {
    state.loading.channels = false
  }
}

const loadAllStatisticData = async () => {
  await Promise.all([getBasicStatisticData(), getTrendStatisticData(), getChannelStatisticData(), getLatestLogsData()])
}

const navigateByPath = (path?: string, permission?: string) => {
  if (!path) return
  if (permission && !rbacStore.hasAnyPermission([permission])) return
  router.push(path)
}

const switchLogTab = (tab: DashboardLogTab) => {
  state.activeLogTab = tab
}

const changeLogPage = (page: number) => {
  const targetPage = Math.min(Math.max(page, 1), activeLogPageCount.value)
  if (state.activeLogTab === 'send') {
    state.sendLogPage = targetPage
    return
  }
  state.consumeLogPage = targetPage
}

const getSendLogChannel = (log: DashboardSendLogItem) => {
  const match = String(log.log || '').match(/实例渠道名[:：]\s*([^\n\r]+)/)
  return match?.[1]?.trim() || '-'
}

const getConsumeLogChannel = (log: DashboardConsumeLogItem) => {
  const match = String(log.send_error || log.raw_message || '').match(/实例渠道名[:：]\s*([^\n\r]+)/)
  return match?.[1]?.trim() || '-'
}

const openSendLogDetail = (log: DashboardSendLogItem) => {
  selectedLogType.value = 'send'
  selectedSendLog.value = log
  selectedConsumeLog.value = null
  logDetailVisible.value = true
}

const openConsumeLogDetail = (log: DashboardConsumeLogItem) => {
  selectedLogType.value = 'consume'
  selectedConsumeLog.value = log
  selectedSendLog.value = null
  logDetailVisible.value = true
}

const formatLogTime = (value: string) => value ? value.replace('T', ' ').slice(0, 16) : '-'

const getSendLogStatusText = (status: number) => Number(status) === 1 ? '成功' : '失败'
const getSendLogStatusTone = (status: number) => Number(status) === 1 ? 'success' : 'danger'
const isConsumeLogMatched = (matched: DashboardConsumeLogItem['matched']) => (
  matched === true || matched === 'true' || matched === 1 || matched === '1'
)
const getConsumeLogStatusText = (log: DashboardConsumeLogItem) => {
  if (log.send_status === 'success' || log.send_status === '1') return '已发送'
  if (log.send_status === 'failed' || log.send_status === '0') return '发送失败'
  if (isConsumeLogMatched(log.matched)) return '已匹配'
  return '未匹配'
}
const getConsumeLogStatusTone = (log: DashboardConsumeLogItem) => {
  const text = getConsumeLogStatusText(log)
  if (text === '发送失败') return 'danger'
  if (text === '未匹配') return 'warning'
  return 'success'
}

const getLogStatusStyle = (tone: string) => ({
  '--dashboard-log-tone': tone === 'danger' ? '#ff5f6d' : tone === 'warning' ? '#ffb020' : '#20c997'
})

const onDateRangeChange = async (value: string[] | null) => {
  if (!Array.isArray(value) || value.length !== 2 || !value[0] || !value[1]) return
  const startMs = new Date(value[0]).getTime()
  const endMs = new Date(value[1]).getTime()
  if (Number.isNaN(startMs) || Number.isNaN(endMs) || endMs < startMs) return
  const diffDays = Math.floor((endMs - startMs) / (24 * 60 * 60 * 1000)) + 1
  suppressNextRouteReload = true
  state.trendDays = Math.min(Math.max(diffDays, 1), 90)
  await router.replace({
    path: route.path,
    query: {
      ...route.query,
      start_date: value[0],
      end_date: value[1],
      start_time: undefined,
      end_time: undefined
    }
  })
  await loadAllStatisticData()
}

onMounted(() => {
  syncSelectedDateRange()
  loadAllStatisticData()
  if (typeof MutationObserver !== 'undefined') {
    themeClassObserver = new MutationObserver(mutations => {
      if (mutations.some(mutation => mutation.type === 'attributes' && mutation.attributeName === 'class')) {
        renderCharts()
      }
    })
    themeClassObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
  }
})

watch(
  () => [route.query.start_time, route.query.end_time, route.query.start_date, route.query.end_date],
  () => {
    if (suppressNextRouteReload) {
      suppressNextRouteReload = false
      return
    }
    syncSelectedDateRange()
    loadAllStatisticData()
  }
)

onUnmounted(() => {
  themeClassObserver?.disconnect()
  themeClassObserver = null
})
</script>

<template>
  <div class="dashboard-page">
    <div class="dashboard-toolbar" :aria-label="messages.filterActions">
      <div class="dashboard-toolbar-summary">
        <span
          class="dashboard-toolbar-summary-icon"
          :class="`dashboard-toolbar-summary-icon-${healthSummary.status}`"
          aria-hidden="true"
        >
          <DoraIcon name="activity" :size="24" />
        </span>
        <span class="dashboard-hero-status" :class="`dashboard-hero-status-${healthSummary.status}`">
          <span class="dashboard-hero-status-dot" aria-hidden="true" />
          {{ toolbarStatusText }}
        </span>
      </div>
      <el-date-picker
        v-model="selectedDateRange"
        type="daterange"
        value-format="YYYY-MM-DD"
        range-separator="→"
        :start-placeholder="messages.startDate"
        :end-placeholder="messages.endDate"
        class="dashboard-date-picker"
        popper-class="dashboard-date-popper"
        placement="bottom-end"
        :popper-options="datePickerPopperOptions"
        :aria-label="messages.dateRange"
        @change="onDateRangeChange"
      />
      <el-button size="small" :loading="isRefreshing" :aria-label="messages.refreshDashboard" @click="loadAllStatisticData">{{ messages.refresh }}</el-button>
    </div>

    <div class="dashboard-stat-grid">
      <StatCard
        v-for="item in metricCards"
        :key="item.id"
        :title="item.label"
        :value="item.value"
        :unit="item.unit"
        :description="item.description"
        :trend-text="item.trendText"
        :trend-type="item.trendType"
        :tone="item.tone"
        :icon-name="item.iconName"
        :route-path="item.action && (!item.action.permission || rbacStore.hasAnyPermission([item.action.permission])) ? item.action.path : undefined"
      />
    </div>

    <div class="dashboard-chart-grid">
      <DashboardTrendPanel
        :loading="state.loading.trend"
        :error="state.error.trend"
        :is-empty="isTrendEmpty"
        :range-label="dashboardContext.rangeLabel"
        :insight-label="trendInsightLabel"
      >
        <div ref="trendChartRef" class="h-full w-full" />
      </DashboardTrendPanel>

      <DashboardChannelPanel :loading="state.loading.channels" :error="state.error.channels" :insight="channelInsight" :can-view-logs="canViewTaskLogs" :can-view-channels="canViewChannels" @open-channel="navigateByPath" @view-all="navigateByPath" />
    </div>

    <div class="dashboard-log-grid" :aria-label="messages.latestLogsRegion">
      <section class="dashboard-log-panel" v-loading="state.loading.logs">
        <div class="dashboard-log-head">
          <div>
            <div class="dashboard-log-title">{{ messages.latestLogs }}</div>
          </div>
          <div class="dashboard-log-tools">
            <div class="dashboard-log-tabs" role="tablist" :aria-label="messages.logTypeSwitch">
              <button type="button" class="dashboard-log-tab" :class="{ 'is-active': state.activeLogTab === 'send' }" role="tab" :aria-selected="state.activeLogTab === 'send'" @click="switchLogTab('send')">{{ messages.sendLogs }}</button>
              <button type="button" class="dashboard-log-tab" :class="{ 'is-active': state.activeLogTab === 'consume' }" role="tab" :aria-selected="state.activeLogTab === 'consume'" @click="switchLogTab('consume')">{{ messages.consumeLogs }}</button>
            </div>
            <button type="button" class="dashboard-log-mini-control" @click="navigateByPath(state.activeLogTab === 'send' ? '/logs/task' : '/logs/consume')">{{ messages.viewAll }}</button>
          </div>
        </div>
        <div class="dashboard-log-tab-content">
          <div v-if="state.error.logs" class="px-4 py-3 text-sm" :class="state.logsStale ? 'text-[var(--el-color-warning)]' : 'text-[var(--el-color-danger)]'" role="alert">
            {{ state.error.logs }}<template v-if="state.logsStale">{{ messages.staleLogsSuffix }}</template>
            <el-button type="primary" link @click="getLatestLogsData">{{ messages.retry }}</el-button>
          </div>
          <div v-if="state.error.logs && !state.logsStale && !state.loading.logs" class="dashboard-log-tab-panel">
            <AppEmptyState :title="messages.latestLogsLoadFailed" :description="state.error.logs" :image-size="72">
              <template #extra><el-button type="primary" @click="getLatestLogsData">{{ messages.reload }}</el-button></template>
            </AppEmptyState>
          </div>
          <div v-show="(!state.error.logs || state.logsStale) && state.activeLogTab === 'send'" class="dashboard-log-tab-panel">
            <div class="dashboard-log-table">
              <div v-for="(log, index) in displayedSendLogs" :key="`send-${log.id}`" class="dashboard-log-row">
                <span>{{ (state.sendLogPage - 1) * 5 + index + 1 }}</span>
                <span class="dashboard-log-name">{{ log.name || '-' }}</span>
                <span class="dashboard-log-channel">{{ getSendLogChannel(log) }}</span>
                <span class="dashboard-log-status" :style="getLogStatusStyle(getSendLogStatusTone(log.status))">{{ getSendLogStatusText(log.status) }}</span>
                <span>{{ formatLogTime(log.created_on) }}</span>
                <button type="button" class="dashboard-log-detail-btn" @click="openSendLogDetail(log)">{{ messages.view }}</button>
              </div>
              <AppEmptyState v-if="!state.sendLogs.length" class="dashboard-log-empty" :description="messages.noSendLogs" :image-size="72" />
            </div>
          </div>
          <div v-show="(!state.error.logs || state.logsStale) && state.activeLogTab === 'consume'" class="dashboard-log-tab-panel">
            <div class="dashboard-log-table">
              <div v-for="(log, index) in displayedConsumeLogs" :key="`consume-${log.id}`" class="dashboard-log-row">
                <span>{{ (state.consumeLogPage - 1) * 5 + index + 1 }}</span>
                <span class="dashboard-log-name">{{ log.subscription_name || '-' }}</span>
                <span class="dashboard-log-channel">{{ getConsumeLogChannel(log) }}</span>
                <span class="dashboard-log-status" :style="getLogStatusStyle(getConsumeLogStatusTone(log))">{{ getConsumeLogStatusText(log) }}</span>
                <span>{{ formatLogTime(log.created_on) }}</span>
                <button type="button" class="dashboard-log-detail-btn" @click="openConsumeLogDetail(log)">{{ messages.view }}</button>
              </div>
              <AppEmptyState v-if="!state.consumeLogs.length" class="dashboard-log-empty" :description="messages.noConsumeLogs" :image-size="72" />
            </div>
          </div>
          <div class="dashboard-log-pagination">
            <span class="dashboard-log-page-actions" :aria-label="messages.pagination">
              <span class="dashboard-log-total">{{ activeLogPaginationText }}</span>
              <button type="button" class="dashboard-log-page-btn" :disabled="activeLogPage <= 1" @click="changeLogPage(activeLogPage - 1)">‹</button>
              <template v-for="item in visibleLogPageItems" :key="item">
                <span v-if="typeof item === 'string'" class="dashboard-log-page-btn is-static">…</span>
                <button v-else type="button" class="dashboard-log-page-btn" :class="{ 'is-active': activeLogPage === item }" @click="changeLogPage(item)">{{ item }}</button>
              </template>
              <button type="button" class="dashboard-log-page-btn" :aria-label="messages.nextPage" :disabled="activeLogPage >= activeLogPageCount" @click="changeLogPage(activeLogPage + 1)">›</button>
              <button type="button" class="dashboard-log-page-btn dashboard-log-page-btn-last" :aria-label="messages.lastPage" :title="messages.lastPageTitle" :disabled="activeLogPage >= activeLogPageCount" @click="changeLogPage(activeLogPageCount)">»</button>
            </span>
          </div>
        </div>
      </section>
    </div>

    <el-dialog v-model="logDetailVisible" :title="selectedLogType === 'send' ? messages.sendLogDetails : messages.consumeLogDetails" width="min(720px, calc(100vw - 32px))" class="dashboard-log-detail-dialog app-nested-dialog" append-to-body>
      <div v-if="selectedLogType === 'send' && selectedSendLog" class="dashboard-log-detail">
        <section class="dashboard-log-detail-summary">
          <div><span>{{ messages.sendTask }}</span><h3>{{ selectedSendLog.name || messages.unnamedTask }}</h3><code>#{{ selectedSendLog.id }}</code></div>
          <span class="dashboard-log-detail-status" :style="getLogStatusStyle(getSendLogStatusTone(selectedSendLog.status))">{{ getSendLogStatusText(selectedSendLog.status) }}</span>
        </section>
        <section class="dashboard-log-detail-section">
          <header><h4>{{ messages.basicInfo }}</h4></header>
          <dl class="dashboard-log-detail-grid">
            <div><dt>{{ messages.taskName }}</dt><dd>{{ selectedSendLog.name || '-' }}</dd></div>
            <div><dt>{{ messages.sendChannel }}</dt><dd>{{ getSendLogChannel(selectedSendLog) }}</dd></div>
            <div><dt>{{ messages.logId }}</dt><dd>{{ selectedSendLog.id }}</dd></div>
            <div><dt>{{ messages.sendTime }}</dt><dd>{{ formatLogTime(selectedSendLog.created_on) }}</dd></div>
          </dl>
        </section>
        <section class="dashboard-log-detail-section">
          <header><h4>{{ messages.logBody }}</h4><span>{{ messages.fullRecord }}</span></header>
          <pre class="dashboard-log-detail-content">{{ selectedSendLog.log || messages.noSendLogContent }}</pre>
        </section>
      </div>

      <div v-else-if="selectedConsumeLog" class="dashboard-log-detail">
        <section class="dashboard-log-detail-summary">
          <div><span>{{ messages.consumeSubscription }}</span><h3>{{ selectedConsumeLog.subscription_name || messages.unnamedSubscription }}</h3><code>#{{ selectedConsumeLog.id }}</code></div>
          <span class="dashboard-log-detail-status" :style="getLogStatusStyle(getConsumeLogStatusTone(selectedConsumeLog))">{{ getConsumeLogStatusText(selectedConsumeLog) }}</span>
        </section>
        <section class="dashboard-log-detail-section">
          <header><h4>{{ messages.basicInfo }}</h4></header>
          <dl class="dashboard-log-detail-grid">
            <div><dt>{{ messages.subscriptionName }}</dt><dd>{{ selectedConsumeLog.subscription_name || '-' }}</dd></div>
            <div><dt>{{ messages.sendChannel }}</dt><dd>{{ getConsumeLogChannel(selectedConsumeLog) }}</dd></div>
            <div><dt>{{ messages.matchResult }}</dt><dd>{{ isConsumeLogMatched(selectedConsumeLog.matched) ? messages.matched : messages.unmatched }}</dd></div>
            <div><dt>{{ messages.consumeTime }}</dt><dd>{{ formatLogTime(selectedConsumeLog.created_on) }}</dd></div>
          </dl>
        </section>
        <section class="dashboard-log-detail-section">
          <header><h4>{{ messages.rawMessage }}</h4><span>{{ messages.consumeContent }}</span></header>
          <pre class="dashboard-log-detail-content">{{ selectedConsumeLog.raw_message || messages.noRawMessage }}</pre>
        </section>
        <section v-if="selectedConsumeLog.send_error" class="dashboard-log-detail-section is-error">
          <header><h4>{{ messages.sendError }}</h4></header>
          <pre class="dashboard-log-detail-content">{{ selectedConsumeLog.send_error }}</pre>
        </section>
      </div>
    </el-dialog>
  </div>
</template>
