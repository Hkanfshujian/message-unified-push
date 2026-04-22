<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { applyTheme, getStoredTheme } from '@/util/theme'
import { CONSTANT } from '../constant.js'
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { usePageState } from '@/store/page_sate.js'
import { useRbacAuthzStore } from '@/store/rbac_authz'
import { useRoute, useRouter } from 'vue-router'
import { request } from '@/api/api'
import Sidebar from '@/components/layout/Sidebar.vue'

const route = useRoute()
const router = useRouter()
const pageState = usePageState()
const rbacAuthzStore = useRbacAuthzStore()
const isAuthenticated = ref(Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME)));
const userAccount = ref('管理员')
const siteConfig = ref<any>({})
const isSidebarCollapsed = ref(false)

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

// 监听localStorage变化
onMounted(async () => {
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
  () => route.path,
  (path) => {
    ensureTabForRoute(path)
  },
  { immediate: true }
)

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
      <div class="border-b bg-[var(--tabbar-bg,white)] h-12 flex items-end px-4 gap-2">
        <div class="flex-1 overflow-x-auto">
          <div class="flex items-end h-10">
            <div
              v-for="tab in tabs"
              :key="tab.path"
              class="relative mr-1"
            >
              <div
                class="tab-item inline-flex items-center h-10 px-4 text-sm cursor-pointer select-none rounded-t border border-b-0 bg-[#fafafa]"
                :class="tab.path === activeTabPath ? 'bg-white' : ''"
                @click="activateTab(tab)"
                @contextmenu="openTabContextMenu($event, tab)"
              >
                <span class="truncate max-w-[120px]">{{ tab.title }}</span>
                <button
                  v-if="tab.closable"
                  class="ml-2 text-xs opacity-60 hover:opacity-100 hover:text-[#ff4d4f]"
                  @click.stop="closeTab(tab)"
                >
                  ×
                </button>
              </div>
              <div
                v-if="tab.path === activeTabPath"
                class="absolute left-0 right-0 bottom-0 h-[2px] bg-[#1890ff]"
              />
            </div>
          </div>
        </div>
        <button
          class="ml-2 mb-1 text-xs text-muted-foreground hover:text-red-500 px-2 py-1 rounded border border-transparent hover:border-red-200 whitespace-nowrap"
          @click="closeAllTabs"
        >
          关闭全部标签
        </button>
      </div>
      <div
        v-if="tabContextMenu.visible"
        class="fixed z-50 w-32 rounded bg-white shadow border text-sm py-1"
        :style="{ left: tabContextMenu.x + 'px', top: tabContextMenu.y + 'px' }"
        @click.stop
      >
        <button
          class="w-full text-left px-3 py-1 hover:bg-[#f5f5f5]"
          @click="closeOtherTabs"
        >
          关闭其他
        </button>
        <button
          class="w-full text-left px-3 py-1 hover:bg-[#f5f5f5]"
          @click="closeAllTabs"
        >
          关闭全部
        </button>
      </div>
      <div class="page-container">
        <div class="main-card">
          <router-view :key="route.fullPath" @click="hideTabContextMenu"></router-view>
        </div>
      </div>
    </main>
  </div>

</template>
