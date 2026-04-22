export const LEGACY_SETTINGS_PATH = '/settings'
export const SYSTEM_SETTINGS_PATH = '/system/settings'
export const PROFILE_SETTINGS_PATH = '/profile/settings'

export interface RouteAccessRule {
  path: string
  requiredPermissions: string[]
}

export const FIRST_ACCESSIBLE_ROUTE_PRIORITY: RouteAccessRule[] = [
  { path: '/', requiredPermissions: ['dashboard:view'] },
  { path: '/templates', requiredPermissions: ['message:template:view'] },
  { path: '/sendways', requiredPermissions: ['message:sendways:view'] },
  { path: '/cronmessages', requiredPermissions: ['message:cron:view'] },
  { path: '/sendlogs', requiredPermissions: ['message:sendlogs:view'] },
  { path: SYSTEM_SETTINGS_PATH, requiredPermissions: ['system:settings:view'] },
  { path: '/system/roles', requiredPermissions: ['system:rbac:role'] },
  { path: '/system/groups', requiredPermissions: ['system:rbac:group'] },
  { path: '/system/permissions', requiredPermissions: ['system:rbac:permission'] },
  { path: '/system/identities', requiredPermissions: ['system:rbac:identity'] },
  { path: PROFILE_SETTINGS_PATH, requiredPermissions: ['profile:settings:view'] }
]

export const resolveLegacySettingsRedirect = (hasSystemSettingsPermission: boolean): string => {
  if (hasSystemSettingsPermission) {
    return SYSTEM_SETTINGS_PATH
  }
  return PROFILE_SETTINGS_PATH
}

export const getFirstAccessibleRoutePath = (
  hasAnyPermission: (permissions: string[]) => boolean
): string => {
  const route = FIRST_ACCESSIBLE_ROUTE_PRIORITY.find(item => hasAnyPermission(item.requiredPermissions))
  return route?.path || ''
}
