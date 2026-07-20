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
import { sendLogsApi } from '@/api/logs'
import { getPageSize } from '@/util/pageUtils'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { downloadBlob, notifyError, notifySuccess } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.sendLogs

interface LogItem {
  id: number
  task_id: string
  type: string
  name: string
  log: string
  created_on: string
  caller_ip?: string
  status: number
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
  tableData: [] as LogItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
  loading: false
})

const timeRange = ref<[string, string] | []>(getTodayRange())
const selectedStatus = ref('all')
const isDrawerOpen = ref(false)
const selectedLog = ref<LogItem | null>(null)
const hasTemplateFilter = computed(() => String(state.optionValue || '').trim() !== '')

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 90, align: 'center' },
  { prop: 'type', label: '类型', width: 120, align: 'center' },
  { prop: 'name', label: '名称', minWidth: 180 },
  { prop: 'log', label: '发信日志', minWidth: 320 },
  { prop: 'created_on', label: '发送时间', minWidth: 170 },
  { prop: 'status', label: '状态', width: 100, align: 'center' },
  { prop: 'actions', label: '操作', width: 120, align: 'center', fixed: 'right' }
]

const toolbarColumns: TableToolbarColumn[] = columns.map(column => ({
  key: column.prop || column.label,
  label: column.label,
  required: column.prop === 'id' || column.prop === 'actions'
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

const getTypeText = (type: string) => {
  if (type === 'template') return '接口调用'
  if (type === 'cron_message') return '定时消息'
  return '系统任务'
}

const getTypeTagType = (type: string) => type === 'template' ? 'info' : 'primary'

const openLogDrawer = (task: LogItem) => {
  selectedLog.value = task
  isDrawerOpen.value = true
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {
    page: String(state.currPage),
    size: String(state.pageSize)
  }
  const name = state.search.trim()
  if (name) nextQuery.name = name
  if (state.optionValue) nextQuery.taskid = state.optionValue
  if (selectedStatus.value !== 'all') nextQuery.status = selectedStatus.value
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
    size: state.pageSize
  }
  if (state.search.trim()) params.name = state.search.trim()
  if (state.optionValue) params.taskid = state.optionValue
  if (selectedStatus.value !== 'all') params.status = selectedStatus.value
  const [startTime, endTime] = timeRange.value
  if (startTime) params.start_time = startTime
  if (endTime) params.end_time = endTime
  return params
}

const queryListData = async (shouldSyncRoute = true) => {
  if (shouldSyncRoute) await syncRouteQuery()
  state.loading = true
  try {
    const rsp = await sendLogsApi.list(buildParams())
    if (rsp?.data?.code === 200) {
      state.tableData = rsp.data.data?.lists || []
      state.total = rsp.data.data?.total || 0
      return
    }
    state.tableData = []
    state.total = 0
    notifyError(rsp?.data?.msg || '获取发信日志失败')
  } catch (error) {
    state.tableData = []
    state.total = 0
    notifyError('获取发信日志时发生错误')
  } finally {
    state.loading = false
  }
}

const filterFunc = async () => {
  state.currPage = 1
  await queryListData()
}

const filterByStatus = async (value: string) => {
  selectedStatus.value = value
  state.currPage = 1
  await queryListData()
}

const handlePaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await queryListData()
}

const clearTemplateFilter = async () => {
  state.optionValue = ''
  state.currPage = 1
  await queryListData()
}

const clearTimeFilter = async () => {
  timeRange.value = getTodayRange()
  state.currPage = 1
  await queryListData()
}

const handleExport = async () => {
  try {
    const rsp = await sendLogsApi.export(buildParams())
    downloadBlob(rsp.data, `sendlogs-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('发信日志导出成功')
  } catch (error) {
    notifyError('发信日志导出失败')
  }
}

onMounted(async () => {
  state.search = route.query.name?.toString() || ''
  state.optionValue = route.query.taskid?.toString() || ''
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.size, state.pageSize)
  const startTime = route.query.start_time?.toString()
  const endTime = route.query.end_time?.toString()
  if (startTime || endTime) timeRange.value = [startTime || '', endTime || '']
  selectedStatus.value = route.query.status?.toString() || 'all'
  await queryListData()
})
</script>

<template>
  <div class="space-y-4">
    <el-alert v-if="hasTemplateFilter" type="info" show-icon :closable="false">
      <template #title>
        {{ messages.templateFilterPrefix }}<span class="font-semibold">{{ state.optionValue }}</span>
        <el-button size="small" text type="primary" @click="clearTemplateFilter">{{ messages.clearFilter }}</el-button>
      </template>
    </el-alert>

    <el-card shadow="never">
      <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
        <div class="flex flex-1 flex-col gap-3 lg:flex-row lg:items-center">
          <el-input v-model="state.search" class="lg:max-w-[240px]" clearable :placeholder="messages.searchPlaceholder" @keyup.enter="filterFunc" @clear="filterFunc">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="selectedStatus" class="lg:max-w-[140px]" @change="filterByStatus">
            <el-option :label="messages.allStatuses" value="all" />
            <el-option :label="messages.success" value="1" />
            <el-option :label="messages.failure" value="0" />
          </el-select>
          <AppDateTimeRange v-model="timeRange" class="lg:max-w-[360px]" @change="filterFunc" />
          <el-button @click="clearTimeFilter">{{ messages.resetTime }}</el-button>
          <el-button :icon="Search" @click="filterFunc">{{ messages.search }}</el-button>
          <el-button :icon="Refresh" @click="queryListData()">{{ messages.refresh }}</el-button>
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
        <template #type="{ row }">
          <el-tag :type="getTypeTagType(String(row.type || 'task'))" effect="light">{{ getTypeText(String(row.type || 'task')) }}</el-tag>
        </template>
        <template #name="{ row }">
          <AppTruncate :text="String(row.name || '-')" :title="messages.name" />
        </template>
        <template #log="{ row }">
          <AppTruncate :text="String(row.log || '-')" :title="messages.log" width="760px" />
        </template>
        <template #created_on="{ row }">
          <span class="text-sm text-muted-foreground">{{ row.created_on || '-' }}</span>
        </template>
        <template #status="{ row }">
          <AppStatusTag :status="row.status" :label-map="{ 1: messages.success, 0: messages.failure }" :success-values="[1]" :danger-values="[0]" />
        </template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'view', label: messages.view, kind: 'view', onClick: () => openLogDrawer(row as unknown as LogItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState :description="messages.empty" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />

    <AppDetailDrawer v-model="isDrawerOpen" :title="messages.detailTitle" size="760px">
      <div v-if="selectedLog" class="send-log-details">
        <section class="send-log-summary">
          <div class="send-log-summary-main">
            <span class="send-log-summary-type">{{ getTypeText(selectedLog.type) }}</span>
            <h3>{{ selectedLog.name || messages.unnamedTask }}</h3>
            <code>#{{ selectedLog.id }}</code>
          </div>
          <AppStatusTag :status="selectedLog.status" :label-map="{ 1: messages.success, 0: messages.failure }" :success-values="[1]" :danger-values="[0]" />
        </section>
        <section class="send-log-detail-section">
          <header><h3>{{ messages.basicInfo }}</h3></header>
          <dl class="send-log-info">
            <div><dt>{{ messages.taskId }}</dt><dd>{{ selectedLog.task_id || '-' }}</dd></div>
            <div><dt>{{ messages.logType }}</dt><dd>{{ getTypeText(selectedLog.type) }}</dd></div>
            <div><dt>{{ messages.sendTime }}</dt><dd>{{ selectedLog.created_on || '-' }}</dd></div>
            <div><dt>{{ messages.callerIp }}</dt><dd>{{ selectedLog.caller_ip || '-' }}</dd></div>
          </dl>
        </section>
        <section class="send-log-detail-section send-log-content-section">
          <header><h3>{{ messages.logBody }}</h3></header>
          <pre class="send-log-detail-content">{{ selectedLog.log || '-' }}</pre>
        </section>
      </div>
    </AppDetailDrawer>
  </div>
</template>

<style scoped>
.send-log-details {
  display: grid;
  gap: 14px;
  max-width: 100%;
}

.send-log-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 14px 16px;
  border: 1px solid color-mix(in srgb, var(--brand-500) 20%, var(--app-overlay-border));
  border-radius: 12px;
  background: color-mix(in srgb, var(--brand-50) 58%, var(--app-overlay-surface));
}

.send-log-summary-main {
  display: grid;
  min-width: 0;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: baseline;
  gap: 8px;
}

.send-log-summary-type {
  color: var(--admin-text-muted);
  font-size: 12px;
  white-space: nowrap;
}

.send-log-summary h3 {
  min-width: 0;
  overflow: hidden;
  margin: 0;
  color: var(--admin-text-primary);
  font-size: 15px;
  font-weight: 720;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.send-log-summary code {
  color: var(--admin-text-muted);
  font-size: 11px;
  white-space: nowrap;
}

.send-log-detail-section {
  overflow: hidden;
  border: 1px solid var(--app-overlay-border);
  border-radius: 12px;
  background: var(--app-overlay-surface);
}

.send-log-detail-section > header {
  display: flex;
  min-height: 38px;
  align-items: center;
  padding: 0 14px;
  border-bottom: 1px solid var(--app-overlay-border);
  background: color-mix(in srgb, var(--app-overlay-surface) 92%, var(--brand-50));
}

.send-log-detail-section > header h3 {
  margin: 0;
  color: var(--admin-text-primary);
  font-size: 13px;
  font-weight: 700;
}

.send-log-info {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.send-log-info > div {
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  min-height: 40px;
  align-items: center;
  padding: 6px 14px;
  border-bottom: 1px solid color-mix(in srgb, var(--app-overlay-border) 72%, transparent);
}

.send-log-info > div:nth-child(odd) {
  border-right: 1px solid color-mix(in srgb, var(--app-overlay-border) 72%, transparent);
}

.send-log-info > div:nth-last-child(-n + 2) {
  border-bottom: 0;
}

.send-log-info dt,
.send-log-info dd {
  margin: 0;
  line-height: 1.45;
}

.send-log-info dt {
  color: var(--admin-text-muted);
  font-size: 12px;
}

.send-log-info dd {
  min-width: 0;
  color: var(--admin-text-primary);
  font-size: 12px;
  font-weight: 550;
  overflow-wrap: anywhere;
}

.send-log-content-section {
  min-height: 180px;
}

.send-log-detail-content {
  box-sizing: border-box;
  min-height: 140px;
  max-height: min(42vh, 380px);
  margin: 0;
  overflow: auto;
  padding: 14px;
  background: color-mix(in srgb, var(--app-overlay-surface) 91%, #000 3%);
  color: var(--admin-text-primary);
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  scrollbar-color: color-mix(in srgb, var(--brand-500) 55%, var(--app-overlay-border)) transparent;
  scrollbar-width: thin;
}

@media (max-width: 760px) {
  .send-log-summary {
    align-items: flex-start;
    flex-direction: column;
    gap: 10px;
    padding: 12px;
  }

  .send-log-summary-main {
    width: 100%;
    grid-template-columns: 1fr auto;
  }

  .send-log-summary-type {
    grid-column: 1 / -1;
  }

  .send-log-info {
    grid-template-columns: 1fr;
  }

  .send-log-info > div,
  .send-log-info > div:nth-child(odd),
  .send-log-info > div:nth-last-child(-n + 2) {
    border-right: 0;
    border-bottom: 1px solid color-mix(in srgb, var(--app-overlay-border) 72%, transparent);
  }

  .send-log-info > div:last-child {
    border-bottom: 0;
  }
}
</style>
