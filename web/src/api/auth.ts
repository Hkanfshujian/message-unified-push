import { request } from './api'
import type { ApiResponse } from '@/types/common'
import type { AuthTokenPayload, PublicAuthConfig } from './api'
import type { RbacGroupRecord } from '@/types/business'

interface AuthConfig {
  register_enabled?: string
  casdoor_enabled?: string
  casdoor_endpoint?: string
  casdoor_client_id?: string
  casdoor_client_secret?: string
  casdoor_redirect_uri?: string
  casdoor_auth_path?: string
  casdoor_token_path?: string
  casdoor_userinfo_path?: string
  casdoor_logout_path?: string
  casdoor_auto_create_user?: string
  casdoor_default_group_code?: string
  casdoor_button_text?: string
  casdoor_button_icon?: string
  local_default_group_code?: string
}

export const authApi = {
  publicConfig: () => request.get<ApiResponse<PublicAuthConfig>>('/auth/public-config', { params: { t: Date.now() } }),
  login: (username: string, passwd: string) => request.post<ApiResponse<AuthTokenPayload>>('/auth', { username, passwd }),
  register: (data: Record<string, unknown>) => request.post<ApiResponse<AuthTokenPayload>>('/auth/register', data),
  groups: () => request.get<ApiResponse<{ lists: RbacGroupRecord[] }>>('/rbac/groups'),
  getConfig: () => request.get<ApiResponse<AuthConfig>>('/system/auth-config'),
  saveConfig: (data: Record<string, unknown>) => request.post<ApiResponse>('/system/auth-config', data)
}
