<script setup lang="ts">
import { computed, useAttrs } from "vue"
import type { HTMLAttributes } from "vue"
import type { ButtonVariants } from "."
import { cn } from "@/lib/utils"
import { buttonVariants } from "."

interface Props {
  as?: string
  asChild?: boolean
  variant?: ButtonVariants["variant"]
  size?: ButtonVariants["size"]
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  ariaLabel?: string
  class?: HTMLAttributes["class"]
}

const props = withDefaults(defineProps<Props>(), {
  as: "button",
  type: "button",
  variant: 'default',
  size: 'default',
  loading: false,
})

const attrs = useAttrs()
const isDisabled = computed(() => Boolean(props.disabled || props.loading))
const renderAs = computed(() => props.asChild ? 'span' : props.as)
</script>

<template>
  <component
    :is="renderAs"
    v-bind="attrs"
    data-slot="button"
    :data-variant="variant || 'default'"
    :data-size="size || 'default'"
    :data-loading="loading ? 'true' : undefined"
    :type="renderAs === 'button' ? type : undefined"
    :disabled="renderAs === 'button' ? isDisabled : undefined"
    :aria-disabled="renderAs !== 'button' && isDisabled ? 'true' : undefined"
    :aria-busy="loading ? 'true' : undefined"
    :aria-label="ariaLabel"
    :tabindex="renderAs !== 'button' && isDisabled ? -1 : undefined"
    :class="cn(buttonVariants({ variant, size }), isDisabled && 'pointer-events-none opacity-60', loading && 'button-loading', props.class)"
  >
    <span v-if="loading" class="button-spinner" aria-hidden="true" />
    <slot />
  </component>
</template>
