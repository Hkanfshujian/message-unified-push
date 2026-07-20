<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { CONSTANT } from '@/constant'
import { createValidationState, validateForm, type InputConfig } from '@/util/validation'
import { channelsApi } from '@/api/channels'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
import { zhCN } from '@/locales/zh-CN'
import {
  MailOutlined,
  ThunderboltOutlined,
  BankOutlined,
  RocketOutlined,
  ApiOutlined,
  MobileOutlined,
  BellOutlined,
  QrcodeOutlined,
  SendOutlined,
  NotificationOutlined,
  SafetyCertificateOutlined,
  AppstoreOutlined
} from '@ant-design/icons-vue'

const messages = zhCN.waysForm

// 组件props
interface Props {
  open?: boolean
  editData?: any // 编辑时传入的数据
  mode?: 'add' | 'edit' // 模式：新增或编辑
}
const props = withDefaults(defineProps<Props>(), {
  open: false,
  editData: null,
  mode: 'add'
})

// 组件emits
const emit = defineEmits<{
  'update:open': [value: boolean]
  'save': [data: any]
}>()

// 前端的页面添加配置
let waysConfigMap = CONSTANT.WAYS_DATA;

// Radio Group 选项 - 根据waysConfigMap动态生成
const channelModeOptions = waysConfigMap.map(item => ({
  value: item.type,
  label: item.label
}))
const channelMode = ref(channelModeOptions[0]?.value || '')

// 当前选中渠道的配置
const currentChannelConfig = computed(() => {
  return waysConfigMap.find(item => item.type === channelMode.value) || null
})

const channelNameInput = computed(() => currentChannelConfig.value?.inputs?.find((input: any) => input.col === 'name'))
const channelAuthInputs = computed(() => currentChannelConfig.value?.inputs?.filter((input: any) => input.col !== 'name') || [])
const sensitiveInputPattern = /(passw|passwd|password|secret|token|access[_-]?key|push[_-]?key|bot[_-]?token|corp[_-]?secret|appsecret|\bkey\b|\biv\b)/i
const isSensitiveInput = (input: any) => sensitiveInputPattern.test(`${input.col || ''} ${input.label || ''} ${input.subLabel || ''}`)

// 表单数据
const formData = ref<Record<string, any>>({})
const testRecipientDialogOpen = ref(false)
const testAllConfirmDialogOpen = ref(false)
const testRecipientInput = ref('')
const testRecipientOverride = ref('')
const isAllRecipient = computed(() => {
  const v = testRecipientInput.value.trim()
  return v.toLowerCase() === 'all' || v === '@all'
})

// 校验状态管理
const validationState = createValidationState()

// 初始化表单数据
const initFormData = () => {
  const config = currentChannelConfig.value
  if (!config) return

  const newFormData: Record<string, any> = {}

  // 如果是编辑模式且有编辑数据，先填充编辑数据
  if (props.mode === 'edit' && props.editData) {
    // 设置渠道类型
    channelMode.value = props.editData.type || channelModeOptions[0]?.value || ''

    // 解析auth数据
    let authData: Record<string, any> = {}
    try {
      authData = props.editData.auth ? JSON.parse(props.editData.auth) : {}
    } catch (e) {
      console.error('解析auth数据失败:', e)
    }

    // 填充基本字段
    newFormData.name = props.editData.name || ''

    // 填充auth中的字段
    Object.keys(authData).forEach(key => {
      newFormData[key] = authData[key]
    })
  }

  // 初始化基本输入字段
  if (config.inputs) {
    config.inputs.forEach((input: any) => {
      if (newFormData[input.col] === undefined) {
        newFormData[input.col] = input.value || ''
      }
    })
  }

  // 初始化任务指令输入字段
  if (config.taskInsInputs) {
    config.taskInsInputs.forEach((input: any) => {
      if (newFormData[input.col] === undefined) {
        newFormData[input.col] = input.value || ''
      }
    })
  }

  // 初始化任务指令单选项
  if (config.taskInsRadios && config.taskInsRadios.length > 0) {
    if (newFormData.taskInsRadio === undefined) {
      newFormData.taskInsRadio = config.taskInsRadios[0].value
    }
  }

  formData.value = newFormData
}

const getAllInputConfigs = (): InputConfig[] => {
  const config = currentChannelConfig.value
  if (!config) return []

  const configs: InputConfig[] = []

  if (config.inputs) {
    configs.push(...config.inputs.map((input: any) => ({
      col: input.col,
      label: input.label,
      subLabel: input.subLabel,
      type: input.type,
      required: input.required !== false,
      minLength: input.minLength,
      maxLength: input.maxLength
    })))
  }

  if (config.taskInsInputs) {
    configs.push(...config.taskInsInputs.map((input: any) => ({
      col: input.col,
      label: input.label,
      subLabel: input.subLabel,
      type: input.type,
      required: input.required !== false,
      minLength: input.minLength,
      maxLength: input.maxLength
    })))
  }

  return configs
}

const validateFormData = () => {
  const inputConfigs = getAllInputConfigs()
  const result = validateForm(formData.value, inputConfigs)
  validationState.setErrors(result.errors)
  return result.isValid
}

// 监听渠道模式变化
const handleChannelModeChange = () => {
  initFormData()
  validationState.clearAllErrors()
}

const focusedChannelIndex = ref(0)

watch(channelMode, (value) => {
  const idx = channelModeOptions.findIndex(option => option.value === value)
  if (idx >= 0) {
    focusedChannelIndex.value = idx
  }
})

const selectChannelMode = (value: string, index: number) => {
  channelMode.value = value
  focusedChannelIndex.value = index
  handleChannelModeChange()
}

const moveChannelFocus = (delta: number) => {
  if (!channelModeOptions.length) return
  const len = channelModeOptions.length
  let next = focusedChannelIndex.value + delta
  if (next < 0) next = len - 1
  if (next >= len) next = 0
  const option = channelModeOptions[next]
  if (option) {
    selectChannelMode(option.value, next)
  }
}

const handleChannelListKeydown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    moveChannelFocus(-1)
  } else if (event.key === 'ArrowDown') {
    event.preventDefault()
    moveChannelFocus(1)
  }
}

// 监听编辑数据变化（仅编辑模式）
watch(() => props.editData, () => {
  if (props.mode === 'edit') {
    initFormData()
  }
}, { immediate: true })

// 初始化表单数据（新增模式）
if (props.mode === 'add') {
  initFormData()
}

// 关闭drawer
const handleClose = () => {
  emit('update:open', false)
}

// 获取最终提交数据
const getFinalData = () => {
  // 根据当前渠道配置的inputs中的col字段，从formData中提取对应的值组成auth对象
  const config = currentChannelConfig.value
  const authData: Record<string, any> = {}
  if (config && config.inputs) {
    config.inputs.forEach((input: any) => {
      if (formData.value[input.col] !== undefined && input.col != 'name') {
        authData[input.col] = formData.value[input.col]
        if (config.type == 'Email' && input.col == 'port') {
          authData[input.col] = parseInt(formData.value[input.col])
        }
        if (config.type == 'Gotify' && input.col == 'priority') {
          authData[input.col] = parseInt(formData.value[input.col])
        }
      }
    })
  }

  let postData: Record<string, any> = {
    auth: JSON.stringify(authData),
    type: channelMode.value,
    name: formData.value.name,
  }

  // 编辑时需要传递ID
  if (props.mode === 'edit' && props.editData && props.editData.id) {
    postData.id = props.editData.id
  }

  return postData
}

// 测试连接
const handleTest = async () => {
  if (!validateFormData()) return
  if (channelMode.value === 'QyWeiXinApp') {
    testRecipientInput.value = ''
    testRecipientDialogOpen.value = true
    return
  }
  await doTestRequest()
}

const doTestRequest = async () => {
  try {
    const postData = getFinalData()
    if (channelMode.value === 'QyWeiXinApp') {
      const toUser = testRecipientOverride.value.trim()
      if (!toUser) {
        notifyError('测试接收者不能为空')
        return
      }
      let authData: Record<string, any> = {}
      try {
        authData = postData.auth ? JSON.parse(postData.auth) : {}
      } catch {
        authData = {}
      }
      authData.to_user = toUser
      postData.auth = JSON.stringify(authData)
    }
    const rsp = await channelsApi.test(postData)
    if (rsp?.data?.code == 200) {
      notifySuccess(rsp.data.msg || '测试成功')
      return
    }
    notifyError(rsp?.data?.msg || '测试失败')
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || error?.message || '测试请求失败')
  }
}

const confirmTestRecipient = async () => {
  const recipient = testRecipientInput.value.trim()
  if (!recipient) {
    notifyError('请输入接收者企微ID')
    return
  }
  if (recipient.toLowerCase() === 'all' || recipient === '@all') {
    testRecipientDialogOpen.value = false
    testAllConfirmDialogOpen.value = true
    return
  }
  await submitQyWeiXinAppTest(recipient)
}

const cancelAllRecipientConfirm = () => {
  testAllConfirmDialogOpen.value = false
  testRecipientDialogOpen.value = true
}

const submitQyWeiXinAppTest = async (recipient: string) => {
  testRecipientDialogOpen.value = false
  testAllConfirmDialogOpen.value = false
  testRecipientOverride.value = recipient
  await doTestRequest()
  testRecipientOverride.value = ''
}

// 保存数据
const handleSave = async () => {
  if (!validateFormData()) return

  try {
    const postData = getFinalData()

    // 根据模式选择API路径和成功消息
    const successMessage = props.mode === 'edit' ? '更新渠道成功！' : '添加渠道成功！'

    const rsp = await (props.mode === 'edit' ? channelsApi.update(postData) : channelsApi.create(postData))
    if (rsp?.data?.code == 200) {
      notifySuccess(successMessage)
      emit('save', postData)
      emit('update:open', false)
      return
    }
    notifyError(rsp?.data?.msg || '保存失败')
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || error?.message || '保存请求失败')
  }
}

// 渠道图标映射
const getChannelIcon = (type: string) => {
  const map: Record<string, any> = {
    'Email': MailOutlined,
    'Dtalk': ThunderboltOutlined,
    'QyWeiXin': BankOutlined,
    'Feishu': RocketOutlined,
    'Custom': ApiOutlined,
    'WeChatOFAccount': QrcodeOutlined,
    'AliyunSMS': MobileOutlined,
    'Telegram': SendOutlined,
    'Bark': BellOutlined,
    'Ntfy': NotificationOutlined,
    'Gotify': SafetyCertificateOutlined,
    'QyWeiXinApp': AppstoreOutlined
  }
  return map[type] || SendOutlined // Default icon
}

// 计算保存按钮文本
const saveButtonText = computed(() => {
  return props.mode === 'edit' ? '更新' : '保存'
})
</script>

<template>
  <div class="ways-form">
    <div class="ways-form-layout" :class="{ 'ways-form-layout-edit': props.mode === 'edit' }">
      <aside v-if="props.mode !== 'edit'" class="ways-form-nav">
        <div class="ways-form-nav-header">
          <h3>{{ messages.channelType }}</h3>
          <span>{{ channelModeOptions.length }}{{ messages.availableTypeSuffix }}</span>
        </div>
        <div class="ways-form-channel-list" tabindex="0" @keydown="handleChannelListKeydown">
          <button
            v-for="(option, index) in channelModeOptions"
            :key="option.value"
            type="button"
            class="ways-form-channel-item"
            :class="{ 'is-active': option.value === channelMode }"
            @click="selectChannelMode(option.value, index)"
          >
            <span class="ways-form-channel-icon"><component :is="getChannelIcon(option.value)" /></span>
            <span class="ways-form-channel-name">{{ option.label }}</span>
            <span v-if="option.value === channelMode" class="ways-form-channel-check" aria-hidden="true">✓</span>
          </button>
        </div>
      </aside>

      <main class="ways-form-main">
        <transition name="fade-config" mode="out-in">
          <div v-if="currentChannelConfig" :key="channelMode" class="ways-form-content">
            <section v-if="channelNameInput || channelAuthInputs.length" class="ways-form-section">
              <header class="ways-form-section-header">
                <div class="ways-form-section-title"><span class="ways-form-identity-icon"><component :is="getChannelIcon(channelMode)" /></span><h4>{{ currentChannelConfig.label || channelMode }}{{ messages.configurationSuffix }}</h4><code>{{ channelMode }}</code><span v-if="currentChannelConfig.dynamicRecipient?.support" class="ways-form-capability">{{ messages.groupSendSupported }}</span></div>
                <p>{{ messages.configurationDescription }}</p>
              </header>
              <div class="ways-form-section-body ways-form-fields">
                <div v-if="channelNameInput" class="ways-form-field ways-form-field-name">
                  <label :for="channelNameInput.col">{{ channelNameInput.subLabel || channelNameInput.label }}</label>
                  <el-input
                    :id="channelNameInput.col"
                    v-model="formData[channelNameInput.col]"
                    clearable
                    :placeholder="channelNameInput.desc || channelNameInput.placeholder || channelNameInput.subLabel || channelNameInput.label"
                    :class="validationState.errors.value[channelNameInput.col] ? 'is-error' : ''"
                    @input="() => validationState.clearFieldError(channelNameInput.col)"
                  />
                  <div v-if="validationState.errors.value[channelNameInput.col]" class="app-form-error">{{ validationState.errors.value[channelNameInput.col] }}</div>
                </div>
                <div v-for="input in channelAuthInputs" :key="input.col" class="ways-form-field" :class="{ 'ways-form-field-wide': input.isTextArea }">
                  <label :for="input.col">
                    {{ input.subLabel || input.label }}
                    <small v-if="input.tips">{{ input.tips }}</small>
                  </label>
                  <el-input
                    v-if="input.isTextArea"
                    :id="input.col"
                    v-model="formData[input.col]"
                    type="textarea"
                    :rows="4"
                    :placeholder="input.desc || input.placeholder || input.subLabel || input.label"
                    :class="validationState.errors.value[input.col] ? 'is-error' : ''"
                    @input="() => validationState.clearFieldError(input.col)"
                  />
                  <el-input
                    v-else
                    :id="input.col"
                    v-model="formData[input.col]"
                    :type="isSensitiveInput(input) ? 'password' : 'text'"
                    :show-password="isSensitiveInput(input)"
                    clearable
                    :placeholder="input.desc || input.placeholder || input.subLabel || input.label"
                    :class="validationState.errors.value[input.col] ? 'is-error' : ''"
                    @input="() => validationState.clearFieldError(input.col)"
                  />
                  <div v-if="validationState.errors.value[input.col]" class="app-form-error">{{ validationState.errors.value[input.col] }}</div>
                </div>
              </div>
            </section>

            <section v-if="currentChannelConfig.dynamicRecipient?.support || currentChannelConfig.tips?.text" class="ways-form-notes">
              <div v-if="currentChannelConfig.dynamicRecipient?.support" class="ways-form-note">
                <MailOutlined />
                <div><strong>{{ messages.dynamicRecipients }}</strong><p>{{ messages.dynamicRecipientPrefix }}<code>recipients</code>{{ messages.dynamicRecipientSuffix }}{{ currentChannelConfig.dynamicRecipient.label }}{{ messages.sentenceEnd }}</p></div>
              </div>
              <el-tooltip v-if="currentChannelConfig.tips?.text" placement="top" effect="dark">
                <template #content><div class="max-w-md text-sm" v-html="currentChannelConfig.tips.desc"></div></template>
                <button type="button" class="ways-form-help">{{ currentChannelConfig.tips.text }} <span>?</span></button>
              </el-tooltip>
            </section>
          </div>
          <div v-else class="ways-form-empty">{{ messages.selectType }}</div>
        </transition>
      </main>
    </div>

    <footer class="ways-form-actions">
      <span>{{ messages.testBeforeSave }}</span>
      <div>
        <el-button @click="handleClose">{{ messages.cancel }}</el-button>
        <el-button @click="handleTest">{{ messages.testConnection }}</el-button>
        <el-button type="primary" @click="handleSave">{{ saveButtonText }}</el-button>
      </div>
    </footer>

    <el-dialog v-model="testRecipientDialogOpen" :title="messages.testRecipient" width="480px" class="app-nested-dialog" append-to-body>
        <div class="space-y-2">
          <label class="text-sm font-medium">{{ messages.recipientLabel }}</label>
          <el-input v-model="testRecipientInput" :placeholder="messages.recipientPlaceholder" clearable />
          <p class="text-xs text-muted-foreground">{{ messages.recipientHelp }}</p>
          <p v-if="testRecipientInput.trim()" class="text-xs" :class="isAllRecipient ? 'text-destructive' : 'text-emerald-600'">
            {{ isAllRecipient ? messages.allRecipientWarning : messages.singleRecipientNotice }}
          </p>
        </div>
        <template #footer>
          <el-button @click="testRecipientDialogOpen = false">{{ messages.cancel }}</el-button>
          <el-button type="primary" @click="confirmTestRecipient">{{ messages.confirmTest }}</el-button>
        </template>
    </el-dialog>

    <el-dialog v-model="testAllConfirmDialogOpen" :title="messages.highRiskConfirm" width="520px" class="app-nested-dialog" append-to-body>
        <p class="text-sm text-destructive">
          {{ messages.highRiskDescription }}
        </p>
        <template #footer>
          <el-button @click="cancelAllRecipientConfirm">{{ messages.cancel }}</el-button>
          <el-button type="danger" @click="submitQyWeiXinAppTest(testRecipientInput.trim())">{{ messages.sendAnyway }}</el-button>
        </template>
    </el-dialog>

  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'WaysForm'
})
</script>

<style scoped>
.ways-form {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  color: var(--foreground);
}

.ways-form-layout {
  display: grid;
  grid-template-columns: 214px minmax(0, 1fr);
  flex: 1;
  min-height: 0;
  background: var(--app-overlay-surface);
}

.ways-form-layout-edit {
  grid-template-columns: minmax(0, 1fr);
}

.ways-form-nav {
  display: flex;
  flex-direction: column;
  min-height: 0;
  border-right: 1px solid var(--app-overlay-border);
  background: color-mix(in srgb, var(--admin-surface-muted) 68%, var(--app-overlay-surface));
}

.ways-form-nav-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;
  padding: 0 14px;
  border-bottom: 1px solid var(--app-overlay-border);
}

.ways-form-nav-header h3 {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
}

.ways-form-nav-header span {
  color: var(--admin-text-muted);
  font-size: 11px;
}

.ways-form-channel-list {
  display: grid;
  gap: 3px;
  min-height: 0;
  padding: 8px;
  overflow-y: auto;
}

.ways-form-channel-item {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 18px;
  align-items: center;
  min-height: 38px;
  padding: 4px 8px;
  border: 1px solid transparent;
  border-radius: 7px;
  background: transparent;
  color: var(--admin-text-muted);
  cursor: pointer;
  text-align: left;
}

.ways-form-channel-item:hover {
  background: color-mix(in srgb, var(--brand-500) 6%, var(--app-overlay-surface));
  color: var(--foreground);
}

.ways-form-channel-item.is-active {
  border-color: color-mix(in srgb, var(--brand-500) 22%, var(--app-overlay-border));
  background: color-mix(in srgb, var(--brand-500) 10%, var(--app-overlay-surface));
  color: var(--brand-700);
}

.ways-form-channel-icon,
.ways-form-identity-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.ways-form-channel-icon :deep(svg) {
  width: 17px;
  height: 17px;
}

.ways-form-channel-name {
  overflow: hidden;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ways-form-channel-check {
  font-size: 12px;
  text-align: right;
}

.ways-form-main {
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  background: var(--glass-inset-bg);
}

.ways-form-content {
  display: grid;
  gap: 12px;
  padding: 16px;
}

.ways-form-section {
  border: 1px solid var(--app-overlay-border);
  border-radius: 9px;
  background: var(--app-overlay-surface);
}

.ways-form-identity-icon {
  flex: none;
  width: 38px;
  height: 38px;
  border-radius: 9px;
  background: color-mix(in srgb, var(--brand-500) 10%, var(--app-overlay-surface));
  color: var(--brand-700);
}

.ways-form-identity-icon :deep(svg) {
  width: 21px;
  height: 21px;
}

.ways-form-section-title {
  min-width: 0;
  flex-wrap: wrap;
}

.ways-form-section-title code,
.ways-form-capability {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 700;
}

.ways-form-section-title code {
  background: var(--admin-surface-muted);
  color: var(--admin-text-muted);
  font-family: monospace;
}

.ways-form-capability {
  background: color-mix(in srgb, var(--el-color-success) 10%, transparent);
  color: var(--el-color-success);
}

.ways-form-section-header p,
.ways-form-note p {
  margin: 4px 0 0;
  color: var(--admin-text-muted);
  font-size: 11px;
  line-height: 1.55;
}

.ways-form-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 38px;
  padding: 0 12px;
  border-bottom: 1px solid var(--app-overlay-border);
}

.ways-form-section-header > div {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ways-form-section-header span {
  color: var(--brand-600);
  font-family: monospace;
  font-size: 10px;
  font-weight: 800;
}

.ways-form-section-header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
}

.ways-form-section-header p {
  margin: 0;
}

.ways-form-section-body {
  padding: 14px;
}

.ways-form-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
}

.ways-form-field {
  min-width: 0;
}

.ways-form-field-name {
  max-width: 560px;
}

.ways-form-field-wide {
  grid-column: 1 / -1;
}

.ways-form-field label {
  display: block;
  margin-bottom: 6px;
  color: var(--foreground);
  font-size: 12px;
  font-weight: 600;
}

.ways-form-field label small {
  margin-left: 5px;
  color: var(--admin-text-muted);
  font-size: 10px;
  font-weight: 400;
}

.ways-form-notes {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border: 1px solid var(--app-overlay-border);
  border-radius: 8px;
  background: color-mix(in srgb, var(--brand-50) 32%, var(--app-overlay-surface));
}

.ways-form-note {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  min-width: 0;
}

.ways-form-note > :deep(svg) {
  flex: none;
  margin-top: 2px;
  color: var(--brand-600);
}

.ways-form-note strong {
  font-size: 12px;
}

.ways-form-note p {
  margin-top: 2px;
}

.ways-form-help {
  flex: none;
  border: 0;
  background: transparent;
  color: var(--admin-text-muted);
  cursor: help;
  font-size: 11px;
}

.ways-form-help span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  margin-left: 4px;
  border: 1px solid var(--app-overlay-border);
  border-radius: 50%;
}

.ways-form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 58px;
  padding: 10px 16px;
  border-top: 1px solid var(--app-overlay-border);
  background: var(--app-overlay-surface);
}

.ways-form-actions > span {
  color: var(--admin-text-muted);
  font-size: 11px;
}

.ways-form-actions > div {
  display: flex;
  gap: 8px;
}

.ways-form-empty {
  margin: 16px;
  padding: 28px;
  border: 1px dashed var(--app-overlay-border);
  border-radius: 9px;
  color: var(--admin-text-muted);
  text-align: center;
}

.fade-config-enter-active,
.fade-config-leave-active {
  transition: opacity 0.15s ease;
}

.fade-config-enter-from,
.fade-config-leave-to {
  opacity: 0;
}

@container app-managed-drawer (max-width: 720px) {
  .ways-form-layout {
    display: block;
    overflow-y: auto;
  }

  .ways-form-nav {
    max-height: 230px;
    border-right: 0;
    border-bottom: 1px solid var(--app-overlay-border);
  }

  .ways-form-main {
    overflow: visible;
  }

  .ways-form-content {
    padding: 12px;
  }

  .ways-form-fields {
    grid-template-columns: 1fr;
  }

  .ways-form-section-header {
    align-items: center;
    flex-direction: row;
  }

  .ways-form-notes,
  .ways-form-actions {
    align-items: flex-start;
    flex-direction: column;
  }

  .ways-form-section-header p,
  .ways-form-actions > span {
    display: none;
  }

  .ways-form-actions > div {
    width: 100%;
  }

  .ways-form-actions :deep(.el-button) {
    flex: 1;
  }
}
</style>
