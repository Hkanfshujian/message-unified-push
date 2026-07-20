import { request } from './api'
import type { ApiResponse } from '@/types/common'

export type StatisticData = any

export const statisticsApi = {
  basic: () => request.get<ApiResponse<StatisticData>>('/statistic', { params: { type: 'basic' } }),
  trend: (days: number) => request.get<ApiResponse<StatisticData>>('/statistic', { params: { type: 'trend', days } }),
  channels: () => request.get<ApiResponse<StatisticData>>('/statistic', { params: { type: 'channels' } }),
  query: (params: Record<string, string | number | undefined>) => request.get<ApiResponse<StatisticData>>('/statistic', { params })
}
