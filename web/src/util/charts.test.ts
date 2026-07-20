import { describe, expect, it } from 'vitest'
import { buildChannelChartOption, buildTrendChartOption, createChartThemeTokens, normalizeTrendSeries, parseTrendDayToIso } from './charts'

describe('chart option builders', () => {
  it('normalizes trend date labels across the requested range', () => {
    const normalized = normalizeTrendSeries([
      { day: '2026-06-24', num: 8, day_succ_num: 7, day_failed_num: 1 },
      { day: '06-25', num: 3, day_succ_num: 3, day_failed_num: 0 }
    ], { start: '2026-06-23', end: '2026-06-25' })

    expect(normalized.map(item => item.day)).toEqual(['06-23', '06-24', '06-25'])
    expect(normalized.map(item => item.num)).toEqual([0, 8, 3])
  })

  it('builds trend series for total, success, and failed counts', () => {
    const option = buildTrendChartOption([
      { day: '06-24', num: 8, day_succ_num: 7, day_failed_num: 1 }
    ])

    expect(option.xAxis).toMatchObject({ type: 'category', data: ['06-24'] })
    expect(option.series).toMatchObject([
      { name: '发送总数', type: 'line', data: [8] },
      { name: '发送成功数', type: 'line', data: [7] },
      { name: '发送失败数', type: 'line', data: [1] }
    ])
  })

  it('builds light and dark chart theme tokens', () => {
    expect(createChartThemeTokens('light')).toMatchObject({
      axisTextColor: '#8a92a6',
      surfaceBorderColor: '#ffffff'
    })
    expect(createChartThemeTokens('dark')).toMatchObject({
      axisTextColor: '#a8b5ca',
      tooltipBackground: 'rgba(28,28,28,0.96)'
    })
  })

  it('builds channel distribution pie data with fallback names', () => {
    const option = buildChannelChartOption([
      { way_name: '企业微信', count_num: 5 },
      { way_name: '', count_num: 2 }
    ])

    expect(option.series).toMatchObject([
      {
        type: 'pie',
        data: [
          { name: '企业微信', value: 5 },
          { name: '未知渠道', value: 2 }
        ]
      }
    ])
  })

  it('parses supported trend day formats', () => {
    expect(parseTrendDayToIso('2026-06-25', 2026)).toBe('2026-06-25')
    expect(parseTrendDayToIso('06-25', 2026)).toBe('2026-06-25')
    expect(parseTrendDayToIso('', 2026)).toBe('')
  })
})
