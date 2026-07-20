<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { templatesApi } from '@/api/templates'
import { notifyError } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.cronMessageForm

interface CronMessageFormData {
  id?: string
  name: string
  cron_expression: string
  template_id: string
}

interface Props {
  modelValue: CronMessageFormData
  mode: 'add' | 'edit'
  loading?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: CronMessageFormData): void
  (e: 'submit'): void
  (e: 'cancel'): void
  (e: 'sendNow'): void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})
const emit = defineEmits<Emits>()

defineOptions({
  name: 'CronMessageForm'
})

// 本地表单数据
const localFormData = reactive<CronMessageFormData>({ ...props.modelValue })

// 监听外部数据变化
watch(() => props.modelValue, (newValue) => {
  Object.assign(localFormData, newValue)
}, { deep: true })

// 监听本地数据变化，同步到外部
watch(localFormData, (newValue) => {
  emit('update:modelValue', newValue)
}, { deep: true })

// 可用的模板列表
const availableTemplates = ref<Array<{ id: string, name: string }>>([])

// 常用的 Cron 表达式模板
const cronTemplates = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每5分钟', value: '*/5 * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天凌晨2点', value: '0 2 * * *' },
  { label: '每周一凌晨2点', value: '0 2 * * 1' },
  { label: '每月1号凌晨2点', value: '0 2 1 * *' }
]

// 加载可用模板
const loadAvailableTemplates = async () => {
  try {
    const rsp = await templatesApi.list({ page: 1, size: 100, status: 'enabled' })
    availableTemplates.value = (rsp.data.data.lists || []).map((tpl: any) => ({
      id: tpl.id,
      name: tpl.name
    }))
  } catch (error) {
    notifyError('加载可用模板失败')
  }
}

// 应用 Cron 模板
const applyCronTemplate = (template: string) => {
  localFormData.cron_expression = template
}

// 提交表单
const handleSubmit = () => {
  emit('submit')
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
}

// 立即发送
const handleSendNow = () => {
  emit('sendNow')
}

// 组件挂载时加载数据
loadAvailableTemplates()
</script>

<template>
  <div class="cron-form">
    <div class="cron-form-content">
      <section class="cron-form-section">
        <header><div><span>{{ messages.sectionNumber }}</span><h4>{{ messages.title }}</h4></div><p>{{ messages.description }}</p></header>
        <div class="cron-form-section-body cron-form-fields">
          <div class="app-form-field">
            <label for="name" class="app-form-label">{{ messages.name }}</label>
            <el-input id="name" v-model="localFormData.name" :placeholder="messages.namePlaceholder" clearable />
          </div>
          <div class="app-form-field">
            <label for="template_id" class="app-form-label">{{ messages.template }}</label>
            <el-select id="template_id" v-model="localFormData.template_id" filterable class="w-full" :placeholder="messages.templatePlaceholder">
              <el-option v-for="template in availableTemplates" :key="template.id" :label="template.name" :value="String(template.id)" />
            </el-select>
          </div>
          <div class="app-form-field cron-form-schedule-field">
            <label for="cron_expression" class="app-form-label">{{ messages.cronExpression }}</label>
            <el-input id="cron_expression" v-model="localFormData.cron_expression" :placeholder="messages.cronPlaceholder" clearable />
            <div class="cron-form-presets">
              <span>{{ messages.commonSchedules }}</span>
              <div><el-button v-for="template in cronTemplates" :key="template.value" size="small" @click="applyCronTemplate(template.value)">{{ template.label }}</el-button></div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <footer class="cron-form-actions">
      <el-button size="small" :disabled="loading" @click="handleSendNow">{{ messages.sendNowOnce }}</el-button>
      <div>
        <el-button size="small" :disabled="loading" @click="handleCancel">{{ messages.cancel }}</el-button>
        <el-button type="primary" size="small" :loading="loading" @click="handleSubmit">{{ mode === 'add' ? messages.create : messages.update }}</el-button>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.cron-form { display: flex; flex-direction: column; width: 100%; height: 100%; min-height: 0; }
.cron-form-content { display: grid; grid-auto-rows: max-content; align-content: start; flex: 1; gap: 12px; min-height: 0; padding: 16px; overflow-y: auto; overscroll-behavior: contain; }
.cron-form-section { overflow: visible; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.cron-form-section header > div { display: flex; align-items: center; gap: 8px; }
.cron-form-section h4 { margin: 0; }
.cron-form-section header p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.cron-form-section header { display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 38px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.cron-form-section header span { color: var(--brand-600); font: 800 10px monospace; }
.cron-form-section h4 { font-size: 13px; }
.cron-form-section header p { margin: 0; }
.cron-form-section-body { padding: 14px; }
.cron-form-fields { display: grid; gap: 14px; }
.cron-form-fields > .app-form-field + .app-form-field { padding-top: 14px; border-top: 1px solid color-mix(in srgb, var(--app-overlay-border) 72%, transparent); }
.cron-form-presets { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: 10px; }
.cron-form-presets > span { color: var(--admin-text-muted); font-size: 11px; }
.cron-form-presets > div, .cron-form-actions > div { display: flex; flex-wrap: wrap; gap: 8px; }
.cron-form-actions { display: flex; flex: none; align-items: center; justify-content: space-between; gap: 12px; min-height: 58px; padding: 10px 16px; border-top: 1px solid var(--app-overlay-border); background: var(--app-overlay-surface); }
@container app-managed-drawer (max-width: 680px) {
  .cron-form-content { padding: 12px; }
  .cron-form-presets { align-items: flex-start; flex-direction: column; }
  .cron-form-section header { align-items: center; flex-direction: row; }
  .cron-form-section header p { display: none; }
}
@container app-managed-drawer (max-width: 460px) {
  .cron-form-actions { align-items: stretch; flex-direction: column; }
  .cron-form-actions > div { width: 100%; }
  .cron-form-actions :deep(.el-button) { flex: 1; margin-left: 0; }
}
</style>

