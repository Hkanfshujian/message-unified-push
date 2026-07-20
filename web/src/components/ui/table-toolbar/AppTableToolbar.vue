<script setup lang="ts">
import { computed } from 'vue'
import { FullScreen, Refresh, Setting } from '@element-plus/icons-vue'
import { getTableToolbarMaterialClasses } from './tableToolbar'
import type { TableToolbarColumn } from './types'

const props = withDefaults(defineProps<{
  title?: string
  columns: TableToolbarColumn[]
  visibleColumns: string[]
  focused?: boolean
  refreshing?: boolean
}>(), {
  title: '列表数据',
  focused: false,
  refreshing: false
})

const emit = defineEmits<{
  (e: 'refresh'): void
  (e: 'update:focused', value: boolean): void
  (e: 'update:visibleColumns', value: string[]): void
}>()

const visibleSet = computed(() => new Set(props.visibleColumns))
const toolbarClass = computed(() => getTableToolbarMaterialClasses({ focused: props.focused, refreshing: props.refreshing }))

const toggleColumn = (column: TableToolbarColumn) => {
  if (column.required) return
  const next = new Set(props.visibleColumns)
  if (next.has(column.key)) next.delete(column.key)
  else next.add(column.key)
  emit('update:visibleColumns', Array.from(next))
}
</script>

<template>
  <div class="app-table-toolbar" :class="toolbarClass">
    <div class="app-table-toolbar-title">
      <span>{{ title }}</span>
      <slot name="summary" />
    </div>
    <div class="app-table-toolbar-actions">
      <el-button :icon="Refresh" :loading="refreshing" @click="emit('refresh')">刷新</el-button>
      <el-button :icon="FullScreen" @click="emit('update:focused', !focused)">{{ focused ? '退出聚焦' : '聚焦' }}</el-button>
      <el-popover placement="bottom-end" trigger="click" width="220">
        <template #reference>
          <el-button :icon="Setting">列设置</el-button>
        </template>
        <div class="space-y-2 app-table-toolbar-columns">
          <el-checkbox
            v-for="column in columns"
            :key="column.key"
            :model-value="visibleSet.has(column.key)"
            :disabled="column.required"
            @update:model-value="() => toggleColumn(column)"
          >
            {{ column.label }}
          </el-checkbox>
        </div>
      </el-popover>
    </div>
  </div>
</template>
