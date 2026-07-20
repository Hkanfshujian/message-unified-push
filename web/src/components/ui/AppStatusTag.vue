<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  status?: string | number | boolean | null
  successValues?: Array<string | number | boolean>
  warningValues?: Array<string | number | boolean>
  dangerValues?: Array<string | number | boolean>
  labelMap?: Record<string, string>
}>(), {
  status: '',
  successValues: () => [1, true, '1', 'success', 'enabled', 'running', 'online'],
  warningValues: () => ['warning', 'pending', 'untested', 'stopped'],
  dangerValues: () => [0, false, '0', 'failed', 'disabled', 'offline', 'error'],
  labelMap: () => ({
    enabled: '启用',
    disabled: '禁用',
    running: '运行中',
    stopped: '已停止',
    success: '成功',
    failed: '失败',
    pending: '处理中',
    untested: '未测试',
    online: '在线',
    offline: '离线',
    true: '启用',
    false: '停用',
    1: '启用',
    0: '停用'
  })
})

const normalized = computed(() => String(props.status ?? ''))

const type = computed(() => {
  const value = props.status ?? ''
  if (props.successValues.includes(value)) return 'success'
  if (props.warningValues.includes(value)) return 'warning'
  if (props.dangerValues.includes(value)) return 'danger'
  return 'info'
})

const label = computed(() => props.labelMap[normalized.value] || normalized.value || '-')
const effect = computed(() => (type.value === 'info' ? 'plain' : 'light'))
const tagClass = computed(() => ['app-status-tag', `app-status-tag-${type.value}`, `app-status-tag-value-${normalized.value || 'empty'}`])
</script>

<template>
  <el-tag :class="tagClass" :type="type" size="small" :effect="effect">{{ label }}</el-tag>
</template>
