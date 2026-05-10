<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  DashboardOutlined,
  MessageOutlined,
  CalendarOutlined,
  CodeOutlined,
  SendOutlined,
  FileTextOutlined,
  SettingOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  KeyOutlined,
  DownOutlined,
  DatabaseOutlined,
  NotificationOutlined,
  AppstoreOutlined,
  FileSearchOutlined,
  LoginOutlined,
  UserOutlined
} from '@ant-design/icons-vue'
import { cn } from '@/lib/utils'
import { useRbacAuthzStore } from '@/store/rbac_authz'

const props = defineProps<{
  isCollapsed: boolean
  siteTitle: string
  siteSlogan: string
  siteSloganInitialEnabled: boolean
  userAccount: string
}>()

const route = useRoute()
const router = useRouter()
const rbacAuthzStore = useRbacAuthzStore()
const openMenus = ref<Record<string, boolean>>({})

interface MenuItem {
  title: string
  path?: string
  name?: string
  icon: any
  requiredPermissions?: string[]
  children?: MenuItem[]
}

const menuItems: MenuItem[] = [
  {
    title: '数据统计',
    path: '/',
    icon: DashboardOutlined,
    requiredPermissions: ['dashboard:view']
  },
  {
    title: '消息管理',
    icon: MessageOutlined,
    children: [
      {
        title: '定时消息',
        path: '/cronmessages',
        name: 'cronmessages',
        icon: CalendarOutlined,
        requiredPermissions: ['message:cron:view']
      },
      {
        title: '订阅消息',
        path: '/message/subscriptions',
        name: 'message-subscriptions',
        icon: NotificationOutlined,
        requiredPermissions: ['data:subscription:view']
      }
    ]
  },
  {
    title: '模板管理',
    path: '/templates',
    icon: CodeOutlined,
    requiredPermissions: ['message:template:view']
  },
  {
    title: '渠道管理',
    path: '/sendways',
    icon: SendOutlined,
    requiredPermissions: ['message:sendways:view']
  },
  {
    title: '数据管理',
    icon: DatabaseOutlined,
    requiredPermissions: ['data:mq-source:view'],
    children: [
      {
        title: '消息队列',
        path: '/data/mq-sources',
        name: 'data-mq-sources',
        icon: AppstoreOutlined,
        requiredPermissions: ['data:mq-source:view']
      }
    ]
  },
  {
    title: '日志管理',
    icon: FileTextOutlined,
    requiredPermissions: ['message:sendlogs:view'],
    children: [
      {
        title: '任务日志',
        path: '/logs/task',
        icon: FileSearchOutlined,
        requiredPermissions: ['message:sendlogs:view']
      },
      {
        title: '登录日志',
        path: '/logs/login',
        icon: LoginOutlined,
        requiredPermissions: ['system:loginlogs:view']
      },
      {
        title: '消费日志',
        path: '/logs/consume',
        icon: FileTextOutlined,
        requiredPermissions: ['data:consume-log:view']
      }
    ]
  },
  {
    title: '系统管理',
    icon: SafetyCertificateOutlined,
    children: [
      {
        title: '用户管理',
        path: '/system/users',
        name: 'system-users',
        icon: UserOutlined,
        requiredPermissions: ['system:rbac:user']
      },
      {
        title: '用户组管理',
        path: '/system/groups',
        name: 'system-groups',
        icon: TeamOutlined,
        requiredPermissions: ['system:rbac:group']
      },
      {
        title: '角色管理',
        path: '/system/roles',
        name: 'system-roles',
        icon: SafetyCertificateOutlined,
        requiredPermissions: ['system:rbac:role']
      },
      {
        title: '权限管理',
        path: '/system/permissions',
        name: 'system-permissions',
        icon: KeyOutlined,
        requiredPermissions: ['system:rbac:permission']
      },
      {
        title: '系统设置',
        path: '/system/settings',
        name: 'system-settings',
        icon: SettingOutlined,
        requiredPermissions: ['system:settings:view']
      }
    ]
  }
]

const canAccessMenuItem = (item: MenuItem) => {
  if (!item.requiredPermissions || item.requiredPermissions.length === 0) {
    return true
  }
  return rbacAuthzStore.hasAnyPermission(item.requiredPermissions)
}

const filteredMenuItems = computed(() => {
  return menuItems
    .map((item) => {
      if (!item.children || item.children.length === 0) {
        return canAccessMenuItem(item) ? item : null
      }
      const visibleChildren = item.children.filter(canAccessMenuItem)
      if (visibleChildren.length === 0) {
        return null
      }
      return {
        ...item,
        children: visibleChildren
      }
    })
    .filter((item): item is MenuItem => Boolean(item))
})

const isActive = (item: MenuItem) => {
  if (item.path) {
    if (item.path === '/') {
      return route.path === '/'
    }
    return route.path.startsWith(item.path)
  }
  if (item.children) {
    return item.children.some(child => child.path && route.path.startsWith(child.path))
  }
  return false
}

const handleItemClick = (item: MenuItem) => {
  if (item.name) {
    router.push({ name: item.name })
    return
  }
  if (item.path) {
    router.push(item.path)
    return
  }
  if (item.children) {
    const key = item.title
    openMenus.value[key] = !openMenus.value[key]
  }
}

const isGroupOpen = (item: MenuItem) => {
  return !!openMenus.value[item.title]
}

watch(
  () => route.path,
  (path) => {
    filteredMenuItems.value.forEach((item) => {
      if (!item.children) return
      const matched = item.children.some(child => child.path && path.startsWith(child.path))
      if (matched) {
        openMenus.value[item.title] = true
      }
    })
  },
  { immediate: true }
)

const sidebarClass = computed(() => {
  return props.isCollapsed ? 'w-[60px]' : 'w-[200px]'
})

const collapsedBrandText = computed(() => {
  if (!props.siteSloganInitialEnabled) return 'M'
  const slogan = (props.siteSlogan || '').trim()
  if (!slogan) return 'M'
  const latinOrDigit = slogan.match(/[A-Za-z0-9]/)
  if (latinOrDigit?.[0]) {
    return latinOrDigit[0].toUpperCase()
  }
  const firstChar = slogan.charAt(0).trim()
  return firstChar ? firstChar.toUpperCase() : 'M'
})

const activeMainClass = 'text-white bg-[linear-gradient(90deg,rgba(35,114,229,0.9)_0%,rgba(32,106,220,0.84)_26%,rgba(27,94,205,0.72)_52%,rgba(22,82,185,0.56)_74%,rgba(18,72,166,0.38)_90%,rgba(15,64,149,0.24)_100%)]'
const activeGroupClass = 'text-white bg-[linear-gradient(90deg,rgba(31,104,216,0.8)_0%,rgba(28,97,207,0.74)_26%,rgba(24,86,191,0.62)_52%,rgba(20,76,174,0.48)_74%,rgba(16,66,156,0.32)_90%,rgba(13,58,140,0.2)_100%)]'
const activeChildClass = 'text-white bg-[linear-gradient(90deg,rgba(33,109,222,0.84)_0%,rgba(30,102,213,0.78)_26%,rgba(26,90,197,0.66)_52%,rgba(21,79,179,0.52)_74%,rgba(17,69,160,0.35)_90%,rgba(14,60,143,0.22)_100%)]'
const menuIconClass = 'text-[17px] leading-none relative z-[1]'

</script>

<template>
  <aside
    :class="cn(
      'sidebar fixed left-0 top-0 z-50 h-screen bg-[var(--sidebar-bg,#001529)] flex flex-col text-white transition-[width] duration-[var(--motion-normal)] ease-in-out',
      sidebarClass
    )"
  >
    <div class="flex items-center h-14 px-4">
      <div v-if="!isCollapsed" class="text-[16px] font-bold text-brand truncate">
        {{ siteTitle }}
      </div>
      <div
        v-else
        class="w-6 h-6 bg-brand text-white rounded flex items-center justify-center font-bold text-[18px]"
      >
        {{ collapsedBrandText }}
      </div>
    </div>
    <div class="h-px bg-white/10" />

    <div class="flex-1 overflow-y-auto py-2">
      <nav class="space-y-1 px-2 text-[14px]">
        <template v-for="(item, index) in filteredMenuItems" :key="index">
          <div v-if="!item.children">
            <button
              @click="handleItemClick(item)"
              :class="cn(
                'w-full flex items-center gap-3 px-3 py-2 rounded-md transition-all duration-[var(--motion-fast)] relative overflow-hidden',
                isActive(item)
                  ? activeMainClass
                  : 'text-white/85 hover:bg-white/10',
                isCollapsed ? 'justify-center' : 'justify-start'
              )"
            >
              <component :is="item.icon" :class="menuIconClass" />
              <span v-if="!isCollapsed" class="truncate relative z-[1]">{{ item.title }}</span>
            </button>
          </div>

          <div v-else>
            <button
              @click="handleItemClick(item)"
              :class="cn(
                'w-full flex items-center gap-3 px-3 py-2 rounded-md transition-all duration-[var(--motion-fast)] relative overflow-hidden',
                isActive(item)
                  ? activeGroupClass
                  : 'text-white/85 hover:bg-white/10',
                isCollapsed ? 'justify-center' : 'justify-start'
              )"
            >
              <component :is="item.icon" :class="menuIconClass" />
              <span v-if="!isCollapsed" class="flex-1 text-left truncate relative z-[1]">{{ item.title }}</span>
              <component
                v-if="!isCollapsed"
                :is="DownOutlined"
                :class="cn(
                  'w-3.5 h-3.5 text-white/70 transition-transform duration-[var(--motion-fast)] relative z-[1]',
                  isGroupOpen(item) ? 'rotate-180' : ''
                )"
              />
            </button>
            <div v-if="isGroupOpen(item)" class="mt-1 space-y-1">
              <button
                v-for="(child, childIndex) in item.children"
                :key="childIndex"
                @click="handleItemClick(child)"
                :class="cn(
                  'w-full flex items-center gap-3 px-3 py-2 rounded-md transition-all duration-[var(--motion-fast)] relative overflow-hidden',
                  isActive(child)
                    ? activeChildClass
                    : 'text-white/75 hover:bg-white/10',
                  isCollapsed ? 'justify-center' : 'justify-start',
                  !isCollapsed ? 'pl-9' : ''
                )"
              >
                <component :is="child.icon" :class="menuIconClass" />
                <span v-if="!isCollapsed" class="truncate relative z-[1]">{{ child.title }}</span>
              </button>
            </div>
          </div>
        </template>
      </nav>
    </div>

  </aside>
</template>
