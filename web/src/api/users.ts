import { request } from './api'
import type { ApiResponse } from '@/types/common'

export const usersApi = {
  getTheme: () => request.get<ApiResponse<Record<string, unknown>>>('/profile/theme'),
  updatePassword: (data: { old_passwd: string; new_passwd: string }) => request.post<ApiResponse>('/profile/password', data)
}
