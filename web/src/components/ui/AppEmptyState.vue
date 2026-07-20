<script setup lang="ts">
import emptyDataUrl from '@/assets/dora-svg-icon/empty-data.svg'

withDefaults(defineProps<{
  title?: string
  description?: string
  imageSize?: number
}>(), {
  description: '当前没有可展示的数据',
  imageSize: 96
})
</script>

<template>
  <el-empty class="app-empty-state" :image="emptyDataUrl" :description="description" :image-size="imageSize">
    <template v-if="title" #default>
      <div class="app-empty-state-title text-sm font-medium text-foreground">{{ title }}</div>
    </template>
    <template v-if="$slots.extra" #extra>
      <slot name="extra" />
    </template>
  </el-empty>
</template>
