import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from '@/store/app'

describe('responsive layout composable store contract', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('keeps desktop content offset aligned with shell sidebar widths', () => {
    const store = useAppStore()
    store.setViewportSize(1280, 800)
    expect(store.navigation.layout.device).toBe('desktop')
    expect(store.navigation.layout.sidebarCollapsed).toBe(false)
    store.setSidebarCollapsed(true)
    expect(store.navigation.layout.sidebarCollapsed).toBe(true)
  })

  it('uses temporary mobile sidebar state at narrow widths', () => {
    const store = useAppStore()
    store.setViewportSize(640, 800)
    expect(store.navigation.layout.device).toBe('mobile')
    expect(store.navigation.layout.mobileSidebarVisible).toBe(false)
    store.openMobileSidebar()
    expect(store.navigation.layout.mobileSidebarVisible).toBe(true)
    store.closeMobileSidebar()
    expect(store.navigation.layout.mobileSidebarVisible).toBe(false)
    expect(store.navigation.layout.sidebarOpened).toBe(false)
  })

  it('keeps mobile resize synchronization predictable', () => {
    const store = useAppStore()
    store.setViewportSize(640, 800)
    store.openMobileSidebar()
    store.setViewportSize(650, 800)
    expect(store.navigation.layout.mobileSidebarVisible).toBe(true)
    store.setViewportSize(1024, 800)
    expect(store.navigation.layout.mobileSidebarVisible).toBe(false)
    expect(store.navigation.layout.sidebarOpened).toBe(true)
  })

  it('keeps compact desktop in collapsed glass shell mode', () => {
    const store = useAppStore()
    store.setViewportSize(900, 800)
    expect(store.navigation.layout.viewportClass).toBe('compact-desktop')
    expect(store.navigation.layout.sidebarCollapsed).toBe(true)
    expect(store.navigation.layout.sidebarOpened).toBe(true)
  })

  it('restores glass navigation availability when moving from mobile overlay to desktop', () => {
    const store = useAppStore()
    store.setViewportSize(700, 800)
    store.openMobileSidebar()
    expect(store.navigation.layout.sidebarOpened).toBe(true)
    store.setViewportSize(1180, 800)
    expect(store.navigation.layout.device).toBe('desktop')
    expect(store.navigation.layout.mobileSidebarVisible).toBe(false)
    expect(store.navigation.layout.sidebarOpened).toBe(true)
    expect(store.navigation.layout.sidebarCollapsed).toBe(false)
  })
})
