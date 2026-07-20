import { describe, expect, it } from 'vitest'
import type { NavigationItem } from '@/types/app'
import { filterMenuItems, findActiveParentTitles, findNavigationTrail, getAccessibleMenuPaths, getRouteBreadcrumbLabels, getRoutePageTitle, isMenuItemActive } from './navigation'

const items: NavigationItem[] = [
  { title: '数据统计', path: '/', requiredPermissions: ['dashboard:view'] },
  {
    title: '系统管理',
    children: [
      { title: '用户管理', path: '/system/users', requiredPermissions: ['system:rbac:user'] },
      { title: '角色管理', path: '/system/roles', requiredPermissions: ['system:rbac:role'] },
      { title: '系统通知', path: '/system/messages', requiredPermissions: ['message:system:view'] },
      { title: '系统设置', path: '/system/settings', exact: true, requiredPermissions: ['system:settings:view'] }
    ]
  }
]

describe('navigation helpers', () => {
  it('filters items by permissions and removes empty parents', () => {
    const filtered = filterMenuItems(items, permissions => permissions.includes('system:rbac:user'))
    expect(filtered.map(item => item.title)).toEqual(['系统管理'])
    expect(filtered[0].children?.map(item => item.title)).toEqual(['用户管理'])
  })

  it('collects accessible route paths from filtered navigation', () => {
    const filtered = filterMenuItems(items, permissions => permissions.includes('dashboard:view') || permissions.includes('system:rbac:role'))
    expect(getAccessibleMenuPaths(filtered)).toEqual(['/', '/system/roles'])
  })

  it('matches dashboard exactly and nested routes by prefix', () => {
    expect(isMenuItemActive(items[0], '/templates')).toBe(false)
    expect(isMenuItemActive(items[0], '/')).toBe(true)
    expect(isMenuItemActive(items[1], '/system/users/detail')).toBe(true)
  })

  it('keeps system notification active as an independent management route', () => {
    const systemMenu = items[1].children || []
    const messageItem = systemMenu.find(item => item.title === '系统通知')!
    const settingsItem = systemMenu.find(item => item.title === '系统设置')!

    expect(isMenuItemActive(messageItem, '/system/messages')).toBe(true)
    expect(isMenuItemActive(settingsItem, '/system/messages')).toBe(false)
  })

  it('finds active parent titles', () => {
    expect(findActiveParentTitles(items, '/system/users')).toEqual(['系统管理'])
  })

  it('builds route trail, breadcrumb labels, and page title', () => {
    expect(findNavigationTrail(items, '/system/users/detail').map(item => item.title)).toEqual(['系统管理', '用户管理'])
    expect(getRouteBreadcrumbLabels(items, '/system/users/detail')).toEqual(['系统管理', '用户管理'])
    expect(getRoutePageTitle(items, '/system/users/detail')).toBe('用户管理')
    expect(getRouteBreadcrumbLabels(items, '/missing', ['兜底'])).toEqual(['兜底'])
  })
})
