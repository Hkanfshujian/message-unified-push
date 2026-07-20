import { describe, expect, it } from 'vitest'
import { FIRST_ACCESSIBLE_ROUTE_PRIORITY, getFirstAccessibleRoutePath, isLoginRoute, isNotFoundRoute, LOGIN_PATH, normalizeRequiredPermissions, NOT_FOUND_PATH, PROFILE_SETTINGS_PATH, resolveAuthRedirect, resolvePermissionDeniedRoute, resolveRouteContextPath, resolveStaleRouteFallback, shouldRestoreRouteContext, SYSTEM_SETTINGS_HOME_PATH, resolveLegacySettingsRedirect } from './guard-utils'

describe('resolveLegacySettingsRedirect', () => {
  it('redirects to system settings when user has system permission', () => {
    expect(resolveLegacySettingsRedirect(true)).toBe(SYSTEM_SETTINGS_HOME_PATH)
  })

  it('redirects to profile settings when user has no system permission', () => {
    expect(resolveLegacySettingsRedirect(false)).toBe(PROFILE_SETTINGS_PATH)
  })
})

describe('route guard helpers', () => {
  it('normalizes required permission metadata', () => {
    expect(normalizeRequiredPermissions('dashboard:view')).toEqual(['dashboard:view'])
    expect(normalizeRequiredPermissions(['dashboard:view', '', 1])).toEqual(['dashboard:view'])
    expect(normalizeRequiredPermissions(undefined)).toEqual([])
  })

  it('detects public route names and paths', () => {
    expect(isLoginRoute('/login')).toBe(true)
    expect(isLoginRoute('/')).toBe(false)
    expect(isNotFoundRoute('404')).toBe(true)
    expect(isNotFoundRoute('dashboard')).toBe(false)
  })

  it('resolves auth redirects for public and protected routes', () => {
    expect(resolveAuthRedirect(false, '/templates')).toBe(LOGIN_PATH)
    expect(resolveAuthRedirect(false, LOGIN_PATH)).toBe('')
    expect(resolveAuthRedirect(true, LOGIN_PATH)).toBe('/')
    expect(resolveAuthRedirect(true, '/templates')).toBe('')
  })
})

describe('getFirstAccessibleRoutePath', () => {
  it('does not include removed or unavailable route paths in fallback priority', () => {
    expect(FIRST_ACCESSIBLE_ROUTE_PRIORITY.map(item => item.path)).not.toContain('/system/identities')
  })

  it('returns system settings when dashboard is not accessible but system settings is accessible', () => {
    const path = getFirstAccessibleRoutePath((permissions) => permissions.includes('system:settings:view'))
    expect(path).toBe(SYSTEM_SETTINGS_HOME_PATH)
  })

  it('returns system notifications as an independent route for notification-only users', () => {
    const path = getFirstAccessibleRoutePath((permissions) => permissions.includes('message:system:view'))
    expect(path).toBe('/system/messages')
  })

  it('returns empty string when user has no permission', () => {
    const path = getFirstAccessibleRoutePath(() => false)
    expect(path).toBe('')
  })

  it('prefers dashboard before later accessible routes', () => {
    const path = getFirstAccessibleRoutePath((permissions) => permissions.includes('dashboard:view') || permissions.includes('system:settings:view'))
    expect(path).toBe('/')
  })
})

describe('resolvePermissionDeniedRoute', () => {
  it('falls back from dashboard to first accessible route', () => {
    expect(resolvePermissionDeniedRoute('/', 'dashboard', () => SYSTEM_SETTINGS_HOME_PATH)).toBe(SYSTEM_SETTINGS_HOME_PATH)
  })

  it('falls back from inaccessible routes to the first accessible route', () => {
    expect(resolvePermissionDeniedRoute('/templates', 'templates', () => SYSTEM_SETTINGS_HOME_PATH)).toBe(SYSTEM_SETTINGS_HOME_PATH)
  })

  it('uses 404 when no safe permission fallback exists', () => {
    expect(resolvePermissionDeniedRoute('/', 'dashboard', () => '')).toBe(NOT_FOUND_PATH)
  })
})

describe('route context restoration helpers', () => {
  const routeTabPaths = ['/', '/templates', '/system/users']

  it('detects routes that should restore workspace context', () => {
    expect(shouldRestoreRouteContext('/system/users/detail', routeTabPaths)).toBe(true)
    expect(shouldRestoreRouteContext('/login', routeTabPaths)).toBe(false)
    expect(shouldRestoreRouteContext('/404', routeTabPaths)).toBe(false)
  })

  it('resolves nested route context path with safe fallback', () => {
    expect(resolveRouteContextPath('/system/users/detail', routeTabPaths)).toBe('/system/users')
    expect(resolveRouteContextPath('/missing', routeTabPaths)).toBe('/')
  })

  it('resolves stale route fallback using workspace context before permission fallback', () => {
    expect(resolveStaleRouteFallback('/system/users/detail', routeTabPaths, () => '/templates')).toBe('/system/users')
    expect(resolveStaleRouteFallback('/removed', routeTabPaths, () => '/templates')).toBe('/templates')
  })

  it('keeps nested route restoration independent from visual shell refresh classes', () => {
    const refreshedRouteTabs = ['/', '/logs/task', '/logs/consume', '/data/mq-sources']
    expect(resolveRouteContextPath('/logs/consume/detail/123', refreshedRouteTabs)).toBe('/logs/consume')
    expect(resolveStaleRouteFallback('/data/mq-sources/edit/abc', refreshedRouteTabs, () => '/')).toBe('/data/mq-sources')
  })
})
