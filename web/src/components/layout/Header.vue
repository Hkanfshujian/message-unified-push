<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import { Button } from '@/components/ui/button'
import MessageDropdownPanel from '@/components/layout/MessageDropdownPanel.vue'
import { useMessageCenterStore } from '@/store/message-center'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.layout.header

const props = defineProps<{
  userAccount: string
  theme: 'light' | 'dark'
  breadcrumb?: string
  isSidebarCollapsed?: boolean
  isMobile?: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-theme'): void
  (e: 'toggle-sidebar'): void
  (e: 'open-profile-settings'): void
  (e: 'logout'): void
}>()

const isUserMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)
const isFullscreen = ref(false)
const themeToggleAnimating = ref(false)
const messageMenuRef = ref<HTMLElement | null>(null)
const messageDropdownRef = ref<HTMLElement | null>(null)
const messageStore = useMessageCenterStore()
const messageDropdownStyle = ref<Record<string, string>>({})
let themeToggleTimer: number | undefined

const closeUserMenu = async (restoreFocus = false) => {
  if (!isUserMenuOpen.value) return
  isUserMenuOpen.value = false
  if (restoreFocus) {
    await nextTick()
    document.getElementById('user-menu-trigger')?.focus()
  }
}

const focusFirstUserMenuItem = async () => {
  await nextTick()
  userMenuRef.value?.querySelector<HTMLElement>('[role="menuitem"]')?.focus()
}

const toggleUserMenu = async () => {
  isUserMenuOpen.value = !isUserMenuOpen.value
  if (isUserMenuOpen.value) await focusFirstUserMenuItem()
}

const handleUserMenuKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isUserMenuOpen.value) {
    event.stopPropagation()
    void closeUserMenu(true)
    return
  }
  if (!isUserMenuOpen.value || !['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
  const items = Array.from(userMenuRef.value?.querySelectorAll<HTMLElement>('[role="menuitem"]') || [])
  if (!items.length) return
  event.preventDefault()
  const currentIndex = items.indexOf(document.activeElement as HTMLElement)
  const nextIndex = event.key === 'Home' ? 0 : event.key === 'End' ? items.length - 1 : event.key === 'ArrowDown' ? (currentIndex + 1) % items.length : (currentIndex <= 0 ? items.length - 1 : currentIndex - 1)
  items[nextIndex]?.focus()
}

const openProfileSettings = () => {
  isUserMenuOpen.value = false
  emit('open-profile-settings')
}

const requestLogout = () => {
  isUserMenuOpen.value = false
  emit('logout')
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null
  if (userMenuRef.value && !userMenuRef.value.contains(event.target as Node)) {
    isUserMenuOpen.value = false
  }
  const clickedInsideTrigger = messageMenuRef.value?.contains(event.target as Node)
  const clickedInsideDropdown = messageDropdownRef.value?.contains(event.target as Node)
  const clickedInsidePreviewDialog = Boolean(target?.closest('.message-preview-dialog, .el-overlay, .el-dialog'))
  if (!clickedInsideTrigger && !clickedInsideDropdown && !clickedInsidePreviewDialog) {
    void closeMessageMenu(false)
  }
}

const syncMessageDropdownPosition = () => {
  if (!messageStore.dropdownOpen || !messageMenuRef.value || typeof window === 'undefined') return
  const rect = messageMenuRef.value.getBoundingClientRect()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const horizontalPadding = 12
  const top = Math.round(rect.bottom + 8)

  if (viewportWidth <= 640) {
    messageDropdownStyle.value = {
      position: 'fixed',
      top: `${top}px`,
      left: `${horizontalPadding}px`,
      right: `${horizontalPadding}px`,
      width: 'auto',
      maxHeight: `${Math.max(280, viewportHeight - top - 12)}px`
    }
    return
  }

  const panelWidth = Math.min(420, viewportWidth - horizontalPadding * 2)
  const preferredLeft = rect.right - panelWidth
  const left = Math.min(
    Math.max(horizontalPadding, Math.round(preferredLeft)),
    Math.max(horizontalPadding, viewportWidth - panelWidth - horizontalPadding)
  )

  messageDropdownStyle.value = {
    position: 'fixed',
    top: `${top}px`,
    left: `${left}px`,
    width: `${panelWidth}px`,
    maxHeight: `${Math.max(320, viewportHeight - top - 16)}px`
  }
}

const closeMessageMenu = async (restoreFocus = true) => {
  if (!messageStore.dropdownOpen) return
  messageStore.dropdownOpen = false
  if (restoreFocus) {
    await nextTick()
    document.getElementById('message-notification-trigger')?.focus()
  }
}

const toggleMessageMenu = async () => {
  if (messageStore.dropdownOpen) {
    await closeMessageMenu()
    return
  }
  messageStore.dropdownOpen = true
  await nextTick()
  syncMessageDropdownPosition()
  messageDropdownRef.value?.querySelector<HTMLElement>('button:not(:disabled), [tabindex]:not([tabindex="-1"])')?.focus()
  await messageStore.fetchDropdownMessages().catch(() => undefined)
}

const handleMessageMenuEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && messageStore.dropdownOpen) {
    event.stopPropagation()
    void closeMessageMenu()
  }
}

const toggleFullscreen = async () => {
  try {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen?.()
    } else {
      await document.exitFullscreen?.()
    }
  } catch {
    syncFullscreenState()
  }
}

const syncFullscreenState = () => {
  isFullscreen.value = Boolean(document.fullscreenElement)
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('fullscreenchange', syncFullscreenState)
  window.addEventListener('resize', syncMessageDropdownPosition)
  window.addEventListener('scroll', syncMessageDropdownPosition, true)
  syncFullscreenState()
  messageStore.initialize().catch(() => undefined)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('fullscreenchange', syncFullscreenState)
  window.removeEventListener('resize', syncMessageDropdownPosition)
  window.removeEventListener('scroll', syncMessageDropdownPosition, true)
  if (themeToggleTimer) window.clearTimeout(themeToggleTimer)
  messageStore.stopPolling()
})

const breadcrumbItems = computed(() => {
  const text = props.breadcrumb?.trim() || '工作台'
  return text
    .split('/')
    .map((item) => item.trim())
    .filter(Boolean)
})

const sidebarButtonTitle = computed(() => props.isMobile ? '打开导航菜单' : (props.isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'))
const messageButtonTitle = computed(() => {
  if (!messageStore.hasUnreadSystemMessages) return '系统通知'
  return messageStore.displayUnreadCount ? `系统通知（${messageStore.displayUnreadCount} 条未读）` : '系统通知（有未读消息）'
})
const fullscreenButtonTitle = computed(() => isFullscreen.value ? '退出全屏' : '全屏')
const themeButtonTitle = computed(() => props.theme === 'dark' ? '切换到白天模式' : '切换到黑夜模式')

const requestThemeToggle = async () => {
  if (themeToggleAnimating.value) return
  themeToggleAnimating.value = true
  emit('toggle-theme')
  await nextTick()
  themeToggleTimer = window.setTimeout(() => {
    themeToggleAnimating.value = false
  }, 260)
}
</script>

<template>
  <header class="admin-header sticky top-0 z-40 w-full">
    <div class="admin-header-inner flex h-14 items-center px-3 gap-2">
      <div class="h-full flex items-center gap-2 flex-1 min-w-0">
        <Button
          variant="icon"
          size="icon"
          id="sidebar-toggle-button"
          class="admin-header-action dora-header-icon h-9 w-9 shrink-0"
          :title="sidebarButtonTitle"
          :aria-label="sidebarButtonTitle"
          @click="emit('toggle-sidebar')"
        >
          <DoraIcon :name="props.isMobile || props.isSidebarCollapsed ? 'menu-fold-right' : 'menu-fold-left'" :size="20" />
          <span class="sr-only">{{ messages.toggleSidebar }}</span>
        </Button>
        <div class="admin-header-breadcrumb flex items-center gap-2 text-sm text-[var(--admin-text-muted)] truncate">
          <template v-for="(item, idx) in breadcrumbItems" :key="`${item}-${idx}`">
            <span v-if="idx > 0" class="text-[var(--admin-text-muted)] shrink-0">/</span>
            <span
              :class="idx === breadcrumbItems.length - 1 ? 'text-foreground font-medium' : 'text-muted-foreground'"
              class="truncate"
            >
              {{ item }}
            </span>
          </template>
        </div>
      </div>

      <div class="admin-header-right h-full flex items-center justify-end gap-1">
        <slot name="actions" />

        <div class="relative" ref="messageMenuRef">
          <Button
            id="message-notification-trigger"
            variant="icon"
            size="icon"
            class="admin-header-action dora-header-icon h-9 w-9 message-header-trigger"
            :title="messageButtonTitle"
            :aria-label="messageButtonTitle"
            aria-haspopup="true"
            aria-controls="message-notification-popover"
            :aria-expanded="messageStore.dropdownOpen"
            @click.stop="toggleMessageMenu"
          >
            <DoraIcon name="notification" :size="18" />
            <span v-if="messageStore.displayUnreadCount" class="message-header-badge" aria-hidden="true">{{ messageStore.displayUnreadCount }}</span>
            <span class="sr-only">{{ messages.systemMessages }}</span>
          </Button>
          <Teleport to="body">
            <div
              v-if="messageStore.dropdownOpen"
              ref="messageDropdownRef"
              class="message-dropdown-layer"
              :style="messageDropdownStyle"
              @keydown="handleMessageMenuEscape"
            >
              <MessageDropdownPanel />
            </div>
          </Teleport>
        </div>

        <Button
          variant="icon"
          size="icon"
          class="admin-header-action dora-header-icon h-9 w-9"
          :title="fullscreenButtonTitle"
          :aria-label="fullscreenButtonTitle"
          :aria-pressed="isFullscreen"
          @click="toggleFullscreen"
        >
          <DoraIcon :name="isFullscreen ? 'fullscreen-exit' : 'fullscreen'" :size="17" />
          <span class="sr-only">{{ messages.toggleFullscreen }}</span>
        </Button>

        <Button
          variant="icon"
          size="icon"
          class="admin-header-action dora-header-icon h-9 w-9"
          :class="{ 'admin-header-action-switching': themeToggleAnimating }"
          :title="themeButtonTitle"
          :aria-label="themeButtonTitle"
          :aria-pressed="props.theme === 'dark'"
          @click="requestThemeToggle"
        >
          <DoraIcon :name="props.theme === 'dark' ? 'sun' : 'moon'" :size="18" />
          <span class="sr-only">{{ messages.toggleTheme }}</span>
        </Button>

        <div class="relative" ref="userMenuRef" @keydown="handleUserMenuKeydown">
          <el-button
            id="user-menu-trigger"
            text
            class="admin-profile-trigger admin-header-action dora-user-trigger h-10 px-3"
            :aria-expanded="isUserMenuOpen"
            aria-haspopup="menu"
            aria-controls="user-profile-menu"
            @click="toggleUserMenu"
            @keydown.down.prevent="!isUserMenuOpen && toggleUserMenu()"
          >
            <div class="admin-profile-avatar h-7 w-7 rounded-full flex items-center justify-center">
              <DoraIcon name="user-circle" :size="18" class="text-[var(--admin-text-muted)]" />
            </div>
            <span class="admin-profile-username mx-1 text-sm font-medium text-foreground max-w-[96px] truncate">{{ userAccount || 'admin' }}</span>
          </el-button>

          <div
            v-if="isUserMenuOpen"
            id="user-profile-menu"
            class="admin-context-menu admin-profile-menu absolute right-0 mt-2 w-56 rounded-xl text-popover-foreground outline-none animate-in fade-in-0 zoom-in-95"
            role="menu"
            :aria-label="messages.userMenu"
          >
            <div class="admin-profile-menu-header">
              <div class="admin-profile-avatar h-10 w-10 rounded-full flex items-center justify-center">
                <DoraIcon name="user-circle" :size="18" />
              </div>
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-foreground">{{ userAccount || 'admin' }}</div>
                <div class="text-xs text-[var(--admin-text-muted)]">{{ messages.currentUser }}</div>
              </div>
            </div>
            <button
              @click="openProfileSettings"
              class="admin-profile-menu-item"
              role="menuitem"
            >
              <DoraIcon name="setting" :size="14" />
              <span>{{ messages.profileSettings }}</span>
            </button>
            <button
              @click="requestLogout"
              class="admin-profile-menu-item danger"
              role="menuitem"
            >
              <DoraIcon name="logout" :size="14" />
              <span>{{ messages.logout }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
