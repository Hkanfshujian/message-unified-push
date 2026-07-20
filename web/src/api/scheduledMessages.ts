import { request } from './api'
import type { ApiResponse } from '@/types/common'

export interface ScheduledMessageRecord extends Record<string, unknown> {
  id: string
  name: string
  cron: string
  cron_expression: string
  template_id: string
  template_name: string
  ins_ids: string[]
  channel_names: string[]
  enable: number
  status: boolean
  created_on: string
  modified_on: string
  next_time?: string
}

interface ScheduledMessageListResult {
  lists: ScheduledMessageRecord[]
  total: number
}

export type ScheduledMessagePayload = Record<string, unknown>
export type ScheduledMessageQuery = Record<string, unknown>

export const scheduledMessagesApi = {
  list: (params: ScheduledMessageQuery) => request.get<ApiResponse<ScheduledMessageListResult>>('/cronmessages/list', { params }),
  create: (data: ScheduledMessagePayload) => request.post<ApiResponse>('/cronmessages/addone', data),
  update: (data: ScheduledMessagePayload) => request.post<ApiResponse>('/cronmessages/edit', data),
  setEnabled: (id: number | string, enabled: boolean) => request.post<ApiResponse>(enabled ? '/cronmessages/start' : '/cronmessages/stop', { id }),
  remove: (id: number | string) => request.post<ApiResponse>('/cronmessages/delete', { id }),
  sendNow: (data: ScheduledMessagePayload) => request.post<ApiResponse>('/cronmessages/sendnow', data),
  export: (params: ScheduledMessageQuery) => request.get<Blob>('/cronmessages/export', { params, responseType: 'blob' })
}
