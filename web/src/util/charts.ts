import type { EChartsOption, SeriesOption } from 'echarts'

export interface SendTrendPoint {
  day: string
  day_failed_num: number
  day_succ_num: number
  num: number
  succ_num?: number
}

export interface ChannelStatPoint {
  count_num: number
  way_name: string
}

export interface ChartThemeTokens {
  axisTextColor: string
  mutedTextColor: string
  gridLineColor: string
  tooltipBackground: string
  tooltipBorderColor: string
  surfaceBorderColor: string
}

type DashboardLineSeriesOption = Extract<SeriesOption, { type?: 'line' }>

export interface TrendDateRange {
  start: string
  end: string
}

export interface TrendInsightSummary {
  total: number
  success: number
  failed: number
  successRate: number
  peakPoint?: SendTrendPoint
  failurePeakPoint?: SendTrendPoint
  empty: boolean
}

export const defaultChartThemeTokens: ChartThemeTokens = {
  axisTextColor: '#8a92a6',
  mutedTextColor: '#6f7b95',
  gridLineColor: 'rgba(229,234,242,0.96)',
  tooltipBackground: '#ffffff',
  tooltipBorderColor: 'rgba(229,234,242,0.96)',
  surfaceBorderColor: '#ffffff'
}

export const createChartThemeTokens = (mode: 'light' | 'dark'): ChartThemeTokens => {
  if (mode === 'dark') {
    return {
      axisTextColor: '#a8b5ca',
      mutedTextColor: '#93a4bd',
      gridLineColor: 'rgba(255,255,255,0.08)',
      tooltipBackground: 'rgba(28,28,28,0.96)',
      tooltipBorderColor: 'rgba(255,255,255,0.10)',
      surfaceBorderColor: '#1c1c1c'
    }
  }

  return { ...defaultChartThemeTokens }
}

export const toIsoDate = (date: Date) => {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

export const parseTrendDayToIso = (input: string, fallbackYear: number) => {
  const text = String(input || '').trim()
  if (!text) return ''
  if (/^\d{4}-\d{2}-\d{2}$/.test(text)) return text
  if (/^\d{2}-\d{2}$/.test(text)) return `${fallbackYear}-${text}`
  const parsed = new Date(text)
  return Number.isNaN(parsed.getTime()) ? '' : toIsoDate(parsed)
}

export const normalizeTrendSeries = (source: SendTrendPoint[], range: TrendDateRange) => {
  const fallbackYear = Number(range.end.slice(0, 4)) || new Date().getFullYear()
  const mapped = new Map<string, SendTrendPoint>()
  source.forEach(item => {
    const key = parseTrendDayToIso(item.day, fallbackYear)
    if (!key) return
    mapped.set(key, {
      day: key,
      day_failed_num: Number(item.day_failed_num || 0),
      day_succ_num: Number(item.day_succ_num || 0),
      num: Number(item.num || 0),
      succ_num: Number(item.succ_num || 0)
    })
  })

  const result: SendTrendPoint[] = []
  const cursor = new Date(range.start)
  const endDate = new Date(range.end)
  while (!Number.isNaN(cursor.getTime()) && !Number.isNaN(endDate.getTime()) && cursor <= endDate) {
    const key = toIsoDate(cursor)
    const current = mapped.get(key) || { day: key, day_failed_num: 0, day_succ_num: 0, num: 0, succ_num: 0 }
    result.push({ ...current, day: key.slice(5) })
    cursor.setDate(cursor.getDate() + 1)
  }
  return result
}

export const hasTrendChartData = (data: SendTrendPoint[]) => {
  return data.some(item => Number(item.num || 0) > 0 || Number(item.day_succ_num || 0) > 0 || Number(item.day_failed_num || 0) > 0)
}

export const hasChannelChartData = (data: ChannelStatPoint[]) => {
  return data.some(item => Number(item.count_num || 0) > 0)
}

export const analyzeTrendSeries = (data: SendTrendPoint[]): TrendInsightSummary => {
  const total = data.reduce((sum, item) => sum + Number(item.num || 0), 0)
  const success = data.reduce((sum, item) => sum + Number(item.day_succ_num || 0), 0)
  const failed = data.reduce((sum, item) => sum + Number(item.day_failed_num || 0), 0)
  const peakPoint = data.reduce<SendTrendPoint | undefined>((current, item) => Number(item.num || 0) > Number(current?.num || 0) ? item : current, undefined)
  const failurePeakPoint = data.reduce<SendTrendPoint | undefined>((current, item) => Number(item.day_failed_num || 0) > Number(current?.day_failed_num || 0) ? item : current, undefined)
  return {
    total,
    success,
    failed,
    successRate: total ? Math.round((success / total) * 100) : 0,
    peakPoint,
    failurePeakPoint,
    empty: total === 0 && success === 0 && failed === 0
  }
}

export const buildTrendChartOption = (data: SendTrendPoint[], tokens: ChartThemeTokens = defaultChartThemeTokens): EChartsOption => {
  const insight = analyzeTrendSeries(data)
  const series: DashboardLineSeriesOption[] = [
    {
      name: '发送总数',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 7,
      areaStyle: { opacity: 0.16 },
      data: data.map(item => item.num || 0)
    },
    {
      name: '发送成功数',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      areaStyle: { opacity: 0.12 },
      data: data.map(item => item.day_succ_num || 0)
    },
    {
      name: '发送失败数',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      lineStyle: { width: 3 },
      areaStyle: { opacity: insight.failed > 0 ? 0.16 : 0.08 },
      data: data.map(item => item.day_failed_num || 0)
    }
  ]

  return {
    color: ['#6d8cff', '#43d6b5', '#ff7b6b'],
    tooltip: {
      trigger: 'axis',
      valueFormatter: value => `${value} 条`,
      backgroundColor: tokens.tooltipBackground,
      borderColor: tokens.tooltipBorderColor,
      textStyle: { color: tokens.axisTextColor },
      extraCssText: 'box-shadow:0 12px 28px rgba(15,23,42,.14);border-radius:10px;'
    },
    legend: {
      top: 0,
      left: 0,
      textStyle: { color: tokens.mutedTextColor }
    },
    grid: { top: 48, right: 16, bottom: 28, left: 48 },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data.map(item => item.day),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: tokens.axisTextColor }
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLabel: { color: tokens.axisTextColor, formatter: '{value} 条' },
      splitLine: { lineStyle: { color: tokens.gridLineColor } }
    },
    series
  }
}

export const buildChannelChartOption = (data: ChannelStatPoint[], tokens: ChartThemeTokens = defaultChartThemeTokens): EChartsOption => {
  return {
    color: ['#2f80ed', '#20c997', '#ffb020', '#ff4d5e', '#7c3aed'],
    tooltip: {
      trigger: 'item',
      formatter: '{b}<br/>{c} 条 ({d}%)',
      backgroundColor: tokens.tooltipBackground,
      borderColor: tokens.tooltipBorderColor,
      textStyle: { color: tokens.axisTextColor }
    },
    series: [
      {
        name: '发送渠道',
        type: 'pie',
        radius: ['52%', '74%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 8, borderColor: tokens.surfaceBorderColor, borderWidth: 3 },
        label: { show: false },
        labelLine: { show: false },
        data: data.map(item => ({ name: item.way_name || '未知渠道', value: Number(item.count_num || 0) }))
      }
    ]
  }
}
