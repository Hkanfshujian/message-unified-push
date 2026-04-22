<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
// @ts-ignore
import { getPageSize } from '@/util/pageUtils'
import SubscriptionForm from './SubscriptionForm.vue'

interface SubscriptionItem {
  id: string
  source_id: string
  source_name: string
  name: string
  topic: string
  tag: string
  group_name: string
  validate_regex: string
  extract_regex: string
  extract_field: string
  extract_rules?: Array<{ field: string; regex: string }>
  template_id: string
  template_name: string
  template_content_type?: string
  consume_mode?: string
  status: string
  total_consumed: number
  total_sent: number
  total_failed: number
  last_consume_time: string
  created_on: string
}

interface MQSourceOption {
  id: string
  name: string
}

interface TemplateOption {
  id: string
  name: string
}

const normalizeTemplateContentType = (value?: string) => {
  const v = String(value || '').trim().toLowerCase()
  if (v === 'html' || v === 'markdown' || v === 'text') return v
  if (v === 'push' || v === 'pull' || v === '') return 'text'
  return 'text'
}

let state = reactive({
  tableData: [] as SubscriptionItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
})

// 过滤条件
const selectedStatus = ref('all')
const selectedSource = ref('all')

// 下拉选项
const sourceOptions = ref<MQSourceOption[]>([])
const templateOptions = ref<TemplateOption[]>([])

// 对话框状态
const isAddDialogOpen = ref(false)
const isEditDialogOpen = ref(false)
const editData = ref<SubscriptionItem | null>(null)

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    running: '运行中',
    stopped: '已停止'
  }
  return statusMap[status] || status
}

// 获取状态样式（运行中:浅绿，已停止:灰底）
const getStatusClass = (status: string) => {
  return status === 'running'
    ? 'bg-green-100 text-green-800 border-green-200'
    : 'bg-gray-100 text-gray-700 border-gray-200'
}

const formatLastConsumeTime = (value: string) => {
  if (!value) return '-'
  // 后端 util.Time 零值会序列化为 0001-01-01...
  if (value.includes('0001-01-01')) return '-'
  return value
}

const getStatClass = (value: number, kind: 'consume' | 'sent' | 'failed') => {
  if (!value) return 'text-muted-foreground'
  if (kind === 'failed') return 'text-red-600'
  if (kind === 'sent') return 'text-green-700'
  return 'text-slate-700 dark:text-slate-200'
}

// 加载数据源选项
const loadSourceOptions = async () => {
  try {
    const res = await request.get('/mq-sources/list', { params: { page: 1, page_size: 100 } })
    if (res.data.code === 200) {
      sourceOptions.value = (res.data.data.list || []).map((item: any) => ({
        id: item.id,
        name: item.name
      }))
    }
  } catch (error) {
    console.error('加载数据源失败:', error)
  }
}

// 加载模板选项
const loadTemplateOptions = async () => {
  try {
    const res = await request.get('/templates/list', { params: { page: 1, size: 100 } })
    if (res.data.code === 200) {
      templateOptions.value = (res.data.data.lists || []).map((item: any) => ({
        id: item.id,
        name: item.name
      }))
    }
  } catch (error) {
    console.error('加载模板失败:', error)
  }
}

// 打开编辑对话框
const openEditDialog = async (item: SubscriptionItem) => {
  try {
    const res = await request.get(`/subscriptions/${item.id}`)
    if (res.data.code === 200 && res.data.data) {
      const detail = res.data.data
      editData.value = {
        ...item,
        ...detail,
        template_content_type: normalizeTemplateContentType(
          detail.template_content_type || detail.consume_mode || item.template_content_type
        )
      }
      isEditDialogOpen.value = true
      return
    }
  } catch (error) {
    // ignore and fallback to list item data
  }
  editData.value = {
    ...item,
    template_content_type: normalizeTemplateContentType(item.template_content_type || item.consume_mode)
  }
  isEditDialogOpen.value = true
}

// 查询数据
const queryListData = async (page: number, pageSize: number, name: string, sourceId: string, status: string) => {
  try {
    const params: any = {
      page,
      page_size: pageSize,
    }
    if (name) params.name = name
    if (sourceId && sourceId !== 'all') params.source_id = sourceId
    if (status && status !== 'all') params.status = status

    const res = await request.get('/subscriptions/list', { params })
    if (res.data.code === 200) {
      state.tableData = res.data.data.list || []
      state.total = res.data.data.total || 0
    } else {
      state.tableData = []
      state.total = 0
    }
  } catch (error) {
    // 静默处理错误，显示空列表
    state.tableData = []
    state.total = 0
  }
}

const queryListDataWithStatus = async () => {
  await queryListData(state.currPage, state.pageSize, state.search, selectedSource.value, selectedStatus.value)
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

const filterBySource = async (value: any) => {
  if (value) {
    selectedSource.value = String(value)
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

// 启动/停止
const handleToggleStatus = async (item: SubscriptionItem) => {
  try {
    const action = item.status === 'running' ? 'stop' : 'start'
    const res = await request.post(`/subscriptions/${item.id}/${action}`)
    if (res.data.code === 200) {
      toast.success(item.status === 'running' ? '订阅已停止' : '订阅启动中...')
      await queryListDataWithStatus()
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '操作失败')
  }
}

// 删除
const handleDelete = async (item: SubscriptionItem) => {
  if (item.status === 'running') {
    toast.warning('请先停止订阅后再删除')
    return
  }

  if (!confirm(`确认删除订阅 "${item.name}"？`)) return

  try {
    const res = await request.post(`/subscriptions/${item.id}/delete`)
    if (res.data.code === 200) {
      toast.success('删除成功')
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
  loadSourceOptions()
  loadTemplateOptions()
  queryListDataWithStatus()
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group">
        <Input
          v-model="state.search"
          placeholder="搜索订阅名称..."
          class="search-input"
          @keyup.enter="filterFunc"
        />
        <Select :model-value="selectedSource" @update:model-value="filterBySource">
          <SelectTrigger class="filter-select w-full">
            <SelectValue placeholder="选择数据源" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部数据源</SelectItem>
              <SelectItem v-for="opt in sourceOptions" :key="opt.id" :value="opt.id">
                {{ opt.name }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Select :model-value="selectedStatus" @update:model-value="filterByStatus">
          <SelectTrigger class="filter-select w-full">
            <SelectValue placeholder="选择订阅状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="running">运行中</SelectItem>
              <SelectItem value="stopped">已停止</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Button variant="outline" @click="filterFunc">查询</Button>
      </div>
      <Button class="primary-btn" v-permission="'data:subscription:add'" @click="isAddDialogOpen = true">新增订阅</Button>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[80px] text-center">序号</TableHead>
            <TableHead class="text-center">ID</TableHead>
            <TableHead class="text-center">订阅名称</TableHead>
            <TableHead class="text-center">Topic</TableHead>
            <TableHead class="text-center">Tag</TableHead>
            <TableHead class="w-[100px] text-center">状态</TableHead>
            <TableHead class="w-[100px] text-center">消费/发送/失败</TableHead>
            <TableHead class="w-[120px] text-center">最后消费时间</TableHead>
            <TableHead class="w-[260px] text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in state.tableData" :key="item.id">
            <TableCell class="font-medium">{{ (state.currPage - 1) * state.pageSize + index + 1 }}</TableCell>
            <TableCell class="font-mono text-sm">{{ item.id }}</TableCell>
            <TableCell class="font-medium">{{ item.name }}</TableCell>
            <TableCell class="font-mono text-sm">{{ item.topic }}</TableCell>
            <TableCell class="font-mono text-sm">{{ item.tag || '*' }}</TableCell>
            <TableCell>
              <Badge variant="outline" class="text-xs font-medium" :class="getStatusClass(item.status)">
                {{ getStatusText(item.status) }}
              </Badge>
            </TableCell>
            <TableCell class="text-right text-sm">
              <span :class="getStatClass(item.total_consumed, 'consume')">{{ item.total_consumed }}</span>
              <span class="text-muted-foreground"> / </span>
              <span :class="getStatClass(item.total_sent, 'sent')">{{ item.total_sent }}</span>
              <span class="text-muted-foreground"> / </span>
              <span :class="getStatClass(item.total_failed, 'failed')">{{ item.total_failed }}</span>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ formatLastConsumeTime(item.last_consume_time) }}
            </TableCell>
            <TableCell class="text-right whitespace-nowrap">
              <div class="inline-flex items-center justify-end gap-2">
                <Button
                v-permission="item.status === 'running' ? 'data:subscription:stop' : 'data:subscription:start'"
                variant="outline"
                size="sm"
                class="min-w-[64px]"
                :class="item.status === 'running'
                  ? 'text-orange-600 border-orange-200 hover:bg-orange-50 hover:text-orange-700'
                  : 'text-emerald-600 border-emerald-200 hover:bg-emerald-50 hover:text-emerald-700'"
                @click="handleToggleStatus(item)"
              >
                {{ item.status === 'running' ? '停止' : '启动' }}
              </Button>
              <Button
                v-permission="'data:subscription:edit'"
                variant="outline"
                size="sm"
                class="min-w-[64px]"
                @click="openEditDialog(item)"
              >
                编辑
              </Button>
              <Button
                v-permission="'data:subscription:delete'"
                variant="outline"
                size="sm"
                class="min-w-[64px] text-red-600 border-red-200 hover:bg-red-50 hover:text-red-700"
                @click="handleDelete(item)"
              >
                删除
              </Button>
              </div>
            </TableCell>
          </TableRow>
          <TableRow v-if="state.tableData.length === 0">
            <TableCell :colspan="9">
              <EmptyTableState title="暂无订阅消息" description="还没有配置任何订阅规则" />
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
      <DialogContent class="sm:max-w-[980px] h-[88vh] sm:h-[760px] max-h-[88vh] overflow-hidden p-0 flex flex-col">
        <DialogHeader class="px-6 pt-5 pb-3 border-b shrink-0">
          <DialogTitle>新增订阅</DialogTitle>
        </DialogHeader>
        <SubscriptionForm
          :source-options="sourceOptions"
          :template-options="templateOptions"
          @success="handleSaveSuccess"
        />
      </DialogContent>
    </Dialog>

    <!-- 编辑对话框 -->
    <Dialog v-model:open="isEditDialogOpen">
      <DialogContent class="sm:max-w-[980px] h-[88vh] sm:h-[760px] max-h-[88vh] overflow-hidden p-0 flex flex-col">
        <DialogHeader class="px-6 pt-5 pb-3 border-b shrink-0">
          <DialogTitle>编辑订阅</DialogTitle>
        </DialogHeader>
        <SubscriptionForm
          v-if="editData"
          :data="editData"
          :source-options="sourceOptions"
          :template-options="templateOptions"
          @success="handleSaveSuccess"
        />
      </DialogContent>
    </Dialog>
  </div>
</template>
