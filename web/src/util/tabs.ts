import type { OpenTabItem } from '@/types/app'

export type RouteTabMap = Record<string, OpenTabItem>

export const getBasePath = (path: string) => {
  if (path === '/') return '/'
  const pure = path.split('?')[0]
  const parts = pure.split('/').filter(Boolean)
  if (parts.length >= 2) return `/${parts[0]}/${parts[1]}`
  if (parts.length === 1) return `/${parts[0]}`
  return '/'
}

export const createDefaultTabs = (): OpenTabItem[] => [
  { title: '数据统计', path: '/', closable: false, fixed: true, icon: 'chart' }
]

export const findRouteTabMeta = (path: string, routeTabMap: RouteTabMap) => {
  if (routeTabMap[path]) return routeTabMap[path]
  const matchedPath = Object.keys(routeTabMap)
    .filter(routePath => routePath !== '/' && (path === routePath || path.startsWith(`${routePath}/`)))
    .sort((a, b) => b.length - a.length)[0]
  if (matchedPath) return routeTabMap[matchedPath]
  const base = getBasePath(path)
  return routeTabMap[base]
}

export const normalizeTabs = (rawTabs: OpenTabItem[], routeTabMap: RouteTabMap) => {
  const unique = new Map<string, OpenTabItem>()
  createDefaultTabs().forEach(tab => unique.set(tab.path, { ...tab }))
  for (const tab of rawTabs) {
    const meta = findRouteTabMeta(tab.path, routeTabMap)
    if (!meta) continue
    unique.set(meta.path, { ...meta, fixed: meta.fixed || !meta.closable })
  }
  return Array.from(unique.values())
}

export const sanitizeTabsForRoutes = (tabs: OpenTabItem[], routeTabMap: RouteTabMap) => normalizeTabs(tabs, routeTabMap)

export const resolveSafeActiveTabPath = (tabs: OpenTabItem[], preferredPath: string) => {
  const matched = tabs.find(tab => tab.path === preferredPath)
  if (matched) return matched.path
  const fixed = tabs.find(tab => !tab.closable || tab.fixed)
  return fixed?.path || tabs[0]?.path || createDefaultTabs()[0].path
}

export const ensureTabForPath = (tabs: OpenTabItem[], path: string, routeTabMap: RouteTabMap) => {
  const meta = findRouteTabMeta(path, routeTabMap)
  if (!meta) return { tabs, activePath: resolveSafeActiveTabPath(tabs, path) }
  const nextTabs = tabs.some(tab => tab.path === meta.path) ? tabs : [...tabs, { ...meta }]
  return { tabs: nextTabs, activePath: meta.path }
}

export const closeTabByPath = (tabs: OpenTabItem[], activePath: string, path: string) => {
  const target = tabs.find(tab => tab.path === path)
  if (!target || !target.closable) return { tabs, activePath }
  const index = tabs.findIndex(tab => tab.path === path)
  const nextTabs = tabs.filter(tab => tab.path !== path)
  if (activePath !== path) return { tabs: nextTabs, activePath }
  const next = nextTabs[index] || nextTabs[index - 1] || nextTabs[0] || createDefaultTabs()[0]
  return { tabs: nextTabs.length ? nextTabs : [next], activePath: next.path }
}

export const closeOtherTabs = (tabs: OpenTabItem[], path: string) => {
  const nextTabs = tabs.filter(tab => !tab.closable || tab.path === path)
  const fallback = nextTabs.find(tab => tab.path === path) || nextTabs[0] || createDefaultTabs()[0]
  return { tabs: nextTabs.length ? nextTabs : [fallback], activePath: fallback.path }
}

export const closeTabsToSide = (tabs: OpenTabItem[], path: string, side: 'left' | 'right') => {
  const index = tabs.findIndex(tab => tab.path === path)
  if (index === -1) return { tabs, activePath: resolveSafeActiveTabPath(tabs, path) }
  const nextTabs = tabs.filter((tab, tabIndex) => {
    if (!tab.closable) return true
    if (tab.path === path) return true
    return side === 'left' ? tabIndex > index : tabIndex < index
  })
  const fallback = nextTabs.find(tab => tab.path === path) || nextTabs[0] || createDefaultTabs()[0]
  return { tabs: nextTabs.length ? nextTabs : [fallback], activePath: fallback.path }
}

export const canCloseTab = (tabs: OpenTabItem[], path: string) => {
  const tab = tabs.find(item => item.path === path)
  return Boolean(tab?.closable)
}

export const getClosableTabsToSideCount = (tabs: OpenTabItem[], path: string, side: 'left' | 'right') => {
  const index = tabs.findIndex(tab => tab.path === path)
  if (index === -1) return 0
  return tabs.filter((tab, tabIndex) => {
    if (!tab.closable || tab.path === path) return false
    return side === 'left' ? tabIndex < index : tabIndex > index
  }).length
}

export const getClosableOtherTabsCount = (tabs: OpenTabItem[], path: string) => {
  return tabs.filter(tab => tab.closable && tab.path !== path).length
}

export const closeAllClosableTabs = (tabs: OpenTabItem[]) => {
  const fixedTabs = tabs.filter(tab => !tab.closable || tab.fixed)
  const nextTabs = fixedTabs.length ? fixedTabs : createDefaultTabs()
  return { tabs: nextTabs, activePath: nextTabs[0].path }
}

export const normalizeRouteTab = (path: string, routeTabMap: RouteTabMap, redirectMap: Record<string, string> = {}) => {
  const redirected = redirectMap[path] || redirectMap[getBasePath(path)] || path
  return findRouteTabMeta(redirected, routeTabMap) || null
}
