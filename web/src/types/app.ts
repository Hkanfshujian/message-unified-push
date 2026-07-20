import type { RouteLocationRaw } from 'vue-router'

export type AuthSource = 'local' | 'casdoor' | ''

export interface AuthSessionState {
  token: string
  isLogin: boolean
  authSource: AuthSource
  username: string
  isLoggingOut: boolean
}

export interface RbacPermissionState {
  userId: number
  username: string
  roles: string[]
  groups: string[]
  permissions: string[]
  isSuperAdmin: boolean
  loaded: boolean
}

export interface SiteConfigState {
  title: string
  login_title: string
  logo: string
  slogan: string
  pagesize: string | number
  channel_test_message?: string
  slogan_initial_enabled?: boolean | string
  cookie_exp_days?: string | number
  [key: string]: unknown
}

export interface OpenTabItem {
  path: string
  fullPath?: string
  title: string
  icon?: DoraIconName
  name?: string
  closable: boolean
  fixed?: boolean
  active?: boolean
  refreshKey?: number
  query?: Record<string, unknown>
  route?: RouteLocationRaw
}

export type LayoutDevice = 'desktop' | 'mobile'

export type LayoutViewportClass = 'mobile' | 'compact-desktop' | 'wide-desktop'

export type LayoutMode = 'expanded' | 'collapsed' | 'temporary'

export interface LayoutState {
  device: LayoutDevice
  viewportWidth: number
  viewportHeight: number
  viewportClass: LayoutViewportClass
  sidebarOpened: boolean
  sidebarCollapsed: boolean
  mobileSidebarVisible: boolean
  lastUserCollapseIntent: boolean | null
}

export interface NavigationItem {
  title: string
  path?: string
  name?: string
  exact?: boolean
  iconName?: DoraIconName
  icon?: unknown
  requiredPermissions?: string[]
  children?: NavigationItem[]
  active?: boolean
  expanded?: boolean
  disabled?: boolean
}

export type DoraIconName =
  | 'activity'
  | 'app'
  | 'calendar'
  | 'chart'
  | 'chevron-left'
  | 'chevron-right'
  | 'close'
  | 'close-slash'
  | 'collapse'
  | 'copy'
  | 'database'
  | 'document'
  | 'expand'
  | 'fullscreen'
  | 'fullscreen-exit'
  | 'menu-fold-left'
  | 'menu-fold-right'
  | 'key'
  | 'login'
  | 'logout'
  | 'menu'
  | 'message'
  | 'notification'
  | 'refresh'
  | 'search'
  | 'security'
  | 'setting'
  | 'sun'
  | 'moon'
  | 'team'
  | 'user'
  | 'user-circle'

export interface TableViewPreference {
  tableKey: string
  visibleColumns: string[]
  focused: boolean
  refreshing: boolean
}

export interface VisualThemeState {
  mode: 'light' | 'dark'
  layoutBackground: string
  surfaceBackground: string
  primaryColor: string
  radiusScale: 'md' | 'lg' | 'xl'
  shadowScale: 'none' | 'sm' | 'md'
}

export interface NavigationState {
  activePath: string
  openTabs: OpenTabItem[]
  sidebarCollapsed: boolean
  layout: LayoutState
}

export interface ChartViewState {
  loading: boolean
  empty: boolean
  error: string
}

export interface RichEditorState {
  content: string
  format: 'html' | 'text' | 'markdown'
  dirty: boolean
  errors: string[]
}
