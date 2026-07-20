<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'info' | ''
  text?: boolean
  link?: boolean
  loading?: boolean
  disabled?: boolean
  confirm?: boolean
  confirmMessage?: string
}>(), {
  type: '',
  text: true,
  link: false,
  loading: false,
  disabled: false,
  confirm: false,
  confirmMessage: '确认执行该操作？'
})

const emit = defineEmits<{
  click: []
}>()

const handleClick = () => emit('click')
const buttonClass = computed(() => [
  'app-action-button',
  props.type === 'primary' ? 'app-action-button-primary' : 'app-action-button-secondary',
  props.type === 'danger' ? 'app-action-button-danger dora-state-danger' : '',
  props.loading ? 'dora-state-loading' : '',
  props.disabled ? 'dora-state-disabled' : '',
  props.confirm ? 'app-action-button-confirm' : ''
])
</script>

<template>
  <el-popconfirm v-if="confirm" :title="confirmMessage" confirm-button-text="确认" cancel-button-text="取消" @confirm="handleClick">
    <template #reference>
      <el-button :type="type" :text="text" :link="link" :loading="loading" :disabled="disabled" :class="buttonClass">
        <slot />
      </el-button>
    </template>
  </el-popconfirm>
  <el-button v-else :type="type" :text="text" :link="link" :loading="loading" :disabled="disabled" :class="buttonClass" @click="handleClick">
    <slot />
  </el-button>
</template>
