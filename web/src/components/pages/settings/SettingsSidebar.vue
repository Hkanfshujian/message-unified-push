<script setup lang="ts">
import {
  KeyOutlined,
  DeleteOutlined,
  SettingOutlined,
  InfoCircleOutlined,
  HistoryOutlined,
  SafetyCertificateOutlined,
  CheckOutlined,
  LoginOutlined,
  CloudServerOutlined,
  ApiOutlined
} from '@ant-design/icons-vue'

interface Props {
  activeTab: string
}

interface Emits {
  (e: 'update:activeTab', value: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const settingsMenu = [
  { id: 'password', name: '重置密码', icon: KeyOutlined, description: '更改您的登录密码' },
  { id: 'clean', name: '数据清理', icon: DeleteOutlined, description: '清理历史数据与日志' },
  { id: 'loginlogs', name: '登录日志', icon: HistoryOutlined, description: '查看最近登录记录' },
  { id: 'site', name: '站点设置', icon: SettingOutlined, description: '配置站点标题与基础信息' },
  { id: 'auth', name: '认证设置', icon: LoginOutlined, description: '配置注册与单点登录' },
  { id: 'storage', name: '存储设置', icon: CloudServerOutlined, description: '管理文件存储配置' },
  { id: 'mqStatusPolicy', name: 'MQ 状态策略', icon: ApiOutlined, description: '配置 MQ 订阅状态策略' },
  { id: 'tokenTool', name: '加解密工具', icon: SafetyCertificateOutlined, description: '管理 Token 编解码工具' },
  { id: 'about', name: '站点关于', icon: InfoCircleOutlined, description: '查看版本信息与说明' }
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
    <div class="mb-4 hidden lg:block">
      <h2 class="text-[16px] font-semibold text-foreground">
        设置
      </h2>
      <p class="mt-1 text-[12px] text-muted-foreground">
        管理系统配置和偏好
      </p>
    </div>

    <nav
      class="space-y-2 pr-1 max-h-[calc(100vh-240px)] overflow-y-auto"
      tabindex="0"
      @keydown="handleKeydown"
    >
      <el-button
        v-for="item in settingsMenu"
        :key="item.id"
        text
        class="settings-menu-button"
        @click="handleClick(item.id)"
        :class="[
          'w-full h-auto min-h-12 justify-between rounded-2xl border transition-all duration-[var(--motion-fast)] relative',
          props.activeTab === item.id
            ? 'border-white/50 bg-[linear-gradient(135deg,var(--brand-600),#0ea5e9)] text-white shadow-[0_14px_30px_rgba(37,99,235,0.24)]'
            : 'border-transparent text-foreground/80 hover:border-[var(--dora-border)] hover:bg-[var(--brand-50)] hover:text-brand-700 hover:shadow-none'
        ]"
      >
        <div class="flex items-center gap-2 min-w-0 text-left">
          <component :is="item.icon" class="w-4 h-4" />
          <span class="truncate">{{ item.name }}</span>
        </div>
        <div
          v-if="props.activeTab === item.id"
          class="flex items-center justify-center"
        >
          <CheckOutlined class="text-[14px] text-white" />
        </div>
      </el-button>
    </nav>
  </div>
</template>

<style scoped>
:deep(.settings-menu-button.el-button) {
  margin-left: 0;
  padding: 10px 12px;
}

:deep(.settings-menu-button.el-button + .settings-menu-button.el-button) {
  margin-left: 0;
}
</style>
