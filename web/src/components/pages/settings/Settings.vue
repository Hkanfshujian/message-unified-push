<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import SettingsSidebar from './SettingsSidebar.vue'
import PasswordSettings from './PasswordSettings.vue'
import CleanSettings from './CleanSettings.vue'
import SiteSettings from './SiteSettings.vue'
import TokenToolSettings from './TokenToolSettings.vue'
import AboutSettings from './AboutSettings.vue'
import LoginLogs from './LoginLogs.vue'

const activeTab = ref('password')
const router = useRouter()

const activeTitle = computed(() => {
  switch (activeTab.value) {
    case 'password':
      return '重置密码'
    case 'clean':
      return '数据清理'
    case 'loginlogs':
      return '登录日志'
    case 'site':
      return '站点设置'
    case 'tokenTool':
      return '加解密工具'
    case 'about':
      return '站点关于'
    default:
      return ''
  }
})

const activeDescription = computed(() => {
  switch (activeTab.value) {
    case 'password':
      return '更改您的登录密码'
    case 'clean':
      return '清理历史数据与日志'
    case 'loginlogs':
      return '查看最近的登录记录'
    case 'site':
      return '配置站点标题、描述等基础信息'
    case 'tokenTool':
      return '管理和测试 Token 编解码工具'
    case 'about':
      return '查看当前站点的版本信息与说明'
    default:
      return ''
  }
})

const handleClose = () => {
  router.back()
}
</script>

<template>
  <div class="p-6 w-full max-w-6xl mx-auto system-settings">
    <!-- 顶部标题与关闭按钮 -->
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-[18px] font-semibold text-[#1f2937] dark:text-slate-100">
        系统设置
      </h1>
      <button
        type="button"
        class="w-7 h-7 flex items-center justify-center rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-slate-800"
        @click="handleClose"
        aria-label="关闭系统设置"
      >
        ×
      </button>
    </div>

    <div class="flex flex-col lg:flex-row gap-6 min-h-[520px] h-full">
      <div class="left-nav lg:flex-[1] lg:max-w-xs w-full">
        <SettingsSidebar
          :active-tab="activeTab"
          @update:active-tab="activeTab = $event"
        />
      </div>

      <div
        class="right-content lg:flex-[3] w-full lg:border-l border-[#e5e7eb] dark:border-slate-700 lg:pl-6 mt-6 lg:mt-0 flex flex-col min-h-0"
      >
        <transition name="settings-fade" mode="out-in">
          <div :key="activeTab" class="flex-1 flex flex-col gap-4">
            <div v-if="activeTitle" class="space-y-1">
              <h2 class="text-[16px] font-semibold text-[#1f2937] dark:text-slate-100">
                {{ activeTitle }}
              </h2>
              <p class="text-[12px] text-[#6b7280] dark:text-slate-400">
                {{ activeDescription }}
              </p>
            </div>

            <div class="flex-1 min-h-0">
              <PasswordSettings v-if="activeTab === 'password'" />
              <CleanSettings v-else-if="activeTab === 'clean'" />
              <SiteSettings v-else-if="activeTab === 'site'" />
              <TokenToolSettings v-else-if="activeTab === 'tokenTool'" />
              <LoginLogs v-else-if="activeTab === 'loginlogs'" />
              <AboutSettings v-else-if="activeTab === 'about'" />
            </div>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-fade-enter-active,
.settings-fade-leave-active {
  transition: opacity 0.15s ease-out, transform 0.15s ease-out;
}

.settings-fade-enter-from,
.settings-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
