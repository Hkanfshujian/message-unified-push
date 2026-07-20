import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMessageCenterStore } from './message-center'
import type { MessageDeliveryEvent, MessageListItem } from '@/api/api'

const api = vi.hoisted(() => ({
  addSystemMessage: vi.fn(),
  deleteMessageForUser: vi.fn(),
  deleteSystemMessage: vi.fn(),
  deleteSystemMessages: vi.fn(),
  editSystemMessage: vi.fn(),
  getMessageCenterMessages: vi.fn(),
  getMessageUnreadCount: vi.fn(),
  getSystemMessages: vi.fn(),
  markAllMessagesRead: vi.fn(),
  markMessageRead: vi.fn(),
  syncMessageCenter: vi.fn()
}))

vi.mock('@/api/api', () => api)

const createApiResponse = <T>(data: T) => ({
  data: {
    code: 0,
    msg: 'ok',
    data
  }
}) as any

const createMessage = (overrides: Partial<MessageListItem> = {}): MessageListItem => ({
  id: 'msg-1',
  category: 'system',
  type: 'announcement',
  title: '系统维护通知',
  summary: '今晚 23:00 进行系统维护',
  content: '维护期间部分服务会短暂抖动',
  time: '2026-07-02 18:30:00',
  is_read: false,
  ...overrides
})

const createEvent = (overrides: Partial<MessageDeliveryEvent> = {}): MessageDeliveryEvent => ({
  event_id: 'evt-1',
  category: 'system',
  message_id: 'msg-1',
  event_type: 'created',
  sync_version: 6,
  occurred_at: '2026-07-02 18:30:00',
  ...overrides
})

describe('message center store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.getMessageUnreadCount.mockResolvedValue(createApiResponse({ unread_count: 2, display_count: '2', sync_version: 3 }))
    api.getMessageCenterMessages.mockResolvedValue(createApiResponse({
      list: [createMessage()],
      total: 1,
      page_num: 1,
      page_size: 20,
      has_more: false,
      sync_version: 4
    }))
    api.markAllMessagesRead.mockResolvedValue(createApiResponse({ unread_count: 0, sync_version: 8 }))
    api.markMessageRead.mockResolvedValue(createApiResponse({ unread_count: 1, sync_version: 5 }))
    api.deleteMessageForUser.mockResolvedValue(createApiResponse({ unread_count: 0, sync_version: 9 }))
    api.syncMessageCenter.mockResolvedValue(createApiResponse({ events: [], unread_count: 2, latest_version: 4 }))
  })

  it('fetches only system messages for the dropdown and records sync metadata', async () => {
    const store = useMessageCenterStore()

    await store.fetchDropdownMessages()

    expect(api.getMessageCenterMessages).toHaveBeenCalledWith({ category: 'system', page_num: 1, page_size: 50 })
    expect(store.latestSystemMessages).toHaveLength(1)
    expect(store.dropdownError).toBe('')
    expect(store.lastSyncAt).not.toBeNull()
  })

  it('caps dropdown cache size to the most recent 50 messages when the backend returns 100+ items', async () => {
    const store = useMessageCenterStore()
    api.getMessageCenterMessages.mockResolvedValueOnce(createApiResponse({
      list: Array.from({ length: 120 }, (_, index) => createMessage({
        id: `msg-${index + 1}`,
        title: `系统通知 ${index + 1}`,
        time: new Date(Date.UTC(2026, 6, 2, 18, 0, 0) - index * 60_000).toISOString().slice(0, 19).replace('T', ' ')
      })),
      total: 120,
      page_num: 1,
      page_size: 120,
      has_more: true,
      sync_version: 11
    }))

    await store.fetchDropdownMessages()

    expect(store.latestSystemMessages).toHaveLength(50)
    expect(store.latestSystemMessages[0]?.id).toBe('msg-1')
    expect(store.latestSystemMessages[store.latestSystemMessages.length - 1]?.id).toBe('msg-50')
  })

  it('marks all cached system messages as read after bulk acknowledge', async () => {
    const store = useMessageCenterStore()
    store.latestSystemMessages = [
      createMessage({ id: 'msg-1', is_read: false }),
      createMessage({ id: 'msg-2', is_read: true, time: '2026-07-01 10:00:00' })
    ]
    store.pageMessages = [
      createMessage({ id: 'msg-1', is_read: false }),
      createMessage({ id: 'msg-3', category: 'push', is_read: false })
    ]

    await store.markAllRead('system')

    expect(api.markAllMessagesRead).toHaveBeenCalledWith({ category: 'system' })
    expect(store.latestSystemMessages.every(item => item.is_read)).toBe(true)
    expect(store.pageMessages.find(item => item.id === 'msg-3')?.is_read).toBe(false)
    expect(store.unreadCount).toBe(0)
  })

  it('refreshes dropdown data when sync returns new system events', async () => {
    const store = useMessageCenterStore()
    api.syncMessageCenter.mockResolvedValueOnce(createApiResponse({
      events: [createEvent()],
      unread_count: 3,
      latest_version: 6
    }))
    api.getMessageCenterMessages.mockResolvedValueOnce(createApiResponse({
      list: [createMessage({ id: 'msg-9', title: '新增告警', is_pinned: true })],
      total: 1,
      page_num: 1,
      page_size: 20,
      has_more: false,
      sync_version: 7
    }))
    api.getMessageUnreadCount.mockResolvedValueOnce(createApiResponse({ unread_count: 3, display_count: '3', sync_version: 7 }))

    await store.pollSync()

    expect(api.syncMessageCenter).toHaveBeenCalledWith(0)
    expect(api.getMessageCenterMessages).toHaveBeenCalledTimes(1)
    expect(api.getMessageUnreadCount).toHaveBeenCalledWith('system')
    expect(store.latestSystemMessages[0]?.id).toBe('msg-9')
    expect(store.unreadCount).toBe(3)
    expect(store.events).toHaveLength(1)
  })
})
