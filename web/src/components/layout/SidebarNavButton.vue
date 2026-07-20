<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
import DoraIcon from '@/components/ui/DoraIcon.vue'
import type { NavigationItem } from '@/types/app'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.layout.sidebar

const props = defineProps<{
  item: NavigationItem
  isActive: boolean
  isCollapsed: boolean
  isOpen: boolean
  level?: number
}>()

const emit = defineEmits<{
  (e: 'select', item: NavigationItem): void
}>()

const hasChildren = computed(() => !!props.item.children?.length)
const buttonLabel = computed(() => hasChildren.value ? `${props.item.title}${props.isOpen ? messages.expandedSuffix : messages.collapsedSuffix}` : props.item.title)
const ariaExpanded = computed(() => hasChildren.value && !props.isCollapsed ? props.isOpen : undefined)
const menuIconClass = 'sidebar-nav-icon text-[17px] leading-none relative z-[1] shrink-0'
const indentationClass = computed(() => {
  if (props.isCollapsed || !props.level) return ''
  if (props.level === 1) return 'pl-9'
  if (props.level === 2) return 'pl-[54px]'
  return 'pl-[72px]'
})
const buttonClass = computed(() => cn(
  'sidebar-nav-button group w-full flex items-center gap-3 px-3 py-2 transition-all duration-[var(--motion-fast)] relative overflow-hidden',
  props.isActive ? 'sidebar-nav-active' : 'sidebar-nav-idle',
  props.isCollapsed ? 'justify-center' : 'justify-start',
  !props.isCollapsed && props.level ? 'sidebar-nav-child' : '',
  indentationClass.value
))
</script>

<template>
  <button
    type="button"
    :class="buttonClass"
    :title="isCollapsed ? item.title : undefined"
    :aria-label="buttonLabel"
    :aria-current="isActive && !hasChildren ? 'page' : undefined"
    :aria-expanded="ariaExpanded"
    @click="emit('select', item)"
  >
    <span v-if="isActive" class="sidebar-nav-active-indicator" aria-hidden="true" />
    <span class="sidebar-nav-interaction-layer" aria-hidden="true" />
    <DoraIcon v-if="item.iconName" :name="item.iconName" :class="menuIconClass" :size="17" />
    <component v-else-if="item.icon" :is="item.icon" :class="menuIconClass" aria-hidden="true" />
    <span v-if="!isCollapsed" class="flex-1 text-left truncate relative z-[1]">{{ item.title }}</span>
    <DoraIcon
      v-if="!isCollapsed && hasChildren"
      name="chevron-right"
      :size="13"
      :class="cn(
        'sidebar-nav-arrow w-3.5 h-3.5 transition-transform duration-[var(--motion-fast)] relative z-[1]',
        isOpen ? 'rotate-90' : ''
      )"
    />
  </button>
</template>
