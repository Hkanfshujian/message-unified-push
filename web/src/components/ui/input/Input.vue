<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { useVModel } from "@vueuse/core"
import { cn } from "@/lib/utils"

const props = defineProps<{
  defaultValue?: string | number
  modelValue?: string | number
  maxLength?: number
  class?: HTMLAttributes["class"]
}>()

const emits = defineEmits<{
  (e: "update:modelValue", payload: string | number): void
}>()

const modelValue = useVModel(props, "modelValue", emits, {
  passive: true,
  defaultValue: props.defaultValue,
})
</script>

<template>
  <input
    v-model="modelValue"
    data-slot="input"
    :maxlength="props.maxLength"
    :class="cn(
      'file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground border-[var(--border-default)] box-border flex h-9 w-full min-w-0 rounded-[var(--radius-control)] border bg-[var(--surface-card)] px-3 py-1 text-base shadow-none transition-[border-color,box-shadow,background-color] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
      'focus-visible:border-[var(--brand)] focus-visible:ring-2 focus-visible:ring-[color-mix(in_srgb,var(--brand)_22%,transparent)]',
      'aria-invalid:ring-red-200/45 aria-invalid:border-red-400',
      props.class,
    )"
  >
</template>
