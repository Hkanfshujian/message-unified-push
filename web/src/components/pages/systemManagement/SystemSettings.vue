<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Settings, Trash2, KeyRound, Info, ShieldCheck, Database, Activity } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const menu = [
  { id: 'site', name: '站点设置', icon: Settings, path: '/system/settings/site' },
  { id: 'auth', name: '认证设置', icon: ShieldCheck, path: '/system/settings/auth' },
  { id: 'storage', name: '存储配置', icon: Database, path: '/system/settings/storage' },
  { id: 'clean', name: '数据清理', icon: Trash2, path: '/system/settings/clean' },
  { id: 'mqStatusPolicy', name: '策略配置', icon: Activity, path: '/system/settings/mq-status-policy' },
  { id: 'tokenTool', name: '加解密工具', icon: KeyRound, path: '/system/settings/token-tool' },
  { id: 'about', name: '站点关于', icon: Info, path: '/system/settings/about' }
]

const titleMap: Record<string, string> = {
  clean: '数据清理',
  site: '站点设置',
  auth: '认证设置',
  storage: '存储配置',
  tokenTool: '加解密工具',
  about: '站点关于',
  mqStatusPolicy: '策略配置'
}

const descMap: Record<string, string> = {
  clean: '清理历史数据与日志',
  site: '配置站点标题、描述等基础信息',
  auth: '配置注册开关、OIDC策略与回调重试参数',
  storage: '配置站点级静态资源存储驱动（本地/S3）',
  tokenTool: '管理和测试 Token 编解码工具',
  about: '查看当前站点的版本信息与说明',
  mqStatusPolicy: '维护消息队列状态的手动/自动更新策略及频率'
}

const activeTab = computed(() => {
  const current = menu.find(item => route.path.startsWith(item.path))
  return current?.id || 'site'
})
const activeTitle = computed(() => titleMap[activeTab.value] || '')
const activeDescription = computed(() => descMap[activeTab.value] || '')
const hideParentHeaderTabs = new Set(['clean', 'tokenTool'])
const showParentHeader = computed(() => !hideParentHeaderTabs.has(activeTab.value))

const handleClose = () => {
  router.back()
}

const handleOpen = (path: string) => {
  if (route.path === path) {
    return
  }
  router.push(path)
}

onMounted(() => {
  // 兜底：首次进入 /system/settings 时强制跳默认子页，避免内容区空白
  if (route.path === '/system/settings' || route.path === '/system/settings/') {
    router.replace('/system/settings/site')
  }
})
</script>

<template>
  <div class="p-4 lg:p-6 w-full system-settings h-full flex flex-col overflow-hidden">
    <div class="flex items-center justify-between mb-4 flex-shrink-0">
      <h1 class="text-[18px] font-semibold text-[#1f2937] dark:text-slate-100">系统设置</h1>
      <button
        type="button"
        class="w-7 h-7 flex items-center justify-center rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-slate-800"
        @click="handleClose"
        aria-label="关闭系统设置"
      >
        ×
      </button>
    </div>
    <div class="flex flex-col lg:flex-row gap-4 flex-1 min-h-0 overflow-hidden">
      <!-- 左侧菜单：固定不滚动 -->
      <div class="left-nav lg:w-[240px] lg:flex-shrink-0 w-full overflow-y-auto lg:overflow-visible">
        <div class="space-y-2">
          <button
            v-for="item in menu"
            :key="item.id"
            type="button"
            class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
            :class="activeTab === item.id ? 'bg-brand text-white' : 'text-gray-700 dark:text-slate-200 hover:bg-[#f1f5f9] dark:hover:bg-slate-800'"
            @click="handleOpen(item.path)"
          >
            <component :is="item.icon" class="w-4 h-4" />
            <span>{{ item.name }}</span>
          </button>
        </div>
      </div>
      <!-- 右侧内容区：独立滚动 -->
      <div class="right-content flex-1 min-w-0 w-full lg:border-l border-[#e5e7eb] dark:border-slate-700 lg:pl-5 mt-4 lg:mt-0 flex flex-col min-h-0 overflow-hidden">
        <transition name="settings-fade" mode="out-in">
          <div :key="route.path" class="flex-1 flex flex-col gap-4 overflow-y-auto pr-2">
            <div v-if="showParentHeader && activeTitle" class="space-y-1 flex-shrink-0">
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
