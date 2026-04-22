<script setup lang="ts">
import type { PaginationListItemProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { reactiveOmit } from "@vueuse/core"
import { PaginationListItem } from "reka-ui"
import { cn } from "@/lib/utils"

const props = defineProps<PaginationListItemProps & {
  class?: HTMLAttributes["class"]
  isActive?: boolean
}>()

const delegatedProps = reactiveOmit(props, "class", "isActive")
</script>

<template>
  <PaginationListItem
    data-slot="pagination-item"
    v-bind="delegatedProps"
    :class="cn('inline-flex items-center justify-center', props.class)"
  >
    <slot />
  </PaginationListItem>
</template>
