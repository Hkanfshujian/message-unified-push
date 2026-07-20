<script setup lang="ts">
import SidebarNavButton from './SidebarNavButton.vue'
import type { NavigationItem } from '@/types/app'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.layout.sidebar

defineProps<{
  item: NavigationItem
  isCollapsed: boolean
  level?: number
  isActive: (item: NavigationItem) => boolean
  isOpen: (item: NavigationItem) => boolean
}>()

const emit = defineEmits<{
  (e: 'select', item: NavigationItem): void
}>()
</script>

<template>
  <div>
    <SidebarNavButton
      :item="item"
      :is-active="isActive(item)"
      :is-collapsed="isCollapsed"
      :is-open="isOpen(item)"
      :level="level"
      @select="emit('select', $event)"
    />
    <div v-if="item.children?.length && isOpen(item) && !isCollapsed" class="sidebar-submenu mt-1 space-y-1" role="group" :aria-label="`${item.title} ${messages.submenuSuffix}`">
      <SidebarItem
        v-for="child in item.children"
        :key="child.path || child.title"
        :item="child"
        :is-collapsed="isCollapsed"
        :level="(level || 0) + 1"
        :is-active="isActive"
        :is-open="isOpen"
        @select="emit('select', $event)"
      />
    </div>
  </div>
</template>
