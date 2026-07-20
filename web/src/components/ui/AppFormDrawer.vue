<script setup lang="ts">
import { computed } from 'vue'

const visible = defineModel<boolean>({ required: true })

const props = withDefaults(defineProps<{
  title: string
  size?: string | number
  width?: string | number
  confirmText?: string
  cancelText?: string
  loading?: boolean
  showFooter?: boolean
  closeOnClickModal?: boolean
  destroyOnClose?: boolean
  bodyMode?: 'scroll' | 'managed'
  density?: 'default' | 'compact'
}>(), {
  size: '640px',
  width: undefined,
  confirmText: '保存',
  cancelText: '取消',
  loading: false,
  showFooter: true,
  closeOnClickModal: false,
  destroyOnClose: true,
  bodyMode: 'scroll',
  density: 'default'
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const drawerSize = computed(() => props.width || props.size)

const handleCancel = () => {
  visible.value = false
  emit('cancel')
}
</script>

<template>
  <el-drawer
    v-model="visible"
    :size="drawerSize"
    direction="rtl"
    :close-on-click-modal="closeOnClickModal"
    :destroy-on-close="destroyOnClose"
    class="app-form-drawer dora-material-overlay"
    :class="[
      { 'dora-state-loading': loading },
      `app-drawer-body-${bodyMode}`,
      `app-drawer-density-${density}`
    ]"
    append-to-body
  >
    <template #header>
      <div class="app-form-drawer-header">
        <div class="min-w-0">
          <h2 class="app-form-drawer-title">{{ title }}</h2>
        </div>
      </div>
    </template>

    <div
      class="app-form-drawer-body dora-material-inset"
      :class="bodyMode === 'managed' ? 'app-managed-drawer-body' : ''"
    >
      <slot />
    </div>

    <template v-if="showFooter || $slots.footer" #footer>
      <div class="app-form-drawer-footer dora-material-panel">
        <slot name="footer">
          <el-button :disabled="loading" @click="handleCancel">{{ cancelText }}</el-button>
          <el-button type="primary" :loading="loading" @click="emit('confirm')">{{ confirmText }}</el-button>
        </slot>
      </div>
    </template>
  </el-drawer>
</template>
