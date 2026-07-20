<script setup lang="ts">
import { inject, computed } from "vue"
import type { HTMLAttributes } from "vue"
import { CloseOutlined } from "@ant-design/icons-vue"
import { cn } from "@/lib/utils"
import { dialogContextKey, type DialogContext } from "./dialogContext"

const props = defineProps<{ class?: HTMLAttributes["class"] }>()
const context = inject(dialogContextKey, null) as DialogContext | null
const isOpen = computed(() => context?.open.value ?? true)
const close = () => context?.setOpen(false)
</script>

<template>
  <Teleport v-if="isOpen" to="body">
    <div
      class="fixed inset-0 z-50 grid place-items-center overflow-y-auto bg-black/55 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
      @click="close"
    >
      <div
        :class="
          cn(
            'glass-dialog-content relative z-50 grid w-full max-w-lg my-8 gap-4 p-6 duration-[var(--motion-fast)] md:w-full',
            props.class,
          )
        "
        @click.stop
      >
        <slot />

        <button
          type="button"
          class="glass-dialog-close absolute top-4 right-4 transition-all"
          @click="close"
        >
          <CloseOutlined class="text-[14px]" />
          <span class="sr-only">Close</span>
        </button>
      </div>
    </div>
  </Teleport>
</template>
