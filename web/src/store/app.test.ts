import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from './app'
import { createDoraThemeTokens } from '@/util/glassTheme'

describe('app layout store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('updates viewport layout state across breakpoints', () => {
    const store = useAppStore()
    store.setViewportSize(760, 800)
    expect(store.navigation.layout.device).toBe('mobile')
    expect(store.navigation.layout.mobileSidebarVisible).toBe(false)
    store.openMobileSidebar()
    expect(store.navigation.layout.mobileSidebarVisible).toBe(true)
    store.setViewportSize(900, 800)
    expect(store.navigation.layout.device).toBe('desktop')
    expect(store.navigation.layout.sidebarCollapsed).toBe(true)
    store.setViewportSize(1200, 800)
    expect(store.navigation.layout.viewportClass).toBe('wide-desktop')
  })

  it('keeps light and dark Dora material expectations stable', () => {
    expect(createDoraThemeTokens('light').materials.base.background).toBe('#ffffff')
    expect(createDoraThemeTokens('dark').materials.base.background).toBe('#1c1c1c')
    expect(createDoraThemeTokens('dark').materials.overlay.blur).toBe('0px')
  })
})
