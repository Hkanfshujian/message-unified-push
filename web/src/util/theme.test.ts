import { describe, expect, it, beforeEach } from 'vitest'
import { applyTheme } from './theme'

describe('applyTheme', () => {
  beforeEach(() => {
    document.documentElement.removeAttribute('style')
    const styleEl = document.getElementById('dynamic-theme-style')
    if (styleEl) {
      styleEl.remove()
    }
    localStorage.removeItem('themeColor')
  })

  it('applies built-in theme and stores theme key', () => {
    applyTheme('blue')
    const styleEl = document.getElementById('dynamic-theme-style')
    expect(styleEl).not.toBeNull()
    expect(styleEl?.textContent).toContain('--brand')
    expect(localStorage.getItem('themeColor')).toBe('blue')
  })

  it('applies custom theme and stores custom key', () => {
    applyTheme('custom:#123456')
    const styleEl = document.getElementById('dynamic-theme-style')
    expect(styleEl?.textContent).toContain('#123456')
    expect(localStorage.getItem('themeColor')).toBe('custom:#123456')
  })
})

