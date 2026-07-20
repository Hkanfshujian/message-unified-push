import { describe, expect, it } from 'vitest'
import { canCloseTab, closeAllClosableTabs, closeOtherTabs, closeTabByPath, closeTabsToSide, createDefaultTabs, ensureTabForPath, getBasePath, getClosableOtherTabsCount, getClosableTabsToSideCount, normalizeRouteTab, normalizeTabs, resolveSafeActiveTabPath, sanitizeTabsForRoutes, type RouteTabMap } from './tabs'

const routeTabMap: RouteTabMap = {
  '/': { title: '数据统计', path: '/', closable: false, fixed: true },
  '/templates': { title: '模板管理', path: '/templates', closable: true },
  '/logs/task': { title: '任务日志', path: '/logs/task', closable: true },
  '/system/users': { title: '用户管理', path: '/system/users', closable: true },
  '/system/settings/site': { title: '系统设置', path: '/system/settings/site', closable: true },
  '/system/messages': { title: '系统通知', path: '/system/messages', closable: true }
}

describe('tab helpers', () => {
  it('gets canonical base paths', () => {
    expect(getBasePath('/')).toBe('/')
    expect(getBasePath('/system/users/detail')).toBe('/system/users')
    expect(getBasePath('/templates?x=1')).toBe('/templates')
  })

  it('normalizes stale tabs and keeps fixed dashboard', () => {
    const tabs = normalizeTabs([
      { title: 'old', path: '/missing', closable: true },
      { title: '模板', path: '/templates', closable: true }
    ], routeTabMap)
    expect(tabs.map(tab => tab.path)).toEqual(['/', '/templates'])
  })

  it('sanitizes stale tabs and resolves safe active path', () => {
    const tabs = sanitizeTabsForRoutes([
      { title: 'old', path: '/missing', closable: true },
      { title: '用户', path: '/system/users', closable: true }
    ], routeTabMap)
    expect(tabs.map(tab => tab.path)).toEqual(['/', '/system/users'])
    expect(resolveSafeActiveTabPath(tabs, '/missing')).toBe('/')
  })

  it('ensures tab for route', () => {
    const result = ensureTabForPath(createDefaultTabs(), '/templates/detail', routeTabMap)
    expect(result.activePath).toBe('/templates')
    expect(result.tabs.some(tab => tab.path === '/templates')).toBe(true)
  })

  it('opens system notification as its own management tab', () => {
    const result = ensureTabForPath(createDefaultTabs(), '/system/messages', routeTabMap)
    expect(result.activePath).toBe('/system/messages')
    expect(result.tabs.some(tab => tab.path === '/system/messages')).toBe(true)
    expect(result.tabs.some(tab => tab.path === '/system/settings')).toBe(false)
  })

  it('closes current tab and selects neighbor', () => {
    const result = closeTabByPath([
      { title: '数据统计', path: '/', closable: false, fixed: true },
      { title: '模板', path: '/templates', closable: true },
      { title: '任务日志', path: '/logs/task', closable: true }
    ], '/templates', '/templates')
    expect(result.activePath).toBe('/logs/task')
    expect(result.tabs.map(tab => tab.path)).toEqual(['/', '/logs/task'])
  })

  it('closes other, left, right, and all closable tabs', () => {
    const tabs = [
      { title: '数据统计', path: '/', closable: false, fixed: true },
      { title: '模板', path: '/templates', closable: true },
      { title: '任务日志', path: '/logs/task', closable: true },
      { title: '用户', path: '/system/users', closable: true }
    ]
    expect(closeOtherTabs(tabs, '/logs/task').tabs.map(tab => tab.path)).toEqual(['/', '/logs/task'])
    expect(closeTabsToSide(tabs, '/logs/task', 'left').tabs.map(tab => tab.path)).toEqual(['/', '/logs/task', '/system/users'])
    expect(closeTabsToSide(tabs, '/logs/task', 'right').tabs.map(tab => tab.path)).toEqual(['/', '/templates', '/logs/task'])
    expect(closeAllClosableTabs(tabs).tabs.map(tab => tab.path)).toEqual(['/'])
    expect(canCloseTab(tabs, '/')).toBe(false)
    expect(canCloseTab(tabs, '/templates')).toBe(true)
    expect(getClosableTabsToSideCount(tabs, '/logs/task', 'left')).toBe(1)
    expect(getClosableOtherTabsCount(tabs, '/templates')).toBe(2)
    expect(closeTabsToSide(tabs, '/missing', 'left').activePath).toBe('/')
  })

  it('normalizes redirected route tabs', () => {
    expect(normalizeRouteTab('/sendlogs', routeTabMap, { '/sendlogs': '/logs/task' })?.path).toBe('/logs/task')
    expect(normalizeRouteTab('/missing', routeTabMap)).toBeNull()
  })
})
