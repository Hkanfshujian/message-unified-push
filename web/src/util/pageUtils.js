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
