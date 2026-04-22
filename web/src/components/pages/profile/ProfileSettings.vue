<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { KeyRound, Palette } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const menu = [
  { id: 'password', name: '修改密码', icon: KeyRound, path: '/profile/settings/password' },
  { id: 'preference', name: '个性设置', icon: Palette, path: '/profile/settings/preference' }
]

const titleMap: Record<string, string> = {
  password: '修改密码',
  preference: '个性设置'
}
const descMap: Record<string, string> = {
  password: '修改你的登录密码',
  preference: '设置个人主题颜色、显示模式与侧边栏样式'
}
const activeTab = computed(() => {
  const current = menu.find(item => route.path.startsWith(item.path))
  return current?.id || 'password'
})
const activeTitle = computed(() => titleMap[activeTab.value] || '')
const activeDescription = computed(() => descMap[activeTab.value] || '')

const handleClose = () => {
  router.back()
}

const handleOpen = (path: string) => {
  if (route.path === path) {
    return
  }
  router.push(path)
}
</script>

<template>
  <div class="p-4 lg:p-6 w-full profile-settings">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-[18px] font-semibold text-[#1f2937] dark:text-slate-100">个人设置</h1>
      <button
        type="button"
        class="w-7 h-7 flex items-center justify-center rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-slate-800"
        @click="handleClose"
        aria-label="关闭个人设置"
      >
        ×
      </button>
    </div>
    <div class="flex flex-col lg:flex-row gap-4 min-h-[calc(100vh-140px)] h-full w-full">
      <div class="left-nav lg:w-[240px] lg:flex-shrink-0 w-full">
        <div class="space-y-2">
          <button
            v-for="item in menu"
            :key="item.id"
            type="button"
            class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
            :class="activeTab === item.id ? 'bg-[#3b82f6] text-white' : 'text-gray-700 dark:text-slate-200 hover:bg-[#f1f5f9] dark:hover:bg-slate-800'"
            @click="handleOpen(item.path)"
          >
            <component :is="item.icon" class="w-4 h-4" />
            <span>{{ item.name }}</span>
          </button>
        </div>
      </div>
      <div class="right-content flex-1 min-w-0 w-full lg:border-l border-[#e5e7eb] dark:border-slate-700 lg:pl-5 mt-4 lg:mt-0 flex flex-col min-h-0">
        <transition name="settings-fade" mode="out-in">
          <div :key="route.path" class="flex-1 flex flex-col gap-4">
            <div v-if="activeTitle" class="space-y-1">
              <h2 class="text-[16px] font-semibold text-[#1f2937] dark:text-slate-100">{{ activeTitle }}</h2>
              <p class="text-[12px] text-[#6b7280] dark:text-slate-400">{{ activeDescription }}</p>
            </div>
            <div class="flex-1 min-h-0">
              <router-view />
            </div>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>
