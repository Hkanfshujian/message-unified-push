<script setup lang="ts">
import { inject, computed } from "vue"
import type { HTMLAttributes } from "vue"
import { CloseOutlined } from "@ant-design/icons-vue"
import { cn } from "@/lib/utils"
import DialogOverlay from "./DialogOverlay.vue"
import { dialogContextKey, type DialogContext } from "./dialogContext"

const props = defineProps<{ class?: HTMLAttributes["class"] }>()
const context = inject(dialogContextKey, null) as DialogContext | null
const isOpen = computed(() => context?.open.value ?? true)
const close = () => context?.setOpen(false)
</script>

<template>
  <Teleport v-if="isOpen" to="body">
    <DialogOverlay @click="close" />
    <div data-slot="dialog-content" :class="cn(
        'glass-dialog-content data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 p-6 duration-[var(--motion-fast)] sm:max-w-lg',
        props.class,
      )" @click.stop>
      <slot />

      <button
        type="button"
        class="glass-dialog-close absolute top-4 right-4 transition-all focus:outline-hidden disabled:pointer-events-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"
        @click="close"
      >
        <CloseOutlined class="text-[14px]" />
        <span class="sr-only">Close</span>
      </button>
    </div>
  </Teleport>
</template>
