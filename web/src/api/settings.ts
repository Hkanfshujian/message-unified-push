import { request } from './api'
import type { ApiResponse } from '@/types/common'

export type SettingsData = Record<string, any>
export type StoragePayload = Record<string, unknown>

export const settingsApi = {
  get: (section: string) => request.get<ApiResponse<SettingsData>>('/settings/getsetting', { params: { section } }),
  set: (section: string, data: SettingsData) => request.post<ApiResponse>('/settings/set', { section, data }),
  reset: () => request.post<ApiResponse>('/settings/reset', {}),
  getStorageConfig: () => request.get<ApiResponse<SettingsData>>('/system/storage-config'),
  saveStorageConfig: (data: StoragePayload) => request.post<ApiResponse>('/system/storage-config', data),
  uploadStorageFile: (data: FormData, signal?: AbortSignal) => request.post<ApiResponse<SettingsData>>('/system/storage-config/upload-file', data, signal ? { signal } : undefined),
  testLocalUpload: (data: FormData, signal?: AbortSignal) => request.post<ApiResponse<SettingsData>>('/system/storage-config/test-local-upload', data, signal ? { signal } : undefined),
  listS3Objects: (profileId: string, path: string) => request.get<ApiResponse<SettingsData>>('/system/storage-config/s3-objects', { params: { profile_id: profileId, path } }),
  listLocalFiles: (profileId: string, path: string) => request.get<ApiResponse<SettingsData>>('/system/storage-config/local-files', { params: { profile_id: profileId, path } }),
  deleteStorageFile: (data: StoragePayload) => request.post<ApiResponse>('/system/storage-config/delete-file', data),
  listLocalDirectories: (params: Record<string, unknown>) => request.get<ApiResponse<SettingsData>>('/system/storage-config/local-directories', { params }),
  createLocalDirectory: (data: StoragePayload) => request.post<ApiResponse>('/system/storage-config/local-directories', data),
  uploadSiteLogo: (data: FormData) => request.post<ApiResponse<SettingsData>>('/system/site-logo/upload', data),
  clearSiteLogo: (deleteSource: boolean) => request.post<ApiResponse>('/system/site-logo/clear', { delete_source: deleteSource })
}
