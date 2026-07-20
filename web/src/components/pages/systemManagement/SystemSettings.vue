<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRbacStore } from '@/store/rbac'
import {
  SettingOutlined,
  DeleteOutlined,
  KeyOutlined,
  InfoCircleOutlined,
  SafetyCertificateOutlined,
  DatabaseOutlined,
  LineChartOutlined
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const rbacStore = useRbacStore()
const menu = [
  { id: 'site', name: '站点设置', icon: SettingOutlined, path: '/system/settings/site', routeName: 'system-settings-site', requiredPermissions: ['system:settings:view'] },
  { id: 'auth', name: '认证设置', icon: SafetyCertificateOutlined, path: '/system/settings/auth', routeName: 'system-settings-auth', requiredPermissions: ['system:settings:view'] },
  { id: 'storage', name: '存储配置', icon: DatabaseOutlined, path: '/system/settings/storage', routeName: 'system-settings-storage', requiredPermissions: ['system:settings:view'] },
  { id: 'clean', name: '数据清理', icon: DeleteOutlined, path: '/system/settings/clean', routeName: 'system-settings-clean', requiredPermissions: ['system:settings:view'] },
  { id: 'mqStatusPolicy', name: '策略配置', icon: LineChartOutlined, path: '/system/settings/mq-status-policy', routeName: 'system-settings-mq-status-policy', requiredPermissions: ['system:settings:view'] },
  { id: 'tokenTool', name: '加解密工具', icon: KeyOutlined, path: '/system/settings/token-tool', routeName: 'system-settings-token-tool', requiredPermissions: ['system:settings:view'] },
  { id: 'about', name: '站点关于', icon: InfoCircleOutlined, path: '/system/settings/about', routeName: 'system-settings-about', requiredPermissions: ['system:settings:view'] }
]

const visibleMenu = computed(() => menu.filter(item => rbacStore.hasAnyPermission(item.requiredPermissions)))

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
  const routeName = String(route.name || '')
  const current = visibleMenu.value.find(item => {
    if (routeName === item.routeName) return true
    return route.path === item.path || route.path.startsWith(`${item.path}/`)
  })
  return current?.id || 'site'
})
const activeTitle = computed(() => titleMap[activeTab.value] || '')
const activeDescription = computed(() => descMap[activeTab.value] || '')
const hideParentHeaderTabs = new Set(['clean', 'tokenTool'])
const showParentHeader = computed(() => !hideParentHeaderTabs.has(activeTab.value))

const handleOpen = (item: typeof menu[number]) => {
  if (route.name === item.routeName || route.path === item.path) {
    return
  }
  router.push({ name: item.routeName })
}

onMounted(() => {
  if (route.path === '/system/settings' || route.path === '/system/settings/') {
    router.replace(visibleMenu.value[0]?.path || '/404')
  }
})
</script>

<template>
  <div class="system-settings app-settings-shell">
    <div class="app-settings-layout">
      <div class="app-settings-nav">
        <div class="space-y-1">
          <button
            v-for="item in visibleMenu"
            :key="item.id"
            type="button"
            class="app-settings-nav-item"
            :class="activeTab === item.id ? 'app-settings-nav-item-active' : 'app-settings-nav-item-idle'"
            @click="handleOpen(item)"
          >
            <component :is="item.icon" class="w-4 h-4" />
            <span>{{ item.name }}</span>
          </button>
        </div>
      </div>
      <div class="app-settings-content">
        <transition name="settings-fade" mode="out-in">
          <div :key="route.path" class="app-settings-panel-scroll">
            <div v-if="showParentHeader && activeTitle" class="app-settings-section-heading">
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
