<script setup lang="ts">
import { computed } from 'vue'
import type { DoraIconName } from '@/types/app'

const props = withDefaults(defineProps<{
  name: DoraIconName
  size?: number | string
}>(), {
  size: 16
})

const sizeValue = computed(() => typeof props.size === 'number' ? `${props.size}px` : props.size)

const paths: Record<DoraIconName, string> = {
  activity: 'M5 12h3l2-7 4 14 2-7h3',
  app: 'M4 4h6v6H4V4Zm10 0h6v6h-6V4ZM4 14h6v6H4v-6Zm10 0h6v6h-6v-6Z',
  calendar: 'M7 3v3m10-3v3M4 9h16M5 5h14a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1Z',
  chart: 'M4 19V5m0 14h16M8 16v-5m4 5V8m4 8v-7',
  'chevron-left': 'M15 6l-6 6 6 6',
  'chevron-right': 'M9 6l6 6-6 6',
  close: 'M7 7l10 10M17 7 7 17',
  'close-slash': 'M8 8l8 8M16 8l-8 8',
  collapse: 'M15 6l-6 6 6 6M4 5h4m-4 7h4m-4 7h4',
  copy: 'M8 8h11v11H8V8Zm-3 8H4V4h12v1',
  database: 'M5 7c0 2 14 2 14 0S5 5 5 7Zm0 0v5c0 2 14 2 14 0V7M5 12v5c0 2 14 2 14 0v-5',
  document: 'M7 3h7l4 4v14H7V3Zm7 0v5h5M10 13h6m-6 4h6',
  expand: 'M9 6H5v4m10-4h4v4M9 18H5v-4m10 4h4v-4',
  fullscreen: 'M8 4H4v4m12-4h4v4M8 20H4v-4m12 4h4v-4',
  'fullscreen-exit': 'M9 5v4H5m10-4v4h4M9 19v-4H5m10 4v-4h4',
  'menu-fold-left': 'M5 6h14M5 12h14M5 18h14M10 9l-3 3 3 3',
  'menu-fold-right': 'M5 6h14M5 12h14M5 18h14M7 9l3 3-3 3',
  key: 'M14 14a5 5 0 1 0-4-4l-6 6v3h3v-2h2v-2h2l3-3Z',
  login: 'M10 17l5-5-5-5m5 5H3m13-8h3a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-3',
  logout: 'M14 17l5-5-5-5m5 5H8m2 8H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h5',
  menu: 'M4 6h16M4 12h16M4 18h16',
  message: 'M4 5h16v11H7l-3 3V5Z',
  notification: 'M6 8a6 6 0 0 1 12 0c0 7 3 7 3 9H3c0-2 3-2 3-9Zm5 12h2',
  refresh: 'M20 7v5h-5M4 17v-5h5M18 12a6 6 0 0 0-10-4l-4 4m2 0a6 6 0 0 0 10 4l4-4',
  search: 'M10 18a8 8 0 1 1 0-16 8 8 0 0 1 0 16Zm6-2 5 5',
  security: 'M12 3 5 6v6c0 4 3 7 7 9 4-2 7-5 7-9V6l-7-3Z',
  setting: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Zm8-3a8 8 0 0 0-.1-1l2-1.5-2-3.5-2.4 1a8 8 0 0 0-1.8-1L15.4 3h-4l-.3 3a8 8 0 0 0-1.8 1L7 6 5 9.5 7 11a8 8 0 0 0 0 2l-2 1.5L7 18l2.3-1a8 8 0 0 0 1.8 1l.3 3h4l.3-3a8 8 0 0 0 1.8-1l2.4 1 2-3.5-2-1.5c.1-.3.1-.7.1-1Z',
  sun: 'M12 4V2m0 20v-2m8-8h2M2 12h2m13.7-5.7 1.4-1.4M4.9 19.1l1.4-1.4m0-11.4L4.9 4.9m14.2 14.2-1.4-1.4M12 16a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z',
  moon: 'M20 15.5A8.5 8.5 0 0 1 8.5 4 7 7 0 1 0 20 15.5Z',
  team: 'M8 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm8-1a3 3 0 1 0 0-6M2 21a6 6 0 0 1 12 0m3 0a5 5 0 0 0-4-5',
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 9a7 7 0 0 1 14 0',
  'user-circle': 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20Zm0-10a3 3 0 1 0 0-6 3 3 0 0 0 0 6Zm-6 7a6 6 0 0 1 12 0'
}
</script>

<template>
  <svg
    class="dora-icon"
    :width="sizeValue"
    :height="sizeValue"
    viewBox="0 0 24 24"
    fill="none"
    aria-hidden="true"
  >
    <path :class="{ 'dora-icon-activity-track': name === 'activity' }" :d="paths[name]" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
    <path v-if="name === 'activity'" class="dora-icon-activity-pulse" :d="paths[name]" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
</template>
