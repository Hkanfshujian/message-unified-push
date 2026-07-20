<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  value: string
  id?: string
  disabled?: boolean
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <input
    data-slot="radio-group-item"
    type="radio"
    :id="id"
    :value="value"
    :disabled="disabled"
    :class="
      cn(
        'border-[var(--glass-inset-border)] text-brand-600 focus-visible:border-[var(--brand-400)] focus-visible:ring-brand-200/40 aria-invalid:ring-red-200/45 aria-invalid:border-red-400 bg-[var(--glass-inset-bg)] aspect-square size-4 shrink-0 rounded-full border shadow-[var(--glass-shadow-inset)] transition-[color,box-shadow,border-color] outline-none focus-visible:ring-4 disabled:cursor-not-allowed disabled:opacity-50',
        props.class,
      )
    "
  >
</template>
