<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import {
  SearchOutlined,
  BellOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  RightOutlined,
  DownOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined
} from '@ant-design/icons-vue'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  userAccount: string
  theme: 'light' | 'dark'
  themePreference: 'light' | 'dark' | 'system'
  breadcrumb?: string
  isSidebarCollapsed?: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-theme'): void
  (e: 'toggle-sidebar'): void
  (e: 'open-profile-settings'): void
  (e: 'logout'): void
}>()

const isUserMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)

const toggleUserMenu = () => {
  isUserMenuOpen.value = !isUserMenuOpen.value
}

const openProfileSettings = () => {
  isUserMenuOpen.value = false
  emit('open-profile-settings')
}

const handleClickOutside = (event: MouseEvent) => {
  if (userMenuRef.value && !userMenuRef.value.contains(event.target as Node)) {
    isUserMenuOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const breadcrumbItems = computed(() => {
  const text = props.breadcrumb?.trim() || '工作台'
  return text
    .split('/')
    .map((item) => item.trim())
    .filter(Boolean)
})
</script>

<template>
  <header class="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <div class="flex h-14 items-center px-4 md:px-6 gap-3">
      <div class="min-w-[220px] flex-1 flex items-center gap-2 overflow-hidden">
        <Button
          variant="ghost"
          size="icon"
          class="h-11 w-11 rounded-md text-[#4a5a7f] hover:bg-[#eef3ff] hover:text-[#36589a] shrink-0"
          :title="props.isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
          @click="emit('toggle-sidebar')"
        >
          <MenuUnfoldOutlined v-if="props.isSidebarCollapsed" class="text-[20px] leading-none [&_svg]:!w-[20px] [&_svg]:!h-[20px]" />
          <MenuFoldOutlined v-else class="text-[20px] leading-none [&_svg]:!w-[20px] [&_svg]:!h-[20px]" />
          <span class="sr-only">切换侧边栏</span>
        </Button>
        <div class="flex items-center gap-1 text-sm font-medium text-foreground truncate">
          <template v-for="(item, idx) in breadcrumbItems" :key="`${item}-${idx}`">
            <RightOutlined v-if="idx > 0" class="text-[12px] text-muted-foreground shrink-0" />
            <span
              :class="idx === breadcrumbItems.length - 1 ? 'text-foreground' : 'text-muted-foreground'"
              class="truncate"
            >
              {{ item }}
            </span>
          </template>
          </div>
      </div>

      <div class="flex justify-end">
        <div class="flex items-center gap-2">
          <slot name="actions" />
        </div>
      </div>

      <div class="ml-2 flex items-center gap-2.5">
        <Button
          variant="ghost"
          size="icon"
          class="h-11 w-11 rounded-full text-[#4f618c] hover:text-[#36589a] hover:bg-[#eef3ff]"
        >
          <SearchOutlined class="text-[20px] [&_svg]:!w-[20px] [&_svg]:!h-[20px]" />
          <span class="sr-only">搜索</span>
        </Button>

        <Button
          variant="ghost"
          size="icon"
          class="relative h-11 w-11 rounded-full text-[#4f618c] hover:text-[#36589a] hover:bg-[#eef3ff]"
        >
          <BellOutlined class="text-[20px] [&_svg]:!w-[20px] [&_svg]:!h-[20px]" />
          <span class="sr-only">通知</span>
        </Button>

        <div class="relative" ref="userMenuRef">
          <Button 
            variant="ghost" 
            class="ml-1 h-11 rounded-full px-3 text-[#4f618c] hover:bg-[#eef3ff]"
            @click="toggleUserMenu"
          >
            <div class="h-9 w-9 rounded-full bg-[#e9eef8] flex items-center justify-center">
              <UserOutlined class="text-[18px] text-[#8c9bb8] [&_svg]:!w-[18px] [&_svg]:!h-[18px]" />
            </div>
            <span class="mx-1 text-sm font-medium text-[#3c4d72] max-w-[96px] truncate">{{ userAccount || 'admin' }}</span>
            <DownOutlined class="text-[14px] text-[#6f81a8] [&_svg]:!w-[14px] [&_svg]:!h-[14px]" />
          </Button>

          <div 
            v-if="isUserMenuOpen"
            class="absolute left-1/2 mt-2 w-36 -translate-x-1/2 rounded-md border bg-popover text-popover-foreground shadow-md outline-none animate-in fade-in-0 zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2"
          >
            <button
              @click="openProfileSettings"
              class="relative flex w-full cursor-pointer select-none items-center justify-center gap-2 rounded-sm px-2 py-1 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50"
            >
              <SettingOutlined class="text-[14px]" />
              <span>个人设置</span>
            </button>
            <button
              @click="isUserMenuOpen = false; emit('logout')"
              class="relative flex w-full cursor-pointer select-none items-center justify-center gap-2 rounded-sm px-2 py-1 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 text-red-600 focus:text-red-600"
            >
              <LogoutOutlined class="text-[14px]" />
              <span>退出登录</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
