import { defineStore } from 'pinia'
import { addSystemMessage, deleteMessageForUser, deleteSystemMessage, deleteSystemMessages, editSystemMessage, getMessageCenterMessages, getMessageUnreadCount, getSystemMessages, markAllMessagesRead, markMessageRead, syncMessageCenter } from '@/api/api'
import type { MessageCategory, MessageDeliveryEvent, MessageListItem, MessageListQuery, SystemMessagePayload } from '@/api/api'
import { getMessageDisplayCount, normalizeMessageList } from '@/util/message-center'

interface MessageCenterState {
  unreadCount: number
  syncVersion: number
  dropdownOpen: boolean
  dropdownLoading: boolean
  dropdownError: string
  pageLoading: boolean
  syncTimer: number | null
  syncing: boolean
  initialized: boolean
  lastSyncAt: number | null
  latestSystemMessages: MessageListItem[]
  pageMessages: MessageListItem[]
  pageTotal: number
  events: MessageDeliveryEvent[]
}

const DROPDOWN_MESSAGE_LIMIT = 50

export const useMessageCenterStore = defineStore('messageCenter', {
  state: (): MessageCenterState => ({
    unreadCount: 0,
    syncVersion: 0,
    dropdownOpen: false,
    dropdownLoading: false,
    dropdownError: '',
    pageLoading: false,
    syncTimer: null,
    syncing: false,
    initialized: false,
    lastSyncAt: null,
    latestSystemMessages: [],
    pageMessages: [],
    pageTotal: 0,
    events: []
  }),
  getters: {
    displayUnreadCount: state => getMessageDisplayCount(state.unreadCount),
    unreadSystemMessages: state => state.latestSystemMessages.filter(item => !item.is_read),
    hasUnreadSystemMessages: state => state.latestSystemMessages.some(item => !item.is_read) || state.unreadCount > 0
  },
  actions: {
    touchSync(version = 0) {
      this.syncVersion = Math.max(this.syncVersion, Number(version || 0))
      this.lastSyncAt = Date.now()
    },
    resetState() {
      this.stopPolling()
      this.unreadCount = 0
      this.syncVersion = 0
      this.dropdownOpen = false
      this.dropdownLoading = false
      this.dropdownError = ''
      this.pageLoading = false
      this.syncing = false
      this.initialized = false
      this.lastSyncAt = null
      this.latestSystemMessages = []
      this.pageMessages = []
      this.pageTotal = 0
      this.events = []
    },
    async initialize() {
      if (this.initialized) {
        this.startPolling()
        return
      }
      await Promise.allSettled([
        this.refreshUnreadCount(),
        this.fetchDropdownMessages()
      ])
      this.initialized = true
      this.startPolling()
    },
    async refreshUnreadCount() {
      const rsp = await getMessageUnreadCount('system')
      const data = rsp.data.data
      this.unreadCount = Number(data?.unread_count || 0)
      this.touchSync(Number(data?.sync_version || 0))
    },
    async fetchDropdownMessages() {
      this.dropdownLoading = true
      this.dropdownError = ''
      try {
        const systemRsp = await getMessageCenterMessages({ category: 'system', page_num: 1, page_size: DROPDOWN_MESSAGE_LIMIT })
        this.latestSystemMessages = normalizeMessageList(systemRsp.data.data?.list || systemRsp.data.data?.lists || []).slice(0, DROPDOWN_MESSAGE_LIMIT)
        this.touchSync(Number(systemRsp.data.data?.sync_version || 0))
      } catch (error) {
        this.dropdownError = '系统通知同步失败，请稍后重试'
        throw error
      } finally {
        this.dropdownLoading = false
      }
    },
    async fetchPageMessages(query: MessageListQuery) {
      this.pageLoading = true
      try {
        const rsp = await getMessageCenterMessages(query)
        this.pageMessages = normalizeMessageList(rsp.data.data?.list || rsp.data.data?.lists || [])
        this.pageTotal = Number(rsp.data.data?.total || 0)
        this.touchSync(Number(rsp.data.data?.sync_version || 0))
      } finally {
        this.pageLoading = false
      }
    },
    async markRead(item: MessageListItem) {
      if (!item || item.is_read) return
      const rsp = await markMessageRead(item.category, item.id)
      this.applyMessageRead(item.category, item.id)
      if (item.category === 'system') {
        this.unreadCount = Number(rsp.data.data?.unread_count ?? Math.max(this.unreadCount - 1, 0))
      }
      this.touchSync(Number(rsp.data.data?.sync_version || 0))
    },
    async markAllRead(category: MessageCategory = 'system') {
      const rsp = await markAllMessagesRead({ category })
      this.latestSystemMessages = this.latestSystemMessages.map(item => ({ ...item, is_read: category === 'all' || category === 'system' ? true : item.is_read }))
      this.pageMessages = this.pageMessages.map(item => ({ ...item, is_read: category === 'all' || category === item.category ? true : item.is_read }))
      this.unreadCount = Number(rsp.data.data?.unread_count || 0)
      this.touchSync(Number(rsp.data.data?.sync_version || 0))
    },
    async deleteMessage(item: MessageListItem) {
      const rsp = await deleteMessageForUser(item.category, item.id)
      this.latestSystemMessages = this.latestSystemMessages.filter(message => !(message.category === item.category && message.id === item.id))
      this.pageMessages = this.pageMessages.filter(message => !(message.category === item.category && message.id === item.id))
      this.pageTotal = Math.max(this.pageTotal - 1, 0)
      this.unreadCount = Number(rsp.data.data?.unread_count ?? this.unreadCount)
      this.touchSync(Number(rsp.data.data?.sync_version || 0))
    },
    applyMessageRead(category: Exclude<MessageCategory, 'all'>, id: string) {
      const updater = (item: MessageListItem) => item.category === category && item.id === id ? { ...item, is_read: true } : item
      this.latestSystemMessages = this.latestSystemMessages.map(updater)
      this.pageMessages = this.pageMessages.map(updater)
    },
    async pollSync() {
      if (this.syncing) return
      this.syncing = true
      try {
        const rsp = await syncMessageCenter(this.syncVersion)
        const data = rsp.data.data
        const events = data?.events || []
        this.events = [...this.events, ...events].slice(-100)
        this.touchSync(Number(data?.latest_version || 0))
        if (events.some(event => event.category === 'system')) {
          await this.fetchDropdownMessages().catch(() => undefined)
        }
        await this.refreshUnreadCount()
      } finally {
        this.syncing = false
      }
    },
    startPolling() {
      if (typeof window === 'undefined') return
      this.stopPolling()
      this.syncTimer = window.setInterval(() => {
        this.pollSync().catch(() => undefined)
      }, 5000)
    },
    stopPolling() {
      if (this.syncTimer) {
        window.clearInterval(this.syncTimer)
        this.syncTimer = null
      }
    },
    async getAdminMessages(params: Record<string, unknown>) {
      return getSystemMessages(params)
    },
    async saveSystemMessage(payload: SystemMessagePayload) {
      if (payload.id) return editSystemMessage(payload)
      return addSystemMessage(payload)
    },
    async removeSystemMessage(id: string) {
      return deleteSystemMessage(id)
    },
    async removeSystemMessages(ids: string[]) {
      return deleteSystemMessages(ids)
    }
  }
})
