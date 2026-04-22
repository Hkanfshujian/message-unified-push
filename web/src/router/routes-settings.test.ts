import { describe, expect, it } from 'vitest'
import { profileSettingsChildren, systemSettingsChildren } from './settings-route-config'

describe('settings nested routes', () => {
  it('contains system settings child routes with permission meta', () => {
    const siteRoute = systemSettingsChildren.find(route => route.path === 'site')
    const authRoute = systemSettingsChildren.find(route => route.path === 'auth')
    const storageRoute = systemSettingsChildren.find(route => route.path === 'storage')
    const cleanRoute = systemSettingsChildren.find(route => route.path === 'clean')

    expect(siteRoute).toBeTruthy()
    expect(authRoute).toBeTruthy()
    expect(storageRoute).toBeTruthy()
    expect(cleanRoute).toBeTruthy()
    expect(siteRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(authRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(storageRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
  })

  it('contains profile settings child routes with proper permission meta', () => {
    const passwordRoute = profileSettingsChildren.find(route => route.path === 'password')
    const preferenceRoute = profileSettingsChildren.find(route => route.path === 'preference')

    expect(passwordRoute).toBeTruthy()
    expect(preferenceRoute).toBeTruthy()
    expect(passwordRoute?.meta?.requiredPermissions).toEqual(['profile:settings:edit'])
    expect(preferenceRoute?.meta?.requiredPermissions).toEqual(['profile:settings:view'])
  })

  it('contains default redirect routes for system and profile settings', () => {
    const systemDefault = systemSettingsChildren.find(route => route.path === '')
    const profileDefault = profileSettingsChildren.find(route => route.path === '')

    expect(systemDefault?.redirect).toBe('/system/settings/site')
    expect(profileDefault?.redirect).toBe('/profile/settings/password')
  })
})
