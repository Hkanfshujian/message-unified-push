<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { rbacApi } from '@/api/rbac'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import { ChevronDown, ChevronRight, FolderTree } from 'lucide-vue-next'

interface PermissionItem {
  id: number
  code: string
  name: string
  type: string
  method: string
  path: string
  status: number
  sort: number
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

const permissionTypeOptions = [
  { label: '全部类型', value: 'all' },
  { label: '菜单', value: 'menu' },
  { label: '操作', value: 'action' },
  { label: '接口', value: 'api' },
]

const state = reactive({
  list: [] as PermissionItem[],
  search: '',
  typeFilter: 'all',
  loading: false,
})

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const permissionTree = ref<PermissionTreeNode[]>([])
const expandedNodeKeys = ref<string[]>([])
const treeFilterKeyword = ref('')
const formData = reactive({
  code: '',
  name: '',
  type: 'api',
  method: 'GET',
  path: '',
  sort: 0,
  status: 1
})

const totalPermissionCount = computed(() => state.list.length)
const totalNodeCount = computed(() => collectNodeKeys(permissionTree.value).length)
const normalizedTreeFilterKeyword = computed(() => treeFilterKeyword.value.trim().toLowerCase())

const permissionTreeTemplate: PermissionTreeTemplateNode[] = [
  { key: 'dashboard', label: '数据统计', matcher: permission => permission.code.startsWith('dashboard:') },
  {
    key: 'message',
    label: '消息管理',
    children: [
      { key: 'message-cron', label: '定时消息', matcher: permission => permission.code.startsWith('message:cron:') }
    ]
  },
  { key: 'template', label: '模板管理', matcher: permission => permission.code.startsWith('message:template:') },
  { key: 'sendways', label: '渠道管理', matcher: permission => permission.code.startsWith('message:sendways:') },
  { key: 'sendlogs', label: '日志管理', matcher: permission => permission.code.startsWith('message:sendlogs:') },
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
      { key: 'system-role', label: '角色管理', matcher: permission => permission.code === 'system:rbac:role' },
      { key: 'system-group', label: '用户组管理', matcher: permission => permission.code === 'system:rbac:group' },
      { key: 'system-permission', label: '权限管理', matcher: permission => permission.code === 'system:rbac:permission' },
      { key: 'system-user', label: '用户管理', matcher: permission => permission.code === 'system:rbac:user' },
      { key: 'system-identity', label: '身份映射', matcher: permission => permission.code === 'system:rbac:identity' }
    ]
  },
  { key: 'profile', label: '个人设置', matcher: permission => permission.code.startsWith('profile:settings:') }
]

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

const sortPermissionTree = (nodes: PermissionTreeNode[]) => {
  nodes.forEach((node) => {
    node.permissions.sort((a, b) => {
      const sortDiff = (a.sort || 0) - (b.sort || 0)
      if (sortDiff !== 0) return sortDiff
      return a.name.localeCompare(b.name, 'zh-CN')
    })
    if (node.children.length > 0) {
      sortPermissionTree(node.children)
    }
  })
}

const buildNodeFromTemplate = (template: PermissionTreeTemplateNode): PermissionTreeNode => ({
  key: template.key,
  label: template.label,
  children: (template.children || []).map(child => buildNodeFromTemplate(child)),
  permissions: []
})

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
  sortPermissionTree(finalRoots)
  return finalRoots
}

const isNodeExpanded = (key: string) => expandedNodeKeys.value.includes(key)

const toggleNodeExpand = (key: string) => {
  if (isNodeExpanded(key)) {
    expandedNodeKeys.value = expandedNodeKeys.value.filter(item => item !== key)
  } else {
    expandedNodeKeys.value.push(key)
  }
}

const expandAllNodes = () => {
  expandedNodeKeys.value = collectNodeKeys(permissionTree.value)
}

const collapseAllNodes = () => {
  expandedNodeKeys.value = []
}

const matchPermission = (permission: PermissionItem, keyword: string) => {
  const name = permission.name.toLowerCase()
  const code = permission.code.toLowerCase()
  const path = (permission.path || '').toLowerCase()
  return name.includes(keyword) || code.includes(keyword) || path.includes(keyword)
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
  const keyword = normalizedTreeFilterKeyword.value
  const hasKeyword = keyword.length > 0
  const rows: PermissionTreeRow[] = []
  const walk = (nodes: PermissionTreeNode[], depth: number) => {
    nodes.forEach((node) => {
      if (hasKeyword && !matchNode(node, keyword)) return
      rows.push({ type: 'node', key: `node:${node.key}`, depth, node })
      const shouldExpand = hasKeyword || isNodeExpanded(node.key)
      if (!shouldExpand) return
      const nodeLabelMatched = hasKeyword && node.label.toLowerCase().includes(keyword)
      const visiblePermissions = hasKeyword
        ? node.permissions.filter(permission => nodeLabelMatched || matchPermission(permission, keyword))
        : node.permissions
      visiblePermissions.forEach((permission) => {
        rows.push({
          type: 'permission',
          key: `permission:${permission.id}`,
          depth: depth + 1,
          permission
        })
      })
      if (node.children.length > 0) {
        walk(node.children, depth + 1)
      }
    })
  }
  walk(permissionTree.value, 0)
  return rows
})

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getPermissions({
      page: 1,
      size: 5000,
      text: state.search,
      type: state.typeFilter === 'all' ? '' : state.typeFilter
    })
    state.list = rsp.data.data?.lists || []
    permissionTree.value = buildPermissionTree(state.list)
    collapseAllNodes()
  } finally {
    state.loading = false
  }
}

const openAdd = () => {
  editingId.value = null
  formData.code = ''
  formData.name = ''
  formData.type = 'api'
  formData.method = 'GET'
  formData.path = ''
  formData.sort = 0
  formData.status = 1
  formOpen.value = true
}

const openEdit = (item: PermissionItem) => {
  editingId.value = item.id
  formData.code = item.code
  formData.name = item.name
  formData.type = item.type
  formData.method = item.method || 'GET'
  formData.path = item.path || ''
  formData.sort = item.sort || 0
  formData.status = item.status
  formOpen.value = true
}

const submitForm = async () => {
  if (!formData.code.trim() || !formData.name.trim()) {
    toast.error('请填写权限编码和权限名称')
    return
  }
  const payload = {
    code: formData.code.trim(),
    name: formData.name.trim(),
    type: formData.type,
    method: formData.method.trim().toUpperCase(),
    path: formData.path.trim(),
    sort: Number(formData.sort) || 0,
    status: Number(formData.status) === 0 ? 0 : 1
  }
  if (editingId.value) {
    await rbacApi.editPermission({ id: editingId.value, ...payload })
    toast.success('编辑权限成功')
  } else {
    await rbacApi.addPermission(payload)
    toast.success('新增权限成功')
  }
  formOpen.value = false
  await queryList()
}

onMounted(async () => {
  await queryList()
})
</script>

<template>
  <div class="h-full flex flex-col">
    <TooltipProvider :delay-duration="120" class="flex-1 min-h-0">
      <div class="border rounded h-full overflow-y-auto">
        <div class="sticky top-0 z-30 p-3 border-b bg-white/90 dark:bg-slate-950/90 backdrop-blur supports-[backdrop-filter]:bg-white/75 space-y-2 relative">
          <div class="toolbar mb-0">
            <div class="search-group">
              <Input v-model="state.search" placeholder="按编码/名称/路径搜索" class="search-input" @keyup.enter="queryList" />
              <Select v-model="state.typeFilter" class="w-full" @update:model-value="queryList">
                <SelectTrigger class="filter-select w-full">
                  <SelectValue placeholder="类型筛选" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem v-for="item in permissionTypeOptions" :key="item.value" :value="item.value">
                      {{ item.label }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
            <Button @click="openAdd">新增权限</Button>
          </div>

          <div class="rounded border px-3 py-2 bg-slate-50/95 dark:bg-slate-900/35 shadow-[0_1px_2px_rgba(15,23,42,0.06)] dark:shadow-[0_1px_2px_rgba(0,0,0,0.22)]">
            <div class="flex flex-wrap items-center gap-2 justify-between">
              <div class="text-sm text-gray-700 dark:text-slate-200">
                当前
                <span class="font-semibold text-brand-600 dark:text-brand-300">{{ totalPermissionCount }}</span>
                个权限点，
                <span class="font-semibold">{{ totalNodeCount }}</span>
                个树节点
              </div>
              <div class="flex items-center gap-2">
                <Button size="sm" variant="outline" @click="expandAllNodes">全部展开</Button>
                <Button size="sm" variant="outline" @click="collapseAllNodes">全部收起</Button>
              </div>
            </div>
            <div class="mt-2">
              <Input
                v-model="treeFilterKeyword"
                placeholder="按模块、权限名、编码或路径筛选当前树"
              />
            </div>
          </div>
          <div class="pointer-events-none absolute left-0 right-0 -bottom-3 h-3 bg-gradient-to-b from-black/8 to-transparent dark:from-black/30 dark:to-transparent"></div>
        </div>
        <div class="p-3">
          <div v-if="!state.loading && state.list.length === 0" class="py-8">
            <EmptyTableState title="暂无权限" description="请先初始化或新增权限" />
          </div>
          <div v-else class="space-y-1">
            <div
              v-for="row in permissionTreeRows"
              :key="row.key"
              class="rounded border border-transparent hover:border-gray-200 dark:hover:border-slate-700"
              :style="{ paddingLeft: `${row.depth * 18}px` }"
            >
              <div v-if="row.type === 'node'" class="h-9 px-2 flex items-center gap-2">
                <button type="button" class="text-gray-500" @click="toggleNodeExpand(row.node.key)">
                  <ChevronDown v-if="isNodeExpanded(row.node.key)" class="w-4 h-4" />
                  <ChevronRight v-else class="w-4 h-4" />
                </button>
                <FolderTree class="w-4 h-4 text-brand-500" />
                <span class="text-sm font-medium text-gray-800 dark:text-slate-100">{{ row.node.label }}</span>
              </div>
              <div v-else class="min-h-10 px-2 py-1 flex items-center justify-between gap-3">
                <Tooltip>
                  <TooltipTrigger as-child>
                    <div class="min-w-0 text-sm text-gray-900 dark:text-slate-100 truncate cursor-help">
                      {{ row.permission.name }}
                    </div>
                  </TooltipTrigger>
                  <TooltipContent class="max-w-[440px]">
                    <div class="space-y-1">
                      <div class="font-medium">{{ row.permission.name }}</div>
                      <div class="text-muted-foreground break-all">{{ row.permission.code }}</div>
                      <div class="text-muted-foreground break-all">
                        {{ row.permission.type }} · {{ row.permission.method || '-' }} · {{ row.permission.path || '-' }}
                      </div>
                    </div>
                  </TooltipContent>
                </Tooltip>
                <Button size="sm" variant="outline" @click="openEdit(row.permission)">编辑</Button>
              </div>
            </div>
            <div
              v-if="permissionTreeRows.length === 0"
              class="h-10 px-2 flex items-center text-sm text-gray-500"
            >
              未匹配到相关权限，请尝试其他关键字
            </div>
          </div>
        </div>
      </div>
    </TooltipProvider>

    <Dialog v-model:open="formOpen">
      <DialogContent class="w-[620px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>{{ editingId ? '编辑权限' : '新增权限' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <Input v-model="formData.code" placeholder="权限编码，例如 message:template:view" />
          <Input v-model="formData.name" placeholder="权限名称" />
          <Select v-model="formData.type" class="w-full">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="选择权限类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="menu">菜单</SelectItem>
                <SelectItem value="action">操作</SelectItem>
                <SelectItem value="api">接口</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
          <Input v-model="formData.method" placeholder="请求方法，如 GET / POST" />
          <Input v-model="formData.path" placeholder="接口路径或前端路由路径" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="formOpen = false">取消</Button>
          <Button @click="submitForm">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
