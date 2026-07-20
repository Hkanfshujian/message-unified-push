import type { NavigationItem } from '@/types/app'

export const menuItems: NavigationItem[] = [
  { title: '数据统计', path: '/', iconName: 'chart', requiredPermissions: ['dashboard:view'] },
  {
    title: '消息管理',
    iconName: 'message',
    children: [
      { title: '定时消息', path: '/cronmessages', name: 'cronmessages', iconName: 'calendar', requiredPermissions: ['message:cron:view'] },
      { title: '订阅消息', path: '/message/subscriptions', name: 'message-subscriptions', iconName: 'notification', requiredPermissions: ['data:subscription:view'] }
    ]
  },
  { title: '模板管理', path: '/templates', iconName: 'document', requiredPermissions: ['message:template:view'] },
  { title: '渠道管理', path: '/sendways', iconName: 'activity', requiredPermissions: ['message:sendways:view'] },
  {
    title: '数据管理',
    iconName: 'database',
    requiredPermissions: ['data:mq-source:view'],
    children: [
      { title: '消息队列', path: '/data/mq-sources', name: 'data-mq-sources', iconName: 'app', requiredPermissions: ['data:mq-source:view'] }
    ]
  },
  {
    title: '日志管理',
    iconName: 'document',
    requiredPermissions: ['message:sendlogs:view'],
    children: [
      { title: '任务日志', path: '/logs/task', iconName: 'search', requiredPermissions: ['message:sendlogs:view'] },
      { title: '登录日志', path: '/logs/login', iconName: 'login', requiredPermissions: ['system:loginlogs:view'] },
      { title: '消费日志', path: '/logs/consume', iconName: 'document', requiredPermissions: ['data:consume-log:view'] }
    ]
  },
  {
    title: '系统管理',
    iconName: 'security',
    children: [
      { title: '用户管理', path: '/system/users', name: 'system-users', iconName: 'user', requiredPermissions: ['system:rbac:user'] },
      { title: '用户组管理', path: '/system/groups', name: 'system-groups', iconName: 'team', requiredPermissions: ['system:rbac:group'] },
      { title: '角色管理', path: '/system/roles', name: 'system-roles', iconName: 'security', requiredPermissions: ['system:rbac:role'] },
      { title: '权限管理', path: '/system/permissions', name: 'system-permissions', iconName: 'key', requiredPermissions: ['system:rbac:permission'] },
      { title: '系统通知', path: '/system/messages', name: 'system-messages', iconName: 'notification', requiredPermissions: ['message:system:view'] },
      { title: '系统设置', path: '/system/settings', name: 'system-settings-site', exact: true, iconName: 'setting', requiredPermissions: ['system:settings:view'] }
    ]
  }
]

export const canAccessMenuItem = (item: NavigationItem, hasAnyPermission: (permissions: string[]) => boolean) => {
  if (!item.requiredPermissions || item.requiredPermissions.length === 0) return true
  return hasAnyPermission(item.requiredPermissions)
}

export const filterMenuItems = (items: NavigationItem[], hasAnyPermission: (permissions: string[]) => boolean): NavigationItem[] => {
  return items
    .map((item) => {
      const children = item.children ? filterMenuItems(item.children, hasAnyPermission) : undefined
      const accessible = canAccessMenuItem(item, hasAnyPermission)
      if (children?.length) return { ...item, children }
      if (accessible && !item.children) return { ...item }
      if (accessible && item.path) return { ...item, children: undefined }
      return null
    })
    .filter((item): item is NavigationItem => Boolean(item))
}

export const getAccessibleMenuPaths = (items: NavigationItem[]): string[] => {
  const paths: string[] = []
  const collect = (nodes: NavigationItem[]) => {
    nodes.forEach((item) => {
      if (item.path) paths.push(item.path)
      if (item.children?.length) collect(item.children)
    })
  }
  collect(items)
  return paths
}

export const isMenuItemActive = (item: NavigationItem, currentPath: string): boolean => {
  if (item.path) {
    if (item.path === '/') return currentPath === '/'
    if (item.exact) return currentPath === item.path
    return currentPath === item.path || currentPath.startsWith(`${item.path}/`)
  }
  return !!item.children?.some(child => isMenuItemActive(child, currentPath))
}

export const findActiveParentTitles = (items: NavigationItem[], currentPath: string): string[] => {
  const result: string[] = []
  for (const item of items) {
    if (item.children?.some(child => isMenuItemActive(child, currentPath))) result.push(item.title)
    result.push(...findActiveParentTitles(item.children || [], currentPath))
  }
  return result
}

export const findNavigationTrail = (items: NavigationItem[], currentPath: string): NavigationItem[] => {
  for (const item of items) {
    if (item.path && isMenuItemActive(item, currentPath)) return [item]
    const childTrail = findNavigationTrail(item.children || [], currentPath)
    if (childTrail.length) return [item, ...childTrail]
  }
  return []
}

export const getRouteBreadcrumbLabels = (items: NavigationItem[], currentPath: string, fallback: string[] = ['数据统计']) => {
  const labels = findNavigationTrail(items, currentPath).map(item => item.title)
  return labels.length ? labels : fallback
}

export const getRoutePageTitle = (items: NavigationItem[], currentPath: string, fallback = '数据统计') => {
  const labels = getRouteBreadcrumbLabels(items, currentPath, [fallback])
  return labels[labels.length - 1] || fallback
}
