<script setup lang="ts">
import AppFormDrawer from './AppFormDrawer.vue'

const visible = defineModel<boolean>({ required: true })

const props = withDefaults(defineProps<{
  title: string
  width?: string
  confirmText?: string
  cancelText?: string
  loading?: boolean
  showFooter?: boolean
  closeOnClickModal?: boolean
}>(), {
  width: '720px',
  confirmText: '保存',
  cancelText: '取消',
  loading: false,
  showFooter: true,
  closeOnClickModal: false
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

</script>

<template>
  <AppFormDrawer
    v-model="visible"
    :size="width"
    :title="title"
    :confirm-text="confirmText"
    :cancel-text="cancelText"
    :loading="loading"
    :show-footer="showFooter"
    :close-on-click-modal="props.closeOnClickModal"
    @confirm="emit('confirm')"
    @cancel="emit('cancel')"
  >
    <slot />
    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </AppFormDrawer>
</template>
