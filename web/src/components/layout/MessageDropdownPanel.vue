<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import { useMessageCenterStore } from '@/store/message-center'
import { useRbacStore } from '@/store'
import type { MessageListItem } from '@/api/api'
import { canNavigateMessage, getMessageTargetPath, getMessageTypeLabel } from '@/util/message-center'
import { notifyError } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const store = useMessageCenterStore()
const rbacStore = useRbacStore()
const router = useRouter()
const copy = zhCN.layout.messageDropdown
const activeFilter = ref<'all' | 'unread' | 'read'>('unread')
const previewOpen = ref(false)
const previewMessage = ref<MessageListItem | null>(null)
const previewShouldMarkRead = ref(false)
const previewClosing = ref(false)

const messages = computed(() => store.latestSystemMessages)
const filterTabs = [
  { key: 'unread' as const, label: '未读' },
  { key: 'read' as const, label: '已读' },
  { key: 'all' as const, label: '全部' }
]

const selectFilter = (key: 'all' | 'unread' | 'read') => {
  activeFilter.value = key
}

const handleTabKeydown = (event: KeyboardEvent, index: number) => {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
  event.preventDefault()
  const nextIndex = event.key === 'Home' ? 0 : event.key === 'End' ? filterTabs.length - 1 : event.key === 'ArrowRight' ? (index + 1) % filterTabs.length : (index === 0 ? filterTabs.length - 1 : index - 1)
  selectFilter(filterTabs[nextIndex]!.key)
  document.getElementById(`message-filter-${filterTabs[nextIndex]!.key}`)?.focus()
}

const filteredMessages = computed(() => {
  if (activeFilter.value === 'unread') return messages.value.filter(item => !item.is_read)
  if (activeFilter.value === 'read') return messages.value.filter(item => item.is_read)
  return messages.value
})

const emptyStateText = computed(() => {
  if (activeFilter.value === 'unread') return '当前没有未读系统通知'
  if (activeFilter.value === 'read') return '当前没有已读系统通知'
  return '暂无系统通知'
})

const footerHint = computed(() => {
  if (messages.value.length >= 50) return '仅展示最近 50 条系统通知'
  return `当前展示 ${filteredMessages.value.length} 条通知`
})

const canOpenSystemMessagesPage = computed(() => rbacStore.hasPermission('message:system:view'))

const previewStatusLabel = computed(() => {
  if (!previewMessage.value) return ''
  return previewShouldMarkRead.value ? '未读消息' : '已读消息'
})

const previewActionLabel = computed(() => {
  const type = previewMessage.value?.type || ''
  if (type === 'audit') return '前往处理'
  if (type === 'warning') return '查看详情'
  return '前往查看'
})

const openMessage = (item: MessageListItem) => {
  previewMessage.value = item
  previewShouldMarkRead.value = !item.is_read
  previewOpen.value = true
}

const completePreviewRead = async () => {
  if (previewClosing.value) return
  previewClosing.value = true
  const item = previewMessage.value
  const shouldMarkRead = previewShouldMarkRead.value
  if (item && shouldMarkRead) {
    try {
      await store.markRead(item)
    } catch {
      notifyError('标记已读失败，消息仍保持未读')
    }
  }
  previewClosing.value = false
}

const resetPreviewState = () => {
  previewMessage.value = null
  previewShouldMarkRead.value = false
}

const closePreview = async () => {
  await completePreviewRead()
  previewOpen.value = false
  resetPreviewState()
}

const openMessageTarget = async () => {
  const item = previewMessage.value
  if (!item) return
  const path = getMessageTargetPath(item)
  if (!path) return
  await completePreviewRead()
  previewOpen.value = false
  resetPreviewState()
  await router.push(path)
}

const handlePreviewBeforeClose = async (done: () => void) => {
  await completePreviewRead()
  done()
  resetPreviewState()
}

const refreshPanel = async () => {
  await store.fetchDropdownMessages().catch(() => undefined)
}

const openSystemMessagesPage = async () => {
  store.dropdownOpen = false
  await router.push('/system/messages').catch(() => undefined)
}

</script>

<template>
  <div id="message-notification-popover" class="message-dropdown-panel" role="region" :aria-label="copy.systemNotifications">
    <div class="message-dropdown-hero">
      <div class="message-dropdown-toolbar">
        <div class="message-dropdown-segment" role="tablist" :aria-label="copy.notificationFilter">
          <button
            v-for="(tab, index) in filterTabs"
            :id="`message-filter-${tab.key}`"
            :key="tab.key"
            type="button"
            role="tab"
            class="message-dropdown-segment-item"
            :class="{ active: activeFilter === tab.key }"
            :aria-selected="activeFilter === tab.key"
            aria-controls="message-filter-panel"
            :tabindex="activeFilter === tab.key ? 0 : -1"
            @click="selectFilter(tab.key)"
            @keydown="handleTabKeydown($event, index)"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="store.dropdownError && messages.length > 0" class="message-dropdown-inline-error" role="status" aria-live="polite">
      {{ store.dropdownError }}{{ copy.cachedDataSuffix }}
    </div>

    <div class="sr-only" role="status" aria-live="polite">
      {{ store.dropdownLoading ? copy.loading : `${emptyStateText}，${filteredMessages.length}${copy.itemUnit}` }}
    </div>

    <div
      id="message-filter-panel"
      role="tabpanel"
      :aria-labelledby="`message-filter-${activeFilter}`"
      class="contents"
    >
    <div v-if="store.dropdownLoading && messages.length === 0" class="message-dropdown-state">
      <div class="message-dropdown-state-title">{{ copy.loading }}</div>
      <div class="message-dropdown-state-text">{{ copy.loadingDescription }}</div>
    </div>
    <div v-else-if="store.dropdownError && messages.length === 0" class="message-dropdown-state">
      <div class="message-dropdown-state-title">{{ copy.syncFailed }}</div>
      <div class="message-dropdown-state-text">{{ store.dropdownError }}</div>
      <button type="button" class="message-dropdown-state-action" @click="refreshPanel">{{ copy.reload }}</button>
    </div>
    <div v-else-if="filteredMessages.length === 0" class="message-dropdown-state">
      <div class="message-dropdown-state-title">{{ copy.noMatches }}</div>
      <div class="message-dropdown-state-text">{{ emptyStateText }}</div>
    </div>
    <ul v-else class="message-dropdown-list">
      <li
        v-for="item in filteredMessages"
        :key="`${item.category}-${item.id}`"
        class="message-dropdown-list-item"
      >
      <button
        type="button"
        class="message-dropdown-item"
        :class="{ unread: !item.is_read, navigable: canNavigateMessage(item) }"
        @click="openMessage(item)"
      >
        <span class="message-dropdown-leading">
          <span class="message-dropdown-dot" :class="{ unread: !item.is_read, read: item.is_read }" aria-hidden="true" />
          <span v-if="item.is_pinned" class="message-dropdown-item-badge">{{ copy.pinned }}</span>
        </span>
        <span class="message-dropdown-content">
          <span class="message-dropdown-item-head">
            <span class="message-dropdown-item-title">{{ item.title }}</span>
            <el-tag size="small" effect="light">{{ getMessageTypeLabel(item.type) }}</el-tag>
          </span>
          <span class="message-dropdown-meta">
            <span v-if="item.source_subject" class="message-dropdown-source">{{ item.source_subject }}</span>
            <span class="message-dropdown-time">{{ item.time }}</span>
          </span>
        </span>
        <span v-if="canNavigateMessage(item)" class="message-dropdown-item-action">
          <span class="message-dropdown-item-action-label">{{ copy.view }}</span>
          <DoraIcon name="chevron-right" :size="14" />
        </span>
      </button>
      </li>
    </ul>
    </div>
    <div v-if="filteredMessages.length > 0" class="message-dropdown-footer-note">
      <span class="message-dropdown-footer-text">{{ footerHint }}</span>
      <button
        v-if="canOpenSystemMessagesPage"
        type="button"
        class="message-dropdown-footer-link"
        @click="openSystemMessagesPage"
      >
        {{ copy.openList }}
        <DoraIcon name="chevron-right" :size="14" />
      </button>
    </div>
  </div>

  <el-dialog
    v-model="previewOpen"
    width="min(680px, calc(100vw - 24px))"
    class="app-nested-dialog message-preview-dialog"
    destroy-on-close
    append-to-body
    :before-close="handlePreviewBeforeClose"
  >
    <template #header>
      <div class="message-preview-header" v-if="previewMessage">
        <div class="message-preview-title-row">
          <span class="message-dropdown-dot" :class="{ unread: !previewMessage.is_read, read: previewMessage.is_read }" aria-hidden="true" />
          <span class="message-preview-title">{{ previewMessage.title }}</span>
          <el-tag size="small" effect="light">{{ getMessageTypeLabel(previewMessage.type) }}</el-tag>
          <span class="message-preview-status" :class="{ unread: previewShouldMarkRead, read: !previewShouldMarkRead }">{{ previewStatusLabel }}</span>
        </div>
      </div>
    </template>

    <div v-if="previewMessage" class="message-preview-body">
      <div class="message-preview-summary">{{ previewMessage.summary || copy.noSummary }}</div>
      <div class="message-preview-content">{{ previewMessage.content || previewMessage.summary || copy.noContent }}</div>
    </div>

    <template #footer>
      <div class="message-preview-footer">
        <el-button @click="closePreview">{{ copy.close }}</el-button>
        <el-button v-if="previewMessage && canNavigateMessage(previewMessage)" type="primary" @click="openMessageTarget">
          {{ previewActionLabel }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>
