import type { DoraIconName } from '@/types/app'

export type EnabledStatus = 'enabled' | 'disabled' | number | boolean
export type ChannelType = 'Email' | 'Dtalk' | 'QyWeiXin' | 'Feishu' | 'Custom' | 'WeChatOFAccount' | 'AliyunSMS' | 'Telegram' | 'Bark' | 'PushMe' | 'Ntfy' | 'Gotify' | 'QyWeiXinApp'
export type TemplateContentType = 'text' | 'html' | 'markdown'

export interface ChannelAuthConfig {
  name?: string
  [key: string]: unknown
}

export interface SendWayRecord {
  id?: number | string
  name: string
  type: ChannelType | string
  auth?: ChannelAuthConfig | string
  status?: EnabledStatus
  created_at?: string
  updated_at?: string
  [key: string]: unknown
}

export interface TemplateRecord {
  id: string
  name: string
  description?: string
  text_template?: string
  html_template?: string
  markdown_template?: string
  placeholders?: string | Record<string, unknown>
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status?: EnabledStatus
  [key: string]: unknown
}

export interface TemplateInstanceRecord {
  id: string
  template_id: string
  channel_id: number | string
  channel_type: string
  content_type: TemplateContentType
  config?: Record<string, unknown> | string
  status?: EnabledStatus
  [key: string]: unknown
}

export interface CronMessageRecord {
  id: string
  name: string
  template_id: string
  cron: string
  enable?: boolean
  next_time?: string
  template_name?: string
  channel_names?: string[] | string
  [key: string]: unknown
}

export interface MQSourceRecord {
  id: string | number
  name: string
  type: string
  namesrv_addr?: string
  access_key?: string
  secret_key?: string
  status?: EnabledStatus
  [key: string]: unknown
}

export interface SubscriptionRecord {
  id: string
  name: string
  mq_source_id: string | number
  topic: string
  tag?: string
  consumer_group: string
  validate_rule?: string
  extract_rule?: string
  template_id?: string
  status?: string
  total_count?: number
  success_count?: number
  fail_count?: number
  [key: string]: unknown
}

export interface LogRecord {
  id: string | number
  task_id?: string
  name?: string
  type?: string
  status?: string
  ip?: string
  created_at?: string
  [key: string]: unknown
}

export type DashboardHealthStatus = 'healthy' | 'warning' | 'critical' | 'empty'
export type DashboardTone = 'blue' | 'green' | 'red' | 'amber' | 'purple' | 'slate'
export type DashboardEventSeverity = 'info' | 'success' | 'warning' | 'critical'

export interface DashboardFilterContext {
  startDate: string
  endDate: string
  rangeLabel: string
  trendDays: number
  routeQuery: Record<string, string>
}

export interface DashboardActionTarget {
  label: string
  path: string
  permission?: string
  disabled?: boolean
}

export interface DashboardHealthSummary {
  status: DashboardHealthStatus
  title: string
  description: string
  lastUpdatedAt: string
  rangeLabel: string
  primaryAction?: DashboardActionTarget
}

export interface CoreMetric {
  id: string
  label: string
  value: string | number
  unit?: string
  scopeLabel: string
  description: string
  trendText: string
  trendType: 'up' | 'down' | 'flat'
  tone: DashboardTone
  iconName: DoraIconName
  sparklineData: number[]
  action?: DashboardActionTarget
}

export interface RankedChannelInsight {
  name: string
  count: number
  percent: number
  color: string
  action?: DashboardActionTarget
}

export interface ChannelInsight {
  total: number
  primaryChannel?: RankedChannelInsight
  rankedChannels: RankedChannelInsight[]
  empty: boolean
  rangeLabel: string
}

export interface ActionableEvent {
  id: string
  severity: DashboardEventSeverity
  title: string
  description: string
  timeLabel: string
  source: string
  action?: DashboardActionTarget
}

export interface RbacRoleRecord {
  id: number
  name: string
  code: string
  status?: EnabledStatus
  [key: string]: unknown
}

export interface RbacUserRecord {
  id: number
  username: string
  nickname?: string
  status?: EnabledStatus
  [key: string]: unknown
}

export interface RbacGroupRecord {
  id: number
  name: string
  code: string
  status?: EnabledStatus
  [key: string]: unknown
}

export interface RbacPermissionRecord {
  id: number
  name: string
  code: string
  method?: string
  path?: string
  type?: string
  status?: EnabledStatus
  [key: string]: unknown
}
