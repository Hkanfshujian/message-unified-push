<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { rbacApi } from '@/api/rbac'
import { getPageSize } from '@/util/pageUtils'
import { useRbacAuthzStore } from '@/store/rbac_authz'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import Pagination from '@/components/ui/Pagination.vue'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'

interface UserItem {
  id: number
  username: string
  channel: string
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
  list: [] as UserItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  loading: false
})

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const formData = reactive({
  username: '',
  password: ''
})

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

const totalPages = computed(() => Math.ceil(state.total / state.pageSize))
const rbacAuthzStore = useRbacAuthzStore()
const canAssignRoles = computed(() => rbacAuthzStore.hasAnyPermission(['system:rbac:role']))
const canAssignGroups = computed(() => rbacAuthzStore.hasAnyPermission(['system:rbac:group']))
const canDelete = computed(() => {
  const name = deleteTarget.value?.username || ''
  return deleteInput.value.trim() === name && name.length > 0
})

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getManageUsers({
      page: state.currPage,
      size: state.pageSize,
      text: state.search
    })
    state.list = rsp.data.data?.lists || []
    state.total = rsp.data.data?.total || 0
  } finally {
    state.loading = false
  }
}

const changePage = async (page: number) => {
  if (page < 1 || page > totalPages.value) return
  state.currPage = page
  await queryList()
}

const handlePageSizeChange = async (size: number) => {
  if (size <= 0) return
  state.pageSize = size
  state.currPage = 1
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
    toast.error('请输入用户名')
    return
  }
  if (!editingId.value && formData.password.trim().length < 6) {
    toast.error('新增用户时密码至少6位')
    return
  }
  if (editingId.value) {
    await rbacApi.editManageUser({
      id: editingId.value,
      username: formData.username.trim(),
      passwd: formData.password.trim()
    })
    toast.success('编辑用户成功')
  } else {
    await rbacApi.addManageUser({
      username: formData.username.trim(),
      passwd: formData.password.trim()
    })
    toast.success('新增用户成功')
  }
  formOpen.value = false
  await queryList()
}

const openDelete = (item: UserItem) => {
  deleteTarget.value = item
  deleteInput.value = ''
  deleteOpen.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value || !canDelete.value) return
  await rbacApi.deleteManageUser({ id: deleteTarget.value.id })
  toast.success('删除用户成功')
  deleteOpen.value = false
  await queryList()
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
  await queryList()
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group">
        <Input v-model="state.search" placeholder="按用户名搜索" class="search-input" @keyup.enter="queryList" />
      </div>
      <Button @click="openAdd">新增用户</Button>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>用户名</TableHead>
          <TableHead class="w-24">渠道</TableHead>
          <TableHead class="text-center">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-if="!state.loading && state.list.length === 0">
          <TableCell colspan="4" class="empty-state">
            <EmptyTableState title="暂无用户" description="可以先新增用户后再做角色或用户组授权" />
          </TableCell>
        </TableRow>
        <TableRow v-for="item in state.list" :key="item.id">
          <TableCell>{{ item.id }}</TableCell>
          <TableCell>{{ item.username }}</TableCell>
          <TableCell>
            <span v-if="item.channel === 'casdoor'" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
              Casdoor
            </span>
            <span v-else-if="item.channel === 'oidc'" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200">
              OIDC
            </span>
            <span v-else class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200">
              本地
            </span>
          </TableCell>
          <TableCell class="text-center space-x-2">
            <Button v-if="canAssignRoles" size="sm" variant="outline" @click="openAssignRoles(item)">分配角色</Button>
            <Button v-if="canAssignGroups" size="sm" variant="outline" @click="openAssignGroups(item)">分配用户组</Button>
            <Button size="sm" variant="outline" @click="openEdit(item)">编辑</Button>
            <Button size="sm" variant="destructive" @click="openDelete(item)">删除</Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    </div>

    <div class="pagination">
      <Pagination
        :total="state.total"
        :current-page="state.currPage"
        :page-size="state.pageSize"
        @page-change="changePage"
        @page-size-change="handlePageSizeChange"
      />
    </div>

    <Dialog v-model:open="formOpen">
      <DialogContent class="w-[560px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>{{ editingId ? '编辑用户' : '新增用户' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <Input v-model="formData.username" placeholder="用户名" />
          <Input v-model="formData.password" type="password" :placeholder="editingId ? '留空则不修改密码' : '密码（至少6位）'" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="formOpen=false">取消</Button>
          <Button @click="submitForm">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteOpen">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除用户</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">
            请输入用户名
            <span class="text-red-500 mx-1">{{ deleteTarget?.username }}</span>
            以确认删除
          </div>
          <Input v-model="deleteInput" placeholder="请输入用户名" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="deleteOpen=false">取消</Button>
          <Button variant="destructive" :disabled="!canDelete" @click="confirmDelete">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

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
