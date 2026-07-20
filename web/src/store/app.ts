import { defineStore } from 'pinia'
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { classifyViewport } from '@/util/layout'
import type { LayoutDevice, LayoutViewportClass, NavigationState, OpenTabItem, SiteConfigState } from '@/types/app'

interface AppStoreState {
  siteConfig: SiteConfigState
  navigation: NavigationState
}

export const useAppStore = defineStore('app', {
  state: (): AppStoreState => ({
    siteConfig: LocalStieConfigUtils.getLocalConfig() as SiteConfigState,
    navigation: {
      activePath: '/',
      openTabs: [],
      sidebarCollapsed: false,
      layout: {
        device: 'desktop',
        viewportWidth: 0,
        viewportHeight: 0,
        viewportClass: 'wide-desktop',
        sidebarOpened: true,
        sidebarCollapsed: false,
        mobileSidebarVisible: false,
        lastUserCollapseIntent: null
      }
    } as NavigationState
  }),
  actions: {
    setSiteConfig(config: SiteConfigState) {
      this.siteConfig = config
    },
    setSidebarCollapsed(collapsed: boolean) {
      this.navigation.sidebarCollapsed = collapsed
      this.navigation.layout.sidebarCollapsed = collapsed
      this.navigation.layout.lastUserCollapseIntent = collapsed
      if (this.navigation.layout.device === 'desktop') {
        this.navigation.layout.sidebarOpened = true
      }
    },
    setViewportSize(width: number, height: number) {
      const viewportClass = classifyViewport(width)
      const previousDevice = this.navigation.layout.device
      const nextDevice: LayoutDevice = viewportClass === 'mobile' ? 'mobile' : 'desktop'
      this.navigation.layout.viewportWidth = width
      this.navigation.layout.viewportHeight = height
      this.navigation.layout.viewportClass = viewportClass
      this.navigation.layout.device = nextDevice
      if (nextDevice === 'mobile') {
        this.navigation.layout.sidebarOpened = false
        if (previousDevice !== 'mobile') {
          this.navigation.layout.mobileSidebarVisible = false
        }
      } else {
        this.navigation.layout.sidebarOpened = true
        this.navigation.layout.mobileSidebarVisible = false
        if (viewportClass === 'compact-desktop') {
          this.navigation.layout.sidebarCollapsed = true
          this.navigation.sidebarCollapsed = true
        } else if (previousDevice === 'mobile' && this.navigation.layout.lastUserCollapseIntent === null) {
          this.navigation.layout.sidebarCollapsed = false
          this.navigation.sidebarCollapsed = false
        } else if (this.navigation.layout.lastUserCollapseIntent !== null) {
          this.navigation.layout.sidebarCollapsed = this.navigation.layout.lastUserCollapseIntent
          this.navigation.sidebarCollapsed = this.navigation.layout.lastUserCollapseIntent
        }
      }
    },
    setDevice(device: LayoutDevice) {
      this.navigation.layout.device = device
    },
    setViewportClass(viewportClass: LayoutViewportClass) {
      this.navigation.layout.viewportClass = viewportClass
    },
    setMobileSidebarVisible(visible: boolean) {
      this.navigation.layout.mobileSidebarVisible = visible
      if (this.navigation.layout.device === 'mobile') {
        this.navigation.layout.sidebarOpened = visible
      }
    },
    openMobileSidebar() {
      this.setMobileSidebarVisible(true)
    },
    closeMobileSidebar() {
      this.setMobileSidebarVisible(false)
    },
    setActivePath(path: string) {
      this.navigation.activePath = path
    },
    setOpenTabs(tabs: OpenTabItem[]) {
      this.navigation.openTabs = tabs
    },
    clearNavigation() {
      this.navigation.openTabs = []
      this.navigation.activePath = '/'
      this.navigation.layout.mobileSidebarVisible = false
    }
  }
})
