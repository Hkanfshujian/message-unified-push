import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(__dirname, 'AppFormDialog.vue'), 'utf-8')

describe('AppFormDialog contract', () => {
  it('delegates cancel behavior and action placement to AppFormDrawer', () => {
    expect(source).toContain("emit('cancel')")
    expect(source).toContain('AppFormDrawer')
    expect(source).toContain(':confirm-text="confirmText"')
    expect(source).toContain(':cancel-text="cancelText"')
  })

  it('passes a single-line title through the drawer wrapper', () => {
    expect(source).toContain(':title="title"')
    expect(source).not.toContain('description')
    expect(source).toContain('closeOnClickModal')

    const drawerSource = readFileSync(resolve(__dirname, 'AppFormDrawer.vue'), 'utf-8')
    expect(drawerSource).not.toContain('description')
    expect(drawerSource).toContain('append-to-body')
    expect(drawerSource).toContain("bodyMode?: 'scroll' | 'managed'")
    expect(drawerSource).toContain("density?: 'default' | 'compact'")
    expect(drawerSource).toContain('v-if="showFooter || $slots.footer"')
  })

  it('keeps compatibility props while routing rendering to right-side drawer', () => {
    expect(source).toContain(':size="width"')
    expect(source).toContain(':loading="loading"')
    expect(source).toContain(':show-footer="showFooter"')
    expect(source).toContain('@confirm="emit(\'confirm\')"')
  })

  it('remains a compatibility wrapper rather than a direct page-level migration dependency', () => {
    const sendWaysSource = readFileSync(resolve(__dirname, '../pages/sendWays/SendWays.vue'), 'utf-8')

    expect(sendWaysSource).toContain('AppFormDrawer')
    expect(sendWaysSource).not.toContain('AppFormDialog')
  })
})
