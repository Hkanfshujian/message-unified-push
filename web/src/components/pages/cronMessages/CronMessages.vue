<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Download, Plus, Refresh, Search } from '@element-plus/icons-vue'
import AddCronMessages from './AddCronMessages.vue'
import EditCronMessages from './EditCronMessages.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import AppTruncate from '@/components/ui/AppTruncate.vue'
import { useRoute, useRouter } from 'vue-router'
import { scheduledMessagesApi } from '@/api/scheduledMessages'
import { useRbacStore } from '@/store'
import { getPageSize } from '@/util/pageUtils'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { appendDateRangeQuery, pickDateRangeQuery } from '@/util/routeQuery'
import { downloadBlob, notifyError, notifySuccess } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.cronMessages

interface CronMessageItem extends Record<string, unknown> {
  id: string
  name: string
  cron: string
  cron_expression: string
  template_id: string
  template_name: string
  ins_ids: string[]
  channel_names: string[]
  enable: number
  status: boolean
  created_on: string
  modified_on: string
  next_time?: string
}

const route = useRoute()
const router = useRouter()
const rbacStore = useRbacStore()

const state = reactive({
  tableData: [] as CronMessageItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false
})

const selectedStatus = ref('all')
const isAddCronMessageDialogOpen = ref(false)
const isEditCronMessageDialogOpen = ref(false)
const editCronMessageData = ref<CronMessageItem | null>(null)
const isDeleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<CronMessageItem | null>(null)

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 120 },
  { prop: 'name', label: '名称', minWidth: 180 },
  { prop: 'template', label: '模板', minWidth: 180 },
  { prop: 'channels', label: '渠道', minWidth: 220 },
  { prop: 'cron', label: 'Cron 表达式', minWidth: 150 },
  { prop: 'next_time', label: '下次执行时间', minWidth: 170 },
  { prop: 'created_on', label: '创建时间', minWidth: 170 },
  { prop: 'actions', label: '操作/状态', width: 220, align: 'center', fixed: 'right' }
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
    await reloadList()
  } finally {
    tableToolbar.refreshing = false
  }
}

const statusOptions = [
  { value: 'all', label: '全部状态' },
  { value: '1', label: '启用' },
  { value: '0', label: '停用' }
]

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : fallback
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {}
  const name = state.search.trim()
  if (name) nextQuery.name = name
  if (selectedStatus.value !== 'all') nextQuery.status = selectedStatus.value
  nextQuery.page = String(state.currPage)
  nextQuery.size = String(state.pageSize)
  appendDateRangeQuery(nextQuery, route.query as Record<string, unknown>)
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

const queryListData = async () => {
  state.loading = true
  try {
    const params: Record<string, unknown> = {
      page: state.currPage,
      size: state.pageSize,
      name: state.search.trim()
    }
    if (selectedStatus.value !== 'all') params.status = selectedStatus.value
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const rsp = await scheduledMessagesApi.list(params)
    if (rsp?.data?.code === 200) {
      state.tableData = rsp.data.data?.lists || []
      state.total = rsp.data.data?.total || 0
      return
    }
    notifyError(rsp?.data?.msg || '获取定时消息列表失败')
  } catch (error) {
    notifyError('获取定时消息列表时发生错误')
  } finally {
    state.loading = false
  }
}

const reloadList = async () => {
  await syncRouteQuery()
  await queryListData()
}

const filterFunc = async () => {
  state.currPage = 1
  await reloadList()
}

const resetFilters = async () => {
  state.search = ''
  selectedStatus.value = 'all'
  state.currPage = 1
  await reloadList()
}

const handlePageChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await reloadList()
}

const openEditCronMessageDialog = (cronMessage: CronMessageItem) => {
  editCronMessageData.value = cronMessage
  isEditCronMessageDialogOpen.value = true
}

const handleSaveCronMessage = async () => {
  isAddCronMessageDialogOpen.value = false
  await queryListData()
}

const handleEditCronMessage = async () => {
  isEditCronMessageDialogOpen.value = false
  await queryListData()
}

const openDeleteConfirm = (cronMessage: CronMessageItem) => {
  deleteTarget.value = cronMessage
  deleteConfirmInput.value = ''
  isDeleteConfirmOpen.value = true
}

const closeDeleteConfirm = () => {
  isDeleteConfirmOpen.value = false
  deleteConfirmInput.value = ''
  deleteTarget.value = null
}

const isDeleteMatch = computed(() => {
  const target = deleteTarget.value?.name || ''
  return deleteConfirmInput.value.trim().toLowerCase() === target.trim().toLowerCase() && target.length > 0
})

const handleViewLogs = (cronMessage: CronMessageItem) => {
  router.push(`/logs/task?taskid=${cronMessage.id}`)
}

const toggleStatus = async (cronMessage: CronMessageItem) => {
  const prevStatus = cronMessage.enable
  const newStatus = prevStatus ? 0 : 1
  const permission = newStatus === 1 ? 'message:cron:start' : 'message:cron:stop'
  if (!rbacStore.hasPermission(permission)) return
  cronMessage.enable = newStatus
  const rsp = await scheduledMessagesApi.setEnabled(cronMessage.id, newStatus === 1)
  if (rsp?.data?.code === 200) {
    notifySuccess(newStatus === 1 ? `已启用定时消息「${cronMessage.name}」` : `已停用定时消息「${cronMessage.name}」`)
    return
  }
  cronMessage.enable = prevStatus
  notifyError(rsp?.data?.msg || '更新定时消息状态失败')
}

const handleDelete = async (id: string) => {
  const rsp = await scheduledMessagesApi.remove(id)
  if (rsp?.data?.code === 200) {
    notifySuccess(rsp.data.msg || '删除成功')
    await queryListData()
    return
  }
  notifyError(rsp?.data?.msg || '删除失败')
}

const handleConfirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  await handleDelete(deleteTarget.value.id)
  closeDeleteConfirm()
}

const handleSendNow = async (cronMessage: CronMessageItem) => {
  const rsp = await scheduledMessagesApi.sendNow({
    id: cronMessage.id,
    template_id: cronMessage.template_id,
    name: cronMessage.name,
    title: cronMessage.name
  })
  if (rsp?.data?.code === 200) {
    notifySuccess(rsp.data.msg || '发送成功')
    return
  }
  notifyError(rsp?.data?.msg || '发送失败')
}

const handleExport = async () => {
  try {
    const params: Record<string, unknown> = {
      page: state.currPage,
      size: state.pageSize,
      name: state.search.trim()
    }
    if (selectedStatus.value !== 'all') params.status = selectedStatus.value
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const rsp = await scheduledMessagesApi.export(params)
    downloadBlob(rsp.data, `cronmessages-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('定时任务导出成功')
  } catch (error) {
    notifyError('定时任务导出失败')
  }
}

onMounted(async () => {
  state.search = route.query.name?.toString() || ''
  selectedStatus.value = route.query.status?.toString() || 'all'
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.size, state.pageSize)
  await syncRouteQuery()
  await queryListData()
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="page-toolbar">
        <el-input v-model="state.search" clearable :placeholder="messages.searchPlaceholder" class="w-full md:!w-[260px]" @keyup.enter="filterFunc">
          <template #prefix><Search /></template>
        </el-input>
        <el-select v-model="selectedStatus" class="w-full md:!w-[140px]" @change="filterFunc">
          <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button :icon="Search" type="primary" @click="filterFunc">{{ messages.search }}</el-button>
        <el-button :icon="Refresh" @click="resetFilters">{{ messages.reset }}</el-button>
        <div class="flex-1" />
        <el-button :icon="Download" @click="handleExport">{{ messages.export }}</el-button>
        <el-button v-permission="'message:cron:add'" :icon="Plus" type="primary" @click="isAddCronMessageDialogOpen = true">{{ messages.addTask }}</el-button>
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
      <AppTable :data="state.tableData" :columns="visibleColumns" :loading="state.loading" :empty-text="messages.empty">
        <template #name="{ row }">
          <AppTruncate :text="row.name" :title="messages.name" />
        </template>
        <template #template="{ row }">
          <AppTruncate :text="row.template_name || row.template_id" :title="messages.template" />
        </template>
        <template #channels="{ row }">
          <AppTruncate :text="(row.channel_names || []).join('、')" :title="messages.channels" />
        </template>
        <template #cron="{ row }">
          <code class="rounded border border-border bg-muted px-2 py-1 text-sm font-mono text-foreground">{{ row.cron }}</code>
        </template>
        <template #next_time="{ row }">
          {{ row.next_time || '-' }}
        </template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: messages.edit, kind: 'write', permission: 'message:cron:edit', onClick: () => openEditCronMessageDialog(row as CronMessageItem) },
            { key: 'stop', label: messages.disable, kind: 'write', permission: 'message:cron:stop', visible: row.enable === 1, danger: true, onClick: () => toggleStatus(row as CronMessageItem) },
            { key: 'start', label: messages.enable, kind: 'write', permission: 'message:cron:start', visible: row.enable !== 1, onClick: () => toggleStatus(row as CronMessageItem) },
            { key: 'delete', label: messages.delete, kind: 'write', permission: 'message:cron:delete', danger: true, onClick: () => openDeleteConfirm(row as CronMessageItem) },
            { key: 'send', label: messages.sendNow, kind: 'write', permission: 'message:cron:sendnow', onClick: () => handleSendNow(row as CronMessageItem) },
            { key: 'logs', label: messages.logs, kind: 'view', permission: 'message:sendlogs:view', onClick: () => handleViewLogs(row as CronMessageItem) }
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

    <AppFormDrawer v-model="isAddCronMessageDialogOpen" :title="messages.addTitle" size="620px" body-mode="managed" :show-footer="false">
      <AddCronMessages v-model:open="isAddCronMessageDialogOpen" @save="handleSaveCronMessage" @cancel="isAddCronMessageDialogOpen = false" />
    </AppFormDrawer>

    <AppFormDrawer v-model="isEditCronMessageDialogOpen" :title="messages.editTitle" size="620px" body-mode="managed" :show-footer="false">
      <EditCronMessages v-model:open="isEditCronMessageDialogOpen" :cron-message="editCronMessageData" @save="handleEditCronMessage" @cancel="isEditCronMessageDialogOpen = false" />
    </AppFormDrawer>

    <el-dialog v-model="isDeleteConfirmOpen" :title="messages.confirmDeleteTitle" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body @closed="closeDeleteConfirm">
      <div class="space-y-2">
        <div class="text-sm text-muted-foreground">
          {{ messages.confirmNamePrefix }}
          <span v-if="deleteTarget?.name" class="text-red-500 font-semibold mx-1">{{ deleteTarget.name }}</span>
          {{ messages.confirmNameSuffix }}
        </div>
        <el-input v-model="deleteConfirmInput" maxlength="50" :placeholder="messages.namePlaceholder" />
        <div v-if="deleteConfirmInput && !isDeleteMatch" class="text-xs text-red-500">{{ messages.nameMismatch }}</div>
      </div>
      <template #footer>
        <el-button @click="closeDeleteConfirm">{{ messages.cancel }}</el-button>
        <el-button type="danger" :disabled="!isDeleteMatch" @click="handleConfirmDelete">{{ messages.confirmDelete }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>
