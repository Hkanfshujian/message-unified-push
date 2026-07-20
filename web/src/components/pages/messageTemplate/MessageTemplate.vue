<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Plus, Search } from '@element-plus/icons-vue'
import TemplateApiViewer from './TemplateApiViewer.vue'
import TemplateInstanceConfig from './TemplateInstanceConfig.vue'
import TemplateEditor from './TemplateEditor.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import AppTruncate from '@/components/ui/AppTruncate.vue'
import { templatesApi } from '@/api/templates'
import { getPageSize } from '@/util/pageUtils'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { appendDateRangeQuery, pickDateRangeQuery } from '@/util/routeQuery'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
import { useRoute, useRouter } from 'vue-router'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.messageTemplate

interface MessageTemplate {
  id: string  // 模板ID是字符串类型（UUID）
  name: string
  description: string
  text_template: string
  html_template: string
  markdown_template: string
  placeholders: string
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status: string
  created_on: string
  modified_on: string
  cron_msg_count?: number
}

const route = useRoute()
const router = useRouter()

let state = reactive({
  tableData: [] as MessageTemplate[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize() as number,
  search: '',
  status: 'all',
  loading: false
})

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 120 },
  { prop: 'name', label: '模板名称', minWidth: 220 },
  { prop: 'description', label: '描述', minWidth: 180 },
  { prop: 'formats', label: '支持格式', minWidth: 180 },
  { prop: 'relations', label: '外部关联', width: 120, align: 'center' },
  { prop: 'status', label: '状态', width: 100, align: 'center' },
  { prop: 'created_on', label: '创建时间', minWidth: 170 },
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

const statusOptions = [
  { value: 'all', label: '全部状态' },
  { value: 'enabled', label: '启用' },
  { value: 'disabled', label: '禁用' }
]

// API代码查看器状态
const isApiViewerOpen = ref(false)
const selectedTemplateForApi = ref<MessageTemplate | null>(null)

// 配置实例状态
const isInstanceConfigOpen = ref(false)
const selectedTemplateForInstance = ref<MessageTemplate | null>(null)

// 模板编辑器状态
const isEditorOpen = ref(false)
const isEditing = ref(false)
const selectedTemplateForEdit = ref<MessageTemplate | null>(null)

const isDeleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<MessageTemplate | null>(null)

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.floor(n)
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {}
  const text = state.search.trim()
  if (text) nextQuery.text = text
  if (state.status && state.status !== 'all') nextQuery.status = state.status
  nextQuery.page = String(state.currPage)
  nextQuery.size = String(state.pageSize)
  appendDateRangeQuery(nextQuery, route.query as Record<string, unknown>)
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

const queryListData = async (page: number, size: number, text = '', status = '') => {
  state.loading = true
  try {
    const params: any = { page, size, text, status }
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const rsp = await templatesApi.list(params)
    if (rsp?.data?.code === 200) {
      state.tableData = rsp.data.data?.lists || []
      state.total = rsp.data.data?.total || 0
      return
    }
    notifyError(rsp?.data?.msg || '获取模板列表失败')
  } catch (error) {
    notifyError('获取模板列表时发生错误')
  } finally {
    state.loading = false
  }
}

const reloadList = async () => {
  await syncRouteQuery()
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const filterFunc = async () => {
  state.currPage = 1
  await reloadList()
}

const handlePageChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await reloadList()
}

const openAddDialog = () => {
  isEditing.value = false
  selectedTemplateForEdit.value = null
  isEditorOpen.value = true
}

const openEditDialog = (template: MessageTemplate) => {
  isEditing.value = true
  selectedTemplateForEdit.value = template
  isEditorOpen.value = true
}

const handleEditorSaved = async () => {
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const deleteTemplate = async (id: string) => {
  const rsp = await templatesApi.remove(id)
  if (rsp.status === 200 && rsp.data.code === 200) {
    notifySuccess(rsp.data.msg)
    const statusParam = state.status === 'all' ? '' : state.status
    await queryListData(state.currPage, state.pageSize, state.search, statusParam)
  }
}

const openDeleteConfirm = (template: MessageTemplate) => {
  // 如果存在外部关联（目前主要是定时消息），先提示并展示关联列表，不进入删除确认流程
  if ((template.cron_msg_count ?? 0) > 0) {
    notifyError('当前模板存在外部关联，请先在相关功能中删除关联后再删除模板')
    openRelationDialog(template)
    return
  }

  deleteTarget.value = template
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

const showDeleteError = computed(() => {
  return deleteConfirmInput.value.length > 0 && !isDeleteMatch.value
})

const handleConfirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  await deleteTemplate(deleteTarget.value.id)
  closeDeleteConfirm()
}

const isRelationDialogOpen = ref(false)
const relationList = ref<Array<{ index: number; type: string; id: string; name: string }>>([])
const relationLoading = ref(false)
const relationTemplateName = ref('')
const relationColumns: AppTableColumn[] = [
  { prop: 'index', label: messages.index, width: 70, align: 'center' },
  { prop: 'type', label: messages.relationType, width: 120 },
  { prop: 'id', label: 'ID', minWidth: 150 },
  { prop: 'name', label: messages.name, minWidth: 180 }
]

const openRelationDialog = async (template: MessageTemplate) => {
  relationTemplateName.value = template.name
  relationList.value = []
  relationLoading.value = true
  isRelationDialogOpen.value = true
  try {
    const rsp = await templatesApi.relations(template.id)
    const list = rsp.data.data?.relations || []
    relationList.value = list.map((item: any, idx: number) => ({
      index: idx + 1,
      type: item.type || '定时消息',
      id: item.id || '',
      name: item.name || ''
    }))
  } catch (error) {
    notifyError('获取关联信息失败')
  } finally {
    relationLoading.value = false
  }
}

// 打开API查看器
const handleViewApi = (template: MessageTemplate) => {
  selectedTemplateForApi.value = template
  isApiViewerOpen.value = true
}

// 打开配置实例
const handleConfigInstance = (template: MessageTemplate) => {
  selectedTemplateForInstance.value = template
  isInstanceConfigOpen.value = true
}

// 查看日志
const handleViewLogs = (template: MessageTemplate) => {
  // 跳转到发信日志页面，携带 taskid 参数（传递模板 id）
  router.push(`/logs/task?taskid=${template.id}`)
}

onMounted(async () => {
  state.search = typeof route.query.text === 'string'
    ? route.query.text
    : (typeof route.query.name === 'string' ? route.query.name : '')
  state.status = typeof route.query.status === 'string' && route.query.status ? route.query.status : 'all'
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.size, state.pageSize)
  await syncRouteQuery()
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="page-toolbar">
        <el-input v-model="state.search" clearable :placeholder="messages.searchPlaceholder" class="w-full md:!w-[280px]" @keyup.enter="filterFunc">
          <template #prefix><Search /></template>
        </el-input>
        <el-select v-model="state.status" class="w-full md:!w-[150px]" @change="filterFunc">
          <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button :icon="Search" type="primary" @click="filterFunc">{{ messages.search }}</el-button>
        <div class="flex-1" />
        <el-button v-permission="'message:template:add'" :icon="Plus" type="primary" @click="openAddDialog">{{ messages.create }}</el-button>
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
        <template #description="{ row }">
          <AppTruncate :text="row.description || '-'" :title="messages.description" width="560px" />
        </template>
        <template #formats="{ row }">
          <div class="flex flex-wrap gap-1">
            <el-tag v-if="row.text_template" size="small" effect="plain">Text</el-tag>
            <el-tag v-if="row.html_template" size="small" effect="plain">HTML</el-tag>
            <el-tag v-if="row.markdown_template" size="small" effect="plain">Markdown</el-tag>
          </div>
        </template>
        <template #relations="{ row }">
          <el-button link type="primary" @click="openRelationDialog(row as MessageTemplate)">{{ row.cron_msg_count ?? 0 }}</el-button>
        </template>
        <template #status="{ row }">
          <AppStatusTag :status="row.status" />
        </template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: messages.edit, kind: 'write', permission: 'message:template:edit', onClick: () => openEditDialog(row as MessageTemplate) },
            { key: 'delete', label: messages.delete, kind: 'write', permission: 'message:template:delete', danger: true, onClick: () => openDeleteConfirm(row as MessageTemplate) },
            { key: 'instance', label: messages.instance, kind: 'write', permission: 'message:template:instance', onClick: () => handleConfigInstance(row as MessageTemplate) },
            { key: 'logs', label: messages.logs, kind: 'view', onClick: () => handleViewLogs(row as MessageTemplate) },
            { key: 'api', label: messages.api, kind: 'view', onClick: () => handleViewApi(row as MessageTemplate) }
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

    <!-- Template editor -->
    <TemplateEditor
      :open="isEditorOpen"
      :is-editing="isEditing"
      :template-data="selectedTemplateForEdit"
      @update:open="isEditorOpen = $event"
      @saved="handleEditorSaved"
    />

    <!-- API code viewer -->
    <TemplateApiViewer 
      :open="isApiViewerOpen" 
      :template-data="selectedTemplateForApi || undefined"
      @update:open="isApiViewerOpen = $event"
    />

    <!-- Instance configuration -->
    <TemplateInstanceConfig 
      :open="isInstanceConfigOpen" 
      :template-data="selectedTemplateForInstance"
      @update:open="isInstanceConfigOpen = $event"
    />

    <el-dialog v-model="isDeleteConfirmOpen" :title="messages.confirmDelete" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body @closed="closeDeleteConfirm">
        <div class="space-y-2">
          <div class="text-sm text-muted-foreground">
            {{ messages.confirmNamePrefix }}
            <span v-if="deleteTarget?.name" class="text-red-500 font-semibold mx-1">{{ deleteTarget.name }}</span>
            {{ messages.confirmNameSuffix }}
          </div>
          <el-input v-model="deleteConfirmInput" :maxlength="25" :placeholder="messages.namePlaceholder" />
          <div v-if="showDeleteError" class="text-xs text-red-500">{{ messages.nameMismatch }}</div>
        </div>
        <template #footer>
          <el-button @click="closeDeleteConfirm">{{ messages.cancel }}</el-button>
          <el-button type="danger" :disabled="!isDeleteMatch" @click="handleConfirmDelete">{{ messages.confirmDelete }}</el-button>
        </template>
    </el-dialog>

    <el-dialog
      v-model="isRelationDialogOpen"
      :title="`${messages.relationTitlePrefix}${relationTemplateName}`"
      width="min(680px, calc(100vw - 24px))"
      class="app-nested-dialog relation-dialog"
      append-to-body
      destroy-on-close
    >
      <div class="relation-dialog-content">
        <div class="relation-dialog-intro">
          <span>以下功能正在使用此模板，删除模板前需要先解除关联。</span>
          <el-tag size="small" effect="plain">共 {{ relationList.length }} 项</el-tag>
        </div>
        <AppTable
          :data="relationList"
          :columns="relationColumns"
          :loading="relationLoading"
          :empty-text="messages.noRelations"
          size="small"
        >
          <template #id="{ row }">
            <code class="relation-dialog-id">{{ row.id || '-' }}</code>
          </template>
          <template #name="{ row }">
            <AppTruncate :text="String(row.name || '-')" :max-length="32" />
          </template>
        </AppTable>
      </div>
      <template #footer>
        <el-button @click="isRelationDialogOpen = false">{{ messages.close }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.relation-dialog-content {
  display: grid;
  gap: 14px;
}

.relation-dialog-intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--muted-foreground);
  font-size: 13px;
  line-height: 1.6;
}

.relation-dialog-id {
  color: var(--foreground);
  font-size: 12px;
  overflow-wrap: anywhere;
}

.relation-dialog-content :deep(.el-table .cell) {
  line-height: 1.4;
  overflow-wrap: anywhere;
}

@media (max-width: 640px) {
  .relation-dialog-intro {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
