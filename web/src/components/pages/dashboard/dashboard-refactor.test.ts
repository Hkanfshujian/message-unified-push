import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { zhCN } from '@/locales/zh-CN'

const dashboardSource = readFileSync(resolve(__dirname, 'Dashboard.vue'), 'utf-8')
const sidebarItemSource = readFileSync(resolve(__dirname, '../../layout/SidebarItem.vue'), 'utf-8')
const sidebarButtonSource = readFileSync(resolve(__dirname, '../../layout/SidebarNavButton.vue'), 'utf-8')
const metricCardSource = readFileSync(resolve(__dirname, 'CardNum.vue'), 'utf-8')
const channelPanelSource = readFileSync(resolve(__dirname, 'DashboardChannelPanel.vue'), 'utf-8')
const cssSource = readFileSync(resolve(__dirname, '../../../index.css'), 'utf-8')

describe('dashboard and sidebar refactor architecture', () => {
  it('splits dashboard shell into focused presentation panels', () => {
    expect(dashboardSource).toContain('dashboard-toolbar')
    expect(dashboardSource).not.toContain('DashboardHero')
    expect(dashboardSource).not.toContain('dashboard-toolbar-title-row')
    expect(dashboardSource).toContain('<DoraIcon name="activity"')
    expect(dashboardSource).toContain('dashboard-hero-status')
    expect(dashboardSource).toContain('DashboardTrendPanel')
    expect(dashboardSource).toContain('DashboardChannelPanel')
    expect(dashboardSource).not.toContain('DashboardActivityPanel')
    expect(dashboardSource).not.toContain('dashboard-message-grid')
  })

  it('delegates sidebar button behavior to an accessible nav button component', () => {
    expect(sidebarItemSource).toContain('SidebarNavButton')
    expect(sidebarButtonSource).toContain(':aria-label="buttonLabel"')
    expect(sidebarButtonSource).toContain(':aria-current="isActive && !hasChildren ? \'page\' : undefined"')
    expect(sidebarButtonSource).toContain('const ariaExpanded = computed')
    expect(sidebarButtonSource).toContain(':aria-expanded="ariaExpanded"')
  })

  it('uses native button semantics for interactive metric cards', () => {
    expect(metricCardSource).toContain('<button')
    expect(metricCardSource).toContain(':aria-label=')
  })

  it('keeps metric cards focused on values without decorative sparklines', () => {
    expect(metricCardSource).not.toContain('sparklinePath')
    expect(metricCardSource).not.toContain('<svg')
    expect(dashboardSource).toContain('deriveCoreMetrics(state.basicData, normalizedTrendData.value')
  })

  it('renders accessible channel distribution, ranking, progress, and permission states without fake filters', () => {
    expect(channelPanelSource).not.toContain('dashboard-channel-filter')
    expect(channelPanelSource).toContain('dashboard-channel-center')
    expect(channelPanelSource).toContain('dashboard-channel-tooltip')
    expect(channelPanelSource).toContain('showSegmentTooltip')
    expect(channelPanelSource).toContain('tabindex="0"')
    expect(channelPanelSource).toContain('@focus="showSegmentTooltip(item, $event)"')
    expect(channelPanelSource).toContain('dashboard-channel-progress')
    expect(channelPanelSource).toContain(':disabled="!canViewLogs || !item.action"')
    expect(channelPanelSource).toContain('v-if="canViewChannels"')
    expect(channelPanelSource).toContain('messages.noChannelPermission')
    expect(zhCN.dashboard.noChannelPermission).toBe('无渠道管理权限')
    expect(channelPanelSource).not.toContain('fallbackChannels')
    expect(cssSource).toContain('--dashboard-panel-content-height: 176px')
    expect(cssSource).toContain('grid-template-columns: 180px minmax(0, 1fr)')
    expect(cssSource).toContain('max-height: 178px')
    expect(cssSource).toContain('overflow-y: auto')
  })

  it('defines responsive and reduced-motion dashboard behavior', () => {
    expect(cssSource).toContain('@media (max-width: 1280px)')
    expect(cssSource).toContain('@media (max-width: 768px)')
    expect(cssSource).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
