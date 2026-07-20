import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const cssSource = readFileSync(resolve(__dirname, '../../index.css'), 'utf-8')

describe('dark mode surface regression guards', () => {
  it('covers Element Plus surfaces and table states with dark overrides', () => {
    expect(cssSource).toContain('.dark .el-card')
    expect(cssSource).toContain('.dark .el-table .el-table__body tr.hover-row > td.el-table__cell')
    expect(cssSource).toContain('background-color: rgba(59, 130, 246, 0.10) !important;')
    expect(cssSource).toContain('.dark .el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell')
    expect(cssSource).toContain('.dark .el-descriptions__label.el-descriptions__cell')
  })

  it('covers poppers, inputs, tags and pagination with non-white dark treatments', () => {
    expect(cssSource).toContain('.dark .el-select-dropdown__item:hover')
    expect(cssSource).toContain('.dark .el-input__wrapper')
    expect(cssSource).toContain('.dark .el-switch__core')
    expect(cssSource).toContain('.dark .el-checkbox__input .el-checkbox__inner')
    expect(cssSource).toContain('.dark .el-radio__inner')
    expect(cssSource).toContain('.dark .el-segmented')
    expect(cssSource).toContain('.dark [data-slot="switch"][data-state="checked"]')
    expect(cssSource).toContain('.dark .el-tag--success.el-tag--light')
    expect(cssSource).toContain('.dark .el-tag--info.el-tag--plain')
    expect(cssSource).toContain('.dark .app-pagination-button.is-active')
  })

  it('covers page-level hover hotspots that previously used light brand blocks', () => {
    expect(cssSource).toContain('.dark .el-alert--info')
    expect(cssSource).toContain('.dark .el-statistic__content')
    expect(cssSource).toContain('.dark .el-message')
    expect(cssSource).toContain('.dark .storage-table-data-row:hover')
    expect(cssSource).toContain('.dark .storage-icon-action')
    expect(cssSource).toContain('.dark .storage-breadcrumb-button:hover')
    expect(cssSource).toContain('.dark .app-form-drawer .el-drawer__header')
    expect(cssSource).toContain('.dark .settings-menu-button:not(.el-button--primary):not(.el-button--danger):hover')
  })

  it('keeps system settings nav active text and icon readable in dark mode', () => {
    expect(cssSource).toContain('.dark .app-settings-nav-item-active')
    expect(cssSource).toContain('.dark .app-settings-nav-item-active span')
    expect(cssSource).toContain('.dark .app-settings-nav-item-active:hover')
    expect(cssSource).toContain('color: #bfdbfe !important;')
  })
})
