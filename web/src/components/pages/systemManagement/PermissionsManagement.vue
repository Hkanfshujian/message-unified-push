<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { rbacApi } from '@/api/rbac'
import AppActionButton from '@/components/ui/AppActionButton.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import { notifySuccess, notifyWarning } from '@/util/uiFeedback'
import { RightOutlined } from '@ant-design/icons-vue'
import type { DoraIconName } from '@/types/app'
import { buildPermissionTree, collectNodeKeys, flattenPermissionTreeRows, type PermissionTreeNode, type PermissionTreeRow } from '@/util/permissionTree'

interface PermissionItem { id: number; code: string; name: string; type: string; method?: string; path?: string; status?: number; sort?: number }

const permissionTypeOptions = [
  { label: '全部类型', value: 'all' },
  { label: '菜单', value: 'menu' },
  { label: '操作', value: 'action' },
  { label: '接口', value: 'api' }
]

const state = reactive({ list: [] as PermissionItem[], search: '', typeFilter: 'all', loading: false })
const formOpen = ref(false)
const editingId = ref<number | null>(null)
const permissionTree = ref<PermissionTreeNode[]>([])
const expandedNodeKeys = ref<string[]>([])
const treeFilterKeyword = ref('')
const formData = reactive({ code: '', name: '', type: 'api', method: 'GET', path: '', sort: 0, status: 1 })

const totalPermissionCount = computed(() => state.list.length)
const totalNodeCount = computed(() => collectNodeKeys(permissionTree.value).length)
const normalizedTreeFilterKeyword = computed(() => treeFilterKeyword.value.trim().toLowerCase())

const permissionTypeLabelMap: Record<string, string> = {
  menu: '菜单',
  action: '操作',
  api: '接口'
}

const isNodeExpanded = (key: string) => expandedNodeKeys.value.includes(key)
const toggleNodeExpand = (key: string) => { expandedNodeKeys.value = isNodeExpanded(key) ? expandedNodeKeys.value.filter(item => item !== key) : [...expandedNodeKeys.value, key] }
const expandAllNodes = () => { expandedNodeKeys.value = collectNodeKeys(permissionTree.value) }
const collapseAllNodes = () => { expandedNodeKeys.value = [] }

const getPermissionTypeLabel = (type: string) => permissionTypeLabelMap[type] || type || '-'
const getPermissionTypeStatus = (type: string) => type === 'api' ? 'online' : type === 'action' ? 'pending' : 'enabled'
const getPermissionMethodLabel = (method?: string) => {
  const normalized = method?.trim().toUpperCase() || ''
  return normalized || 'ROUTE'
}
const getPermissionMethodClass = (method?: string) => {
  const normalized = method?.trim().toUpperCase() || ''
  if (normalized === 'GET') return 'bg-emerald-50 text-emerald-700 border-emerald-100 dark:bg-emerald-500/10 dark:text-emerald-300 dark:border-emerald-500/20'
  if (normalized === 'POST') return 'bg-blue-50 text-blue-700 border-blue-100 dark:bg-blue-500/10 dark:text-blue-300 dark:border-blue-500/20'
  if (normalized === 'PUT' || normalized === 'PATCH') return 'bg-amber-50 text-amber-700 border-amber-100 dark:bg-amber-500/10 dark:text-amber-300 dark:border-amber-500/20'
  if (normalized === 'DELETE') return 'bg-rose-50 text-rose-700 border-rose-100 dark:bg-rose-500/10 dark:text-rose-300 dark:border-rose-500/20'
  return 'bg-slate-50 text-slate-600 border-slate-100 dark:bg-white/5 dark:text-slate-300 dark:border-white/10'
}
const getPermissionPathText = (permission: Pick<PermissionItem, 'path' | 'type'>) => permission.path || (permission.type === 'menu' ? '前端菜单权限' : '未配置路径')
const getNodeIconName = (key: string): DoraIconName => {
  if (key.includes('dashboard')) return 'chart'
  if (key.includes('message')) return 'message'
  if (key.includes('template')) return 'document'
  if (key.includes('sendways')) return 'activity'
  if (key.includes('data')) return 'database'
  if (key.includes('system')) return 'setting'
  if (key.includes('profile')) return 'user'
  return 'app'
}

const permissionTreeRows = computed<PermissionTreeRow<PermissionItem>[]>(() => flattenPermissionTreeRows<PermissionItem>({ nodes: permissionTree.value, keyword: normalizedTreeFilterKeyword.value, expandedKeys: expandedNodeKeys.value }))

const queryList = async () => {
  state.loading = true
  try {
    const rsp = await rbacApi.getPermissions({ page: 1, size: 5000, text: state.search, type: state.typeFilter === 'all' ? '' : state.typeFilter })
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
  formData.status = item.status ?? 1
  formOpen.value = true
}

const submitForm = async () => {
  if (!formData.code.trim() || !formData.name.trim()) {
    notifyWarning('请填写权限编码和权限名称')
    return
  }
  const payload = { code: formData.code.trim(), name: formData.name.trim(), type: formData.type, method: formData.method.trim().toUpperCase(), path: formData.path.trim(), sort: Number(formData.sort) || 0, status: Number(formData.status) === 0 ? 0 : 1 }
  if (editingId.value) {
    await rbacApi.editPermission({ id: editingId.value, ...payload })
    notifySuccess('编辑权限成功')
  } else {
    await rbacApi.addPermission(payload)
    notifySuccess('新增权限成功')
  }
  formOpen.value = false
  await queryList()
}

onMounted(queryList)
</script>

<template>
  <div class="h-full flex flex-col space-y-4">
    <el-card shadow="never">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-col gap-2 md:flex-row md:items-center">
          <el-input v-model="state.search" class="md:w-[280px]" clearable placeholder="按编码/名称/路径搜索" @keyup.enter="queryList" @clear="queryList">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="state.typeFilter" class="md:w-[150px]" @change="queryList">
            <el-option v-for="item in permissionTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-button :icon="Search" @click="queryList">查询</el-button>
          <el-button :icon="Refresh" @click="queryList">刷新</el-button>
        </div>
        <el-button v-permission="'system:rbac:permission'" type="primary" :icon="Plus" @click="openAdd">新增权限</el-button>
      </div>
    </el-card>

    <el-card shadow="never" class="min-h-0 flex-1" body-class="!p-0 h-full">
      <div class="flex h-full flex-col">
        <div class="border-b weak-divider bg-[linear-gradient(180deg,color-mix(in_srgb,var(--brand-600)_5%,transparent),transparent)] p-4 space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="space-y-1">
              <div class="text-[15px] font-semibold text-foreground">权限树视图</div>
              <div class="text-xs text-muted-foreground">当前 <span class="font-semibold text-brand">{{ totalPermissionCount }}</span> 个权限点，<span class="font-semibold text-foreground">{{ totalNodeCount }}</span> 个树节点</div>
            </div>
            <div class="flex items-center gap-2">
              <el-button size="small" @click="expandAllNodes">全部展开</el-button>
              <el-button size="small" @click="collapseAllNodes">全部收起</el-button>
            </div>
          </div>
          <el-input v-model="treeFilterKeyword" placeholder="按模块、权限名、编码或路径筛选当前树" clearable />
        </div>
        <div v-loading="state.loading" class="min-h-0 flex-1 overflow-y-auto bg-[var(--surface-muted)] p-4">
          <AppEmptyState v-if="!state.loading && state.list.length === 0" description="请先初始化或新增权限" />
          <div v-else class="space-y-2">
            <div v-for="row in permissionTreeRows" :key="row.key" :style="{ paddingLeft: `${row.depth * 18}px` }">
              <div v-if="row.type === 'node'" class="group flex min-h-11 items-center gap-3 rounded-xl border border-[var(--dora-border)] bg-[var(--dora-container-bg)] px-3 py-2 shadow-[0_10px_28px_rgba(15,23,42,0.04)] transition-all duration-200 hover:-translate-y-0.5 hover:border-[color-mix(in_srgb,var(--brand-600)_28%,var(--dora-border))] hover:shadow-[0_16px_36px_rgba(15,23,42,0.08)]">
                <button type="button" class="permission-tree-toggle relative inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-transparent bg-transparent text-muted-foreground transition-all duration-200 before:absolute before:inset-1 before:rounded-full before:bg-[color-mix(in_srgb,var(--brand-600)_7%,transparent)] before:opacity-0 before:transition-opacity hover:text-brand hover:before:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[color-mix(in_srgb,var(--brand-600)_32%,transparent)]" :aria-label="isNodeExpanded(row.node.key) ? '收起模块' : '展开模块'" @click="toggleNodeExpand(row.node.key)">
                  <span class="permission-tree-toggle-icon relative z-10 inline-flex transition-transform duration-200" :class="isNodeExpanded(row.node.key) ? 'rotate-90 text-brand permission-tree-toggle-icon-expanded' : 'rotate-0'">
                    <RightOutlined class="text-[13px]" />
                  </span>
                </button>
                <span class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-[linear-gradient(135deg,var(--brand-600),#0ea5e9)] text-white shadow-[0_10px_22px_rgba(37,99,235,0.24)]">
                  <DoraIcon :name="getNodeIconName(row.node.key)" :size="16" />
                </span>
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-semibold text-foreground">{{ row.node.label }}</div>
                  <div class="text-xs text-muted-foreground">{{ row.node.permissions.length }} 个直属权限 · {{ row.node.children.length }} 个子模块</div>
                </div>
              </div>
              <div v-else class="group grid min-h-[68px] grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-xl border border-[var(--dora-border)] bg-[var(--dora-container-bg)] px-3 py-2.5 transition-all duration-200 hover:border-[color-mix(in_srgb,var(--brand-600)_22%,var(--dora-border))] hover:bg-[var(--dora-container-bg)] hover:shadow-[0_12px_30px_rgba(15,23,42,0.06)]">
                <el-tooltip placement="top" effect="dark">
                  <template #content>
                    <div class="max-w-[440px] space-y-1">
                      <div class="font-medium">{{ row.permission.name }}</div>
                      <div>{{ row.permission.code }}</div>
                      <div>{{ row.permission.type }} · {{ row.permission.method || '-' }} · {{ row.permission.path || '-' }}</div>
                    </div>
                  </template>
                  <div class="min-w-0 cursor-help space-y-2">
                    <div class="flex min-w-0 items-center gap-2">
                      <span class="truncate text-sm font-semibold text-foreground">{{ row.permission.name }}</span>
                      <span class="shrink-0 rounded-full border px-2 py-0.5 font-mono text-[10px] font-semibold leading-none tracking-wide" :class="getPermissionMethodClass(row.permission.method)">{{ getPermissionMethodLabel(row.permission.method) }}</span>
                    </div>
                    <div class="grid min-w-0 grid-cols-[minmax(120px,0.9fr)_minmax(120px,1.1fr)] gap-2 text-xs text-muted-foreground">
                      <div class="min-w-0 rounded-lg bg-[var(--surface-muted)] px-2 py-1">
                        <span class="mr-1 text-[11px] text-muted-foreground/70">编码</span>
                        <span class="font-mono text-foreground/80">{{ row.permission.code }}</span>
                      </div>
                      <div class="min-w-0 rounded-lg bg-[var(--surface-muted)] px-2 py-1">
                        <span class="mr-1 text-[11px] text-muted-foreground/70">路径</span>
                        <span class="truncate font-mono text-foreground/80">{{ getPermissionPathText(row.permission) }}</span>
                      </div>
                    </div>
                  </div>
                </el-tooltip>
                <div class="flex shrink-0 items-center gap-2">
                  <AppStatusTag :status="getPermissionTypeStatus(row.permission.type)" :label-map="{ enabled: getPermissionTypeLabel(row.permission.type), pending: getPermissionTypeLabel(row.permission.type), online: getPermissionTypeLabel(row.permission.type) }" />
                  <AppActionButton v-permission="'system:rbac:permission'" @click="openEdit(row.permission)">编辑</AppActionButton>
                </div>
              </div>
            </div>
            <div v-if="permissionTreeRows.length === 0" class="h-10 px-2 flex items-center text-sm text-muted-foreground">未匹配到相关权限，请尝试其他关键字</div>
          </div>
        </div>
      </div>
    </el-card>

    <AppFormDrawer v-model="formOpen" :title="editingId ? '编辑权限' : '新增权限'" size="620px" :show-footer="false">
      <div class="permission-form-content">
      <el-form label-position="top" class="permission-form-sections">
        <section class="permission-form-section">
          <header><div><span>01</span><h3>权限定义</h3></div><p>定义可被角色授权的菜单、操作或接口能力</p></header>
          <div class="permission-form-section-body">
            <div class="permission-subsection-title">基础信息</div>
            <el-form-item label="权限编码" required><el-input v-model="formData.code" placeholder="权限编码，例如 message:template:view" /></el-form-item>
            <el-form-item label="权限名称" required><el-input v-model="formData.name" placeholder="权限名称" /></el-form-item>
            <el-form-item label="权限类型"><el-select v-model="formData.type" class="w-full"><el-option label="菜单" value="menu" /><el-option label="操作" value="action" /><el-option label="接口" value="api" /></el-select></el-form-item>
            <div class="permission-subsection-title">匹配规则 <span>配置请求方法及接口或前端路由路径</span></div>
            <el-form-item label="请求方法"><el-input v-model="formData.method" placeholder="请求方法，如 GET / POST" /></el-form-item>
            <el-form-item label="路径"><el-input v-model="formData.path" placeholder="接口路径或前端路由路径" /></el-form-item>
            <div class="permission-subsection-title">展示配置 <span>控制可用状态与同级展示顺序</span></div>
            <div class="permission-display-grid">
              <el-form-item label="状态"><el-select v-model="formData.status" class="w-full"><el-option label="启用" :value="1" /><el-option label="停用" :value="0" /></el-select></el-form-item>
              <el-form-item label="排序"><el-input-number v-model="formData.sort" :min="0" class="!w-full" /></el-form-item>
            </div>
          </div>
        </section>
      </el-form>
      </div>
      <template #footer>
        <el-button @click="formOpen = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </AppFormDrawer>
  </div>
</template>

<style scoped>
.permission-form-content, .permission-form-sections { display: grid; gap: 12px; }
.permission-form-section { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.permission-form-section header p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.permission-form-section header { display: flex; align-items: center; justify-content: space-between; min-height: 38px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.permission-form-section header > div { display: flex; align-items: center; gap: 8px; }
.permission-form-section header span { color: var(--brand-600); font-family: monospace; font-size: 10px; font-weight: 800; }
.permission-form-section header h3 { margin: 0; font-size: 13px; font-weight: 700; }
.permission-form-section header p { margin: 0; }
.permission-form-section-body { display: grid; gap: 12px; padding: 14px; }
.permission-form-section-body :deep(.el-form-item) { margin-bottom: 0; }
.permission-subsection-title { display: flex; align-items: baseline; gap: 8px; margin-top: 2px; padding-top: 12px; border-top: 1px solid var(--app-overlay-border); font-size: 12px; font-weight: 700; }
.permission-subsection-title:first-child { margin-top: 0; padding-top: 0; border-top: 0; }
.permission-subsection-title span { color: var(--admin-text-muted); font-size: 10px; font-weight: 400; }
.permission-display-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.permission-tree-toggle {
  color: color-mix(in srgb, var(--admin-text-muted) 88%, var(--foreground) 12%);
}

.dark .permission-tree-toggle {
  border-color: rgba(148, 163, 184, 0.12);
  background: rgba(15, 23, 42, 0.20);
  color: #cbd5e1;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

.dark .permission-tree-toggle::before {
  background: rgba(59, 130, 246, 0.14);
}

.dark .permission-tree-toggle:hover,
.dark .permission-tree-toggle:focus-visible {
  border-color: rgba(96, 165, 250, 0.22);
  background: rgba(59, 130, 246, 0.10);
  color: #bfdbfe;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.dark .permission-tree-toggle-icon {
  color: inherit;
}

.dark .permission-tree-toggle-icon-expanded {
  color: #93c5fd !important;
}

@media (max-width: 760px) {
  .permission-display-grid { grid-template-columns: 1fr; }
  .permission-form-section header { align-items: center; flex-direction: row; }
  .permission-form-section header p { display: none; }
}
</style>
