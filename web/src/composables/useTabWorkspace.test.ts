import { describe, expect, it } from 'vitest'
import { closeAllClosableTabs, closeOtherTabs, closeTabByPath, createDefaultTabs, ensureTabForPath, resolveSafeActiveTabPath, sanitizeTabsForRoutes, type RouteTabMap } from '@/util/tabs'

const routeTabMap: RouteTabMap = {
  '/': { title: '数据统计', path: '/', closable: false, fixed: true, icon: 'chart' },
  '/templates': { title: '模板管理', path: '/templates', closable: true, icon: 'document' },
  '/logs/task': { title: '任务日志', path: '/logs/task', closable: true, icon: 'search' }
}

describe('tab workspace composable helper contract', () => {
  it('restores direct nested routes to canonical workspace tabs', () => {
    const result = ensureTabForPath(createDefaultTabs(), '/templates/detail/1', routeTabMap)
    expect(result.activePath).toBe('/templates')
    expect(result.tabs.map(tab => tab.path)).toEqual(['/', '/templates'])
    expect(result.tabs.find(tab => tab.path === '/templates')?.icon).toBe('document')
  })

  it('does not require replacing the browser route when nested route maps to a canonical tab', () => {
    const result = ensureTabForPath(createDefaultTabs(), '/templates/detail/1', routeTabMap)
    expect(result.activePath).not.toBe('/templates/detail/1')
    expect(result.activePath).toBe('/templates')
  })

  it('removes stale tabs and keeps a safe default tab', () => {
    const tabs = sanitizeTabsForRoutes([
      { title: '旧页面', path: '/old', closable: true },
      { title: '任务日志', path: '/logs/task', closable: true }
    ], routeTabMap)
    expect(tabs.map(tab => tab.path)).toEqual(['/', '/logs/task'])
    expect(tabs.map(tab => tab.icon)).toEqual(['chart', 'search'])
    expect(resolveSafeActiveTabPath(tabs, '/old')).toBe('/')
  })

  it('protects fixed tabs during close actions', () => {
    const tabs = [
      { title: '数据统计', path: '/', closable: false, fixed: true },
      { title: '模板管理', path: '/templates', closable: true },
      { title: '任务日志', path: '/logs/task', closable: true }
    ]
    expect(closeOtherTabs(tabs, '/templates').tabs.map(tab => tab.path)).toEqual(['/', '/templates'])
    expect(closeAllClosableTabs(tabs).tabs.map(tab => tab.path)).toEqual(['/'])
    expect(closeTabByPath(tabs, '/templates', '/templates').activePath).toBe('/logs/task')
  })

  it('keeps visual refresh stable when closing stale and refreshed workspace tabs', () => {
    const tabs = [
      { title: '数据统计', path: '/', closable: false, fixed: true },
      { title: '旧页面', path: '/legacy', closable: true },
      { title: '任务日志', path: '/logs/task', closable: true }
    ]
    const sanitized = sanitizeTabsForRoutes(tabs, routeTabMap)
    expect(sanitized.map(tab => tab.path)).toEqual(['/', '/logs/task'])
    expect(closeAllClosableTabs(sanitized).activePath).toBe('/')
  })
})
