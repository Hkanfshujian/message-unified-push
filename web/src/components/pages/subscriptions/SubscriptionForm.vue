<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { subscriptionsApi } from '@/api/subscriptions'
import { notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'

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
    status?: string
  } | null
  sourceOptions: MQSourceOption[]
  templateOptions: TemplateOption[]
}

const props = withDefaults(defineProps<Props>(), {
  data: null
})

const emit = defineEmits<{
  success: []
  cancel: []
}>()

const isEdit = computed(() => !!props.data)
const isRunningEdit = computed(() => props.data?.status === 'running')
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
    notifyWarning('请输入测试消息内容')
    return
  }

  testResult.value = {
    validateMatched: null,
    extractResult: null,
    error: null
  }

  isTestingRegex.value = true
  try {
    const rsp = await subscriptionsApi.testRegex({
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
    await subscriptionsApi.testRegex({
      message: testMessage.value || '{"department":"平台研发部","name":"kanfa.hu","text":"demo"}',
      validate_regex: formData.validate_regex,
      extract_rules: normalizedRules
    })
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
    notifyWarning('请选择数据源')
    return
  }
  if (!formData.name) {
    notifyWarning('请输入订阅名称')
    return
  }
  if (!formData.topic) {
    notifyWarning('请输入 Topic')
    return
  }
  if (!formData.group_name) {
    notifyWarning('请输入 Group Name')
    return
  }
  if (!formData.template_id) {
    notifyWarning('请选择消息模板')
    return
  }
  await checkRuleSyntax()
  if (validateSyntaxError.value || extractSyntaxError.value) {
    notifyWarning('规则语法校验未通过，请先修正')
    return
  }

  isSubmitting.value = true
  try {
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

    const res = await (isEdit.value
      ? subscriptionsApi.update(props.data?.id || '', payload)
      : subscriptionsApi.create(payload))
    if (res.data.code === 200) {
      notifySuccess(isEdit.value ? '编辑成功' : '新增成功')
      emit('success')
    } else {
      notifyError(normalizeRuleErrorMessage(res?.data?.msg || '操作失败'))
    }
  } catch (error: any) {
    const msg = error?.response?.data?.msg || '操作失败'
    notifyError(normalizeRuleErrorMessage(msg))
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="subscription-form-shell">
    <div class="subscription-form-layout">
      <aside class="app-form-side-nav subscription-form-nav">
        <div class="space-y-1">
          <button
            v-for="item in sectionItems"
            :key="item.key"
            type="button"
            class="app-form-side-nav-item"
            :class="activeSection === item.key ? 'app-form-side-nav-item-active' : 'app-form-side-nav-item-idle'"
            @click="scrollToSection(item.key)"
          >
            {{ item.label }}
          </button>
        </div>
      </aside>

      <div class="subscription-form-scroll space-y-4">
        <el-select v-model="activeSection" class="subscription-form-section-select w-full" aria-label="选择配置区域">
          <el-option v-for="item in sectionItems" :key="item.key" :label="item.label" :value="item.key" />
        </el-select>
        <el-alert
          v-if="isRunningEdit"
          title="订阅运行中，仅允许调整规则和模板配置；如需修改数据源、Topic、Tag 或消费者组，请先停止订阅。"
          type="warning"
          show-icon
          :closable="false"
        />
        <section v-show="activeSection === 'basic'" id="sub-section-basic" class="app-form-section space-y-4 subscription-section">
          <header class="subscription-section-head"><div><span>01</span><h4>基本信息</h4><el-tag v-if="isRunningEdit" size="small" type="success">运行中</el-tag></div><p>确定消息来源、消费主题和消费者组</p></header>
          <div class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <label for="source_id" class="text-sm font-medium">数据源 <span class="text-destructive">*</span></label>
        <el-select id="source_id" v-model="formData.source_id" class="w-full" filterable placeholder="选择数据源" :disabled="isRunningEdit">
          <el-option v-for="opt in sourceOptions" :key="opt.id" :label="opt.name" :value="opt.id" />
        </el-select>
      </div>

      <div class="space-y-2">
        <label for="name" class="text-sm font-medium">订阅名称 <span class="text-destructive">*</span></label>
        <el-input
          id="name"
          v-model="formData.name"
          placeholder="例如：订单异常告警订阅"
          maxlength="200"
          :disabled="isRunningEdit"
        />
      </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
              <label for="topic" class="text-sm font-medium">Topic <span class="text-destructive">*</span></label>
              <el-input
                id="topic"
                v-model="formData.topic"
                placeholder="例如：ORDER_EXCEPTION"
                maxlength="200"
                :disabled="isRunningEdit"
              />
            </div>

            <div class="space-y-2">
              <label for="tag" class="text-sm font-medium">Tag</label>
              <el-input
                id="tag"
                v-model="formData.tag"
                placeholder="可选，例如：prod 或 tag1||tag2"
                maxlength="200"
                :disabled="isRunningEdit"
              />
              <p class="text-xs text-muted-foreground">
                多个 Tag 用 || 分隔，留空表示订阅全部
              </p>
            </div>
            </div>

            <div class="space-y-2">
            <label for="group_name" class="text-sm font-medium">Consumer Group <span class="text-destructive">*</span></label>
            <el-input
              id="group_name"
              v-model="formData.group_name"
              placeholder="例如：mq_consumer_group"
              maxlength="200"
              :disabled="isRunningEdit"
            />
            <p class="text-xs text-muted-foreground">
              消费者组名称，同一组内负载均衡消费
            </p>
            </div>
          </div>
        </section>

        <section v-show="activeSection === 'regex'" id="sub-section-regex" class="app-form-section space-y-4 subscription-section">
          <header class="subscription-section-head"><div><span>02</span><h4>正则配置</h4></div><p>验证消息并提取模板渲染所需字段</p></header>
          <div class="space-y-4">
            <div class="space-y-2">
      <label for="validate_regex" class="text-sm font-medium">验证正则</label>
      <div class="relative">
        <el-input
          id="validate_regex"
          v-model="formData.validate_regex"
          type="textarea"
          placeholder='可选。DSL 示例：dsl:contains($.department, "研发部") && exists($.name)'
          :rows="2"
          @focus="activeRuleField = 'validate'"
          @blur="handleRuleBlur"
        />
        <div
          v-if="validateSuggestions.length > 0"
          class="glass-suggestion-panel absolute z-20 mt-1 w-full p-2 space-y-1 h-[158px] overflow-y-auto overscroll-contain"
          @mouseenter="handleSuggestionMouseEnter"
          @mouseleave="handleSuggestionMouseLeave"
          @wheel.stop
        >
          <button
            v-for="item in validateSuggestions"
            :key="`v-${item.name}`"
            type="button"
            class="glass-suggestion-item w-full text-left px-2 py-1.5 transition-all"
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

            <div class="extract-rule-builder">
      <div class="extract-rule-toolbar">
        <div class="extract-rule-toolbar-copy">
          <div><strong>提取字段</strong><span>{{ formData.extract_rules.length }} 组</span></div>
          <p>将消息内容转换为模板变量</p>
        </div>
        <el-button type="primary" plain size="small" @click="addExtractRule">新增字段</el-button>
      </div>
      <div class="extract-rule-presets" aria-label="常用字段模板">
        <span>常用字段</span>
        <button type="button" @click="addExtractRuleTemplate('to_user', 'dsl:pick($.to_user)')">to_user</button>
        <button type="button" @click="addExtractRuleTemplate('name', 'dsl:pick($.name)')">name</button>
        <button type="button" @click="addExtractRuleTemplate('text', 'dsl:pick($.text)')">text</button>
        <small>点击后新增对应字段</small>
      </div>
      <div class="extract-rule-list">
      <article
        v-for="(rule, idx) in formData.extract_rules"
        :key="`extract-rule-${idx}`"
        class="extract-rule-card"
      >
        <header class="extract-rule-card-head">
          <div class="extract-rule-order"><span>{{ String(idx + 1).padStart(2, '0') }}</span><strong>{{ rule.field || '未命名字段' }}</strong></div>
          <div class="extract-rule-actions" aria-label="字段组操作">
            <button type="button" :disabled="idx === 0" title="上移" @click="moveExtractRule(idx, -1)">↑</button>
            <button type="button" :disabled="idx === formData.extract_rules.length - 1" title="下移" @click="moveExtractRule(idx, 1)">↓</button>
            <button type="button" @click="copyExtractRule(idx)">复制</button>
            <button type="button" class="is-danger" @click="requestRemoveExtractRule(idx)">删除</button>
          </div>
        </header>
        <div class="extract-rule-fields">
          <div class="extract-rule-field-name">
            <label :for="`extract_field_${idx}`">字段名称</label>
            <el-input
              :id="`extract_field_${idx}`"
              v-model="rule.field"
              placeholder="例如：to_user"
            />
            <span>模板中通过 ${字段名} 引用</span>
          </div>
          <div class="extract-rule-expression">
            <div class="extract-rule-expression-head">
              <label :for="`extract_regex_${idx}`">提取表达式</label>
              <div class="extract-rule-json-presets">
                <span>JSON 模板</span>
                <button type="button" @mousedown.prevent='insertJsonTemplate("findIdsByValue($, \"target\", \"id\", \"|\")")'>按值查 ID</button>
                <button type="button" @mousedown.prevent='insertJsonTemplate("findByField($, \"value\", \"target\", \"id\", \"|\")")'>按字段取值</button>
                <button type="button" @mousedown.prevent='insertJsonTemplate("findByFieldRaw($, \"value\", \"target\", \"|\")")'>返回原始对象</button>
              </div>
            </div>
            <div class="relative">
            <el-input
              :id="`extract_regex_${idx}`"
              v-model="rule.regex"
              type="textarea"
              placeholder='例如：dsl:pick($.name)'
              :rows="3"
              @focus="activeRuleField = 'extract'; activeExtractRuleIndex = idx"
              @blur="handleRuleBlur"
            />
            <div
              v-if="activeRuleField === 'extract' && activeExtractRuleIndex === idx && extractSuggestions.length > 0"
              class="glass-suggestion-panel absolute z-20 mt-1 w-full p-2 space-y-1 h-[158px] overflow-y-auto overscroll-contain"
              @mouseenter="handleSuggestionMouseEnter"
              @mouseleave="handleSuggestionMouseLeave"
              @wheel.stop
            >
              <button
                v-for="item in extractSuggestions"
                :key="`e-${idx}-${item.name}`"
                type="button"
                class="glass-suggestion-item w-full text-left px-2 py-1.5 transition-all"
                @mousedown.prevent="applySuggestion('extract', item.snippet)"
              >
                <div class="text-xs font-medium">{{ item.name }}</div>
                <div class="text-[11px] text-muted-foreground">{{ item.tip }}：{{ item.snippet }}</div>
              </button>
            </div>
          </div>
        </div>
        </div>
      </article>
      </div>
      <p class="extract-rule-help">
        每个字段对应一个模板变量；企业微信应用的动态接收者字段请使用 <code>to_user</code>
      </p>
      <p v-if="extractSyntaxError" class="text-xs text-destructive">{{ extractSyntaxError }}</p>
            </div>

            <p v-if="isSyntaxChecking" class="text-xs text-muted-foreground">规则语法校验中...</p>
          </div>
        </section>

        <!-- 正则测试区域 -->
        <section v-show="activeSection === 'test'" id="sub-section-test" class="app-form-section subscription-section">
      <header class="subscription-section-head"><div><span>03</span><h4>正则测试</h4></div><p>使用样例消息验证匹配与字段提取结果</p></header>
      <div class="subscription-test-body space-y-3">
      <div class="flex items-center justify-between mb-3">
        <span class="text-xs text-muted-foreground">输入样例并用后端规则测试</span>
        <el-button plain size="small" :loading="isTestingRegex" @click="runRegexTest">
          {{ isTestingRegex ? '测试中...' : '测试' }}
        </el-button>
      </div>
      <div class="space-y-3">
        <div class="space-y-2">
          <label for="test_message" class="text-xs font-medium">测试消息内容</label>
          <el-input
            id="test_message"
            v-model="testMessage"
            type="textarea"
            placeholder='粘贴示例消息内容进行测试，例如：{"order_id":"12345","status":"created"}'
            :rows="3"
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
              <el-tag v-if="testResult.validateMatched === true" type="success">匹配</el-tag>
              <el-tag v-else-if="testResult.validateMatched === false" type="danger">不匹配</el-tag>
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
                  <el-tag effect="plain">{{ key }}</el-tag>
                  <span class="text-xs text-muted-foreground">=</span>
                  <span class="text-xs font-mono break-all">{{ value }}</span>
                </div>
              </div>
            </div>
            <div v-if="dynamicRecipientPreview.length > 0" class="text-sm">
              <span class="text-muted-foreground">动态接收者预览:</span>
              <div class="mt-2 flex flex-wrap gap-2">
                <el-tag v-for="(recipient, idx) in dynamicRecipientPreview" :key="`recipient-${idx}`" type="info" effect="light">
                  {{ recipient }}
                </el-tag>
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

        <section v-show="activeSection === 'template'" id="sub-section-template" class="app-form-section space-y-2 subscription-section">
      <header class="subscription-section-head"><div><span>04</span><h4>模板配置</h4></div><p>选择渲染格式和最终发送模板</p></header>
      <div class="space-y-2">
      <label class="text-sm font-medium">模板内容格式 <span class="text-destructive">*</span></label>
      <el-segmented v-model="formData.template_content_type" :options="['text', 'html', 'markdown']" />
      <p class="text-xs text-muted-foreground">
        订阅发送时按所选格式渲染模板内容
      </p>

      <label for="template_id" class="text-sm font-medium">消息模板 <span class="text-destructive">*</span></label>
      <el-select id="template_id" v-model="formData.template_id" class="w-full" filterable placeholder="选择消息模板">
        <el-option v-for="opt in templateOptions" :key="opt.id" :label="opt.name" :value="opt.id" />
      </el-select>
      <p class="text-xs text-muted-foreground">
        选择用于发送消息的模板，支持 ${variable} 变量替换
      </p>
      </div>
        </section>
      </div>
    </div>

    <el-dialog v-model="isDeleteRuleDialogOpen" title="确认删除字段组" width="min(480px, calc(100vw - 24px))" class="app-nested-dialog" append-to-body>
        <div class="space-y-3">
          <p class="text-sm text-muted-foreground">
            请输入字段名
            <span class="font-medium text-foreground">
              {{ pendingDeleteRuleIndex !== null ? formData.extract_rules[pendingDeleteRuleIndex]?.field || '(空字段名)' : '' }}
            </span>
            以确认删除。
          </p>
          <el-input
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
        <template #footer>
          <el-button @click="resetDeleteRuleDialog">取消</el-button>
          <el-button type="danger" :disabled="!canConfirmDeleteRule" @click="confirmRemoveExtractRule">
            确认删除
          </el-button>
        </template>
    </el-dialog>

    <div class="app-form-actions shrink-0 py-2.5">
      <el-button type="primary" native-type="submit" :loading="isSubmitting">
        {{ isSubmitting ? '提交中...' : (isEdit ? '保存' : '创建') }}
      </el-button>
    </div>
  </form>
</template>

<style scoped>
.subscription-section { border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.subscription-form-shell { display: flex; width: 100%; height: 100%; min-height: 0; flex-direction: column; }
.subscription-form-layout { display: grid; grid-template-columns: 190px minmax(0, 1fr); min-height: 0; flex: 1; }
.subscription-form-nav { display: block; overflow: hidden; }
.subscription-form-scroll { min-height: 0; padding: 10px 12px; overflow-y: auto; overscroll-behavior: contain; }
.subscription-form-section-select { display: none; }
.subscription-section-head p { margin: 4px 0 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.subscription-section { padding: 0 !important; }
.subscription-section > :not(.subscription-section-head) { margin-inline: 14px; }
.subscription-section-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 38px; padding: 0 12px; border-bottom: 1px solid var(--app-overlay-border); }
.subscription-section-head > div { display: flex; align-items: center; gap: 8px; }
.subscription-section-head span { color: var(--brand-600); font: 800 10px monospace; }
.subscription-section-head h4 { margin: 0; font-size: 13px; }
.subscription-section-head p { margin: 0; }
.subscription-test-body { padding-block: 14px; }
.extract-rule-builder { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 10px; background: var(--app-overlay-surface); }
.extract-rule-toolbar { display: flex; min-height: 58px; align-items: center; justify-content: space-between; gap: 16px; padding: 10px 14px; border-bottom: 1px solid var(--app-overlay-border); }
.extract-rule-toolbar-copy > div { display: flex; align-items: center; gap: 8px; }
.extract-rule-toolbar-copy strong { font-size: 13px; }
.extract-rule-toolbar-copy span { border-radius: 999px; padding: 2px 7px; background: color-mix(in srgb, var(--brand-500) 10%, transparent); color: var(--brand-700); font-size: 10px; font-weight: 700; }
.extract-rule-toolbar-copy p { margin: 3px 0 0; color: var(--admin-text-muted); font-size: 11px; }
.extract-rule-presets { display: flex; min-height: 42px; align-items: center; gap: 6px; padding: 6px 14px; border-bottom: 1px solid var(--app-overlay-border); background: color-mix(in srgb, var(--app-overlay-surface) 92%, var(--brand-50)); }
.extract-rule-presets > span, .extract-rule-json-presets > span { margin-right: 2px; color: var(--admin-text-muted); font-size: 11px; }
.extract-rule-presets button, .extract-rule-json-presets button { border: 1px solid color-mix(in srgb, var(--brand-500) 18%, var(--app-overlay-border)); border-radius: 6px; padding: 4px 8px; background: var(--app-overlay-surface); color: var(--admin-text-primary); cursor: pointer; font-size: 11px; line-height: 1.35; }
.extract-rule-presets button:hover, .extract-rule-json-presets button:hover { border-color: color-mix(in srgb, var(--brand-500) 52%, var(--app-overlay-border)); color: var(--brand-700); }
.extract-rule-presets small { margin-left: auto; color: var(--admin-text-muted); font-size: 10px; }
.extract-rule-list { display: grid; gap: 10px; padding: 12px; }
.extract-rule-card { overflow: hidden; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.extract-rule-card-head { display: flex; min-height: 42px; align-items: center; justify-content: space-between; gap: 12px; padding: 6px 10px 6px 12px; border-bottom: 1px solid var(--app-overlay-border); background: var(--admin-surface-muted); }
.extract-rule-order { display: flex; min-width: 0; align-items: center; gap: 8px; }
.extract-rule-order span { color: var(--brand-600); font: 800 10px monospace; }
.extract-rule-order strong { min-width: 0; overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.extract-rule-actions { display: flex; align-items: center; gap: 2px; }
.extract-rule-actions button { min-width: 30px; height: 28px; border: 0; border-radius: 6px; padding: 0 7px; background: transparent; color: var(--admin-text-muted); cursor: pointer; font-size: 11px; }
.extract-rule-actions button:hover:not(:disabled) { background: var(--app-overlay-surface); color: var(--brand-700); }
.extract-rule-actions button.is-danger:hover { background: color-mix(in srgb, var(--el-color-danger) 10%, transparent); color: var(--el-color-danger); }
.extract-rule-actions button:disabled { cursor: not-allowed; opacity: .35; }
.extract-rule-fields { display: grid; grid-template-columns: minmax(180px, .7fr) minmax(0, 1.7fr); gap: 14px; padding: 12px; }
.extract-rule-field-name, .extract-rule-expression { min-width: 0; }
.extract-rule-field-name { display: grid; align-content: start; gap: 6px; }
.extract-rule-fields label { color: var(--admin-text-primary); font-size: 12px; font-weight: 650; }
.extract-rule-field-name > span { color: var(--admin-text-muted); font-size: 10px; }
.extract-rule-expression-head { display: flex; min-height: 24px; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 6px; }
.extract-rule-json-presets { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 4px; }
.extract-rule-json-presets button { padding: 3px 6px; }
.extract-rule-help { margin: 0; padding: 0 14px 12px; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
.extract-rule-help code { color: var(--brand-700); font-size: 10px; }
.subscription-actions { display: flex; align-items: center; justify-content: space-between; gap: 16px; min-height: 58px; padding: 10px 16px; border-top: 1px solid var(--app-overlay-border); background: var(--app-overlay-surface); }
.subscription-actions > span { color: var(--admin-text-muted); font-size: 11px; }
.subscription-actions > div { display: flex; gap: 8px; }
@container app-managed-drawer (max-width: 760px) {
  .subscription-form-layout { grid-template-columns: 1fr; }
  .subscription-form-nav { display: none; }
  .subscription-form-section-select { display: flex; }
  .subscription-form-scroll { padding: 10px 12px; }
  .subscription-section-head { align-items: center; flex-direction: row; }
  .subscription-actions, .extract-rule-toolbar { align-items: flex-start; flex-direction: column; }
  .subscription-section-head p, .subscription-actions > span, .extract-rule-presets small { display: none; }
  .subscription-actions > div { width: 100%; }
  .subscription-actions :deep(.el-button) { flex: 1; }
  .extract-rule-presets { align-items: flex-start; flex-wrap: wrap; }
  .extract-rule-presets > span { width: 100%; }
  .extract-rule-card-head { align-items: flex-start; flex-direction: column; }
  .extract-rule-actions { width: 100%; justify-content: flex-end; }
  .extract-rule-fields { grid-template-columns: 1fr; }
  .extract-rule-expression-head { align-items: flex-start; flex-direction: column; }
  .extract-rule-json-presets { justify-content: flex-start; }
}
</style>
