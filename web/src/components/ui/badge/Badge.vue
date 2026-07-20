<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import type { BadgeVariants } from "."
import { cn } from "@/lib/utils"
import { badgeVariants } from "."

const props = defineProps<{
  as?: string
  variant?: BadgeVariants["variant"]
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <component
    :is="as || 'span'"
    data-slot="badge"
    :class="cn(badgeVariants({ variant }), props.class)"
  >
    <slot />
  </component>
</template>
