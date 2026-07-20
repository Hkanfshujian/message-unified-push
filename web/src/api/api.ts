import type { AxiosRequestConfig, AxiosResponse } from 'axios'
import { handleException, logout, request } from './client'
import type { ApiResponse } from '@/types/common'
import type { SiteConfigState } from '@/types/app'

export { handleException, logout, request }

export interface LoginPayload {
  username: string
  passwd: string
}

export interface RegisterPayload {
  username: string
  passwd: string
  confirm_passwd: string
}

export interface AuthTokenPayload {
  token: string
}

export interface PublicAuthConfig {
  register_enabled?: string
  casdoor_enabled?: string
  casdoor_button_text?: string
  casdoor_button_icon?: string
}

export interface CurrentUserProfile {
  id?: number
  username?: string
  nickname?: string
  roles?: string[]
  groups?: string[]
  permissions?: string[]
  is_super_admin?: boolean
  [key: string]: unknown
}

export type MessageCategory = 'all' | 'system' | 'push'
export type MessageReadStatus = 'all' | 'unread' | 'read'

export interface MessageTargetScopeInput {
  target_type: 'all' | 'user' | 'role' | 'department' | 'position' | 'group'
  target_id: string
}

export interface MessageListItem {
  id: string
  category: 'system' | 'push'
  type: string
  title: string
  summary: string
  content?: string
  time: string
  is_read: boolean
  is_pinned?: boolean
  thumbnail_url?: string
  thumbnail_ratio?: '1:1' | '16:9'
  source_subject?: string
  target_url?: string
}

export interface MessageListQuery {
  category?: MessageCategory
  type?: string
  read_status?: MessageReadStatus
  start_time?: string
  end_time?: string
  page_num?: number
  page_size?: number
}

export interface MessageListResponse {
  list: MessageListItem[]
  lists?: MessageListItem[]
  total: number
  page_num: number
  page_size: number
  has_more: boolean
  sync_version: number
}

export interface MessageUnreadResponse {
  unread_count: number
  display_count: string
  sync_version: number
}

export interface MessageMutationResponse {
  unread_count: number
  sync_version: number
}

export interface MessageDeliveryEvent {
  event_id: string
  category: 'system' | 'push'
  message_id: string
  event_type: 'created' | 'read' | 'deleted' | 'expired' | 'scope_changed'
  sync_version: number
  occurred_at: string
}

export interface MessageSyncResponse {
  events: MessageDeliveryEvent[]
  unread_count: number
  latest_version: number
}

export interface SystemMessagePayload {
  id?: string
  type: string
  title: string
  summary?: string
  content?: string
  status?: 'published' | 'draft'
  publish_time?: string
  effective_start_time?: string
  effective_end_time?: string
  is_pinned?: boolean
  target_scopes: MessageTargetScopeInput[]
}

export const apiGet = <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
  return request.get<ApiResponse<T>>(url, config)
}

export const apiPost = <T = unknown, D = unknown>(url: string, data?: D, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
  return request.post<ApiResponse<T>>(url, data, config)
}

export const login = (payload: LoginPayload) => apiPost<AuthTokenPayload, LoginPayload>('/auth', payload)

export const register = (payload: RegisterPayload) => apiPost<AuthTokenPayload, RegisterPayload>('/auth/register', payload)

export const getPublicAuthConfig = () => apiGet<PublicAuthConfig>('/auth/public-config', {
  params: { t: Date.now() },
  meta: { silentBizToast: true, silentErrorToast: true }
})

export const getCurrentUser = () => apiGet<CurrentUserProfile>('/profile/current')

export const getSiteSetting = (section = 'site_config') => apiGet<SiteConfigState>('/settings/getsetting', {
  params: { section }
})

export const setSetting = <T extends Record<string, unknown>>(section: string, data: T) => apiPost('/settings/set', {
  section,
  data
})

export const resetSetting = (section: string) => apiPost('/settings/reset', { section })

export const getMessageUnreadCount = (category: MessageCategory = 'all') => apiGet<MessageUnreadResponse>('/message-center/unread-count', { params: { category } })

export const getMessageCenterMessages = (params: MessageListQuery) => apiGet<MessageListResponse>('/message-center/messages', { params })

export const markMessageRead = (category: Exclude<MessageCategory, 'all'>, messageId: string) => apiPost<MessageMutationResponse>('/message-center/messages/read', { category, message_id: messageId })

export const markAllMessagesRead = (payload: Pick<MessageListQuery, 'category' | 'type' | 'start_time' | 'end_time'> = {}) => apiPost<MessageMutationResponse>('/message-center/messages/read-all', payload)

export const deleteMessageForUser = (category: Exclude<MessageCategory, 'all'>, messageId: string) => apiPost<MessageMutationResponse>('/message-center/messages/delete', { category, message_id: messageId })

export const syncMessageCenter = (afterVersion = 0) => apiGet<MessageSyncResponse>('/message-center/sync', { params: { after_version: afterVersion }, meta: { silentBizToast: true, silentErrorToast: true } })

export const getSystemMessages = (params: Record<string, unknown>) => apiGet('/system-messages/list', { params })

export const addSystemMessage = (payload: SystemMessagePayload) => apiPost('/system-messages/add', payload)

export const editSystemMessage = (payload: SystemMessagePayload) => apiPost('/system-messages/edit', payload)

export const deleteSystemMessage = (id: string) => apiPost('/system-messages/delete', { id })

export const deleteSystemMessages = (ids: string[]) => apiPost('/system-messages/delete', { ids })
