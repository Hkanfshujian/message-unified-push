<script setup lang="ts">
import { ref, reactive } from 'vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import CronMessageForm from './CronMessageForm.vue'

interface Props {
  open: boolean
}

interface Emits {
  (e: 'save', data: any): void
  (e: 'cancel'): void
  (e: 'update:open', value: boolean): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

defineOptions({
  name: 'AddCronMessages'
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
  loading.value = true
  try {
    let postData = {
      "name": formData.name,
      "title": formData.name,
      "cron": formData.cron_expression,
      "template_id": formData.template_id,
      "url": ""
    }

    const rsp = await request.post('/cronmessages/addone', postData)
    if (rsp.data.code === 200) {
      toast.success(rsp.data.msg)
      setTimeout(() => {
        window.location.reload()
      }, 1000)
    } else {
      toast.error(rsp.data.msg)
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

// 立即发送（新增模式也支持，可以在创建前测试发送效果）
const handleSendNow = async () => {
  // 验证必填字段
  if (!formData.template_id) {
    toast.error('请先选择关联的消息模板')
    return
  }

  loading.value = true
  try {
    const postData = {
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
</script>

<template>
  <CronMessageForm
    :model-value="formData"
    @update:model-value="(val) => {
      console.log('Received update:model-value:', val);
      Object.assign(formData, val);
    }"
    mode="add"
    :loading="loading"
    @submit="handleSubmit"
    @cancel="handleCancel"
    @send-now="handleSendNow"
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'AddCronMessages'
})
</script>
