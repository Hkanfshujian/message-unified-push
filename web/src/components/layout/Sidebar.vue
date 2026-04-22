<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  LayoutDashboard,
  MessageSquare,
  Clock,
  FileText,
  Share2,
  ScrollText,
  Settings,
  Shield,
  Users,
  KeyRound,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  User,
  Database,
  Rss
} from 'lucide-vue-next'
import { cn } from '@/lib/utils'
import { useRbacAuthzStore } from '@/store/rbac_authz'

const props = defineProps<{
  isCollapsed: boolean
  siteTitle: string
  siteSlogan: string
  siteSloganInitialEnabled: boolean
  userAccount: string
}>()

const emit = defineEmits<{
  (e: 'toggle-collapse'): void
  (e: 'logout'): void
}>()

const route = useRoute()
const router = useRouter()
const rbacAuthzStore = useRbacAuthzStore()
const openMenus = ref<Record<string, boolean>>({})
const isUserMenuOpen = ref(false)

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
    icon: LayoutDashboard,
    requiredPermissions: ['dashboard:view']
  },
  {
    title: '消息管理',
    icon: MessageSquare,
    children: [
      {
        title: '定时消息',
        path: '/cronmessages',
        name: 'cronmessages',
        icon: Clock,
        requiredPermissions: ['message:cron:view']
      },
      {
        title: '订阅消息',
        path: '/message/subscriptions',
        name: 'message-subscriptions',
        icon: Rss,
        requiredPermissions: ['data:subscription:view']
      }
    ]
  },
  {
    title: '模板管理',
    path: '/templates',
    icon: FileText,
    requiredPermissions: ['message:template:view']
  },
  {
    title: '渠道管理',
    path: '/sendways',
    icon: Share2,
    requiredPermissions: ['message:sendways:view']
  },
  {
    title: '数据管理',
    icon: Database,
    requiredPermissions: ['data:mq-source:view'],
    children: [
      {
        title: '消息队列',
        path: '/data/mq-sources',
        name: 'data-mq-sources',
        icon: Database,
        requiredPermissions: ['data:mq-source:view']
      }
    ]
  },
  {
    title: '日志管理',
    icon: ScrollText,
    requiredPermissions: ['message:sendlogs:view'],
    children: [
      {
        title: '任务日志',
        path: '/logs/task',
        icon: ScrollText,
        requiredPermissions: ['message:sendlogs:view']
      },
      {
        title: '登录日志',
        path: '/logs/login',
        icon: ScrollText,
        requiredPermissions: ['system:loginlogs:view']
      },
      {
        title: '消费日志',
        path: '/logs/consume',
        icon: ScrollText,
        requiredPermissions: ['data:consume-log:view']
      }
    ]
  },
  {
    title: '系统管理',
    icon: Shield,
    children: [
      {
        title: '用户管理',
        path: '/system/users',
        name: 'system-users',
        icon: User,
        requiredPermissions: ['system:rbac:user']
      },
      {
        title: '用户组管理',
        path: '/system/groups',
        name: 'system-groups',
        icon: Users,
        requiredPermissions: ['system:rbac:group']
      },
      {
        title: '角色管理',
        path: '/system/roles',
        name: 'system-roles',
        icon: Users,
        requiredPermissions: ['system:rbac:role']
      },
      {
        title: '权限管理',
        path: '/system/permissions',
        name: 'system-permissions',
        icon: KeyRound,
        requiredPermissions: ['system:rbac:permission']
      },
      {
        title: '系统设置',
        path: '/system/settings',
        name: 'system-settings',
        icon: Settings,
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

const toggleUserMenu = () => {
  isUserMenuOpen.value = !isUserMenuOpen.value
}

const openProfileSettings = () => {
  isUserMenuOpen.value = false
  router.push('/profile/settings')
}
</script>

<template>
  <aside
    :class="cn(
      'sidebar fixed left-0 top-0 z-50 h-screen bg-[var(--sidebar-bg,#001529)] flex flex-col text-white transition-[width] duration-300 ease-in-out',
      sidebarClass
    )"
  >
    <div class="flex items-center justify-between h-14 px-4">
      <div v-if="!isCollapsed" class="text-[16px] font-bold text-[#1890ff] truncate">
        {{ siteTitle }}
      </div>
      <div
        v-else
        class="w-6 h-6 bg-[#1890ff] text-white rounded flex items-center justify-center font-bold text-[18px]"
      >
        {{ collapsedBrandText }}
      </div>
      <button
        @click="emit('toggle-collapse')"
        class="text-white/70 hover:text-[#1890ff] transition-colors text-[16px] flex items-center justify-center"
      >
        <ChevronLeft v-if="!isCollapsed" class="w-4 h-4" />
        <ChevronRight v-else class="w-4 h-4" />
      </button>
    </div>
    <div class="h-px bg-[#434343]" />

    <div class="flex-1 overflow-y-auto py-2">
      <nav class="space-y-1 px-2 text-[14px]">
        <template v-for="(item, index) in filteredMenuItems" :key="index">
          <div v-if="!item.children">
            <button
              @click="handleItemClick(item)"
              :class="cn(
                'w-full flex items-center gap-3 px-3 py-2 rounded transition-colors relative',
                isActive(item)
                  ? 'bg-[#1890ff] text-white'
                  : 'text-white/85 hover:bg-[#1f2d3d]',
                isCollapsed ? 'justify-center' : 'justify-start'
              )"
            >
              <span
                v-if="isActive(item)"
                class="absolute left-0 top-0 h-full w-[3px] bg-[#40a9ff]"
              />
              <component :is="item.icon" class="w-5 h-5" />
              <span v-if="!isCollapsed" class="truncate">{{ item.title }}</span>
            </button>
          </div>

          <div v-else>
            <button
              @click="handleItemClick(item)"
              :class="cn(
                'w-full flex items-center gap-3 px-3 py-2 rounded transition-colors relative',
                isActive(item)
                  ? 'bg-[#1f2d3d] text-white'
                  : 'text-white/85 hover:bg-[#1f2d3d]',
                isCollapsed ? 'justify-center' : 'justify-start'
              )"
            >
              <component :is="item.icon" class="w-5 h-5" />
              <span v-if="!isCollapsed" class="flex-1 text-left truncate">{{ item.title }}</span>
              <component
                v-if="!isCollapsed"
                :is="ChevronDown"
                class="w-4 h-4 text-white/70"
              />
            </button>
            <div v-if="isGroupOpen(item)" class="mt-1 space-y-1">
              <button
                v-for="(child, childIndex) in item.children"
                :key="childIndex"
                @click="handleItemClick(child)"
                :class="cn(
                  'w-full flex items-center gap-3 px-3 py-2 rounded transition-colors relative',
                  isActive(child)
                    ? 'bg-[#1890ff] text-white'
                    : 'text-white/75 hover:bg-[#1f2d3d]',
                  isCollapsed ? 'justify-center' : 'justify-start',
                  !isCollapsed ? 'pl-9' : ''
                )"
              >
                <span
                  v-if="isActive(child)"
                  class="absolute left-0 top-0 h-full w-[3px] bg-[#40a9ff]"
                />
              <component :is="child.icon" class="w-5 h-5" />
                <span v-if="!isCollapsed" class="truncate">{{ child.title }}</span>
              </button>
            </div>
          </div>
        </template>
      </nav>
    </div>

    <div class="mt-auto px-3 py-3">
      <div class="h-px bg-[#434343] mb-3" />
      <div
        class="relative flex items-center gap-3 cursor-pointer"
        @click="toggleUserMenu"
      >
        <div
          :class="[
            'rounded-full bg-white/10 flex items-center justify-center',
            isCollapsed ? 'w-6 h-6' : 'w-8 h-8'
          ]"
        >
          <User class="w-4 h-4 text-white" />
        </div>
        <div v-if="!isCollapsed" class="flex-1 text-sm font-medium text-white truncate">
          {{ userAccount }}
        </div>
        <div
          v-if="!isCollapsed"
          class="flex items-center justify-center"
        >
          <ChevronDown class="w-4 h-4 text-white/70" />
        </div>
        <div
          v-if="isUserMenuOpen"
          class="absolute bottom-10 left-0 w-[160px] rounded bg-[#0f1b2d] border border-[#1f2d3d] shadow-lg"
        >
          <button
            @click="openProfileSettings"
            class="w-full text-left px-3 py-2 text-sm text-white/90 hover:bg-[#1f2d3d]"
          >
            个人设置
          </button>
          <button
            @click="emit('logout')"
            class="w-full text-left px-3 py-2 text-sm text-white/90 hover:bg-[#1f2d3d]"
          >
            退出登录
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>
