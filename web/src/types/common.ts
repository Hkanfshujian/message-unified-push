import type { RouteRecordRaw } from 'vue-router'

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PageQuery {
  page?: number
  page_size?: number
  pageSize?: number
  keyword?: string
  name?: string
  status?: string | number
  type?: string
  start_time?: string
  end_time?: string
  [key: string]: unknown
}

export interface PageResult<T> {
  list: T[]
  total: number
  page?: number
  page_size?: number
}

export interface SelectOption<T = string | number> {
  label: string
  value: T
  disabled?: boolean
}

export interface DateRangeValue {
  start?: string
  end?: string
}

export interface RouteMetaAuth {
  title?: string
  icon?: string
  permission?: string
  permissions?: string[]
  requiresAuth?: boolean
  hidden?: boolean
  affix?: boolean
}

export type AppRouteRecord = RouteRecordRaw & {
  meta?: RouteMetaAuth
  children?: AppRouteRecord[]
}

export interface RequestMeta {
  silentBizToast?: boolean
  silentErrorToast?: boolean
}

export interface ValidationResult {
  valid: boolean
  message?: string
}
