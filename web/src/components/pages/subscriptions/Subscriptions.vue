<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Download, Plus, Refresh, Search } from '@element-plus/icons-vue'
import AppEmptyState from '@/components/ui/AppEmptyState.vue'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppPagination from '@/components/ui/AppPagination.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import AppStatusTag from '@/components/ui/AppStatusTag.vue'
import AppTable, { type AppTableColumn } from '@/components/ui/AppTable.vue'
import AppTruncate from '@/components/ui/AppTruncate.vue'
import { mqApi } from '@/api/mq'
import { subscriptionsApi } from '@/api/subscriptions'
import { templatesApi } from '@/api/templates'
import { useRoute, useRouter } from 'vue-router'
import { getPageSize } from '@/util/pageUtils'
import { appendDateRangeQuery, pickDateRangeQuery } from '@/util/routeQuery'
import { confirmAction, downloadBlob, notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'
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
  loading: false,
})
const route = useRoute()
const router = useRouter()

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

const columns: AppTableColumn[] = [
  { prop: 'index', label: '序号', width: 80, align: 'center' },
  { prop: 'id', label: 'ID', minWidth: 130 },
  { prop: 'name', label: '订阅名称', minWidth: 180 },
  { prop: 'topic', label: 'Topic', minWidth: 170 },
  { prop: 'tag', label: 'Tag', minWidth: 120 },
  { prop: 'status', label: '状态', width: 100, align: 'center' },
  { prop: 'stats', label: '消费/发送/失败', width: 150, align: 'center' },
  { prop: 'last_consume_time', label: '最后消费时间', minWidth: 170 },
  { prop: 'actions', label: '操作', width: 190, align: 'center', fixed: 'right' }
]

const parsePositiveNumber = (value: unknown, fallback: number) => {
  const n = Number(value)
  if (!Number.isFinite(n) || n <= 0) return fallback
  return Math.floor(n)
}

const buildRouteQuery = () => {
  const nextQuery: Record<string, string> = {}
  const name = state.search.trim()
  if (name) nextQuery.name = name
  if (selectedSource.value && selectedSource.value !== 'all') nextQuery.source_id = selectedSource.value
  if (selectedStatus.value && selectedStatus.value !== 'all') nextQuery.status = selectedStatus.value
  nextQuery.page = String(state.currPage)
  nextQuery.page_size = String(state.pageSize)
  appendDateRangeQuery(nextQuery, route.query as Record<string, unknown>)
  return nextQuery
}

const syncRouteQuery = async () => {
  await router.replace({ path: route.path, query: buildRouteQuery() })
}

const formatLastConsumeTime = (value: string) => {
  if (!value) return '-'
  if (value.includes('0001-01-01')) return '-'
  return value
}

const getStatClass = (value: number, kind: 'consume' | 'sent' | 'failed') => {
  if (!value) return 'text-muted-foreground'
  if (kind === 'failed') return 'text-red-600'
  if (kind === 'sent') return 'text-green-700'
  return 'text-foreground'
}

const loadSourceOptions = async () => {
  try {
    const res = await mqApi.list({ page: 1, page_size: 100 })
    if (res.data.code === 200) {
      sourceOptions.value = (res.data.data.list || []).map((item: any) => ({
        id: item.id,
        name: item.name
      }))
    }
  } catch (error) {
    notifyError('加载数据源失败')
  }
}

const loadTemplateOptions = async () => {
  try {
    const res = await templatesApi.list({ page: 1, size: 100 })
    if (res.data.code === 200) {
      templateOptions.value = (res.data.data.lists || []).map((item: any) => ({
        id: item.id,
        name: item.name
      }))
    }
  } catch (error) {
    notifyError('加载模板失败')
  }
}

const openEditDialog = async (item: SubscriptionItem) => {
  try {
    const res = await subscriptionsApi.detail(item.id)
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
  }
  editData.value = {
    ...item,
    template_content_type: normalizeTemplateContentType(item.template_content_type || item.consume_mode)
  }
  isEditDialogOpen.value = true
}

const queryListData = async (page: number, pageSize: number, name: string, sourceId: string, status: string) => {
  state.loading = true
  try {
    const params: any = {
      page,
      page_size: pageSize,
    }
    if (name) params.name = name
    if (sourceId && sourceId !== 'all') params.source_id = sourceId
    if (status && status !== 'all') params.status = status
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime

    const res = await subscriptionsApi.list(params)
    if (res.data.code === 200) {
      state.tableData = res.data.data.list || []
      state.total = res.data.data.total || 0
    } else {
      state.tableData = []
      state.total = 0
      notifyError(res.data.msg || '获取订阅列表失败')
    }
  } catch (error) {
    state.tableData = []
    state.total = 0
    notifyError('获取订阅列表时发生错误')
  } finally {
    state.loading = false
  }
}

const queryListDataWithStatus = async () => {
  await queryListData(state.currPage, state.pageSize, state.search, selectedSource.value, selectedStatus.value)
}

const handlePaginationChange = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  state.currPage = page
  state.pageSize = pageSize
  await syncRouteQuery()
  await queryListDataWithStatus()
}

const filterFunc = async () => {
  state.currPage = 1
  await syncRouteQuery()
  await queryListDataWithStatus()
}

const filterBySource = async (value: any) => {
  if (value) {
    selectedSource.value = String(value)
    state.currPage = 1
    await syncRouteQuery()
    await queryListDataWithStatus()
  }
}

const filterByStatus = async (value: any) => {
  if (value) {
    selectedStatus.value = String(value)
    state.currPage = 1
    await syncRouteQuery()
    await queryListDataWithStatus()
  }
}

const handleToggleStatus = async (item: SubscriptionItem) => {
  try {
    const action = item.status === 'running' ? 'stop' : 'start'
    const res = await subscriptionsApi.setState(item.id, action)
    if (res.data.code === 200) {
      notifySuccess(item.status === 'running' ? '订阅已停止' : '订阅启动中...')
      await queryListDataWithStatus()
    } else {
      notifyError(res.data.msg || '操作失败')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '操作失败')
  }
}

const handleDelete = async (item: SubscriptionItem) => {
  if (item.status === 'running') {
    notifyWarning('请先停止订阅后再删除')
    return
  }

  try {
    await confirmAction(`确认删除订阅“${item.name}”？`, '删除订阅')
  } catch (error) {
    return
  }

  try {
    const res = await subscriptionsApi.remove(item.id)
    if (res.data.code === 200) {
      notifySuccess('删除成功')
      await queryListDataWithStatus()
    } else {
      notifyError(res.data.msg || '删除失败')
    }
  } catch (error: any) {
    notifyError(error.response?.data?.msg || '删除失败')
  }
}

const handleSaveSuccess = () => {
  isAddDialogOpen.value = false
  isEditDialogOpen.value = false
  queryListDataWithStatus()
}

const handleExport = async () => {
  try {
    const params: Record<string, unknown> = {
      page: state.currPage,
      page_size: state.pageSize
    }
    if (state.search.trim()) params.name = state.search.trim()
    if (selectedSource.value !== 'all') params.source_id = selectedSource.value
    if (selectedStatus.value !== 'all') params.status = selectedStatus.value
    const { startTime, endTime } = pickDateRangeQuery(route.query as Record<string, unknown>)
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime

    const rsp = await subscriptionsApi.export(params)
    downloadBlob(rsp.data, `subscriptions-${new Date().toISOString().slice(0, 10)}.csv`)
    notifySuccess('订阅数据导出成功')
  } catch (error) {
    notifyError('订阅数据导出失败')
  }
}

onMounted(() => {
  state.search = typeof route.query.name === 'string' ? route.query.name : ''
  selectedSource.value = typeof route.query.source_id === 'string' && route.query.source_id ? route.query.source_id : 'all'
  selectedStatus.value = typeof route.query.status === 'string' && route.query.status ? route.query.status : 'all'
  state.currPage = parsePositiveNumber(route.query.page, 1)
  state.pageSize = parsePositiveNumber(route.query.page_size, state.pageSize)
  syncRouteQuery()
  loadSourceOptions()
  loadTemplateOptions()
  queryListDataWithStatus()
})
</script>

<template>
  <div class="space-y-4">
    <el-card shadow="never">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-1 flex-col gap-3 md:flex-row md:items-center">
          <el-input
            v-model="state.search"
            class="md:max-w-[260px]"
            clearable
            placeholder="搜索订阅名称..."
            @keyup.enter="filterFunc"
            @clear="filterFunc"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="selectedSource" class="md:max-w-[220px]" filterable @change="filterBySource">
            <el-option label="全部数据源" value="all" />
            <el-option v-for="opt in sourceOptions" :key="opt.id" :label="opt.name" :value="opt.id" />
          </el-select>
          <el-select v-model="selectedStatus" class="md:max-w-[160px]" @change="filterByStatus">
            <el-option label="全部状态" value="all" />
            <el-option label="运行中" value="running" />
            <el-option label="已停止" value="stopped" />
          </el-select>
          <el-button :icon="Search" @click="filterFunc">查询</el-button>
          <el-button :icon="Refresh" @click="queryListDataWithStatus">刷新</el-button>
        </div>
        <div class="flex flex-wrap gap-2">
          <el-button :icon="Download" @click="handleExport">导出</el-button>
          <el-button v-permission="'data:subscription:add'" type="primary" :icon="Plus" @click="isAddDialogOpen = true">新增订阅</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" body-class="!p-0">
      <AppTable :data="state.tableData as unknown as Record<string, unknown>[]" :columns="columns" :loading="state.loading" empty-text="还没有配置任何订阅规则">
        <template #index="{ index }">
          {{ (state.currPage - 1) * state.pageSize + index + 1 }}
        </template>
        <template #id="{ row }">
          <AppTruncate :text="String(row.id || '-')" class="font-mono text-sm" />
        </template>
        <template #name="{ row }">
          <div class="font-medium">{{ row.name || '-' }}</div>
          <div class="text-xs text-muted-foreground">{{ row.source_name || '-' }}</div>
        </template>
        <template #topic="{ row }">
          <AppTruncate :text="String(row.topic || '-')" class="font-mono text-sm" />
        </template>
        <template #tag="{ row }">
          <span class="font-mono text-sm">{{ row.tag || '*' }}</span>
        </template>
        <template #status="{ row }">
          <AppStatusTag :status="row.status" />
        </template>
        <template #stats="{ row }">
          <span :class="getStatClass(Number(row.total_consumed || 0), 'consume')">{{ row.total_consumed || 0 }}</span>
          <span class="text-muted-foreground"> / </span>
          <span :class="getStatClass(Number(row.total_sent || 0), 'sent')">{{ row.total_sent || 0 }}</span>
          <span class="text-muted-foreground"> / </span>
          <span :class="getStatClass(Number(row.total_failed || 0), 'failed')">{{ row.total_failed || 0 }}</span>
        </template>
        <template #last_consume_time="{ row }">
          <span class="text-sm text-muted-foreground">{{ formatLastConsumeTime(String(row.last_consume_time || '')) }}</span>
        </template>
        <template #actions="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: '编辑', kind: 'write', permission: 'data:subscription:edit', onClick: () => openEditDialog(row as unknown as SubscriptionItem) },
            { key: 'delete', label: '删除', kind: 'write', permission: 'data:subscription:delete', danger: true, onClick: () => handleDelete(row as unknown as SubscriptionItem) },
            { key: 'toggle', label: row.status === 'running' ? '停止' : '启动', kind: 'write', permission: row.status === 'running' ? 'data:subscription:stop' : 'data:subscription:start', danger: row.status === 'running', onClick: () => handleToggleStatus(row as unknown as SubscriptionItem) }
          ]" />
        </template>
        <template #empty>
          <AppEmptyState description="还没有配置任何订阅规则" />
        </template>
      </AppTable>
    </el-card>

    <AppPagination v-model:current-page="state.currPage" v-model:page-size="state.pageSize" :total="state.total" @change="handlePaginationChange" />

    <AppFormDrawer v-model="isAddDialogOpen" title="新增订阅" size="min(980px, 96vw)" body-mode="managed" density="compact" :show-footer="false">
      <SubscriptionForm :source-options="sourceOptions" :template-options="templateOptions" @success="handleSaveSuccess" />
    </AppFormDrawer>

    <AppFormDrawer v-model="isEditDialogOpen" title="编辑订阅" size="min(980px, 96vw)" body-mode="managed" density="compact" :show-footer="false">
      <SubscriptionForm v-if="editData" :data="editData" :source-options="sourceOptions" :template-options="templateOptions" @success="handleSaveSuccess" @cancel="isEditDialogOpen = false" />
    </AppFormDrawer>
  </div>
</template>
