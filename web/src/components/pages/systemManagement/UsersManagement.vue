<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { rbacApi } from '@/api/rbac'
import { getPageSize } from '@/util/pageUtils'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTableToolbar from '@/components/ui/table-toolbar/AppTableToolbar.vue'
import { createTableToolbarState, getVisibleToolbarColumns } from '@/components/ui/table-toolbar/tableToolbar'
import type { TableToolbarColumn } from '@/components/ui/table-toolbar/types'
import { notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'

interface UserItem extends Record<string, unknown> {
  id: number
  username: string
  channel: string
}

interface RoleItem { id: number; code: string; name: string }
interface GroupItem { id: number; code: string; name: string }

const state = reactive({
  list: [] as UserItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false
})

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const formData = reactive({ username: '', password: '' })
const deleteOpen = ref(false)
const deleteTarget = ref<UserItem | null>(null)
const deleteInput = ref('')
const roleAssignOpen = ref(false)
const groupAssignOpen = ref(false)
const selectedUser = ref<UserItem | null>(null)
const roleList = ref<RoleItem[]>([])
const groupList = ref<GroupItem[]>([])
const selectedRoleIds = ref<number[]>([])
const selectedGroupIds = ref<number[]>([])
const assignLoading = ref(false)

const canDelete = computed(() => {
  const name = deleteTarget.value?.username || ''
  return deleteInput.value.trim() === name && name.length > 0
})

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 90, align: 'center' },
  { prop: 'username', label: '用户名', minWidth: 180 },
  { prop: 'channel', label: '渠道', width: 110, align: 'center' },
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
    await queryList()
  } finally {
    tableToolbar.refreshing = false
  }
}

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getManageUsers({ page: state.currPage, size: state.pageSize, text: state.search })
    state.list = rsp.data.data?.lists || []
    state.total = rsp.data.data?.total || 0
  } catch (error) {
    state.list = []
    state.total = 0
    notifyError('获取用户列表失败')
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
  formData.username = ''
  formData.password = ''
  formOpen.value = true
}

const openEdit = (item: UserItem) => {
  editingId.value = item.id
  formData.username = item.username
  formData.password = ''
  formOpen.value = true
}

const submitForm = async () => {
  if (!formData.username.trim()) {
    notifyWarning('请输入用户名')
    return
  }
  if (!editingId.value && formData.password.trim().length < 6) {
    notifyWarning('新增用户时密码至少6位')
    return
  }
  try {
    if (editingId.value) {
      await rbacApi.editManageUser({ id: editingId.value, username: formData.username.trim(), passwd: formData.password.trim() })
      notifySuccess('编辑用户成功')
    } else {
      await rbacApi.addManageUser({ username: formData.username.trim(), passwd: formData.password.trim() })
      notifySuccess('新增用户成功')
    }
    formOpen.value = false
    await queryList()
  } catch (error) {
    notifyError('保存用户失败')
  }
}

const openDelete = (item: UserItem) => {
  deleteTarget.value = item
  deleteInput.value = ''
  deleteOpen.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value || !canDelete.value) return
  try {
    await rbacApi.deleteManageUser({ id: deleteTarget.value.id })
    notifySuccess('删除用户成功')
    deleteOpen.value = false
    await queryList()
  } catch (error) {
    notifyError('删除用户失败')
  }
}

const openAssignRoles = async (user: UserItem) => {
  selectedUser.value = user
  roleAssignOpen.value = true
  assignLoading.value = true
  try {
    const [roleRsp, relationRsp] = await Promise.all([rbacApi.getRoles({ page: 1, size: 200 }), rbacApi.getUserRoleIDs(user.id)])
    roleList.value = roleRsp.data.data?.lists || []
    selectedRoleIds.value = relationRsp.data.data?.role_ids || []
  } finally {
    assignLoading.value = false
  }
}

const openAssignGroups = async (user: UserItem) => {
  selectedUser.value = user
  groupAssignOpen.value = true
  assignLoading.value = true
  try {
    const [groupRsp, relationRsp] = await Promise.all([rbacApi.getGroups({ page: 1, size: 200 }), rbacApi.getUserGroupIDs(user.id)])
    groupList.value = groupRsp.data.data?.lists || []
    selectedGroupIds.value = relationRsp.data.data?.group_ids || []
  } finally {
    assignLoading.value = false
  }
}

const submitUserRoleAssign = async () => {
  if (!selectedUser.value) return
  await rbacApi.assignUserRoles({ user_id: selectedUser.value.id, role_ids: selectedRoleIds.value })
  notifySuccess('用户角色授权成功')
  roleAssignOpen.value = false
}

const submitUserGroupAssign = async () => {
  if (!selectedUser.value) return
  await rbacApi.assignUserGroups({ user_id: selectedUser.value.id, group_ids: selectedGroupIds.value })
  notifySuccess('用户组授权成功')
  groupAssignOpen.value = false
}

const getChannelTagType = (channel: string) => {
  if (channel === 'casdoor') return 'primary'
  if (channel === 'oidc') return 'warning'
  return 'info'
}

onMounted(queryList)
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div class="flex gap-2">
          <el-input v-model="state.search" class="w-[260px]" clearable placeholder="按用户名搜索" @keyup.enter="handleSearch" @clear="handleSearch">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button :icon="Search" @click="handleSearch">查询</el-button>
          <el-button :icon="Refresh" @click="refreshTable">刷新</el-button>
        </div>
        <el-button v-permission="'system:rbac:user'" type="primary" :icon="Plus" @click="openAdd">新增用户</el-button>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0" :class="tableToolbar.focused ? 'app-table-focused-card' : ''">
      <AppTableToolbar
        title="系统用户"
        :columns="toolbarColumns"
        v-model:visible-columns="tableToolbar.visibleColumns"
        v-model:focused="tableToolbar.focused"
        :refreshing="tableToolbar.refreshing || state.loading"
        @refresh="refreshTable"
      >
        <template #summary>
          <span class="text-xs text-muted-foreground">共 {{ state.total }} 条</span>
        </template>
      </AppTableToolbar>
      <AppTable :data="state.list" :columns="visibleColumns" :loading="state.loading" empty-text="可以先新增用户后再做角色或用户组授权">
        <template #channel="{ row }">
          <el-tag :type="getChannelTagType(String(row.channel || 'local'))" effect="light">{{ row.channel === 'casdoor' ? 'Casdoor' : row.channel === 'oidc' ? 'OIDC' : '本地' }}</el-tag>
        </template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: '编辑', kind: 'write', permission: 'system:rbac:user', onClick: () => openEdit(row as UserItem) },
            { key: 'delete', label: '删除', kind: 'write', permission: 'system:rbac:user', danger: true, onClick: () => openDelete(row as UserItem) },
            { key: 'roles', label: '分配角色', kind: 'write', permission: 'system:rbac:role', onClick: () => openAssignRoles(row as UserItem) },
            { key: 'groups', label: '分配用户组', kind: 'write', permission: 'system:rbac:group', onClick: () => openAssignGroups(row as UserItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState description="可以先新增用户后再做角色或用户组授权" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />

    <AppFormDrawer v-model="formOpen" :title="editingId ? '编辑用户' : '新增用户'" size="560px" :show-footer="false">
      <div class="management-drawer-content">
        <el-form label-position="top" class="management-section">
          <header><div><span>01</span><h3>登录身份</h3></div><p>创建本地系统用户，保存后可继续分配角色和用户组</p></header>
          <div class="management-section-body">
            <el-form-item label="用户名" required>
              <el-input v-model="formData.username" placeholder="用户名" />
              <div class="management-field-help">用户名用于登录与授权对象识别。</div>
            </el-form-item>
            <el-form-item :label="editingId ? '密码' : '密码'" :required="!editingId">
              <el-input v-model="formData.password" type="password" show-password :placeholder="editingId ? '留空则不修改密码' : '密码（至少6位）'" />
            </el-form-item>
            <div class="management-note">{{ editingId ? '仅在输入新密码时覆盖现有密码。' : '新用户密码至少 6 位；保存后不会在页面中再次展示。' }}</div>
          </div>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="formOpen = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </AppFormDrawer>

    <el-dialog v-model="deleteOpen" title="确认删除用户" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body>
      <div class="space-y-3">
        <el-alert type="warning" :closable="false" show-icon>请输入用户名 {{ deleteTarget?.username }} 以确认删除</el-alert>
        <el-input v-model="deleteInput" placeholder="请输入用户名" />
      </div>
      <template #footer>
        <el-button @click="deleteOpen = false">取消</el-button>
        <el-button type="danger" :disabled="!canDelete" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>

    <AppFormDrawer v-model="roleAssignOpen" :title="`用户角色授权 - ${selectedUser?.username || ''}`" size="760px" :show-footer="false">
      <div class="management-drawer-content">
        <section class="management-identity"><div class="management-identity-icon">用</div><div><div class="management-identity-title"><h3>{{ selectedUser?.username || '当前用户' }}</h3><span>ID {{ selectedUser?.id || '-' }}</span><strong>已选 {{ selectedRoleIds.length }} / {{ roleList.length }}</strong></div><p>角色权限将覆盖到该用户；与用户组带来的权限共同生效。</p></div></section>
        <section class="management-section"><header><div><span>01</span><h3>角色范围</h3></div><p>勾选需要直接授予的角色</p></header><div class="management-section-body">
        <el-checkbox-group v-if="roleList.length > 0" v-model="selectedRoleIds" v-loading="assignLoading" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <el-checkbox v-for="role in roleList" :key="role.id" :label="role.id" border class="!m-0 !w-full">{{ role.name }}（{{ role.code }}）</el-checkbox>
        </el-checkbox-group>
        <AppEmptyState v-if="!assignLoading && roleList.length === 0" description="当前没有可分配角色" />
        <div class="management-note">保存后将覆盖当前用户的直接角色关系，不会修改用户组授权。</div>
        </div></section>
      </div>
      <template #footer>
        <el-button @click="roleAssignOpen = false">取消</el-button>
        <el-button type="primary" @click="submitUserRoleAssign">保存授权</el-button>
      </template>
    </AppFormDrawer>

    <AppFormDrawer v-model="groupAssignOpen" :title="`用户组授权 - ${selectedUser?.username || ''}`" size="760px" :show-footer="false">
      <div class="management-drawer-content">
        <section class="management-identity"><div class="management-identity-icon">用</div><div><div class="management-identity-title"><h3>{{ selectedUser?.username || '当前用户' }}</h3><span>ID {{ selectedUser?.id || '-' }}</span><strong>已选 {{ selectedGroupIds.length }} / {{ groupList.length }}</strong></div><p>用户将继承所选用户组关联的角色和权限。</p></div></section>
        <section class="management-section"><header><div><span>01</span><h3>用户组范围</h3></div><p>勾选该用户需要加入的用户组</p></header><div class="management-section-body">
        <el-checkbox-group v-if="groupList.length > 0" v-model="selectedGroupIds" v-loading="assignLoading" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <el-checkbox v-for="group in groupList" :key="group.id" :label="group.id" border class="!m-0 !w-full">{{ group.name }}（{{ group.code }}）</el-checkbox>
        </el-checkbox-group>
        <AppEmptyState v-if="!assignLoading && groupList.length === 0" description="当前没有可分配用户组" />
        <div class="management-note">保存后将覆盖当前用户的用户组关系，不会修改直接角色授权。</div>
        </div></section>
      </div>
      <template #footer>
        <el-button @click="groupAssignOpen = false">取消</el-button>
        <el-button type="primary" @click="submitUserGroupAssign">保存授权</el-button>
      </template>
    </AppFormDrawer>
  </div>
</template>

<style scoped>
.management-drawer-content { display: grid; gap: 12px; }
.management-identity, .management-section { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.management-identity { display: flex; align-items: center; gap: 14px; padding: 15px 17px; }
.management-identity-icon { display: inline-flex; flex: none; align-items: center; justify-content: center; width: 38px; height: 38px; border-radius: 9px; background: color-mix(in srgb, var(--brand-500) 10%, transparent); color: var(--brand-700); font-weight: 800; }
.management-identity > div:last-child { min-width: 0; }
.management-identity-title { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.management-identity-title h3 { margin: 0; font-size: 15px; font-weight: 700; }
.management-identity-title span, .management-identity-title strong { padding: 2px 6px; border-radius: 4px; background: var(--admin-surface-muted); color: var(--admin-text-muted); font-size: 10px; font-weight: 700; }
.management-identity p, .management-section header p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.management-section header { display: flex; align-items: center; justify-content: space-between; min-height: 38px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.management-section header > div { display: flex; align-items: center; gap: 8px; }
.management-section header span { color: var(--brand-600); font-family: monospace; font-size: 10px; font-weight: 800; }
.management-section header h3 { margin: 0; font-size: 13px; font-weight: 700; }
.management-section header p { margin: 0; }
.management-section-body { display: grid; gap: 12px; padding: 14px; }
.management-note { padding: 9px 11px; border-radius: 7px; background: color-mix(in srgb, var(--brand-50) 36%, var(--app-overlay-surface)); color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.management-field-help { margin-top: 5px; color: var(--admin-text-muted); font-size: 11px; line-height: 1.5; }
.management-section-body :deep(.el-form-item) { margin-bottom: 0; }
@media (max-width: 760px) { .management-section header { align-items: center; flex-direction: row; } .management-section header p { display: none; } }
</style>
