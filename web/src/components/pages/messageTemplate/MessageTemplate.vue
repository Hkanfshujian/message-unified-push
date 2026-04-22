<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import ClickableTruncate from '@/components/ui/ClickableTruncate.vue'
import TemplateApiViewer from './TemplateApiViewer.vue'
import TemplateInstanceConfig from './TemplateInstanceConfig.vue'
import TemplateEditor from './TemplateEditor.vue'
import { request } from '@/api/api'
import { getPageSize } from '@/util/pageUtils'
import { toast } from 'vue-sonner'
import { useRouter } from 'vue-router'

interface MessageTemplate {
  id: string  // 模板ID是字符串类型（UUID）
  name: string
  description: string
  text_template: string
  html_template: string
  markdown_template: string
  placeholders: string
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status: string
  created_on: string
  modified_on: string
  cron_msg_count?: number
}

const router = useRouter()

let state = reactive({
  tableData: [] as MessageTemplate[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize() as number,
  search: '',
  status: 'all'
})

// API代码查看器状态
const isApiViewerOpen = ref(false)
const selectedTemplateForApi = ref<MessageTemplate | null>(null)

// 配置实例状态
const isInstanceConfigOpen = ref(false)
const selectedTemplateForInstance = ref<MessageTemplate | null>(null)

// 模板编辑器状态
const isEditorOpen = ref(false)
const isEditing = ref(false)
const selectedTemplateForEdit = ref<MessageTemplate | null>(null)

const isDeleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<MessageTemplate | null>(null)

const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

const queryListData = async (page: number, size: number, text = '', status = '') => {
  const params: any = { page, size, text, status }
  const rsp = await request.get('/templates/list', { params })
  state.tableData = rsp.data.data.lists || []
  state.total = rsp.data.data.total || 0
}

const changePage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    state.currPage = page
    const statusParam = state.status === 'all' ? '' : state.status
    await queryListData(state.currPage, state.pageSize, state.search, statusParam)
  }
}

const handlePageSizeChange = async (size: number) => {
  if (size <= 0) return
  state.pageSize = size
  state.currPage = 1
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const filterFunc = async () => {
  state.currPage = 1
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const openAddDialog = () => {
  isEditing.value = false
  selectedTemplateForEdit.value = null
  isEditorOpen.value = true
}

const openEditDialog = (template: MessageTemplate) => {
  isEditing.value = true
  selectedTemplateForEdit.value = template
  isEditorOpen.value = true
}

const handleEditorSaved = async () => {
  // 刷新列表
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const deleteTemplate = async (id: string) => {
  const rsp = await request.post('/templates/delete', { id })
  if (rsp.status === 200 && rsp.data.code === 200) {
    toast.success(rsp.data.msg)
    // 刷新列表，处理status参数
    const statusParam = state.status === 'all' ? '' : state.status
    await queryListData(state.currPage, state.pageSize, state.search, statusParam)
  }
}

const openDeleteConfirm = (template: MessageTemplate) => {
  // 如果存在外部关联（目前主要是定时消息），先提示并展示关联列表，不进入删除确认流程
  if ((template.cron_msg_count ?? 0) > 0) {
    toast.error('当前模板存在外部关联，请先在相关功能中删除关联后再删除模板')
    openRelationDialog(template)
    return
  }

  deleteTarget.value = template
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

const handleConfirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  await deleteTemplate(deleteTarget.value.id)
  closeDeleteConfirm()
}

const isRelationDialogOpen = ref(false)
const relationList = ref<Array<{ index: number; type: string; id: string; name: string }>>([])
const relationLoading = ref(false)
const relationTemplateName = ref('')

const openRelationDialog = async (template: MessageTemplate) => {
  relationTemplateName.value = template.name
  relationList.value = []
  relationLoading.value = true
  isRelationDialogOpen.value = true
  try {
    const rsp = await request.get('/templates/relations', { params: { id: template.id } })
    const list = rsp.data.data?.relations || []
    relationList.value = list.map((item: any, idx: number) => ({
      index: idx + 1,
      type: item.type || '定时消息',
      id: item.id || '',
      name: item.name || ''
    }))
  } catch (error) {
    toast.error('获取关联信息失败')
  } finally {
    relationLoading.value = false
  }
}

// 打开API查看器
const handleViewApi = (template: MessageTemplate) => {
  selectedTemplateForApi.value = template
  isApiViewerOpen.value = true
}

// 打开配置实例
const handleConfigInstance = (template: MessageTemplate) => {
  selectedTemplateForInstance.value = template
  isInstanceConfigOpen.value = true
}

// 查看日志
const handleViewLogs = (template: MessageTemplate) => {
  // 跳转到发信日志页面，携带 taskid 参数（传递模板 id）
  router.push(`/logs/task?taskid=${template.id}`)
}

onMounted(async () => {
  await queryListData(1, state.pageSize)
})
</script>

<template>
  <div class="space-y-2">
    <div class="toolbar">
      <div class="search-group">
        <Input
          v-model="state.search"
          placeholder="搜索..."
          class="search-input"
          @keyup.enter="filterFunc"
          @blur="filterFunc"
        />

        <Select v-model="state.status" class="w-full" @update:model-value="filterFunc">
          <SelectTrigger class="filter-select w-full">
            <SelectValue placeholder="选择状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="enabled">启用</SelectItem>
              <SelectItem value="disabled">禁用</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <Button v-permission="'message:template:add'" @click="openAddDialog" class="primary-btn">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        新建模板
      </Button>
    </div>

    <!-- 表格 -->
    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
      <TableHeader>
        <TableRow>
          <TableHead class="w-24">ID</TableHead>
          <TableHead class="w-[288px] min-w-[288px] whitespace-normal">模板名称</TableHead>
          <TableHead class="w-[240px]">描述</TableHead>
          <TableHead class="w-[200px]">支持格式</TableHead>
          <TableHead class="w-[140px] whitespace-nowrap">外部关联</TableHead>
          <TableHead>状态</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">创建时间</TableHead>
          <TableHead class="text-center">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="7" class="empty-state">
            <EmptyTableState 
              title="暂无消息模板" 
              description="还没有创建任何消息模板，点击右上角按钮创建新模板" 
            >
              <template #icon>
                <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                </svg>
              </template>
            </EmptyTableState>
          </TableCell>
        </TableRow>
        
        <!-- 数据行 -->
        <TableRow v-for="item in state.tableData" :key="item.id">
          <TableCell>{{ item.id }}</TableCell>
          <TableCell class="table-cell w-[288px] min-w-[288px] whitespace-normal">
            {{ item.name }}
          </TableCell>
          <TableCell>
            <ClickableTruncate :text="item.description || '-'" wrapper-class="max-w-[80px] sm:max-w-[130px]" preview-title="模板描述" />
          </TableCell>
          <TableCell>
            <div class="flex gap-1">
              <Badge v-if="item.text_template" variant="secondary">Text</Badge>
              <Badge v-if="item.html_template" variant="secondary">HTML</Badge>
              <Badge v-if="item.markdown_template" variant="secondary">Markdown</Badge>
            </div>
          </TableCell>
          <TableCell class="whitespace-nowrap">
            <Button
              size="sm"
              variant="ghost"
              class="h-7 px-2 text-xs text-brand-600 hover:text-brand-700 hover:bg-brand-50"
              @click="openRelationDialog(item)"
            >
              {{ item.cron_msg_count ?? 0 }}
            </Button>
          </TableCell>
          <TableCell>
            <Badge :variant="item.status === 'enabled' ? 'default' : 'secondary'">
              {{ item.status === 'enabled' ? '启用' : '禁用' }}
            </Badge>
          </TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ item.created_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="handleViewLogs(item)">日志</Button>
            <Button size="sm" variant="outline" @click="handleViewApi(item)">接口</Button>
            <Button v-permission="'message:template:edit'" size="sm" variant="outline" @click="openEditDialog(item)">编辑</Button>
            <Button v-permission="'message:template:instance'" size="sm" variant="outline" @click="handleConfigInstance(item)">实例</Button>
            <Button v-permission="'message:template:delete'" size="sm" variant="destructive" @click="openDeleteConfirm(item)">删除</Button>
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

    <!-- 模板编辑器 -->
    <TemplateEditor
      :open="isEditorOpen"
      :is-editing="isEditing"
      :template-data="selectedTemplateForEdit"
      @update:open="isEditorOpen = $event"
      @saved="handleEditorSaved"
    />

    <!-- API代码查看器 -->
    <TemplateApiViewer 
      :open="isApiViewerOpen" 
      :template-data="selectedTemplateForApi || undefined"
      @update:open="isApiViewerOpen = $event"
    />

    <!-- 配置实例 -->
    <TemplateInstanceConfig 
      :open="isInstanceConfigOpen" 
      :template-data="selectedTemplateForInstance"
      @update:open="isInstanceConfigOpen = $event"
    />

    <Dialog :open="isDeleteConfirmOpen" @update:open="(value) => value ? (isDeleteConfirmOpen = true) : closeDeleteConfirm()">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除模板</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">
            请输入要删除的模板名称
            <span v-if="deleteTarget?.name" class="text-red-500 font-semibold mx-1">{{ deleteTarget.name }}</span>
            以确认操作
          </div>
          <Input
            v-model="deleteConfirmInput"
            :max-length="25"
            placeholder="请输入模板名称"
            class="confirm-delete-input"
          />
          <div v-if="showDeleteError" class="error-tip">名称不匹配，请重新输入</div>
        </div>
        <DialogFooter class="flex justify-end gap-2 mt-4">
          <Button type="button" class="cancel-btn" @click="closeDeleteConfirm">取消</Button>
          <Button type="button" class="danger-btn" :disabled="!isDeleteMatch" @click="handleConfirmDelete">
            确认删除
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="isRelationDialogOpen" @update:open="(value) => isRelationDialogOpen = value">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>模板关联信息：{{ relationTemplateName }}</DialogTitle>
        </DialogHeader>
        <div v-if="relationLoading" class="py-4 text-sm text-muted-foreground">
          加载中...
        </div>
        <div v-else class="max-h-[320px] overflow-auto mt-2">
          <Table class="text-xs">
            <TableHeader>
              <TableRow>
                <TableHead class="w-12">序号</TableHead>
                <TableHead class="w-24">关联类型</TableHead>
                <TableHead class="w-32">ID</TableHead>
                <TableHead>名称</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="relationList.length === 0">
                <TableCell colspan="4" class="text-center text-muted-foreground py-4">
                  暂无关联记录
                </TableCell>
              </TableRow>
              <TableRow v-for="item in relationList" :key="item.index">
                <TableCell>{{ item.index }}</TableCell>
                <TableCell>{{ item.type }}</TableCell>
                <TableCell>{{ item.id }}</TableCell>
                <TableCell>{{ item.name }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        <DialogFooter>
          <Button type="button" size="sm" variant="outline" @click="isRelationDialogOpen = false">
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
