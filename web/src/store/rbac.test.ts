import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useRbacStore } from './rbac'
import { filterMenuItems, getAccessibleMenuPaths, menuItems } from '@/util/navigation'

describe('rbac navigation visibility', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('hides empty parent groups when user lacks child permissions', () => {
    const store = useRbacStore()
    store.permissions = ['dashboard:view']
    const filtered = filterMenuItems(menuItems, store.hasAnyPermission)
    expect(filtered.map(item => item.title)).toEqual(['数据统计'])
    expect(getAccessibleMenuPaths(filtered)).toEqual(['/'])
  })

  it('shows allowed children while preserving active parent context', () => {
    const store = useRbacStore()
    store.permissions = ['system:rbac:user', 'system:rbac:role']
    const filtered = filterMenuItems(menuItems, store.hasAnyPermission)
    const system = filtered.find(item => item.title === '系统管理')
    expect(system?.children?.map(item => item.title)).toEqual(['用户管理', '角色管理'])
    expect(getAccessibleMenuPaths(filtered)).toContain('/system/users')
  })

  it('allows super admin to see all permission-gated navigation entries', () => {
    const store = useRbacStore()
    store.isSuperAdmin = true
    const filtered = filterMenuItems(menuItems, store.hasAnyPermission)
    expect(getAccessibleMenuPaths(filtered)).toContain('/system/settings')
    expect(getAccessibleMenuPaths(filtered)).toContain('/logs/consume')
  })

  it('keeps permission-filtered visual shell routes stable for refreshed list pages', () => {
    const store = useRbacStore()
    store.permissions = ['message:sendlogs:view', 'data:consume-log:view', 'data:mq-source:view']
    const filtered = filterMenuItems(menuItems, store.hasAnyPermission)
    expect(getAccessibleMenuPaths(filtered)).toEqual(['/data/mq-sources', '/logs/task', '/logs/consume'])
  })
})
