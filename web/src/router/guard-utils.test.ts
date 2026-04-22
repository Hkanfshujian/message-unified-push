import { describe, expect, it } from 'vitest'
import { getFirstAccessibleRoutePath, PROFILE_SETTINGS_PATH, SYSTEM_SETTINGS_PATH, resolveLegacySettingsRedirect } from './guard-utils'

describe('resolveLegacySettingsRedirect', () => {
  it('redirects to system settings when user has system permission', () => {
    expect(resolveLegacySettingsRedirect(true)).toBe(SYSTEM_SETTINGS_PATH)
  })

  it('redirects to profile settings when user has no system permission', () => {
    expect(resolveLegacySettingsRedirect(false)).toBe(PROFILE_SETTINGS_PATH)
  })
})

describe('getFirstAccessibleRoutePath', () => {
  it('returns system settings when dashboard is not accessible but system settings is accessible', () => {
    const path = getFirstAccessibleRoutePath((permissions) => permissions.includes('system:settings:view'))
    expect(path).toBe(SYSTEM_SETTINGS_PATH)
  })

  it('returns empty string when user has no permission', () => {
    const path = getFirstAccessibleRoutePath(() => false)
    expect(path).toBe('')
  })
})
