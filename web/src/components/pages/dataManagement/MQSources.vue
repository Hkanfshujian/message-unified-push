<script setup lang="ts">
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
// @ts-ignore
import { getPageSize } from '@/util/pageUtils'
import MQSourceForm from './MQSourceForm.vue'

interface MQSourceItem {
  id: string
  name: string
  type: string
  namesrv_addr: string
  access_key: string
  secret_key: string
  enabled: number
  last_test_status: string
  last_test_time: string
  test_error: string
  created_on: string
  binding_count?: number
}

let state = reactive({
  tableData: [] as MQSourceItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
})

// 状态过滤
const selectedStatus = ref('all')
const STATUS_UNTESTED = '__untested__'

// 队列类型过滤
const selectedType = ref('all')

// 队列类型选项
const typeOptions = [
  { value: 'all', label: '全部类型' },
  { value: 'rocketmq', label: 'RocketMQ' },
  { value: 'kafka', label: 'Kafka' },
  { value: 'rabbitmq', label: 'RabbitMQ' }
]

// Sheet 相关状态
const isAddDialogOpen = ref(false)
const isEditDialogOpen = ref(false)
const isTestDialogOpen = ref(false)
const isDeleteConfirmOpen = ref(false)

const editData = ref<MQSourceItem | null>(null)
const testResult = ref<{ success: boolean; message?: string; error?: string } | null>(null)
const deleteTarget = ref<MQSourceItem | null>(null)
const deleteConfirmInput = ref('')
let autoRefreshTimer: number | null = null

const closeTransientUi = () => {
  isAddDialogOpen.value = false
  isEditDialogOpen.value = false
  isTestDialogOpen.value = false
  isDeleteConfirmOpen.value = false
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer !== null) {
    window.clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

const setupAutoRefreshByPolicy = async () => {
  stopAutoRefresh()
  try {
    const rsp = await request.get('/settings/getsetting', {
      params: { section: 'mq_status_policy' }
    })
    const data = rsp?.data?.data || {}
    const enabled = data.enabled === 'true'
    const intervalSeconds = Number(data.interval_seconds || 300)
    if (!enabled || Number.isNaN(intervalSeconds) || intervalSeconds < 10) {
      return
    }
    autoRefreshTimer = window.setInterval(() => {
      queryListDataWithStatus()
    }, intervalSeconds * 1000)
  } catch {
    // 忽略策略获取异常，保持手动刷新
  }
}

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

// 获取队列类型文本
const getTypeText = (type: string) => {
  const typeMap: Record<string, string> = {
    rocketmq: 'RocketMQ',
    kafka: 'Kafka',
    rabbitmq: 'RabbitMQ'
  }
  return typeMap[type] || type
}

// 获取状态文本（仅两态：在线/离线）
const getStatusText = (status: string) => {
  return status === 'success' ? '在线' : '离线'
}

// 获取状态样式（在线:浅绿，离线:灰底）
const getStatusClass = (status: string) => {
  return status === 'success'
    ? 'bg-green-100 text-green-800 border-green-200'
    : 'bg-gray-100 text-gray-700 border-gray-200'
}

// 打开编辑对话框
const openEditDialog = (item: MQSourceItem) => {
  editData.value = { ...item }
  isEditDialogOpen.value = true
}

// 打开测试对话框
const openTestDialog = (item: MQSourceItem) => {
  editData.value = { ...item }
  testResult.value = null
  isTestDialogOpen.value = true
}

// 打开删除确认
const openDeleteConfirm = (item: MQSourceItem) => {
  deleteTarget.value = item
  deleteConfirmInput.value = ''
  isDeleteConfirmOpen.value = true
}

const closeDeleteConfirm = () => {
  isDeleteConfirmOpen.value = false
  deleteConfirmInput.value = ''
  deleteTarget.value = null
}

const isDeleteMatch = computed(() => {
  const target = deleteTarget.value?.name || ''
  return deleteConfirmInput.value.trim().toLowerCase() === target.trim().toLowerCase() && target.length > 0
})

const showDeleteError = computed(() => {
  return deleteConfirmInput.value.length > 0 && !isDeleteMatch.value
})

// 查询数据
const queryListData = async (page: number, pageSize: number, name: string, type: string, status: string) => {
  try {
    const params: any = {
      page,
      page_size: pageSize,
    }
    if (name) params.name = name
    if (type && type !== 'all') params.type = type
    if (status === STATUS_UNTESTED) {
      params.status = 'untested'
    } else if (status && status !== 'all') {
      params.status = status
    }

    const res = await request.get('/mq-sources/list', { params })
    if (res.data.code === 200) {
      state.tableData = res.data.data.list || []
      state.total = res.data.data.total || 0
    } else {
      // 只有非 200 状态码才显示错误
      state.tableData = []
      state.total = 0
    }
  } catch (error) {
    // 网络错误时静默处理，显示空列表
    state.tableData = []
    state.total = 0
  }
}

const queryListDataWithStatus = async () => {
  const normalizedStatus = selectedStatus.value === STATUS_UNTESTED ? 'failed' : selectedStatus.value
  await queryListData(state.currPage, state.pageSize, state.search, selectedType.value, normalizedStatus)
}

// 分页
const changePage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    state.currPage = page
    await queryListDataWithStatus()
  }
}

const handlePageSizeChange = async (size: number) => {
  if (size <= 0) return
  state.pageSize = size
  state.currPage = 1
  await queryListDataWithStatus()
}

// 过滤
const filterFunc = async () => {
  state.currPage = 1
  await queryListDataWithStatus()
}

const filterByType = async (value: any) => {
  if (value) {
    selectedType.value = String(value)
    state.currPage = 1
    await queryListDataWithStatus()
  }
}

const filterByStatus = async (value: any) => {
  if (value) {
    selectedStatus.value = String(value)
    state.currPage = 1
    await queryListDataWithStatus()
  }
}

// 测试连接
const handleTestConnection = async () => {
  if (!editData.value) return
  
  try {
    const res = await request.post(`/mq-sources/${editData.value.id}/test`)
    if (res.data.code === 200) {
      testResult.value = res.data.data
      if (testResult.value?.success) {
        toast.success('连接测试成功')
      } else {
        toast.error(testResult.value?.error || '连接测试失败')
      }
      // 刷新列表
      await queryListDataWithStatus()
    }
  } catch (error) {
    toast.error('连接测试失败')
    testResult.value = { success: false, error: '网络错误' }
  }
}

// 删除
const handleDelete = async () => {
  if (!deleteTarget.value) return

  try {
    const res = await request.post(`/mq-sources/${deleteTarget.value.id}/delete`)
    if (res.data.code === 200) {
      toast.success('删除成功')
      closeDeleteConfirm()
      await queryListDataWithStatus()
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '删除失败')
  }
}

// 新增/编辑成功回调
const handleSaveSuccess = () => {
  isAddDialogOpen.value = false
  isEditDialogOpen.value = false
  queryListDataWithStatus()
}

onMounted(() => {
  queryListDataWithStatus()
  setupAutoRefreshByPolicy()
})

onBeforeRouteLeave(() => {
  closeTransientUi()
  stopAutoRefresh()
})

onUnmounted(() => {
  closeTransientUi()
  stopAutoRefresh()
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group">
        <Input
          v-model="state.search"
          placeholder="搜索数据源名称..."
          class="search-input"
          @keyup.enter="filterFunc"
        />
        <Select :model-value="selectedType" @update:model-value="filterByType">
          <SelectTrigger class="filter-select w-full">
            <SelectValue placeholder="队列类型" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Select :model-value="selectedStatus" @update:model-value="filterByStatus">
          <SelectTrigger class="filter-select w-full">
            <SelectValue placeholder="状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="success">在线</SelectItem>
              <SelectItem value="failed">离线</SelectItem>
              <SelectItem :value="STATUS_UNTESTED">离线</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Button variant="outline" @click="filterFunc">查询</Button>
      </div>
      <Button class="primary-btn" v-permission="'data:mq-source:add'" @click="isAddDialogOpen = true">新增数据源</Button>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[80px] text-center">序号</TableHead>
            <TableHead class="text-center">ID</TableHead>
            <TableHead class="text-center">队列名称</TableHead>
            <TableHead class="w-[120px] text-center">队列类型</TableHead>
            <TableHead class="text-center">队列地址</TableHead>
            <TableHead class="w-[120px] text-center">外部绑定</TableHead>
            <TableHead class="w-[120px] text-center">状态</TableHead>
            <TableHead class="w-[120px] text-center">最后测试时间</TableHead>
            <TableHead class="w-[240px] text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in state.tableData" :key="item.id">
            <TableCell class="font-medium">{{ (state.currPage - 1) * state.pageSize + index + 1 }}</TableCell>
            <TableCell class="font-mono text-sm">{{ item.id }}</TableCell>
            <TableCell class="font-medium">{{ item.name }}</TableCell>
            <TableCell>
              <Badge variant="outline">{{ getTypeText(item.type) }}</Badge>
            </TableCell>
            <TableCell class="font-mono text-sm max-w-[300px] truncate" :title="item.namesrv_addr">
              {{ item.namesrv_addr }}
            </TableCell>
            <TableCell>
              <Badge variant="secondary">{{ item.binding_count || 0 }} 个订阅</Badge>
            </TableCell>
            <TableCell>
              <Badge variant="outline" class="text-xs font-medium" :class="getStatusClass(item.last_test_status)">
                {{ getStatusText(item.last_test_status) }}
              </Badge>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ item.last_test_time || '-' }}
            </TableCell>
            <TableCell class="text-right whitespace-nowrap">
              <div class="inline-flex items-center justify-end gap-2">
                <Button
                v-permission="'data:mq-source:test'"
                variant="outline"
                size="sm"
                class="min-w-[76px]"
                @click="openTestDialog(item)"
              >
                测试连接
              </Button>
              <Button
                v-permission="'data:mq-source:edit'"
                variant="outline"
                size="sm"
                class="min-w-[64px]"
                @click="openEditDialog(item)"
              >
                编辑
              </Button>
              <Button
                v-permission="'data:mq-source:delete'"
                variant="outline"
                size="sm"
                class="min-w-[64px] text-red-600 border-red-200 hover:bg-red-50 hover:text-red-700"
                @click="openDeleteConfirm(item)"
              >
                删除
              </Button>
              </div>
            </TableCell>
          </TableRow>
          <TableRow v-if="state.tableData.length === 0">
            <TableCell :colspan="9">
              <EmptyTableState title="暂无数据源" description="还没有配置任何消息队列数据源" />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <div class="pagination">
      <Pagination
        :current-page="state.currPage"
        :total-pages="totalPages"
        :page-size="state.pageSize"
        :total="state.total"
        @change-page="changePage"
        @change-page-size="handlePageSizeChange"
      />
    </div>

    <!-- 新增对话框 -->
    <Dialog v-model:open="isAddDialogOpen">
      <DialogContent class="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>新增数据源</DialogTitle>
        </DialogHeader>
        <MQSourceForm @success="handleSaveSuccess" />
      </DialogContent>
    </Dialog>

    <!-- 编辑对话框 -->
    <Dialog v-model:open="isEditDialogOpen">
      <DialogContent class="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>编辑数据源</DialogTitle>
        </DialogHeader>
        <MQSourceForm v-if="editData" :data="editData" @success="handleSaveSuccess" />
      </DialogContent>
    </Dialog>

    <!-- 测试连接对话框 -->
    <Dialog v-model:open="isTestDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>测试连接 - {{ editData?.name }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div class="text-sm text-muted-foreground">
            正在测试连接到: <code class="bg-muted px-1 py-0.5 rounded">{{ editData?.namesrv_addr }}</code>
          </div>
          <Button @click="handleTestConnection" :disabled="!editData">
            开始测试
          </Button>
          <div v-if="testResult" class="p-3 rounded-md" :class="testResult.success ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'">
            {{ testResult.success ? '✅ ' + (testResult.message || '连接成功') : '❌ ' + (testResult.error || '连接失败') }}
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <!-- 删除确认对话框 -->
    <Dialog v-model:open="isDeleteConfirmOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <p class="text-sm text-muted-foreground">
            请输入数据源名称 <strong>{{ deleteTarget?.name }}</strong> 以确认删除：
          </p>
          <Input
            v-model="deleteConfirmInput"
            placeholder="输入数据源名称"
          />
          <p v-if="showDeleteError" class="text-sm text-destructive">
            输入的名称不匹配
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="closeDeleteConfirm">取消</Button>
          <Button
            variant="destructive"
            :disabled="!isDeleteMatch"
            @click="handleDelete"
          >
            删除
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
