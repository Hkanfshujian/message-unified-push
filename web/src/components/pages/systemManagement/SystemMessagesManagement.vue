<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import AppActionButton from '@/components/ui/AppActionButton.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { useMessageCenterStore } from '@/store/message-center'
import { useRbacStore } from '@/store'
import { rbacApi } from '@/api/rbac'
import type { MessageTargetScopeInput, SystemMessagePayload } from '@/api/api'
import { getPageSize } from '@/util/pageUtils'
import { confirmAction, notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'

interface SystemMessageRecord extends Omit<SystemMessagePayload, 'id' | 'status'> {
  id: string
  status: 'published' | 'draft' | 'deleted'
  created_on?: string
  modified_on?: string
}

interface SystemMessageListResponse {
  list?: SystemMessageRecord[]
  lists?: SystemMessageRecord[]
  total?: number
}

const store = useMessageCenterStore()
const rbacStore = useRbacStore()
const state = reactive({ list: [] as SystemMessageRecord[], total: 0, page: 1, pageSize: getPageSize(), keyword: '', type: '', status: '', publishRange: [] as [string, string] | [], loading: false, saving: false, error: '', stale: false })
const formOpen = ref(false)
const editingId = ref('')
const readonlyMode = ref(false)
const selectedRows = ref<SystemMessageRecord[]>([])
const roleOptions = ref<Array<{ id: number; name: string; code: string }>>([])
const groupOptions = ref<Array<{ id: number; name: string; code: string }>>([])
const userOptions = ref<Array<{ id: number; username: string }>>([])
const form = reactive<SystemMessagePayload>({ type: 'announcement', title: '', summary: '', content: '', status: 'draft', publish_time: '', effective_start_time: '', effective_end_time: '', is_pinned: false, target_scopes: [{ target_type: 'all', target_id: 'all' }] })
let filterTimer: number | undefined

const columns: AppTableColumn[] = [
  { prop: 'title', label: '标题', minWidth: 220 },
  { prop: 'publish_time', label: '发布时间', minWidth: 170 },
  { prop: 'status', label: '状态', width: 120 },
  { prop: 'actions', label: '操作', width: 160, align: 'center', fixed: 'right' }
]
const toolbarColumns: TableToolbarColumn[] = columns.map(column => ({ key: column.prop || column.label, label: column.label, required: column.prop === 'title' || column.prop === 'actions' }))
const tableToolbar = reactive(createTableToolbarState(toolbarColumns))
const visibleColumns = computed(() => getVisibleToolbarColumns(columns, tableToolbar.visibleColumns))
const statusLabelMap = { draft: '草稿', deleted: '已删除', published: '已发布' }

const canAdd = computed(() => rbacStore.hasPermission('message:system:add'))
const canEdit = computed(() => rbacStore.hasPermission('message:system:edit'))
const canDelete = computed(() => rbacStore.hasPermission('message:system:delete'))
const canViewTargets = computed(() => rbacStore.hasAnyPermission(['message:system:target:view', 'system:rbac:role', 'system:rbac:group', 'system:rbac:user']))

const getTargetOptions = (targetType: string) => {
  if (targetType === 'all') return [{ label: '全部', value: 'all' }]
  if (targetType === 'role') return roleOptions.value.map(item => ({ label: `${item.name}（${item.code}）`, value: String(item.id) }))
  if (targetType === 'group' || targetType === 'department' || targetType === 'position') return groupOptions.value.map(item => ({ label: `${item.name}（${item.code}）`, value: String(item.id) }))
  return userOptions.value.map(item => ({ label: item.username, value: String(item.id) }))
}

const queryList = async () => {
  state.loading = true
  state.error = ''
  try {
    const [startTime, endTime] = state.publishRange
    const rsp = await store.getAdminMessages({ page_num: state.page, page_size: state.pageSize, keyword: state.keyword, type: state.type, status: state.status, start_time: startTime, end_time: endTime })
    const data = (rsp.data.data || {}) as SystemMessageListResponse
    state.list = data.list || data.lists || []
    state.total = Number(data.total || 0)
    state.stale = false
  } catch {
    state.error = '获取系统通知失败，请检查网络后重试'
    state.stale = state.list.length > 0
    notifyError('获取系统通知失败')
  } finally {
    state.loading = false
  }
}

const resetForm = () => {
  editingId.value = ''
  readonlyMode.value = false
  form.id = undefined
  form.type = 'announcement'
  form.title = ''
  form.summary = ''
  form.content = ''
  form.status = 'draft'
  form.publish_time = ''
  form.effective_start_time = ''
  form.effective_end_time = ''
  form.is_pinned = false
  form.target_scopes = [{ target_type: 'all', target_id: 'all' }]
}

const openAdd = () => {
  if (!canAdd.value) return
  resetForm()
  formOpen.value = true
}

const openEdit = (item: SystemMessageRecord, readonly = item.status === 'published' || !canEdit.value) => {
  resetForm()
  editingId.value = item.id
  readonlyMode.value = readonly
  Object.assign(form, item)
  form.id = item.id
  if (!Array.isArray(form.target_scopes) || form.target_scopes.length === 0) form.target_scopes = [{ target_type: 'all', target_id: 'all' }]
  formOpen.value = true
}

const addScope = () => form.target_scopes.push({ target_type: 'user', target_id: '' })
const removeScope = (index: number) => form.target_scopes.splice(index, 1)

const save = async () => {
  if (readonlyMode.value) return
  if (editingId.value && !canEdit.value) return
  if (!editingId.value && !canAdd.value) return
  if (!form.title?.trim()) {
    notifyWarning('请输入通知标题')
    return
  }
  const scopes = form.target_scopes.filter(item => item.target_type && item.target_id)
  if (scopes.length === 0) {
    notifyWarning('请至少选择一个目标范围')
    return
  }
  state.saving = true
  try {
    await store.saveSystemMessage({ ...form, target_scopes: scopes as MessageTargetScopeInput[] })
    notifySuccess(editingId.value ? '系统通知已更新' : '系统通知已创建')
    formOpen.value = false
    await queryList()
  } catch {
    notifyError(editingId.value ? '更新系统通知失败' : '创建系统通知失败')
  } finally {
    state.saving = false
  }
}

const remove = async (item: SystemMessageRecord) => {
  if (!canDelete.value) return
  try {
    await confirmAction(`确认删除通知「${item.title}」吗？删除后将不会再出现在用户消息列表中。`, '删除系统通知')
    await store.removeSystemMessage(item.id)
    notifySuccess('系统通知已删除')
    await queryList()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') notifyError('删除系统通知失败')
  }
}

const batchRemove = async () => {
  if (!canDelete.value) return
  if (selectedRows.value.length === 0) {
    notifyWarning('请先选择要删除的通知')
    return
  }
  try {
    await confirmAction(`确认删除选中的 ${selectedRows.value.length} 条系统通知吗？`, '批量删除系统通知')
    await store.removeSystemMessages(selectedRows.value.map(item => item.id))
    selectedRows.value = []
    notifySuccess('已批量删除系统通知')
    await queryList()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') notifyError('批量删除系统通知失败')
  }
}

const resetFilters = async () => {
  state.keyword = ''
  state.type = ''
  state.status = ''
  state.publishRange = []
  state.page = 1
  await queryList()
}

const loadOptions = async () => {
  if (!canViewTargets.value) return
  const [roles, groups, users] = await Promise.all([rbacApi.getRoles({ page: 1, size: 200 }), rbacApi.getGroups({ page: 1, size: 200 }), rbacApi.getManageUsers({ page: 1, size: 200 })])
  roleOptions.value = roles.data.data?.lists || []
  groupOptions.value = groups.data.data?.lists || []
  userOptions.value = users.data.data?.lists || []
}

const handlePaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.page = page
  state.pageSize = pageSize
  await queryList()
}

const refreshTable = async () => {
  if (tableToolbar.refreshing) return
  tableToolbar.refreshing = true
  try {
    await queryList()
  } finally {
    tableToolbar.refreshing = false
  }
}

const rowActionLabel = (row: SystemMessageRecord) => row.status === 'published' || !canEdit.value ? '查看' : '编辑'
const targetPlaceholder = (targetType: string) => targetType === 'all' ? '全部' : canViewTargets.value ? '选择目标' : '暂无目标范围读取权限'

watch(() => [state.keyword, state.type, state.status, state.publishRange?.[0], state.publishRange?.[1]], () => {
  if (filterTimer) window.clearTimeout(filterTimer)
  filterTimer = window.setTimeout(() => {
    state.page = 1
    queryList()
  }, 300)
})

onMounted(async () => {
  await Promise.all([loadOptions(), queryList()])
})

onUnmounted(() => {
  if (filterTimer) window.clearTimeout(filterTimer)
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never" class="system-message-filter-card" body-class="system-message-filter-card-body">
      <div class="system-message-toolbar">
        <div class="system-message-toolbar-main">
          <div class="system-message-toolbar-copy">
            <div class="system-message-toolbar-title">通知筛选</div>
            <div class="system-message-toolbar-desc">按标题、发布时间、状态和类型快速过滤系统通知列表。</div>
          </div>
        </div>

        <div class="system-message-filter-grid">
          <div class="system-message-filter-row system-message-filter-row--primary">
            <el-input v-model="state.keyword" class="system-message-filter-input" clearable placeholder="按通知标题关键词筛选" @keyup.enter="queryList"><template #prefix><el-icon><Search /></el-icon></template></el-input>
            <el-select v-model="state.status" class="system-message-filter-select" clearable placeholder="状态"><el-option label="已发布" value="published" /><el-option label="草稿" value="draft" /></el-select>
            <el-select v-model="state.type" class="system-message-filter-select" clearable placeholder="类型"><el-option label="公告" value="announcement" /><el-option label="预警" value="warning" /><el-option label="审核" value="audit" /></el-select>
            <el-date-picker v-model="state.publishRange" type="datetimerange" format="YYYY-MM-DD" start-placeholder="发布时间开始" end-placeholder="发布时间结束" value-format="YYYY-MM-DDTHH:mm" class="system-message-filter-date" />
            <div class="system-message-query-actions">
              <el-button type="primary" :icon="Search" @click="queryList">查询</el-button>
              <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
            </div>
          </div>
        </div>
      </div>
    </el-card>
    <el-card shadow="never" body-class="!p-0" :class="tableToolbar.focused ? 'app-table-focused-card' : ''">
      <AppTableToolbar
        title="系统通知"
        :columns="toolbarColumns"
        v-model:visible-columns="tableToolbar.visibleColumns"
        v-model:focused="tableToolbar.focused"
        :refreshing="tableToolbar.refreshing || state.loading"
        @refresh="refreshTable"
      >
        <template #summary>
          <span class="text-xs text-muted-foreground">
            <template v-if="canDelete">已选 {{ selectedRows.length }} 条，</template>共 {{ state.total }} 条
          </span>
        </template>
        <template #default />
      </AppTableToolbar>
      <div class="system-message-list-toolbar px-4 pt-3">
        <div class="system-message-list-summary">管理通知发布、目标范围与状态。</div>
        <div class="system-message-actions">
          <el-button v-if="canAdd" type="primary" :icon="Plus" @click="openAdd">新增通知</el-button>
          <el-button v-if="canDelete" type="danger" plain :disabled="selectedRows.length === 0" @click="batchRemove">批量删除</el-button>
        </div>
      </div>
      <AppTable
        :data="state.list as unknown as Record<string, unknown>[]"
        :columns="visibleColumns"
        :loading="state.loading"
        :selection="canDelete"
        :error="state.error"
        :stale="state.stale"
        empty-text="暂无系统通知"
        @retry="queryList"
        @selection-change="selectedRows = $event as unknown as SystemMessageRecord[]"
      >
        <template #status="{ row }">
          <AppStatusTag :status="String(row.status || '')" :label-map="statusLabelMap" :success-values="['published']" :danger-values="['deleted']" :warning-values="[]" />
        </template>
        <template #actions="{ row }">
          <AppRowActions>
            <AppActionButton type="primary" @click="openEdit(row as unknown as SystemMessageRecord, row.status === 'published' || !canEdit)">{{ rowActionLabel(row as unknown as SystemMessageRecord) }}</AppActionButton>
            <AppActionButton v-if="canDelete" type="danger" @click="remove(row as unknown as SystemMessageRecord)">删除</AppActionButton>
          </AppRowActions>
        </template>
        <template #empty><AppEmptyState description="暂无系统通知" /></template>
      </AppTable>
    </el-card>
    <AppPagination v-model:current-page="state.page" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />
    <AppFormDrawer v-model="formOpen" :title="readonlyMode ? '查看系统通知' : editingId ? '编辑系统通知' : '新增系统通知'" size="760px" :show-footer="false">
      <div class="system-message-drawer-content">
      <div v-if="readonlyMode" class="management-sections">
        <section class="management-section">
          <header><div><span>01</span><h3>通知内容与发布</h3></div><p>通知识别、正文与当前发布状态</p></header>
          <div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div class="sm:col-span-2"><div class="app-form-label">通知标题</div><div class="mt-1 text-sm">{{ form.title || '-' }}</div></div>
            <div><div class="app-form-label">通知类型</div><div class="mt-1 text-sm">{{ form.type === 'warning' ? '预警' : form.type === 'audit' ? '审核' : '公告' }}</div></div>
            <div><div class="app-form-label">状态</div><div class="mt-1 text-sm">{{ statusLabelMap[form.status as keyof typeof statusLabelMap] || form.status || '-' }}</div></div>
            <div class="sm:col-span-2"><div class="app-form-label">通知摘要</div><div class="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">{{ form.summary || '暂无摘要' }}</div></div>
            <div><div class="app-form-label">发布时间</div><div class="mt-1 text-sm">{{ form.publish_time || '-' }}</div></div>
            <div><div class="app-form-label">置顶</div><div class="mt-1 text-sm">{{ form.is_pinned ? '是' : '否' }}</div></div>
            <div class="sm:col-span-2"><div class="app-form-label">通知内容</div><div class="mt-1 whitespace-pre-wrap rounded-lg bg-muted/40 p-3 text-sm leading-6">{{ form.content || '暂无内容' }}</div></div>
          </div>
        </section>
        <section class="management-section">
          <header><div><span>02</span><h3>可见范围</h3></div><p>通知将对以下目标生效</p></header>
          <div class="mt-3 flex flex-wrap gap-2">
            <el-tag v-for="(scope, index) in form.target_scopes" :key="index" effect="plain">
              {{ scope.target_type === 'all' ? '全部用户' : (getTargetOptions(scope.target_type).find(option => option.value === String(scope.target_id))?.label || scope.target_id) }}
            </el-tag>
          </div>
        </section>
      </div>
      <el-form v-else label-position="top" class="management-sections">
        <section class="management-section">
          <header><div><span>01</span><h3>通知内容与发布</h3></div><p>配置通知内容、分类和发布策略</p></header>
          <div class="mt-3">
            <el-form-item label="通知标题" required><el-input v-model="form.title" placeholder="请输入通知标题" /></el-form-item>
            <el-form-item label="通知摘要"><el-input v-model="form.summary" placeholder="用于列表展示的简短摘要" /></el-form-item>
            <div class="system-message-form-grid">
              <el-form-item label="通知类型"><el-select v-model="form.type"><el-option label="公告" value="announcement" /><el-option label="预警" value="warning" /><el-option label="审核" value="audit" /></el-select></el-form-item>
              <el-form-item label="状态"><el-select v-model="form.status" :disabled="Boolean(editingId && form.status === 'published')"><el-option label="草稿" value="draft" /><el-option label="已发布" value="published" /></el-select></el-form-item>
            </div>
            <div class="management-subsection-title">内容与发布设置</div>
            <el-form-item label="发布时间"><el-date-picker v-model="form.publish_time" type="datetime" value-format="YYYY-MM-DDTHH:mm" placeholder="选择发布时间" class="w-full" /></el-form-item>
            <el-form-item label="通知内容" required><el-input v-model="form.content" type="textarea" :rows="6" placeholder="请输入通知正文" /></el-form-item>
            <el-form-item label="置顶"><el-switch v-model="form.is_pinned" /></el-form-item>
          </div>
        </section>
        <section class="management-section">
          <header><div><span>02</span><h3>可见范围</h3></div><p>多个范围叠加生效，保存前确认覆盖对象</p></header>
          <el-form-item label="目标范围" required class="mt-3">
            <div class="w-full space-y-2">
              <el-alert v-if="!canViewTargets" type="warning" :closable="false" show-icon title="当前账号没有目标范围读取权限，仅可发布给全部用户。" />
              <div v-for="(scope, index) in form.target_scopes" :key="index" class="flex flex-col gap-2 sm:flex-row">
                <el-select v-model="scope.target_type" class="w-full sm:!w-[140px]" @change="scope.target_id = scope.target_type === 'all' ? 'all' : ''">
                  <el-option label="全部" value="all" /><el-option label="用户" value="user" :disabled="!canViewTargets" /><el-option label="角色" value="role" :disabled="!canViewTargets" /><el-option label="用户组" value="group" :disabled="!canViewTargets" /><el-option label="部门" value="department" :disabled="!canViewTargets" /><el-option label="岗位" value="position" :disabled="!canViewTargets" />
                </el-select>
                <el-select v-model="scope.target_id" filterable class="min-w-0 flex-1" :placeholder="targetPlaceholder(scope.target_type)" :disabled="scope.target_type !== 'all' && !canViewTargets">
                  <el-option v-for="option in getTargetOptions(scope.target_type)" :key="option.value" :label="option.label" :value="option.value" />
                </el-select>
                <el-button v-if="form.target_scopes.length > 1" type="danger" text @click="removeScope(index)">移除</el-button>
              </div>
              <el-button text type="primary" @click="addScope">添加范围</el-button>
            </div>
          </el-form-item>
        </section>
      </el-form>
      </div>
      <template #footer><el-button @click="formOpen = false">{{ readonlyMode ? '关闭' : '取消' }}</el-button><el-button v-if="!readonlyMode" type="primary" :loading="state.saving" @click="save">保存</el-button></template>
    </AppFormDrawer>
  </div>
</template>

<style scoped>
.system-message-drawer-content,
.management-sections { display: grid; gap: 12px; }
.management-section { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.management-section > header p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.management-section { padding: 0 14px 14px; }
.management-section > header { display: flex; align-items: center; justify-content: space-between; min-height: 38px; margin: 0 -14px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.management-section > header > div { display: flex; align-items: center; gap: 8px; }
.management-section > header span { color: var(--brand-600); font-family: monospace; font-size: 10px; font-weight: 800; }
.management-section > header h3 { margin: 0; font-size: 13px; font-weight: 700; }
.management-section > header p { margin: 0; }
.management-subsection-title { margin: 2px 0 14px; padding-top: 14px; border-top: 1px solid var(--app-overlay-border); font-size: 12px; font-weight: 700; }
@media (max-width: 760px) {
  .management-section > header { align-items: center; flex-direction: row; }
  .management-section > header p { display: none; }
  .system-message-form-grid { grid-template-columns: 1fr; }
}
</style>
