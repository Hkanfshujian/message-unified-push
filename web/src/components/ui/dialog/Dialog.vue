<script setup lang="ts">
import { computed, provide } from 'vue'
import { dialogContextKey } from './dialogContext'

const props = defineProps<{
  open?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

provide(dialogContextKey, {
  open: computed(() => props.open !== false),
  setOpen: (value: boolean) => emit('update:open', value)
})
</script>

<template>
  <div data-slot="dialog">
    <slot />
  </div>
</template>
