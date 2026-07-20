<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { Download, Plus, Refresh, Search } from '@element-plus/icons-vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import AppTruncate from '@/components/ui/AppTruncate.vue'
import { mqApi } from '@/api/mq'
import { settingsApi } from '@/api/settings'
import { getPageSize } from '@/util/pageUtils'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { appendDateRangeQuery, pickDateRangeQuery } from '@/util/routeQuery'
import { downloadBlob, notifyError, notifySuccess } from '@/util/uiFeedback'
import MQSourceForm from './MQSourceForm.vue'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.mqSources

interface MQSourceItem extends Record<string, unknown> {
  id: string
  name: string
  type: string
  namesrv_addr: string
  access_key: string
  secret_key: string
  enabled: number
  last_test_status: string
  last_test_time: string
  test_error: string
  created_on: string
  binding_count?: number
}

const route = useRoute()
const router = useRouter()
const STATUS_UNTESTED = '__untested__'

const state = reactive({
  tableData: [] as MQSourceItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false,
  error: '',
  stale: false
})

const selectedStatus = ref('all')
const selectedType = ref('all')
const isAddDialogOpen = ref(false)
const isEditDialogOpen = ref(false)
const isTestDialogOpen = ref(false)
const isDeleteConfirmOpen = ref(false)
const editData = ref<MQSourceItem | null>(null)
const testResult = ref<{ success: boolean; message?: string; error?: string } | null>(null)
const deleteTarget = ref<MQSourceItem | null>(null)
const deleteConfirmInput = ref('')
let autoRefreshTimer: number | null = null

const typeOptions = [
  { value: 'all', label: '全部类型' },
  { value: 'rocketmq', label: 'RocketMQ' },
  { value: 'kafka', label: 'Kafka' },
  { value: 'rabbitmq', label: 'RabbitMQ' }
]

const statusOptions = [
  { value: 'all', label: '全部状态' },
  { value: 'success', label: '在线' },
  { value: 'failed', label: '离线' },
  { value: STATUS_UNTESTED, label: '未测试' }
]

const columns: AppTableColumn[] = [
  { prop: 'index', label: '序号', width: 80, align: 'center' },
  { prop: 'id', label: 'ID', minWidth: 130 },
  { prop: 'name', label: '队列名称', minWidth: 180 },
  { prop: 'type', label: '队列类型', width: 120, align: 'center' },
  { prop: 'namesrv_addr', label: '队列地址', minWidth: 260 },
  { prop: 'binding_count', label: '外部绑定', width: 120, align: 'center' },
  { prop: 'last_test_status', label: '状态', width: 110, align: 'center' },
  { prop: 'last_test_time', label: '最后测试时间', minWidth: 170 },
  { prop: 'actions', label: '操作', width: 190, align: 'center', fixed: 'right' }
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
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : fallback
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {
    page: String(state.currPage),
    page_size: String(state.pageSize)
  }
  const name = state.search.trim()
  if (name) nextQuery.name = name
  if (selectedType.value !== 'all') nextQuery.type = selectedType.value
  if (selectedStatus.value !== 'all') nextQuery.status = selectedStatus.value
  appendDateRangeQuery(nextQuery, route.query as Record<string, unknown>)
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

const getTypeText = (type: string) => typeOptions.find((item) => item.value === type)?.label || type

const normalizedStatus = (status: string) => status === 'success' ? 'online' : (status === 'failed' ? 'offline' : 'untested')

const isDeleteMatch = computed(() => {
  const target = deleteTarget.value?.name || ''
  return deleteConfirmInput.value.trim().toLowerCase() === target.trim().toLowerCase() && target.length > 0
})

const closeTransientUi = () => {
  isAddDialogOpen.value = false
  isEditDialogOpen.value = false
  isTestDialogOpen.value = false
  isDeleteConfirmOpen.value = false
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer !== null) {
    window.clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

const queryListData = async (shouldSyncRoute = true) => {
  if (shouldSyncRoute) await syncRouteQuery()
  state.loading = true
  state.error = ''
  try {
    const params: Record<string, unknown> = {
      page: state.currPage,
      page_size: state.pageSize
    }
    if (state.search.trim()) params.name = state.search.trim()
    if (selectedType.value !== 'all') params.type = selectedType.value
    if (selectedStatus.value === STATUS_UNTESTED) params.status = 'untested'
    else if (selectedStatus.value !== 'all') params.status = selectedStatus.value
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime

    const res = await mqApi.list(params)
    if (res?.data?.code === 200) {
      state.tableData = res.data.data?.list || []
      state.total = res.data.data?.total || 0
      state.stale = false
      return
    }
    state.error = res?.data?.msg || '获取数据源列表失败'
    state.stale = state.tableData.length > 0
    notifyError(state.error)
  } catch (error) {
    state.error = '获取数据源列表时发生错误，请检查网络后重试'
    state.stale = state.tableData.length > 0
    notifyError('获取数据源列表时发生错误')
  } finally {
    state.loading = false
  }
}

const setupAutoRefreshByPolicy = async () => {
  stopAutoRefresh()
  try {
    const rsp = await settingsApi.get('mq_status_policy')
    const data = rsp?.data?.data || {}
    const enabled = data.enabled === 'true'
    const intervalSeconds = Number(data.interval_seconds || 300)
    if (!enabled || Number.isNaN(intervalSeconds) || intervalSeconds < 10) return
    autoRefreshTimer = window.setInterval(() => {
      queryListData(false)
    }, intervalSeconds * 1000)
  } catch {}
}

const filterFunc = async () => {
  state.currPage = 1
  await queryListData()
}

const resetFilters = async () => {
  state.search = ''
  selectedType.value = 'all'
  selectedStatus.value = 'all'
  state.currPage = 1
  await queryListData()
}

const handlePageChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await queryListData()
}

const openEditDialog = (item: MQSourceItem) => {
  editData.value = { ...item }
  isEditDialogOpen.value = true
}

const openTestDialog = (item: MQSourceItem) => {
  editData.value = { ...item }
  testResult.value = null
  isTestDialogOpen.value = true
}

const openDeleteConfirm = (item: MQSourceItem) => {
  deleteTarget.value = item
  deleteConfirmInput.value = ''
  isDeleteConfirmOpen.value = true
}

const closeDeleteConfirm = () => {
  isDeleteConfirmOpen.value = false
  deleteConfirmInput.value = ''
  deleteTarget.value = null
}

const handleTestConnection = async () => {
  if (!editData.value) return
  try {
    const res = await mqApi.test(editData.value.id)
    if (res?.data?.code === 200) {
      testResult.value = res.data.data
      if (testResult.value?.success) notifySuccess('连接测试成功')
      else notifyError(testResult.value?.error || '连接测试失败')
      await queryListData(false)
      return
    }
    notifyError(res?.data?.msg || '连接测试失败')
  } catch (error) {
    notifyError('连接测试失败')
    testResult.value = { success: false, error: '网络错误' }
  }
}

const handleDelete = async () => {
  if (!deleteTarget.value) return
  try {
    const res = await mqApi.remove(deleteTarget.value.id)
    if (res?.data?.code === 200) {
      notifySuccess('删除成功')
      closeDeleteConfirm()
      await queryListData(false)
      return
    }
    notifyError(res?.data?.msg || '删除失败')
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '删除失败')
  }
}

const handleSaveSuccess = async () => {
  isAddDialogOpen.value = false
  isEditDialogOpen.value = false
  await queryListData(false)
}

const handleExport = async () => {
  try {
    const params: Record<string, unknown> = { page: state.currPage, page_size: state.pageSize }
    if (state.search.trim()) params.name = state.search.trim()
    if (selectedType.value !== 'all') params.type = selectedType.value
    if (selectedStatus.value !== 'all') params.status = selectedStatus.value
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const rsp = await mqApi.export(params)
    downloadBlob(rsp.data, `mq-sources-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('数据源导出成功')
  } catch (error) {
    notifyError('数据源导出失败')
  }
}

onMounted(() => {
  state.search = route.query.name?.toString() || ''
  selectedType.value = route.query.type?.toString() || 'all'
  selectedStatus.value = route.query.status?.toString() || 'all'
  if (!typeOptions.some((item) => item.value === selectedType.value)) selectedType.value = 'all'
  if (!statusOptions.some((item) => item.value === selectedStatus.value)) selectedStatus.value = 'all'
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.page_size, state.pageSize)
  queryListData(false)
  setupAutoRefreshByPolicy()
})

onBeforeRouteLeave(() => {
  closeTransientUi()
  stopAutoRefresh()
})

onUnmounted(() => {
  closeTransientUi()
  stopAutoRefresh()
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="page-toolbar">
        <el-input v-model="state.search" clearable :placeholder="messages.searchPlaceholder" class="w-full md:!w-[260px]" @keyup.enter="filterFunc">
          <template #prefix><Search /></template>
        </el-input>
        <el-select v-model="selectedType" class="w-full md:!w-[160px]" @change="filterFunc">
          <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
        <el-select v-model="selectedStatus" class="w-full md:!w-[150px]" @change="filterFunc">
          <el-option v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
        <el-button :icon="Search" type="primary" @click="filterFunc">{{ messages.search }}</el-button>
        <el-button :icon="Refresh" @click="resetFilters">{{ messages.reset }}</el-button>
        <div class="flex-1" />
        <el-button :icon="Download" @click="handleExport">{{ messages.export }}</el-button>
        <el-button v-permission="'data:mq-source:add'" :icon="Plus" type="primary" @click="isAddDialogOpen = true">{{ messages.add }}</el-button>
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
      <AppTable :data="state.tableData" :columns="visibleColumns" :loading="state.loading" :error="state.error" :stale="state.stale" :empty-text="messages.empty" @retry="queryListData">
        <template #index="{ index }">{{ (state.currPage - 1) * state.pageSize + index + 1 }}</template>
        <template #name="{ row }"><AppTruncate :text="row.name" :title="messages.queueName" /></template>
        <template #type="{ row }"><el-tag effect="plain">{{ getTypeText(row.type as string) }}</el-tag></template>
        <template #namesrv_addr="{ row }"><AppTruncate :text="row.namesrv_addr" :title="messages.queueAddress" /></template>
        <template #binding_count="{ row }"><el-tag type="info" effect="plain">{{ row.binding_count || 0 }}{{ messages.subscriptionUnit }}</el-tag></template>
        <template #last_test_status="{ row }"><AppStatusTag :status="normalizedStatus(row.last_test_status as string)" /></template>
        <template #last_test_time="{ row }">{{ row.last_test_time || '-' }}</template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: messages.edit, kind: 'write', permission: 'data:mq-source:edit', onClick: () => openEditDialog(row as MQSourceItem) },
            { key: 'delete', label: messages.delete, kind: 'write', permission: 'data:mq-source:delete', danger: true, onClick: () => openDeleteConfirm(row as MQSourceItem) },
            { key: 'test', label: messages.test, kind: 'write', permission: 'data:mq-source:test', onClick: () => openTestDialog(row as MQSourceItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState :description="messages.empty" />
        </template>
      </AppTable>
      <div class="px-4">
        <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePageChange" />
      </div>
    </el-card>

    <AppFormDrawer v-model="isAddDialogOpen" :title="messages.add" size="640px" body-mode="managed" :show-footer="false">
      <MQSourceForm @success="handleSaveSuccess" />
    </AppFormDrawer>

    <AppFormDrawer v-model="isEditDialogOpen" :title="messages.editTitle" size="640px" body-mode="managed" :show-footer="false">
      <MQSourceForm v-if="editData" :data="editData" @success="handleSaveSuccess" />
    </AppFormDrawer>

    <el-dialog v-model="isTestDialogOpen" :title="`${messages.testTitlePrefix}${editData?.name || ''}`" width="min(520px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body>
      <div class="space-y-4">
        <div class="text-sm text-muted-foreground">
          {{ messages.testingAddress }}<code class="break-all rounded bg-muted px-1 py-0.5">{{ editData?.namesrv_addr }}</code>
        </div>
        <el-alert v-if="testResult" :type="testResult.success ? 'success' : 'error'" :title="testResult.success ? (testResult.message || messages.connectionSucceeded) : (testResult.error || messages.connectionFailed)" show-icon :closable="false" />
        <p v-else>{{ messages.testHelp }}</p>
      </div>
      <template #footer>
        <el-button @click="isTestDialogOpen = false">{{ messages.close }}</el-button>
        <el-button type="primary" :disabled="!editData" @click="handleTestConnection">{{ messages.startTest }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="isDeleteConfirmOpen" :title="messages.confirmDelete" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body @closed="closeDeleteConfirm">
      <div class="space-y-2">
        <p class="text-sm text-muted-foreground">{{ messages.confirmNamePrefix }}<strong>{{ deleteTarget?.name }}</strong>{{ messages.confirmNameSuffix }}</p>
        <el-input v-model="deleteConfirmInput" :placeholder="messages.nameInputPlaceholder" />
        <p v-if="deleteConfirmInput && !isDeleteMatch" class="text-sm text-red-500">{{ messages.nameMismatch }}</p>
      </div>
      <template #footer>
        <el-button @click="closeDeleteConfirm">{{ messages.cancel }}</el-button>
        <el-button type="danger" :disabled="!isDeleteMatch" @click="handleDelete">{{ messages.delete }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.mq-test-dialog { display: grid; gap: 12px; }
.mq-test-identity, .mq-test-target, .mq-test-result { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.mq-test-identity { display: flex; align-items: center; gap: 12px; padding: 14px; }
.mq-test-mark { display: grid; place-items: center; flex: none; width: 36px; height: 36px; border-radius: 8px; background: color-mix(in srgb, var(--brand-500) 10%, transparent); color: var(--brand-700); font: 800 10px monospace; }
.mq-test-identity h3 { margin: 0; font-size: 14px; }
.mq-test-identity p, .mq-test-result p { margin: 3px 0 0; color: var(--admin-text-muted); font-size: 11px; }
.mq-test-target header, .mq-test-result header { display: flex; align-items: center; justify-content: space-between; padding: 9px 12px; border-bottom: 1px solid var(--app-overlay-border); }
.mq-test-target h4, .mq-test-result h4 { margin: 0; font-size: 12px; }
.mq-test-target header span, .mq-test-result header span { color: var(--admin-text-muted); font-size: 10px; }
.mq-test-target dl { margin: 0; }
.mq-test-target dl > div { display: grid; grid-template-columns: 100px minmax(0, 1fr); padding: 9px 12px; border-top: 1px solid var(--app-overlay-border); }
.mq-test-target dl > div:first-child { border-top: 0; }
.mq-test-target dt { color: var(--admin-text-muted); font-size: 11px; }
.mq-test-target dd { margin: 0; font-size: 11px; overflow-wrap: anywhere; }
.mq-test-result > :not(header) { margin: 12px; }
</style>
