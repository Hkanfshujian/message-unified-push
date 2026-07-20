import { computed, ref, watch, type Ref } from 'vue'
import type { RouteLocationNormalizedLoadedGeneric, Router } from 'vue-router'
import { useAppStore } from '@/store/app'
import type { OpenTabItem } from '@/types/app'
import { closeAllClosableTabs, closeOtherTabs as closeOtherTabItems, closeTabByPath, closeTabsToSide, createDefaultTabs, ensureTabForPath, normalizeTabs, resolveSafeActiveTabPath, sanitizeTabsForRoutes, type RouteTabMap } from '@/util/tabs'

export interface TabContextMenuState {
  visible: boolean
  x: number
  y: number
  path: string | null
}

export const useTabWorkspace = (options: {
  route: RouteLocationNormalizedLoadedGeneric
  router: Router
  routeTabMap: RouteTabMap
  storageKey: string
}) => {
  const appStore = useAppStore()
  const refreshKey = ref(0)
  const refreshingPath = ref<string | null>(null)
  const tabContextMenu = ref<TabContextMenuState>({ visible: false, x: 0, y: 0, path: null })

  const loadTabsFromStorage = (): OpenTabItem[] => {
    try {
      const raw = localStorage.getItem(options.storageKey)
      if (!raw) return createDefaultTabs()
      const parsed = JSON.parse(raw)
      if (!Array.isArray(parsed)) return createDefaultTabs()
      return normalizeTabs(parsed, options.routeTabMap)
    } catch {
      return createDefaultTabs()
    }
  }

  const tabs = ref<OpenTabItem[]>(loadTabsFromStorage()) as Ref<OpenTabItem[]>
  const activeTabPath = ref(options.route.path)

  const persistTabs = () => {
    appStore.setOpenTabs(tabs.value)
    try {
      localStorage.setItem(options.storageKey, JSON.stringify(tabs.value))
    } catch { }
  }

  const clearTabsCache = () => {
    try {
      localStorage.removeItem(options.storageKey)
    } catch { }
  }

  const resetTabsToDefault = () => {
    tabs.value = createDefaultTabs()
    activeTabPath.value = '/'
    clearTabsCache()
  }

  const ensureTabForRoute = (path: string) => {
    const result = ensureTabForPath(tabs.value, path, options.routeTabMap)
    tabs.value = result.tabs
    activeTabPath.value = result.activePath
  }

  const sanitizeWorkspaceTabs = () => {
    tabs.value = sanitizeTabsForRoutes(tabs.value, options.routeTabMap)
    activeTabPath.value = resolveSafeActiveTabPath(tabs.value, activeTabPath.value)
  }

  const activateTab = (tab: OpenTabItem) => {
    if (tab.path !== options.route.path) options.router.push(tab.path)
  }

  const closeTab = (tab: OpenTabItem) => {
    const result = closeTabByPath(tabs.value, activeTabPath.value, tab.path)
    tabs.value = result.tabs
    if (activeTabPath.value !== result.activePath) options.router.push(result.activePath)
    activeTabPath.value = result.activePath
  }

  const hideTabContextMenu = () => {
    tabContextMenu.value.visible = false
  }

  const openTabContextMenu = (event: MouseEvent | KeyboardEvent, tab: OpenTabItem) => {
    event.preventDefault()
    const target = event.currentTarget as HTMLElement | null
    const rect = target?.getBoundingClientRect()
    const x = event instanceof MouseEvent ? event.clientX : Math.round(rect?.left || 0)
    const y = event instanceof MouseEvent ? event.clientY : Math.round(rect?.bottom || 0)
    tabContextMenu.value = { visible: true, x, y, path: tab.path }
  }

  const closeOtherTabs = () => {
    const currentPath = tabContextMenu.value.path || activeTabPath.value
    const result = closeOtherTabItems(tabs.value, currentPath)
    tabs.value = result.tabs
    if (options.route.path !== result.activePath) options.router.push(result.activePath)
    activeTabPath.value = result.activePath
    hideTabContextMenu()
  }

  const closeLeftTabs = () => {
    const currentPath = tabContextMenu.value.path || activeTabPath.value
    const result = closeTabsToSide(tabs.value, currentPath, 'left')
    tabs.value = result.tabs
    if (options.route.path !== result.activePath) options.router.push(result.activePath)
    activeTabPath.value = result.activePath
    hideTabContextMenu()
  }

  const closeRightTabs = () => {
    const currentPath = tabContextMenu.value.path || activeTabPath.value
    const result = closeTabsToSide(tabs.value, currentPath, 'right')
    tabs.value = result.tabs
    if (options.route.path !== result.activePath) options.router.push(result.activePath)
    activeTabPath.value = result.activePath
    hideTabContextMenu()
  }

  const closeAllTabs = () => {
    const result = closeAllClosableTabs(tabs.value)
    tabs.value = result.tabs
    activeTabPath.value = result.activePath
    options.router.push(result.activePath)
    hideTabContextMenu()
  }

  const refreshActiveTab = () => {
    refreshingPath.value = activeTabPath.value
    refreshKey.value += 1
    hideTabContextMenu()
    window.setTimeout(() => {
      refreshingPath.value = null
    }, 380)
  }

  watch(tabs, persistTabs, { deep: true })

  const routerViewKey = computed(() => `${options.route.fullPath}:${refreshKey.value}`)

  return {
    tabs,
    activeTabPath,
    refreshingPath,
    tabContextMenu,
    routerViewKey,
    ensureTabForRoute,
    activateTab,
    closeTab,
    openTabContextMenu,
    hideTabContextMenu,
    closeOtherTabs,
    closeLeftTabs,
    closeRightTabs,
    closeAllTabs,
    refreshActiveTab,
    resetTabsToDefault,
    clearTabsCache,
    sanitizeWorkspaceTabs
  }
}
