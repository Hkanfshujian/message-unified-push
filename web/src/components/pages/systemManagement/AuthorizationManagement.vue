<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { rbacApi } from '@/api/rbac'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import { notifySuccess } from '@/util/uiFeedback'

interface UserItem extends Record<string, unknown> { id: number; username: string }
interface RoleItem { id: number; code: string; name: string }
interface GroupItem { id: number; code: string; name: string }

const state = reactive({ users: [] as UserItem[], loading: false })
const roleAssignOpen = ref(false)
const groupAssignOpen = ref(false)
const selectedUser = ref<UserItem | null>(null)
const roleList = ref<RoleItem[]>([])
const groupList = ref<GroupItem[]>([])
const selectedRoleIds = ref<number[]>([])
const selectedGroupIds = ref<number[]>([])
const assignLoading = ref(false)

const columns: AppTableColumn[] = [
  { prop: 'id', label: '用户ID', width: 100, align: 'center' },
  { prop: 'username', label: '用户名', minWidth: 220 },
  { prop: 'actions', label: '操作', width: 150, align: 'center', fixed: 'right' }
]

const queryUsers = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getUsers({ page: 1, size: 500 })
    state.users = rsp.data.data?.lists || []
  } finally {
    state.loading = false
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

onMounted(queryUsers)
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="flex justify-end">
        <el-button :icon="Refresh" @click="queryUsers">刷新</el-button>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0">
      <AppTable :data="state.users" :columns="columns" :loading="state.loading" empty-text="请先创建用户后再分配角色或用户组">
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'roles', label: '分配角色', kind: 'write', permission: 'system:rbac:role', onClick: () => openAssignRoles(row as UserItem) },
            { key: 'groups', label: '分配用户组', kind: 'write', permission: 'system:rbac:group', onClick: () => openAssignGroups(row as UserItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState description="请先创建用户后再分配角色或用户组" />
        </template>
      </AppTable>
    </el-card>

    <AppFormDrawer v-model="roleAssignOpen" :title="`用户角色授权 - ${selectedUser?.username || ''}`" size="760px" :show-footer="false">
      <div class="space-y-4">
        <div class="assignment-toolbar">
          <div class="assignment-object">
            <span class="assignment-object-type">用户</span>
            <strong>{{ selectedUser?.username || '当前用户' }}</strong>
            <span class="assignment-object-id">ID {{ selectedUser?.id || '-' }}</span>
          </div>
          <span class="assignment-summary">已选 <strong>{{ selectedRoleIds.length }}</strong> / {{ roleList.length }}</span>
        </div>
        <el-checkbox-group v-if="roleList.length > 0" v-model="selectedRoleIds" v-loading="assignLoading" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <el-checkbox v-for="role in roleList" :key="role.id" :label="role.id" border class="!m-0 !w-full">{{ role.name }}（{{ role.code }}）</el-checkbox>
        </el-checkbox-group>
        <AppEmptyState v-if="!assignLoading && roleList.length === 0" description="当前没有可分配角色" />
      </div>
      <template #footer>
        <el-button @click="roleAssignOpen = false">取消</el-button>
        <el-button type="primary" @click="submitUserRoleAssign">保存授权</el-button>
      </template>
    </AppFormDrawer>

    <AppFormDrawer v-model="groupAssignOpen" :title="`用户组授权 - ${selectedUser?.username || ''}`" size="760px" :show-footer="false">
      <div class="space-y-4">
        <div class="assignment-toolbar">
          <div class="assignment-object">
            <span class="assignment-object-type">用户</span>
            <strong>{{ selectedUser?.username || '当前用户' }}</strong>
            <span class="assignment-object-id">ID {{ selectedUser?.id || '-' }}</span>
          </div>
          <span class="assignment-summary">已选 <strong>{{ selectedGroupIds.length }}</strong> / {{ groupList.length }}</span>
        </div>
        <el-checkbox-group v-if="groupList.length > 0" v-model="selectedGroupIds" v-loading="assignLoading" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <el-checkbox v-for="group in groupList" :key="group.id" :label="group.id" border class="!m-0 !w-full">{{ group.name }}（{{ group.code }}）</el-checkbox>
        </el-checkbox-group>
        <AppEmptyState v-if="!assignLoading && groupList.length === 0" description="当前没有可分配用户组" />
      </div>
      <template #footer>
        <el-button @click="groupAssignOpen = false">取消</el-button>
        <el-button type="primary" @click="submitUserGroupAssign">保存授权</el-button>
      </template>
    </AppFormDrawer>
  </div>
</template>

<style scoped>
.assignment-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 40px;
  padding: 7px 10px;
  border: 1px solid var(--line-weak);
  border-radius: 10px;
  background: var(--surface-muted);
  font-size: 13px;
}

.assignment-object {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 7px;
}

.assignment-object strong {
  overflow: hidden;
  color: var(--foreground);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.assignment-object-type,
.assignment-object-id {
  flex: none;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--dora-container-bg);
  color: var(--admin-text-muted);
  font-size: 11px;
  font-weight: 700;
}

.assignment-summary {
  flex: none;
  color: var(--admin-text-muted);
}

.assignment-summary strong {
  color: var(--brand-600);
}

@media (max-width: 480px) {
  .assignment-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
