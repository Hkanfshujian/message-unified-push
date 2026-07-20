<script setup lang="ts">
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import { ref, computed, watch, nextTick } from 'vue'
import { templatesApi } from '@/api/templates'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'

const messages = zhCN.templateEditor

interface Placeholder {
  key: string
  label: string
  default: string
}

interface TemplateData {
  id?: string  // 模板ID是字符串类型（UUID）
  name: string
  description: string
  text_template: string
  html_template: string
  markdown_template: string
  placeholders: string
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status: string
}

// 组件props
interface Props {
  open?: boolean
  isEditing?: boolean
  templateData?: TemplateData | null
}
const props = withDefaults(defineProps<Props>(), {
  open: false,
  isEditing: false,
  templateData: null
})

// 组件emits
const emit = defineEmits<{
  'update:open': [value: boolean]
  'saved': []
}>()

// Textarea refs for inserting placeholders
const textTemplateRef = ref<any>(null)
const htmlTemplateRef = ref<any>(null)
const markdownTemplateRef = ref<any>(null)

// 表单数据
const formData = ref<TemplateData>({
  name: '',
  description: '',
  text_template: '',
  html_template: '',
  markdown_template: '',
  placeholders: '[]',
  at_mobiles: '',
  at_user_ids: '',
  is_at_all: false,
  status: 'enabled'
})

const nameTooLong = ref(false)
const maxNameUnits = 18

const getNameUnits = (value: string) => {
  let units = 0
  for (const char of value) {
    units += /[^\x00-\xff]/.test(char) ? 1 : 0.5
  }
  return units
}

const truncateName = (value: string) => {
  let units = 0
  let result = ''
  for (const char of value) {
    const nextUnits = units + (/[^\x00-\xff]/.test(char) ? 1 : 0.5)
    if (nextUnits > maxNameUnits) break
    units = nextUnits
    result += char
  }
  return { value: result, units }
}

const handleNameInput = (value: string | number) => {
  const text = String(value || '')
  const { value: trimmed, units } = truncateName(text)
  nameTooLong.value = units >= maxNameUnits && trimmed !== text
  formData.value.name = trimmed
  if (!nameTooLong.value && getNameUnits(trimmed) <= maxNameUnits) {
    nameTooLong.value = false
  }
}

// 使用独立的响应式数组来管理占位符，避免频繁的 JSON 序列化
const placeholdersList = ref<Placeholder[]>([])

// 过滤出有效的占位符（key 不为空）
const validPlaceholders = computed(() => {
  return placeholdersList.value.filter(ph => ph.key && ph.key.trim())
})

// 预览数据
const previewData = ref({
  text: '',
  html: '',
  markdown: '',
  params: {} as Record<string, string>
})

// 是否显示预览
const showPreview = ref(false)

// 预览防抖定时器
let previewDebounceTimer: number | null = null

// 刷新预览
const refreshPreview = async () => {
  if (!props.isEditing || !formData.value.id) {
    // 新建模板时，直接使用当前输入的内容作为预览
    previewData.value.text = replacePreviewPlaceholders(formData.value.text_template)
    previewData.value.html = replacePreviewPlaceholders(formData.value.html_template)
    previewData.value.markdown = replacePreviewPlaceholders(formData.value.markdown_template)
    return
  }

  try {
    const rsp = await templatesApi.preview({
      id: formData.value.id,
      params: previewData.value.params
    })
    previewData.value.text = rsp.data.data.text || ''
    previewData.value.html = rsp.data.data.html || ''
    previewData.value.markdown = rsp.data.data.markdown || ''
  } catch (error: any) {
    console.error('预览失败:', error)
  }
}

// 替换预览占位符（用于新建模板）
const replacePreviewPlaceholders = (template: string) => {
  if (!template) return ''
  let result = template
  Object.keys(previewData.value.params).forEach(key => {
    const value = previewData.value.params[key] || `{{${key}}}`
    result = result.replace(new RegExp(`{{${key}}}`, 'g'), value)
  })
  return result
}

// 监听模板内容变化，自动刷新预览（防抖）
watch([
  () => formData.value.text_template,
  () => formData.value.html_template,
  () => formData.value.markdown_template,
  () => previewData.value.params
], () => {
  if (!showPreview.value) return
  
  if (previewDebounceTimer) {
    clearTimeout(previewDebounceTimer)
  }
  previewDebounceTimer = window.setTimeout(() => {
    refreshPreview()
  }, 500)
}, { deep: true })

// 监听占位符列表变化，同步到 formData（使用防抖）
let placeholderDebounceTimer: number | null = null
watch(placeholdersList, () => {
  if (placeholderDebounceTimer) {
    clearTimeout(placeholderDebounceTimer)
  }
  placeholderDebounceTimer = window.setTimeout(() => {
    formData.value.placeholders = JSON.stringify(placeholdersList.value)
  }, 300)
}, { deep: true })

// 检查占位符 key 是否重复
const isDuplicateKey = (key: string, currentIndex: number): boolean => {
  if (!key.trim()) return false
  return placeholdersList.value.some((p, index) => 
    index !== currentIndex && p.key.trim() === key.trim()
  )
}

// 获取重复的 key 列表
const getDuplicateKeys = computed(() => {
  const keys = placeholdersList.value.map(p => p.key.trim()).filter(k => k)
  const duplicates = new Set<string>()
  const seen = new Set<string>()
  
  keys.forEach(key => {
    if (seen.has(key)) {
      duplicates.add(key)
    }
    seen.add(key)
  })
  
  return duplicates
})

// 添加占位符
const addPlaceholder = () => {
  placeholdersList.value.push({ key: '', label: '', default: '' })
}

// 删除占位符
const removePlaceholder = (index: number) => {
  placeholdersList.value.splice(index, 1)
}

// 插入占位符到模板
const insertPlaceholder = async (type: 'text' | 'html' | 'markdown', key: string) => {
  const placeholder = `{{${key}}}`
  let targetRef: any = null
  
  if (type === 'text') targetRef = textTemplateRef.value
  else if (type === 'html') targetRef = htmlTemplateRef.value
  else if (type === 'markdown') targetRef = markdownTemplateRef.value
  
  if (!targetRef) return
  
  await nextTick()
  
  const textarea = targetRef?.textarea || targetRef?.$el?.querySelector?.('textarea') || targetRef?.$el || targetRef
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const text = formData.value[`${type}_template`]
  
  const before = text.substring(0, start)
  const after = text.substring(end)
  
  formData.value[`${type}_template`] = before + placeholder + after
  
  await nextTick()
  textarea.focus()
  const newPosition = start + placeholder.length
  textarea.setSelectionRange(newPosition, newPosition)
}

// 重置表单
const resetForm = () => {
  formData.value = {
    id: '',
    name: '',
    description: '',
    text_template: '',
    html_template: '',
    markdown_template: '',
    placeholders: '[]',
    at_mobiles: '',
    at_user_ids: '',
    is_at_all: false,
    status: 'enabled'
  }
  placeholdersList.value = []
}

// 加载模板数据
const loadTemplateData = (template: TemplateData) => {
  formData.value = {
    id: template.id,
    name: template.name,
    description: template.description,
    text_template: template.text_template,
    html_template: template.html_template,
    markdown_template: template.markdown_template,
    placeholders: template.placeholders,
    at_mobiles: template.at_mobiles || '',
    at_user_ids: template.at_user_ids || '',
    is_at_all: Boolean(template.is_at_all),
    status: template.status
  }
  
  // 解析占位符
  try {
    placeholdersList.value = JSON.parse(template.placeholders || '[]')
  } catch {
    placeholdersList.value = []
  }
  
  // 初始化预览参数
  previewData.value.params = {}
  placeholdersList.value.forEach(p => {
    previewData.value.params[p.key] = p.default || ''
  })
}

const idError = ref('')

const validateTemplateIdFormat = (id: string) => {
  if (!id) {
    return ''
  }
  const regex = /^TP[a-zA-Z0-9]{10}$/
  if (!regex.test(id)) {
    return '模板ID必须以TP开头，且后面为10位字母或数字（共12位）'
  }
  return ''
}

// 保存模板
const saveTemplate = async () => {
  if (!formData.value.name.trim()) {
    notifyError('请输入模板名称')
    return
  }
  
  // 验证至少填写一种格式的模板内容
  if (!formData.value.text_template && !formData.value.html_template && !formData.value.markdown_template) {
    notifyError('至少需要填写一种格式的模板内容')
    return
  }
  
  // 验证占位符 key 不能为空且不能重复
  const emptyKeys = placeholdersList.value.filter(p => p.key.trim() === '')
  if (emptyKeys.length > 0) {
    notifyError('占位符 key 不能为空')
    return
  }
  
  if (getDuplicateKeys.value.size > 0) {
    const duplicates = Array.from(getDuplicateKeys.value).join('、')
    notifyError(`占位符 key 不能重复：${duplicates}`)
    return
  }

  // 模板ID校验（仅在新建时允许手动指定）
  if (!props.isEditing) {
    idError.value = validateTemplateIdFormat(formData.value.id || '')
    if (idError.value) {
      notifyError(idError.value)
      return
    }
  } else {
    idError.value = ''
  }

  // 同步占位符数据
  formData.value.placeholders = JSON.stringify(placeholdersList.value)

  try {
    const response = await (props.isEditing ? templatesApi.update(formData.value) : templatesApi.create(formData.value))
    if (response.data.code === 200) {
      notifySuccess(props.isEditing ? '更新模板成功' : '添加模板成功')
      emit('update:open', false)
      emit('saved')
    } else {
      notifyError(response.data.msg || '操作失败')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || error.response?.data?.message || '操作失败')
  }
}

// 监听对话框打开状态
watch(() => props.open, (newVal) => {
  if (newVal) {
    if (props.isEditing && props.templateData) {
      loadTemplateData(props.templateData)
    } else {
      resetForm()
    }
  }
})
</script>

<template>
  <AppFormDrawer :model-value="open" :title="isEditing ? messages.editTitle : messages.createTitle" size="min(960px, 96vw)" @update:model-value="(value: boolean) => $emit('update:open', value)">
    <div class="template-editor-content">
      <section class="template-editor-card space-y-4">
        <div class="template-editor-section-head">
          <div>
            <h3 class="app-form-section-title">{{ messages.basicInfo }}</h3>
            <p class="app-form-section-description">{{ messages.basicDescription }}</p>
          </div>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-10 gap-4">
          <div class="md:col-span-7 app-form-field template-name-input">
            <label for="name" class="app-form-label">{{ messages.templateName }}<span class="text-destructive">{{ messages.required }}</span></label>
            <el-input id="name" :model-value="formData.name" maxlength="50" :placeholder="messages.namePlaceholder" @input="handleNameInput" />
            <div v-if="nameTooLong" class="app-form-error">{{ messages.nameTooLong }}</div>
          </div>
          <div class="md:col-span-3 app-form-field">
            <label class="app-form-label">{{ messages.status }}</label>
            <el-select v-model="formData.status" class="w-full">
              <el-option :label="messages.enabled" value="enabled" />
              <el-option :label="messages.disabled" value="disabled" />
            </el-select>
          </div>
        </div>

        <div class="app-form-field">
          <label for="templateId" class="app-form-label">{{ messages.templateId }}<span class="app-form-optional">{{ messages.optional }}</span></label>
          <el-input id="templateId" v-model="formData.id" :readonly="isEditing" :placeholder="messages.idPlaceholder" />
          <div v-if="idError" class="app-form-error">{{ idError }}</div>
        </div>

        <div class="app-form-field">
          <label for="description" class="app-form-label">{{ messages.description }}</label>
          <el-input id="description" v-model="formData.description" type="textarea" :rows="3" :placeholder="messages.descriptionPlaceholder" />
        </div>
      </section>

      <section class="template-editor-card space-y-3">
        <div class="template-editor-section-head">
          <div>
            <h3 class="app-form-section-title">{{ messages.placeholders }}</h3>
            <p class="app-form-section-description">{{ messages.placeholdersDescription }}</p>
          </div>
          <el-button size="small" plain @click="addPlaceholder">{{ messages.addPlaceholder }}</el-button>
        </div>
        <div v-for="(placeholder, index) in placeholdersList" :key="index" class="template-placeholder-row">
          <div class="min-w-0 flex-1">
            <el-input v-model="placeholder.key" :placeholder="messages.keyPlaceholder" :class="isDuplicateKey(placeholder.key, index) ? 'is-error' : ''" />
            <p v-if="isDuplicateKey(placeholder.key, index)" class="app-form-error">{{ messages.duplicateKey }}</p>
          </div>
          <el-input v-model="placeholder.label" :placeholder="messages.labelPlaceholder" class="flex-1" />
          <el-input v-model="placeholder.default" :placeholder="messages.defaultPlaceholder" class="flex-1" />
          <el-button size="small" type="danger" text @click="removePlaceholder(index)">{{ messages.delete }}</el-button>
        </div>
        <p class="template-editor-help">
          {{ messages.placeholderHelpPrefix }}<code class="template-editor-code" v-text="'{{key}}'"></code>{{ messages.placeholderHelpMiddle }}<code class="template-editor-code" v-text="'{{username}}'"></code>
        </p>

        <div class="template-editor-subsection">
          <div class="template-editor-section-head">
            <div>
              <h3 class="app-form-section-title">{{ messages.mentionSettings }}</h3>
              <p class="app-form-section-description">{{ messages.mentionDescription }}</p>
            </div>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
            <div class="template-editor-check-item">
              <el-checkbox v-model="formData.is_at_all">{{ messages.mentionAll }}</el-checkbox>
            </div>
            <el-input v-model="formData.at_mobiles" :placeholder="messages.mobilePlaceholder" />
            <el-input v-model="formData.at_user_ids" :placeholder="messages.userIdPlaceholder" />
          </div>
        </div>
      </section>

      <section class="template-editor-card space-y-3">
        <div class="template-editor-section-head">
          <div>
            <h3 class="app-form-section-title">{{ messages.content }}</h3>
            <p class="app-form-section-description">{{ messages.contentDescription }}</p>
          </div>
          <el-button size="small" plain @click="showPreview = !showPreview; if (showPreview) refreshPreview()">
            {{ showPreview ? messages.hidePreview : messages.showPreview }}
          </el-button>
        </div>
        <div v-if="showPreview && validPlaceholders.length > 0" class="template-preview-params space-y-3">
          <label class="app-form-label">{{ messages.fillPlaceholderParameters }}</label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <div v-for="ph in validPlaceholders" :key="ph.key" class="flex gap-2 items-center">
              <label class="template-preview-key">{{ ph.key }}</label>
              <el-input v-model="previewData.params[ph.key]" :placeholder="ph.default || `${messages.parameterPlaceholderPrefix}${ph.key}`" size="small" />
            </div>
          </div>
        </div>
      
      <el-tabs model-value="text" class="template-editor-tabs w-full">
        <el-tab-pane label="Text" name="text">
          <div class="template-editor-code-panel space-y-3">
            <div class="app-form-field">
              <label class="app-form-label">{{ messages.textTemplate }}</label>
              <div v-if="validPlaceholders.length > 0" class="flex gap-1 overflow-x-auto pb-1 scrollbar-thin">
                <el-button v-for="ph in validPlaceholders" :key="ph.key" size="small" @click="insertPlaceholder('text', ph.key)">{{ ph.key }}</el-button>
              </div>
            </div>
            <el-input ref="textTemplateRef" v-model="formData.text_template" type="textarea" :placeholder="messages.textPlaceholder" :rows="showPreview ? 10 : 15" />
            <div v-if="showPreview" class="app-form-field">
              <label class="app-form-label">{{ messages.preview }}</label>
              <div class="template-preview-panel">
                <pre class="whitespace-pre-wrap text-sm">{{ previewData.text || messages.noContent }}</pre>
              </div>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="HTML" name="html">
          <div class="template-editor-card space-y-3">
            <div class="app-form-field">
              <label class="app-form-label">{{ messages.htmlTemplate }}</label>
              <div v-if="validPlaceholders.length > 0" class="flex gap-1 overflow-x-auto pb-1 scrollbar-thin">
                <el-button v-for="ph in validPlaceholders" :key="ph.key" size="small" @click="insertPlaceholder('html', ph.key)">{{ ph.key }}</el-button>
              </div>
            </div>
            <el-input ref="htmlTemplateRef" v-model="formData.html_template" type="textarea" :placeholder="messages.htmlPlaceholder" :rows="showPreview ? 10 : 15" />
            <div v-if="showPreview" class="app-form-field">
              <label class="app-form-label">{{ messages.htmlPreview }}</label>
              <div class="template-preview-panel">
                <div v-html="previewData.html || messages.noContent"></div>
              </div>
              <p class="template-editor-help">{{ messages.htmlHelp }}</p>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="Markdown" name="markdown">
          <div class="template-editor-code-panel space-y-3">
            <div class="app-form-field">
              <label class="app-form-label">{{ messages.markdownTemplate }}</label>
              <div v-if="validPlaceholders.length > 0" class="flex gap-1 overflow-x-auto pb-1 scrollbar-thin">
                <el-button v-for="ph in validPlaceholders" :key="ph.key" size="small" @click="insertPlaceholder('markdown', ph.key)">{{ ph.key }}</el-button>
              </div>
            </div>
            <el-input ref="markdownTemplateRef" v-model="formData.markdown_template" type="textarea" :placeholder="messages.markdownPlaceholder" :rows="showPreview ? 10 : 15" />
            <div v-if="showPreview" class="app-form-field">
              <label class="app-form-label">{{ messages.markdownPreview }}</label>
              <div class="template-preview-panel">
                <pre class="whitespace-pre-wrap text-sm">{{ previewData.markdown || messages.noContent }}</pre>
              </div>
              <p class="template-editor-help">{{ messages.markdownHelp }}</p>
            </div>
          </div>
        </el-tab-pane>
        </el-tabs>
      </section>
    </div>
    <template #footer>
      <span class="template-editor-footer-note">{{ messages.footerHelp }}</span>
      <el-button @click="$emit('update:open', false)">{{ messages.cancel }}</el-button>
      <el-button type="primary" @click="saveTemplate">{{ messages.save }}</el-button>
    </template>
  </AppFormDrawer>
</template>

<style scoped>
.template-editor-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.template-editor-footer-note { margin-right: auto; color: var(--admin-text-muted); font-size: 11px; }

.template-editor-card {
  border: 1px solid var(--glass-inset-border);
  border-radius: var(--admin-radius-lg);
  background: var(--glass-panel-bg);
  box-shadow: var(--glass-shadow-inset);
  padding: 16px;
}

.template-editor-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.template-placeholder-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr) auto;
  align-items: flex-start;
  gap: 8px;
}

.template-editor-check-item,
.template-preview-params,
.template-preview-panel {
  border: 1px solid var(--glass-inset-border);
  border-radius: var(--admin-radius-lg);
  background: var(--glass-inset-bg);
  box-shadow: var(--glass-shadow-inset);
}

.template-editor-check-item {
  display: flex;
  min-height: 32px;
  align-items: center;
  padding: 0 12px;
}

.template-preview-params,
.template-preview-panel {
  padding: 12px;
}

.template-preview-key {
  width: 96px;
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--admin-text-muted);
}

.template-editor-help {
  font-size: 12px;
  line-height: 1.6;
  color: var(--admin-text-muted);
}

.template-editor-code {
  border-radius: 6px;
  background: var(--glass-active-bg);
  padding: 1px 5px;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.app-form-optional {
  margin-left: 4px;
  font-size: 12px;
  font-weight: 400;
  color: var(--admin-text-muted);
}

.template-editor-tabs :deep(.el-tabs__header) {
  margin-bottom: 12px;
}

.template-editor-tabs :deep(.el-tabs__nav-wrap::after) {
  background: var(--glass-inset-border);
}

.template-editor-tabs :deep(.el-tabs__item) {
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 768px) {
  .template-editor-card {
    padding: 14px;
  }

  .template-editor-section-head,
  .template-placeholder-row {
    grid-template-columns: 1fr;
  }

  .template-editor-section-head {
    flex-direction: column;
  }
}
</style>
