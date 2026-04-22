<script setup lang="ts">
import { KeyIcon, TrashIcon, SettingsIcon, InfoIcon, HistoryIcon, KeyRoundIcon, Check } from 'lucide-vue-next'

interface Props {
  activeTab: string
}

interface Emits {
  (e: 'update:activeTab', value: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const settingsMenu = [
  { id: 'password', name: '重置密码', icon: KeyIcon, description: '更改您的登录密码' },
  { id: 'clean', name: '数据清理', icon: TrashIcon, description: '清理历史数据与日志' },
  { id: 'loginlogs', name: '登录日志', icon: HistoryIcon, description: '查看最近登录记录' },
  { id: 'site', name: '站点设置', icon: SettingsIcon, description: '配置站点标题与基础信息' },
  { id: 'tokenTool', name: '加解密工具', icon: KeyRoundIcon, description: '管理 Token 编解码工具' },
  { id: 'about', name: '站点关于', icon: InfoIcon, description: '查看版本信息与说明' }
]

const handleClick = (id: string) => {
  emit('update:activeTab', id)
}

const handleKeydown = (event: KeyboardEvent) => {
  const index = settingsMenu.findIndex(item => item.id === props.activeTab)
  const length = settingsMenu.length
  if (length === 0) return
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    const next = (index + 1 + length) % length
    emit('update:activeTab', settingsMenu[next].id)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    const prev = (index - 1 + length) % length
    emit('update:activeTab', settingsMenu[prev].id)
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    const current = index >= 0 ? settingsMenu[index].id : settingsMenu[0].id
    emit('update:activeTab', current)
  }
}
</script>

<script lang="ts">
export default {
  name: 'SettingsSidebar'
}
</script>

<template>
  <div class="w-full h-full">
    <div class="mb-4">
      <h2 class="text-[16px] font-semibold text-[#1f2937] dark:text-slate-100">
        设置
      </h2>
      <p class="mt-1 text-[12px] text-[#6b7280] dark:text-slate-400">
        管理系统配置和偏好
      </p>
    </div>

    <nav
      class="space-y-2 pr-1 max-h-[calc(100vh-240px)] overflow-y-auto"
      tabindex="0"
      @keydown="handleKeydown"
    >
      <button
        v-for="item in settingsMenu"
        :key="item.id"
        type="button"
        @click="handleClick(item.id)"
        :class="[
          'w-full flex items-center justify-between px-3 py-2 h-10 rounded-md text-sm transition-colors relative',
          props.activeTab === item.id
            ? 'bg-[#3b82f6] text-white'
            : 'text-gray-700 dark:text-slate-200 hover:bg-[#f1f5f9] dark:hover:bg-slate-800'
        ]"
      >
        <div class="flex items-center gap-2">
          <component :is="item.icon" class="w-4 h-4" />
          <span>{{ item.name }}</span>
        </div>
        <div
          v-if="props.activeTab === item.id"
          class="flex items-center justify-center"
        >
          <Check class="w-3.5 h-3.5 text-white" />
        </div>
      </button>
    </nav>
  </div>
</template>
