<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { rbacApi } from '@/api/rbac'
import { getPageSize } from '@/util/pageUtils'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import Pagination from '@/components/ui/Pagination.vue'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'

interface GroupItem {
  id: number
  code: string
  name: string
  description: string
  status: number
}

interface RoleItem {
  id: number
  code: string
  name: string
}

interface UserItem {
  id: number
  username: string
}

const state = reactive({
  list: [] as GroupItem[],
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
const deleteTarget = ref<GroupItem | null>(null)
const deleteInput = ref('')

const roleAssignOpen = ref(false)
const memberAssignOpen = ref(false)
const assignGroup = ref<GroupItem | null>(null)
const roleList = ref<RoleItem[]>([])
const userList = ref<UserItem[]>([])
const selectedRoleIds = ref<number[]>([])
const selectedUserIds = ref<number[]>([])

const totalPages = computed(() => Math.ceil(state.total / state.pageSize))
const isDeleteMatch = computed(() => {
  const name = deleteTarget.value?.name || ''
  return deleteInput.value.trim().toLowerCase() === name.trim().toLowerCase() && name.length > 0
})

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getGroups({
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
    toast.error('请填写用户组编码和用户组名称')
    return
  }
  const payload = {
    code: formData.code.trim(),
    name: formData.name.trim(),
    description: formData.description.trim(),
    status: Number(formData.status) === 0 ? 0 : 1
  }
  if (editingId.value) {
    await rbacApi.editGroup({ id: editingId.value, ...payload })
    toast.success('编辑用户组成功')
  } else {
    await rbacApi.addGroup(payload)
    toast.success('新增用户组成功')
  }
  formOpen.value = false
  await queryList()
}

const openDelete = (item: GroupItem) => {
  deleteTarget.value = item
  deleteInput.value = ''
  deleteOpen.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  await rbacApi.deleteGroup({ id: deleteTarget.value.id })
  toast.success('删除用户组成功')
  deleteOpen.value = false
  await queryList()
}

const openAssignRoles = async (item: GroupItem) => {
  assignGroup.value = item
  roleAssignOpen.value = true
  const [roleRsp, selectedRsp] = await Promise.all([
    rbacApi.getRoles({ page: 1, size: 200 }),
    rbacApi.getGroupRoleIDs(item.id)
  ])
  roleList.value = roleRsp.data.data?.lists || []
  selectedRoleIds.value = selectedRsp.data.data?.role_ids || []
}

const openAssignMembers = async (item: GroupItem) => {
  assignGroup.value = item
  memberAssignOpen.value = true
  const [userRsp, selectedRsp] = await Promise.all([
    rbacApi.getUsers({ page: 1, size: 500 }),
    rbacApi.getGroupMemberIDs(item.id)
  ])
  userList.value = userRsp.data.data?.lists || []
  selectedUserIds.value = selectedRsp.data.data?.user_ids || []
}

const toggleRole = (roleId: number, checked: boolean) => {
  if (checked) {
    if (!selectedRoleIds.value.includes(roleId)) {
      selectedRoleIds.value.push(roleId)
    }
  } else {
    selectedRoleIds.value = selectedRoleIds.value.filter(id => id !== roleId)
  }
}

const toggleUser = (userId: number, checked: boolean) => {
  if (checked) {
    if (!selectedUserIds.value.includes(userId)) {
      selectedUserIds.value.push(userId)
    }
  } else {
    selectedUserIds.value = selectedUserIds.value.filter(id => id !== userId)
  }
}

const submitAssignRoles = async () => {
  if (!assignGroup.value) return
  await rbacApi.assignGroupRoles({
    group_id: assignGroup.value.id,
    role_ids: selectedRoleIds.value
  })
  toast.success('用户组角色授权成功')
  roleAssignOpen.value = false
}

const submitAssignMembers = async () => {
  if (!assignGroup.value) return
  await rbacApi.assignGroupMembers({
    group_id: assignGroup.value.id,
    user_ids: selectedUserIds.value
  })
  toast.success('用户组成员授权成功')
  memberAssignOpen.value = false
}

onMounted(async () => {
  await queryList()
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group">
        <Input v-model="state.search" placeholder="按编码或名称搜索" class="search-input" @keyup.enter="queryList" />
      </div>
      <Button @click="openAdd">新增用户组</Button>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>用户组编码</TableHead>
          <TableHead>用户组名称</TableHead>
          <TableHead>状态</TableHead>
          <TableHead>描述</TableHead>
          <TableHead class="text-center">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-if="!state.loading && state.list.length === 0">
          <TableCell colspan="6" class="empty-state">
            <EmptyTableState title="暂无用户组" description="请先创建用户组并绑定角色或成员" />
          </TableCell>
        </TableRow>
        <TableRow v-for="item in state.list" :key="item.id">
          <TableCell>{{ item.id }}</TableCell>
          <TableCell>{{ item.code }}</TableCell>
          <TableCell>{{ item.name }}</TableCell>
          <TableCell>{{ item.status === 1 ? '启用' : '禁用' }}</TableCell>
          <TableCell>{{ item.description || '-' }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="openAssignRoles(item)">分配角色</Button>
            <Button size="sm" variant="outline" @click="openAssignMembers(item)">分配成员</Button>
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
          <DialogTitle>{{ editingId ? '编辑用户组' : '新增用户组' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <Input v-model="formData.code" placeholder="用户组编码，例如 group_ops" />
          <Input v-model="formData.name" placeholder="用户组名称" />
          <Input v-model="formData.description" placeholder="用户组描述" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="formOpen = false">取消</Button>
          <Button @click="submitForm">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="deleteOpen">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除用户组</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">
            请输入用户组名称
            <span class="text-red-500 mx-1">{{ deleteTarget?.name }}</span>
            以确认删除
          </div>
          <Input v-model="deleteInput" placeholder="请输入用户组名称" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="deleteOpen = false">取消</Button>
          <Button variant="destructive" :disabled="!isDeleteMatch" @click="confirmDelete">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="roleAssignOpen">
      <DialogContent class="w-[760px] max-w-[92vw]">
        <DialogHeader>
          <DialogTitle>用户组角色授权 - {{ assignGroup?.name }}</DialogTitle>
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
          <Button @click="submitAssignRoles">保存授权</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="memberAssignOpen">
      <DialogContent class="w-[760px] max-w-[92vw]">
        <DialogHeader>
          <DialogTitle>用户组成员授权 - {{ assignGroup?.name }}</DialogTitle>
        </DialogHeader>
        <div class="max-h-[420px] overflow-y-auto border rounded p-3">
          <div v-if="userList.length === 0" class="text-sm text-muted-foreground">暂无可分配用户</div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <label v-for="user in userList" :key="user.id" class="flex items-center gap-2 text-sm border rounded px-2 py-1">
              <input
                type="checkbox"
                :checked="selectedUserIds.includes(user.id)"
                @change="(event) => toggleUser(user.id, (event.target as HTMLInputElement).checked)"
              >
              <span>{{ user.username }}</span>
            </label>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="memberAssignOpen = false">取消</Button>
          <Button @click="submitAssignMembers">保存授权</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
