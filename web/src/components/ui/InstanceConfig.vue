<script setup lang="ts">
import { ref, computed, watch, reactive } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from "@/components/ui/badge"
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import Pagination from '@/components/ui/Pagination.vue'
import { CONSTANT } from '@/constant'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { generateBizUniqueID } from '@/util/uuid'

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
    toast.error('请选择发送渠道')
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
        toast.error(`该${entityName}已存在动态接收实例，一个${entityName}只能配置一个动态接收实例`)
        return
      }
      if (insTableData.value.length > 0) {
        toast.error(`动态接收实例不能与固定接收实例混合使用，请先删除所有固定实例`)
        return
      }
    }
    
    // 如果要添加固定接收实例，但已有动态接收实例
    if (formData.value.allowMultiRecip !== true && hasDynamicInstance) {
      toast.error(`该${entityName}已配置动态接收实例，不能再添加固定接收实例`)
      return
    }
  }

  // 验证内容类型
  const contentType = formData.value.templ_type
  if (!contentType) {
    toast.error('请选择消息格式')
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
      toast.error(`模板的 ${contentType} 格式内容为空，无法添加此类型的实例`)
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
      toast.success(response.data.msg)
      // 重新加载实例列表
      await queryInsListData()
      // 清空表单
      selectedChannel.value = null
      formData.value = { allowMultiRecip: false }
    } else {
      toast.error(response.data.msg || '添加实例失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '添加实例失败')
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
      toast.error(`关闭动态接收者前请先填写固定接收者（${recipientField}）`)
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
      toast.success(response.data.msg || '更新成功')
      await queryInsListData()
      dynamicConfirmOpen.value = false
      dynamicToggleTarget.value = null
      dynamicRecipientInput.value = ''
    } else {
      toast.error(response.data.msg || '更新失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '更新失败')
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
const handleDeleteIns = async (insId: string) => {
  try {
    const response = await request.post(apiConfig.value.deleteIns, { id: insId }, {
      meta: {
        silentBizToast: true,
        silentErrorToast: true
      }
    } as any)
    if (response.status === 200 && response.data.code === 200) {
      toast.success(response.data.msg)
      await queryInsListData()
    } else {
      toast.error(response.data.msg || '删除失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '删除失败')
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
      toast.success(response.data.msg)
      // 重新加载确保数据同步
      await queryInsListData()
    } else {
      toast.error(response.data.msg || '更新失败')
      // 失败时恢复原状态
      if (insIndex !== -1) {
        insTableData.value[insIndex].enable = currentStatus
      }
    }
  } catch (error: any) {
    console.error('状态切换失败:', error)
    toast.error(error.response?.data?.msg || '更新失败')
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
    toast.error('获取渠道列表失败')
  } finally {
    channelState.loading = false
  }
}

const handleChannelPageChange = async (page: number) => {
  await queryChannelList(page)
}

const handleChannelPageSizeChange = async (size: number) => {
  channelState.pageSize = size
  await queryChannelList(1)
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
  <div class="space-y-4" :class="{ 'px-4 pb-4': inDialog }">
    <!-- 信息展示区域 -->
    <div v-if="data" class="p-3 bg-muted rounded-lg space-y-1">
      <div class="flex items-baseline gap-2">
        <span class="text-base font-semibold">{{ data[apiConfig.nameField] }}</span>
        <Badge variant="outline" class="text-xs">{{ data.id }}</Badge>
      </div>
      <div class="text-xs text-muted-foreground">
        为此模板配置发送实例
      </div>
    </div>

    <!-- 添加实例表单 -->
    <div class="space-y-4">
      <div class="space-y-2">
        <Label class="text-sm font-medium">选择发送渠道</Label>
        <div class="grid grid-cols-1 md:grid-cols-12 gap-2">
          <Input
            v-model="channelFilters.name"
            class="md:col-span-6"
            placeholder="按渠道名称模糊搜索..."
            @keyup.enter="handleChannelSearch"
          />
          <select
            v-model="channelFilters.type"
            class="h-9 rounded-md border border-input bg-transparent px-3 text-sm md:col-span-4"
            @change="handleChannelSearch"
          >
            <option v-for="option in channelTypeOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
          <Button class="md:col-span-2" size="sm" variant="outline" @click="handleChannelSearch">搜索</Button>
        </div>
      </div>

      <div class="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-12">选择</TableHead>
              <TableHead class="w-16">序号</TableHead>
              <TableHead class="w-40">渠道ID</TableHead>
              <TableHead>渠道名称</TableHead>
              <TableHead class="w-32">渠道类型</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="(channel, idx) in channelState.list"
              :key="channel.id"
              class="cursor-pointer"
              :class="selectedChannel?.id === channel.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''"
              @click="handleSelectChannel(channel)"
            >
              <TableCell @click.stop>
                <input
                  type="checkbox"
                  class="h-4 w-4"
                  :checked="isChannelChecked(channel.id)"
                  @change="(event) => handleChannelCheck(channel, (event.target as HTMLInputElement).checked)"
                />
              </TableCell>
              <TableCell>{{ (channelState.currPage - 1) * channelState.pageSize + idx + 1 }}</TableCell>
              <TableCell class="font-mono text-xs">{{ channel.id }}</TableCell>
              <TableCell>{{ channel.name }}</TableCell>
              <TableCell>{{ formatChannelTypeLabel(channel.type) }}</TableCell>
            </TableRow>
            <TableRow v-if="!channelState.loading && channelState.list.length === 0">
              <TableCell colspan="5" class="h-20 text-center text-muted-foreground">
                暂无匹配渠道
              </TableCell>
            </TableRow>
            <TableRow v-if="channelState.loading">
              <TableCell colspan="5" class="h-20 text-center text-muted-foreground">
                加载中...
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <Pagination
        :total="channelState.total"
        :current-page="channelState.currPage"
        :page-size="channelState.pageSize"
        @page-change="handleChannelPageChange"
        @page-size-change="handleChannelPageSizeChange"
      />

      <div class="flex justify-between items-center">
        <div class="text-xs text-muted-foreground">
          {{ selectedChannel ? `已选择：${selectedChannel.name}（${formatChannelTypeLabel(selectedChannel.type)}）` : '请选择一个渠道后再添加实例' }}
        </div>
        <Button size="sm" variant="outline" @click="handleAddSubmit">添加实例</Button>
      </div>
    </div>

    <!-- 渠道配置表单 -->
    <div v-if="currentChannelConfig" class="mt-4">
      <!-- 动态接收者勾选框 -->
      <div v-if="currentDynamicRecipient?.support" class="mb-4 p-3 border rounded-lg bg-gray-50 dark:bg-gray-800/50">
        <div class="flex items-center space-x-2">
          <Switch 
            :model-value="formData.allowMultiRecip" 
            @update:model-value="(val: boolean) => formData.allowMultiRecip = val"
            :id="`allow-multi-${selectedChannel?.id || 'none'}`" 
          />
          <Label :for="`allow-multi-${selectedChannel?.id || 'none'}`" class="text-sm font-medium cursor-pointer">
            动态接收者模式
          </Label>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ml-8">
          {{ formData.allowMultiRecip ? '支持动态接收者，发送时通过API指定接收者列表（群发模式）' : '固定接收者模式，需要在下方配置固定接收者' }}
        </p>
        <p v-if="formData.allowMultiRecip" class="text-xs text-orange-500 dark:text-orange-400 mt-1 ml-8 font-medium">
          ⚠️ 注意：一个模板只能配置一个动态接收实例，且不能与固定接收实例混合使用
        </p>
      </div>

      <!-- 接收者输入字段 -->
      <div v-if="shouldShowRecipientInput" class="mb-2">
        <Label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">实例配置</Label>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ currentDynamicRecipient?.label }}
            </label>
            <Input 
              v-model="formData[currentDynamicRecipientField]" 
              :placeholder="`请输入${currentDynamicRecipient?.desc}`"
              type="text" 
              class="text-sm" 
            />
          </div>
        </div>
        <p v-if="isQyWeiXinFixedRecipientEmpty" class="mt-2 text-xs text-destructive">
          企业微信应用固定模式下 to_user 不能为空。为空将无法保存，且系统不会兜底为 @all。
        </p>
      </div>
      
      <!-- 实例配置输入字段（排除动态接收者字段） -->
      <div v-if="currentChannelConfig.taskInsInputs && currentChannelConfig.taskInsInputs.length > 0" class="mb-2">
        <Label class="text-sm font-medium mb-1">实例配置</Label>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div 
            v-for="input in currentChannelConfig.taskInsInputs.filter((inp: any) => inp.col !== currentDynamicRecipient?.field)" 
            :key="input.col" 
            class="space-y-2"
          >
            <label class="text-xs font-medium text-muted-foreground">{{ input.label || input.desc }}</label>
            <Input 
              v-model="formData[input.col]" 
              :placeholder="input.desc || `请输入${input.label}`"
              :type="input.type || 'text'" 
              class="w-full" 
            />
          </div>
        </div>
      </div>

      <!-- 单选框 -->
      <div v-if="currentChannelConfig.taskInsRadios && currentChannelConfig.taskInsRadios.length > 0" class="mt-4">
        <Label class="text-sm font-medium mb-2">消息格式</Label>
        <RadioGroup v-model="formData.templ_type" class="flex gap-4">
          <div v-for="radio in currentChannelConfig.taskInsRadios" :key="radio.subLabel" class="flex items-center space-x-2">
            <RadioGroupItem :value="radio.subLabel" :id="radio.subLabel" />
            <Label :for="radio.subLabel" class="text-sm cursor-pointer">{{ radio.subLabel }}</Label>
          </div>
        </RadioGroup>
      </div>
    </div>

    <!-- 关联的实例表 -->
    <div class="mt-4">
      <h3 class="text-sm font-medium mb-3">已经关联的实例</h3>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>渠道名称</TableHead>
            <TableHead>内容类型</TableHead>
            <TableHead>接收者</TableHead>
            <TableHead>动态接收者</TableHead>
            <TableHead class="text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="ins in insTableData" :key="ins.id">
            <TableCell>
              <div class="font-medium">{{ ins.way_name || '未命名' }}</div>
              <div class="text-xs text-muted-foreground">{{ ins.way_type }}</div>
            </TableCell>
            <TableCell>
              <Badge variant="secondary">{{ ins.content_type }}</Badge>
            </TableCell>
            <TableCell>
              <Badge v-if="formatInsConfigDisplay(ins)" variant="secondary">{{ formatInsConfigDisplay(ins) }}</Badge>
              <span v-else class="text-sm text-muted-foreground">-</span>
            </TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <Switch
                  :model-value="isDynamicRecipientEnabled(ins)"
                  @update:model-value="() => handleToggleDynamicRecipient(ins)"
                />
                <span class="text-xs" :class="isDynamicRecipientEnabled(ins) ? 'text-emerald-600' : 'text-muted-foreground'">
                  {{ isDynamicRecipientEnabled(ins) ? '开启' : '关闭' }}
                </span>
              </div>
            </TableCell>
            <TableCell class="text-center">
              <div class="flex items-center justify-center gap-2">
                <Switch 
                  :model-value="ins.enable === 1" 
                  @update:model-value="() => handleToggleEnable(ins.id, ins.enable)" 
                />
                <Button 
                  size="sm" 
                  variant="outline" 
                  class="text-red-500 border-red-300 hover:bg-red-50 hover:border-red-400 hover:text-red-600 hover:shadow-md transition-all duration-200" 
                  @click="handleDeleteIns(ins.id)"
                >
                  删除
                </Button>
              </div>
            </TableCell>
          </TableRow>
          <TableRow v-if="!insTableData || insTableData.length === 0">
            <TableCell :colspan="5" class="h-24">
              <EmptyTableState title="暂无实例" description="还没有配置任何实例，请先添加" />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <Dialog v-model:open="dynamicConfirmOpen">
      <DialogContent class="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle>确认操作</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">
          确认将动态接收者模式{{ dynamicToggleNext ? '开启' : '关闭' }}吗？
        </p>
        <div v-if="shouldRequireRecipientInput" class="space-y-2">
          <Label class="text-sm">固定接收者（to_user）</Label>
          <Input
            v-model="dynamicRecipientInput"
            placeholder="关闭动态接收者前请输入固定接收者"
          />
          <p class="text-xs text-muted-foreground">关闭后将按该固定接收者发送</p>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="cancelToggleDynamicRecipient">取消</Button>
          <Button
            type="button"
            :variant="dynamicToggleNext ? 'default' : 'destructive'"
            @click="confirmToggleDynamicRecipient"
          >
            确认
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
