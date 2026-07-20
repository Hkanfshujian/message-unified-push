<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { rbacApi } from '@/api/rbac'
import { getPageSize } from '@/util/pageUtils'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import { notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'
import { DownOutlined, RightOutlined, ApartmentOutlined } from '@ant-design/icons-vue'
import { buildPermissionTree, collectNodeKeys, collectNodePermissionIds, flattenPermissionTreeRows, type PermissionTreeNode, type PermissionTreeRow } from '@/util/permissionTree'

interface RoleItem {
  id: number
  code: string
  name: string
  description: string
  status: number
  created_on: string
  modified_on: string
}

interface RoleRow extends RoleItem, Record<string, unknown> {}

interface PermissionItem {
  id: number
  code: string
  name: string
  type: string
}

const state = reactive({
  list: [] as RoleRow[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false,
})

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const formData = reactive({
  code: '',
  name: '',
  description: '',
  status: 1
})

const deleteOpen = ref(false)
const deleteTarget = ref<RoleItem | null>(null)
const deleteInput = ref('')

const assignOpen = ref(false)
const assignRole = ref<RoleItem | null>(null)
const permissionList = ref<PermissionItem[]>([])
const originalPermissionIds = ref<number[]>([])
const selectedPermissionIds = ref<number[]>([])
const permissionTree = ref<PermissionTreeNode[]>([])
const expandedNodeKeys = ref<string[]>([])
const activeNodeKey = ref('')
const assignFilterKeyword = ref('')
const showSelectedOnly = ref(false)

const isDeleteMatch = computed(() => {
  const roleName = deleteTarget.value?.name || ''
  return deleteInput.value.trim().toLowerCase() === roleName.trim().toLowerCase() && roleName.length > 0
})

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 90, align: 'center' },
  { prop: 'code', label: '角色编码', minWidth: 180 },
  { prop: 'name', label: '角色名称', minWidth: 160 },
  { prop: 'status', label: '状态', width: 100, align: 'center' },
  { prop: 'description', label: '描述', minWidth: 220 },
  { prop: 'actions', label: '操作', width: 190, align: 'center', fixed: 'right' }
]

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getRoles({
      page: state.currPage,
      size: state.pageSize,
      text: state.search
    })
    state.list = rsp.data.data?.lists || []
    state.total = rsp.data.data?.total || 0
  } catch (error) {
    notifyError('获取角色列表失败')
  } finally {
    state.loading = false
  }
}

const handleSearch = async () => {
  state.currPage = 1
  await queryList()
}

const handlePaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await queryList()
}

const openAdd = () => {
  editingId.value = null
  formData.code = ''
  formData.name = ''
  formData.description = ''
  formData.status = 1
  formOpen.value = true
}

const openEdit = (item: RoleItem) => {
  editingId.value = item.id
  formData.code = item.code
  formData.name = item.name
  formData.description = item.description || ''
  formData.status = item.status
  formOpen.value = true
}

const submitForm = async () => {
  if (!formData.code.trim() || !formData.name.trim()) {
    notifyWarning('请填写角色编码和角色名称')
    return
  }
  const payload = {
    code: formData.code.trim(),
    name: formData.name.trim(),
    description: formData.description.trim(),
    status: Number(formData.status) === 0 ? 0 : 1
  }
  if (editingId.value) {
    await rbacApi.editRole({ id: editingId.value, ...payload })
    notifySuccess('编辑角色成功')
  } else {
    await rbacApi.addRole(payload)
    notifySuccess('新增角色成功')
  }
  formOpen.value = false
  await queryList()
}

const openDelete = (item: RoleItem) => {
  deleteTarget.value = item
  deleteInput.value = ''
  deleteOpen.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  await rbacApi.deleteRole({ id: deleteTarget.value.id })
  deleteOpen.value = false
  notifySuccess('删除角色成功')
  await queryList()
}

const openAssignPermissions = async (item: RoleItem) => {
  assignRole.value = item
  assignOpen.value = true
  assignFilterKeyword.value = ''
  showSelectedOnly.value = false
  activeNodeKey.value = ''
  const [permissionsRsp, selectedRsp] = await Promise.all([
    rbacApi.getPermissions({ page: 1, size: 5000 }),
    rbacApi.getRolePermissionIDs(item.id),
  ])
  permissionList.value = permissionsRsp.data.data?.lists || []
  permissionTree.value = buildPermissionTree(permissionList.value)
  expandedNodeKeys.value = collectNodeKeys(permissionTree.value)
  originalPermissionIds.value = selectedRsp.data.data?.permission_ids || []
  selectedPermissionIds.value = [...originalPermissionIds.value]
  activeNodeKey.value = permissionTree.value[0]?.key || ''
}

const togglePermission = (permissionId: number, checked: boolean) => {
  if (checked) {
    if (!selectedPermissionIds.value.includes(permissionId)) {
      selectedPermissionIds.value.push(permissionId)
    }
  } else {
    selectedPermissionIds.value = selectedPermissionIds.value.filter(id => id !== permissionId)
  }
}

const submitAssignPermissions = async () => {
  if (!assignRole.value) return
  await rbacApi.assignRolePermissions({
    role_id: assignRole.value.id,
    permission_ids: selectedPermissionIds.value
  })
  notifySuccess('角色权限授权成功')
  assignOpen.value = false
}

const isNodeExpanded = (key: string) => expandedNodeKeys.value.includes(key)

const toggleNodeExpand = (key: string) => {
  if (isNodeExpanded(key)) {
    expandedNodeKeys.value = expandedNodeKeys.value.filter(item => item !== key)
  } else {
    expandedNodeKeys.value.push(key)
  }
}

const isNodeFullySelected = (node: PermissionTreeNode) => {
  const ids = collectNodePermissionIds(node)
  if (ids.length === 0) return false
  return ids.every(id => selectedPermissionIds.value.includes(id))
}

const isNodePartiallySelected = (node: PermissionTreeNode) => {
  const ids = collectNodePermissionIds(node)
  if (ids.length === 0) return false
  const selectedCount = ids.filter(id => selectedPermissionIds.value.includes(id)).length
  return selectedCount > 0 && selectedCount < ids.length
}

const toggleNodePermissions = (node: PermissionTreeNode, checked: boolean) => {
  const ids = collectNodePermissionIds(node)
  if (checked) {
    const merged = new Set([...selectedPermissionIds.value, ...ids])
    selectedPermissionIds.value = Array.from(merged)
  } else {
    const idSet = new Set(ids)
    selectedPermissionIds.value = selectedPermissionIds.value.filter(id => !idSet.has(id))
  }
}

const getNodeSelectionState = (node: PermissionTreeNode): 'full' | 'partial' | 'none' => {
  if (isNodeFullySelected(node)) {
    return 'full'
  }
  if (isNodePartiallySelected(node)) {
    return 'partial'
  }
  return 'none'
}

const isNodeChecked = (node: PermissionTreeNode) => {
  return getNodeSelectionState(node) !== 'none'
}

const getNodeToggleHint = (node: PermissionTreeNode) => {
  return isNodeChecked(node) ? '点击将取消该分组下全部权限' : '点击将全选该分组下全部权限'
}

const totalPermissionCount = computed(() => permissionList.value.length)
const selectedPermissionCount = computed(() => selectedPermissionIds.value.length)
const totalNodeCount = computed(() => collectNodeKeys(permissionTree.value).length)
const normalizedFilterKeyword = computed(() => assignFilterKeyword.value.trim().toLowerCase())
const originalPermissionIdSet = computed(() => new Set(originalPermissionIds.value))
const selectedPermissionIdSet = computed(() => new Set(selectedPermissionIds.value))
const addedPermissionCount = computed(() => selectedPermissionIds.value.filter(id => !originalPermissionIdSet.value.has(id)).length)
const removedPermissionCount = computed(() => originalPermissionIds.value.filter(id => !selectedPermissionIdSet.value.has(id)).length)
const changedPermissionCount = computed(() => addedPermissionCount.value + removedPermissionCount.value)
const sensitivePermissionCount = computed(() => permissionList.value.filter(item => selectedPermissionIdSet.value.has(item.id) && isSensitivePermission(item)).length)
const activeNode = computed(() => findNodeByKey(permissionTree.value, activeNodeKey.value) || permissionTree.value[0] || null)
const activeNodePermissionIds = computed(() => activeNode.value ? collectNodePermissionIds(activeNode.value) : [])
const activeNodeSelectedCount = computed(() => activeNodePermissionIds.value.filter(id => selectedPermissionIdSet.value.has(id)).length)
const activeNodePermissionCount = computed(() => activeNodePermissionIds.value.length)
const activeNodeProgress = computed(() => activeNodePermissionCount.value ? Math.round((activeNodeSelectedCount.value / activeNodePermissionCount.value) * 100) : 0)

const expandAllNodes = () => {
  expandedNodeKeys.value = collectNodeKeys(permissionTree.value)
}

const collapseAllNodes = () => {
  expandedNodeKeys.value = []
}

const selectAllPermissions = () => {
  selectedPermissionIds.value = permissionList.value.map(item => item.id)
}

const clearAllPermissions = () => {
  selectedPermissionIds.value = []
}

const resetPermissionChanges = () => {
  selectedPermissionIds.value = [...originalPermissionIds.value]
}

const findNodeByKey = (nodes: PermissionTreeNode[], key: string): PermissionTreeNode | null => {
  for (const node of nodes) {
    if (node.key === key) return node
    const child = findNodeByKey(node.children, key)
    if (child) return child
  }
  return null
}

const getNodeDirectPermissionCount = (node: PermissionTreeNode) => node.permissions.length
const getNodeTotalPermissionCount = (node: PermissionTreeNode) => collectNodePermissionIds(node).length

const getPermissionTypeLabel = (type: string) => type === 'menu' ? '菜单' : type === 'action' ? '操作' : '接口'
const getPermissionTypeClass = (type: string) => type === 'menu' ? 'rbac-type-menu' : type === 'action' ? 'rbac-type-action' : 'rbac-type-api'
const isSensitivePermission = (permission: PermissionItem) => /delete|remove|stop|disable|rbac|user|permission|delete|删除|停用|权限|用户/.test(`${permission.code}:${permission.name}`.toLowerCase())
const getPermissionChangeType = (permission: PermissionItem) => {
  const wasSelected = originalPermissionIdSet.value.has(permission.id)
  const isSelected = selectedPermissionIdSet.value.has(permission.id)
  if (!wasSelected && isSelected) return '新增'
  if (wasSelected && !isSelected) return '移除'
  if (isSelected) return '保留'
  return '未授权'
}

const permissionTreeRows = computed<PermissionTreeRow[]>(() => flattenPermissionTreeRows({
  nodes: permissionTree.value,
  keyword: normalizedFilterKeyword.value,
  expandedKeys: expandedNodeKeys.value,
  selectedIds: selectedPermissionIds.value,
  selectedOnly: showSelectedOnly.value,
  rootKey: activeNodeKey.value
}))

onMounted(async () => {
  await queryList()
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div class="flex gap-2">
          <el-input v-model="state.search" class="w-[280px]" clearable placeholder="按编码或名称搜索" @keyup.enter="handleSearch" @clear="handleSearch">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button :icon="Search" @click="handleSearch">查询</el-button>
          <el-button :icon="Refresh" @click="queryList">刷新</el-button>
        </div>
        <el-button v-permission="'system:rbac:role'" type="primary" :icon="Plus" @click="openAdd">新增角色</el-button>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0">
      <AppTable :data="state.list" :columns="columns" :loading="state.loading" empty-text="请先创建角色并完成授权">
        <template #status="{ row }"><AppStatusTag :status="row.status" /></template>
        <template #description="{ row }">{{ row.description || '-' }}</template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: '编辑', kind: 'write', permission: 'system:rbac:role', onClick: () => openEdit(row as RoleItem) },
            { key: 'delete', label: '删除', kind: 'write', permission: 'system:rbac:role', danger: true, onClick: () => openDelete(row as RoleItem) },
            { key: 'authorize', label: '授权', kind: 'write', permission: 'system:rbac:role', onClick: () => openAssignPermissions(row as RoleItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState description="请先创建角色并完成授权" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />

    <AppFormDrawer v-model="formOpen" :title="editingId ? '编辑角色' : '新增角色'" size="560px" :show-footer="false">
      <div class="role-form-content">
        <el-form label-position="top" class="role-form-section"><header><div><span>01</span><h3>角色信息</h3></div><p>编码用于权限关联，名称用于授权时识别</p></header><div class="role-form-section-body">
          <el-form-item label="角色编码" required><el-input v-model="formData.code" placeholder="角色编码，例如 role_admin" /></el-form-item>
          <el-form-item label="角色名称" required><el-input v-model="formData.name" placeholder="角色名称" /></el-form-item>
          <el-form-item label="角色描述"><el-input v-model="formData.description" placeholder="角色描述" /></el-form-item>
          <div class="role-form-note">{{ editingId ? '保存仅更新角色信息，已配置的权限矩阵保持不变。' : '创建角色后可进入权限授权矩阵配置能力范围。' }}</div>
        </div></el-form>
      </div>
      <template #footer>
        <el-button @click="formOpen = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </AppFormDrawer>

    <el-dialog v-model="deleteOpen" title="确认删除角色" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body>
      <div class="space-y-3">
        <el-alert type="warning" :closable="false" show-icon>请输入角色名称 {{ deleteTarget?.name }} 以确认删除</el-alert>
        <el-input v-model="deleteInput" placeholder="请输入角色名称" />
      </div>
      <template #footer>
        <el-button @click="deleteOpen = false">取消</el-button>
        <el-button type="danger" :disabled="!isDeleteMatch" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>

    <AppFormDrawer v-model="assignOpen" :title="`角色权限授权 - ${assignRole?.name || ''}`" size="min(1180px, 96vw)" :show-footer="false" body-mode="managed">
      <div class="rbac-matrix-workbench">
        <section class="role-identity rbac-role-identity"><div class="role-identity-icon">角</div><div><div class="role-identity-title"><h3>{{ assignRole?.name || '当前角色' }}</h3><span>{{ assignRole?.code || '-' }}</span><strong>已选 {{ selectedPermissionCount }} / {{ totalPermissionCount }}</strong></div><p>权限矩阵保存后将覆盖该角色现有权限；用户与用户组通过角色继承最终能力。</p></div></section>
        <div v-if="permissionList.length > 0" class="rbac-matrix-toolbar">
          <div class="rbac-matrix-summary">
            <span>已选 <strong>{{ selectedPermissionCount }}</strong> / {{ totalPermissionCount }}</span>
            <span>模块 {{ totalNodeCount }}</span>
            <span v-if="changedPermissionCount > 0" class="rbac-summary-change">变更 {{ changedPermissionCount }}</span>
          </div>
          <div class="rbac-matrix-actions">
            <el-button size="small" @click="expandAllNodes">全部展开</el-button>
            <el-button size="small" @click="collapseAllNodes">全部收起</el-button>
            <el-button size="small" @click="selectAllPermissions">全选权限</el-button>
            <el-button size="small" @click="clearAllPermissions">清空权限</el-button>
            <el-button size="small" :disabled="changedPermissionCount === 0" @click="resetPermissionChanges">还原变更</el-button>
            <el-button size="small" :type="showSelectedOnly ? 'primary' : 'default'" @click="showSelectedOnly = !showSelectedOnly">仅看已选</el-button>
          </div>
          <el-input v-model="assignFilterKeyword" class="rbac-matrix-search" placeholder="按模块、权限名称或编码筛选，例如 模板 / 渠道 / 用户管理" clearable />
        </div>

        <AppEmptyState v-if="permissionList.length === 0" description="当前没有可分配权限" :image-size="72" />
        <div v-else class="rbac-matrix-layout">
          <aside class="rbac-resource-panel">
            <div class="rbac-panel-title">资源模块</div>
            <button
              v-for="node in permissionTree"
              :key="node.key"
              type="button"
              class="rbac-resource-item"
              :class="activeNodeKey === node.key ? 'active' : ''"
              @click="activeNodeKey = node.key; expandedNodeKeys = Array.from(new Set([...expandedNodeKeys, node.key]))"
            >
              <span>{{ node.label }}</span>
              <strong>{{ getNodeTotalPermissionCount(node) }}</strong>
            </button>
          </aside>

          <section class="rbac-permission-matrix">
            <div class="rbac-matrix-head">
              <div>
                <div class="rbac-panel-title">权限矩阵</div>
                <div class="rbac-panel-desc">当前模块 {{ activeNodeSelectedCount }} / {{ activeNodePermissionCount }}，覆盖率 {{ activeNodeProgress }}%</div>
              </div>
              <div class="rbac-matrix-progress"><span :style="{ width: `${activeNodeProgress}%` }" /></div>
            </div>
            <div class="rbac-matrix-table">
              <div v-for="row in permissionTreeRows" :key="row.key" :style="{ paddingLeft: `${row.depth * 14}px` }">
                <div v-if="row.type === 'node'" class="rbac-matrix-node-row">
                  <button type="button" class="rbac-node-check" :class="getNodeSelectionState(row.node)" :title="getNodeToggleHint(row.node)" @click="toggleNodePermissions(row.node, !isNodeChecked(row.node))">
                    <span v-if="isNodeChecked(row.node)">✓</span>
                  </button>
                  <button type="button" class="rbac-node-expand" @click="toggleNodeExpand(row.node.key)">
                    <DownOutlined v-if="isNodeExpanded(row.node.key)" class="text-[13px]" />
                    <RightOutlined v-else class="text-[13px]" />
                    <ApartmentOutlined class="text-[14px] text-brand-500" />
                    <span>{{ row.node.label }}</span>
                  </button>
                  <span class="rbac-node-count">{{ getNodeDirectPermissionCount(row.node) }} / {{ getNodeTotalPermissionCount(row.node) }}</span>
                </div>
                <label v-else class="rbac-matrix-permission-row">
                  <input type="checkbox" :checked="selectedPermissionIds.includes(row.permission.id)" @change="(event) => togglePermission(row.permission.id, (event.target as HTMLInputElement).checked)">
                  <span class="rbac-permission-main">
                    <span class="rbac-permission-name">{{ row.permission.name }}</span>
                    <span class="rbac-permission-code">{{ row.permission.code }}</span>
                  </span>
                  <span class="rbac-permission-type" :class="getPermissionTypeClass(row.permission.type)">{{ getPermissionTypeLabel(row.permission.type) }}</span>
                  <span v-if="isSensitivePermission(row.permission)" class="rbac-sensitive-tag">敏感</span>
                  <span class="rbac-change-tag">{{ getPermissionChangeType(row.permission) }}</span>
                </label>
              </div>
              <div v-if="permissionTreeRows.length === 0" class="rbac-empty-row">未匹配到相关权限，请尝试其他关键字</div>
            </div>
          </section>

          <aside class="rbac-impact-panel">
            <div class="rbac-panel-title">影响预览</div>
            <div class="rbac-impact-card"><span>新增权限</span><strong class="text-emerald-600">{{ addedPermissionCount }}</strong></div>
            <div class="rbac-impact-card"><span>移除权限</span><strong class="text-rose-600">{{ removedPermissionCount }}</strong></div>
            <div class="rbac-impact-card"><span>敏感权限</span><strong class="text-amber-600">{{ sensitivePermissionCount }}</strong></div>
            <el-alert v-if="sensitivePermissionCount > 0" type="warning" :closable="false" show-icon title="当前授权包含用户、权限、删除或停用类敏感能力，请保存前确认影响范围。" />
            <el-alert v-if="changedPermissionCount === 0" type="info" :closable="false" show-icon title="当前授权未产生变更。" />
          </aside>
        </div>
      </div>
      <template #footer>
        <el-button @click="assignOpen = false">取消</el-button>
        <el-button :disabled="changedPermissionCount === 0" @click="resetPermissionChanges">还原变更</el-button>
        <el-button type="primary" @click="submitAssignPermissions">保存授权</el-button>
      </template>
    </AppFormDrawer>
  </div>
</template>

<style scoped>
.role-form-content { display: grid; gap: 12px; }
.role-identity, .role-form-section { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.role-identity { display: flex; align-items: center; gap: 14px; padding: 15px 17px; }
.role-identity-icon { display: inline-flex; flex: none; align-items: center; justify-content: center; width: 38px; height: 38px; border-radius: 9px; background: color-mix(in srgb, var(--brand-500) 10%, transparent); color: var(--brand-700); font-weight: 800; }
.role-identity > div:last-child { min-width: 0; }
.role-identity-title { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.role-identity-title h3 { margin: 0; font-size: 15px; font-weight: 700; }
.role-identity-title span, .role-identity-title strong { padding: 2px 6px; border-radius: 4px; background: var(--admin-surface-muted); color: var(--admin-text-muted); font-size: 10px; font-weight: 700; }
.role-identity p, .role-form-section header p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.role-form-section header { display: flex; align-items: center; justify-content: space-between; min-height: 38px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.role-form-section header > div { display: flex; align-items: center; gap: 8px; }
.role-form-section header span { color: var(--brand-600); font-family: monospace; font-size: 10px; font-weight: 800; }
.role-form-section header h3 { margin: 0; font-size: 13px; font-weight: 700; }
.role-form-section header p { margin: 0; }
.role-form-section-body { display: grid; gap: 12px; padding: 14px; }
.role-form-section-body :deep(.el-form-item) { margin-bottom: 0; }
.role-form-note { color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.rbac-role-identity { flex: none; }
.rbac-matrix-workbench {
  min-height: 0;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 12px;
}

.rbac-matrix-toolbar {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) auto;
  gap: 10px 12px;
  padding: 12px;
  border: 1px solid var(--line-weak);
  border-radius: 12px;
  background: var(--surface-muted);
}

.rbac-matrix-summary,
.rbac-matrix-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.rbac-matrix-summary {
  color: var(--admin-text-muted);
  font-size: 13px;
}

.rbac-matrix-summary strong,
.rbac-summary-change {
  color: var(--brand-600);
  font-weight: 700;
}

.rbac-matrix-actions {
  justify-content: flex-end;
}

.rbac-matrix-search {
  grid-column: 1 / -1;
}

.rbac-matrix-layout {
  min-height: 0;
  display: grid;
  flex: 1;
  grid-template-columns: 210px minmax(0, 1fr) 220px;
  gap: 12px;
  overflow: hidden;
}

.rbac-resource-panel,
.rbac-permission-matrix,
.rbac-impact-panel {
  min-width: 0;
  min-height: 0;
  border: 1px solid var(--line-weak);
  border-radius: 12px;
  background: var(--dora-container-bg);
}

.rbac-resource-panel,
.rbac-impact-panel {
  padding: 12px;
}

.rbac-panel-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--foreground);
}

.rbac-panel-desc {
  margin-top: 2px;
  color: var(--admin-text-muted);
  font-size: 12px;
}

.rbac-resource-item {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 8px;
  padding: 9px 10px;
  border: 1px solid transparent;
  border-radius: 10px;
  background: var(--surface-muted);
  color: var(--foreground);
  font-size: 13px;
  text-align: left;
}

.rbac-resource-item.active {
  border-color: color-mix(in srgb, var(--brand-600) 24%, var(--line-weak));
  background: color-mix(in srgb, var(--brand-600) 10%, transparent);
  color: var(--brand-600);
  font-weight: 700;
}

.rbac-matrix-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px;
  border-bottom: 1px solid var(--line-weak);
}

.rbac-matrix-progress {
  width: 120px;
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--dora-layout-bg) 72%, transparent);
}

.rbac-matrix-progress span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--brand-600), #0ea5e9);
}

.rbac-permission-matrix {
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.rbac-matrix-table {
  min-height: 0;
  flex: 1;
  overflow: auto;
  padding: 10px;
  overscroll-behavior: contain;
}

.rbac-matrix-node-row,
.rbac-matrix-permission-row {
  display: grid;
  align-items: center;
  gap: 8px;
  min-height: 40px;
  margin-bottom: 6px;
  border-radius: 10px;
}

.rbac-matrix-node-row {
  grid-template-columns: 22px minmax(0, 1fr) auto;
  padding: 7px 9px;
  background: var(--surface-muted);
}

.rbac-matrix-permission-row {
  grid-template-columns: 18px minmax(0, 1fr) auto auto auto;
  padding: 8px 10px;
  border: 1px solid transparent;
  cursor: pointer;
}

.rbac-matrix-permission-row:hover {
  border-color: color-mix(in srgb, var(--brand-600) 18%, var(--line-weak));
  background: color-mix(in srgb, var(--brand-600) 5%, transparent);
}

.rbac-node-check {
  width: 18px;
  height: 18px;
  border: 1px solid var(--line-weak);
  border-radius: 5px;
  color: transparent;
  background: var(--dora-container-bg);
  font-size: 11px;
  font-weight: 800;
}

.rbac-node-check.full,
.rbac-node-check.partial {
  color: #fff;
  border-color: transparent;
  background: linear-gradient(135deg, var(--brand-600), #0ea5e9);
}

.rbac-node-check.partial span {
  opacity: .72;
}

.rbac-node-expand {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 0;
  background: transparent;
  color: var(--foreground);
  font-size: 13px;
  font-weight: 700;
}

.rbac-node-expand span,
.rbac-permission-name,
.rbac-permission-code {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rbac-node-count,
.rbac-change-tag,
.rbac-sensitive-tag,
.rbac-permission-type {
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
}

.rbac-node-count,
.rbac-change-tag {
  color: var(--admin-text-muted);
  background: color-mix(in srgb, var(--dora-layout-bg) 74%, transparent);
}

.rbac-permission-main {
  min-width: 0;
  display: grid;
  gap: 2px;
}

.rbac-permission-name {
  color: var(--foreground);
  font-size: 13px;
  font-weight: 600;
}

.rbac-permission-code {
  color: var(--admin-text-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
}

.rbac-type-menu {
  color: #0369a1;
  background: rgba(14, 165, 233, .12);
}

.rbac-type-action {
  color: #1d4ed8;
  background: rgba(37, 99, 235, .12);
}

.rbac-type-api {
  color: #047857;
  background: rgba(16, 185, 129, .12);
}

.rbac-sensitive-tag {
  color: #b45309;
  background: rgba(245, 158, 11, .14);
}

.rbac-impact-panel {
  display: grid;
  align-content: start;
  gap: 10px;
}

.rbac-impact-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px;
  border: 1px solid var(--line-weak);
  border-radius: 10px;
  background: var(--surface-muted);
}

.rbac-impact-card span {
  color: var(--admin-text-muted);
  font-size: 12px;
}

.rbac-impact-card strong {
  font-size: 20px;
}

.rbac-empty-row {
  height: 48px;
  display: flex;
  align-items: center;
  color: var(--admin-text-muted);
  font-size: 13px;
}

@media (max-width: 1180px) {
  .rbac-matrix-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(0, 1fr) auto;
  }

  .rbac-resource-panel {
    display: flex;
    gap: 8px;
    overflow-x: auto;
  }

  .rbac-impact-panel {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .rbac-impact-panel .rbac-panel-title,
  .rbac-impact-panel .el-alert {
    grid-column: 1 / -1;
  }

  .rbac-resource-panel .rbac-panel-title {
    display: none;
  }

  .rbac-resource-item {
    min-width: 150px;
    margin-top: 0;
  }
}

@media (max-width: 720px) {
  .rbac-matrix-workbench {
    overflow: hidden;
  }

  .rbac-matrix-toolbar {
    grid-template-columns: 1fr;
  }

  .rbac-matrix-actions {
    justify-content: flex-start;
  }

  .rbac-impact-panel {
    grid-template-columns: 1fr;
  }

  .rbac-matrix-permission-row {
    grid-template-columns: 18px minmax(0, 1fr) auto;
  }

  .rbac-sensitive-tag,
  .rbac-change-tag {
    display: none;
  }
}
</style>
