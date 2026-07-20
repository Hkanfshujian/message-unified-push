import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { buttonVariants } from './index'

const buttonSource = readFileSync(resolve(__dirname, 'Button.vue'), 'utf-8')
const headerSource = readFileSync(resolve(__dirname, '../../layout/Header.vue'), 'utf-8')
const cssSource = readFileSync(resolve(__dirname, '../../../index.css'), 'utf-8')

describe('button interaction contract', () => {
  it('exposes icon variant for natural header action feedback', () => {
    expect(buttonVariants({ variant: 'icon', size: 'icon' })).toContain('bg-transparent')
    expect(buttonVariants({ variant: 'icon', size: 'icon' })).toContain('hover:bg-[var(--brand-50)]')
  })

  it('covers loading and non-button disabled states', () => {
    expect(buttonSource).toContain('data-loading')
    expect(buttonSource).toContain('aria-busy')
    expect(buttonSource).toContain('aria-disabled')
    expect(buttonSource).toContain('button-spinner')
  })

  it('uses shared button primitive for selected header icon actions', () => {
    expect(headerSource).toContain('import { Button }')
    expect(headerSource).toContain('variant="icon"')
    expect(headerSource).toContain(':aria-pressed="props.theme === \'dark\'"')
    expect(headerSource).toContain('requestThemeToggle')
  })

  it('keeps header icon hover, focus, active and switch animations styled', () => {
    expect(cssSource).toContain('.dora-header-icon[data-slot="button"]::before')
    expect(cssSource).toContain('.admin-header-action:focus-visible')
    expect(cssSource).toContain('.admin-header-action:active')
    expect(cssSource).toContain('@keyframes dora-header-icon-switch')
  })
})
