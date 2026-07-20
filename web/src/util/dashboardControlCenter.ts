import { toIsoDate, type ChannelStatPoint, type SendTrendPoint } from '@/util/charts'
import type { ActionableEvent, ChannelInsight, CoreMetric, DashboardActionTarget, DashboardFilterContext, DashboardHealthSummary, RankedChannelInsight } from '@/types/business'

export interface DashboardBasicSource {
  message_total_num: number
  today_total_num: number
  today_succ_num: number
  today_failed_num: number
}

const channelColors = ['#2f80ed', '#20c997', '#ffb020', '#ff4d5e', '#7c3aed', '#06b6d4', '#6366f1', '#f97316', '#64748b']

const safeNumber = (value: unknown) => {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed : 0
}

const parseDate = (value: string) => {
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? null : parsed
}

const formatDateTime = (date: Date) => {
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export const normalizeDashboardFilterContext = (startDate: string, endDate: string, fallbackDays = 15): DashboardFilterContext => {
  const today = toIsoDate(new Date())
  const normalizedEnd = parseDate(endDate) ? endDate : today
  const end = parseDate(normalizedEnd) || new Date()
  const start = parseDate(startDate)
  const fallbackStart = new Date(end)
  fallbackStart.setDate(fallbackStart.getDate() - Math.max(fallbackDays - 1, 0))
  const normalizedStart = start && start <= end ? startDate : toIsoDate(fallbackStart)
  const startMs = parseDate(normalizedStart)?.getTime() || end.getTime()
  const endMs = end.getTime()
  const trendDays = Math.min(Math.max(Math.floor((endMs - startMs) / 86400000) + 1, 1), 90)
  return {
    startDate: normalizedStart,
    endDate: normalizedEnd,
    rangeLabel: `${normalizedStart} 至 ${normalizedEnd}`,
    trendDays,
    routeQuery: {
      start_time: `${normalizedStart} 00:00:00`,
      end_time: `${normalizedEnd} 23:59:59`
    }
  }
}

export const buildTaskLogAction = (context: DashboardFilterContext, extraQuery: Record<string, string> = {}, label = '查看任务日志'): DashboardActionTarget => ({
  label,
  path: `/logs/task?query=${encodeURIComponent(JSON.stringify({ ...context.routeQuery, ...extraQuery }))}`,
  permission: 'message:sendlogs:view'
})

export const buildChannelAction = (context: DashboardFilterContext, channelName?: string): DashboardActionTarget => ({
  label: channelName ? '查看渠道日志' : '查看全部渠道',
  path: channelName ? `/logs/task?query=${encodeURIComponent(JSON.stringify({ ...context.routeQuery, name: channelName }))}` : '/sendways',
  permission: channelName ? 'message:sendlogs:view' : 'message:sendways:view'
})

export const calculateTrendTotals = (trendData: SendTrendPoint[]) => {
  const total = trendData.reduce((sum, item) => sum + safeNumber(item.num), 0)
  const success = trendData.reduce((sum, item) => sum + safeNumber(item.day_succ_num), 0)
  const failed = trendData.reduce((sum, item) => sum + safeNumber(item.day_failed_num), 0)
  const successRate = total ? Math.round((success / total) * 100) : 0
  return { total, success, failed, successRate }
}

export const deriveDashboardHealthSummary = (trendData: SendTrendPoint[], context: DashboardFilterContext, lastUpdatedAt: string): DashboardHealthSummary => {
  const totals = calculateTrendTotals(trendData)
  if (!totals.total) {
    return {
      status: 'empty',
      title: '当前范围暂无推送数据',
      description: `所选 ${context.rangeLabel} 内没有发送记录，可调整时间范围或检查任务配置。`,
      lastUpdatedAt,
      rangeLabel: context.rangeLabel
    }
  }
  if (totals.failed > 0) {
    return {
      status: totals.successRate >= 90 ? 'warning' : 'critical',
      title: totals.successRate >= 90 ? '推送运行存在少量失败' : '推送运行需要立即关注',
      description: `当前范围发送 ${totals.total} 条，失败 ${totals.failed} 条，成功率 ${totals.successRate}%。`,
      lastUpdatedAt,
      rangeLabel: context.rangeLabel,
      primaryAction: buildTaskLogAction(context, { status: '0' }, '查看失败日志')
    }
  }
  return {
    status: 'healthy',
    title: '推送运行稳定',
    description: `当前范围发送 ${totals.total} 条，成功率 ${totals.successRate}%，暂未发现失败堆积。`,
    lastUpdatedAt,
    rangeLabel: context.rangeLabel,
    primaryAction: buildTaskLogAction(context, {}, '查看发送明细')
  }
}

export const deriveCoreMetrics = (basicData: DashboardBasicSource, trendData: SendTrendPoint[], channels: ChannelStatPoint[], context: DashboardFilterContext): CoreMetric[] => {
  const totals = calculateTrendTotals(trendData)
  const activeChannels = channels.filter(item => safeNumber(item.count_num) > 0).length
  const totalSparkline = trendData.map(item => safeNumber(item.num))
  const successSparkline = trendData.map(item => safeNumber(item.day_succ_num))
  const failedSparkline = trendData.map(item => safeNumber(item.day_failed_num))
  const channelSparkline = totalSparkline.length ? totalSparkline.map(() => activeChannels) : []
  const cumulative = safeNumber(basicData.message_total_num)
  return [
    {
      id: 'range-total',
      label: '范围发送',
      value: totals.total,
      unit: '条',
      scopeLabel: context.rangeLabel,
      description: `统计 ${context.trendDays} 天`,
      trendText: totals.total ? `峰值 ${Math.max(...totalSparkline)} 条` : '暂无发送',
      trendType: 'flat',
      tone: 'blue',
      iconName: 'chart',
      sparklineData: totalSparkline,
      action: buildTaskLogAction(context)
    },
    {
      id: 'success-rate',
      label: '成功率',
      value: `${totals.successRate}%`,
      scopeLabel: context.rangeLabel,
      description: `成功 ${totals.success} 条`,
      trendText: totals.failed ? `失败 ${totals.failed} 条` : '无失败记录',
      trendType: totals.failed ? 'down' : 'up',
      tone: 'green',
      iconName: 'security',
      sparklineData: successSparkline,
      action: buildTaskLogAction(context, { status: '1' }, '查看成功日志')
    },
    {
      id: 'failed-count',
      label: '失败待查',
      value: totals.failed,
      unit: '条',
      scopeLabel: context.rangeLabel,
      description: totals.failed ? '需要排查失败明细' : '当前无需处理',
      trendText: totals.failed ? '点击处理' : '运行稳定',
      trendType: totals.failed ? 'down' : 'flat',
      tone: totals.failed ? 'red' : 'slate',
      iconName: 'close',
      sparklineData: failedSparkline,
      action: totals.failed ? buildTaskLogAction(context, { status: '0' }, '查看失败日志') : undefined
    },
    {
      id: 'active-channels',
      label: '活跃渠道',
      value: activeChannels || cumulative,
      unit: activeChannels ? '个' : '条累计',
      scopeLabel: activeChannels ? context.rangeLabel : '累计统计',
      description: activeChannels ? `渠道总量 ${channels.length} 个` : `累计发送 ${cumulative} 条`,
      trendText: activeChannels ? '查看渠道' : '暂无渠道发送',
      trendType: 'flat',
      tone: 'purple',
      iconName: 'activity',
      sparklineData: channelSparkline,
      action: buildChannelAction(context)
    }
  ]
}

export const deriveChannelInsight = (channels: ChannelStatPoint[], context: DashboardFilterContext): ChannelInsight => {
  const valid = channels
    .map(item => ({ name: item.way_name || '未知渠道', count: safeNumber(item.count_num) }))
    .filter(item => item.count > 0)
    .sort((a, b) => b.count - a.count)
  const total = valid.reduce((sum, item) => sum + item.count, 0)
  const rankedChannels: RankedChannelInsight[] = valid.map((item, index) => ({
    ...item,
    percent: total ? Math.round((item.count / total) * 100) : 0,
    color: channelColors[index % channelColors.length],
    action: buildChannelAction(context, item.name)
  }))
  return {
    total,
    primaryChannel: rankedChannels[0],
    rankedChannels,
    empty: total === 0,
    rangeLabel: context.rangeLabel
  }
}

export const deriveActionableEvents = (health: DashboardHealthSummary, channels: ChannelInsight, trendData: SendTrendPoint[], context: DashboardFilterContext): ActionableEvent[] => {
  const totals = calculateTrendTotals(trendData)
  if (health.status === 'empty') {
    return [{ id: 'empty-range', severity: 'info', title: '当前范围暂无数据', description: '可切换日期范围，或检查模板、定时任务和订阅是否已启用。', timeLabel: context.rangeLabel, source: 'empty-state' }]
  }
  const events: ActionableEvent[] = []
  if (totals.failed > 0) {
    events.push({ id: 'failed-send', severity: totals.successRate >= 90 ? 'warning' : 'critical', title: '存在失败发送需要处理', description: `失败 ${totals.failed} 条，建议进入任务日志查看失败原因。`, timeLabel: context.rangeLabel, source: 'failed-metric', action: buildTaskLogAction(context, { status: '0' }, '处理失败') })
  } else {
    events.push({ id: 'healthy-send', severity: 'success', title: '发送运行平稳', description: `当前范围成功发送 ${totals.success} 条，未发现失败堆积。`, timeLabel: context.rangeLabel, source: 'health-summary', action: buildTaskLogAction(context, {}, '查看明细') })
  }
  if (channels.primaryChannel) {
    events.push({ id: 'primary-channel', severity: 'info', title: `主渠道：${channels.primaryChannel.name}`, description: `该渠道发送 ${channels.primaryChannel.count} 条，占比 ${channels.primaryChannel.percent}%。`, timeLabel: context.rangeLabel, source: 'channel-insight', action: channels.primaryChannel.action })
  }
  return events
}

export const createLastUpdatedLabel = () => formatDateTime(new Date())
