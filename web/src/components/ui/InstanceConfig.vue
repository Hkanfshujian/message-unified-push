<script setup lang="ts">
import { ref, computed, watch, reactive } from 'vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import { CONSTANT } from '@/constant'
import { request } from '@/api/api'
import { useRbacStore } from '@/store'
import { generateBizUniqueID } from '@/util/uuid'
import { confirmAction, notifyError, notifySuccess } from '@/util/uiFeedback'

// 组件props
interface Props {
  // 关联的数据（模板数据）
  data: any
  // 是否在对话框中显示（用于模板）
  inDialog?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  inDialog: false
})
const rbacStore = useRbacStore()

// API 配置映射
const apiConfig = computed(() => {
  return {
    addIns: '/templates/ins/addone',
    getIns: '/templates/ins/get',
    deleteIns: '/templates/ins/delete',
    updateEnable: '/templates/ins/update_enable',
    updateConfig: '/templates/ins/update_config',
    idField: 'template_id',
    nameField: 'name'
  }
})

// 前端的页面添加配置
const waysConfigMap = CONSTANT.WAYS_DATA

// 渠道列表筛选与分页
const channelFilters = reactive({
  name: '',
  type: 'all'
})
const channelState = reactive({
  list: [] as Array<{ id: string, name: string, type: string; created_on?: string }>,
  total: 0,
  currPage: 1,
  pageSize: 10,
  loading: false
})
const selectedChannel = ref<{ id: string, name: string, type: string } | null>(null)

const channelTypeOptions = computed(() => {
  const options = [{ value: 'all', label: '根据类型筛选' }]
  waysConfigMap.forEach((item: any) => {
    options.push({ value: item.type, label: item.label })
  })
  return options
})

const channelTypeLabelMap = computed(() => {
  const map = new Map<string, string>()
  waysConfigMap.forEach((item: any) => {
    map.set(item.type, item.label)
  })
  return map
})

const formatChannelTypeLabel = (type?: string) => {
  if (!type) {
    return '-'
  }
  return channelTypeLabelMap.value.get(type) || type
}

// 当前选中渠道的配置
const currentChannelConfig = computed(() => {
  const type = selectedChannel.value?.type
  // 再根据type找到配置
  return waysConfigMap.find((item: any) => item.type === type) || null
})

// 表单数据
const formData = ref<Record<string, any>>({
  allowMultiRecip: false  // 默认false为固定模式，true为动态模式
})

const currentDynamicRecipient = computed(() => {
  if (currentChannelConfig.value?.dynamicRecipient?.support) {
    return currentChannelConfig.value.dynamicRecipient
  }
  // 兼容企业微信应用：constant.js 未配置 dynamicRecipient
  if (selectedChannel.value?.type === 'QyWeiXinApp') {
    return {
      support: true,
      field: 'to_user',
      label: '固定接收者（to_user）',
      desc: 'to_user（多个接收者用 | 分隔）'
    }
  }
  return null
})
const canManageTemplateInstance = computed(() => rbacStore.hasPermission('message:template:instance'))

const currentDynamicRecipientField = computed(() => currentDynamicRecipient.value?.field || '')

// 是否显示接收者输入框
const shouldShowRecipientInput = computed(() => {
  // 支持动态接收者 且 未勾选（固定模式）时显示输入框
  return !!currentDynamicRecipient.value && !formData.value.allowMultiRecip
})

const isQyWeiXinAppSelected = computed(() => selectedChannel.value?.type === 'QyWeiXinApp')

const isQyWeiXinFixedRecipientEmpty = computed(() => {
  if (!isQyWeiXinAppSelected.value) return false
  if (formData.value.allowMultiRecip) return false
  const field = currentDynamicRecipient.value?.field
  if (!field) return false
  return !String(formData.value[field] || '').trim()
})

// 选择渠道
const handleSelectChannel = (channel: { id: string, name: string, type: string }) => {
  selectedChannel.value = channel
  // 数据加载后，text/html单选设置默认选中（这里选第一个）
  if (currentChannelConfig.value?.taskInsRadios.length > 0) {
    formData.value.templ_type = currentChannelConfig.value?.taskInsRadios[0].subLabel
  }
  // 重置动态接收者设置：企业微信应用首次添加默认开启动态接收者
  if (channel.type === 'QyWeiXinApp') {
    const hasQyWeiXinAppInstance = insTableData.value.some(ins => ins.way_type === 'QyWeiXinApp')
    formData.value.allowMultiRecip = !hasQyWeiXinAppInstance
  } else {
    formData.value.allowMultiRecip = false
  }
}

const isChannelChecked = (channelId: string) => {
  return selectedChannel.value?.id === channelId
}

const handleChannelCheck = (channel: { id: string, name: string, type: string }, checked: boolean) => {
  if (checked) {
    handleSelectChannel(channel)
    return
  }
  if (selectedChannel.value?.id === channel.id) {
    selectedChannel.value = null
    formData.value.allowMultiRecip = false
  }
}

// 添加单条实例配置
const handleAddSubmit = async () => {
  // 验证是否选择了渠道
  if (!selectedChannel.value) {
    notifyError('请选择发送渠道')
    return
  }

  // 检查动态接收和固定接收不能混合使用
  if (insTableData.value.length > 0) {
    const hasDynamicInstance = insTableData.value.some(ins => {
      try {
        const config = JSON.parse(ins.config)
        return config.allowMultiRecip === true
      } catch {
        return false
      }
    })
    
    const entityName = '模板'
    
    // 如果要添加动态接收实例，但已有其他实例
    if (formData.value.allowMultiRecip === true) {
      if (hasDynamicInstance) {
        notifyError(`该${entityName}已存在动态接收实例，一个${entityName}只能配置一个动态接收实例`)
        return
      }
      if (insTableData.value.length > 0) {
        notifyError(`动态接收实例不能与固定接收实例混合使用，请先删除所有固定实例`)
        return
      }
    }
    
    // 如果要添加固定接收实例，但已有动态接收实例
    if (formData.value.allowMultiRecip !== true && hasDynamicInstance) {
      notifyError(`该${entityName}已配置动态接收实例，不能再添加固定接收实例`)
      return
    }
  }

  // 验证内容类型
  const contentType = formData.value.templ_type
  if (!contentType) {
    notifyError('请选择消息格式')
    return
  }

  // 仅模板需要验证对应格式的内容是否为空
  const templateFieldMap: Record<string, string> = {
    'text': 'text_template',
    'html': 'html_template',
    'markdown': 'markdown_template'
  }
  
  const fieldName = templateFieldMap[contentType.toLowerCase()]
  if (fieldName) {
    const templateContent = props.data?.[fieldName] || ''
    // 检查是否为空（去除所有空白字符后检查）
    if (!templateContent.trim()) {
      notifyError(`模板的 ${contentType} 格式内容为空，无法添加此类型的实例`)
      return
    }
  }

  // 组建表单数据
  let postData: Record<string, any> = {
    "id": generateBizUniqueID('IN'),
    "enable": 1,
    [apiConfig.value.idField]: props.data.id,
    "way_id": selectedChannel.value.id,
    "way_type": selectedChannel.value.type,
    "way_name": selectedChannel.value.name,
    "content_type": formData.value.templ_type,
    "config": JSON.stringify(formData.value),
  }

  try {
    const response = await request.post(apiConfig.value.addIns, postData, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (response.status === 200 && response.data.code === 200) {
      notifySuccess(response.data.msg)
      // 重新加载实例列表
      await queryInsListData()
      // 清空表单
      selectedChannel.value = null
      formData.value = { allowMultiRecip: false }
    } else {
      notifyError(response.data.msg || '添加实例失败')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '添加实例失败')
  }
}

// 实例表格数据
const insTableData = ref<any[]>([])

// 格式化额外信息列的值
const formatInsConfigDisplay = (row: any) => {
  if (!row.config) {
    return "-"
  }
  let config = JSON.parse(row.config)
  
  // 检查是否为动态接收者模式
  if (config.allowMultiRecip === true) {
    return "动态接收"
  }
  
  // 固定模式，根据 constant.js 配置动态获取接收者字段
  const channelConfig = CONSTANT.WAYS_DATA.find((item: any) => item.type === row.way_type)
  let recipientField = channelConfig?.dynamicRecipient?.field || ''
  // 兼容企业微信应用：constant.js 未配置 dynamicRecipient 时，固定使用 to_user
  if (!recipientField && row.way_type === 'QyWeiXinApp') {
    recipientField = 'to_user'
  }
  if (recipientField) {
    return config[recipientField] || ""
  }

  if (channelConfig?.taskInsInputs && Array.isArray(channelConfig.taskInsInputs) && channelConfig.taskInsInputs.length === 0) {
    return "无需配置"
  }
  return ""
}

const isDynamicRecipientEnabled = (row: any) => {
  if (!row?.config) return false
  try {
    const config = JSON.parse(row.config)
    return config.allowMultiRecip === true
  } catch {
    return false
  }
}

const dynamicConfirmOpen = ref(false)
const dynamicToggleTarget = ref<any | null>(null)
const dynamicToggleNext = ref(false)
const dynamicRecipientInput = ref('')

const getChannelConfigByWayType = (wayType: string) => {
  return waysConfigMap.find((item: any) => item.type === wayType) || null
}

const getDynamicRecipientField = (ins: any) => {
  const channelConfig = getChannelConfigByWayType(ins?.way_type)
  if (channelConfig?.dynamicRecipient?.field) {
    return channelConfig.dynamicRecipient.field
  }
  // 兼容企业微信应用：constant.js 未配置 dynamicRecipient 时，固定使用 to_user
  if (ins?.way_type === 'QyWeiXinApp') {
    return 'to_user'
  }
  return ''
}

const shouldRequireRecipientInput = computed(() => {
  const ins = dynamicToggleTarget.value
  if (!ins) return false
  if (dynamicToggleNext.value) return false
  return !!getDynamicRecipientField(ins)
})

const handleToggleDynamicRecipient = (ins: any) => {
  const current = isDynamicRecipientEnabled(ins)
  let config: Record<string, any> = {}
  try {
    config = ins?.config ? JSON.parse(ins.config) : {}
  } catch {
    config = {}
  }
  const recipientField = getDynamicRecipientField(ins)
  dynamicRecipientInput.value = String(config[recipientField] || '').trim()
  dynamicToggleTarget.value = ins
  dynamicToggleNext.value = !current
  dynamicConfirmOpen.value = true
}

const cancelToggleDynamicRecipient = () => {
  dynamicConfirmOpen.value = false
  dynamicToggleTarget.value = null
  dynamicRecipientInput.value = ''
}

const confirmToggleDynamicRecipient = async () => {
  const ins = dynamicToggleTarget.value
  if (!ins) return
  const next = dynamicToggleNext.value

  let config: Record<string, any> = {}
  try {
    config = ins?.config ? JSON.parse(ins.config) : {}
  } catch {
    config = {}
  }
  const recipientField = getDynamicRecipientField(ins)
  if (!next && recipientField) {
    const recipient = dynamicRecipientInput.value.trim()
    if (!recipient) {
      notifyError(`关闭动态接收者前请先填写固定接收者（${recipientField}）`)
      return
    }
    config[recipientField] = recipient
  }
  config.allowMultiRecip = next

  try {
    const response = await request.post(apiConfig.value.updateConfig, {
      id: ins.id,
      way_type: ins.way_type,
      config: JSON.stringify(config)
    }, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (response.status === 200 && response.data.code === 200) {
      notifySuccess(response.data.msg || '更新成功')
      await queryInsListData()
      dynamicConfirmOpen.value = false
      dynamicToggleTarget.value = null
      dynamicRecipientInput.value = ''
    } else {
      notifyError(response.data.msg || '更新失败')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '更新失败')
  }
}

// 查询实例列表数据
const queryInsListData = async () => {
  if (!props.data?.id) return
  
  try {
    const response = await request.get(apiConfig.value.getIns, {
      params: { id: props.data.id },
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (response.status === 200 && response.data.code === 200) {
      // 模板返回 ins_list，任务返回 ins_data
      const insList = response.data.data.ins_list || response.data.data.ins_data || []
      insTableData.value = insList
    }
  } catch (error) {
    console.error('获取实例列表失败', error)
  }
}

// 删除实例
const handleDeleteIns = async (ins: any) => {
  await confirmAction(`确认删除实例“${ins?.way_name || ins?.id || '未命名'}”？该操作不可恢复。`, '删除实例')
  try {
    const response = await request.post(apiConfig.value.deleteIns, { id: ins.id }, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (response.status === 200 && response.data.code === 200) {
      notifySuccess(response.data.msg)
      await queryInsListData()
    } else {
      notifyError(response.data.msg || '删除失败')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '删除失败')
  }
}

// 切换实例启用状态
const handleToggleEnable = async (insId: string, currentStatus: number | string) => {
  const isEnabled = Number(currentStatus) === 1
  const newStatus = isEnabled ? 0 : 1
  
  // 立即更新本地状态，提供即时反馈
  const insIndex = insTableData.value.findIndex(ins => ins.id === insId)
  if (insIndex !== -1) {
    insTableData.value[insIndex].enable = newStatus
  }
  
  try {
    const response = await request.post(apiConfig.value.updateEnable, {
      ins_id: insId,
      status: newStatus
    }, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    
    if (response.status === 200 && response.data.code === 200) {
      notifySuccess(response.data.msg)
      // 重新加载确保数据同步
      await queryInsListData()
    } else {
      notifyError(response.data.msg || '更新失败')
      // 失败时恢复原状态
      if (insIndex !== -1) {
        insTableData.value[insIndex].enable = currentStatus
      }
    }
  } catch (error: any) {
    console.error('状态切换失败:', error)
    notifyError(error.response?.data?.msg || '更新失败')
    // 失败时恢复原状态
    if (insIndex !== -1) {
      insTableData.value[insIndex].enable = currentStatus
    }
  }
}

const queryChannelList = async (page = 1) => {
  channelState.loading = true
  try {
    const response = await request.get('/sendways/list', {
      params: {
        page,
        size: channelState.pageSize,
        name: channelFilters.name,
        type: channelFilters.type === 'all' ? '' : channelFilters.type
      },
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (response.status === 200 && response.data.code === 200) {
      channelState.list = response.data.data?.lists || []
      channelState.total = response.data.data?.total || 0
      channelState.currPage = page
    }
  } catch (error) {
    notifyError('获取渠道列表失败')
  } finally {
    channelState.loading = false
  }
}

const handleChannelPaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  channelState.currPage = page
  channelState.pageSize = pageSize
  await queryChannelList(page)
}

const handleChannelSearch = async () => {
  await queryChannelList(1)
}

// 监听数据变化，自动加载实例列表与渠道列表
watch(() => props.data?.id, (newVal) => {
  if (newVal) {
    selectedChannel.value = null
    queryInsListData()
    queryChannelList(1)
  }
}, { immediate: true })

// 暴露方法供父组件调用
defineExpose({
  queryInsListData
})
</script>

<template>
  <div class="instance-config" :class="{ 'is-in-dialog': inDialog }">
    <section v-if="data" class="instance-config-section instance-config-summary">
      <div class="instance-config-summary-main">
        <div class="instance-config-eyebrow">当前模板</div>
        <div class="instance-config-summary-title">
          <h3>{{ data[apiConfig.nameField] }}</h3>
          <el-tag effect="plain" size="small">{{ data.id }}</el-tag>
        </div>
        <p>选择发送渠道并完成实例配置，添加后即可用于此模板。</p>
      </div>
      <div class="instance-config-summary-count">
        <strong>{{ insTableData.length }}</strong>
        <span>已关联实例</span>
      </div>
    </section>

    <section class="instance-config-section">
      <div class="instance-config-section-head">
        <div>
          <h3>选择发送渠道</h3>
          <p>从已有渠道中筛选并选择一个渠道。</p>
        </div>
      </div>

      <div class="instance-config-filter-grid">
        <el-input
          v-model="channelFilters.name"
          placeholder="按渠道名称搜索"
          @keyup.enter="handleChannelSearch"
        />
        <el-select v-model="channelFilters.type" @change="handleChannelSearch">
          <el-option v-for="option in channelTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button type="primary" @click="handleChannelSearch">搜索</el-button>
      </div>

      <div class="instance-config-table-shell">
        <el-table :data="channelState.list" v-loading="channelState.loading" row-key="id" @row-click="handleSelectChannel">
          <el-table-column label="选择" width="64">
            <template #default="{ row }">
              <el-checkbox :model-value="isChannelChecked(row.id)" @change="(checked: boolean) => handleChannelCheck(row, checked)" @click.stop />
            </template>
          </el-table-column>
          <el-table-column label="序号" width="72">
            <template #default="{ $index }">{{ (channelState.currPage - 1) * channelState.pageSize + $index + 1 }}</template>
          </el-table-column>
          <el-table-column label="渠道ID" prop="id" width="180" />
          <el-table-column label="渠道名称" prop="name" min-width="150" />
          <el-table-column label="渠道类型" width="150">
            <template #default="{ row }">{{ formatChannelTypeLabel(row.type) }}</template>
          </el-table-column>
          <template #empty>当前没有匹配渠道</template>
        </el-table>
      </div>

      <div class="instance-config-pagination">
        <AppPagination
          v-model:current-page="channelState.currPage"
          v-model:page-size="channelState.pageSize"
          :total="channelState.total"
          compact
          @change="handleChannelPaginationChange"
        />
      </div>
    </section>

    <section class="instance-config-section instance-config-action-section">
      <div class="instance-config-section-head">
        <div>
          <h3>添加实例</h3>
          <p>确认当前选择后，将渠道作为新的发送实例关联到模板。</p>
        </div>
      </div>
      <div class="instance-config-action-bar">
        <div class="instance-config-selection-status" :class="{ 'is-selected': selectedChannel }">
          <span class="instance-config-status-dot" />
          <span>{{ selectedChannel ? `已选择：${selectedChannel.name}（${formatChannelTypeLabel(selectedChannel.type)}）` : '尚未选择发送渠道' }}</span>
        </div>
        <el-button type="primary" :disabled="!selectedChannel" @click="handleAddSubmit">添加实例</el-button>
      </div>
    </section>

    <section v-if="currentChannelConfig" class="instance-config-section">
      <div class="instance-config-section-head">
        <div>
          <h3>渠道配置</h3>
          <p>设置接收者模式、渠道参数和消息格式。</p>
        </div>
      </div>

      <div v-if="currentDynamicRecipient?.support" class="instance-config-mode-panel">
        <div class="flex items-center gap-2">
          <el-switch
            :id="`allow-multi-${selectedChannel?.id || 'none'}`"
            :model-value="formData.allowMultiRecip"
            @update:model-value="(val: boolean) => formData.allowMultiRecip = val"
          />
          <label :for="`allow-multi-${selectedChannel?.id || 'none'}`" class="text-sm font-medium cursor-pointer">
            动态接收者模式
          </label>
        </div>
        <p>{{ formData.allowMultiRecip ? '支持动态接收者，发送时通过 API 指定接收者列表（群发模式）' : '固定接收者模式，需要在下方配置固定接收者' }}</p>
        <p v-if="formData.allowMultiRecip" class="instance-config-warning">
          一个模板只能配置一个动态接收实例，且不能与固定接收实例混合使用。
        </p>
      </div>

      <div class="instance-config-fields">
        <div v-if="shouldShowRecipientInput" class="instance-config-field-group">
          <div class="instance-config-field-title">接收者配置</div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-xs font-medium text-muted-foreground">{{ currentDynamicRecipient?.label }}</label>
              <el-input
                v-model="formData[currentDynamicRecipientField]"
                :placeholder="`请输入${currentDynamicRecipient?.desc}`"
                type="text"
              />
            </div>
          </div>
          <p v-if="isQyWeiXinFixedRecipientEmpty" class="mt-2 text-xs text-destructive">
            企业微信应用固定模式下 to_user 不能为空。为空将无法保存，且系统不会兜底为 @all。
          </p>
        </div>

        <div v-if="currentChannelConfig.taskInsInputs && currentChannelConfig.taskInsInputs.length > 0" class="instance-config-field-group">
          <div class="instance-config-field-title">实例配置</div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div
              v-for="input in currentChannelConfig.taskInsInputs.filter((inp: any) => inp.col !== currentDynamicRecipient?.field)"
              :key="input.col"
              class="space-y-2"
            >
              <label class="text-xs font-medium text-muted-foreground">{{ input.label || input.desc }}</label>
              <el-input
                v-model="formData[input.col]"
                :placeholder="input.desc || `请输入${input.label}`"
                :type="input.type || 'text'"
                class="w-full"
              />
            </div>
          </div>
        </div>

        <div v-if="currentChannelConfig.taskInsRadios && currentChannelConfig.taskInsRadios.length > 0" class="instance-config-field-group">
          <div class="instance-config-field-title">消息格式</div>
          <el-radio-group v-model="formData.templ_type" class="flex flex-wrap gap-x-4 gap-y-2">
            <el-radio v-for="radio in currentChannelConfig.taskInsRadios" :key="radio.subLabel" :value="radio.subLabel">{{ radio.subLabel }}</el-radio>
          </el-radio-group>
        </div>
      </div>
    </section>

    <section class="instance-config-section">
      <div class="instance-config-section-head">
        <div>
          <h3>已关联实例</h3>
          <p>查看和管理此模板当前关联的发送实例。</p>
        </div>
        <el-tag effect="plain" size="small">共 {{ insTableData.length }} 条</el-tag>
      </div>
      <div class="instance-config-table-shell instance-config-instance-table">
        <el-table :data="insTableData" row-key="id">
          <el-table-column label="渠道名称" min-width="150">
            <template #default="{ row: ins }">
              <div class="font-medium">{{ ins.way_name || '未命名' }}</div>
              <div class="text-xs text-muted-foreground">{{ ins.way_type }}</div>
            </template>
          </el-table-column>
          <el-table-column label="内容类型" min-width="110">
            <template #default="{ row: ins }">
              <el-tag type="info">{{ ins.content_type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="接收者" min-width="140">
            <template #default="{ row: ins }">
              <el-tag v-if="formatInsConfigDisplay(ins)" type="info">{{ formatInsConfigDisplay(ins) }}</el-tag>
              <span v-else class="text-sm text-muted-foreground">-</span>
            </template>
          </el-table-column>
          <el-table-column label="动态接收者" min-width="140">
            <template #default="{ row: ins }">
              <div class="flex items-center gap-2">
                <el-switch
                  :model-value="isDynamicRecipientEnabled(ins)"
                  :disabled="!canManageTemplateInstance"
                  @update:model-value="() => handleToggleDynamicRecipient(ins)"
                />
                <span class="text-xs" :class="isDynamicRecipientEnabled(ins) ? 'text-emerald-600' : 'text-muted-foreground'">
                  {{ isDynamicRecipientEnabled(ins) ? '开启' : '关闭' }}
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="190" align="center">
            <template #default="{ row: ins }">
              <AppRowActions :actions="[
                { key: 'stop', label: '停用', kind: 'write', permission: 'message:template:instance', visible: ins.enable === 1, danger: true, onClick: () => handleToggleEnable(ins.id, ins.enable) },
                { key: 'start', label: '启用', kind: 'write', permission: 'message:template:instance', visible: ins.enable !== 1, onClick: () => handleToggleEnable(ins.id, ins.enable) },
                { key: 'delete', label: '删除', kind: 'write', permission: 'message:template:instance', danger: true, onClick: () => handleDeleteIns(ins) }
              ]" />
            </template>
          </el-table-column>
          <template #empty>
            <AppEmptyState description="还没有配置任何实例，请先添加" />
          </template>
        </el-table>
      </div>
    </section>

    <el-dialog v-model="dynamicConfirmOpen" title="确认操作" width="420px" class="app-nested-dialog" append-to-body>
      <p class="text-sm text-muted-foreground">
        确认将动态接收者模式{{ dynamicToggleNext ? '开启' : '关闭' }}吗？
      </p>
      <div v-if="shouldRequireRecipientInput" class="space-y-2">
        <label class="text-sm">固定接收者（to_user）</label>
        <el-input v-model="dynamicRecipientInput" placeholder="关闭动态接收者前请输入固定接收者" />
        <p class="text-xs text-muted-foreground">关闭后将按该固定接收者发送</p>
      </div>
      <template #footer>
        <el-button type="button" @click="cancelToggleDynamicRecipient">取消</el-button>
        <el-button :type="dynamicToggleNext ? 'primary' : 'danger'" @click="confirmToggleDynamicRecipient">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.instance-config {
  display: grid;
  gap: 14px;
  min-width: 0;
}

.instance-config-section {
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--dora-border, var(--border));
  border-radius: var(--dora-surface-radius, 8px);
  background: var(--app-overlay-surface, var(--card));
}

.instance-config-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.instance-config-summary-main {
  min-width: 0;
}

.instance-config-eyebrow {
  margin-bottom: 4px;
  color: var(--dora-text-muted, var(--muted-foreground));
  font-size: 12px;
  font-weight: 600;
}

.instance-config-summary-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.instance-config-summary-title h3,
.instance-config-section-head h3 {
  margin: 0;
  color: var(--dora-text, var(--foreground));
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.instance-config-summary-title h3 {
  overflow: hidden;
  font-size: 17px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instance-config-summary-main p,
.instance-config-section-head p {
  margin: 4px 0 0;
  color: var(--dora-text-muted, var(--muted-foreground));
  font-size: 12px;
  line-height: 1.55;
}

.instance-config-summary-count {
  display: grid;
  flex: 0 0 auto;
  min-width: 88px;
  padding-left: 20px;
  border-left: 1px solid var(--dora-border, var(--border));
  text-align: right;
}

.instance-config-summary-count strong {
  color: var(--dora-text, var(--foreground));
  font-size: 20px;
  line-height: 1.2;
}

.instance-config-summary-count span {
  margin-top: 3px;
  color: var(--dora-text-muted, var(--muted-foreground));
  font-size: 12px;
}

.instance-config-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.instance-config-filter-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(180px, 0.8fr) auto;
  gap: 10px;
  margin-bottom: 12px;
}

.instance-config-filter-grid .el-button {
  min-width: 88px;
}

.instance-config-table-shell {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--dora-border, var(--border));
}

.instance-config-table-shell :deep(.el-table) {
  min-width: 680px;
  border-radius: 0;
}

.instance-config-table-shell :deep(.el-table::before) {
  display: none;
}

.instance-config-instance-table :deep(.el-table) {
  min-width: 760px;
}

.instance-config-pagination :deep(.app-pagination-wrap) {
  justify-content: flex-start;
  padding: 10px 0 0;
}

.instance-config-pagination :deep(.app-pagination-bar) {
  width: auto;
  max-width: 100%;
  gap: 8px;
}

.instance-config-action-section {
  padding-top: 14px;
  padding-bottom: 14px;
}

.instance-config-action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--dora-border, var(--border)) 78%, transparent);
  background: color-mix(in srgb, var(--app-overlay-surface, var(--card)) 92%, var(--muted));
}

.instance-config-selection-status {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 8px;
  color: var(--dora-text-muted, var(--muted-foreground));
  font-size: 13px;
}

.instance-config-selection-status span:last-child {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instance-config-status-dot {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: var(--dora-border, var(--border));
}

.instance-config-selection-status.is-selected {
  color: var(--dora-text, var(--foreground));
  font-weight: 500;
}

.instance-config-selection-status.is-selected .instance-config-status-dot {
  background: var(--el-color-success);
}

.instance-config-mode-panel {
  padding: 12px;
  border: 1px solid var(--dora-border, var(--border));
  background: color-mix(in srgb, var(--app-overlay-surface, var(--card)) 90%, var(--muted));
}

.instance-config-mode-panel > p {
  margin: 6px 0 0 32px;
  color: var(--dora-text-muted, var(--muted-foreground));
  font-size: 12px;
  line-height: 1.55;
}

.instance-config-mode-panel .instance-config-warning {
  color: var(--el-color-warning-dark-2);
  font-weight: 500;
}

.instance-config-fields {
  display: grid;
  gap: 16px;
  margin-top: 16px;
}

.instance-config-field-group + .instance-config-field-group {
  padding-top: 16px;
  border-top: 1px solid color-mix(in srgb, var(--dora-border, var(--border)) 72%, transparent);
}

.instance-config-field-title {
  margin-bottom: 10px;
  color: var(--dora-text, var(--foreground));
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 760px) {
  .instance-config {
    gap: 10px;
  }

  .instance-config-section {
    padding: 12px;
  }

  .instance-config-summary {
    align-items: flex-start;
  }

  .instance-config-summary-title {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }

  .instance-config-summary-count {
    min-width: 70px;
    padding-left: 12px;
  }

  .instance-config-filter-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .instance-config-filter-grid .el-button {
    width: 100%;
  }

  .instance-config-action-bar {
    align-items: stretch;
    flex-direction: column;
  }

  .instance-config-action-bar .el-button {
    width: 100%;
  }

  .instance-config-selection-status span:last-child {
    white-space: normal;
  }

  .instance-config-pagination :deep(.app-pagination-bar) {
    flex-wrap: wrap;
  }
}
</style>
