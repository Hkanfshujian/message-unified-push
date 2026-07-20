<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SettingsSidebar from './SettingsSidebar.vue'
import PasswordSettings from './PasswordSettings.vue'
import CleanSettings from './CleanSettings.vue'
import SiteSettings from './SiteSettings.vue'
import AuthSettings from './AuthSettings.vue'
import StorageSettings from './StorageSettings.vue'
import MQStatusPolicySettings from './MQStatusPolicySettings.vue'
import TokenToolSettings from './TokenToolSettings.vue'
import AboutSettings from './AboutSettings.vue'
import LoginLogs from './LoginLogs.vue'

const settingsSections = [
  { id: 'password', title: '重置密码', description: '更改您的登录密码' },
  { id: 'clean', title: '数据清理', description: '清理历史数据与日志' },
  { id: 'loginlogs', title: '登录日志', description: '查看最近的登录记录' },
  { id: 'site', title: '站点设置', description: '配置站点标题、描述、Logo、分页与主题等基础信息' },
  { id: 'auth', title: '认证设置', description: '配置注册入口、默认用户组与 Casdoor 单点登录参数' },
  { id: 'storage', title: '存储设置', description: '管理本地与 S3 存储配置、默认存储和文件浏览能力' },
  { id: 'mqStatusPolicy', title: 'MQ 状态策略', description: '配置订阅状态检测、异常判定与恢复策略' },
  { id: 'tokenTool', title: '加解密工具', description: '管理和测试 Token 编解码工具' },
  { id: 'about', title: '站点关于', description: '查看当前站点的版本信息与说明' }
] as const

type SettingsSectionId = typeof settingsSections[number]['id']

const router = useRouter()
const route = useRoute()

const getTabFromQuery = (): SettingsSectionId => {
  const rawTab = Array.isArray(route.query.tab) ? route.query.tab[0] : route.query.tab
  return settingsSections.some(item => item.id === rawTab) ? rawTab as SettingsSectionId : 'password'
}

const activeTab = ref<SettingsSectionId>(getTabFromQuery())

watch(
  () => route.query.tab,
  () => {
    const nextTab = getTabFromQuery()
    if (activeTab.value !== nextTab) {
      activeTab.value = nextTab
    }
  }
)

watch(activeTab, tab => {
  const currentTab = Array.isArray(route.query.tab) ? route.query.tab[0] : route.query.tab
  if (currentTab === tab) return
  router.replace({ query: { ...route.query, tab } })
})

const activeSection = computed(() => {
  return settingsSections.find(item => item.id === activeTab.value) || settingsSections[0]
})

const handleTabChange = (tab: string) => {
  if (settingsSections.some(item => item.id === tab)) {
    activeTab.value = tab as SettingsSectionId
  }
}

</script>

<template>
  <div class="p-6 w-full max-w-7xl mx-auto system-settings">
    <el-card shadow="never" class="settings-shell-card">
      <div class="flex flex-col lg:flex-row gap-6 min-h-[560px] h-full">
        <div class="left-nav lg:w-[280px] w-full shrink-0">
        <SettingsSidebar
          :active-tab="activeTab"
          @update:active-tab="handleTabChange"
        />
      </div>

      <div
        class="right-content lg:flex-[3] w-full lg:border-l weak-divider lg:pl-6 mt-6 lg:mt-0 flex flex-col min-h-0"
      >
        <transition name="settings-fade" mode="out-in">
          <div :key="activeTab" class="flex-1 flex flex-col gap-4">
            <div class="space-y-1">
              <h2 class="text-[16px] font-semibold text-foreground">
                {{ activeSection.title }}
              </h2>
              <p class="text-[12px] text-muted-foreground">
                {{ activeSection.description }}
              </p>
            </div>

            <div class="flex-1 min-h-0">
              <PasswordSettings v-if="activeTab === 'password'" />
              <CleanSettings v-else-if="activeTab === 'clean'" />
              <SiteSettings v-else-if="activeTab === 'site'" />
              <AuthSettings v-else-if="activeTab === 'auth'" />
              <StorageSettings v-else-if="activeTab === 'storage'" />
              <MQStatusPolicySettings v-else-if="activeTab === 'mqStatusPolicy'" />
              <TokenToolSettings v-else-if="activeTab === 'tokenTool'" />
              <LoginLogs v-else-if="activeTab === 'loginlogs'" />
              <AboutSettings v-else-if="activeTab === 'about'" />
            </div>
          </div>
        </transition>
      </div>
    </div>
    </el-card>
  </div>
</template>

<style scoped>
:deep(.settings-shell-card > .el-card__body) {
  min-height: 620px;
}

.settings-fade-enter-active,
.settings-fade-leave-active {
  transition: opacity var(--motion-fast) ease-out, transform var(--motion-fast) ease-out;
}

.settings-fade-enter-from,
.settings-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
