<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { Download, Refresh, Search } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import AppActionButton from '@/components/ui/AppActionButton.vue'
import AppDateTimeRange from '@/components/ui/AppDateTimeRange.vue'
import AppDetailDrawer from '@/components/ui/AppDetailDrawer.vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTruncate from '@/components/ui/AppTruncate.vue'
import { loginLogsApi } from '@/api/logs'
import { getPageSize } from '@/util/pageUtils'
import { downloadBlob, notifyError, notifySuccess } from '@/util/uiFeedback'

interface LoginLog {
  id: number
  user_id: number
  username: string
  ip: string
  ua: string
  created_on: string
}

const route = useRoute()
const router = useRouter()

const getTodayRange = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return [`${year}-${month}-${day}T00:00`, `${year}-${month}-${day}T23:59`] as [string, string]
}

const loading = ref(false)
const ipDialogOpen = ref(false)
const ipLoading = ref(false)
const selectedIp = ref('')
const ipInfo = ref<Record<string, unknown> | null>(null)
const search = ref('')
const timeRange = ref<[string, string] | []>(getTodayRange())

const state = reactive({
  logs: [] as LoginLog[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize()
})

const columns: AppTableColumn[] = [
  { prop: 'username', label: '用户名', minWidth: 160 },
  { prop: 'ip', label: 'IP', minWidth: 150 },
  { prop: 'ua', label: 'User-Agent', minWidth: 360 },
  { prop: 'created_on', label: '登录时间', minWidth: 170 }
]

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.floor(n)
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {
    page: String(state.currPage),
    page_size: String(state.pageSize)
  }
  const [startTime, endTime] = timeRange.value
  if (startTime) nextQuery.start_time = startTime
  if (endTime) nextQuery.end_time = endTime
  const keyword = search.value.trim()
  if (keyword) nextQuery.keyword = keyword
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

const buildParams = () => {
  const params: Record<string, unknown> = {
    page: state.currPage,
    page_size: state.pageSize
  }
  const [startTime, endTime] = timeRange.value
  if (startTime) params.start_time = startTime
  if (endTime) params.end_time = endTime
  return params
}

const fetchLogs = async () => {
  loading.value = true
  try {
    await syncRouteQuery()
    const rsp = await loginLogsApi.recent(buildParams())
    const data = rsp.data
    if (data?.code === 200 && data.data) {
      state.logs = data.data.lists || []
      state.total = data.data.total || 0
      return
    }
    state.logs = []
    state.total = 0
    notifyError(data?.msg || '获取登录日志失败')
  } catch (error) {
    state.logs = []
    state.total = 0
    notifyError('获取登录日志时发生错误')
  } finally {
    loading.value = false
  }
}

const filterFunc = async () => {
  state.currPage = 1
  await fetchLogs()
}

const handlePaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await fetchLogs()
}

const clearTimeFilter = async () => {
  timeRange.value = getTodayRange()
  state.currPage = 1
  await fetchLogs()
}

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
    ipInfo.value = await rsp.json()
  } catch (error) {
    notifyError('获取IP信息失败')
  } finally {
    ipLoading.value = false
  }
}

const formatUtcOffset = (offset: unknown) => {
  const num = typeof offset === 'number' ? offset : parseInt(String(offset), 10)
  if (Number.isNaN(num)) return '-'
  const hours = Math.floor(num / 3600)
  const sign = hours >= 0 ? '+' : '-'
  return `UTC${sign}${Math.abs(hours)}`
}

const ipDisplayRows = computed(() => {
  if (!ipInfo.value) return [] as Array<{ label: string, value: string }>
  const info = ipInfo.value
  const rows: Array<{ label: string, value: string }> = []
  const push = (label: string, value: unknown) => {
    if (value !== undefined && value !== null && String(value)) rows.push({ label, value: String(value) })
  }
  const country = [info.country, info.country_code ? `(${info.country_code})` : ''].filter(Boolean).join(' ')
  const tz = [info.timezone, info.offset != null ? formatUtcOffset(info.offset) : ''].filter(Boolean).join(' · ')
  const asn = [info.asn, info.asn_organization].filter(Boolean).join(' - ')
  const coord = [info.latitude, info.longitude].filter(v => v !== undefined && v !== null).join(', ')
  push('IP', info.ip || selectedIp.value)
  push('国家/地区', country)
  push('时区', tz)
  push('ISP', info.isp)
  push('组织', info.organization)
  push('ASN', asn)
  push('大洲', info.continent_code)
  push('坐标', coord)
  return rows
})

const handleExport = async () => {
  try {
    const rsp = await loginLogsApi.export(buildParams())
    downloadBlob(rsp.data, `loginlogs-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('登录日志导出成功')
  } catch (error) {
    notifyError('登录日志导出失败')
  }
}

onMounted(async () => {
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.page_size, state.pageSize)
  const startTime = route.query.start_time?.toString()
  const endTime = route.query.end_time?.toString()
  if (startTime || endTime) timeRange.value = [startTime || '', endTime || '']
  search.value = route.query.keyword?.toString() || ''
  await fetchLogs()
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
        <div class="flex flex-1 flex-col gap-3 lg:flex-row lg:items-center">
          <el-input v-model="search" class="lg:max-w-[260px]" clearable placeholder="搜索用户名/IP/UA..." @keyup.enter="filterFunc" @clear="filterFunc">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <AppDateTimeRange v-model="timeRange" class="lg:max-w-[360px]" @change="filterFunc" />
          <el-button @click="clearTimeFilter">重置时间</el-button>
          <el-button :icon="Search" :loading="loading" @click="filterFunc">查询</el-button>
          <el-button :icon="Refresh" :loading="loading" @click="fetchLogs">刷新</el-button>
        </div>
        <el-button :icon="Download" @click="handleExport">导出</el-button>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0">
      <AppTable :data="displayLogs as unknown as Record<string, unknown>[]" :columns="columns" :loading="loading" empty-text="当前没有可展示的登录日志记录">
        <template #ip="{ row }">
          <AppActionButton @click="openIpInfo(String(row.ip || ''))">{{ row.ip || '-' }}</AppActionButton>
        </template>
        <template #ua="{ row }">
          <AppTruncate :text="String(row.ua || '-')" title="User-Agent" width="760px" />
        </template>
        <template #created_on="{ row }">
          <span class="text-sm text-muted-foreground">{{ row.created_on || '-' }}</span>
        </template>
        <template #empty>
          <AppEmptyState description="当前没有可展示的登录日志记录" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />
  </div>

  <AppDetailDrawer v-model="ipDialogOpen" title="IP 信息" size="520px">
    <div v-if="ipLoading" class="py-8 text-center text-sm text-muted-foreground">加载中...</div>
    <div v-else class="space-y-3 text-sm">
      <el-descriptions v-if="ipDisplayRows.length" :column="1" border>
        <el-descriptions-item v-for="(row, idx) in ipDisplayRows" :key="idx" :label="row.label">
          <span class="break-all" :class="{ 'font-mono': row.label === 'IP' || row.label === '坐标' }">{{ row.value }}</span>
        </el-descriptions-item>
      </el-descriptions>
      <AppEmptyState v-else description="未获取到该 IP 的地理信息" />
      <div class="text-xs text-muted-foreground">
        数据来源：<a :href="`https://api.ip.sb/geoip/${encodeURIComponent(selectedIp)}`" target="_blank" rel="noreferrer" class="underline">api.ip.sb</a>
      </div>
    </div>
  </AppDetailDrawer>
</template>
