import { describe, expect, it } from 'vitest'
import { createDoraThemeTokens, createGlassThemeTokens, getDoraMaterialClass, getDoraThemeClass, getGlassMaterialClass, getGlassThemeClass } from './glassTheme'

describe('Dora theme tokens', () => {
  it('builds light and dark material levels without glass blur', () => {
    expect(createDoraThemeTokens('light').materials.panel.background).toBe('#ffffff')
    expect(createDoraThemeTokens('dark').materials.panel.background).toBe('#1c1c1c')
    expect(createDoraThemeTokens('light').materials.active.border).toContain('37, 99, 235')
    expect(createDoraThemeTokens('dark').materials.danger.blur).toBe('0px')
  })

  it('exposes fallback and reduced motion state hooks', () => {
    const tokens = createDoraThemeTokens('light')
    expect(tokens.fallbackClass).toBe('dora-no-backdrop-filter')
    expect(tokens.reducedMotionClass).toBe('dora-reduced-motion')
  })

  it('resolves stable material and theme class names', () => {
    expect(getDoraMaterialClass('overlay')).toBe('dora-material-overlay')
    expect(getDoraThemeClass('dark')).toBe('dora-theme-dark')
  })

  it('keeps legacy helper names as Dora-compatible wrappers', () => {
    expect(createGlassThemeTokens('light').materials.panel.background).toBe('#ffffff')
    expect(getGlassMaterialClass('overlay')).toBe('dora-material-overlay')
    expect(getGlassThemeClass('dark')).toBe('dora-theme-dark')
  })
})
