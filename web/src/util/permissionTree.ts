export interface PermissionTreeItem {
  id: number
  code: string
  name: string
  type: string
  method?: string
  path?: string
  status?: number
  sort?: number
}

export interface PermissionTreeNode<T extends PermissionTreeItem = PermissionTreeItem> {
  key: string
  label: string
  children: PermissionTreeNode<T>[]
  permissions: T[]
}

export interface PermissionTreeTemplateNode<T extends PermissionTreeItem = PermissionTreeItem> {
  key: string
  label: string
  matcher?: (permission: T) => boolean
  children?: PermissionTreeTemplateNode<T>[]
}

export type PermissionTreeRow<T extends PermissionTreeItem = PermissionTreeItem> =
  | { type: 'node'; key: string; depth: number; node: PermissionTreeNode<T> }
  | { type: 'permission'; key: string; depth: number; permission: T }

export const permissionTreeTemplate: PermissionTreeTemplateNode[] = [
  { key: 'dashboard', label: '数据统计', matcher: permission => permission.code.startsWith('dashboard:') },
  {
    key: 'message',
    label: '消息管理',
    children: [
      { key: 'message-cron', label: '定时消息', matcher: permission => permission.code.startsWith('message:cron:') },
      { key: 'message-center', label: '消息中心', matcher: permission => permission.code.startsWith('message:center:') },
      { key: 'message-system', label: '系统通知', matcher: permission => permission.code.startsWith('message:system:') }
    ]
  },
  { key: 'template', label: '模板管理', matcher: permission => permission.code.startsWith('message:template:') },
  { key: 'sendways', label: '渠道管理', matcher: permission => permission.code.startsWith('message:sendways:') },
  { key: 'sendlogs', label: '日志管理', matcher: permission => permission.code.startsWith('message:sendlogs:') },
  { key: 'data', label: '数据管理', matcher: permission => permission.code.startsWith('data:') },
  {
    key: 'system',
    label: '系统管理',
    matcher: permission => permission.code === 'system:rbac:view',
    children: [
      { key: 'system-settings', label: '系统设置', matcher: permission => permission.code.startsWith('system:settings:') || permission.code === 'system:loginlogs:view' },
      { key: 'system-role', label: '角色管理', matcher: permission => permission.code === 'system:rbac:role' },
      { key: 'system-group', label: '用户组管理', matcher: permission => permission.code === 'system:rbac:group' },
      { key: 'system-permission', label: '权限管理', matcher: permission => permission.code === 'system:rbac:permission' },
      { key: 'system-user', label: '用户管理', matcher: permission => permission.code === 'system:rbac:user' },
      { key: 'system-identity', label: '身份映射', matcher: permission => permission.code === 'system:rbac:identity' }
    ]
  },
  { key: 'profile', label: '个人设置', matcher: permission => permission.code.startsWith('profile:settings:') }
]

const buildNodeFromTemplate = <T extends PermissionTreeItem>(template: PermissionTreeTemplateNode<T>): PermissionTreeNode<T> => ({
  key: template.key,
  label: template.label,
  children: (template.children || []).map(child => buildNodeFromTemplate(child)),
  permissions: []
})

const findBestMatchedNode = <T extends PermissionTreeItem>(permission: T, template: PermissionTreeTemplateNode<T>, node: PermissionTreeNode<T>): PermissionTreeNode<T> | null => {
  for (let i = 0; i < (template.children || []).length; i += 1) {
    const matchedChild = findBestMatchedNode(permission, template.children![i], node.children[i])
    if (matchedChild) return matchedChild
  }
  return template.matcher?.(permission) ? node : null
}

const sortPermissionTree = <T extends PermissionTreeItem>(nodes: PermissionTreeNode<T>[]) => {
  nodes.forEach((node) => {
    node.permissions.sort((a, b) => (a.sort || 0) - (b.sort || 0) || a.name.localeCompare(b.name, 'zh-CN'))
    if (node.children.length > 0) sortPermissionTree(node.children)
  })
}

const prunePermissionTree = <T extends PermissionTreeItem>(nodes: PermissionTreeNode<T>[]): PermissionTreeNode<T>[] => nodes
  .map(node => ({ ...node, children: prunePermissionTree(node.children) }))
  .filter(node => node.permissions.length > 0 || node.children.length > 0)

export const buildPermissionTree = <T extends PermissionTreeItem>(permissions: T[], template: PermissionTreeTemplateNode<T>[] = permissionTreeTemplate as PermissionTreeTemplateNode<T>[]) => {
  const roots = template.map(item => buildNodeFromTemplate(item))
  const fallbackNode: PermissionTreeNode<T> = { key: 'others', label: '其他权限', children: [], permissions: [] }
  permissions.forEach((permission) => {
    let matched = false
    for (let i = 0; i < template.length; i += 1) {
      const targetNode = findBestMatchedNode(permission, template[i], roots[i])
      if (targetNode) {
        targetNode.permissions.push(permission)
        matched = true
        break
      }
    }
    if (!matched) fallbackNode.permissions.push(permission)
  })
  const finalRoots = prunePermissionTree(roots)
  if (fallbackNode.permissions.length > 0) finalRoots.push(fallbackNode)
  sortPermissionTree(finalRoots)
  return finalRoots
}

export const collectNodeKeys = <T extends PermissionTreeItem>(nodes: PermissionTreeNode<T>[]) => {
  const keys: string[] = []
  const walk = (items: PermissionTreeNode<T>[]) => items.forEach((item) => { keys.push(item.key); if (item.children.length) walk(item.children) })
  walk(nodes)
  return keys
}

export const collectNodePermissionIds = <T extends PermissionTreeItem>(node: PermissionTreeNode<T>) => {
  const ids = node.permissions.map(item => item.id)
  node.children.forEach(child => ids.push(...collectNodePermissionIds(child)))
  return ids
}

export const matchPermissionKeyword = <T extends PermissionTreeItem>(permission: T, keyword: string) => {
  const normalized = keyword.trim().toLowerCase()
  return permission.name.toLowerCase().includes(normalized) || permission.code.toLowerCase().includes(normalized) || (permission.path || '').toLowerCase().includes(normalized)
}

export const matchPermissionNode = <T extends PermissionTreeItem>(node: PermissionTreeNode<T>, keyword: string): boolean => {
  const normalized = keyword.trim().toLowerCase()
  return node.label.toLowerCase().includes(normalized) || node.permissions.some(permission => matchPermissionKeyword(permission, normalized)) || node.children.some(child => matchPermissionNode(child, normalized))
}

export const flattenPermissionTreeRows = <T extends PermissionTreeItem>(params: {
  nodes: PermissionTreeNode<T>[]
  keyword?: string
  expandedKeys?: string[]
  selectedIds?: number[]
  selectedOnly?: boolean
  rootKey?: string
}) => {
  const keyword = params.keyword?.trim().toLowerCase() || ''
  const hasKeyword = keyword.length > 0
  const expandedKeySet = new Set(params.expandedKeys || [])
  const selectedIdSet = new Set(params.selectedIds || [])
  const rows: PermissionTreeRow<T>[] = []
  const walk = (nodes: PermissionTreeNode<T>[], depth: number) => {
    nodes.forEach((node) => {
      if (hasKeyword && !matchPermissionNode(node, keyword)) return
      const nodeLabelMatched = hasKeyword && node.label.toLowerCase().includes(keyword)
      const visiblePermissions = node.permissions.filter((permission) => {
        if (params.selectedOnly && !selectedIdSet.has(permission.id)) return false
        if (!hasKeyword) return true
        return nodeLabelMatched || matchPermissionKeyword(permission, keyword)
      })
      const beforeLength = rows.length
      rows.push({ type: 'node', key: `node:${node.key}`, depth, node })
      const shouldExpand = hasKeyword || params.selectedOnly || expandedKeySet.has(node.key)
      if (shouldExpand) {
        visiblePermissions.forEach(permission => rows.push({ type: 'permission', key: `permission:${permission.id}`, depth: depth + 1, permission }))
        walk(node.children, depth + 1)
      }
      if (params.selectedOnly && rows.length === beforeLength + 1 && visiblePermissions.length === 0) rows.splice(beforeLength, 1)
    })
  }
  const sourceNodes = params.rootKey ? params.nodes.filter(node => node.key === params.rootKey) : params.nodes
  walk(sourceNodes.length > 0 ? sourceNodes : params.nodes, 0)
  return rows
}
