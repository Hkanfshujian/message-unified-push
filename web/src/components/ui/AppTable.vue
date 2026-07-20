<script setup lang="ts" generic="T extends Record<string, unknown>">
import { computed } from 'vue'
import AppEmptyState from './AppEmptyState.vue'

export interface AppTableColumn {
  prop?: string
  label: string
  width?: string | number
  minWidth?: string | number
  align?: 'left' | 'center' | 'right'
  fixed?: boolean | 'left' | 'right'
  formatter?: (row: Record<string, unknown>) => unknown
}

const props = withDefaults(defineProps<{
  data: T[]
  columns: AppTableColumn[]
  loading?: boolean
  rowKey?: string
  emptyText?: string
  border?: boolean
  stripe?: boolean
  size?: 'large' | 'default' | 'small'
  selection?: boolean
  selectionWidth?: string | number
  error?: string
  stale?: boolean
  retryText?: string
}>(), {
  loading: false,
  rowKey: 'id',
  emptyText: '当前没有可展示的数据',
  border: false,
  stripe: true,
  size: 'default',
  selection: false,
  selectionWidth: 46,
  error: '',
  stale: false,
  retryText: '重试'
})

const emit = defineEmits<{
  (e: 'selection-change', rows: T[]): void
  (e: 'retry'): void
}>()

const tableClass = computed(() => [
  'app-data-table',
  'dora-material-base',
  props.stripe ? 'app-data-table-striped' : '',
  props.border ? 'app-data-table-bordered' : ''
])
</script>

<template>
  <div v-if="error && data.length > 0" class="px-4 py-2 text-sm text-[var(--el-color-warning)]" role="status">
    {{ stale ? `${error}，当前展示上次成功加载的数据。` : error }}
    <el-button type="primary" link @click="emit('retry')">{{ retryText }}</el-button>
  </div>
  <el-table
    v-loading="loading"
    :data="data"
    :row-key="rowKey"
    :border="border"
    :stripe="stripe"
    :size="size"
    class="w-full"
    :class="tableClass"
    @selection-change="emit('selection-change', $event as T[])"
  >
    <el-table-column v-if="selection" type="selection" :width="selectionWidth" />
    <el-table-column
      v-for="column in columns"
      :key="`${column.prop || column.label}`"
      :prop="column.prop"
      :label="column.label"
      :width="column.width"
      :min-width="column.minWidth"
      :align="column.align"
      :fixed="column.fixed"
      :class-name="column.prop === 'actions' ? 'app-table-actions-column' : ''"
      :label-class-name="column.prop === 'actions' ? 'app-table-actions-column' : ''"
    >
      <template #default="scope">
        <slot :name="column.prop || column.label" :row="scope.row" :column="column" :index="scope.$index">
          {{ column.formatter ? column.formatter(scope.row) : (column.prop ? scope.row[column.prop] ?? '-' : '-') }}
        </slot>
      </template>
    </el-table-column>
    <template #empty>
      <AppEmptyState v-if="error && !loading" title="数据加载失败" :description="error">
        <template #extra>
          <el-button type="primary" @click="emit('retry')">{{ retryText }}</el-button>
        </template>
      </AppEmptyState>
      <slot v-else name="empty">
        <AppEmptyState :description="emptyText" />
      </slot>
    </template>
  </el-table>
</template>
