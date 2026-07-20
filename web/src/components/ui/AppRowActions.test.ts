import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/ui/AppRowActions.vue'), 'utf8')

describe('AppRowActions contract', () => {
  it('uses AppActionButton for direct actions and Element Plus dropdown items for more actions', () => {
    expect(source).toContain('<AppActionButton')
    expect(source).toContain('<el-dropdown-item')
    expect(source).toContain('v-for="action in allocation.direct"')
  })

  it('guards duplicate async clicks and clears pending in finally', () => {
    expect(source).toContain('pendingKeys.value.has(action.key)')
    expect(source).toContain('await action.onClick()')
    expect(source).toContain('} finally {')
    expect(source).toContain('clearedPending.delete(action.key)')
  })

  it('keeps the legacy default and more slot fallback', () => {
    expect(source).toContain('<slot />')
    expect(source).toContain('<slot name="more" />')
  })
})
