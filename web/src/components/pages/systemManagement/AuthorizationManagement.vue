<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { rbacApi } from '@/api/rbac'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'

interface UserItem {
  id: number
  username: string
}

interface RoleItem {
  id: number
  code: string
  name: string
}

interface GroupItem {
  id: number
  code: string
  name: string
}

const state = reactive({
  users: [] as UserItem[],
  loading: false
})

const roleAssignOpen = ref(false)
const groupAssignOpen = ref(false)
const selectedUser = ref<UserItem | null>(null)
const roleList = ref<RoleItem[]>([])
const groupList = ref<GroupItem[]>([])
const selectedRoleIds = ref<number[]>([])
const selectedGroupIds = ref<number[]>([])

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
  const [roleRsp, relationRsp] = await Promise.all([
    rbacApi.getRoles({ page: 1, size: 200 }),
    rbacApi.getUserRoleIDs(user.id)
  ])
  roleList.value = roleRsp.data.data?.lists || []
  selectedRoleIds.value = relationRsp.data.data?.role_ids || []
}

const openAssignGroups = async (user: UserItem) => {
  selectedUser.value = user
  groupAssignOpen.value = true
  const [groupRsp, relationRsp] = await Promise.all([
    rbacApi.getGroups({ page: 1, size: 200 }),
    rbacApi.getUserGroupIDs(user.id)
  ])
  groupList.value = groupRsp.data.data?.lists || []
  selectedGroupIds.value = relationRsp.data.data?.group_ids || []
}

const toggleRole = (id: number, checked: boolean) => {
  if (checked) {
    if (!selectedRoleIds.value.includes(id)) selectedRoleIds.value.push(id)
  } else {
    selectedRoleIds.value = selectedRoleIds.value.filter(item => item !== id)
  }
}

const toggleGroup = (id: number, checked: boolean) => {
  if (checked) {
    if (!selectedGroupIds.value.includes(id)) selectedGroupIds.value.push(id)
  } else {
    selectedGroupIds.value = selectedGroupIds.value.filter(item => item !== id)
  }
}

const submitUserRoleAssign = async () => {
  if (!selectedUser.value) return
  await rbacApi.assignUserRoles({
    user_id: selectedUser.value.id,
    role_ids: selectedRoleIds.value
  })
  toast.success('用户角色授权成功')
  roleAssignOpen.value = false
}

const submitUserGroupAssign = async () => {
  if (!selectedUser.value) return
  await rbacApi.assignUserGroups({
    user_id: selectedUser.value.id,
    group_ids: selectedGroupIds.value
  })
  toast.success('用户组授权成功')
  groupAssignOpen.value = false
}

onMounted(async () => {
  await queryUsers()
})
</script>

<template>
  <div class="space-y-2">
    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="w-20">用户ID</TableHead>
            <TableHead>用户名</TableHead>
            <TableHead class="text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="!state.loading && state.users.length === 0">
            <TableCell colspan="3" class="empty-state">
              <EmptyTableState title="暂无用户" description="请先创建用户后再分配角色或用户组" />
            </TableCell>
          </TableRow>
          <TableRow v-for="user in state.users" :key="user.id">
            <TableCell>{{ user.id }}</TableCell>
            <TableCell>{{ user.username }}</TableCell>
            <TableCell class="text-center space-x-2">
              <Button size="sm" variant="outline" @click="openAssignRoles(user)">分配角色</Button>
              <Button size="sm" variant="outline" @click="openAssignGroups(user)">分配用户组</Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <Dialog v-model:open="roleAssignOpen">
      <DialogContent class="w-[760px] max-w-[92vw]">
        <DialogHeader>
          <DialogTitle>用户角色授权 - {{ selectedUser?.username }}</DialogTitle>
        </DialogHeader>
        <div class="max-h-[420px] overflow-y-auto border rounded p-3">
          <div v-if="roleList.length === 0" class="text-sm text-muted-foreground">暂无可分配角色</div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <label v-for="role in roleList" :key="role.id" class="flex items-center gap-2 text-sm border rounded px-2 py-1">
              <input
                type="checkbox"
                :checked="selectedRoleIds.includes(role.id)"
                @change="(event) => toggleRole(role.id, (event.target as HTMLInputElement).checked)"
              >
              <span>{{ role.name }}（{{ role.code }}）</span>
            </label>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="roleAssignOpen = false">取消</Button>
          <Button @click="submitUserRoleAssign">保存授权</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="groupAssignOpen">
      <DialogContent class="w-[760px] max-w-[92vw]">
        <DialogHeader>
          <DialogTitle>用户组授权 - {{ selectedUser?.username }}</DialogTitle>
        </DialogHeader>
        <div class="max-h-[420px] overflow-y-auto border rounded p-3">
          <div v-if="groupList.length === 0" class="text-sm text-muted-foreground">暂无可分配用户组</div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <label v-for="group in groupList" :key="group.id" class="flex items-center gap-2 text-sm border rounded px-2 py-1">
              <input
                type="checkbox"
                :checked="selectedGroupIds.includes(group.id)"
                @change="(event) => toggleGroup(group.id, (event.target as HTMLInputElement).checked)"
              >
              <span>{{ group.name }}（{{ group.code }}）</span>
            </label>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="groupAssignOpen = false">取消</Button>
          <Button @click="submitUserGroupAssign">保存授权</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
