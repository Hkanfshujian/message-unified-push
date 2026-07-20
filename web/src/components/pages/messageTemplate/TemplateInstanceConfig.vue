<script setup lang="ts">
import InstanceConfig from '@/components/ui/InstanceConfig.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.messageTemplate

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
  <AppFormDrawer :model-value="open" :title="messages.configureSendInstance" size="min(900px, 96vw)" :show-footer="false" @update:model-value="(value: boolean) => $emit('update:open', value)">
    <InstanceConfig :data="templateData" :in-dialog="true" />
  </AppFormDrawer>
</template>
