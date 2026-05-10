<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { KeyOutlined, BgColorsOutlined } from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const menu = [
  { id: 'password', name: '修改密码', icon: KeyOutlined, path: '/profile/settings/password' },
  { id: 'preference', name: '个性设置', icon: BgColorsOutlined, path: '/profile/settings/preference' }
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
      <h1 class="text-[18px] font-semibold text-foreground">个人设置</h1>
      <button
        type="button"
        class="w-7 h-7 flex items-center justify-center rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors duration-[var(--motion-fast)]"
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
            class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors duration-[var(--motion-fast)]"
            :class="activeTab === item.id ? 'bg-brand text-white' : 'text-foreground/80 hover:bg-muted'"
            @click="handleOpen(item.path)"
          >
            <component :is="item.icon" class="w-4 h-4" />
            <span>{{ item.name }}</span>
          </button>
        </div>
      </div>
      <div class="right-content flex-1 min-w-0 w-full lg:border-l weak-divider lg:pl-5 mt-4 lg:mt-0 flex flex-col min-h-0">
        <transition name="settings-fade" mode="out-in">
          <div :key="route.path" class="flex-1 flex flex-col gap-4">
            <div v-if="activeTitle" class="space-y-1">
              <h2 class="text-[16px] font-semibold text-foreground">{{ activeTitle }}</h2>
              <p class="text-[12px] text-muted-foreground">{{ activeDescription }}</p>
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
