import { describe, expect, it } from 'vitest'
import { analyzeTrendSeries, buildChannelChartOption, buildTrendChartOption, createChartThemeTokens, hasChannelChartData, hasTrendChartData } from '@/util/charts'

describe('dashboard chart empty state mapping', () => {
  it('treats empty trend series as empty', () => {
    expect(hasTrendChartData([])).toBe(false)
    expect(hasTrendChartData([
      { day: '06-24', num: 0, day_succ_num: 0, day_failed_num: 0 }
    ])).toBe(false)
  })

  it('detects populated trend series', () => {
    expect(hasTrendChartData([
      { day: '06-24', num: 0, day_succ_num: 1, day_failed_num: 0 }
    ])).toBe(true)
  })

  it('maps channel zero values to empty state', () => {
    expect(hasChannelChartData([])).toBe(false)
    expect(hasChannelChartData([
      { way_name: '企业微信', count_num: 0 }
    ])).toBe(false)
  })

  it('detects populated channel data', () => {
    expect(hasChannelChartData([
      { way_name: '企业微信', count_num: 3 }
    ])).toBe(true)
  })

  it('uses DoraCMS chart colors for trend cards', () => {
    const option = buildTrendChartOption([
      { day: '06-24', num: 5, day_succ_num: 4, day_failed_num: 1 }
    ])
    expect(option.color).toEqual(['#6d8cff', '#43d6b5', '#ff7b6b'])
  })

  it('does not render intrusive peak badges on trend lines', () => {
    const option = buildTrendChartOption([
      { day: '06-24', num: 1, day_succ_num: 1, day_failed_num: 0 },
      { day: '06-25', num: 8, day_succ_num: 7, day_failed_num: 1 }
    ])
    const series = Array.isArray(option.series) ? option.series : []
    expect(series.every(item => !('markPoint' in item))).toBe(true)
  })

  it('derives trend insight summary for control center metrics', () => {
    expect(analyzeTrendSeries([
      { day: '06-24', num: 10, day_succ_num: 8, day_failed_num: 2 },
      { day: '06-25', num: 4, day_succ_num: 4, day_failed_num: 0 }
    ])).toMatchObject({ total: 14, success: 12, failed: 2, successRate: 86, empty: false })
  })

  it('uses compact DoraCMS donut settings and light card border for channel cards', () => {
    const option = buildChannelChartOption([
      { way_name: '企业微信', count_num: 3 }
    ], createChartThemeTokens('light'))
    const series = Array.isArray(option.series) ? option.series[0] : undefined
    expect(option.color).toEqual(['#2f80ed', '#20c997', '#ffb020', '#ff4d5e', '#7c3aed'])
    expect(series).toMatchObject({ radius: ['52%', '74%'], center: ['50%', '50%'], label: { show: false }, itemStyle: { borderColor: '#ffffff', borderWidth: 3 } })
  })

  it('keeps dark chart tokens aligned with Dora dark surfaces', () => {
    expect(createChartThemeTokens('dark')).toMatchObject({
      tooltipBackground: 'rgba(28,28,28,0.96)',
      surfaceBorderColor: '#1c1c1c'
    })
  })
})
