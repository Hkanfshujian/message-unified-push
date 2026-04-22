<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import CronMessageForm from './CronMessageForm.vue'

interface CronMessageItem {
  id: string
  name: string
  cron: string
  template_id: string
  enable: number
  status: boolean
}

interface Props {
  open: boolean
  cronMessage: CronMessageItem | null
}

interface Emits {
  (e: 'save', data: any): void
  (e: 'cancel'): void
  (e: 'update:open', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

defineOptions({
  name: 'EditCronMessages'
})

// 表单数据
const formData = reactive({
  name: '',
  cron_expression: '',
  template_id: ''
})


// 加载状态
const loading = ref(false)

// 提交表单
const handleSubmit = async () => {
  if (!props.cronMessage) {
    toast.error('未找到要编辑的定时消息')
    return
  }
  
  
  loading.value = true
  try {
    let postData = {
      "name": formData.name,
      "id": props.cronMessage.id,
      "cron": formData.cron_expression,
      "template_id": formData.template_id,
      "title": formData.name,
      "url": "",
      "enable": props.cronMessage.enable,
    }

    const rsp = await request.post('/cronmessages/edit', postData)
    if (rsp.data.code === 200) {
      toast.success(rsp.data.msg)
      setTimeout(() => {
        window.location.reload()
      }, 1000)
    } else {
        toast.success(rsp.data.msg)
      }
  } finally {
    loading.value = false
  }
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
  emit('update:open', false)
}

// 立即发送
const handleSendNow = async () => {
  // 验证必填字段
  if (!formData.template_id) {
    toast.error('请先选择关联的消息模板')
    return
  }

  loading.value = true
  try {
    const postData = {
      id: props.cronMessage?.id,
      template_id: formData.template_id,
      name: formData.name,
      title: formData.name
    }

    const rsp = await request.post('/cronmessages/sendnow', postData)
    if (rsp.data.code === 200) {
      toast.success(rsp.data.msg)
    } else {
      toast.error(rsp.data.msg)
    }
  } catch (error) {
    toast.error('发送失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 监听 cronMessage 变化，更新表单数据
watch(
  () => props.cronMessage,
  (newCronMessage) => {
    if (newCronMessage) {
      formData.name = newCronMessage.name
      formData.cron_expression = newCronMessage.cron
      formData.template_id = newCronMessage.template_id
    }
  },
  { immediate: true }
)
</script>

<template>
  <CronMessageForm
    :model-value="formData"
    @update:model-value="(val) => Object.assign(formData, val)"
    mode="edit"
    :loading="loading"
    @submit="handleSubmit"
    @cancel="handleCancel"
    @send-now="handleSendNow"
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EditCronMessages'
})
</script>
