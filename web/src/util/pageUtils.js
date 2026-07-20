// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig';

/**
 * 获取分页大小配置
 * @returns {number} 分页大小，默认为20
 */
export const getPageSize = () => {
  try {
    const config = LocalStieConfigUtils.getLocalConfig()
    return config?.pagesize ? Number(config.pagesize) : 20
  } catch (error) {
    return 20
  }
}

export const normalizePageQuery = (query = {}) => {
  const page = Number(query.page || query.pageNum || 1)
  const pageSize = Number(query.page_size || query.pageSize || getPageSize())
  return {
    ...query,
    page: Number.isFinite(page) && page > 0 ? page : 1,
    page_size: Number.isFinite(pageSize) && pageSize > 0 ? pageSize : getPageSize()
  }
}

export const serializeDateRange = (range = []) => {
  if (!Array.isArray(range) || range.length < 2) {
    return {}
  }
  return {
    start_time: range[0],
    end_time: range[1]
  }
}

export const buildExportFilename = (prefix, ext = 'csv') => {
  const date = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')
  return `${prefix}-${date}.${ext}`
}

export const removeEmptyParams = (params = {}) => {
  return Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== '' && value !== undefined && value !== null)
  )
}
