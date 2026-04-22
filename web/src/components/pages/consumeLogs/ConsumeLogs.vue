<script setup lang="ts">
import { ref, computed, reactive, onMounted, nextTick } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import DateTimePicker from '@/components/ui/DateTimePicker.vue'
import { request } from '@/api/api'
// @ts-ignore
import { getPageSize } from '@/util/pageUtils'

interface ConsumeLogItem {
  id: string
  subscription_id: string
  subscription_name: string
  raw_message: string
  matched: number
  extracted_values: string
  send_status: number
  send_error: string
  created_on: string
}

let state = reactive({
  tableData: [] as ConsumeLogItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
})

// 过滤条件
const selectedMatched = ref('all')
const selectedSendStatus = ref('all')

// 获取当天时间范围
const getTodayRange = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return {
    start: `${year}-${month}-${day}T00:00`,
    end: `${year}-${month}-${day}T23:59`
  }
}

// 时间范围过滤 - 默认当天
const todayRange = getTodayRange()
const startTime = ref(todayRange.start)
const endTime = ref(todayRange.end)

// Sheet 状态
const isSheetOpen = ref(false)
const selectedLog = ref<ConsumeLogItem | null>(null)

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

// 获取匹配状态文本
const getMatchedText = (matched: number) => {
  return matched === 1 ? '已匹配' : '未匹配'
}

// 获取匹配状态样式
const getMatchedClass = (matched: number) => {
  return matched === 1
    ? 'bg-green-100 text-green-800 border-green-200'
    : 'bg-gray-100 text-gray-700 border-gray-200'
}

// 获取发送状态文本
const getSendStatusText = (status: number) => {
  const statusMap: Record<number, string> = {
    0: '未发送',
    1: '发送成功',
    2: '发送失败'
  }
  return statusMap[status] || '未知'
}

// 获取发送状态样式
const getSendStatusClass = (status: number) => {
  if (status === 1) return 'bg-green-100 text-green-800 border-green-200'
  if (status === 2) return 'bg-red-100 text-red-700 border-red-200'
  return 'bg-gray-100 text-gray-700 border-gray-200'
}

// 打开日志详情
const openLogSheet = (item: ConsumeLogItem) => {
  selectedLog.value = item
  isSheetOpen.value = true
}

// 查询数据
const queryListData = async (page: number, size: number, subscriptionName = '', matched = '', sendStatus = '') => {
  try {
    let params: any = { page, page_size: size }
    
    if (subscriptionName) params.subscription_name = subscriptionName
    if (matched && matched !== 'all') params.matched = matched
    if (sendStatus && sendStatus !== 'all') params.send_status = sendStatus
    
    // 添加时间范围参数
    if (startTime.value) {
      params.start_time = startTime.value
    }
    if (endTime.value) {
      params.end_time = endTime.value
    }

    const res = await request.get('/consume-logs/list', { params })
    
    state.tableData = []
    await nextTick()
    
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
  await queryListData(state.currPage, state.pageSize, state.search, selectedMatched.value, selectedSendStatus.value)
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

const filterByMatched = async (value: any) => {
  if (value) {
    selectedMatched.value = String(value)
    state.currPage = 1
    await queryListDataWithStatus()
  }
}

const filterBySendStatus = async (value: any) => {
  if (value) {
    selectedSendStatus.value = String(value)
    state.currPage = 1
    await queryListDataWithStatus()
  }
}

// 清除时间过滤 - 恢复当天
const clearTimeFilter = async () => {
  const today = getTodayRange()
  startTime.value = today.start
  endTime.value = today.end
  await filterFunc()
}

// 格式化提取的字段
const formatExtractedValues = (values: string) => {
  try {
    const parsed = JSON.parse(values)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return values
  }
}

onMounted(() => {
  queryListDataWithStatus()
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group" style="flex-wrap: wrap; gap: 8px;">
        <Input
          v-model="state.search"
          placeholder="搜索订阅名称..."
          style="width: 180px;"
          @keyup.enter="filterFunc"
        />
        <Select :model-value="selectedMatched" style="width: 100px;" @update:model-value="filterByMatched">
          <SelectTrigger style="height: 36px;">
            <SelectValue placeholder="匹配状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="1">已匹配</SelectItem>
              <SelectItem value="0">未匹配</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Select :model-value="selectedSendStatus" style="width: 100px;" @update:model-value="filterBySendStatus">
          <SelectTrigger style="height: 36px;">
            <SelectValue placeholder="发送状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="1">发送成功</SelectItem>
              <SelectItem value="2">发送失败</SelectItem>
              <SelectItem value="0">未发送</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <!-- 时间范围过滤 -->
        <div class="flex items-center gap-1">
          <DateTimePicker
            v-model="startTime"
            placeholder="开始时间"
            style="width: 180px;"
            @change="filterFunc"
          />
          <span class="text-sm text-muted-foreground px-1">至</span>
          <DateTimePicker
            v-model="endTime"
            placeholder="结束时间"
            style="width: 180px;"
            @change="filterFunc"
          />
          <Button 
            v-if="startTime || endTime" 
            size="sm" 
            variant="ghost" 
            @click="clearTimeFilter"
          >
            清除
          </Button>
        </div>
        <Button variant="outline" @click="filterFunc">查询</Button>
      </div>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[80px] text-center">序号</TableHead>
            <TableHead class="text-center">ID</TableHead>
            <TableHead class="text-center">订阅名称</TableHead>
            <TableHead class="w-[150px] text-center">原始消息</TableHead>
            <TableHead class="w-[100px] text-center">匹配状态</TableHead>
            <TableHead class="w-[100px] text-center">发送状态</TableHead>
            <TableHead class="w-[150px] text-center">消费时间</TableHead>
            <TableHead class="w-[120px] text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in state.tableData" :key="item.id">
            <TableCell class="font-medium">{{ (state.currPage - 1) * state.pageSize + index + 1 }}</TableCell>
            <TableCell class="font-mono text-sm">{{ item.id }}</TableCell>
            <TableCell>{{ item.subscription_name }}</TableCell>
            <TableCell>
              <div class="max-w-[300px] truncate font-mono text-sm" :title="item.raw_message">
                {{ item.raw_message }}
              </div>
            </TableCell>
            <TableCell>
              <Badge variant="outline" class="text-xs font-medium" :class="getMatchedClass(item.matched)">
                {{ getMatchedText(item.matched) }}
              </Badge>
            </TableCell>
            <TableCell>
              <Badge variant="outline" class="text-xs font-medium" :class="getSendStatusClass(item.send_status)">
                {{ getSendStatusText(item.send_status) }}
              </Badge>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ item.created_on }}
            </TableCell>
            <TableCell class="text-right whitespace-nowrap">
              <Button
                variant="outline"
                size="sm"
                class="min-w-[76px]"
                @click="openLogSheet(item)"
              >
                查看详情
              </Button>
            </TableCell>
          </TableRow>
          <TableRow v-if="state.tableData.length === 0">
            <TableCell :colspan="8">
              <EmptyTableState title="暂无消费日志" description="当前没有可展示的消费日志记录" />
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

    <!-- 日志详情 Sheet -->
    <Sheet v-model:open="isSheetOpen">
      <SheetContent class="sm:max-w-xl">
        <SheetHeader>
          <SheetTitle>消费日志详情</SheetTitle>
        </SheetHeader>
        <div v-if="selectedLog" class="mt-6 space-y-4">
          <div>
            <h3 class="text-sm font-semibold mb-2">基本信息</h3>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">日志 ID:</span>
                <span class="font-mono">{{ selectedLog.id }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">订阅名称:</span>
                <span>{{ selectedLog.subscription_name }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">匹配状态:</span>
                <Badge variant="outline" class="text-xs font-medium" :class="getMatchedClass(selectedLog.matched)">
                  {{ getMatchedText(selectedLog.matched) }}
                </Badge>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">发送状态:</span>
                <Badge variant="outline" class="text-xs font-medium" :class="getSendStatusClass(selectedLog.send_status)">
                  {{ getSendStatusText(selectedLog.send_status) }}
                </Badge>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">消费时间:</span>
                <span>{{ selectedLog.created_on }}</span>
              </div>
            </div>
          </div>

          <div>
            <h3 class="text-sm font-semibold mb-2">原始消息</h3>
            <div class="p-3 bg-muted rounded-md">
              <pre class="text-sm whitespace-pre-wrap font-mono">{{ selectedLog.raw_message }}</pre>
            </div>
          </div>

          <div v-if="selectedLog.matched === 1 && selectedLog.extracted_values">
            <h3 class="text-sm font-semibold mb-2">提取的字段</h3>
            <div class="p-3 bg-muted rounded-md">
              <pre class="text-sm whitespace-pre-wrap font-mono">{{ formatExtractedValues(selectedLog.extracted_values) }}</pre>
            </div>
          </div>

          <div v-if="selectedLog.send_status === 2 && selectedLog.send_error">
            <h3 class="text-sm font-semibold mb-2">发送错误</h3>
            <div class="p-3 bg-red-50 text-red-800 rounded-md">
              <pre class="text-sm whitespace-pre-wrap">{{ selectedLog.send_error }}</pre>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
