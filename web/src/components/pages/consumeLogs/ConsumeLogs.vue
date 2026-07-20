<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Download, Refresh, Search } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import AppDateTimeRange from '@/components/ui/AppDateTimeRange.vue'
import AppDetailDrawer from '@/components/ui/AppDetailDrawer.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import AppTruncate from '@/components/ui/AppTruncate.vue'
import { consumeLogsApi } from '@/api/logs'
import { getPageSize } from '@/util/pageUtils'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { downloadBlob, notifyError, notifySuccess } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.consumeLogs

interface ConsumeLogItem {
  id: string
  subscription_id: string
  subscription_name: string
  raw_message: string
  matched: number
  extracted_values: string
  send_status: number
  send_error: string
  created_on: string
}

interface ConsumeStats {
  total_consume: number
  total_matched: number
  total_sent: number
  total_failed: number
}

const route = useRoute()
const router = useRouter()

const getTodayRange = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return [`${year}-${month}-${day}T00:00`, `${year}-${month}-${day}T23:59`] as [string, string]
}

const state = reactive({
  tableData: [] as ConsumeLogItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false,
  statsLoading: false
})

const selectedMatched = ref('all')
const selectedSendStatus = ref('all')
const timeRange = ref<[string, string] | []>(getTodayRange())
const isDrawerOpen = ref(false)
const detailLoading = ref(false)
const selectedLog = ref<ConsumeLogItem | null>(null)
const stats = ref<ConsumeStats>({ total_consume: 0, total_matched: 0, total_sent: 0, total_failed: 0 })

const columns: AppTableColumn[] = [
  { prop: 'index', label: '序号', width: 80, align: 'center' },
  { prop: 'id', label: 'ID', minWidth: 100 },
  { prop: 'subscription_name', label: '订阅名称', minWidth: 180 },
  { prop: 'raw_message', label: '原始消息', minWidth: 300 },
  { prop: 'matched', label: '匹配状态', width: 110, align: 'center' },
  { prop: 'send_status', label: '发送状态', width: 120, align: 'center' },
  { prop: 'created_on', label: '消费时间', minWidth: 170 },
  { prop: 'actions', label: '操作', width: 90, align: 'center', fixed: 'right' }
]

const toolbarColumns: TableToolbarColumn[] = columns.map(column => ({
  key: column.prop || column.label,
  label: column.label,
  required: column.prop === 'index' || column.prop === 'actions'
}))

const tableToolbar = reactive(createTableToolbarState(toolbarColumns))
const visibleColumns = computed(() => getVisibleToolbarColumns(columns, tableToolbar.visibleColumns))

const refreshTable = async () => {
  if (tableToolbar.refreshing) return
  tableToolbar.refreshing = true
  try {
    await queryListData()
  } finally {
    tableToolbar.refreshing = false
  }
}

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.floor(n)
}

const matchedLabelMap = { 1: '已匹配', 0: '未匹配' }
const sendStatusLabelMap = { 0: '未发送', 1: '发送成功', 2: '发送失败' }

const formatExtractedValues = (values: string) => {
  try {
    const parsed = JSON.parse(values)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return values || '-'
  }
}

const extractedEntries = computed(() => {
  if (!selectedLog.value?.extracted_values) return [] as Array<[string, unknown]>
  try {
    const parsed = JSON.parse(selectedLog.value.extracted_values)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return []
    return Object.entries(parsed)
  } catch {
    return []
  }
})

const openLogDrawer = async (item: ConsumeLogItem) => {
  selectedLog.value = item
  isDrawerOpen.value = true
  detailLoading.value = true
  try {
    const rsp = await consumeLogsApi.detail(item.id)
    if (rsp?.data?.code === 200 && rsp.data.data) selectedLog.value = { ...item, ...rsp.data.data }
  } catch (error) {
  } finally {
    detailLoading.value = false
  }
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {
    page: String(state.currPage),
    page_size: String(state.pageSize)
  }
  const name = state.search.trim()
  if (name) nextQuery.subscription_name = name
  if (selectedMatched.value !== 'all') nextQuery.matched = selectedMatched.value
  if (selectedSendStatus.value !== 'all') nextQuery.send_status = selectedSendStatus.value
  const [startTime, endTime] = timeRange.value
  if (startTime) nextQuery.start_time = startTime
  if (endTime) nextQuery.end_time = endTime
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

const buildParams = () => {
  const params: Record<string, unknown> = {
    page: state.currPage,
    page_size: state.pageSize
  }
  if (state.search.trim()) params.subscription_name = state.search.trim()
  if (selectedMatched.value !== 'all') params.matched = selectedMatched.value
  if (selectedSendStatus.value !== 'all') params.send_status = selectedSendStatus.value
  const [startTime, endTime] = timeRange.value
  if (startTime) params.start_time = startTime
  if (endTime) params.end_time = endTime
  return params
}

const loadStats = async () => {
  state.statsLoading = true
  try {
    const rsp = await consumeLogsApi.stats()
    if (rsp?.data?.code === 200) {
      stats.value = {
        total_consume: Number(rsp.data.data?.total_consume || 0),
        total_matched: Number(rsp.data.data?.total_matched || 0),
        total_sent: Number(rsp.data.data?.total_sent || 0),
        total_failed: Number(rsp.data.data?.total_failed || 0)
      }
    }
  } catch (error) {
    notifyError('获取消费统计失败')
  } finally {
    state.statsLoading = false
  }
}

const queryListData = async (shouldSyncRoute = true) => {
  if (shouldSyncRoute) await syncRouteQuery()
  state.loading = true
  try {
    const res = await consumeLogsApi.list(buildParams())
    if (res.data.code === 200) {
      state.tableData = res.data.data.list || []
      state.total = res.data.data.total || 0
      return
    }
    state.tableData = []
    state.total = 0
    notifyError(res.data.msg || '获取消费日志失败')
  } catch (error) {
    state.tableData = []
    state.total = 0
    notifyError('获取消费日志时发生错误')
  } finally {
    state.loading = false
  }
}

const filterFunc = async () => {
  state.currPage = 1
  await queryListData()
}

const filterByMatched = async (value: string) => {
  selectedMatched.value = value
  state.currPage = 1
  await queryListData()
}

const filterBySendStatus = async (value: string) => {
  selectedSendStatus.value = value
  state.currPage = 1
  await queryListData()
}

const handlePaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await queryListData()
}

const clearTimeFilter = async () => {
  timeRange.value = getTodayRange()
  state.currPage = 1
  await queryListData()
}

const handleExport = async () => {
  try {
    const rsp = await consumeLogsApi.export(buildParams())
    downloadBlob(rsp.data, `consume-logs-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('消费日志导出成功')
  } catch (error) {
    notifyError('消费日志导出失败')
  }
}

onMounted(async () => {
  state.search = route.query.subscription_name?.toString() || ''
  selectedMatched.value = route.query.matched?.toString() || 'all'
  selectedSendStatus.value = route.query.send_status?.toString() || 'all'
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.page_size, state.pageSize)
  const startTime = route.query.start_time?.toString()
  const endTime = route.query.end_time?.toString()
  if (startTime || endTime) timeRange.value = [startTime || '', endTime || '']
  await Promise.all([queryListData(), loadStats()])
})
</script>

<template>
  <div class="space-y-4">
    <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
      <el-card v-loading="state.statsLoading" shadow="never"><el-statistic :title="messages.totalConsumed" :value="stats.total_consume" /></el-card>
      <el-card v-loading="state.statsLoading" shadow="never"><el-statistic :title="messages.matched" :value="stats.total_matched" /></el-card>
      <el-card v-loading="state.statsLoading" shadow="never"><el-statistic :title="messages.sent" :value="stats.total_sent" /></el-card>
      <el-card v-loading="state.statsLoading" shadow="never"><el-statistic :title="messages.failed" :value="stats.total_failed" /></el-card>
    </div>

    <el-card shadow="never">
      <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
        <div class="flex flex-1 flex-col gap-3 lg:flex-row lg:items-center">
          <el-input v-model="state.search" class="lg:max-w-[220px]" clearable :placeholder="messages.searchPlaceholder" @keyup.enter="filterFunc" @clear="filterFunc">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="selectedMatched" class="lg:max-w-[140px]" @change="filterByMatched">
            <el-option :label="messages.allMatches" value="all" />
            <el-option :label="messages.matched" value="1" />
            <el-option :label="messages.unmatched" value="0" />
          </el-select>
          <el-select v-model="selectedSendStatus" class="lg:max-w-[150px]" @change="filterBySendStatus">
            <el-option :label="messages.allSendStatuses" value="all" />
            <el-option :label="messages.sent" value="1" />
            <el-option :label="messages.failed" value="2" />
            <el-option :label="messages.notSent" value="0" />
          </el-select>
          <AppDateTimeRange v-model="timeRange" class="lg:max-w-[360px]" @change="filterFunc" />
          <el-button @click="clearTimeFilter">{{ messages.resetTime }}</el-button>
          <el-button :icon="Search" @click="filterFunc">{{ messages.search }}</el-button>
          <el-button :icon="Refresh" @click="refreshTable">{{ messages.refresh }}</el-button>
        </div>
        <el-button :icon="Download" @click="handleExport">{{ messages.export }}</el-button>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0" :class="tableToolbar.focused ? 'app-table-focused-card' : ''">
      <AppTableToolbar
        :title="messages.title"
        :columns="toolbarColumns"
        v-model:visible-columns="tableToolbar.visibleColumns"
        v-model:focused="tableToolbar.focused"
        :refreshing="tableToolbar.refreshing || state.loading"
        @refresh="refreshTable"
      >
        <template #summary>
          <span class="text-xs text-muted-foreground">{{ messages.totalPrefix }}{{ state.total }}{{ messages.itemUnit }}</span>
        </template>
      </AppTableToolbar>
      <AppTable :data="state.tableData as unknown as Record<string, unknown>[]" :columns="visibleColumns" :loading="state.loading" :empty-text="messages.empty">
        <template #index="{ index }">{{ (state.currPage - 1) * state.pageSize + index + 1 }}</template>
        <template #id="{ row }"><AppTruncate :text="String(row.id || '-')" class="font-mono text-sm" /></template>
        <template #subscription_name="{ row }">{{ row.subscription_name || '-' }}</template>
        <template #raw_message="{ row }"><AppTruncate :text="String(row.raw_message || '-')" :title="messages.rawMessage" width="760px" /></template>
        <template #matched="{ row }">
          <AppStatusTag :status="row.matched" :label-map="matchedLabelMap" :success-values="[1]" :warning-values="[0]" />
        </template>
        <template #send_status="{ row }">
          <AppStatusTag :status="row.send_status" :label-map="sendStatusLabelMap" :success-values="[1]" :danger-values="[2]" :warning-values="[0]" />
        </template>
        <template #created_on="{ row }"><span class="text-sm text-muted-foreground">{{ row.created_on || '-' }}</span></template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'view', label: messages.viewDetails, kind: 'view', onClick: () => openLogDrawer(row as unknown as ConsumeLogItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState :description="messages.empty" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />

    <AppDetailDrawer v-model="isDrawerOpen" :title="messages.detailTitle" size="680px">
      <div v-if="selectedLog" v-loading="detailLoading" :element-loading-text="messages.detailLoading" class="consume-log-details">
        <section class="consume-log-identity">
          <div class="consume-log-mark">{{ messages.mark }}</div>
          <div><span>{{ messages.title }}</span><h3>{{ selectedLog.subscription_name || messages.unnamedSubscription }}</h3><code>{{ selectedLog.id }}</code><p>{{ messages.detailDescription }}</p></div>
          <AppStatusTag :status="selectedLog.send_status" :label-map="sendStatusLabelMap" :success-values="[1]" :danger-values="[2]" :warning-values="[0]" />
        </section>
        <section class="consume-log-detail-section">
          <h3 class="consume-log-detail-title">{{ messages.summary }}</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="messages.logId">{{ selectedLog.id }}</el-descriptions-item>
            <el-descriptions-item :label="messages.subscriptionName">{{ selectedLog.subscription_name || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="messages.matchStatus">
              <AppStatusTag :status="selectedLog.matched" :label-map="matchedLabelMap" :success-values="[1]" :warning-values="[0]" />
            </el-descriptions-item>
            <el-descriptions-item :label="messages.sendStatus">
              <AppStatusTag :status="selectedLog.send_status" :label-map="sendStatusLabelMap" :success-values="[1]" :danger-values="[2]" :warning-values="[0]" />
            </el-descriptions-item>
            <el-descriptions-item :label="messages.consumedAt">{{ selectedLog.created_on }}</el-descriptions-item>
          </el-descriptions>
        </section>
        <section class="consume-log-detail-section">
          <h3 class="consume-log-detail-title">{{ messages.rawMessage }}</h3>
          <pre class="consume-log-detail-content">{{ selectedLog.raw_message }}</pre>
        </section>
        <section v-if="selectedLog.extracted_values" class="consume-log-detail-section">
          <h3 class="consume-log-detail-title">{{ messages.extractedFields }}</h3>
          <div v-if="extractedEntries.length" class="consume-log-fields">
            <div v-for="([key, value], idx) in extractedEntries" :key="`${key}-${idx}`" class="consume-log-field">
              <el-tag effect="plain">{{ key }}</el-tag>
              <span class="break-all font-mono">{{ value }}</span>
            </div>
          </div>
          <pre v-else class="consume-log-detail-content">{{ formatExtractedValues(selectedLog.extracted_values) }}</pre>
        </section>
        <section v-if="selectedLog.send_status === 2 && selectedLog.send_error" class="consume-log-detail-section">
          <h3 class="consume-log-detail-title">{{ messages.sendError }}</h3>
          <pre class="consume-log-detail-content consume-log-error-detail">{{ selectedLog.send_error }}</pre>
        </section>
      </div>
    </AppDetailDrawer>
  </div>
</template>

<style scoped>
.consume-log-details {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.consume-log-detail-section {
  border: 1px solid var(--glass-inset-border);
  border-radius: var(--admin-radius-lg);
  background: var(--glass-inset-bg);
  padding: 14px;
}

.consume-log-detail-title {
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: 600;
}

.consume-log-detail-content {
  margin: 0;
  overflow-x: auto;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
}

.consume-log-fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.consume-log-field {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  border-top: 1px solid var(--glass-inset-border);
  padding-top: 8px;
  font-size: 13px;
}

.consume-log-field:first-child {
  border-top: 0;
  padding-top: 0;
}

@media (max-width: 760px) { .consume-log-identity { grid-template-columns: 38px minmax(0, 1fr); } .consume-log-identity > :deep(.el-tag) { grid-column: 1 / -1; width: fit-content; } }
</style>
