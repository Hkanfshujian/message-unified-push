import { request } from './api'
import type { ApiResponse } from '@/types/common'

export interface MQSourceApiRecord extends Record<string, unknown> {
  id: string
  name: string
  type: string
  namesrv_addr: string
  access_key: string
  secret_key: string
  enabled: number
  last_test_status: string
  last_test_time: string
  test_error: string
  created_on: string
  binding_count?: number
}

interface MQSourceListResult {
  list: MQSourceApiRecord[]
  total: number
}

interface MQConnectionTestResult {
  success: boolean
  message?: string
  error?: string
}

export type MQPayload = Record<string, unknown>
export type MQQuery = Record<string, unknown>

export const mqApi = {
  list: (params: MQQuery) => request.get<ApiResponse<MQSourceListResult>>('/mq-sources/list', { params }),
  create: (data: MQPayload) => request.post<ApiResponse>('/mq-sources/add', data),
  update: (id: number | string, data: MQPayload) => request.post<ApiResponse>(`/mq-sources/${id}/edit`, data),
  testConfig: (data: MQPayload) => request.post<ApiResponse<MQConnectionTestResult>>('/mq-sources/test-config', data),
  test: (id: number | string) => request.post<ApiResponse<MQConnectionTestResult>>(`/mq-sources/${id}/test`),
  remove: (id: number | string) => request.post<ApiResponse>(`/mq-sources/${id}/delete`),
  export: (params: MQQuery) => request.get<Blob>('/mq-sources/export', { params, responseType: 'blob' })
}
