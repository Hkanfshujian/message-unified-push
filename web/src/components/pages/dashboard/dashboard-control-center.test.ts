import { describe, expect, it } from 'vitest'
import { deriveActionableEvents, deriveChannelInsight, deriveCoreMetrics, deriveDashboardHealthSummary, normalizeDashboardFilterContext } from '@/util/dashboardControlCenter'

const context = normalizeDashboardFilterContext('2026-06-24', '2026-06-25', 15)
const trendData = [
  { day: '06-24', num: 10, day_succ_num: 9, day_failed_num: 1 },
  { day: '06-25', num: 5, day_succ_num: 5, day_failed_num: 0 }
]

describe('dashboard control center view models', () => {
  it('normalizes date range context for drill-down query usage', () => {
    expect(context).toMatchObject({ rangeLabel: '2026-06-24 至 2026-06-25', trendDays: 2 })
    expect(context.routeQuery).toMatchObject({ start_time: '2026-06-24 00:00:00', end_time: '2026-06-25 23:59:59' })
  })

  it('derives warning health summary and failure action from real trend data', () => {
    const summary = deriveDashboardHealthSummary(trendData, context, '12:00:00')
    expect(summary.status).toBe('warning')
    expect(summary.primaryAction?.path).toContain('%22status%22')
  })

  it('derives four scoped metrics without today-only wording', () => {
    const metrics = deriveCoreMetrics({ message_total_num: 100, today_total_num: 0, today_succ_num: 0, today_failed_num: 0 }, trendData, [{ way_name: '企业微信', count_num: 3 }], context)
    expect(metrics).toHaveLength(4)
    expect(metrics.map(item => item.scopeLabel)).toContain('2026-06-24 至 2026-06-25')
  })

  it('derives real channel ranking without fallback channels', () => {
    const insight = deriveChannelInsight([{ way_name: '企业微信', count_num: 3 }], context)
    expect(insight.rankedChannels).toHaveLength(1)
    expect(insight.primaryChannel?.percent).toBe(100)
  })

  it('keeps all real channels sorted by usage instead of truncating to top five', () => {
    const insight = deriveChannelInsight([
      { way_name: 'A', count_num: 10 },
      { way_name: 'B', count_num: 9 },
      { way_name: 'C', count_num: 8 },
      { way_name: 'D', count_num: 7 },
      { way_name: 'E', count_num: 6 },
      { way_name: 'F', count_num: 5 }
    ], context)
    expect(insight.rankedChannels.map(item => item.name)).toEqual(['A', 'B', 'C', 'D', 'E', 'F'])
  })

  it('derives actionable events from health and channel insights', () => {
    const health = deriveDashboardHealthSummary(trendData, context, '12:00:00')
    const channel = deriveChannelInsight([{ way_name: '企业微信', count_num: 3 }], context)
    const events = deriveActionableEvents(health, channel, trendData, context)
    expect(events.some(item => item.action?.path.includes('/logs/task'))).toBe(true)
  })
})
