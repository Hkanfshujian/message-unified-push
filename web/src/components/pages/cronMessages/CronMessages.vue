<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import ClickableTruncate from '@/components/ui/ClickableTruncate.vue'
import AddCronMessages from './AddCronMessages.vue'
import EditCronMessages from './EditCronMessages.vue'
import { toast } from 'vue-sonner'

import { useRoute, useRouter } from 'vue-router';
import { request } from '@/api/api';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface CronMessageItem {
  id: string
  name: string
  cron: string
  cron_expression: string
  template_id: string
  template_name: string
  ins_ids: string[]
  channel_names: string[]
  enable: number
  status: boolean
  created_on: string
  modified_on: string
  next_time?: string
}

const route = useRoute();
const router = useRouter();

let state = reactive({
  tableData: [] as CronMessageItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
})

// 状态过滤
const selectedStatus = ref('all')


// 新增定时消息 Dialog 状态
const isAddCronMessageDialogOpen = ref(false)

// 编辑定时消息 Dialog 状态
const isEditCronMessageDialogOpen = ref(false)
const editCronMessageData = ref<CronMessageItem | null>(null)
const isDeleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<CronMessageItem | null>(null)

// 处理保存新定时消息
const handleSaveCronMessage = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

// 打开编辑定时消息Dialog
const openEditCronMessageDialog = (cronMessage: CronMessageItem) => {
  editCronMessageData.value = cronMessage
  isEditCronMessageDialogOpen.value = true
}

// 处理编辑定时消息保存
const handleEditCronMessage = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

const openDeleteConfirm = (cronMessage: CronMessageItem) => {
  deleteTarget.value = cronMessage
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

// 处理查看日志
const handleViewLogs = (cronMessage: CronMessageItem) => {
  // 跳转到定时消息日志页面，携带cronMessageId参数
  router.push(`/sendlogs?taskid=${cronMessage.id}`)
}

// 切换状态
const toggleStatus = async (cronMessage: CronMessageItem) => {
  const prevStatus = cronMessage.enable
  const newStatus = prevStatus ? 0 : 1
  cronMessage.enable = newStatus
  const rsp = await request.post('/cronmessages/edit', cronMessage)
  if (rsp.data.code === 200) {
    toast.success(newStatus === 1 ? `已启用定时消息「${cronMessage.name}」` : `已停用定时消息「${cronMessage.name}」`)
    return
  }
  cronMessage.enable = prevStatus
  toast.error(rsp.data.msg || '更新定时消息状态失败')
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

//触发过滤筛选
const filterFunc = async () => {
  await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);
}

// 查询数据（包含状态过滤）
const queryListDataWithStatus = async () => {
  const statusParam = selectedStatus.value === 'all' ? '' : selectedStatus.value;
  await queryListData(state.currPage, state.pageSize, state.search, '', '', statusParam);
}

const queryListData = async (page: number, size: number, name = '', taskType = '', query = '', status = '') => {
  let params: any = { page: page, size: size, name: name, type: taskType, query: query };
  if (status !== '') {
    params.status = status;
  }
  const rsp = await request.get('/cronmessages/list', { params: params });
  state.tableData = rsp?.data?.data?.lists || [];
  state.total = rsp?.data?.data?.total || 0;
}

// 删除定时消息
const handleDelete = async (id: string) => {
  const rsp = await request.post('/cronmessages/delete', { id: id });
  if (rsp.status == 200 && await rsp.data.code == 200) {
    toast.success(rsp.data.msg);
    setTimeout(() => {
      window.location.reload();
    }, 1000);
  }
}

const handleConfirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  await handleDelete(deleteTarget.value.id)
  closeDeleteConfirm()
}


onMounted(async () => {
  // 初始化查询
  state.search = route.query.name?.toString() || '';
  await queryListData(
    1,
    state.pageSize,
    route.query.name?.toString() || '',
    route.query.task_type?.toString() || ''
  );
});

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
      </div>
      <Dialog v-model:open="isAddCronMessageDialogOpen">
        <DialogTrigger as-child>
          <Button v-permission="'message:cron:add'" variant="default" class="primary-btn">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            新增任务
          </Button>
        </DialogTrigger>

        <DialogContent class="w-[500px] max-w-[90vw]">
          <DialogHeader>
            <DialogTitle>新增定时消息</DialogTitle>
          </DialogHeader>

          <div class="px-4 pb-4">
            <AddCronMessages v-model:open="isAddCronMessageDialogOpen" @save="handleSaveCronMessage"
              @cancel="() => isAddCronMessageDialogOpen = false" />
          </div>
        </DialogContent>
      </Dialog>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <div class="min-w-full">
        <Table class="data-table border-collapse">
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">ID</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>模板</TableHead>
              <TableHead>渠道</TableHead>
              <TableHead>Cron表达式</TableHead>
              <TableHead>下次执行时间</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead class="text-center">操作</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            <!-- 空数据展示 -->
            <TableRow v-if="(state.tableData || []).length === 0">
              <TableCell colspan="8" class="empty-state">
                <EmptyTableState title="暂无定时消息" description="还没有配置任何定时消息，请先添加定时消息" />
              </TableCell>
            </TableRow>

            <!-- 数据行 -->
            <TableRow v-for="cronMessage in (state.tableData || [])" :key="cronMessage.id">
              <TableCell>{{ cronMessage.id }}</TableCell>
              <TableCell>
                <ClickableTruncate :text="cronMessage.name" wrapper-class="max-w-[180px] sm:max-w-[260px]" preview-title="名称" />
              </TableCell>
              <TableCell>
                <ClickableTruncate :text="cronMessage.template_name || cronMessage.template_id" wrapper-class="max-w-[180px] sm:max-w-[240px]" preview-title="模板" />
              </TableCell>
              <TableCell>
                <ClickableTruncate :text="(cronMessage.channel_names || []).join('、')" wrapper-class="max-w-[220px] sm:max-w-[320px]" preview-title="渠道" />
              </TableCell>
              <TableCell>
                <code class="max-w-[90px] sm:max-w-[9px] rounded text-sm font-mono bg-muted text-foreground border border-border">
                  {{ cronMessage.cron }}
                </code>
              </TableCell>
              <TableCell>{{ cronMessage.next_time || '-' }}</TableCell>
              <TableCell>{{ cronMessage.created_on }}</TableCell>
              <!-- <TableCell class="text-center">
                <Badge :class="cronMessage.status === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600'">
                  {{ getStatusText(cronMessage.status) }}
                </Badge>
              </TableCell> -->
              <TableCell class="text-center space-x-2">
                <Button size="sm" variant="outline" @click="handleViewLogs(cronMessage)">日志</Button>
                <Button v-permission="'message:cron:edit'" size="sm" variant="outline" @click="openEditCronMessageDialog(cronMessage)">编辑</Button>
                <Button v-permission="'message:cron:delete'" size="sm" variant="outline" class="text-red-500 border-red-300 hover:bg-red-50 
                  hover:border-red-400 hover:text-red-600 hover:shadow-md
                   transition-all duration-200" @click="openDeleteConfirm(cronMessage)">删除</Button>
                <Switch v-permission="'message:cron:edit'" :model-value="cronMessage.enable === 1" @update:model-value="toggleStatus(cronMessage)" />

              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
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

    <!-- 编辑定时消息Dialog -->
    <Dialog v-model:open="isEditCronMessageDialogOpen">
      <DialogContent class="w-[500px] max-w-[90vw] max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader class="flex-shrink-0">
          <DialogTitle>编辑定时消息</DialogTitle>
        </DialogHeader>

        <div class="px-4 pb-4 flex-1 overflow-y-auto">
          <EditCronMessages v-model:open="isEditCronMessageDialogOpen" :cron-message="editCronMessageData"
            @save="handleEditCronMessage" />
        </div>
      </DialogContent>
    </Dialog>

    <Dialog :open="isDeleteConfirmOpen" @update:open="(value) => value ? (isDeleteConfirmOpen = true) : closeDeleteConfirm()">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除定时消息</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">
            请输入要删除的定时消息名称
            <span v-if="deleteTarget?.name" class="text-red-500 font-semibold mx-1">{{ deleteTarget.name }}</span>
            以确认操作
          </div>
          <Input
            v-model="deleteConfirmInput"
            :max-length="50"
            placeholder="请输入定时消息名称"
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
  </div>
</template>
