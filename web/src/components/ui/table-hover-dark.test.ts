import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const cssSource = readFileSync(resolve(__dirname, '../../index.css'), 'utf-8')

describe('dark table hover styling', () => {
  it('uses low-alpha dark hover colors instead of bright row backgrounds', () => {
    expect(cssSource).toContain('--el-table-row-hover-bg-color: rgba(59, 130, 246, 0.10);')
    expect(cssSource).toContain('.dark .app-data-table .el-table__body tr.hover-row > td.el-table__cell')
    expect(cssSource).toContain('background: rgba(59, 130, 246, 0.10) !important;')
  })

  it('keeps fixed columns and tags readable in dark table hover states', () => {
    expect(cssSource).toContain('.dark .app-data-table .el-table__body tr:hover > td.el-table-fixed-column--right')
    expect(cssSource).toContain('background: rgb(24, 34, 52) !important;')
    expect(cssSource).toContain('.dark .app-data-table .el-tag--light')
    expect(cssSource).toContain('background: rgba(59, 130, 246, 0.13);')
  })
})
