<script setup lang="ts">
import { onMounted, ref, computed, reactive } from 'vue'
import { Dialog, DialogContent, DialogHeader } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import DateTimePicker from '@/components/ui/DateTimePicker.vue'
import { toast } from 'vue-sonner'

// @ts-ignore
import { request } from '@/api/api'
// @ts-ignore
import { getPageSize } from '@/util/pageUtils'

interface LoginLog {
  id: number
  user_id: number
  username: string
  ip: string
  ua: string
  created_on: string
}

const loading = ref(false)
const ipDialogOpen = ref(false)
const ipLoading = ref(false)
const selectedIp = ref('')
const ipInfo = ref<any>(null)
const search = ref('')

// 分页状态
const state = reactive({
  logs: [] as LoginLog[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
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

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

const fetchLogs = async () => {
  loading.value = true
  try {
    const params: any = {
      page: state.currPage,
      page_size: state.pageSize,
    }
    if (startTime.value) {
      params.start_time = startTime.value
    }
    if (endTime.value) {
      params.end_time = endTime.value
    }
    const rsp = await request.get('/loginlogs/recent', { params })
    const data = rsp.data
    if (data && data.code === 200 && data.data) {
      state.logs = data.data.lists || []
      state.total = data.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

// 分页
const changePage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    state.currPage = page
    await fetchLogs()
  }
}

const handlePageSizeChange = async (size: number) => {
  if (size <= 0) return
  state.pageSize = size
  state.currPage = 1
  await fetchLogs()
}

// 清除时间过滤 - 恢复当天
const clearTimeFilter = async () => {
  const today = getTodayRange()
  startTime.value = today.start
  endTime.value = today.end
  state.currPage = 1
  await fetchLogs()
}

onMounted(fetchLogs)

const displayLogs = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return state.logs
  return state.logs.filter((item) =>
    item.username?.toLowerCase().includes(keyword) ||
    item.ip?.toLowerCase().includes(keyword) ||
    item.ua?.toLowerCase().includes(keyword)
  )
})

const openIpInfo = async (ip: string) => {
  selectedIp.value = ip
  ipDialogOpen.value = true
  ipLoading.value = true
  ipInfo.value = null
  try {
    const rsp = await fetch(`https://api.ip.sb/geoip/${encodeURIComponent(ip)}`)
    if (!rsp.ok) throw new Error('请求失败')
    const data = await rsp.json()
    ipInfo.value = data
  } catch (e) {
    toast.error('获取IP信息失败')
  } finally {
    ipLoading.value = false
  }
}

const formatUtcOffset = (offset: any) => {
  const num = typeof offset === 'number' ? offset : parseInt(offset, 10)
  if (Number.isNaN(num)) return '-'
  const hours = Math.floor(num / 3600)
  const sign = hours >= 0 ? '+' : '-'
  const abs = Math.abs(hours)
  return `UTC${sign}${abs}`
}

const ipDisplayRows = computed(() => {
  if (!ipInfo.value) return [] as Array<{ label: string, value: string }>
  const info = ipInfo.value
  const rows: Array<{ label: string, value: string }> = []

  const ip = info.ip || selectedIp.value
  if (ip) rows.push({ label: 'IP', value: String(ip) })

  const country = [info.country, info.country_code ? `(${info.country_code})` : ''].filter(Boolean).join(' ')
  if (country) rows.push({ label: '国家/地区', value: country })

  const tz = [info.timezone, info.offset != null ? formatUtcOffset(info.offset) : ''].filter(Boolean).join(' · ')
  if (tz) rows.push({ label: '时区', value: tz })

  if (info.isp) rows.push({ label: 'ISP', value: String(info.isp) })
  if (info.organization) rows.push({ label: '组织', value: String(info.organization) })

  const asn = [info.asn, info.asn_organization].filter(Boolean).join(' - ')
  if (asn) rows.push({ label: 'ASN', value: asn })

  const continent = info.continent_code
  if (continent) rows.push({ label: '大洲', value: String(continent) })

  const coordParts = [] as string[]
  if (info.latitude != null) coordParts.push(String(info.latitude))
  if (info.longitude != null) coordParts.push(String(info.longitude))
  const coord = coordParts.join(', ')
  if (coord) rows.push({ label: '坐标', value: coord })

  return rows
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group" style="flex-wrap: wrap; gap: 8px;">
        <Input
          v-model="search"
          placeholder="搜索用户名/IP/UA..."
          style="width: 200px;"
        />
        <!-- 时间范围过滤 -->
        <div class="flex items-center gap-1">
          <DateTimePicker
            v-model="startTime"
            placeholder="开始时间"
            style="width: 180px;"
            @change="fetchLogs"
          />
          <span class="text-sm text-muted-foreground px-1">至</span>
          <DateTimePicker
            v-model="endTime"
            placeholder="结束时间"
            style="width: 180px;"
            @change="fetchLogs"
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
        <Button variant="outline" @click="fetchLogs" :disabled="loading">
          {{ loading ? '刷新中...' : '刷新' }}
        </Button>
      </div>
    </div>
    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="text-center">用户名</TableHead>
            <TableHead class="text-center">IP</TableHead>
            <TableHead class="text-center">UA</TableHead>
            <TableHead class="w-[160px] text-center">登录时间</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="!displayLogs.length">
            <TableCell :colspan="4">
              <EmptyTableState title="暂无登录日志" description="当前没有可展示的登录日志记录" />
            </TableCell>
          </TableRow>
          <TableRow v-for="item in displayLogs" :key="item.id">
            <TableCell>{{ item.username }}</TableCell>
            <TableCell>
              <button class="text-brand hover:underline" @click="openIpInfo(item.ip)">{{ item.ip }}</button>
            </TableCell>
            <TableCell class="truncate max-w-[220px] sm:max-w-[420px]" :title="item.ua">{{ item.ua }}</TableCell>
            <TableCell class="whitespace-nowrap w-[160px]">{{ item.created_on }}</TableCell>
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
  </div>

  <Dialog :open="ipDialogOpen" @update:open="val => (ipDialogOpen = val)">
    <DialogContent class="w-[90vw] max-w-[90vw] sm:max-w-lg max-h-[80vh] overflow-y-auto">

      <DialogHeader>
        <!-- <VisuallyHidden> -->
        <!-- <DialogTitle>IP 信息</DialogTitle> -->
        <!-- </VisuallyHidden> -->
      </DialogHeader>

      <div v-if="ipLoading" class="text-sm text-muted-foreground">加载中...</div>
      <div v-else class="space-y-3 text-sm">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2">
          <template v-for="(row, idx) in ipDisplayRows" :key="idx">
            <div class="text-muted-foreground">{{ row.label }}</div>
            <div class="break-all" :class="{ 'font-mono': row.label === 'IP' || row.label === '坐标' }">{{ row.value }}
            </div>
          </template>
        </div>
        <div class="text-xs text-muted-foreground mt-2">数据来源：<a
            :href="'https://api.ip.sb/geoip/' + encodeURIComponent(selectedIp)" target="_blank" rel="noreferrer"
            class="underline">api.ip.sb</a></div>
      </div>
    </DialogContent>
  </Dialog>
</template>
