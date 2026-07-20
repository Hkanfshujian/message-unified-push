import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { zhCN } from '@/locales/zh-CN'

const runtimePanelSource = readFileSync(resolve(__dirname, 'DashboardRuntimePanel.vue'), 'utf-8')
const cssSource = readFileSync(resolve(__dirname, '../../../index.css'), 'utf-8')

describe('dashboard runtime overview styling', () => {
  it('uses semantic list roles for runtime metric tiles', () => {
    expect(runtimePanelSource).toContain('role="list"')
    expect(runtimePanelSource).toContain('role="listitem"')
    expect(runtimePanelSource).toContain(':aria-label="messages.runtimeMetrics"')
    expect(zhCN.dashboard.runtimeMetrics).toBe('站点运行核心指标')
  })

  it('uses adaptive runtime grid columns for narrow side panels', () => {
    expect(cssSource).toContain('grid-template-columns: repeat(auto-fit, minmax(92px, 1fr));')
  })

  it('keeps runtime tiles aligned with dark dashboard surfaces', () => {
    expect(cssSource).toContain('.dark .dashboard-runtime-tile')
    expect(cssSource).toContain('background: linear-gradient(180deg, rgba(30, 41, 59, 0.58), rgba(15, 23, 42, 0.34));')
    expect(cssSource).toContain('.dark .dashboard-runtime-blue .dashboard-runtime-value')
    expect(cssSource).toContain('.dark .dashboard-runtime-green .dashboard-runtime-value')
  })
})
