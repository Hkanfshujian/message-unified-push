import { request } from './api'
import type { ApiResponse } from '@/types/common'

export type SubscriptionPayload = Record<string, unknown>
export type SubscriptionQuery = Record<string, unknown>

export interface SubscriptionApiRecord {
  id: string
  source_id: string
  source_name: string
  name: string
  topic: string
  tag: string
  group_name: string
  validate_regex: string
  extract_regex: string
  extract_field: string
  extract_rules?: Array<{ field: string; regex: string }>
  template_id: string
  template_name: string
  template_content_type?: string
  consume_mode?: string
  status: string
  total_consumed: number
  total_sent: number
  total_failed: number
  last_consume_time: string
  created_on: string
}

interface SubscriptionListResult { list: SubscriptionApiRecord[]; total: number }

interface RegexTestResult {
  validate_matched: boolean
  extracted_values?: Record<string, string>
}

export const subscriptionsApi = {
  list: (params: SubscriptionQuery) => request.get<ApiResponse<SubscriptionListResult>>('/subscriptions/list', { params }),
  detail: (id: number | string) => request.get<ApiResponse<SubscriptionApiRecord>>(`/subscriptions/${id}`),
  create: (data: SubscriptionPayload) => request.post<ApiResponse>('/subscriptions/add', data, { meta: { silentBizToast: true, silentErrorToast: true } }),
  update: (id: number | string, data: SubscriptionPayload) => request.post<ApiResponse>(`/subscriptions/${id}/edit`, data, { meta: { silentBizToast: true, silentErrorToast: true } }),
  setState: (id: number | string, action: 'start' | 'stop') => request.post<ApiResponse>(`/subscriptions/${id}/${action}`),
  remove: (id: number | string) => request.post<ApiResponse>(`/subscriptions/${id}/delete`),
  testRegex: (data: SubscriptionPayload) => request.post<ApiResponse<RegexTestResult>>('/subscriptions/regex-test', data),
  export: (params: SubscriptionQuery) => request.get<Blob>('/subscriptions/export', { params, responseType: 'blob' })
}
