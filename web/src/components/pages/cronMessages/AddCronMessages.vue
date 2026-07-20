<script setup lang="ts">
import { ref, reactive } from 'vue'
import { scheduledMessagesApi } from '@/api/scheduledMessages'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
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
  if (!formData.name.trim()) {
    notifyError('请输入定时消息名称')
    return
  }
  if (!formData.template_id) {
    notifyError('请选择关联的消息模板')
    return
  }
  if (!formData.cron_expression.trim()) {
    notifyError('请输入 Cron 表达式')
    return
  }
  loading.value = true
  try {
    let postData = {
      "name": formData.name,
      "title": formData.name,
      "cron": formData.cron_expression,
      "template_id": formData.template_id,
      "url": ""
    }

    const rsp = await scheduledMessagesApi.create(postData)
    if (rsp.data.code === 200) {
      notifySuccess(rsp.data.msg || '创建定时消息成功')
      emit('save', postData)
      emit('update:open', false)
    } else {
      notifyError(rsp.data.msg || '创建定时消息失败')
    }
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || '创建定时消息失败')
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
    notifyError('请先选择关联的消息模板')
    return
  }

  loading.value = true
  try {
    const postData = {
      template_id: formData.template_id,
      name: formData.name,
      title: formData.name
    }

    const rsp = await scheduledMessagesApi.sendNow(postData)
    if (rsp.data.code === 200) {
      notifySuccess(rsp.data.msg || '发送成功')
    } else {
      notifyError(rsp.data.msg || '发送失败')
    }
  } catch (error) {
    notifyError('发送失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <CronMessageForm
    :model-value="formData"
    @update:model-value="(val) => {
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
