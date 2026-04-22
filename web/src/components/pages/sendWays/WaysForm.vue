<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { CONSTANT } from '@/constant'
import { createValidationState } from '@/util/validation'
// import { validateForm, createValidationState, type InputConfig } from '@/util/validation'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from '@/components/ui/tooltip'
import {
  Mail,
  Zap,
  Building2,
  Bird,
  PlugZap,
  Smartphone,
  BellRing,
  QrCode,
  Send,
  Megaphone,
  ShieldCheck,
  Briefcase
} from 'lucide-vue-next'

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

// // 获取所有输入字段配置
// const getAllInputConfigs = (): InputConfig[] => {
//   const config = currentChannelConfig.value
//   if (!config) return []

//   const configs: InputConfig[] = []

//   // 基本输入字段
//   if (config.inputs) {
//     configs.push(...config.inputs.map((input: any) => ({
//       col: input.col,
//       label: input.label,
//       subLabel: input.subLabel,
//       type: input.type,
//       required: input.required !== false,
//       minLength: input.minLength,
//       maxLength: input.maxLength
//     })))
//   }

//   // 任务指令输入字段
//   if (config.taskInsInputs) {
//     configs.push(...config.taskInsInputs.map((input: any) => ({
//       col: input.col,
//       label: input.label,
//       subLabel: input.subLabel,
//       type: input.type,
//       required: input.required !== false,
//       minLength: input.minLength,
//       maxLength: input.maxLength
//     })))
//   }

//   return configs
// }

// // 校验表单
// const validateFormData = () => {
//   const inputConfigs = getAllInputConfigs()
//   const result = validateForm(formData.value, inputConfigs)

//   validationState.setErrors(result.errors)
//   return result.isValid
// }

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
        toast.error('测试接收者不能为空')
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
    const rsp = await request.post('/sendways/test', postData, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (rsp?.data?.code == 200) {
      toast.success(rsp.data.msg || '测试成功')
      return
    }
    toast.error(rsp?.data?.msg || '测试失败')
  } catch (error: any) {
    toast.error(error?.response?.data?.msg || error?.message || '测试请求失败')
  }
}

const confirmTestRecipient = async () => {
  const recipient = testRecipientInput.value.trim()
  if (!recipient) {
    toast.error('请输入接收者企微ID')
    return
  }
  if (recipient.toLowerCase() === 'all' || recipient === '@all') {
    testAllConfirmDialogOpen.value = true
    return
  }
  await submitQyWeiXinAppTest(recipient)
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
  // if (!validateFormData()) { return }

  try {
    const postData = getFinalData()

    // 根据模式选择API路径和成功消息
    const apiUrl = props.mode === 'edit' ? '/sendways/edit' : '/sendways/add'
    const successMessage = props.mode === 'edit' ? '更新渠道成功！' : '添加渠道成功！'

    const rsp = await request.post(apiUrl, postData)
    if (rsp?.data?.code == 200) {
      toast.success(successMessage)
      setTimeout(() => {
        window.location.reload()
      }, 1000)
      return
    }
    toast.error(rsp?.data?.msg || '保存失败')
  } catch (error: any) {
    toast.error(error?.response?.data?.msg || error?.message || '保存请求失败')
  }
}

// 渠道图标映射
const getChannelIcon = (type: string) => {
  const map: Record<string, any> = {
    'Email': Mail,
    'Dtalk': Zap,
    'QyWeiXin': Building2,
    'Feishu': Bird,
    'Custom': PlugZap,
    'WeChatOFAccount': QrCode,
    'AliyunSMS': Smartphone,
    'Telegram': Send,
    'Bark': BellRing,
    'Ntfy': Megaphone,
    'Gotify': ShieldCheck,
    'QyWeiXinApp': Briefcase
  }
  return map[type] || Send // Default icon
}

// 计算保存按钮文本
const saveButtonText = computed(() => {
  return props.mode === 'edit' ? '更新' : '保存'
})
</script>

<template>
  <div class="w-full h-full flex flex-col">
    <div class="flex flex-col lg:flex-row gap-6 flex-1">
      <div v-if="props.mode !== 'edit'" class="lg:w-2/5 w-full">
        <div
          class="mt-3 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-[#111827] p-2 max-h-[60vh] overflow-y-auto focus:outline-none"
          tabindex="0"
          @keydown="handleChannelListKeydown"
        >
          <button
            v-for="(option, index) in channelModeOptions"
            :key="option.value"
            type="button"
            class="group flex h-9 w-full items-center justify-between rounded-md px-2 text-sm mb-1 last:mb-0"
            :class="option.value === channelMode
              ? 'bg-[#3b82f6] text-white'
              : 'text-gray-800 dark:text-gray-100 hover:bg-[#f1f5f9] dark:hover:bg-slate-700'"
            @click="selectChannelMode(option.value, index)"
          >
            <div class="flex items-center gap-2">
              <div class="w-5 h-5 flex items-center justify-center">
                <component
                  :is="getChannelIcon(option.value)"
                  class="w-5 h-5"
                  :class="option.value === channelMode ? 'text-white' : 'text-gray-700 dark:text-gray-300'"
                />
              </div>
              <span>{{ option.label }}</span>
            </div>
            <svg
              v-if="option.value === channelMode"
              class="h-3.5 w-3.5 text-white"
              viewBox="0 0 16 16"
              aria-hidden="true"
            >
              <path
                fill="currentColor"
                d="M6.173 12.414 2.4 8.64l1.414-1.414L6.173 9.586l6.01-6.01 1.414 1.414-7.424 7.424z"
              />
            </svg>
          </button>
        </div>
      </div>

      <div
        :class="props.mode === 'edit' ? 'w-full' : 'lg:w-3/5 w-full'"
        class="flex flex-col min-h-0"
      >
        <transition name="fade-config" mode="out-in">
          <div v-if="currentChannelConfig" :key="channelMode" class="mt-2 lg:mt-0">
            <div class="mb-4">
              <div
                v-if="props.mode === 'edit'"
                class="flex items-center gap-1.5 text-sm text-gray-700 dark:text-gray-300"
              >
                <span class="font-medium">{{ currentChannelConfig?.label || channelMode }}</span>
                <span
                  v-if="currentChannelConfig?.dynamicRecipient?.support"
                  class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-brand/10 text-brand dark:bg-brand/20 dark:text-brand"
                >
                  群发
                </span>
              </div>
              <div v-else>
                <div class="flex items-center gap-2">
                  <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100">
                    {{ currentChannelConfig?.label || '请选择通道' }}
                  </h3>
                  <span
                    v-if="currentChannelConfig?.dynamicRecipient?.support"
                    class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-brand/10 text-brand dark:bg-brand/20 dark:text-brand"
                  >
                    群发
                  </span>
                </div>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  根据选择的发信通道配置认证信息和必要参数
                </p>
              </div>
            </div>

            <div
              v-if="currentChannelConfig.dynamicRecipient?.support"
              class="mb-4 p-3 bg-[#1e3a8a] border border-[#1e3a8a] rounded-md"
            >
              <div class="flex items-start gap-2">
                <Mail class="w-4 h-4 text-white mt-0.5" />
                <div class="flex-1 space-y-1">
                  <p class="text-xs text-white font-medium">
                    支持群发模式 - 可在配置实例时启用"动态接收者"，通过 API 的
                    <code class="px-1 py-0.5 bg-white/10 rounded text-[11px] text-white">recipients</code>
                    参数指定多个{{ currentChannelConfig.dynamicRecipient.label }}
                  </p>
                  <p class="text-[11px] text-white/80">
                    适用：邮件群发、公众号批量推送、营销通知等
                  </p>
                </div>
              </div>
            </div>

            <div v-if="currentChannelConfig.inputs && currentChannelConfig.inputs.length > 0" class="mb-8">
              <h4 class="text-base font-medium mb-4 text-gray-800 dark:text-gray-100">基本配置</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div
                  v-for="input in currentChannelConfig.inputs"
                  :key="input.col"
                  class="space-y-2"
                  :class="{
                    'md:col-span-2': input.isTextArea
                  }"
                >
                  <Label :for="input.col" class="text-sm font-medium">
                    {{ input.subLabel || input.label }}
                    <span v-if="input.tips" class="text-xs text-gray-500 ml-1">({{ input.tips }})</span>
                  </Label>
                  <Textarea
                    v-if="input.isTextArea"
                    :id="input.col"
                    v-model="formData[input.col]"
                    :placeholder="input.desc || input.placeholder || input.subLabel || input.label"
                    :class="[
                      'w-full rounded-md border border-[#d1d5da] bg-white dark:bg-slate-900 placeholder:text-gray-500 focus:border-2 focus:border-[#3b82f6] focus:ring-0 focus-visible:ring-0 transition-transform transition-colors duration-150 focus:scale-[1.01]',
                      validationState.errors.value[input.col] ? 'border-red-500 focus:border-red-500' : ''
                    ]"
                    @input="() => validationState.clearFieldError(input.col)"
                  />
                  <Input
                    v-else
                    :id="input.col"
                    v-model="formData[input.col]"
                    :placeholder="input.desc || input.placeholder || input.subLabel || input.label"
                    :class="[
                      'w-full rounded-md border border-[#d1d5da] placeholder:text-gray-500 focus:border-2 focus:border-[#3b82f6] focus:ring-0 focus-visible:ring-0 transition-transform transition-colors duration-150 focus:scale-[1.01]',
                      validationState.errors.value[input.col] ? 'border-red-500 focus:border-red-500' : ''
                    ]"
                    @input="() => validationState.clearFieldError(input.col)"
                  />
                  <div v-if="validationState.errors.value[input.col]" class="text-red-500 text-xs mt-1">
                    {{ validationState.errors.value[input.col] }}
                  </div>
                </div>
              </div>

              <div class="mt-2 ml-4" v-if="currentChannelConfig.tips && currentChannelConfig.tips.text">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger class="text-sm hover:text-gray-700 inline-flex items-center gap-1">
                      {{ currentChannelConfig.tips.text }}
                      <span
                        class="cursor-help inline-flex items-center justify-center w-4 h-4 rounded-full border border-gray-300 hover:border-gray-400 text-xs"
                      >
                        ?
                      </span>
                    </TooltipTrigger>
                    <TooltipContent class="max-w-md">
                      <div class="text-sm" v-html="currentChannelConfig.tips.desc"></div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            </div>
          </div>
          <div v-else class="mt-6 p-6 bg-gray-50 dark:bg-slate-900/40 rounded-lg">
            <p class="text-gray-500 dark:text-gray-400">请选择一个渠道类型开始配置</p>
          </div>
        </transition>
      </div>
    </div>

    <div
      class="flex justify-end gap-3 mt-8 pt-4 border-t sticky bottom-0 bg-white dark:bg-slate-950"
    >
      <Button variant="outline" class="transition-transform hover:-translate-y-0.5" @click="handleClose">
        取消
      </Button>
      <Button class="transition-transform hover:-translate-y-0.5" @click="handleTest">
        测试
      </Button>
      <Button
        class="transition-transform hover:-translate-y-0.5 bg-gradient-to-r from-brand-500 to-brand-600 hover:from-brand-600 hover:to-brand-700 text-white"
        @click="handleSave"
      >
        {{ saveButtonText }}
      </Button>
    </div>

    <Dialog v-model:open="testRecipientDialogOpen">
      <DialogContent class="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>测试接收者</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <Label>请输入接收者企微ID</Label>
          <Input v-model="testRecipientInput" placeholder="例如：zhangsan" />
          <p class="text-xs text-muted-foreground">仅用于本次测试发送，不会写入渠道默认配置。</p>
          <p v-if="testRecipientInput.trim()" class="text-xs" :class="isAllRecipient ? 'text-destructive' : 'text-emerald-600'">
            {{ isAllRecipient ? '当前输入为全员发送（高风险）' : '当前输入为单用户发送' }}
          </p>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="testRecipientDialogOpen = false">取消</Button>
          <Button type="button" @click="confirmTestRecipient">确认测试</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="testAllConfirmDialogOpen">
      <DialogContent class="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>高风险确认</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-destructive">
          你输入的是 all/@all，这将把测试消息发送给企业微信应用可触达的所有成员，风险较高，请再次确认。
        </p>
        <DialogFooter>
          <Button type="button" variant="outline" @click="testAllConfirmDialogOpen = false">取消</Button>
          <Button type="button" variant="destructive" @click="submitQyWeiXinAppTest(testRecipientInput.trim())">仍然发送</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'WaysForm'
})
</script>

<style scoped>
.fade-config-enter-active,
.fade-config-leave-active {
  transition: opacity 0.15s ease-out, transform 0.15s ease-out;
}

.fade-config-enter-from,
.fade-config-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
