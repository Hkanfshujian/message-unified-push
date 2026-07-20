type MessageCategory = 'all' | 'system' | 'push'

interface MessageListItem {
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

export const getMessageDisplayCount = (count: number) => {
  if (!Number.isFinite(count) || count <= 0) return ''
  return count > 99 ? '99+' : String(Math.floor(count))
}

export const getMessageTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    announcement: '公告',
    warning: '预警',
    audit: '审核',
    cron_sent: '定时发送',
    subscription_triggered: '订阅触发',
    template_send: '接口调用'
  }
  return map[type] || type || '消息'
}

export const getMessageCategoryLabel = (category: MessageCategory) => {
  if (category === 'system') return '系统通知'
  if (category === 'push') return '最新推送'
  return '全部消息'
}

export const getMessageTone = (item: Pick<MessageListItem, 'type' | 'category'>) => {
  if (item.type === 'warning') return 'warning'
  if (item.type === 'audit') return 'primary'
  if (item.category === 'push') return 'success'
  return 'info'
}

export const normalizeMessageList = (items: MessageListItem[] = []) => {
  return [...items].sort((a, b) => {
    if (a.is_pinned !== b.is_pinned) return a.is_pinned ? -1 : 1
    return new Date(b.time || 0).getTime() - new Date(a.time || 0).getTime()
  })
}

export const canNavigateMessage = (item: MessageListItem) => Boolean(item.target_url && item.target_url.trim())

export const getMessageTargetPath = (item: MessageListItem) => item.target_url?.trim() || ''

export const getMessageThumbnailRatioClass = (item: MessageListItem) => item.thumbnail_ratio === '1:1' ? 'message-thumb-square' : 'message-thumb-wide'
