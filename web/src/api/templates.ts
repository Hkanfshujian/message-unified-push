import { request } from './api'
import type { ApiResponse } from '@/types/common'

export type TemplatePayload = Record<string, unknown>
export type TemplateQuery = Record<string, unknown>

export interface TemplateApiRecord {
  id: string
  name: string
  description: string
  text_template: string
  html_template: string
  markdown_template: string
  placeholders: string
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status: string
  created_on: string
  modified_on: string
  cron_msg_count?: number
}

interface TemplateListResult { lists: TemplateApiRecord[]; total: number }
interface TemplatePreviewResult { text?: string; html?: string; markdown?: string }
interface TemplateRelationRecord { type?: string; id?: string; name?: string }
interface TemplateRelationsResult { relations: TemplateRelationRecord[] }

export const templatesApi = {
  list: (params: TemplateQuery) => request.get<ApiResponse<TemplateListResult>>('/templates/list', { params }),
  create: (data: TemplatePayload) => request.post<ApiResponse>('/templates/add', data),
  update: (data: TemplatePayload) => request.post<ApiResponse>('/templates/edit', data),
  preview: (data: TemplatePayload) => request.post<ApiResponse<TemplatePreviewResult>>('/templates/preview', data),
  remove: (id: number | string) => request.post<ApiResponse>('/templates/delete', { id }),
  relations: (id: number | string) => request.get<ApiResponse<TemplateRelationsResult>>('/templates/relations', { params: { id } })
}
