<script setup lang="ts">
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import InstanceConfig from '@/components/ui/InstanceConfig.vue'

// 组件props
interface Props {
  open?: boolean
  templateData?: any // 模板数据
}
withDefaults(defineProps<Props>(), {
  open: false,
  templateData: null
})

// 组件emits
defineEmits<{
  'update:open': [value: boolean]
}>()
</script>

<template>
  <Dialog :open="open" @update:open="(value) => $emit('update:open', value)">
    <DialogContent class="w-[min(855px,98vw)] !max-w-[98vw] sm:!max-w-[98vw] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader class="flex-shrink-0 border-b border-border/60 pb-3">
        <DialogTitle>配置发送实例</DialogTitle>
      </DialogHeader>

      <div class="flex-1 overflow-y-auto mt-4 pr-1 sm:pr-2">
        <InstanceConfig 
          :data="templateData" 
          :in-dialog="true"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>
