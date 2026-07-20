import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(__dirname, 'PermissionsManagement.vue'), 'utf-8')

describe('permissions management dark toggle visibility', () => {
  it('adds a dedicated class for permission tree expand buttons', () => {
    expect(source).toContain('permission-tree-toggle')
    expect(source).toContain('permission-tree-toggle-icon')
  })

  it('keeps expand buttons visible in dark mode', () => {
    expect(source).toContain('.dark .permission-tree-toggle')
    expect(source).toContain('background: rgba(15, 23, 42, 0.20);')
    expect(source).toContain('.dark .permission-tree-toggle:hover')
    expect(source).toContain('color: #bfdbfe;')
  })
})
