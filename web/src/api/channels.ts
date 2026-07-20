import { request } from './api'
import type { ApiResponse } from '@/types/common'

export type ChannelPayload = Record<string, unknown>
export type ChannelQuery = Record<string, unknown>

interface ChannelApiRecord extends Record<string, unknown> {
  id: number
  name: string
  type: string
  auth?: string
  config?: string
  created_on: string
  modified_on: string
  status: number
}

type ChannelListResult = { lists: ChannelApiRecord[]; total: number }

export const channelsApi = {
  list: (params: ChannelQuery) => request.get<ApiResponse<ChannelListResult>>('/sendways/list', { params }),
  create: (data: ChannelPayload) => request.post<ApiResponse>('/sendways/addone', data),
  update: (data: ChannelPayload) => request.post<ApiResponse>('/sendways/edit', data),
  remove: (id: number | string) => request.post<ApiResponse>('/sendways/delete', { id }),
  test: (data: ChannelPayload) => request.post<ApiResponse>('/sendways/test', data, { meta: { silentBizToast: true, silentErrorToast: true } }),
  export: (params: ChannelQuery) => request.get<Blob>('/sendways/export', { params, responseType: 'blob' })
}
