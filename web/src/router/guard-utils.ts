export const LEGACY_SETTINGS_PATH = '/settings'
export const SYSTEM_SETTINGS_PATH = '/system/settings'
export const SYSTEM_SETTINGS_HOME_PATH = '/system/settings/site'
export const PROFILE_SETTINGS_PATH = '/profile/settings'
export const LOGIN_PATH = '/login'
export const NOT_FOUND_PATH = '/404'

export interface RouteAccessRule {
  path: string
  requiredPermissions: string[]
}

export const FIRST_ACCESSIBLE_ROUTE_PRIORITY: RouteAccessRule[] = [
  { path: '/', requiredPermissions: ['dashboard:view'] },
  { path: '/templates', requiredPermissions: ['message:template:view'] },
  { path: '/sendways', requiredPermissions: ['message:sendways:view'] },
  { path: '/cronmessages', requiredPermissions: ['message:cron:view'] },
  { path: '/logs/task', requiredPermissions: ['message:sendlogs:view'] },
  { path: '/system/messages', requiredPermissions: ['message:system:view'] },
  { path: SYSTEM_SETTINGS_HOME_PATH, requiredPermissions: ['system:settings:view'] },
  { path: '/system/roles', requiredPermissions: ['system:rbac:role'] },
  { path: '/system/groups', requiredPermissions: ['system:rbac:group'] },
  { path: '/system/permissions', requiredPermissions: ['system:rbac:permission'] },
  { path: PROFILE_SETTINGS_PATH, requiredPermissions: ['profile:settings:view'] }
]

export const resolveLegacySettingsRedirect = (hasSystemSettingsPermission: boolean): string => {
  if (hasSystemSettingsPermission) {
    return SYSTEM_SETTINGS_HOME_PATH
  }
  return PROFILE_SETTINGS_PATH
}

export const getFirstAccessibleRoutePath = (
  hasAnyPermission: (permissions: string[]) => boolean
): string => {
  const route = FIRST_ACCESSIBLE_ROUTE_PRIORITY.find(item => hasAnyPermission(item.requiredPermissions))
  return route?.path || ''
}

export const normalizeRequiredPermissions = (permissions: unknown): string[] => {
  if (Array.isArray(permissions)) {
    return permissions.filter((item): item is string => typeof item === 'string' && item.trim() !== '')
  }
  if (typeof permissions === 'string' && permissions.trim() !== '') {
    return [permissions]
  }
  return []
}

export const isLoginRoute = (path: string): boolean => path === LOGIN_PATH

export const isNotFoundRoute = (name: unknown): boolean => name === '404'

export const sanitizeInternalRedirect = (value: unknown, fallback = '/'): string => {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//') || value === LOGIN_PATH) return fallback
  return value
}

export const resolveAuthRedirect = (isAuthenticated: boolean, targetPath: string): string => {
  if (!isAuthenticated && !isLoginRoute(targetPath)) return LOGIN_PATH
  if (isAuthenticated && isLoginRoute(targetPath)) return '/'
  return ''
}

export const resolvePermissionDeniedRoute = (
  targetPath: string,
  targetName: unknown,
  getFirstAccessibleRoute: () => string
): string => {
  const fallbackRoute = getFirstAccessibleRoute()
  if (fallbackRoute && fallbackRoute !== targetPath) return fallbackRoute
  if (targetPath === '/' || targetName === 'dashboard') return NOT_FOUND_PATH
  return NOT_FOUND_PATH
}

export const shouldRestoreRouteContext = (targetPath: string, routeTabPaths: string[]) => {
  if (!targetPath || targetPath === LOGIN_PATH || targetPath === NOT_FOUND_PATH) return false
  return routeTabPaths.some(path => path === targetPath || (path !== '/' && targetPath.startsWith(`${path}/`)))
}

export const resolveRouteContextPath = (targetPath: string, routeTabPaths: string[], fallbackPath = '/') => {
  const matched = routeTabPaths
    .filter(path => path === targetPath || (path !== '/' && targetPath.startsWith(`${path}/`)))
    .sort((a, b) => b.length - a.length)[0]
  return matched || fallbackPath
}

export const resolveStaleRouteFallback = (
  targetPath: string,
  routeTabPaths: string[],
  getFirstAccessibleRoute: () => string,
  fallbackPath = '/'
) => {
  if (shouldRestoreRouteContext(targetPath, routeTabPaths)) {
    return resolveRouteContextPath(targetPath, routeTabPaths, fallbackPath)
  }
  return getFirstAccessibleRoute() || fallbackPath
}
