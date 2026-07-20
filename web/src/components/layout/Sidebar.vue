<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { cn } from '@/lib/utils'
import { useRbacStore } from '@/store/rbac'
import { filterMenuItems, findActiveParentTitles, isMenuItemActive, menuItems } from '@/util/navigation'
import SidebarItem from './SidebarItem.vue'
import doraLogoUrl from '@/assets/dora-svg-icon/logo.svg'
import type { NavigationItem } from '@/types/app'

const props = defineProps<{
  isCollapsed: boolean
  isMobile?: boolean
  isMobileVisible?: boolean
  siteTitle: string
  siteSlogan: string
  siteLogo: string
  siteSloganInitialEnabled: boolean
  userAccount: string
}>()

const emit = defineEmits<{
  (e: 'navigate'): void
}>()

const route = useRoute()
const router = useRouter()
const rbacStore = useRbacStore()
const openMenus = ref<Record<string, boolean>>({})
const logoLoadFailed = ref(false)
const sidebarRef = ref<HTMLElement | null>(null)

const filteredMenuItems = computed(() => filterMenuItems(menuItems, permissions => rbacStore.hasAnyPermission(permissions)))

const isActive = (item: NavigationItem) => isMenuItemActive(item, route.path)

const handleItemClick = (item: NavigationItem) => {
  if (item.children?.length && !item.path && !item.name) {
    openMenus.value[item.title] = !openMenus.value[item.title]
    return
  }
  if (item.name) {
    router.push({ name: item.name })
    emit('navigate')
    return
  }
  if (item.path) {
    router.push(item.path)
    emit('navigate')
  }
}

const isGroupOpen = (item: NavigationItem) => !!openMenus.value[item.title]

watch(
  () => route.path,
  (path) => {
    findActiveParentTitles(filteredMenuItems.value, path).forEach((title) => {
      openMenus.value[title] = true
    })
  },
  { immediate: true }
)

const sidebarClass = computed(() => {
  if (props.isMobile) return props.isMobileVisible ? 'w-[var(--dora-mobile-sider-width)] translate-x-0' : 'w-[var(--dora-mobile-sider-width)] -translate-x-full'
  return props.isCollapsed ? 'w-[var(--dora-sider-collapsed-width)]' : 'w-[var(--dora-sider-width)]'
})

const collapsedBrandText = computed(() => {
  if (!props.siteSloganInitialEnabled) return 'M'
  const slogan = (props.siteSlogan || '').trim()
  if (!slogan) return 'M'
  const latinOrDigit = slogan.match(/[A-Za-z0-9]/)
  if (latinOrDigit?.[0]) return latinOrDigit[0].toUpperCase()
  const firstChar = slogan.charAt(0).trim()
  return firstChar ? firstChar.toUpperCase() : 'M'
})

const brandLogo = computed(() => {
  if (logoLoadFailed.value) return doraLogoUrl
  return props.siteLogo || doraLogoUrl
})

watch(
  () => props.siteLogo,
  () => {
    logoLoadFailed.value = false
  }
)

watch(
  () => [props.isMobile, props.isMobileVisible],
  async ([isMobile, visible]) => {
    if (!isMobile || !visible) return
    await nextTick()
    sidebarRef.value?.querySelector<HTMLElement>('button, [href], [tabindex]:not([tabindex="-1"])')?.focus()
  }
)

const handleLogoError = () => {
  if (brandLogo.value !== doraLogoUrl) {
    logoLoadFailed.value = true
  }
}
</script>

<template>
  <aside
    ref="sidebarRef"
    :aria-hidden="isMobile && !isMobileVisible"
    :inert="isMobile && !isMobileVisible"
    :class="cn(
      'sidebar fixed left-0 top-0 z-50 h-screen flex flex-col text-white transition-[width,transform] duration-[var(--motion-normal)] ease-in-out',
      sidebarClass
    )"
  >
    <div class="sidebar-brand flex items-center h-[var(--dora-header-height)] px-4 gap-3">
      <div v-if="!isCollapsed || isMobile" class="sidebar-brand-mark w-8 h-8 text-white rounded-xl flex items-center justify-center font-bold text-[16px] shrink-0">
        <img v-if="brandLogo" :src="brandLogo" alt="logo" class="h-6 w-6 object-contain" @error="handleLogoError" />
        <span v-else>{{ collapsedBrandText }}</span>
      </div>
      <div v-if="!isCollapsed || isMobile" class="min-w-0">
        <div class="sidebar-brand-title text-[16px] font-semibold truncate">
          {{ siteTitle }}
        </div>
        <div v-if="siteSlogan" class="sidebar-brand-slogan text-[11px] truncate mt-0.5">
          {{ siteSlogan }}
        </div>
      </div>
      <div
        v-else
        class="sidebar-brand-mark w-8 h-8 text-white rounded-xl flex items-center justify-center font-bold text-[16px]"
      >
        <img v-if="brandLogo" :src="brandLogo" alt="logo" class="h-6 w-6 object-contain" @error="handleLogoError" />
        <span v-else>{{ collapsedBrandText }}</span>
      </div>
    </div>
    <div class="flex-1 overflow-y-auto py-3">
      <nav class="space-y-1 px-2 text-[14px]">
        <SidebarItem
          v-for="item in filteredMenuItems"
          :key="item.path || item.title"
          :item="item"
          :is-collapsed="isCollapsed && !isMobile"
          :is-active="isActive"
          :is-open="isGroupOpen"
          @select="handleItemClick"
        />
      </nav>
    </div>
  </aside>
</template>
