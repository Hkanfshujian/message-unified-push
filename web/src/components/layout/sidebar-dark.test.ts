import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const cssSource = readFileSync(resolve(__dirname, '../../index.css'), 'utf-8')

describe('dark sidebar navigation styling', () => {
  it('uses dark-specific active and hover treatments instead of light brand blocks', () => {
    expect(cssSource).toContain('.dark .sidebar-nav-active')
    expect(cssSource).toContain('linear-gradient(90deg, rgba(37, 99, 235, 0.28), rgba(59, 130, 246, 0.11))')
    expect(cssSource).toContain('.dark .sidebar-nav-idle:hover')
    expect(cssSource).toContain('background: rgba(59, 130, 246, 0.10);')
  })

  it('keeps the global dark hover override away from sidebar navigation', () => {
    const globalOverrideStart = cssSource.indexOf('.dark .admin-tab-item-idle:hover')
    expect(globalOverrideStart).toBeGreaterThan(-1)
    const globalOverride = cssSource.slice(globalOverrideStart, globalOverrideStart + 260)
    expect(globalOverride).not.toContain('.dark .sidebar-nav-idle:hover')
  })
})
