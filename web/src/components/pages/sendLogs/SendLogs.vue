<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import ClickableTruncate from '@/components/ui/ClickableTruncate.vue'
import Pagination from '@/components/ui/Pagination.vue'
import DateTimePicker from '@/components/ui/DateTimePicker.vue'

import { useRoute, useRouter } from 'vue-router';
import { request } from '@/api/api';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface LogItem {
  id: number
  task_id: string
  type: string  // 类型：task 或 template
  name: string  // 任务或模板名称
  log: string
  created_on: string
  caller_ip?: string
  status: number
}

const route = useRoute();
const router = useRouter();

let state = reactive({
  tableData: [] as LogItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',  // 保存 taskid，用于过滤
})

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

// 状态过滤
const selectedStatus = ref('all')
// Sheet 相关状态
const isSheetOpen = ref(false)
const selectedLog = ref('')
const selectedTaskName = ref('')
// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))
const hasTemplateFilter = computed(() => String(state.optionValue || '').trim() !== '')

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.floor(n)
}

const getStatusText = (status: number) => {
  return status === 1 ? '成功' : '失败'
}

// 获取类型文本
const getTypeText = (type: string) => {
  if (type === 'template') return '接口调用'
  if (type === 'cron_message') return '定时消息'
  return '系统任务'
}

// 获取类型徽章样式
const getTypeBadgeVariant = (type: string) => {
  return type === 'template' ? 'secondary' : 'default'
}

// 获取显示名称
const getDisplayName = (task: LogItem) => {
  return task.name || '-'
}

// 打开日志详情Sheet
const openLogSheet = (task: LogItem) => {
  selectedLog.value = formatLogDisplayHtml(task);
  selectedTaskName.value = getDisplayName(task)
  isSheetOpen.value = true
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {
    page: String(state.currPage),
    size: String(state.pageSize),
  }
  const name = state.search.trim()
  if (name) nextQuery.name = name
  if (state.optionValue) nextQuery.taskid = state.optionValue
  if (selectedStatus.value && selectedStatus.value !== 'all') nextQuery.status = selectedStatus.value
  if (startTime.value) nextQuery.start_time = startTime.value
  if (endTime.value) nextQuery.end_time = endTime.value
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

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

// 格式化处理显示的日志文本
const formatLogDisplayHtml = (task: LogItem) => {
  let log = task.log;
  log += '\n';
  if (task.caller_ip) {
    log += `调用来源IP：${task.caller_ip}`;
  };
  return log;
}

//触发过滤筛选
const filterFunc = async () => {
  state.currPage = 1
  await queryListDataWithStatus()
}

// 按状态过滤
const filterByStatus = async (value: any) => {
  selectedStatus.value = value;
  state.currPage = 1; // 重置到第一页
  await queryListDataWithStatus()
}

const queryListData = async (page: number, size: number, name = '', taskid = '') => {
  const params: any = { page, size, name, taskid }
  if (selectedStatus.value !== '' && selectedStatus.value !== 'all') {
    params.status = selectedStatus.value
  }
  if (startTime.value) {
    params.start_time = startTime.value
  }
  if (endTime.value) {
    params.end_time = endTime.value
  }

  const rsp = await request.get('/sendlogs/list', { params })
  state.tableData = rsp.data.data.lists || []
  state.total = rsp.data.data.total
}

const queryListDataWithStatus = async (shouldSyncRoute = true) => {
  if (shouldSyncRoute) {
    await syncRouteQuery()
  }
  await queryListData(state.currPage, state.pageSize, state.search, state.optionValue)
}

const clearTemplateFilter = async () => {
  state.optionValue = ''
  state.currPage = 1
  await queryListDataWithStatus()
}

// 清除时间过滤 - 恢复当天
const clearTimeFilter = async () => {
  const today = getTodayRange()
  startTime.value = today.start
  endTime.value = today.end
  state.currPage = 1
  await queryListDataWithStatus()
}

onMounted(async () => {
  state.search = route.query.name?.toString() || ''
  state.optionValue = route.query.taskid?.toString() || ''
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.size, state.pageSize)
  startTime.value = route.query.start_time?.toString() || startTime.value
  endTime.value = route.query.end_time?.toString() || endTime.value

  const explicitStatus = route.query.status?.toString()
  if (explicitStatus && explicitStatus !== 'all') {
    selectedStatus.value = explicitStatus
  } else {
    selectedStatus.value = 'all'
    const legacyQuery = route.query.query?.toString() || ''
    if (legacyQuery) {
      try {
        const queryObj = JSON.parse(decodeURIComponent(legacyQuery))
        if (queryObj.status !== undefined) {
          selectedStatus.value = queryObj.status.toString()
        }
      } catch (error) {
        console.warn('解析旧版query参数失败:', error)
      }
    }
  }

  await queryListDataWithStatus()
})
</script>

<template>
  <div class="space-y-2">
    <div v-if="hasTemplateFilter" class="rounded-lg border border-blue-200/70 bg-blue-50/80 px-3 py-2 text-xs text-blue-800 dark:border-blue-900/60 dark:bg-blue-900/20 dark:text-blue-200">
      当前按模板过滤中：<span class="font-semibold">{{ state.optionValue }}</span>
      <Button size="sm" variant="outline" class="ml-2 h-6 px-2 text-xs" @click="clearTemplateFilter">清除过滤</Button>
    </div>
    <div class="toolbar">
      <div class="search-group" style="flex-wrap: wrap; gap: 8px;">
        <Input
          v-model="state.search"
          placeholder="搜索..."
          style="width: 200px;"
          @keyup.enter="filterFunc"
          @blur="filterFunc"
        />
        <Select v-model="selectedStatus" style="width: 100px;" @update:model-value="filterByStatus">
          <SelectTrigger style="height: 36px;">
            <SelectValue placeholder="选择状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="1">成功</SelectItem>
              <SelectItem value="0">失败</SelectItem>
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
      </div>
    </div>

    <!-- 表格 -->
    <div class="rounded border weak-divider overflow-x-auto">
      <Table class="data-table border-collapse">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead class="w-24">类型</TableHead>
          <TableHead>名称</TableHead>
          <TableHead>发信日志</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">发送时间</TableHead>
          <TableHead class="text-center">详情/状态</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="6" class="empty-state">
            <EmptyTableState 
              title="暂无发信日志" 
              description="还没有任何发信日志记录" 
            >
              <template #icon>
                <svg class="w-8 h-8 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                </svg>
              </template>
            </EmptyTableState>
          </TableCell>
        </TableRow>
        
        <!-- 数据行 -->
        <TableRow v-for="task in state.tableData" :key="task.id">
          <TableCell>{{ task.id }}</TableCell>
          <TableCell>
            <Badge :variant="getTypeBadgeVariant(task.type || 'task')">
              {{ getTypeText(task.type || 'task') }}
            </Badge>
          </TableCell>
          <TableCell>
            <ClickableTruncate :text="getDisplayName(task)" wrapper-class="max-w-[220px] sm:max-w-[360px]" preview-title="名称" />
          </TableCell>
          <TableCell>
            <ClickableTruncate :text="task.log" wrapper-class="max-w-[320px] sm:max-w-[480px]" preview-title="发信日志" />
          </TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ task.created_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="openLogSheet(task)">查看</Button>
            <!-- <Button size="sm" variant="destructive">删除</Button> -->
            <Badge :class="task.status === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600'">
              {{ getStatusText(task.status) }}
            </Badge>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    </div>

    <!-- 分页 -->
    <div class="pagination">
      <Pagination 
        :total="state.total" 
        :current-page="state.currPage" 
        :page-size="state.pageSize" 
        @page-change="changePage"
        @page-size-change="handlePageSizeChange"
      />
    </div>

    <!-- 日志详情Sheet -->
    <Sheet v-model:open="isSheetOpen" class="lg:w-[900px] ">
      <SheetContent class="lg:w-[900px]">
        <SheetHeader>
          <SheetTitle>{{ selectedTaskName }} - 发信日志详情</SheetTitle>
        </SheetHeader>
        <div class="mt-4">
          <div class="rounded-lg p-4 bg-muted/40 dark:bg-white/5 ring-1 ring-border/50 shadow-sm max-h-[82vh] overflow-y-auto break-words">
            <pre class="whitespace-pre-wrap text-sm font-mono leading-relaxed text-foreground">{{ selectedLog }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
