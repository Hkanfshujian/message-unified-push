<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, watch, type ComponentPublicInstance } from 'vue'
import { CONSTANT } from '../constant.js'
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { usePageState } from '@/store/page_sate.js'
import { useAppStore, useRbacStore, useSessionStore } from '@/store'
import { useRoute, useRouter } from 'vue-router'
import { request } from '@/api/api'
import { statisticsApi } from '@/api/statistics'
import { usersApi } from '@/api/users'
import { confirmAction, notifyError, notifyInfo, notifySuccess } from '@/util/uiFeedback'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import { useResponsiveLayout } from '@/composables/useResponsiveLayout'
import { useTabWorkspace } from '@/composables/useTabWorkspace'
import { canCloseTab, findRouteTabMeta, getBasePath, getClosableOtherTabsCount, getClosableTabsToSideCount, type RouteTabMap } from '@/util/tabs'
import { useMessageCenterStore } from '@/store/message-center'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.appShell

const route = useRoute()
const router = useRouter()
const pageState = usePageState()
const appStore = useAppStore()
const rbacStore = useRbacStore()
const sessionStore = useSessionStore()
const messageCenterStore = useMessageCenterStore()
const isAuthenticated = computed(() => sessionStore.isLogin)
const userAccount = ref('管理员')
const siteConfig = ref<any>({})
const showTabBar = ref(true)
const tabScrollRef = ref<HTMLElement | null>(null)
const tabContextMenuRef = ref<HTMLElement | null>(null)
const tabItemRefs = new Map<string, HTMLElement>()
const canScrollTabsLeft = ref(false)
const canScrollTabsRight = ref(false)
const dashboardDateStart = ref('')
const dashboardDateEnd = ref('')
const isExporting = ref(false)

const TABS_STORAGE_KEY = CONSTANT.STORE_OPEN_TABS_NAME || 'message_nest_open_tabs_v1'

const routeTabMap: RouteTabMap = {
  '/': { title: '数据统计', path: '/', closable: false, fixed: true, icon: 'chart' },
  '/cronmessages': { title: '定时消息', path: '/cronmessages', closable: true, icon: 'calendar' },
  '/templates': { title: '模板管理', path: '/templates', closable: true, icon: 'document' },
  '/sendways': { title: '渠道管理', path: '/sendways', closable: true, icon: 'activity' },
  '/logs': { title: '日志管理', path: '/logs', closable: true, icon: 'document' },
  '/logs/task': { title: '任务日志', path: '/logs/task', closable: true, icon: 'search' },
  '/logs/login': { title: '登录日志', path: '/logs/login', closable: true, icon: 'login' },
  '/logs/consume': { title: '消费日志', path: '/logs/consume', closable: true, icon: 'document' },
  '/system/settings/site': { title: '系统设置', path: '/system/settings/site', closable: true, icon: 'setting' },
  '/system/messages': { title: '系统通知', path: '/system/messages', closable: true, icon: 'notification' },
  '/profile/settings': { title: '个人设置', path: '/profile/settings', closable: true, icon: 'user' },
  '/system/roles': { title: '角色管理', path: '/system/roles', closable: true, icon: 'security' },
  '/system/groups': { title: '用户组管理', path: '/system/groups', closable: true, icon: 'team' },
  '/system/permissions': { title: '权限管理', path: '/system/permissions', closable: true, icon: 'key' },
  '/system/users': { title: '用户管理', path: '/system/users', closable: true, icon: 'user' },
  '/system/relations': { title: '授权关系', path: '/system/relations', closable: true, icon: 'security' },
  '/data/mq-sources': { title: '消息队列', path: '/data/mq-sources', closable: true, icon: 'database' },
  '/message/subscriptions': { title: '订阅消息', path: '/message/subscriptions', closable: true, icon: 'notification' }
}

const {
  layoutState,
  isMobileLayout,
  isSidebarCollapsed,
  isMobileSidebarVisible,
  contentOffset,
  toggleSidebar,
  closeMobileSidebar
} = useResponsiveLayout()

const {
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
  resetTabsToDefault
} = useTabWorkspace({ route, router, routeTabMap, storageKey: TABS_STORAGE_KEY })

const contextMenuTargetPath = computed(() => tabContextMenu.value.path || activeTabPath.value)
const canCloseContextTab = computed(() => canCloseTab(tabs.value, contextMenuTargetPath.value))
const closableLeftCount = computed(() => getClosableTabsToSideCount(tabs.value, contextMenuTargetPath.value, 'left'))
const closableRightCount = computed(() => getClosableTabsToSideCount(tabs.value, contextMenuTargetPath.value, 'right'))
const closableOtherCount = computed(() => getClosableOtherTabsCount(tabs.value, contextMenuTargetPath.value))
const closableTabsCount = computed(() => tabs.value.filter(tab => tab.closable).length)

const updateTabOverflow = () => {
  const el = tabScrollRef.value
  if (!el) {
    canScrollTabsLeft.value = false
    canScrollTabsRight.value = false
    return
  }
  const maxScrollLeft = Math.max(el.scrollWidth - el.clientWidth, 0)
  canScrollTabsLeft.value = el.scrollLeft > 1
  canScrollTabsRight.value = el.scrollLeft < maxScrollLeft - 1
}

const scrollTabs = (direction: 'left' | 'right') => {
  const el = tabScrollRef.value
  if (!el) return
  el.scrollBy({ left: direction === 'left' ? -240 : 240, behavior: 'smooth' })
  window.setTimeout(updateTabOverflow, 260)
}

const setTabItemRef = (path: string, el: Element | ComponentPublicInstance | null) => {
  if (el instanceof HTMLElement) {
    tabItemRefs.set(path, el)
    return
  }
  tabItemRefs.delete(path)
}

const scrollActiveTabIntoView = () => {
  const activeTabEl = tabItemRefs.get(activeTabPath.value)
  if (!activeTabEl) return
  activeTabEl.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'smooth' })
  window.setTimeout(updateTabOverflow, 260)
}

const handleWindowResize = () => updateTabOverflow()

const focusAndActivateTab = async (index: number) => {
  const tab = tabs.value[index]
  if (!tab) return
  activateTab(tab)
  await nextTick()
  tabItemRefs.get(tab.path)?.focus()
}

const handleTabKeydown = (event: KeyboardEvent, index: number) => {
  let targetIndex = index
  if (event.key === 'ArrowRight') targetIndex = (index + 1) % tabs.value.length
  else if (event.key === 'ArrowLeft') targetIndex = (index - 1 + tabs.value.length) % tabs.value.length
  else if (event.key === 'Home') targetIndex = 0
  else if (event.key === 'End') targetIndex = tabs.value.length - 1
  else if (event.key === 'F10' && event.shiftKey) {
    openTabContextMenu(event, tabs.value[index])
    return
  } else return
  event.preventDefault()
  void focusAndActivateTab(targetIndex)
}

const closeMobileSidebarWithFocusRestore = async () => {
  if (!isMobileSidebarVisible.value) return
  closeMobileSidebar()
  await nextTick()
  document.getElementById('sidebar-toggle-button')?.focus()
}

const handleGlobalEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isMobileSidebarVisible.value) closeMobileSidebarWithFocusRestore()
}

watch(
  () => tabContextMenu.value.visible,
  async visible => {
    if (!visible) return
    await nextTick()
    tabContextMenuRef.value?.querySelector<HTMLElement>('button:not(:disabled)')?.focus()
  }
)

const closeWorkspaceTab = async (tab: (typeof tabs.value)[number]) => {
  closeTab(tab)
  await nextTick()
  tabItemRefs.get(activeTabPath.value)?.focus()
}

const closeAllWorkspaceTabs = async () => {
  closeAllTabs()
  await nextTick()
  tabItemRefs.get(activeTabPath.value)?.focus()
}

const closeContextTab = async () => {
  const target = tabs.value.find(tab => tab.path === contextMenuTargetPath.value)
  if (!target || !target.closable) return
  closeTab(target)
  hideTabContextMenu()
  await nextTick()
  tabItemRefs.get(activeTabPath.value)?.focus()
}

const routeBreadcrumbMap: Record<string, string[]> = {
  '/': ['数据统计'],
  '/cronmessages': ['消息管理', '定时消息'],
  '/message/subscriptions': ['消息管理', '订阅消息'],
  '/templates': ['模板管理'],
  '/sendways': ['渠道管理'],
  '/data/mq-sources': ['数据管理', '消息队列'],
  '/logs': ['日志管理'],
  '/logs/task': ['日志管理', '任务日志'],
  '/logs/login': ['日志管理', '登录日志'],
  '/logs/consume': ['日志管理', '消费日志'],
  '/system/users': ['系统管理', '用户管理'],
  '/system/groups': ['系统管理', '用户组管理'],
  '/system/roles': ['系统管理', '角色管理'],
  '/system/permissions': ['系统管理', '权限管理'],
  '/system/settings/site': ['系统管理', '系统设置'],
  '/system/messages': ['系统管理', '系统通知'],
  '/system/relations': ['系统管理', '授权关系'],
  '/profile/settings': ['个人设置']
}
// 主题：明暗模式与跟随系统
type ThemePreference = 'light' | 'dark' | 'system'

const getInitialThemePreference = (): ThemePreference => {
  try {
    const storedPref = localStorage.getItem('themePreference') as ThemePreference | null
    if (storedPref === 'light' || storedPref === 'dark' || storedPref === 'system') return storedPref
    const legacy = localStorage.getItem('theme') as 'light' | 'dark' | null
    if (legacy === 'light' || legacy === 'dark') return legacy
    return 'system'
  } catch {
    return 'system'
  }
}

const themePreference = ref<ThemePreference>(getInitialThemePreference())
const theme = ref<'light' | 'dark'>('light')

const applyColorModeFromPreference = () => {
  const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  const effective: 'light' | 'dark' = themePreference.value === 'system'
    ? (systemDark ? 'dark' : 'light')
    : themePreference.value
  theme.value = effective
  if (effective === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  try { localStorage.setItem('themePreference', themePreference.value) } catch { }
}

const toggleTheme = () => {
  if (themePreference.value === 'system') {
    themePreference.value = theme.value === 'dark' ? 'light' : 'dark'
  } else {
    themePreference.value = themePreference.value === 'dark' ? 'light' : 'dark'
  }
  applyColorModeFromPreference()
}

// 从JWT中解析用户名
const parseJwtUsername = (token: string): string => {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.username || payload.user || payload.name || '管理员'
  } catch (error) {
    console.error('解析JWT失败:', error)
    return '管理员'
  }
}

// 更新用户账号信息
const updateUserAccount = () => {
  if (sessionStore.token) {
    userAccount.value = parseJwtUsername(sessionStore.token)
  } else {
    userAccount.value = '管理员'
  }
}

const resolveLogoPath = (logoValue: string) => {
  const raw = (logoValue || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw) || raw.startsWith('data:')) return raw
  return raw.startsWith('/') ? raw : `/${raw}`
}

// 更新favicon
const updateFavicon = (logoValue: string) => {
  if (!logoValue) return
  let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  if (logoValue.trimStart().startsWith('<')) {
    const svgBlob = new Blob([logoValue], { type: 'image/svg+xml' })
    link.href = URL.createObjectURL(svgBlob)
    link.type = 'image/svg+xml'
    return
  }
  link.href = resolveLogoPath(logoValue)
  link.type = logoValue.toLowerCase().includes('.svg') ? 'image/svg+xml' : 'image/png'
}

// 获取本地配置
const getLocalConfig = () => {
  try {
    const localConfig = LocalStieConfigUtils.getLocalConfig()
    if (localConfig) {
      siteConfig.value = localConfig
      // 更新页面状态中的配置数据
      if (pageState.setSiteConfigData) {
        pageState.setSiteConfigData(localConfig)
      }
      appStore.setSiteConfig(localConfig)
      // 更新网站标题
      if (localConfig.title) {
        document.title = localConfig.title
      }
      // 更新favicon
      if (localConfig.logo) {
        updateFavicon(localConfig.logo)
      }
    }
  } catch (error) {
    console.error('获取本地配置失败:', error)
  }
}

// 获取最新配置并更新
const getLatestConfig = async () => {
  try {
    const latestConfig = await LocalStieConfigUtils.getLatestLocalConfig()
    if (latestConfig) {
      siteConfig.value = latestConfig
      // 更新页面状态中的配置数据
      if (pageState.setSiteConfigData) {
        pageState.setSiteConfigData(latestConfig)
      }
      appStore.setSiteConfig(latestConfig)
      // 更新网站标题
      if (latestConfig.title) {
        document.title = latestConfig.title
      }
      // 更新favicon
      if (latestConfig.logo) {
        updateFavicon(latestConfig.logo)
      }
    }
  } catch (error) {
    console.error('获取最新配置失败:', error)
    // 如果获取最新配置失败，尝试使用本地配置
    getLocalConfig()
  }
}

const loadUserThemePreference = async () => {
  if (!isAuthenticated.value) return
  try {
    const rsp = await usersApi.getTheme()
    const data = rsp?.data?.data || {}
    if (data.theme_mode === 'light' || data.theme_mode === 'dark' || data.theme_mode === 'system') {
      themePreference.value = data.theme_mode
      applyColorModeFromPreference()
    }
  } catch {
  }
}

// 退出登录
const logout = async () => {
  try {
    await confirmAction('退出后将清空当前会话和工作区标签，确认退出登录吗？', '退出登录')
  } catch {
    return
  }
  rbacStore.clear()
  messageCenterStore.resetState()
  resetTabsToDefault()
  await sessionStore.logout()
}

const openProfileSettings = () => {
  const target = '/profile/settings'
  ensureTabForRoute(target)
  if (route.path !== target) {
    router.push(target)
  }
}

// 监听localStorage变化
onMounted(async () => {
  initDashboardDateRange()
  hydrateHeaderDateRangeFromRoute()

  // 初始化主题并监听系统主题变化
  // 注意：如果用户已登录，主题设置会在 loadUserThemePreference 中加载
  // 这里先应用默认/本地设置，避免无主题状态
  applyColorModeFromPreference()
  try {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const handleSystemChange = () => {
      if (themePreference.value === 'system') applyColorModeFromPreference()
    }
    // 新浏览器
    if (media.addEventListener) {
      media.addEventListener('change', handleSystemChange)
    } else if ((media as any).addListener) {
      // 兼容旧浏览器
      ; (media as any).addListener(handleSystemChange)
    }
  } catch { }

  // 初始化用户账号信息
  updateUserAccount();

  await nextTick()
  updateTabOverflow()
  window.addEventListener('resize', handleWindowResize)
  window.addEventListener('keydown', handleGlobalEscape)

  // 初始化配置信息
  getLocalConfig();

  // 如果已认证，获取最新配置和用户显示模式
  if (isAuthenticated.value) {
    await loadUserThemePreference()
    getLatestConfig();
    rbacStore.loadCurrentUserPermissions().catch(() => undefined)
  }

  // 定期检查token状态
  const checkAuth = () => {
    const wasAuthenticated = isAuthenticated.value;
    sessionStore.syncFromStorage();
    // 如果认证状态发生变化，更新用户账号信息和配置
    if (wasAuthenticated !== isAuthenticated.value) {
      updateUserAccount();
      if (isAuthenticated.value) {
        // 用户刚登录，获取最新配置
        getLatestConfig();
        loadUserThemePreference();
        rbacStore.loadCurrentUserPermissions().catch(() => undefined)
      } else {
        // 用户退出登录，使用本地配置
        getLocalConfig();
        rbacStore.clear()
        resetTabsToDefault()
      }
    }
  };
  // 监听storage事件
  window.addEventListener('storage', checkAuth);
  // 定期检查（处理同一页面内的变化）
  const interval = setInterval(checkAuth, 1000);

  // 清理函数
  return () => {
    window.removeEventListener('storage', checkAuth);
    clearInterval(interval);
  };
});

onUnmounted(() => {
  window.removeEventListener('resize', handleWindowResize)
  window.removeEventListener('keydown', handleGlobalEscape)
})

// 计算属性：站点标题
const siteTitle = computed(() => {
  return siteConfig.value?.title || '消息管理系统'
})

const siteSlogan = computed(() => {
  return siteConfig.value?.slogan || ''
})

const siteLogo = computed(() => {
  return resolveLogoPath(siteConfig.value?.logo || '/dora-logo.svg')
})

const siteSloganInitialEnabled = computed(() => {
  return String(siteConfig.value?.slogan_initial_enabled || 'false') === 'true'
})

const currentBreadcrumb = computed(() => {
  const base = findRouteTabMeta(route.path, routeTabMap)?.path || getBasePath(route.path)
  const segments = routeBreadcrumbMap[base]
  if (segments && segments.length) {
    return segments.join(' / ')
  }
  return routeTabMap[base]?.title || '数据统计'
})

const formatDate = (date: Date) => {
  return date.toISOString().split('T')[0]
}

const initDashboardDateRange = () => {
  const today = new Date()
  const before = new Date()
  before.setDate(today.getDate() - 6)
  dashboardDateEnd.value = formatDate(today)
  dashboardDateStart.value = formatDate(before)
}

type HeaderDateQueryMode = 'date' | 'datetime'

const headerActionConfig: Record<string, { showDateRange: boolean; showDownload: boolean; dateQueryMode?: HeaderDateQueryMode }> = {
  '/': { showDateRange: true, showDownload: true, dateQueryMode: 'date' },
  '/sendlogs': { showDateRange: true, showDownload: true, dateQueryMode: 'datetime' },
  '/logs': { showDateRange: true, showDownload: true, dateQueryMode: 'datetime' },
  '/logs/task': { showDateRange: true, showDownload: true, dateQueryMode: 'datetime' },
  '/logs/login': { showDateRange: true, showDownload: true, dateQueryMode: 'datetime' },
  '/logs/consume': { showDateRange: true, showDownload: true, dateQueryMode: 'datetime' },
  '/data/mq-sources': { showDateRange: true, showDownload: true, dateQueryMode: 'date' },
  '/cronmessages': { showDateRange: true, showDownload: true, dateQueryMode: 'date' },
  '/message/subscriptions': { showDateRange: true, showDownload: true, dateQueryMode: 'date' },
  '/templates': { showDateRange: true, showDownload: true, dateQueryMode: 'date' },
  '/sendways': { showDateRange: true, showDownload: true, dateQueryMode: 'date' }
}

const currentHeaderActions = computed(() => {
  const base = getBasePath(route.path)
  return headerActionConfig[base] || null
})

watch(
  () => route.path,
  (path) => {
    appStore.setActivePath(path)
    ensureTabForRoute(path)
    hydrateHeaderDateRangeFromRoute()
  },
  { immediate: true }
)

function hydrateHeaderDateRangeFromRoute() {
  if (!currentHeaderActions.value?.showDateRange) return
  const isDateTimeMode = currentHeaderActions.value?.dateQueryMode === 'datetime'
  const startTime = typeof route.query.start_time === 'string' ? route.query.start_time : ''
  const endTime = typeof route.query.end_time === 'string' ? route.query.end_time : ''
  const startDate = typeof route.query.start_date === 'string' ? route.query.start_date : ''
  const endDate = typeof route.query.end_date === 'string' ? route.query.end_date : ''
  const queryStart = isDateTimeMode ? (startTime || startDate) : (startDate || startTime)
  const queryEnd = isDateTimeMode ? (endTime || endDate) : (endDate || endTime)
  if (queryStart && queryEnd) {
    dashboardDateStart.value = queryStart.slice(0, 10)
    dashboardDateEnd.value = queryEnd.slice(0, 10)
    return
  }
  if (!dashboardDateStart.value || !dashboardDateEnd.value) {
    initDashboardDateRange()
  }
}

const syncHeaderDateRangeToRoute = () => {
  if (!currentHeaderActions.value?.showDateRange) return
  const isDateTimeMode = currentHeaderActions.value?.dateQueryMode === 'datetime'
  const nextQuery = { ...route.query } as Record<string, any>
  if (isDateTimeMode) {
    if (dashboardDateStart.value) nextQuery.start_time = `${dashboardDateStart.value}T00:00`
    else delete nextQuery.start_time
    if (dashboardDateEnd.value) nextQuery.end_time = `${dashboardDateEnd.value}T23:59`
    else delete nextQuery.end_time
    delete nextQuery.start_date
    delete nextQuery.end_date
  } else {
    if (dashboardDateStart.value) nextQuery.start_date = dashboardDateStart.value
    else delete nextQuery.start_date
    if (dashboardDateEnd.value) nextQuery.end_date = dashboardDateEnd.value
    else delete nextQuery.end_date
    delete nextQuery.start_time
    delete nextQuery.end_time
  }
  router.replace({ path: route.path, query: nextQuery })
}

const csvEscape = (value: unknown) => {
  const text = String(value ?? '')
  if (text.includes(',') || text.includes('"') || text.includes('\n')) {
    return `"${text.replace(/"/g, '""')}"`
  }
  return text
}

const toCsvContent = (rows: Array<Array<unknown>>) => {
  return '\uFEFF' + rows.map(row => row.map(csvEscape).join(',')).join('\n')
}

const downloadCsv = (fileName: string, rows: Array<Array<unknown>>) => {
  const csv = toCsvContent(rows)
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', fileName)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const parseFilenameFromDisposition = (contentDisposition: unknown, fallbackName: string) => {
  if (typeof contentDisposition !== 'string' || !contentDisposition.trim()) return fallbackName
  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1])
    } catch {
      return utf8Match[1]
    }
  }
  const plainMatch = contentDisposition.match(/filename="?([^";]+)"?/i)
  if (plainMatch?.[1]) {
    return plainMatch[1]
  }
  return fallbackName
}

const downloadBlobFile = (fileName: string, blob: Blob) => {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', fileName)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const tryBackendCsvExport = async (
  endpoint: string,
  params: Record<string, unknown>,
  fallbackName: string
) => {
  try {
    const rsp = await request.get(endpoint, {
      params,
      responseType: 'blob',
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    const contentType = String(rsp.headers?.['content-type'] || '').toLowerCase()
    if (!contentType.includes('text/csv') && !contentType.includes('application/octet-stream')) {
      return false
    }
    const fileName = parseFilenameFromDisposition(rsp.headers?.['content-disposition'], fallbackName)
    downloadBlobFile(fileName, rsp.data as Blob)
    return true
  } catch (error) {
    console.warn(`后端导出失败，回退前端CSV: ${endpoint}`, error)
    return false
  }
}

const getDateDiffDays = (start: string, end: string) => {
  const startMs = new Date(start).getTime()
  const endMs = new Date(end).getTime()
  if (Number.isNaN(startMs) || Number.isNaN(endMs) || endMs < startMs) return 30
  const dayMs = 24 * 60 * 60 * 1000
  return Math.min(Math.max(Math.floor((endMs - startMs) / dayMs) + 1, 1), 90)
}

const getQueryString = (key: string) => {
  const value = route.query[key]
  return typeof value === 'string' ? value : ''
}

const setParamIfPresent = (params: Record<string, any>, key: string, value: string) => {
  if (value) params[key] = value
}

const setParamIfNotAll = (params: Record<string, any>, key: string, value: string) => {
  if (value && value !== 'all') params[key] = value
}

const setRouteDateTimeRange = (params: Record<string, any>) => {
  setParamIfPresent(params, 'start_time', getQueryString('start_time'))
  setParamIfPresent(params, 'end_time', getQueryString('end_time'))
}

const setRouteDateTimeRangeWithDashboardFallback = (params: Record<string, any>) => {
  const startTime = getQueryString('start_time')
  const endTime = getQueryString('end_time')
  if (startTime) {
    params.start_time = startTime
  } else if (dashboardDateStart.value) {
    params.start_time = `${dashboardDateStart.value}T00:00`
  }
  if (endTime) {
    params.end_time = endTime
  } else if (dashboardDateEnd.value) {
    params.end_time = `${dashboardDateEnd.value}T23:59`
  }
}

const exportListCsvWithBackendFallback = async (
  backendEndpoint: string,
  listEndpoint: string,
  params: Record<string, any>,
  fallbackName: string,
  listErrorMessage: string,
  buildRows: (data: any) => Array<Array<unknown>>
) => {
  const backendDone = await tryBackendCsvExport(backendEndpoint, params, fallbackName)
  if (backendDone) return
  const rsp = await request.get(listEndpoint, { params })
  if (rsp?.data?.code !== 200) {
    throw new Error(listErrorMessage)
  }
  const rows = buildRows(rsp.data.data || {})
  downloadCsv(fallbackName, rows)
}

const buildCsvRows = <T>(
  header: Array<unknown>,
  list: T[],
  toRow: (item: T) => Array<unknown>
) => {
  const rows: Array<Array<unknown>> = [header]
  list.forEach(item => rows.push(toRow(item)))
  return rows
}

const exportDashboardCsv = async () => {
  const days = getDateDiffDays(dashboardDateStart.value, dashboardDateEnd.value)
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `dashboard-report-${datePart}.csv`
  const backendDone = await tryBackendCsvExport('/statistic/export', { days }, fileName)
  if (backendDone) return
  const [basicRsp, trendRsp, channelRsp] = await Promise.all([
    statisticsApi.basic(),
    statisticsApi.trend(days),
    statisticsApi.channels()
  ])

  if (basicRsp?.data?.code !== 200 || trendRsp?.data?.code !== 200 || channelRsp?.data?.code !== 200) {
    throw new Error('获取导出数据失败')
  }

  const basic = basicRsp.data.data || {}
  const trend = trendRsp.data.data?.latest_send_data || []
  const channels = channelRsp.data.data?.way_cate_data || []
  const rows: Array<Array<unknown>> = [
    ['模块', '指标', '值'],
    ['概览', '发送日志数', basic.message_total_num || 0],
    ['概览', '今日发送数', basic.today_total_num || 0],
    ['概览', '今日成功数', basic.today_succ_num || 0],
    ['概览', '今日失败数', basic.today_failed_num || 0],
    [],
    ['趋势', '日期', '发送总数', '发送成功数', '发送失败数']
  ]

  trend.forEach((item: any) => {
    rows.push(['趋势', item.day || '-', item.num || 0, item.day_succ_num || 0, item.day_failed_num || 0])
  })

  rows.push([])
  rows.push(['渠道', '渠道名称', '发送数'])
  channels.forEach((item: any) => {
    rows.push(['渠道', item.way_name || '-', item.count_num || 0])
  })
  downloadCsv(fileName, rows)
}

const exportSendLogsCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `sendlogs-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    size: 1000,
    name: getQueryString('name'),
    taskid: getQueryString('taskid')
  }
  const status = getQueryString('status')
  if (status && status !== 'all') {
    params.status = status
  } else if (getQueryString('query')) {
    // 兼容旧版 sendLogs query(JSON) 参数
    params.query = getQueryString('query')
  }
  setRouteDateTimeRangeWithDashboardFallback(params)
  await exportListCsvWithBackendFallback(
    '/sendlogs/export',
    '/sendlogs/list',
    params,
    fileName,
    '获取日志导出数据失败',
    (data: any) => {
      const list = data?.lists || []
      return buildCsvRows(
        ['ID', '类型', '名称', '日志', '发送时间', '状态'],
        list,
        (item: any) => [
          item.id || '',
          item.type || '',
          item.name || '',
          item.log || '',
          item.created_on || '',
          item.status === 1 ? '成功' : '失败'
        ]
      )
    },
  )
}

const exportCronMessagesCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `cronmessages-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    size: 1000,
    name: getQueryString('name')
  }
  setParamIfPresent(params, 'status', getQueryString('status'))
  await exportListCsvWithBackendFallback(
    '/cronmessages/export',
    '/cronmessages/list',
    params,
    fileName,
    '获取定时任务导出数据失败',
    (data: any) => {
      const list = data?.lists || []
      return buildCsvRows(
        ['ID', '名称', '模板', '渠道', 'Cron表达式', '下次执行时间', '创建时间', '启用状态'],
        list,
        (item: any) => [
          item.id || '',
          item.name || '',
          item.template_name || item.template_id || '',
          Array.isArray(item.channel_names) ? item.channel_names.join('、') : '',
          item.cron || item.cron_expression || '',
          item.next_time || '',
          item.created_on || '',
          Number(item.enable) === 1 ? '启用' : '停用'
        ]
      )
    },
  )
}

const exportSubscriptionsCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `subscriptions-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    page_size: 1000
  }
  setParamIfPresent(params, 'name', getQueryString('name'))
  setParamIfNotAll(params, 'source_id', getQueryString('source_id'))
  setParamIfNotAll(params, 'status', getQueryString('status'))
  await exportListCsvWithBackendFallback(
    '/subscriptions/export',
    '/subscriptions/list',
    params,
    fileName,
    '获取订阅导出数据失败',
    (data: any) => {
      const list = data?.list || []
      return buildCsvRows(
        ['ID', '订阅名称', '数据源', 'Topic', 'Tag', '状态', '消费总数', '发送总数', '失败总数', '最后消费时间'],
        list,
        (item: any) => [
          item.id || '',
          item.name || '',
          item.source_name || '',
          item.topic || '',
          item.tag || '*',
          item.status === 'running' ? '运行中' : '已停止',
          item.total_consumed || 0,
          item.total_sent || 0,
          item.total_failed || 0,
          item.last_consume_time || ''
        ]
      )
    },
  )
}

const exportTemplatesCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `templates-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    size: 1000,
    text: getQueryString('text') || getQueryString('name')
  }
  setParamIfNotAll(params, 'status', getQueryString('status'))
  await exportListCsvWithBackendFallback(
    '/templates/export',
    '/templates/list',
    params,
    fileName,
    '获取模板导出数据失败',
    (data: any) => {
      const list = data?.lists || []
      return buildCsvRows(
        ['ID', '模板名称', '描述', '状态', '创建时间', '更新时间'],
        list,
        (item: any) => [
          item.id || '',
          item.name || '',
          item.description || '',
          item.status === 'enabled' ? '启用' : '禁用',
          item.created_on || '',
          item.modified_on || ''
        ]
      )
    },
  )
}

const exportSendWaysCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `sendways-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    size: 1000,
    name: getQueryString('name')
  }
  setParamIfNotAll(params, 'type', getQueryString('channel_type'))
  setParamIfNotAll(params, 'status', getQueryString('status'))
  await exportListCsvWithBackendFallback(
    '/sendways/export',
    '/sendways/list',
    params,
    fileName,
    '获取渠道导出数据失败',
    (data: any) => {
      const list = data?.lists || []
      return buildCsvRows(
        ['ID', '渠道名称', '渠道类型', '状态', '创建时间', '更新时间'],
        list,
        (item: any) => [
          item.id || '',
          item.name || '',
          item.type || '',
          Number(item.status) === 1 ? '启用' : '停用',
          item.created_on || '',
          item.modified_on || ''
        ]
      )
    },
  )
}

const exportLoginLogsCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `loginlogs-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    page_size: 1000
  }
  setRouteDateTimeRange(params)
  await exportListCsvWithBackendFallback(
    '/loginlogs/export',
    '/loginlogs/recent',
    params,
    fileName,
    '获取登录日志导出数据失败',
    (data: any) => {
      const list = data?.lists || []
      return buildCsvRows(
        ['ID', '用户ID', '用户名', 'IP', 'UA', '登录时间'],
        list,
        (item: any) => [
          item.id || '',
          item.user_id || '',
          item.username || '',
          item.ip || '',
          item.ua || '',
          item.created_on || ''
        ]
      )
    },
  )
}

const exportConsumeLogsCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `consume-logs-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    page_size: 1000
  }
  setParamIfPresent(params, 'subscription_name', getQueryString('subscription_name'))
  setParamIfNotAll(params, 'matched', getQueryString('matched'))
  setParamIfNotAll(params, 'send_status', getQueryString('send_status'))
  setRouteDateTimeRange(params)
  await exportListCsvWithBackendFallback(
    '/consume-logs/export',
    '/consume-logs/list',
    params,
    fileName,
    '获取消费日志导出数据失败',
    (data: any) => {
      const list = data?.list || []
      return buildCsvRows(
        ['ID', '订阅ID', '订阅名称', '原始消息', '匹配状态', '发送状态', '发送错误', '消费时间'],
        list,
        (item: any) => [
          item.id || '',
          item.subscription_id || '',
          item.subscription_name || '',
          item.raw_message || '',
          Number(item.matched) === 1 ? '已匹配' : '未匹配',
          Number(item.send_status) === 1 ? '发送成功' : (Number(item.send_status) === 2 ? '发送失败' : '未发送'),
          item.send_error || '',
          item.created_on || ''
        ]
      )
    },
  )
}

const exportMqSourcesCsv = async () => {
  const datePart = new Date().toISOString().slice(0, 10)
  const fileName = `mq-sources-report-${datePart}.csv`
  const params: Record<string, any> = {
    page: 1,
    page_size: 1000
  }
  setParamIfPresent(params, 'name', getQueryString('name'))
  setParamIfNotAll(params, 'type', getQueryString('type'))
  const status = getQueryString('status')
  if (status && status !== 'all') {
    params.status = status === '__untested__' ? 'untested' : status
  }
  await exportListCsvWithBackendFallback(
    '/mq-sources/export',
    '/mq-sources/list',
    params,
    fileName,
    '获取数据源导出数据失败',
    (data: any) => {
      const list = data?.list || []
      return buildCsvRows(
        ['ID', '数据源名称', '类型', '地址', '状态', '绑定订阅数', '最后测试时间', '创建时间', '更新时间'],
        list,
        (item: any) => {
          const statusText = item.last_test_status === 'success'
            ? '在线'
            : (item.last_test_status === 'failed' ? '离线' : '未测试')
          return [
          item.id || '',
          item.name || '',
          item.type || '',
          item.namesrv_addr || '',
          statusText,
          item.binding_count ?? 0,
          item.last_test_time || '',
          item.created_on || '',
          item.modified_on || ''
          ]
        }
      )
    },
  )
}

const handleHeaderDownload = async () => {
  if (isExporting.value) return
  const base = getBasePath(route.path)
  isExporting.value = true
  try {
    if (base === '/') {
      await exportDashboardCsv()
      notifySuccess('统计报表导出成功')
      return
    }
    if (base === '/sendlogs' || base === '/logs' || base === '/logs/task') {
      await exportSendLogsCsv()
      notifySuccess('日志导出成功')
      return
    }
    if (base === '/logs/login') {
      await exportLoginLogsCsv()
      notifySuccess('登录日志导出成功')
      return
    }
    if (base === '/logs/consume') {
      await exportConsumeLogsCsv()
      notifySuccess('消费日志导出成功')
      return
    }
    if (base === '/data/mq-sources') {
      await exportMqSourcesCsv()
      notifySuccess('数据源导出成功')
      return
    }
    if (base === '/cronmessages') {
      await exportCronMessagesCsv()
      notifySuccess('定时任务导出成功')
      return
    }
    if (base === '/message/subscriptions') {
      await exportSubscriptionsCsv()
      notifySuccess('订阅导出成功')
      return
    }
    if (base === '/templates') {
      await exportTemplatesCsv()
      notifySuccess('模板导出成功')
      return
    }
    if (base === '/sendways') {
      await exportSendWaysCsv()
      notifySuccess('渠道导出成功')
      return
    }
    notifyInfo('当前页面暂不支持导出')
  } catch (error) {
    console.error('导出失败:', error)
    notifyError('导出失败，请稍后重试')
  } finally {
    isExporting.value = false
  }
}

void syncHeaderDateRangeToRoute
void handleHeaderDownload

const layoutTabbarHeight = computed(() => (showTabBar.value ? 'var(--dora-tabbar-height)' : '0px'))
const contentMarginLeft = computed(() => `${contentOffset.value}px`)

watch(
  () => [tabs.value.length, activeTabPath.value, layoutState.value.viewportWidth],
  async () => {
    await nextTick()
    scrollActiveTabIntoView()
    updateTabOverflow()
  },
  { flush: 'post' }
)
</script>


<template>
  <router-view v-if="!isAuthenticated || route.path == '/login' || route.path == 'login'"></router-view>

  <div class="layout" v-else>
    <div
      v-if="isMobileSidebarVisible"
      class="mobile-sidebar-mask"
      @click="closeMobileSidebarWithFocusRestore"
    />
    <Sidebar
      :is-collapsed="isSidebarCollapsed"
      :is-mobile="isMobileLayout"
      :is-mobile-visible="isMobileSidebarVisible"
      :site-title="siteTitle"
      :site-slogan="siteSlogan"
      :site-logo="siteLogo"
      :site-slogan-initial-enabled="siteSloganInitialEnabled"
      :user-account="userAccount"
      @navigate="closeMobileSidebarWithFocusRestore"
    />
    <main
      class="content dora-no-horizontal-overflow"
      :style="{ marginLeft: contentMarginLeft }"
      :data-device="layoutState.device"
    >
      <Header
        :user-account="userAccount"
        :theme="theme"
        :breadcrumb="currentBreadcrumb"
        :is-sidebar-collapsed="isSidebarCollapsed"
        :is-mobile="isMobileLayout"
        @toggle-theme="toggleTheme"
        @toggle-sidebar="toggleSidebar"
        @open-profile-settings="openProfileSettings"
        @logout="logout"
      />
      <div v-if="showTabBar" class="admin-tabbar flex items-center px-4 gap-2">
        <button
          type="button"
          class="admin-tab-scroll-button inline-flex h-8 w-8 items-center justify-center"
          :disabled="!canScrollTabsLeft"
          :aria-label="messages.scrollTabsLeft"
          @click="scrollTabs('left')"
        >
          <DoraIcon name="chevron-left" :size="15" />
        </button>
        <div ref="tabScrollRef" class="admin-tab-scroll flex-1 overflow-x-auto" @scroll="updateTabOverflow">
          <div class="admin-tab-list flex items-center h-full gap-0 pr-3" role="tablist" :aria-label="messages.workspaceTabs">
            <div
              v-for="(tab, tabIndex) in tabs"
              :key="tab.path"
              class="relative"
            >
              <div
                :ref="(el) => setTabItemRef(tab.path, el)"
                class="tab-item admin-tab-item relative inline-flex items-center justify-center h-8 px-3 text-[14px] cursor-pointer select-none whitespace-nowrap transition-[background-color,border-color,color] duration-200 ease-out"
                :class="[
                  tab.path === activeTabPath
                    ? 'admin-tab-item-active'
                    : 'admin-tab-item-idle',
                  tab.closable ? 'admin-tab-item-closable' : ''
                ]"
                role="tab"
                :id="`workspace-tab-${tabIndex}`"
                :tabindex="tab.path === activeTabPath ? 0 : -1"
                :aria-controls="'workspace-tabpanel'"
                :aria-selected="tab.path === activeTabPath"
                :aria-label="`${tab.title}${tab.fixed ? messages.fixedTabSuffix : ''}${messages.tabActionsSuffix}`"
                @click="activateTab(tab)"
                @keydown.enter.prevent="activateTab(tab)"
                @keydown.space.prevent="activateTab(tab)"
                @keydown="handleTabKeydown($event, tabIndex)"
                @contextmenu="openTabContextMenu($event, tab)"
              >
                <DoraIcon v-if="tab.icon" :name="tab.icon" :size="14" class="admin-tab-page-icon" />
                <span v-else-if="tab.fixed" class="admin-tab-pin" aria-hidden="true" />
                <span class="truncate max-w-[150px]">{{ tab.title }}</span>
                <span v-if="refreshingPath === tab.path" class="admin-tab-refreshing" :aria-label="messages.refreshing" />
                <button
                  v-if="tab.closable"
                  type="button"
                  class="admin-tab-close inline-flex items-center justify-center overflow-hidden text-[var(--admin-text-muted)]"
                  :aria-label="messages.closeTab"
                  @click.stop="closeWorkspaceTab(tab)"
                >
                  <DoraIcon name="close-slash" :size="16" />
                </button>
              </div>
            </div>
          </div>
        </div>
        <button
          type="button"
          class="admin-tab-scroll-button inline-flex h-8 w-8 items-center justify-center"
          :disabled="!canScrollTabsRight"
          :aria-label="messages.scrollTabsRight"
          @click="scrollTabs('right')"
        >
          <DoraIcon name="chevron-right" :size="15" />
        </button>
        <button
          class="admin-tab-close-all h-8 px-3 text-[13px] disabled:opacity-45 disabled:cursor-not-allowed whitespace-nowrap transition-colors duration-[var(--motion-fast)]"
          :disabled="closableTabsCount === 0"
          @click="closeAllWorkspaceTabs"
        >
          {{ messages.closeAllTabs }}
        </button>
      </div>
      <div
        v-if="tabContextMenu.visible"
        ref="tabContextMenuRef"
        class="admin-context-menu admin-workspace-menu fixed z-50 w-40 rounded-md text-sm py-1"
        :style="{ left: tabContextMenu.x + 'px', top: tabContextMenu.y + 'px' }"
        role="menu"
        @click.stop
        @keydown.esc="hideTabContextMenu"
      >
        <button
          class="admin-workspace-menu-item"
          role="menuitem"
          @click="refreshActiveTab"
        >
          {{ messages.refreshCurrent }}
        </button>
        <button
          class="admin-workspace-menu-item"
          role="menuitem"
          :disabled="closableLeftCount === 0"
          @click="closeLeftTabs"
        >
          {{ messages.closeLeft }}
        </button>
        <button
          class="admin-workspace-menu-item"
          role="menuitem"
          :disabled="closableRightCount === 0"
          @click="closeRightTabs"
        >
          {{ messages.closeRight }}
        </button>
        <button
          class="admin-workspace-menu-item"
          role="menuitem"
          :disabled="closableOtherCount === 0"
          @click="closeOtherTabs"
        >
          {{ messages.closeOthers }}
        </button>
        <button
          class="admin-workspace-menu-item danger"
          role="menuitem"
          :disabled="!canCloseContextTab"
          @click="closeContextTab"
        >
          {{ messages.closeCurrent }}
        </button>
      </div>
      <div class="admin-shell-body">
        <div class="page-container" :style="{ '--layout-tabbar-height': layoutTabbarHeight }">
          <div
            id="workspace-tabpanel"
            class="main-card dora-contained-scroll"
            role="tabpanel"
            tabindex="0"
            :aria-labelledby="`workspace-tab-${Math.max(tabs.findIndex(tab => tab.path === activeTabPath), 0)}`"
          >
            <router-view :key="routerViewKey" @click="hideTabContextMenu"></router-view>
          </div>
        </div>
        <footer class="admin-shell-footer">
          Copyright MIT © 2026 {{ siteTitle || messages.defaultSiteTitle }}
        </footer>
      </div>
    </main>
  </div>

</template>
