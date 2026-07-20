<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { mqApi } from '@/api/mq'
import { notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.mqSourceForm

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
    notifyWarning('请输入数据源名称')
    return
  }
  if (!formData.namesrv_addr) {
    notifyWarning('请输入队列地址')
    return
  }

  isSubmitting.value = true
  try {
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

    const res = await (isEdit.value ? mqApi.update(props.data?.id || '', payload) : mqApi.create(payload))
    if (res.data.code === 200) {
      notifySuccess(isEdit.value ? '编辑成功' : '新增成功')
      emit('success')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '操作失败')
  } finally {
    isSubmitting.value = false
  }
}

// 测试连接
const handleTestConnection = async () => {
  if (!formData.namesrv_addr) {
    notifyWarning('请输入队列地址')
    return
  }

  isTesting.value = true
  testResult.value = null
  
  try {
    const res = await mqApi.testConfig({
      type: formData.type,
      namesrv_addr: formData.namesrv_addr,
      access_key: formData.enableAuth ? formData.access_key : '',
      secret_key: formData.enableAuth ? formData.secret_key : ''
    })
    
    if (res.data.code === 200) {
      testResult.value = res.data.data
      if (testResult.value?.success) {
        notifySuccess('连接测试成功')
      } else {
        notifyError(testResult.value?.error || '连接测试失败')
      }
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.msg || '连接测试失败'
    testResult.value = { success: false, error: errorMsg }
    notifyError(errorMsg)
  } finally {
    isTesting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="mq-source-form">
    <div class="mq-source-content">
    <section class="mq-source-section">
      <header><div><span>{{ messages.sectionOne }}</span><h4>{{ messages.configuration }}</h4></div><p>{{ messages.configurationDescription }}</p></header>
      <div class="mq-source-body mq-source-config">
      <div class="app-form-grid">
    <div class="app-form-field">
      <label for="name" class="app-form-label">{{ messages.name }}<span class="text-destructive">{{ messages.required }}</span></label>
      <el-input
        id="name"
        v-model="formData.name"
        :placeholder="messages.namePlaceholder"
        maxlength="200"
        clearable
      />
    </div>

    <div class="app-form-field">
      <label for="type" class="app-form-label">{{ messages.type }}<span class="text-destructive">{{ messages.required }}</span></label>
      <el-select v-model="formData.type" class="w-full" :placeholder="messages.typePlaceholder">
        <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </el-select>
    </div>
    </div>
    <div class="mq-source-subsection">
      <h5>{{ messages.connectionAndAuth }}</h5>
    <div class="app-form-field">
      <label for="namesrv_addr" class="app-form-label">{{ messages.address }}<span class="text-destructive">{{ messages.required }}</span></label>
      <el-input
        id="namesrv_addr"
        v-model="formData.namesrv_addr"
        :placeholder="messages.addressPlaceholder"
        maxlength="500"
        clearable
      />
      <p class="app-form-help">
        {{ messages.addressHelp }}
      </p>
    </div>

    <div class="app-form-field app-form-inline-control">
      <div class="flex items-center justify-between">
        <label for="enableAuth" class="app-form-label">{{ messages.enableAuth }}</label>
        <el-switch id="enableAuth" v-model="formData.enableAuth" />
      </div>
      <p class="app-form-help">
        {{ messages.authHelp }}
      </p>
    </div>

    <div v-if="formData.enableAuth" class="app-form-grid">
      <div class="app-form-field">
        <label for="access_key" class="app-form-label">Access Key</label>
        <el-input
          id="access_key"
          v-model="formData.access_key"
          :placeholder="messages.accessKeyPlaceholder"
          maxlength="200"
          clearable
        />
      </div>

      <div class="app-form-field">
        <label for="secret_key" class="app-form-label">Secret Key</label>
        <el-input
          id="secret_key"
          v-model="formData.secret_key"
          type="password"
          :placeholder="messages.secretKeyPlaceholder"
          maxlength="200"
          show-password
        />
      </div>
    </div>

    <div v-if="isEdit" class="app-form-field">
      <label for="enabled" class="app-form-label">{{ messages.enabledStatus }}</label>
      <el-select v-model="formData.enabled" class="w-full" :placeholder="messages.statusPlaceholder">
        <el-option :label="messages.enabled" :value="1" />
        <el-option :label="messages.disabled" :value="0" />
      </el-select>
    </div>

      </div>
      </div>
    </section>

    <section class="mq-source-section">
      <header><div><span>{{ messages.sectionTwo }}</span><h4>{{ messages.testFeedback }}</h4></div><p>{{ messages.testDescription }}</p></header>
      <div class="mq-source-feedback">
        <el-alert v-if="testResult" :type="testResult.success ? 'success' : 'error'" :title="testResult.success ? (testResult.message || messages.connectionSucceeded) : (testResult.error || messages.connectionFailed)" show-icon :closable="false" />
        <p v-else>{{ messages.notTested }}</p>
      </div>
    </section>
    </div>

    <div class="mq-source-actions">
      <span>{{ messages.submitHelp }}</span>
      <div>
        <el-button type="button" :disabled="isTesting || !formData.namesrv_addr" @click="handleTestConnection">{{ isTesting ? messages.testing : messages.testConnection }}</el-button>
        <el-button type="primary" native-type="submit" :loading="isSubmitting">{{ isSubmitting ? messages.submitting : (isEdit ? messages.save : messages.create) }}</el-button>
      </div>
    </div>
  </form>
</template>

<style scoped>
.mq-source-form { display: flex; flex-direction: column; width: 100%; height: 100%; min-height: 0; }
.mq-source-content { display: grid; grid-auto-rows: max-content; align-content: start; flex: 1; gap: 12px; min-height: 0; padding: 16px; overflow-y: auto; overscroll-behavior: contain; }
.mq-source-section { overflow: visible; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.mq-source-section header > div { display: flex; align-items: center; gap: 8px; }
.mq-source-section h4 { margin: 0; }
.mq-source-section header p, .mq-source-feedback p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.mq-source-section header { display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 38px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.mq-source-section header span { color: var(--brand-600); font: 800 10px monospace; }
.mq-source-section h4 { font-size: 13px; }
.mq-source-section header p { margin: 0; }
.mq-source-body, .mq-source-feedback { padding: 14px; }
.mq-source-config { display: grid; gap: 16px; }
.mq-source-subsection { display: grid; gap: 14px; padding-top: 14px; border-top: 1px solid color-mix(in srgb, var(--app-overlay-border) 72%, transparent); }
.mq-source-subsection h5 { margin: 0; color: var(--admin-text-primary); font-size: 12px; font-weight: 700; }
.mq-source-actions { display: flex; flex: none; align-items: center; justify-content: space-between; gap: 16px; min-height: 58px; padding: 10px 16px; border-top: 1px solid var(--app-overlay-border); background: var(--app-overlay-surface); }
.mq-source-actions > span { min-width: 0; color: var(--admin-text-muted); font-size: 11px; }
.mq-source-actions > div { display: flex; flex: none; gap: 8px; }
@container app-managed-drawer (max-width: 680px) { .mq-source-content { padding: 12px; } .mq-source-section header { align-items: center; flex-direction: row; } .mq-source-section header p { display: none; } .mq-source-actions > span { max-width: 260px; } }
@container app-managed-drawer (max-width: 460px) { .mq-source-actions { align-items: stretch; flex-direction: column; } .mq-source-actions > span { display: none; } .mq-source-actions > div { width: 100%; } .mq-source-actions :deep(.el-button) { flex: 1; } }
</style>
