<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  modelValue?: string
  class?: HTMLAttributes["class"]
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const updateModelValue = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  if (target?.value) {
    emit('update:modelValue', target.value)
  }
}
</script>

<template>
  <div
    data-slot="radio-group"
    role="radiogroup"
    :class="cn('grid gap-3', props.class)"
    @change="updateModelValue"
  >
    <slot />
  </div>
</template>
