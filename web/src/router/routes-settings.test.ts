import { describe, expect, it } from 'vitest'
import { profileSettingsChildren, systemSettingsChildren } from './settings-route-config'

describe('settings nested routes', () => {
  it('contains system settings child routes with permission meta', () => {
    const siteRoute = systemSettingsChildren.find(route => route.path === 'site')
    const authRoute = systemSettingsChildren.find(route => route.path === 'auth')
    const storageRoute = systemSettingsChildren.find(route => route.path === 'storage')
    const cleanRoute = systemSettingsChildren.find(route => route.path === 'clean')
    const mqPolicyRoute = systemSettingsChildren.find(route => route.path === 'mq-status-policy')
    const tokenToolRoute = systemSettingsChildren.find(route => route.path === 'token-tool')
    const aboutRoute = systemSettingsChildren.find(route => route.path === 'about')

    expect(siteRoute).toBeTruthy()
    expect(authRoute).toBeTruthy()
    expect(storageRoute).toBeTruthy()
    expect(cleanRoute).toBeTruthy()
    expect(mqPolicyRoute).toBeTruthy()
    expect(tokenToolRoute).toBeTruthy()
    expect(aboutRoute).toBeTruthy()
    expect(siteRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(authRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(storageRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(cleanRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(mqPolicyRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(tokenToolRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
    expect(aboutRoute?.meta?.requiredPermissions).toEqual(['system:settings:view'])
  })

  it('does not contain system notifications inside settings shell', () => {
    expect(systemSettingsChildren.find(route => route.path === 'messages')).toBeUndefined()
  })

  it('contains profile settings child routes with proper permission meta', () => {
    const passwordRoute = profileSettingsChildren.find(route => route.path === 'password')

    expect(passwordRoute).toBeTruthy()
    expect(passwordRoute?.meta?.requiredPermissions).toEqual(['profile:settings:edit'])
  })

  it('contains default redirect routes for system and profile settings', () => {
    const systemDefault = systemSettingsChildren.find(route => route.path === '')
    const profileDefault = profileSettingsChildren.find(route => route.path === '')

    expect(typeof systemDefault?.redirect).toBe('function')
    expect(profileDefault?.redirect).toBe('/profile/settings/password')
  })

  it('uses lazy components for nested settings pages', () => {
    const nestedRoutes = [...systemSettingsChildren, ...profileSettingsChildren].filter(route => route.path !== '')

    expect(nestedRoutes.every(route => typeof route.component === 'function')).toBe(true)
  })
})
