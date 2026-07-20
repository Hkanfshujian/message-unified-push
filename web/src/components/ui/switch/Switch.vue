<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  modelValue?: boolean
  defaultValue?: boolean
  disabled?: boolean
  id?: string
  class?: HTMLAttributes["class"]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()
</script>

<template>
  <button
    data-slot="switch"
    type="button"
    role="switch"
    :id="id"
    :aria-checked="modelValue || false"
    :disabled="disabled"
    :data-state="modelValue ? 'checked' : 'unchecked'"
    :class="cn(
      'peer inline-flex h-5 w-9 shrink-0 items-center rounded-full border transition-all outline-none focus-visible:ring-4 focus-visible:ring-brand-200/35 disabled:cursor-not-allowed disabled:opacity-50',
      modelValue ? 'border-white/40 bg-[linear-gradient(135deg,var(--brand-600),#0ea5e9)] shadow-[0_10px_20px_rgba(37,99,235,0.22),inset_0_1px_0_rgba(255,255,255,0.28)]' : 'border-[var(--glass-inset-border)] bg-[var(--glass-inset-bg)] shadow-[var(--glass-shadow-inset)]',
      props.class,
    )"
    @click="emit('update:modelValue', !modelValue)"
  >
    <span
      data-slot="switch-thumb"
      :data-state="modelValue ? 'checked' : 'unchecked'"
      :class="cn('pointer-events-none block size-4 rounded-full ring-0 transition-transform shadow-[0_4px_10px_rgba(15,23,42,0.18)]', modelValue ? 'translate-x-[calc(100%+1px)] bg-white' : 'translate-x-0.5 bg-[var(--glass-panel-bg)]')"
    >
      <slot name="thumb" />
    </span>
  </button>
</template>
