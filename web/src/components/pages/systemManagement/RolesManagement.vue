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
import { ChevronDown, ChevronRight, FolderTree } from 'lucide-vue-next'

interface RoleItem {
  id: number
  code: string
  name: string
  description: string
  status: number
  created_on: string
  modified_on: string
}

interface PermissionItem {
  id: number
  code: string
  name: string
  type: string
}

interface PermissionTreeNode {
  key: string
  label: string
  children: PermissionTreeNode[]
  permissions: PermissionItem[]
}

interface PermissionTreeTemplateNode {
  key: string
  label: string
  matcher?: (permission: PermissionItem) => boolean
  children?: PermissionTreeTemplateNode[]
}

type PermissionTreeRow =
  | { type: 'node'; key: string; depth: number; node: PermissionTreeNode }
  | { type: 'permission'; key: string; depth: number; permission: PermissionItem }

const state = reactive({
  list: [] as RoleItem[],
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
const selectedPermissionIds = ref<number[]>([])
const permissionTree = ref<PermissionTreeNode[]>([])
const expandedNodeKeys = ref<string[]>([])
const assignFilterKeyword = ref('')
const showSelectedOnly = ref(false)

const totalPages = computed(() => Math.ceil(state.total / state.pageSize))
const isDeleteMatch = computed(() => {
  const roleName = deleteTarget.value?.name || ''
  return deleteInput.value.trim().toLowerCase() === roleName.trim().toLowerCase() && roleName.length > 0
})

const permissionTreeTemplate: PermissionTreeTemplateNode[] = [
  {
    key: 'dashboard',
    label: '数据统计',
    matcher: permission => permission.code.startsWith('dashboard:')
  },
  {
    key: 'message',
    label: '消息管理',
    children: [
      {
        key: 'message-cron',
        label: '定时消息',
        matcher: permission => permission.code.startsWith('message:cron:')
      }
    ]
  },
  {
    key: 'template',
    label: '模板管理',
    matcher: permission => permission.code.startsWith('message:template:')
  },
  {
    key: 'sendways',
    label: '渠道管理',
    matcher: permission => permission.code.startsWith('message:sendways:')
  },
  {
    key: 'sendlogs',
    label: '日志管理',
    matcher: permission => permission.code.startsWith('message:sendlogs:')
  },
  {
    key: 'system',
    label: '系统管理',
    matcher: permission => permission.code === 'system:rbac:view',
    children: [
      {
        key: 'system-settings',
        label: '系统设置',
        matcher: permission => permission.code.startsWith('system:settings:') || permission.code === 'system:loginlogs:view'
      },
      {
        key: 'system-role',
        label: '角色管理',
        matcher: permission => permission.code === 'system:rbac:role'
      },
      {
        key: 'system-group',
        label: '用户组管理',
        matcher: permission => permission.code === 'system:rbac:group'
      },
      {
        key: 'system-permission',
        label: '权限管理',
        matcher: permission => permission.code === 'system:rbac:permission'
      },
      {
        key: 'system-user',
        label: '用户管理',
        matcher: permission => permission.code === 'system:rbac:user'
      },
      {
        key: 'system-identity',
        label: '身份映射',
        matcher: permission => permission.code === 'system:rbac:identity'
      }
    ]
  },
  {
    key: 'profile',
    label: '个人设置',
    matcher: permission => permission.code.startsWith('profile:settings:')
  }
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
    toast.error('请填写角色编码和角色名称')
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
    toast.success('编辑角色成功')
  } else {
    await rbacApi.addRole(payload)
    toast.success('新增角色成功')
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
  toast.success('删除角色成功')
  await queryList()
}

const buildNodeFromTemplate = (template: PermissionTreeTemplateNode): PermissionTreeNode => {
  return {
    key: template.key,
    label: template.label,
    permissions: [],
    children: (template.children || []).map(child => buildNodeFromTemplate(child))
  }
}

const sortPermissionsInTree = (nodes: PermissionTreeNode[]) => {
  nodes.forEach((node) => {
    node.permissions.sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
    if (node.children.length > 0) {
      sortPermissionsInTree(node.children)
    }
  })
}

const findBestMatchedNode = (
  permission: PermissionItem,
  template: PermissionTreeTemplateNode,
  node: PermissionTreeNode
): PermissionTreeNode | null => {
  for (let i = 0; i < (template.children || []).length; i += 1) {
    const childTemplate = template.children![i]
    const childNode = node.children[i]
    const matchedChild = findBestMatchedNode(permission, childTemplate, childNode)
    if (matchedChild) {
      return matchedChild
    }
  }
  if (template.matcher && template.matcher(permission)) {
    return node
  }
  return null
}

const buildPermissionTree = (permissions: PermissionItem[]): PermissionTreeNode[] => {
  const roots = permissionTreeTemplate.map(item => buildNodeFromTemplate(item))
  const fallbackNode: PermissionTreeNode = {
    key: 'others',
    label: '其他权限',
    children: [],
    permissions: []
  }
  permissions.forEach((permission) => {
    let matched = false
    for (let i = 0; i < permissionTreeTemplate.length; i += 1) {
      const template = permissionTreeTemplate[i]
      const node = roots[i]
      const targetNode = findBestMatchedNode(permission, template, node)
      if (targetNode) {
        targetNode.permissions.push(permission)
        matched = true
        break
      }
    }
    if (!matched) {
      fallbackNode.permissions.push(permission)
    }
  })
  const finalRoots = roots.filter(item => item.permissions.length > 0 || item.children.length > 0)
  if (fallbackNode.permissions.length > 0) {
    finalRoots.push(fallbackNode)
  }
  sortPermissionsInTree(finalRoots)
  return finalRoots
}

const collectNodeKeys = (nodes: PermissionTreeNode[]): string[] => {
  const keys: string[] = []
  const walk = (items: PermissionTreeNode[]) => {
    items.forEach((item) => {
      keys.push(item.key)
      if (item.children.length > 0) {
        walk(item.children)
      }
    })
  }
  walk(nodes)
  return keys
}

const openAssignPermissions = async (item: RoleItem) => {
  assignRole.value = item
  assignOpen.value = true
  assignFilterKeyword.value = ''
  showSelectedOnly.value = false
  const [permissionsRsp, selectedRsp] = await Promise.all([
    rbacApi.getPermissions({ page: 1, size: 200 }),
    rbacApi.getRolePermissionIDs(item.id),
  ])
  permissionList.value = permissionsRsp.data.data?.lists || []
  permissionTree.value = buildPermissionTree(permissionList.value)
  expandedNodeKeys.value = collectNodeKeys(permissionTree.value)
  selectedPermissionIds.value = selectedRsp.data.data?.permission_ids || []
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
  toast.success('角色权限授权成功')
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

const collectNodePermissionIds = (node: PermissionTreeNode): number[] => {
  const ids = node.permissions.map(item => item.id)
  node.children.forEach((child) => {
    ids.push(...collectNodePermissionIds(child))
  })
  return ids
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

const matchPermission = (permission: PermissionItem, keyword: string) => {
  const name = permission.name.toLowerCase()
  const code = permission.code.toLowerCase()
  return name.includes(keyword) || code.includes(keyword)
}

const matchNode = (node: PermissionTreeNode, keyword: string): boolean => {
  if (node.label.toLowerCase().includes(keyword)) {
    return true
  }
  if (node.permissions.some(permission => matchPermission(permission, keyword))) {
    return true
  }
  return node.children.some(child => matchNode(child, keyword))
}

const permissionTreeRows = computed<PermissionTreeRow[]>(() => {
  const keyword = normalizedFilterKeyword.value
  const hasKeyword = keyword.length > 0
  const selectedOnly = showSelectedOnly.value
  const buildNodeRows = (node: PermissionTreeNode, depth: number): PermissionTreeRow[] => {
    const rows: PermissionTreeRow[] = []
    const childRows: PermissionTreeRow[] = []
    node.children.forEach((child) => {
      childRows.push(...buildNodeRows(child, depth + 1))
    })
    const hasVisibleChildren = childRows.length > 0
    const nodeLabelMatched = hasKeyword && node.label.toLowerCase().includes(keyword)
    const visiblePermissions = hasKeyword
      ? node.permissions.filter(permission => {
        if (selectedOnly && !selectedPermissionIds.value.includes(permission.id)) return false
        return nodeLabelMatched || matchPermission(permission, keyword)
      })
      : node.permissions.filter(permission => {
        if (!selectedOnly) return true
        return selectedPermissionIds.value.includes(permission.id)
      })
    const visibleNode = visiblePermissions.length > 0 || hasVisibleChildren
    if (!visibleNode) {
      return rows
    }
    rows.push({
      type: 'node',
      key: `node:${node.key}`,
      depth,
      node
    })
    const shouldExpand = hasKeyword || selectedOnly || isNodeExpanded(node.key)
    if (!shouldExpand) {
      return rows
    }
    visiblePermissions.forEach((permission) => {
      rows.push({
        type: 'permission',
        key: `permission:${permission.id}`,
        depth: depth + 1,
        permission
      })
    })
    rows.push(...childRows)
    return rows
  }
  const rows: PermissionTreeRow[] = []
  permissionTree.value.forEach((node) => {
    rows.push(...buildNodeRows(node, 0))
  })
  return rows
})

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
      <Button @click="openAdd">新增角色</Button>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="w-20">ID</TableHead>
            <TableHead>角色编码</TableHead>
            <TableHead>角色名称</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>描述</TableHead>
            <TableHead class="text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="!state.loading && state.list.length === 0">
            <TableCell colspan="6" class="empty-state">
              <EmptyTableState title="暂无角色" description="请先创建角色并完成授权" />
            </TableCell>
          </TableRow>
          <TableRow v-for="item in state.list" :key="item.id">
            <TableCell>{{ item.id }}</TableCell>
            <TableCell>{{ item.code }}</TableCell>
            <TableCell>{{ item.name }}</TableCell>
            <TableCell>{{ item.status === 1 ? '启用' : '禁用' }}</TableCell>
            <TableCell>{{ item.description || '-' }}</TableCell>
            <TableCell class="text-center space-x-2">
              <Button size="sm" variant="outline" @click="openAssignPermissions(item)">授权</Button>
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
          <DialogTitle>{{ editingId ? '编辑角色' : '新增角色' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <Input v-model="formData.code" placeholder="角色编码，例如 role_admin" />
          <Input v-model="formData.name" placeholder="角色名称" />
          <Input v-model="formData.description" placeholder="角色描述" />
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
          <DialogTitle>确认删除角色</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">
            请输入角色名称
            <span class="text-red-500 mx-1">{{ deleteTarget?.name }}</span>
            以确认删除
          </div>
          <Input v-model="deleteInput" placeholder="请输入角色名称" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="deleteOpen = false">取消</Button>
          <Button variant="destructive" :disabled="!isDeleteMatch" @click="confirmDelete">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="assignOpen">
      <DialogContent class="w-[860px] max-w-[94vw]">
        <DialogHeader>
          <DialogTitle>角色权限授权 - {{ assignRole?.name }}</DialogTitle>
        </DialogHeader>
        <div v-if="permissionList.length > 0" class="rounded border px-3 py-2 bg-slate-50 dark:bg-slate-900/40">
          <div class="flex flex-wrap items-center gap-2 justify-between">
            <div class="text-sm text-gray-700 dark:text-slate-200">
              已选
              <span class="font-semibold text-brand-600 dark:text-brand-300">{{ selectedPermissionCount }}</span>
              /
              <span class="font-semibold">{{ totalPermissionCount }}</span>
              项权限，当前树节点
              <span class="font-semibold">{{ totalNodeCount }}</span>
              个
            </div>
            <div class="flex items-center gap-2">
              <Button size="sm" variant="outline" @click="expandAllNodes">全部展开</Button>
              <Button size="sm" variant="outline" @click="collapseAllNodes">全部收起</Button>
              <Button size="sm" variant="outline" @click="selectAllPermissions">全选权限</Button>
              <Button size="sm" variant="outline" @click="clearAllPermissions">清空权限</Button>
              <Button size="sm" :variant="showSelectedOnly ? 'default' : 'outline'" @click="showSelectedOnly = !showSelectedOnly">
                仅看已选
              </Button>
            </div>
          </div>
          <div class="mt-2">
            <Input
              v-model="assignFilterKeyword"
              placeholder="按模块或权限名称筛选，例如 模板 / 渠道 / 用户管理"
            />
          </div>
        </div>
        <div class="max-h-[480px] overflow-y-auto border rounded p-3">
          <div v-if="permissionList.length === 0" class="text-sm text-muted-foreground">暂无可分配权限</div>
          <div v-else class="space-y-1">
            <div
              v-for="row in permissionTreeRows"
              :key="row.key"
              class="rounded border border-transparent hover:border-gray-200 dark:hover:border-slate-700"
              :style="{ paddingLeft: `${row.depth * 18}px` }"
            >
              <div v-if="row.type === 'node'" class="h-9 px-2 flex items-center gap-2">
                <button
                  type="button"
                  class="w-4 h-4 rounded border flex items-center justify-center text-[11px] font-bold leading-none transition-colors"
                  :class="isNodeChecked(row.node)
                    ? 'border-brand-500 bg-brand-500 text-white'
                    : 'border-gray-300 text-transparent dark:border-slate-600'"
                  :title="getNodeToggleHint(row.node)"
                  @click="toggleNodePermissions(row.node, !isNodeChecked(row.node))"
                >
                  <span v-if="isNodeChecked(row.node)">✓</span>
                </button>
                <button
                  type="button"
                  class="flex items-center gap-1 text-sm font-medium text-gray-800 dark:text-slate-100"
                  @click="toggleNodeExpand(row.node.key)"
                >
                  <ChevronDown v-if="isNodeExpanded(row.node.key)" class="w-4 h-4 text-gray-500" />
                  <ChevronRight v-else class="w-4 h-4 text-gray-500" />
                  <FolderTree class="w-4 h-4 text-brand-500" />
                  <span>{{ row.node.label }}</span>
                </button>
              </div>
              <label v-else class="h-9 px-2 flex items-center gap-2 text-sm cursor-pointer">
                <input
                  type="checkbox"
                  :checked="selectedPermissionIds.includes(row.permission.id)"
                  @change="(event) => togglePermission(row.permission.id, (event.target as HTMLInputElement).checked)"
                >
                <span class="text-gray-900 dark:text-slate-100">{{ row.permission.name }}</span>
              </label>
            </div>
            <div
              v-if="permissionTreeRows.length === 0"
              class="h-10 px-2 flex items-center text-sm text-gray-500"
            >
              未匹配到相关权限，请尝试其他关键字
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="assignOpen = false">取消</Button>
          <Button @click="submitAssignPermissions">保存授权</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
