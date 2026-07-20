<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Download, Plus, Refresh, Search } from '@element-plus/icons-vue'
import AddWays from './AddWays.vue'
import EditWays from './EditWays.vue'
import AppDetailDrawer from '@/components/ui/AppDetailDrawer.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import { channelsApi } from '@/api/channels'
import { CONSTANT } from '@/constant'
import { getPageSize } from '@/util/pageUtils'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { appendDateRangeQuery, pickDateRangeQuery } from '@/util/routeQuery'
import { downloadBlob, notifyError, notifySuccess } from '@/util/uiFeedback'
import { useRoute, useRouter } from 'vue-router'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.sendWays

interface WayItem extends Record<string, unknown> {
  id: number
  name: string
  type: string
  auth?: string
  config?: string
  created_on: string
  modified_on: string
  status: number
}

const route = useRoute()
const router = useRouter()

const state = reactive({
  tableData: [] as WayItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false
})

const selectedStatus = ref('all')
const selectedChannelType = ref('all')
const isConfigDrawerOpen = ref(false)
const selectedConfig = ref('')
const selectedChannelName = ref('')
const selectedChannel = ref<WayItem | null>(null)
const isAddChannelDrawerOpen = ref(false)
const isEditChannelDrawerOpen = ref(false)
const editChannelData = ref<WayItem | null>(null)
const isDeleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<WayItem | null>(null)

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 90 },
  { prop: 'name', label: '渠道名称', minWidth: 180 },
  { prop: 'type', label: '发信方式类型', minWidth: 160 },
  { prop: 'created_on', label: '创建时间', minWidth: 170 },
  { prop: 'modified_on', label: '更新时间', minWidth: 170 },
  { prop: 'status', label: '状态', width: 100, align: 'center' },
  { prop: 'actions', label: '操作', width: 220, align: 'center', fixed: 'right' }
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

const channelTypeOptions = computed(() => [
  { value: 'all', label: '全部类型' },
  ...CONSTANT.WAYS_DATA.map((item: { type: string; label?: string; name?: string }) => ({ value: item.type, label: item.label || item.name || item.type }))
])

const statusOptions = [
  { value: 'all', label: '全部状态' },
  { value: '1', label: '启用' },
  { value: '0', label: '停用' }
]

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : fallback
}

const getWayTypeText = (type: string) => {
  const wayData = CONSTANT.WAYS_DATA.find((item: { type: string }) => item.type === type)
  return wayData ? wayData.label : type
}

const selectedChannelConfig = computed(() => CONSTANT.WAYS_DATA.find((way: any) => way.type === selectedChannel.value?.type))
const sensitiveConfigKeyPattern = /(passw|passwd|password|secret|token|access[_-]?key|push[_-]?key|bot[_-]?token|corp[_-]?secret|appsecret|^key$|^iv$)/i
const selectedConfigEntries = computed(() => {
  let config: Record<string, unknown> = {}
  try {
    config = JSON.parse(selectedConfig.value || '{}')
  } catch {
    return []
  }
  const labelMap = new Map((selectedChannelConfig.value?.inputs || []).map((input: any) => [input.col, input.subLabel || input.label]))
  return Object.entries(config).map(([key, value]) => ({
    key,
    label: labelMap.get(key) || key,
    value: sensitiveConfigKeyPattern.test(key)
      ? '••••••••'
      : typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value ?? '')
  }))
})

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {}
  const name = state.search.trim()
  if (name) nextQuery.name = name
  if (selectedChannelType.value !== 'all') nextQuery.channel_type = selectedChannelType.value
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
      name: state.search.trim(),
      type: selectedChannelType.value === 'all' ? '' : selectedChannelType.value
    }
    if (selectedStatus.value !== 'all') params.status = selectedStatus.value
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const rsp = await channelsApi.list(params)
    if (rsp?.data?.code === 200) {
      state.tableData = rsp.data.data?.lists || []
      state.total = rsp.data.data?.total || 0
      return
    }
    notifyError(rsp?.data?.msg || '获取渠道列表失败')
  } catch (error) {
    notifyError('获取渠道列表时发生错误')
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
  selectedChannelType.value = 'all'
  selectedStatus.value = 'all'
  state.currPage = 1
  await reloadList()
}

const handlePageChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await reloadList()
}

const openEditChannelDrawer = (channel: WayItem) => {
  editChannelData.value = channel
  isEditChannelDrawerOpen.value = true
}

const openConfigDrawer = (channel: WayItem) => {
  selectedChannel.value = channel
  selectedChannelName.value = channel.name || '未命名渠道'
  selectedConfig.value = channel.auth || channel.config || '{}'
  isConfigDrawerOpen.value = true
}

const handleSaveChannel = async () => {
  isAddChannelDrawerOpen.value = false
  await queryListData()
}

const handleEditChannel = async () => {
  isEditChannelDrawerOpen.value = false
  await queryListData()
}

const openDeleteConfirm = (channel: WayItem) => {
  deleteTarget.value = channel
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
  return target.length > 0 && deleteConfirmInput.value.trim().toLowerCase() === target.trim().toLowerCase()
})

const handleConfirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  const rsp = await channelsApi.remove(deleteTarget.value.id)
  if (rsp?.data?.code === 200) {
    notifySuccess(rsp.data.msg || '删除成功')
    closeDeleteConfirm()
    await queryListData()
    return
  }
  notifyError(rsp?.data?.msg || '删除失败')
}

const handleExport = async () => {
  try {
    const params: Record<string, unknown> = {
      page: state.currPage,
      size: state.pageSize,
      name: state.search.trim(),
      type: selectedChannelType.value === 'all' ? '' : selectedChannelType.value
    }
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const rsp = await channelsApi.export(params)
    downloadBlob(rsp.data, `sendways-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('渠道导出成功')
  } catch (error) {
    notifyError('渠道导出失败')
  }
}

onMounted(async () => {
  state.search = route.query.name?.toString() || ''
  selectedChannelType.value = route.query.channel_type?.toString() || 'all'
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
        <el-select v-model="selectedChannelType" class="w-full md:!w-[180px]" @change="filterFunc">
          <el-option v-for="option in channelTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-select v-model="selectedStatus" class="w-full md:!w-[140px]" @change="filterFunc">
          <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button :icon="Search" type="primary" @click="filterFunc">{{ messages.search }}</el-button>
        <el-button :icon="Refresh" @click="resetFilters">{{ messages.reset }}</el-button>
        <div class="flex-1" />
        <el-button :icon="Download" @click="handleExport">{{ messages.export }}</el-button>
        <el-button v-permission="'message:sendways:add'" :icon="Plus" type="primary" @click="isAddChannelDrawerOpen = true">{{ messages.add }}</el-button>
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
        <template #type="{ row }">
          <el-tag effect="plain">{{ getWayTypeText(row.type as string) }}</el-tag>
        </template>
        <template #status="{ row }"><AppStatusTag :status="row.status" /></template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: messages.edit, kind: 'write', permission: 'message:sendways:edit', onClick: () => openEditChannelDrawer(row as WayItem) },
            { key: 'delete', label: messages.delete, kind: 'write', permission: 'message:sendways:delete', danger: true, onClick: () => openDeleteConfirm(row as WayItem) },
            { key: 'view', label: messages.view, kind: 'view', onClick: () => openConfigDrawer(row as WayItem) }
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

    <AppFormDrawer v-model="isAddChannelDrawerOpen" :title="messages.addTitle" size="min(860px, 96vw)" body-mode="managed" density="compact" :show-footer="false">
      <AddWays v-model:open="isAddChannelDrawerOpen" @save="handleSaveChannel" />
    </AppFormDrawer>

    <AppFormDrawer v-model="isEditChannelDrawerOpen" :title="messages.editTitle" size="min(860px, 96vw)" body-mode="managed" density="compact" :show-footer="false">
      <EditWays v-model:open="isEditChannelDrawerOpen" :edit-data="editChannelData" @save="handleEditChannel" />
    </AppFormDrawer>

    <el-dialog v-model="isDeleteConfirmOpen" :title="messages.confirmDelete" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body @closed="closeDeleteConfirm">
      <div class="space-y-2">
        <div class="text-sm text-muted-foreground">{{ messages.confirmDeletePrefix }}{{ deleteTarget?.name }}</div>
        <el-input v-model="deleteConfirmInput" maxlength="50" :placeholder="messages.namePlaceholder" />
        <div v-if="deleteConfirmInput && !isDeleteMatch" class="text-xs text-red-500">{{ messages.nameMismatch }}</div>
      </div>
      <template #footer>
        <el-button @click="closeDeleteConfirm">{{ messages.cancel }}</el-button>
        <el-button type="danger" :disabled="!isDeleteMatch" @click="handleConfirmDelete">{{ messages.confirmDelete }}</el-button>
      </template>
    </el-dialog>

    <AppDetailDrawer v-model="isConfigDrawerOpen" :title="`${messages.viewTitlePrefix}${selectedChannelName}`" size="760px">
      <div v-if="selectedChannel" class="send-way-detail">
        <section class="send-way-detail-summary">
          <div>
            <span class="send-way-detail-kicker">{{ messages.channelType }}</span>
            <h3>{{ getWayTypeText(selectedChannel.type) }}</h3>
            <code>{{ selectedChannel.type }}</code>
          </div>
          <AppStatusTag :status="selectedChannel.status" />
        </section>

        <section class="send-way-detail-section">
          <header><h4>{{ messages.basicInfo }}</h4></header>
          <dl class="send-way-detail-grid">
            <div><dt>{{ messages.channelName }}</dt><dd>{{ selectedChannel.name || '-' }}</dd></div>
            <div><dt>{{ messages.channelId }}</dt><dd>{{ selectedChannel.id }}</dd></div>
            <div><dt>{{ messages.createdAt }}</dt><dd>{{ selectedChannel.created_on || '-' }}</dd></div>
            <div><dt>{{ messages.updatedAt }}</dt><dd>{{ selectedChannel.modified_on || '-' }}</dd></div>
          </dl>
        </section>

        <section class="send-way-detail-section">
          <header><h4>{{ messages.connectionConfig }}</h4><span>{{ selectedConfigEntries.length }}{{ messages.configItemUnit }}</span></header>
          <dl v-if="selectedConfigEntries.length" class="send-way-detail-config">
            <div v-for="item in selectedConfigEntries" :key="item.key">
              <dt><span>{{ item.label }}</span><code>{{ item.key }}</code></dt>
              <dd>{{ item.value || '-' }}</dd>
            </div>
          </dl>
          <pre v-else class="channel-config-pre">{{ selectedConfig }}</pre>
        </section>
      </div>
    </AppDetailDrawer>
  </div>
</template>

<style scoped>
.send-way-detail {
  display: grid;
  gap: 14px;
  max-width: 920px;
  margin: 0 auto;
}

.send-way-detail-summary,
.send-way-detail-section {
  overflow: hidden;
  border: 1px solid var(--app-overlay-border);
  border-radius: 9px;
  background: var(--app-overlay-surface);
}

.send-way-detail-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
}

.send-way-detail-kicker {
  display: block;
  margin-bottom: 4px;
  color: var(--admin-text-muted);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.send-way-detail-summary h3 {
  display: inline;
  margin: 0 8px 0 0;
  font-size: 16px;
  font-weight: 700;
}

.send-way-detail-summary code,
.send-way-detail-config dt code {
  color: var(--admin-text-muted);
  font-family: monospace;
  font-size: 10px;
}

.send-way-detail-section > header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 46px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--app-overlay-border);
}

.send-way-detail-section > header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
}

.send-way-detail-section > header span {
  color: var(--admin-text-muted);
  font-size: 11px;
}

.send-way-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.send-way-detail-grid > div {
  padding: 13px 16px;
}

.send-way-detail-grid > div:nth-child(even) {
  border-left: 1px solid var(--app-overlay-border);
}

.send-way-detail-grid > div:nth-child(n + 3) {
  border-top: 1px solid var(--app-overlay-border);
}

.send-way-detail-grid dt,
.send-way-detail-config dt {
  margin-bottom: 5px;
  color: var(--admin-text-muted);
  font-size: 11px;
  font-weight: 600;
}

.send-way-detail-grid dd,
.send-way-detail-config dd {
  margin: 0;
  color: var(--foreground);
  font-size: 12px;
  line-height: 1.55;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.send-way-detail-config {
  margin: 0;
  padding: 5px 0;
}

.send-way-detail-config > div {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  min-height: 44px;
  margin: 0 10px;
  padding: 8px 6px;
  border-bottom: 1px solid var(--app-overlay-border);
}

.send-way-detail-config > div:last-child {
  border-bottom: 0;
}

.send-way-detail-config dt {
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin: 0;
  padding: 0 10px;
}

.send-way-detail-config dt span {
  color: var(--foreground);
  font-size: 12px;
}

.send-way-detail-config dd {
  display: flex;
  align-items: center;
  padding: 0 12px;
}

.channel-config-pre {
  margin: 12px;
  padding: 14px;
  overflow: auto;
  border-radius: 7px;
  background: var(--admin-surface-muted);
  font-family: monospace;
  font-size: 11px;
  white-space: pre-wrap;
}

@media (max-width: 760px) {
  .send-way-detail-grid {
    grid-template-columns: 1fr;
  }

  .send-way-detail-grid > div:nth-child(even) {
    border-left: 0;
  }

  .send-way-detail-grid > div + div {
    border-top: 1px solid var(--app-overlay-border);
  }

  .send-way-detail-config > div {
    grid-template-columns: 1fr;
    gap: 5px;
  }

  .send-way-detail-config dd {
    padding-inline: 10px;
  }
}
</style>
