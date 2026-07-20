<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { KeyOutlined } from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const menu = [
  { id: 'password', name: '修改密码', icon: KeyOutlined, path: '/profile/settings/password' }
]

const titleMap: Record<string, string> = {
  password: '修改密码'
}
const descMap: Record<string, string> = {
  password: '修改你的登录密码'
}
const activeTab = computed(() => {
  const current = menu.find(item => route.path.startsWith(item.path))
  return current?.id || 'password'
})
const activeTitle = computed(() => titleMap[activeTab.value] || '')
const activeDescription = computed(() => descMap[activeTab.value] || '')

const handleOpen = (path: string) => {
  if (route.path === path) {
    return
  }
  router.push(path)
}
</script>

<template>
  <div class="profile-settings app-profile-shell">
    <div class="app-profile-card">
      <div class="app-profile-layout">
      <div class="app-profile-nav">
        <div class="space-y-1">
          <el-button
            v-for="item in menu"
            :key="item.id"
            text
            class="app-profile-nav-item"
            :class="activeTab === item.id ? 'app-profile-nav-item-active' : 'app-profile-nav-item-idle'"
            @click="handleOpen(item.path)"
          >
            <component :is="item.icon" class="w-4 h-4" />
            <span>{{ item.name }}</span>
          </el-button>
        </div>
      </div>
      <div class="app-profile-content">
        <transition name="settings-fade" mode="out-in">
          <div :key="route.path" class="app-profile-panel-scroll">
            <div v-if="activeTitle" class="app-profile-section-heading">
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
  </div>
</template>
