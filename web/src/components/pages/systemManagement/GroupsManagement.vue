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

interface GroupItem extends Record<string, unknown> { id: number; code: string; name: string; description: string; status: number }
interface RoleItem { id: number; code: string; name: string }
interface UserItem { id: number; username: string }

const state = reactive({
  list: [] as GroupItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false
})

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const formData = reactive({ code: '', name: '', description: '', status: 1 })
const deleteOpen = ref(false)
const deleteTarget = ref<GroupItem | null>(null)
const deleteInput = ref('')
const roleAssignOpen = ref(false)
const memberAssignOpen = ref(false)
const assignGroup = ref<GroupItem | null>(null)
const roleList = ref<RoleItem[]>([])
const userList = ref<UserItem[]>([])
const selectedRoleIds = ref<number[]>([])
const selectedUserIds = ref<number[]>([])
const assignLoading = ref(false)

const isDeleteMatch = computed(() => {
  const name = deleteTarget.value?.name || ''
  return deleteInput.value.trim().toLowerCase() === name.trim().toLowerCase() && name.length > 0
})

const columns: AppTableColumn[] = [
  { prop: 'id', label: 'ID', width: 90, align: 'center' },
  { prop: 'code', label: '用户组编码', minWidth: 180 },
  { prop: 'name', label: '用户组名称', minWidth: 160 },
  { prop: 'status', label: '状态', width: 100, align: 'center' },
  { prop: 'description', label: '描述', minWidth: 220 },
  { prop: 'actions', label: '操作', width: 220, align: 'center', fixed: 'right' }
]

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getGroups({ page: state.currPage, size: state.pageSize, text: state.search })
    state.list = rsp.data.data?.lists || []
    state.total = rsp.data.data?.total || 0
  } catch (error) {
    state.list = []
    state.total = 0
    notifyError('获取用户组列表失败')
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

const openEdit = (item: GroupItem) => {
  editingId.value = item.id
  formData.code = item.code
  formData.name = item.name
  formData.description = item.description || ''
  formData.status = item.status
  formOpen.value = true
}

const submitForm = async () => {
  if (!formData.code.trim() || !formData.name.trim()) {
    notifyWarning('请填写用户组编码和用户组名称')
    return
  }
  const payload = { code: formData.code.trim(), name: formData.name.trim(), description: formData.description.trim(), status: Number(formData.status) === 0 ? 0 : 1 }
  try {
    if (editingId.value) {
      await rbacApi.editGroup({ id: editingId.value, ...payload })
      notifySuccess('编辑用户组成功')
    } else {
      await rbacApi.addGroup(payload)
      notifySuccess('新增用户组成功')
    }
    formOpen.value = false
    await queryList()
  } catch (error) {
    notifyError('保存用户组失败')
  }
}

const openDelete = (item: GroupItem) => {
  deleteTarget.value = item
  deleteInput.value = ''
  deleteOpen.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  try {
    await rbacApi.deleteGroup({ id: deleteTarget.value.id })
    notifySuccess('删除用户组成功')
    deleteOpen.value = false
    await queryList()
  } catch (error) {
    notifyError('删除用户组失败')
  }
}

const openAssignRoles = async (item: GroupItem) => {
  assignGroup.value = item
  roleAssignOpen.value = true
  assignLoading.value = true
  try {
    const [roleRsp, selectedRsp] = await Promise.all([rbacApi.getRoles({ page: 1, size: 200 }), rbacApi.getGroupRoleIDs(item.id)])
    roleList.value = roleRsp.data.data?.lists || []
    selectedRoleIds.value = selectedRsp.data.data?.role_ids || []
  } finally {
    assignLoading.value = false
  }
}

const openAssignMembers = async (item: GroupItem) => {
  assignGroup.value = item
  memberAssignOpen.value = true
  assignLoading.value = true
  try {
    const [userRsp, selectedRsp] = await Promise.all([rbacApi.getUsers({ page: 1, size: 500 }), rbacApi.getGroupMemberIDs(item.id)])
    userList.value = userRsp.data.data?.lists || []
    selectedUserIds.value = selectedRsp.data.data?.user_ids || []
  } finally {
    assignLoading.value = false
  }
}

const submitAssignRoles = async () => {
  if (!assignGroup.value) return
  await rbacApi.assignGroupRoles({ group_id: assignGroup.value.id, role_ids: selectedRoleIds.value })
  notifySuccess('用户组角色授权成功')
  roleAssignOpen.value = false
}

const submitAssignMembers = async () => {
  if (!assignGroup.value) return
  await rbacApi.assignGroupMembers({ group_id: assignGroup.value.id, user_ids: selectedUserIds.value })
  notifySuccess('用户组成员授权成功')
  memberAssignOpen.value = false
}

onMounted(queryList)
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
        <el-button v-permission="'system:rbac:group'" type="primary" :icon="Plus" @click="openAdd">新增用户组</el-button>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0">
      <AppTable :data="state.list" :columns="columns" :loading="state.loading" empty-text="请先创建用户组并绑定角色或成员">
        <template #status="{ row }"><AppStatusTag :status="row.status" /></template>
        <template #description="{ row }">{{ row.description || '-' }}</template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: '编辑', kind: 'write', permission: 'system:rbac:group', onClick: () => openEdit(row as GroupItem) },
            { key: 'delete', label: '删除', kind: 'write', permission: 'system:rbac:group', danger: true, onClick: () => openDelete(row as GroupItem) },
            { key: 'roles', label: '分配角色', kind: 'write', permission: 'system:rbac:group', onClick: () => openAssignRoles(row as GroupItem) },
            { key: 'members', label: '分配成员', kind: 'write', permission: 'system:rbac:group', onClick: () => openAssignMembers(row as GroupItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState description="请先创建用户组并绑定角色或成员" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />

    <AppFormDrawer v-model="formOpen" :title="editingId ? '编辑用户组' : '新增用户组'" size="560px" :show-footer="false">
      <div class="management-drawer-content">
        <el-form label-position="top" class="management-section"><header><div><span>01</span><h3>用户组信息</h3></div><p>编码用于系统识别，名称用于界面展示</p></header><div class="management-section-body">
          <el-form-item label="用户组编码" required><el-input v-model="formData.code" placeholder="用户组编码，例如 group_ops" /></el-form-item>
          <el-form-item label="用户组名称" required><el-input v-model="formData.name" placeholder="用户组名称" /></el-form-item>
          <el-form-item label="用户组描述"><el-input v-model="formData.description" placeholder="用户组描述" /></el-form-item>
          <div class="management-note">成员将继承该用户组关联角色的权限，保存基本信息不会改变授权关系。</div>
        </div></el-form>
      </div>
      <template #footer>
        <el-button @click="formOpen = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </AppFormDrawer>

    <el-dialog v-model="deleteOpen" title="确认删除用户组" width="min(420px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body>
      <div class="space-y-3">
        <el-alert type="warning" :closable="false" show-icon>请输入用户组名称 {{ deleteTarget?.name }} 以确认删除</el-alert>
        <el-input v-model="deleteInput" placeholder="请输入用户组名称" />
      </div>
      <template #footer>
        <el-button @click="deleteOpen = false">取消</el-button>
        <el-button type="danger" :disabled="!isDeleteMatch" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>

    <AppFormDrawer v-model="roleAssignOpen" :title="`用户组角色授权 - ${assignGroup?.name || ''}`" size="760px" :show-footer="false">
      <div class="management-drawer-content">
        <section class="management-identity"><div class="management-identity-icon">组</div><div><div class="management-identity-title"><h3>{{ assignGroup?.name || '当前用户组' }}</h3><span>{{ assignGroup?.code || '-' }}</span><strong>已选 {{ selectedRoleIds.length }} / {{ roleList.length }}</strong></div><p>组内全部成员将继承所选角色的权限。</p></div></section>
        <section class="management-section"><header><div><span>01</span><h3>角色范围</h3></div><p>勾选该用户组需要关联的角色</p></header><div class="management-section-body">
        <el-checkbox-group v-if="roleList.length > 0" v-model="selectedRoleIds" v-loading="assignLoading" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <el-checkbox v-for="role in roleList" :key="role.id" :label="role.id" border class="!m-0 !w-full">{{ role.name }}（{{ role.code }}）</el-checkbox>
        </el-checkbox-group>
        <AppEmptyState v-if="!assignLoading && roleList.length === 0" description="当前没有可分配角色" />
        <div class="management-note">保存后覆盖当前用户组的角色关系，成员关系保持不变。</div>
        </div></section>
      </div>
      <template #footer>
        <el-button @click="roleAssignOpen = false">取消</el-button>
        <el-button type="primary" @click="submitAssignRoles">保存授权</el-button>
      </template>
    </AppFormDrawer>

    <AppFormDrawer v-model="memberAssignOpen" :title="`用户组成员授权 - ${assignGroup?.name || ''}`" size="760px" :show-footer="false">
      <div class="management-drawer-content">
        <section class="management-identity"><div class="management-identity-icon">组</div><div><div class="management-identity-title"><h3>{{ assignGroup?.name || '当前用户组' }}</h3><span>{{ assignGroup?.code || '-' }}</span><strong>已选 {{ selectedUserIds.length }} / {{ userList.length }}</strong></div><p>所选用户将加入该组并继承组角色权限。</p></div></section>
        <section class="management-section"><header><div><span>01</span><h3>成员范围</h3></div><p>勾选需要纳入该用户组的用户</p></header><div class="management-section-body">
        <el-checkbox-group v-if="userList.length > 0" v-model="selectedUserIds" v-loading="assignLoading" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <el-checkbox v-for="user in userList" :key="user.id" :label="user.id" border class="!m-0 !w-full">{{ user.username }}</el-checkbox>
        </el-checkbox-group>
        <AppEmptyState v-if="!assignLoading && userList.length === 0" description="当前没有可分配用户" />
        <div class="management-note">保存后覆盖当前用户组的成员关系，角色关系保持不变。</div>
        </div></section>
      </div>
      <template #footer>
        <el-button @click="memberAssignOpen = false">取消</el-button>
        <el-button type="primary" @click="submitAssignMembers">保存授权</el-button>
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
.management-section-body :deep(.el-form-item) { margin-bottom: 0; }
@media (max-width: 760px) { .management-section header { align-items: center; flex-direction: row; } .management-section header p { display: none; } }
</style>
