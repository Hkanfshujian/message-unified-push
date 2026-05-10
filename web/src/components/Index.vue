<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import { applyTheme, getStoredTheme } from '@/util/theme'
import { CONSTANT } from '../constant.js'
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { usePageState } from '@/store/page_sate.js'
import { useRbacAuthzStore } from '@/store/rbac_authz'
import { useRoute, useRouter } from 'vue-router'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'

const route = useRoute()
const router = useRouter()
const pageState = usePageState()
const rbacAuthzStore = useRbacAuthzStore()
const isAuthenticated = ref(Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME)));
const userAccount = ref('管理员')
const siteConfig = ref<any>({})
const isSidebarCollapsed = ref(false)
const showTabBar = ref(true)
const dashboardDateStart = ref('')
const dashboardDateEnd = ref('')
const isExporting = ref(false)

interface TabItem {
  title: string
  path: string
  closable: boolean
}

const TABS_STORAGE_KEY = CONSTANT.STORE_OPEN_TABS_NAME || 'message_nest_open_tabs_v1'

const routeTabMap: Record<string, TabItem> = {
  '/': { title: '数据统计', path: '/', closable: false },
  '/cronmessages': { title: '定时消息', path: '/cronmessages', closable: true },
  '/templates': { title: '模板管理', path: '/templates', closable: true },
  '/sendways': { title: '渠道管理', path: '/sendways', closable: true },
  '/logs': { title: '日志管理', path: '/logs', closable: true },
  '/logs/task': { title: '任务日志', path: '/logs/task', closable: true },
  '/logs/login': { title: '登录日志', path: '/logs/login', closable: true },
  '/logs/consume': { title: '消费日志', path: '/logs/consume', closable: true },
  '/system/settings': { title: '系统设置', path: '/system/settings', closable: true },
  '/profile/settings': { title: '个人设置', path: '/profile/settings', closable: true },
  '/system/roles': { title: '角色管理', path: '/system/roles', closable: true },
  '/system/groups': { title: '用户组管理', path: '/system/groups', closable: true },
  '/system/permissions': { title: '权限管理', path: '/system/permissions', closable: true },
  '/system/users': { title: '用户管理', path: '/system/users', closable: true },
  '/system/relations': { title: '授权关系', path: '/system/relations', closable: true },
  '/data/mq-sources': { title: '消息队列', path: '/data/mq-sources', closable: true },
  '/message/subscriptions': { title: '订阅消息', path: '/message/subscriptions', closable: true }
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
  '/system/settings': ['系统管理', '系统设置'],
  '/system/relations': ['系统管理', '授权关系'],
  '/profile/settings': ['个人设置']
}
const createDefaultTabs = (): TabItem[] => {
  return [{ title: '数据统计', path: '/', closable: false }]
}

const clearTabsCache = () => {
  try {
    localStorage.removeItem(TABS_STORAGE_KEY)
  } catch {
  }
}

const resetTabsToDefault = () => {
  tabs.value = createDefaultTabs()
  activeTabPath.value = '/'
  clearTabsCache()
}

const normalizeTabs = (rawTabs: TabItem[]) => {
  const unique = new Map<string, TabItem>()
  // 固定首页标签始终存在
  unique.set('/', createDefaultTabs()[0])
  for (const t of rawTabs) {
    const base = getBasePath(t.path)
    const meta = routeTabMap[base]
    if (!meta) continue
    unique.set(meta.path, { ...meta })
  }
  return Array.from(unique.values())
}

const loadTabsFromStorage = (): TabItem[] => {
  try {
    const raw = localStorage.getItem(TABS_STORAGE_KEY)
    if (!raw) return createDefaultTabs()
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return createDefaultTabs()
    return normalizeTabs(parsed)
  } catch {
    return createDefaultTabs()
  }
}

const tabs = ref<TabItem[]>(loadTabsFromStorage())
const activeTabPath = ref(route.path)

function getBasePath(path: string) {
  if (path === '/') return '/'
  const pure = path.split('?')[0]
  const parts = pure.split('/').filter(Boolean)
  
  // 对于二级及以上路径，返回前两级
  if (parts.length >= 2) {
    return `/${parts[0]}/${parts[1]}`
  }
  // 只有一级路径
  if (parts.length === 1) {
    return `/${parts[0]}`
  }
  return '/'
}

const ensureTabForRoute = (path: string) => {
  const base = getBasePath(path)
  const meta = routeTabMap[base]
  if (!meta) {
    activeTabPath.value = path
    return
  }
  if (!tabs.value.some(t => t.path === meta.path)) {
    tabs.value.push({ ...meta })
  }
  activeTabPath.value = meta.path
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

const applyThemeFromPreference = () => {
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
  applyThemeFromPreference()
}

// 切换侧边栏
const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const activateTab = (tab: TabItem) => {
  if (tab.path !== route.path) {
    router.push(tab.path)
  }
}

const closeTab = (tab: TabItem) => {
  if (!tab.closable) return
  const index = tabs.value.findIndex(t => t.path === tab.path)
  if (index === -1) return
  tabs.value.splice(index, 1)
  if (activeTabPath.value === tab.path) {
    const next = tabs.value[index] || tabs.value[index - 1] || tabs.value[0]
    if (next) {
      router.push(next.path)
    }
  }
}

const tabContextMenu = ref<{
  visible: boolean
  x: number
  y: number
  path: string | null
}>({
  visible: false,
  x: 0,
  y: 0,
  path: null
})

const openTabContextMenu = (event: MouseEvent, tab: TabItem) => {
  event.preventDefault()
  if (!tab.closable) return
  tabContextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    path: tab.path
  }
}

const hideTabContextMenu = () => {
  tabContextMenu.value.visible = false
}

const closeOtherTabs = () => {
  const currentPath = tabContextMenu.value.path
  if (!currentPath) return
  tabs.value = tabs.value.filter(t => !t.closable || t.path === currentPath)
  activeTabPath.value = currentPath
  hideTabContextMenu()
}

const closeAllTabs = () => {
  const fixedTabs = tabs.value.filter(t => !t.closable)
  const fallback = fixedTabs[0] || { title: '数据统计', path: '/', closable: false }
  const remaining = fixedTabs.length ? fixedTabs : [fallback]
  tabs.value = remaining
  activeTabPath.value = remaining[0].path
  router.push(remaining[0].path)
  hideTabContextMenu()
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
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME)
  if (token) {
    userAccount.value = parseJwtUsername(token)
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
  link.type = 'image/png'
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
      // 更新网站标题
      if (localConfig.title) {
        document.title = localConfig.title
      }
      // 更新favicon
      if (localConfig.logo) {
        updateFavicon(localConfig.logo)
      }
      // 更新主题色
      if (localConfig.theme_color) {
        applyTheme(localConfig.theme_color)
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
      // 更新网站标题
      if (latestConfig.title) {
        document.title = latestConfig.title
      }
      // 更新favicon
      if (latestConfig.logo) {
        updateFavicon(latestConfig.logo)
      }
      // 更新主题色
      if (latestConfig.theme_color) {
        applyTheme(latestConfig.theme_color)
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
    const rsp = await request.get('/profile/theme')
    const data = rsp?.data?.data || {}
    if (data.theme_color) {
      applyTheme(data.theme_color)
    }
    if (data.theme_mode === 'light' || data.theme_mode === 'dark' || data.theme_mode === 'system') {
      themePreference.value = data.theme_mode
      applyThemeFromPreference()
    }
    // 应用侧边栏背景色
    if (data.sidebar_bg) {
      document.documentElement.style.setProperty('--sidebar-bg', data.sidebar_bg)
    }
  } catch {
  }
}

// 退出登录
const logout = async () => {
  localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
  rbacAuthzStore.clearAuthzData()
  resetTabsToDefault()
  isAuthenticated.value = false
  localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME)
  router.push('/login')
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
  applyThemeFromPreference()
  try {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const handleSystemChange = () => {
      if (themePreference.value === 'system') applyThemeFromPreference()
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

  // 初始化配置信息
  getLocalConfig();

  // 如果已认证，获取最新配置和用户个性设置
  if (isAuthenticated.value) {
    // 先加载用户个性设置（包含主题色），避免闪烁
    await loadUserThemePreference()
    getLatestConfig();
    rbacAuthzStore.fetchCurrentUserPermissions({ silent: true })
  } else {
    // 未登录状态下应用本地存储的主题色
    applyTheme(getStoredTheme())
  }

  // 定期检查token状态
  const checkAuth = () => {
    const wasAuthenticated = isAuthenticated.value;
    isAuthenticated.value = Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME));
    // 如果认证状态发生变化，更新用户账号信息和配置
    if (wasAuthenticated !== isAuthenticated.value) {
      updateUserAccount();
      if (isAuthenticated.value) {
        // 用户刚登录，获取最新配置
        getLatestConfig();
        loadUserThemePreference();
        rbacAuthzStore.fetchCurrentUserPermissions({ silent: true })
      } else {
        // 用户退出登录，使用本地配置
        getLocalConfig();
        rbacAuthzStore.clearAuthzData()
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

watch(
  tabs,
  (newTabs) => {
    try {
      localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(newTabs))
    } catch { }
  },
  { deep: true }
)

// 计算属性：站点标题
const siteTitle = computed(() => {
  return siteConfig.value?.title || '消息管理系统'
})

const siteSlogan = computed(() => {
  return siteConfig.value?.slogan || ''
})

const siteSloganInitialEnabled = computed(() => {
  return String(siteConfig.value?.slogan_initial_enabled || 'false') === 'true'
})

const currentBreadcrumb = computed(() => {
  const base = getBasePath(route.path)
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
    request.get('/statistic?type=basic'),
    request.get(`/statistic?type=trend&days=${days}`),
    request.get('/statistic?type=channels')
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
      toast.success('统计报表导出成功')
      return
    }
    if (base === '/sendlogs' || base === '/logs' || base === '/logs/task') {
      await exportSendLogsCsv()
      toast.success('日志导出成功')
      return
    }
    if (base === '/logs/login') {
      await exportLoginLogsCsv()
      toast.success('登录日志导出成功')
      return
    }
    if (base === '/logs/consume') {
      await exportConsumeLogsCsv()
      toast.success('消费日志导出成功')
      return
    }
    if (base === '/data/mq-sources') {
      await exportMqSourcesCsv()
      toast.success('数据源导出成功')
      return
    }
    if (base === '/cronmessages') {
      await exportCronMessagesCsv()
      toast.success('定时任务导出成功')
      return
    }
    if (base === '/message/subscriptions') {
      await exportSubscriptionsCsv()
      toast.success('订阅导出成功')
      return
    }
    if (base === '/templates') {
      await exportTemplatesCsv()
      toast.success('模板导出成功')
      return
    }
    if (base === '/sendways') {
      await exportSendWaysCsv()
      toast.success('渠道导出成功')
      return
    }
    toast.info('当前页面暂不支持导出')
  } catch (error) {
    console.error('导出失败:', error)
    toast.error('导出失败，请稍后重试')
  } finally {
    isExporting.value = false
  }
}

void syncHeaderDateRangeToRoute
void handleHeaderDownload

const layoutTopOffset = computed(() => (showTabBar.value ? '160px' : '112px'))
</script>


<template>
  <router-view v-if="!isAuthenticated || route.path == '/login' || route.path == 'login'"></router-view>

  <div class="layout" v-else>
    <Sidebar
      :is-collapsed="isSidebarCollapsed"
      :site-title="siteTitle"
      :site-slogan="siteSlogan"
      :site-slogan-initial-enabled="siteSloganInitialEnabled"
      :user-account="userAccount"
      @toggle-collapse="toggleSidebar"
      @logout="logout"
    />
    <main
      class="content"
      :class="isSidebarCollapsed ? 'ml-16' : 'ml-[200px]'"
    >
      <Header
        :user-account="userAccount"
        :theme="theme"
        :theme-preference="themePreference"
        :breadcrumb="currentBreadcrumb"
        :is-sidebar-collapsed="isSidebarCollapsed"
        @toggle-theme="toggleTheme"
        @toggle-sidebar="toggleSidebar"
        @open-profile-settings="openProfileSettings"
        @logout="logout"
      />
      <div v-if="showTabBar" class="border-b weak-divider bg-background/95 h-12 flex items-center px-3 gap-2">
        <div class="flex-1 overflow-x-auto">
          <div class="flex items-center h-10 gap-1.5">
            <div
              v-for="tab in tabs"
              :key="tab.path"
              class="relative"
            >
              <div
                class="tab-item relative inline-flex items-center h-9 px-4 text-[14px] cursor-pointer select-none rounded-[8px] border-x border-y-0 transition-[background-color,box-shadow,color,transform,border-color] duration-200 ease-out hover:-translate-y-[1px]"
                :class="tab.path === activeTabPath
                  ? 'bg-white text-[#24314d] border-x-[#d7e4fb] shadow-[0_4px_12px_rgba(31,71,142,0.12)]'
                  : 'bg-transparent text-muted-foreground border-x-[#e8eef9] hover:bg-muted/40 hover:text-foreground hover:border-x-[#dbe6fb]'"
                @click="activateTab(tab)"
                @contextmenu="openTabContextMenu($event, tab)"
              >
                <span class="truncate max-w-[150px]">{{ tab.title }}</span>
                <button
                  v-if="tab.closable"
                  class="ml-2 inline-flex h-4 w-4 items-center justify-center rounded-sm text-[#8b98b6] hover:bg-[#eef3ff] hover:text-red-500 transition-colors duration-[var(--motion-fast)]"
                  aria-label="关闭标签"
                  @click.stop="closeTab(tab)"
                >
                    <CloseOutlined class="text-[12px]" />
                </button>
                <span
                  class="pointer-events-none absolute left-1.5 right-1.5 bottom-0 h-[2.5px] rounded-full bg-[linear-gradient(90deg,rgba(79,141,255,0)_0%,rgba(79,141,255,0.72)_22%,rgba(64,129,248,1)_50%,rgba(79,141,255,0.72)_78%,rgba(79,141,255,0)_100%)] origin-center transition-all duration-200 ease-out"
                  :class="tab.path === activeTabPath ? 'opacity-100 scale-x-100' : 'opacity-0 scale-x-0'"
                />
              </div>
            </div>
          </div>
        </div>
        <button
          class="ml-2 h-8 px-3 rounded-md text-[13px] border border-[#e5ecf9] bg-white text-[#5e6e90] hover:text-red-500 hover:border-red-200 hover:bg-red-50/40 disabled:opacity-45 disabled:cursor-not-allowed disabled:hover:text-[#5e6e90] disabled:hover:border-[#e5ecf9] disabled:hover:bg-white whitespace-nowrap transition-colors duration-[var(--motion-fast)]"
          :disabled="tabs.length <= 1"
          @click="closeAllTabs"
        >
          关闭全部标签
        </button>
      </div>
      <div
        v-if="tabContextMenu.visible"
        class="fixed z-50 w-32 rounded-md bg-white shadow-[0_10px_26px_rgba(15,42,91,0.18)] border border-[#e6edf9] text-sm py-1"
        :style="{ left: tabContextMenu.x + 'px', top: tabContextMenu.y + 'px' }"
        @click.stop
      >
        <button
          class="w-full text-left px-3 py-1.5 hover:bg-[#f3f7ff]"
          @click="closeOtherTabs"
        >
          关闭其他
        </button>
        <button
          class="w-full text-left px-3 py-1.5 hover:bg-[#f3f7ff]"
          @click="closeAllTabs"
        >
          关闭全部
        </button>
      </div>
      <div class="page-container" :style="{ '--layout-tabbar-height': layoutTopOffset }">
        <div class="main-card">
          <router-view :key="route.fullPath" @click="hideTabContextMenu"></router-view>
        </div>
      </div>
    </main>
  </div>

</template>
