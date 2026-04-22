<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'

interface MQSourceOption {
  id: string
  name: string
}

interface TemplateOption {
  id: string
  name: string
}

interface Props {
  data?: {
    id: string
    source_id: string
    name: string
    topic: string
    tag: string
    group_name: string
    validate_regex: string
    extract_regex: string
    extract_field: string
    extract_rules?: Array<{ field: string; regex: string }>
    template_id: string
    template_content_type?: string
    consume_mode?: string
  } | null
  sourceOptions: MQSourceOption[]
  templateOptions: TemplateOption[]
}

const props = withDefaults(defineProps<Props>(), {
  data: null
})

const emit = defineEmits<{
  success: []
}>()

const isEdit = computed(() => !!props.data)
const createExtractRule = (field = '', regex = '') => ({
  field,
  regex
})

const normalizeTemplateContentType = (value?: string) => {
  const v = String(value || '').trim().toLowerCase()
  if (v === 'html' || v === 'markdown' || v === 'text') return v
  if (v === 'push' || v === 'pull' || v === '') return 'text'
  return 'text'
}

const formData = reactive({
  source_id: props.data?.source_id || '',
  name: props.data?.name || '',
  topic: props.data?.topic || '',
  tag: props.data?.tag || '',
  group_name: props.data?.group_name || 'mq_consumer_group',
  validate_regex: props.data?.validate_regex || '',
  extract_regex: props.data?.extract_regex || '',
  extract_field: props.data?.extract_field || '',
  extract_rules: props.data?.extract_rules?.length
    ? props.data.extract_rules.map(r => createExtractRule(r.field, r.regex))
    : (props.data?.extract_field || props.data?.extract_regex
      ? [createExtractRule(props.data?.extract_field || '', props.data?.extract_regex || '')]
      : [createExtractRule('to_user', '')]),
  template_id: props.data?.template_id || '',
  template_content_type: normalizeTemplateContentType(props.data?.template_content_type || props.data?.consume_mode)
})

const isSubmitting = ref(false)
const validateSyntaxError = ref('')
const extractSyntaxError = ref('')
const isSyntaxChecking = ref(false)
let syntaxTimer: number | null = null
const activeRuleField = ref<'validate' | 'extract' | null>(null)
const activeExtractRuleIndex = ref<number | null>(null)
const isSuggestionHovering = ref(false)
const isDeleteRuleDialogOpen = ref(false)
const pendingDeleteRuleIndex = ref<number | null>(null)
const deleteRuleConfirmInput = ref('')

// 正则测试相关
const testMessage = ref('')
const testResult = ref<{
  validateMatched: boolean | null
  extractResult: Record<string, string> | null
  error: string | null
} | null>(null)
const isTestingRegex = ref(false)
const extractedEntries = computed(() => Object.entries(testResult.value?.extractResult || {}))
const dynamicRecipientPreview = computed(() => {
  const data = testResult.value?.extractResult || {}
  const raw = String((data as any).to_user || '').trim()
  if (!raw) return []
  const tokens = raw
    .split(/[|,;\s]+/)
    .map(v => v.trim())
    .filter(Boolean)
  return Array.from(new Set(tokens))
})
const activeSection = ref('basic')

const sectionItems = [
  { key: 'basic', label: '基本信息' },
  { key: 'regex', label: '正则配置' },
  { key: 'test', label: '正则测试' },
  { key: 'template', label: '模板配置' }
]

const stripRulePrefix = (rule: string) => {
  const raw = String(rule || '').trim()
  const lower = raw.toLowerCase()
  if (lower.startsWith('dsl:')) return raw.slice(4).trim()
  return raw
}

const ensureDSLRule = (rule: string) => {
  const core = stripRulePrefix(rule)
  if (!core) return ''
  return `dsl:${core}`
}

const dslFunctionHints = [
  { name: 'contains', snippet: 'contains($.department, "研发部")', tip: '包含判断' },
  { name: 'equals', snippet: 'equals($.status, "created")', tip: '相等判断' },
  { name: 'exists', snippet: 'exists($.name)', tip: '字段存在且非空' },
  { name: 'regex', snippet: 'regex($.text, ".*告警.*")', tip: '正则匹配' },
  { name: 'in', snippet: 'in($.level, "P5", "P6")', tip: '集合包含' },
  { name: 'gt', snippet: 'gt($.cost, 100)', tip: '大于比较' },
  { name: 'gte', snippet: 'gte($.cost, 100)', tip: '大于等于比较' },
  { name: 'lt', snippet: 'lt($.cost, 100)', tip: '小于比较' },
  { name: 'lte', snippet: 'lte($.cost, 100)', tip: '小于等于比较' },
  { name: 'between', snippet: 'between($.cost, 50, 100)', tip: '区间比较' },
  { name: 'empty', snippet: 'empty($.name)', tip: '空值判断' },
  { name: 'notEmpty', snippet: 'notEmpty($.name)', tip: '非空判断' },
  { name: 'pick', snippet: 'pick($.name)', tip: '提取 JSON 字段值' },
  { name: 'lower', snippet: 'lower(pick($.name))', tip: '转小写' },
  { name: 'upper', snippet: 'upper(pick($.name))', tip: '转大写' },
  { name: 'trim', snippet: 'trim(pick($.name))', tip: '去空白' },
  { name: 'replace', snippet: 'replace(pick($.name), ".", "_")', tip: '字符串替换' },
  { name: 'concat', snippet: 'concat("user:", pick($.name))', tip: '拼接字符串' },
  { name: 'split', snippet: 'split(pick($.email), "@", 0)', tip: '分割并取索引' },
  { name: 'regexAll', snippet: 'regexAll(raw, "(?s)\\\\{[^{}]*\\"id\\"\\\\s*:\\\\s*(\\\\d+)[^{}]*\\"value\\"\\\\s*:\\\\s*\\"target\\"[^{}]*\\\\}", 1, "|")', tip: '提取全部匹配' },
  { name: 'findIdsByValue', snippet: 'findIdsByValue($, "target", "id", "|")', tip: '按 JSON 递归查找 id' },
  { name: 'arrayLen', snippet: 'arrayLen($.children)', tip: '获取数组长度' },
  { name: 'valuesByKey', snippet: 'valuesByKey($, "id", "|")', tip: '递归提取指定 key 值' },
  { name: 'findByField', snippet: 'findByField($, "value", "target", "id", "|")', tip: '按字段过滤并返回目标字段' },
  { name: 'findByFieldRaw', snippet: 'findByFieldRaw($, "value", "target", "|")', tip: '按字段过滤并返回对象 JSON' },
  { name: 'len', snippet: 'len(pick($.name))', tip: '字符串长度' },
  { name: 'substr', snippet: 'substr(pick($.name), 0, 4)', tip: '字符串截取' },
  { name: 'toInt', snippet: 'toInt(pick($.count))', tip: '转整数' },
  { name: 'toFloat', snippet: 'toFloat(pick($.price))', tip: '转浮点' },
  { name: 'add', snippet: 'add(toFloat($.a), toFloat($.b))', tip: '数值相加' },
  { name: 'sub', snippet: 'sub(toFloat($.a), toFloat($.b))', tip: '数值相减' },
  { name: 'mul', snippet: 'mul(toFloat($.a), toFloat($.b))', tip: '数值相乘' },
  { name: 'div', snippet: 'div(toFloat($.a), toFloat($.b))', tip: '数值相除' },
  { name: 'default', snippet: 'default(pick($.name), "unknown")', tip: '空值回退' },
  { name: 'coalesce', snippet: 'coalesce(pick($.name), pick($.nickname), "unknown")', tip: '多值回退' },
  { name: 'if', snippet: 'if(contains($.department, "研发部"), pick($.name), "")', tip: '条件表达式' }
]

const getCurrentDSLToken = (rule: string) => {
  const raw = String(rule || '').trim().toLowerCase()
  if (!raw.startsWith('dsl:')) return ''
  const core = stripRulePrefix(rule)
  const match = core.match(/([a-zA-Z_][a-zA-Z0-9_]*)$/)
  return match ? match[1] : ''
}

const getSuggestions = (rule: string) => {
  const token = getCurrentDSLToken(rule)
  if (token === '') return dslFunctionHints
  return dslFunctionHints
    .filter((item) => item.name.toLowerCase().startsWith(token.toLowerCase()))
}

const validateSuggestions = computed(() =>
  activeRuleField.value === 'validate' ? getSuggestions(formData.validate_regex) : []
)
const extractSuggestions = computed(() =>
  activeRuleField.value === 'extract' && activeExtractRuleIndex.value !== null
    ? getSuggestions(formData.extract_rules[activeExtractRuleIndex.value]?.regex || '')
    : []
)

const applySuggestion = (field: 'validate' | 'extract', snippet: string) => {
  const current = field === 'validate'
    ? formData.validate_regex
    : (formData.extract_rules[activeExtractRuleIndex.value ?? 0]?.regex || '')
  const normalized = ensureDSLRule(current || 'dsl:')
  const core = stripRulePrefix(normalized)
  const match = core.match(/([a-zA-Z_][a-zA-Z0-9_]*)$/)
  const token = match ? match[1] : ''
  const nextCore = token ? `${core.slice(0, core.length - token.length)}${snippet}` : `${core}${snippet}`
  if (field === 'validate') {
    formData.validate_regex = `dsl:${nextCore}`
  } else {
    const idx = activeExtractRuleIndex.value ?? 0
    if (!formData.extract_rules[idx]) {
      formData.extract_rules[idx] = createExtractRule()
    }
    formData.extract_rules[idx].regex = `dsl:${nextCore}`
  }
}

const handleRuleBlur = () => {
  window.setTimeout(() => {
    if (isSuggestionHovering.value) return
    activeRuleField.value = null
    activeExtractRuleIndex.value = null
  }, 120)
}

const handleSuggestionMouseEnter = () => {
  isSuggestionHovering.value = true
}

const handleSuggestionMouseLeave = () => {
  isSuggestionHovering.value = false
  // 光标离开建议面板后，若输入框已失焦，则延迟收起
  window.setTimeout(() => {
    const activeEl = document.activeElement as HTMLElement | null
    const id = activeEl?.id || ''
    const isRuleInputFocused = id === 'validate_regex' || id.startsWith('extract_regex_')
    if (!isRuleInputFocused) {
      activeRuleField.value = null
      activeExtractRuleIndex.value = null
    }
  }, 80)
}

const addExtractRule = () => {
  formData.extract_rules.push(createExtractRule('', ''))
  activeRuleField.value = null
  activeExtractRuleIndex.value = null
}

const addExtractRuleTemplate = (field: string, regex: string) => {
  formData.extract_rules.push(createExtractRule(field, ensureDSLRule(regex)))
  activeRuleField.value = null
  activeExtractRuleIndex.value = null
}

const insertJsonTemplate = (snippet: string) => {
  const idx = activeExtractRuleIndex.value
  if (idx === null || !formData.extract_rules[idx]) {
    return
  }
  formData.extract_rules[idx].regex = ensureDSLRule(snippet)
}

const removeExtractRule = (idx: number) => {
  if (formData.extract_rules.length <= 1) {
    formData.extract_rules[0] = createExtractRule('to_user', '')
    return
  }
  formData.extract_rules.splice(idx, 1)
  if (activeExtractRuleIndex.value === idx) {
    activeExtractRuleIndex.value = null
    activeRuleField.value = null
  }
}

const requestRemoveExtractRule = (idx: number) => {
  pendingDeleteRuleIndex.value = idx
  deleteRuleConfirmInput.value = ''
  isDeleteRuleDialogOpen.value = true
}

const resetDeleteRuleDialog = () => {
  isDeleteRuleDialogOpen.value = false
  pendingDeleteRuleIndex.value = null
  deleteRuleConfirmInput.value = ''
}

const canConfirmDeleteRule = computed(() => {
  const idx = pendingDeleteRuleIndex.value
  if (idx === null) return false
  const target = formData.extract_rules[idx]
  if (!target) return false
  const targetField = (target.field || '').trim()
  if (!targetField) return true
  return deleteRuleConfirmInput.value.trim() === targetField
})

const deleteRuleTargetField = computed(() => {
  const idx = pendingDeleteRuleIndex.value
  if (idx === null) return ''
  return (formData.extract_rules[idx]?.field || '').trim()
})

const deleteRuleMatchStatusText = computed(() => {
  if (!isDeleteRuleDialogOpen.value) return ''
  if (!deleteRuleTargetField.value) return '字段名为空，可直接确认删除'
  if (!deleteRuleConfirmInput.value.trim()) return '请输入字段名进行确认'
  return canConfirmDeleteRule.value ? '字段名已匹配，可删除' : '字段名未匹配'
})

const confirmRemoveExtractRule = () => {
  const idx = pendingDeleteRuleIndex.value
  if (idx === null) return
  if (!canConfirmDeleteRule.value) return
  removeExtractRule(idx)
  resetDeleteRuleDialog()
}

const copyExtractRule = (idx: number) => {
  const rule = formData.extract_rules[idx]
  if (!rule) return
  formData.extract_rules.splice(idx + 1, 0, createExtractRule(rule.field, rule.regex))
}

const moveExtractRule = (idx: number, direction: -1 | 1) => {
  const target = idx + direction
  if (target < 0 || target >= formData.extract_rules.length) return
  const current = formData.extract_rules[idx]
  formData.extract_rules[idx] = formData.extract_rules[target]
  formData.extract_rules[target] = current
  if (activeExtractRuleIndex.value === idx) {
    activeExtractRuleIndex.value = target
  } else if (activeExtractRuleIndex.value === target) {
    activeExtractRuleIndex.value = idx
  }
}

const scrollToSection = (key: string) => {
  activeSection.value = key
}

watch(
  () => props.data,
  (val) => {
    formData.source_id = val?.source_id || ''
    formData.name = val?.name || ''
    formData.topic = val?.topic || ''
    formData.tag = val?.tag || ''
    formData.group_name = val?.group_name || 'mq_consumer_group'
    formData.validate_regex = val?.validate_regex || ''
    formData.extract_regex = val?.extract_regex || ''
    formData.extract_field = val?.extract_field || ''
    formData.extract_rules = val?.extract_rules?.length
      ? val.extract_rules.map(r => createExtractRule(r.field, r.regex))
      : (val?.extract_field || val?.extract_regex
        ? [createExtractRule(val?.extract_field || '', val?.extract_regex || '')]
        : [createExtractRule('to_user', '')])
    formData.template_id = val?.template_id || ''
    formData.template_content_type = normalizeTemplateContentType(val?.template_content_type || val?.consume_mode)
    formData.validate_regex = ensureDSLRule(formData.validate_regex)
    formData.extract_rules = formData.extract_rules.map(r => createExtractRule(r.field, ensureDSLRule(r.regex)))
  },
  { immediate: true, deep: true }
)

// 运行正则测试
const runRegexTest = async () => {
  if (!testMessage.value) {
    toast.warning('请输入测试消息内容')
    return
  }

  testResult.value = {
    validateMatched: null,
    extractResult: null,
    error: null
  }

  isTestingRegex.value = true
  try {
    const rsp = await request.post('/subscriptions/regex-test', {
      message: testMessage.value,
      validate_regex: formData.validate_regex,
      extract_rules: formData.extract_rules
        .filter(r => (r.field || '').trim() !== '' || (r.regex || '').trim() !== '')
        .map(r => ({
          field: (r.field || '').trim(),
          regex: ensureDSLRule(r.regex || '')
        }))
    })
    const data = rsp?.data?.data || {}
    testResult.value = {
      validateMatched: data.validate_matched ?? true,
      extractResult: data.extracted_values || null,
      error: null
    }
  } catch (e: any) {
    testResult.value.error = e?.response?.data?.msg || e?.message || '正则测试失败'
  } finally {
    isTestingRegex.value = false
  }
}

// 当正则变化时清空测试结果
watch([() => formData.validate_regex, () => JSON.stringify(formData.extract_rules)], () => {
  if (testResult.value) {
    testResult.value = null
  }
})

const checkRuleSyntax = async () => {
  const normalizedRules = formData.extract_rules
    .filter(r => (r.field || '').trim() !== '' || (r.regex || '').trim() !== '')
    .map(r => ({
      field: (r.field || '').trim(),
      regex: ensureDSLRule(r.regex || '')
    }))

  if (!formData.validate_regex && normalizedRules.length === 0) {
    validateSyntaxError.value = ''
    extractSyntaxError.value = ''
    return
  }
  isSyntaxChecking.value = true
  try {
    await request.post('/subscriptions/regex-test', {
      message: testMessage.value || '{"department":"平台研发部","name":"kanfa.hu","text":"demo"}',
      validate_regex: formData.validate_regex,
      extract_rules: normalizedRules
    }, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    validateSyntaxError.value = ''
    extractSyntaxError.value = ''
  } catch (e: any) {
    const msg = e?.response?.data?.msg || e?.message || '规则语法错误'
    if (msg.includes('验证正则')) {
      validateSyntaxError.value = msg
      extractSyntaxError.value = ''
    } else if (msg.includes('提取正则') || msg.includes('提取')) {
      extractSyntaxError.value = msg
      validateSyntaxError.value = ''
    } else {
      // 未明确归属时，两侧都提示同一错误，避免漏报
      validateSyntaxError.value = formData.validate_regex ? msg : ''
      extractSyntaxError.value = formData.extract_rules.some(r => (r.field || '').trim() !== '' || (r.regex || '').trim() !== '') ? msg : ''
    }
  } finally {
    isSyntaxChecking.value = false
  }
}

const normalizeRuleErrorMessage = (msg: string) => {
  const raw = String(msg || '')
  if (!raw) return '规则语法错误'
  if (raw.includes('DSL 布尔表达式非法')) {
    return '验证规则语法不完整，请按函数形式输入，如：dsl:contains($.department, "研发部")'
  }
  if (raw.includes('DSL 函数不支持')) {
    return '规则函数不支持，请检查函数名（支持 contains/equals/exists/regex/if 等）'
  }
  if (raw.includes('DSL 参数语法不完整') || raw.includes('DSL 参数括号不匹配')) {
    return '规则参数或括号不完整，请检查逗号、括号和引号是否闭合'
  }
  return raw
}

watch(
  [() => formData.validate_regex, () => JSON.stringify(formData.extract_rules)],
  () => {
    formData.validate_regex = ensureDSLRule(formData.validate_regex)
    formData.extract_rules = formData.extract_rules.map(r => createExtractRule(r.field, ensureDSLRule(r.regex)))
    if (syntaxTimer) {
      window.clearTimeout(syntaxTimer)
      syntaxTimer = null
    }
    syntaxTimer = window.setTimeout(() => {
      checkRuleSyntax()
    }, 450)
  }
)

const handleSubmit = async () => {
  // 验证必填项
  if (!formData.source_id) {
    toast.warning('请选择数据源')
    return
  }
  if (!formData.name) {
    toast.warning('请输入订阅名称')
    return
  }
  if (!formData.topic) {
    toast.warning('请输入 Topic')
    return
  }
  if (!formData.group_name) {
    toast.warning('请输入 Group Name')
    return
  }
  if (!formData.template_id) {
    toast.warning('请选择消息模板')
    return
  }
  await checkRuleSyntax()
  if (validateSyntaxError.value || extractSyntaxError.value) {
    toast.warning('规则语法校验未通过，请先修正')
    return
  }

  isSubmitting.value = true
  try {
    const url = isEdit.value
      ? `/subscriptions/${props.data?.id}/edit`
      : '/subscriptions/add'
    
    const payload = {
      source_id: formData.source_id,
      name: formData.name,
      topic: formData.topic,
      tag: formData.tag,
      group_name: formData.group_name,
      validate_regex: formData.validate_regex,
      extract_rules: formData.extract_rules
        .filter(r => (r.field || '').trim() !== '' || (r.regex || '').trim() !== '')
        .map(r => ({
          field: (r.field || '').trim(),
          regex: ensureDSLRule(r.regex || '')
        })),
      template_id: formData.template_id,
      template_content_type: normalizeTemplateContentType(formData.template_content_type),
      consume_mode: normalizeTemplateContentType(formData.template_content_type)
    }

    const res = await request.post(url, payload, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (res.data.code === 200) {
      toast.success(isEdit.value ? '编辑成功' : '新增成功')
      emit('success')
    } else {
      toast.error(normalizeRuleErrorMessage(res?.data?.msg || '操作失败'))
    }
  } catch (error: any) {
    const msg = error?.response?.data?.msg || '操作失败'
    toast.error(normalizeRuleErrorMessage(msg))
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="h-full flex flex-col min-h-0">
    <div class="min-h-0 flex-1 grid grid-cols-1 md:grid-cols-[180px_minmax(0,1fr)]">
      <aside class="hidden md:block border-r bg-muted/20 p-3 overflow-y-auto">
        <div class="space-y-1">
          <button
            v-for="item in sectionItems"
            :key="item.key"
            type="button"
            class="w-full text-left px-3 py-2 text-sm rounded transition-colors"
            :class="activeSection === item.key ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
            @click="scrollToSection(item.key)"
          >
            {{ item.label }}
          </button>
        </div>
      </aside>

      <div class="overflow-y-auto p-4 md:p-5 space-y-6">
        <section v-show="activeSection === 'basic'" id="sub-section-basic" class="space-y-4">
          <h4 class="text-sm font-semibold text-muted-foreground">基本信息</h4>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="source_id">数据源 <span class="text-destructive">*</span></Label>
        <Select v-model="formData.source_id">
          <SelectTrigger>
            <SelectValue placeholder="选择数据源" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem v-for="opt in sourceOptions" :key="opt.id" :value="opt.id">
                {{ opt.name }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div class="space-y-2">
        <Label for="name">订阅名称 <span class="text-destructive">*</span></Label>
        <Input
          id="name"
          v-model="formData.name"
          placeholder="例如：订单异常告警订阅"
          maxlength="200"
        />
      </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="topic">Topic <span class="text-destructive">*</span></Label>
              <Input
                id="topic"
                v-model="formData.topic"
                placeholder="例如：ORDER_EXCEPTION"
                maxlength="200"
              />
            </div>

            <div class="space-y-2">
              <Label for="tag">Tag</Label>
              <Input
                id="tag"
                v-model="formData.tag"
                placeholder="可选，例如：prod 或 tag1||tag2"
                maxlength="200"
              />
              <p class="text-xs text-muted-foreground">
                多个 Tag 用 || 分隔，留空表示订阅全部
              </p>
            </div>
            </div>

            <div class="space-y-2">
            <Label for="group_name">Consumer Group <span class="text-destructive">*</span></Label>
            <Input
              id="group_name"
              v-model="formData.group_name"
              placeholder="例如：mq_consumer_group"
              maxlength="200"
            />
            <p class="text-xs text-muted-foreground">
              消费者组名称，同一组内负载均衡消费
            </p>
            </div>
          </div>
        </section>

        <section v-show="activeSection === 'regex'" id="sub-section-regex" class="space-y-4">
          <h4 class="text-sm font-semibold text-muted-foreground">正则配置</h4>
          <div class="space-y-4">
            <div class="space-y-2">
      <Label for="validate_regex">验证正则</Label>
      <div class="relative">
        <Textarea
          id="validate_regex"
          v-model="formData.validate_regex"
          placeholder='可选。DSL 示例：dsl:contains($.department, "研发部") && exists($.name)'
          rows="2"
          @focus="activeRuleField = 'validate'"
          @blur="handleRuleBlur"
        />
        <div
          v-if="validateSuggestions.length > 0"
          class="absolute z-20 mt-1 w-full rounded-md border bg-background p-2 shadow-sm space-y-1 h-[158px] overflow-y-auto overscroll-contain"
          @mouseenter="handleSuggestionMouseEnter"
          @mouseleave="handleSuggestionMouseLeave"
          @wheel.stop
        >
          <button
            v-for="item in validateSuggestions"
            :key="`v-${item.name}`"
            type="button"
            class="w-full text-left rounded px-2 py-1.5 hover:bg-muted transition-colors"
            @mousedown.prevent="applySuggestion('validate', item.snippet)"
          >
            <div class="text-xs font-medium">{{ item.name }}</div>
            <div class="text-[11px] text-muted-foreground">{{ item.tip }}：{{ item.snippet }}</div>
          </button>
        </div>
      </div>
      <p class="text-xs text-muted-foreground">
        仅支持 DSL，建议使用 dsl: 前缀（contains/equals/exists/regex/gt/gte/lt/lte/&&/||/!）
      </p>
      <p v-if="validateSyntaxError" class="text-xs text-destructive">{{ validateSyntaxError }}</p>
            </div>

            <div class="space-y-3 rounded-xl border bg-card/40 p-3 md:p-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <Label class="text-sm font-semibold">提取字段组</Label>
          <Badge variant="secondary">共 {{ formData.extract_rules.length }} 组</Badge>
        </div>
        <Button type="button" size="sm" variant="outline" class="h-8 px-3" @click="addExtractRule">新增字段组</Button>
      </div>
      <div class="flex flex-wrap items-center gap-2 rounded-lg border border-dashed border-border/70 bg-muted/40 p-2">
        <span class="text-xs text-muted-foreground">快速模板:</span>
        <Button type="button" size="sm" variant="ghost" class="h-7 px-2 text-xs" @click="addExtractRuleTemplate('to_user', 'dsl:pick($.to_user)')">
          to_user
        </Button>
        <Button type="button" size="sm" variant="ghost" class="h-7 px-2 text-xs" @click="addExtractRuleTemplate('name', 'dsl:pick($.name)')">
          name
        </Button>
        <Button type="button" size="sm" variant="ghost" class="h-7 px-2 text-xs" @click="addExtractRuleTemplate('text', 'dsl:pick($.text)')">
          text
        </Button>
      </div>
      <div
        v-for="(rule, idx) in formData.extract_rules"
        :key="`extract-rule-${idx}`"
        class="rounded-lg border border-border/70 bg-background p-3 space-y-3 shadow-sm transition-all hover:shadow-md"
      >
        <div class="flex items-center justify-between">
          <Badge variant="outline" class="font-normal bg-muted/40">字段组 {{ idx + 1 }}</Badge>
          <div class="flex items-center gap-1 rounded-md border bg-muted/30 p-1">
            <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-xs" :disabled="idx === 0" @click="moveExtractRule(idx, -1)">上移</Button>
            <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-xs" :disabled="idx === formData.extract_rules.length - 1" @click="moveExtractRule(idx, 1)">下移</Button>
            <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-xs" @click="copyExtractRule(idx)">复制</Button>
            <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-xs text-destructive hover:text-destructive" @click="requestRemoveExtractRule(idx)">删除</Button>
          </div>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-start">
          <div class="space-y-1 md:col-span-4">
            <Label :for="`extract_field_${idx}`">提取字段名</Label>
            <Input
              :id="`extract_field_${idx}`"
              v-model="rule.field"
              placeholder="例如：to_user"
            />
          </div>
          <div class="space-y-1 md:col-span-8 relative">
            <Label :for="`extract_regex_${idx}`">提取规则（DSL）</Label>
            <div class="mb-1 flex flex-wrap gap-1">
              <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-[11px]" @mousedown.prevent='insertJsonTemplate("findIdsByValue($, \"target\", \"id\", \"|\")")'>
                JSON: findIdsByValue
              </Button>
              <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-[11px]" @mousedown.prevent='insertJsonTemplate("findByField($, \"value\", \"target\", \"id\", \"|\")")'>
                JSON: findByField
              </Button>
              <Button type="button" size="sm" variant="ghost" class="h-6 px-2 text-[11px]" @mousedown.prevent='insertJsonTemplate("findByFieldRaw($, \"value\", \"target\", \"|\")")'>
                JSON: findByFieldRaw
              </Button>
            </div>
            <Textarea
              :id="`extract_regex_${idx}`"
              v-model="rule.regex"
              placeholder='例如：dsl:pick($.name)'
              rows="2"
              @focus="activeRuleField = 'extract'; activeExtractRuleIndex = idx"
              @blur="handleRuleBlur"
            />
            <div
              v-if="activeRuleField === 'extract' && activeExtractRuleIndex === idx && extractSuggestions.length > 0"
              class="absolute z-20 mt-1 w-full rounded-md border bg-background p-2 shadow-sm space-y-1 h-[158px] overflow-y-auto overscroll-contain"
              @mouseenter="handleSuggestionMouseEnter"
              @mouseleave="handleSuggestionMouseLeave"
              @wheel.stop
            >
              <button
                v-for="item in extractSuggestions"
                :key="`e-${idx}-${item.name}`"
                type="button"
                class="w-full text-left rounded px-2 py-1.5 hover:bg-muted transition-colors"
                @mousedown.prevent="applySuggestion('extract', item.snippet)"
              >
                <div class="text-xs font-medium">{{ item.name }}</div>
                <div class="text-[11px] text-muted-foreground">{{ item.tip }}：{{ item.snippet }}</div>
              </button>
            </div>
          </div>
        </div>
      </div>
      <p class="text-xs text-muted-foreground leading-5">
        每个字段组定义一个变量（字段名 + 提取规则）。建议将接收者字段命名为 to_user（企业微信应用定向发送）
      </p>
      <p v-if="extractSyntaxError" class="text-xs text-destructive">{{ extractSyntaxError }}</p>
            </div>

            <p v-if="isSyntaxChecking" class="text-xs text-muted-foreground">规则语法校验中...</p>
          </div>
        </section>

        <!-- 正则测试区域 -->
        <section v-show="activeSection === 'test'" id="sub-section-test" class="border rounded-lg p-4 bg-muted/30">
      <h4 class="text-sm font-medium mb-3">正则测试</h4>
      <div class="space-y-3">
      <div class="flex items-center justify-between mb-3">
        <span class="text-xs text-muted-foreground">输入样例并用后端规则测试</span>
        <Button type="button" variant="outline" size="sm" :disabled="isTestingRegex" @click="runRegexTest">
          {{ isTestingRegex ? '测试中...' : '测试' }}
        </Button>
      </div>
      <div class="space-y-3">
        <div class="space-y-2">
          <Label for="test_message" class="text-xs">测试消息内容</Label>
          <Textarea
            id="test_message"
            v-model="testMessage"
            placeholder='粘贴示例消息内容进行测试，例如：{"order_id":"12345","status":"created"}'
            rows="3"
          />
        </div>

        <!-- 测试结果 -->
        <div v-if="testResult" class="space-y-2">
          <div v-if="testResult.error" class="text-sm text-destructive">
            错误: {{ testResult.error }}
          </div>
          <template v-else>
            <div class="flex items-center gap-2 text-sm">
              <span>验证结果:</span>
              <Badge v-if="testResult.validateMatched === true" variant="default" class="bg-green-500">匹配</Badge>
              <Badge v-else-if="testResult.validateMatched === false" variant="destructive">不匹配</Badge>
              <span v-else class="text-muted-foreground">未设置验证正则</span>
            </div>
            <div v-if="extractedEntries.length > 0" class="text-sm">
              <span class="text-muted-foreground">提取结果:</span>
              <div class="mt-2 space-y-2">
                <div
                  v-for="([key, value], idx) in extractedEntries"
                  :key="`${key}-${idx}`"
                  class="flex items-center gap-2 p-2 bg-background rounded border"
                >
                  <Badge variant="outline">{{ key }}</Badge>
                  <span class="text-xs text-muted-foreground">=</span>
                  <span class="text-xs font-mono break-all">{{ value }}</span>
                </div>
              </div>
            </div>
            <div v-if="dynamicRecipientPreview.length > 0" class="text-sm">
              <span class="text-muted-foreground">动态接收者预览:</span>
              <div class="mt-2 flex flex-wrap gap-2">
                <Badge v-for="(recipient, idx) in dynamicRecipientPreview" :key="`recipient-${idx}`" variant="secondary">
                  {{ recipient }}
                </Badge>
              </div>
              <p class="mt-1 text-xs text-muted-foreground">
                仅当提取字段名为 to_user 且企业微信应用开启动态接收模式时，按以上接收者定向推送
              </p>
            </div>
            <div v-else-if="formData.extract_rules.some(r => (r.field || '').trim() !== '' || (r.regex || '').trim() !== '')" class="text-sm text-muted-foreground">
              提取结果: 未匹配到内容
            </div>
          </template>
        </div>
      </div>
      </div>
        </section>

        <section v-show="activeSection === 'template'" id="sub-section-template" class="space-y-2">
      <h4 class="text-sm font-semibold text-muted-foreground">模板配置</h4>
      <div class="space-y-2">
      <Label>模板内容格式 <span class="text-destructive">*</span></Label>
      <div class="inline-flex rounded-lg border bg-muted/30 p-1 gap-1">
        <button
          type="button"
          class="px-4 py-1.5 text-sm rounded-md transition-colors border"
          :class="formData.template_content_type === 'text'
            ? 'bg-primary text-primary-foreground border-primary shadow-sm'
            : 'bg-background text-muted-foreground border-transparent hover:text-foreground'"
          @click="formData.template_content_type = 'text'"
        >
          Text
        </button>
        <button
          type="button"
          class="px-4 py-1.5 text-sm rounded-md transition-colors border"
          :class="formData.template_content_type === 'html'
            ? 'bg-primary text-primary-foreground border-primary shadow-sm'
            : 'bg-background text-muted-foreground border-transparent hover:text-foreground'"
          @click="formData.template_content_type = 'html'"
        >
          HTML
        </button>
        <button
          type="button"
          class="px-4 py-1.5 text-sm rounded-md transition-colors border"
          :class="formData.template_content_type === 'markdown'
            ? 'bg-primary text-primary-foreground border-primary shadow-sm'
            : 'bg-background text-muted-foreground border-transparent hover:text-foreground'"
          @click="formData.template_content_type = 'markdown'"
        >
          Markdown
        </button>
      </div>
      <p class="text-xs text-muted-foreground">
        订阅发送时按所选格式渲染模板内容
      </p>

      <Label for="template_id">消息模板 <span class="text-destructive">*</span></Label>
      <Select v-model="formData.template_id">
        <SelectTrigger>
          <SelectValue placeholder="选择消息模板" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem v-for="opt in templateOptions" :key="opt.id" :value="opt.id">
              {{ opt.name }}
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
      <p class="text-xs text-muted-foreground">
        选择用于发送消息的模板，支持 ${variable} 变量替换
      </p>
      </div>
        </section>
      </div>
    </div>

    <Dialog v-model:open="isDeleteRuleDialogOpen">
      <DialogContent class="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>确认删除字段组</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <p class="text-sm text-muted-foreground">
            请输入字段名
            <span class="font-medium text-foreground">
              {{ pendingDeleteRuleIndex !== null ? formData.extract_rules[pendingDeleteRuleIndex]?.field || '(空字段名)' : '' }}
            </span>
            以确认删除。
          </p>
          <Input
            v-model="deleteRuleConfirmInput"
            placeholder="请输入字段名确认删除"
            :disabled="pendingDeleteRuleIndex === null || !(formData.extract_rules[pendingDeleteRuleIndex]?.field || '').trim()"
          />
          <p
            v-if="deleteRuleMatchStatusText"
            class="text-xs"
            :class="canConfirmDeleteRule ? 'text-emerald-600' : 'text-muted-foreground'"
          >
            {{ deleteRuleMatchStatusText }}
          </p>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="resetDeleteRuleDialog">取消</Button>
          <Button type="button" variant="destructive" :disabled="!canConfirmDeleteRule" @click="confirmRemoveExtractRule">
            确认删除
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <div class="shrink-0 border-t px-4 py-3 bg-background flex justify-end gap-2">
      <Button type="submit" :disabled="isSubmitting">
        {{ isSubmitting ? '提交中...' : (isEdit ? '保存' : '创建') }}
      </Button>
    </div>
  </form>
</template>
