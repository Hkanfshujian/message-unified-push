import { request } from './api'
import type { ApiResponse } from '@/types/common'

export type QueryParams = Record<string, unknown>

export interface SendLogRecord {
  id: number
  task_id: string
  type: string
  name: string
  log: string
  created_on: string
  caller_ip?: string
  status: number
}

export interface ConsumeLogRecord {
  id: string
  subscription_id: string
  subscription_name: string
  raw_message: string
  matched: number
  extracted_values: string
  send_status: number
  send_error: string
  created_on: string
}

export interface LoginLogRecord {
  id: number
  user_id: number
  username: string
  ip: string
  ua: string
  created_on: string
}

interface ListsResult<T> {
  lists: T[]
  list: T[]
  total: number
}

interface ConsumeStatsResult {
  total_consume: number
  total_matched: number
  total_sent: number
  total_failed: number
}

export const sendLogsApi = {
  list: (params: QueryParams) => request.get<ApiResponse<ListsResult<SendLogRecord>>>('/sendlogs/list', { params }),
  export: (params: QueryParams) => request.get<Blob>('/sendlogs/export', { params, responseType: 'blob' })
}

export const consumeLogsApi = {
  detail: (id: number | string) => request.get<ApiResponse<ConsumeLogRecord>>(`/consume-logs/${id}`, { meta: { silentErrorToast: true } }),
  stats: () => request.get<ApiResponse<ConsumeStatsResult>>('/consume-logs/stats'),
  list: (params: QueryParams) => request.get<ApiResponse<ListsResult<ConsumeLogRecord>>>('/consume-logs/list', { params }),
  export: (params: QueryParams) => request.get<Blob>('/consume-logs/export', { params, responseType: 'blob' })
}

export const loginLogsApi = {
  recent: (params: QueryParams) => request.get<ApiResponse<ListsResult<LoginLogRecord>>>('/loginlogs/recent', { params }),
  export: (params: QueryParams) => request.get<Blob>('/loginlogs/export', { params, responseType: 'blob' })
}
