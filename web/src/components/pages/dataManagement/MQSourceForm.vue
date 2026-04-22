<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'

interface Props {
  data?: {
    id: string
    name: string
    type: string
    namesrv_addr: string
    access_key: string
    secret_key: string
    enabled: number
  } | null
}

const props = withDefaults(defineProps<Props>(), {
  data: null
})

const emit = defineEmits<{
  success: []
}>()

const isEdit = computed(() => !!props.data)

const formData = reactive({
  name: props.data?.name || '',
  type: props.data?.type || 'rocketmq',
  namesrv_addr: props.data?.namesrv_addr || '',
  access_key: props.data?.access_key || '',
  secret_key: props.data?.secret_key || '',
  enabled: props.data?.enabled ?? 1,
  enableAuth: !!(props.data?.access_key && props.data.access_key.length > 0)
})

// 监听认证开关，关闭时清空 AK/SK
watch(() => formData.enableAuth, (val) => {
  if (!val) {
    formData.access_key = ''
    formData.secret_key = ''
  }
})

const isSubmitting = ref(false)
const isTesting = ref(false)
const testResult = ref<{ success: boolean; message?: string; error?: string } | null>(null)

const typeOptions = [
  { value: 'rocketmq', label: 'RocketMQ' },
  { value: 'kafka', label: 'Kafka' },
  { value: 'rabbitmq', label: 'RabbitMQ' }
]

const handleSubmit = async () => {
  if (!formData.name) {
    toast.warning('请输入数据源名称')
    return
  }
  if (!formData.namesrv_addr) {
    toast.warning('请输入队列地址')
    return
  }

  isSubmitting.value = true
  try {
    const url = isEdit.value
      ? `/mq-sources/${props.data?.id}/edit`
      : '/mq-sources/add'
    
    const method = isEdit ? 'post' : 'post'
    const payload: any = {
      name: formData.name,
      type: formData.type,
      namesrv_addr: formData.namesrv_addr,
      access_key: formData.enableAuth ? formData.access_key : '',
      secret_key: formData.enableAuth ? formData.secret_key : ''
    }
    
    if (isEdit.value) {
      payload.enabled = formData.enabled
    }

    const res = await request[method](url, payload)
    if (res.data.code === 200) {
      toast.success(isEdit.value ? '编辑成功' : '新增成功')
      emit('success')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '操作失败')
  } finally {
    isSubmitting.value = false
  }
}

// 测试连接
const handleTestConnection = async () => {
  if (!formData.namesrv_addr) {
    toast.warning('请输入队列地址')
    return
  }

  isTesting.value = true
  testResult.value = null
  
  try {
    const res = await request.post('/mq-sources/test-config', {
      type: formData.type,
      namesrv_addr: formData.namesrv_addr,
      access_key: formData.enableAuth ? formData.access_key : '',
      secret_key: formData.enableAuth ? formData.secret_key : ''
    })
    
    if (res.data.code === 200) {
      testResult.value = res.data.data
      if (testResult.value?.success) {
        toast.success('连接测试成功')
      } else {
        toast.error(testResult.value?.error || '连接测试失败')
      }
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.msg || '连接测试失败'
    testResult.value = { success: false, error: errorMsg }
    toast.error(errorMsg)
  } finally {
    isTesting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="space-y-4">
    <div class="space-y-2">
      <Label for="name">数据源名称 <span class="text-destructive">*</span></Label>
      <Input
        id="name"
        v-model="formData.name"
        placeholder="例如：生产环境 RocketMQ"
        maxlength="200"
      />
    </div>

    <div class="space-y-2">
      <Label for="type">队列类型 <span class="text-destructive">*</span></Label>
      <Select v-model="formData.type">
        <SelectTrigger>
          <SelectValue placeholder="选择队列类型" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>

    <div class="space-y-2">
      <Label for="namesrv_addr">队列地址 <span class="text-destructive">*</span></Label>
      <Input
        id="namesrv_addr"
        v-model="formData.namesrv_addr"
        placeholder="例如：127.0.0.1:9876 或 http://mq.example.com:9876"
        maxlength="500"
      />
      <p class="text-xs text-muted-foreground">
        RocketMQ NameServer 地址，多个地址用分号分隔
      </p>
    </div>

    <div class="space-y-2">
      <div class="flex items-center justify-between">
        <Label for="enableAuth">开启认证</Label>
        <Switch id="enableAuth" v-model="formData.enableAuth" />
      </div>
      <p class="text-xs text-muted-foreground">
        如果 RocketMQ 开启了 ACL 鉴权，请启用此项并填写 AK/SK
      </p>
    </div>

    <template v-if="formData.enableAuth">
      <div class="space-y-2">
        <Label for="access_key">Access Key</Label>
        <Input
          id="access_key"
          v-model="formData.access_key"
          placeholder="请输入 Access Key"
          maxlength="200"
        />
      </div>

      <div class="space-y-2">
        <Label for="secret_key">Secret Key</Label>
        <Input
          id="secret_key"
          v-model="formData.secret_key"
          type="password"
          placeholder="请输入 Secret Key"
          maxlength="200"
        />
      </div>
    </template>

    <div v-if="isEdit" class="space-y-2">
      <Label for="enabled">启用状态</Label>
      <Select v-model="formData.enabled">
        <SelectTrigger>
          <SelectValue placeholder="选择状态" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem :value="1">启用</SelectItem>
          <SelectItem :value="0">禁用</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="flex justify-end gap-2 pt-4">
      <Button
        type="button"
        variant="outline"
        :disabled="isTesting || !formData.namesrv_addr"
        @click="handleTestConnection"
      >
        {{ isTesting ? '测试中...' : '测试连接' }}
      </Button>
      <Button type="submit" :disabled="isSubmitting">
        {{ isSubmitting ? '提交中...' : (isEdit ? '保存' : '创建') }}
      </Button>
    </div>

    <!-- 测试结果 -->
    <div
      v-if="testResult"
      class="p-3 rounded-md text-sm"
      :class="testResult.success ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'"
    >
      <div class="flex items-center gap-2">
        <span v-if="testResult.success">✓</span>
        <span v-else>✗</span>
        <span>{{ testResult.success ? '连接成功' : testResult.error }}</span>
      </div>
    </div>
  </form>
</template>
