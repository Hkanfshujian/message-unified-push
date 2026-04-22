<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '@/components/ui/dialog'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import AddWays from './AddWays.vue'
import EditWays from './EditWays.vue'
import { toast } from 'vue-sonner'

import { useRoute } from 'vue-router';
import { request } from '@/api/api';
import { CONSTANT } from '@/constant';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface WayItem {
  id: number
  name: string
  type: string
  config: string
  created_on: string
  modified_on: string
  status: number
}

const router = useRoute();

let state = reactive({
  tableData: [] as WayItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
})

// 状态过滤
const selectedStatus = ref('all')

// 渠道类型过滤
const selectedChannelType = ref('all')

// 渠道类型选项 - 从CONSTANT.WAYS_DATA生成
const channelTypeOptions = computed(() => {
  const options = [{ value: 'all', label: '全部类型' }]
  CONSTANT.WAYS_DATA.forEach(item => {
    options.push({ value: item.type, label: item.label })
  })
  return options
})

// Sheet 相关状态
const isSheetOpen = ref(false)
const selectedConfig = ref('')
const selectedChannelName = ref('')

// 新增渠道 Sheet 状态
const isAddChannelDrawerOpen = ref(false)

// 编辑渠道 Sheet 状态
const isEditChannelDrawerOpen = ref(false)
const editChannelData = ref<WayItem | null>(null)
const isDeleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<WayItem | null>(null)

// 处理保存新渠道
const handleSaveChannel = (_data: any) => {
  // 这里可以添加实际的保存逻辑
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

const getWayTypeText = (type: string) => {
  const wayData = CONSTANT.WAYS_DATA.find(item => item.type === type)
  return wayData ? wayData.label : type
}

// 打开编辑渠道Drawer
const openEditChannelDrawer = (channel: WayItem) => {
  editChannelData.value = channel
  isEditChannelDrawerOpen.value = true
}

// 处理编辑渠道保存
const handleEditChannel = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

const openDeleteConfirm = (channel: WayItem) => {
  deleteTarget.value = channel
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

// 按渠道类型过滤
const filterByChannelType = async (value: any) => {
  if (value) {
    selectedChannelType.value = String(value);
    state.currPage = 1; // 重置到第一页
    await queryListDataWithStatus();
  }
}

// 查询数据（包含状态过滤）
const queryListDataWithStatus = async () => {
  const statusParam = selectedStatus.value === 'all' ? '' : selectedStatus.value;
  const channelTypeParam = selectedChannelType.value === 'all' ? '' : selectedChannelType.value;
  await queryListData(state.currPage, state.pageSize, state.search, channelTypeParam, '', statusParam);
}

const queryListData = async (page: number, size: number, name = '', channelType = '', query = '', status = '') => {
  let params: any = { page: page, size: size, name: name, type: channelType, query: query };
  if (status !== '') {
    params.status = status;
  }
  const rsp = await request.get('/sendways/list', { params: params });
  state.tableData = await rsp.data.data.lists;
  state.total = await rsp.data.data.total;
}
// 删除渠道
const handleDelete = async (id: number) => {
  const rsp = await request.post('/sendways/delete', { id: id });
  if (rsp.status == 200 && await rsp.data.code == 200) {
    // state.tableData.splice(index, 1);
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
  state.search = router.query.name?.toString() || '';
  await queryListData(
    1,
    state.pageSize,
    router.query.name?.toString() || '',
    router.query.channel_type?.toString() || ''
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

        <Select v-model="selectedChannelType" class="w-full" @update:model-value="filterByChannelType">
          <SelectTrigger class="filter-select w-full">
            <SelectValue placeholder="选择渠道类型" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem v-for="option in channelTypeOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <Dialog v-model:open="isAddChannelDrawerOpen">
        <DialogTrigger as-child>
          <Button v-permission="'message:sendways:add'" variant="default" class="primary-btn">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            新增渠道
          </Button>
        </DialogTrigger>

        <DialogContent
          class="!max-w-none w-[90vw] lg:w-[60vw] h-[80vh] flex flex-col p-0 gap-0">
          <DialogHeader class="p-6 pb-3 border-b text-left">
            <DialogTitle class="text-[18px]">新增发信渠道</DialogTitle>
          </DialogHeader>

          <div class="flex-1 overflow-y-auto px-6 pb-6">
            <AddWays v-model:open="isAddChannelDrawerOpen" @save="handleSaveChannel" />
          </div>
        </DialogContent>
      </Dialog>
    </div>

    <!-- 表格 -->
    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <Table class="data-table border-collapse">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>渠道名称</TableHead>
          <TableHead>发信方式类型</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">创建时间</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">更新时间</TableHead>
          <TableHead class="text-center">操作/状态</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="6" class="empty-state">
            <EmptyTableState title="暂无发信方式" description="还没有配置任何发信方式，请先添加发信方式" />
          </TableCell>
        </TableRow>

        <!-- 数据行 -->
        <TableRow v-for="channel in state.tableData" :key="channel.id">
          <TableCell>{{ channel.id }}</TableCell>
          <TableCell>{{ channel.name }}</TableCell>
          <TableCell>
            <Badge variant="outline">{{ getWayTypeText(channel.type) }}</Badge>
          </TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ channel.created_on }}</TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ channel.modified_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button v-permission="'message:sendways:edit'" size="sm" variant="outline" @click="openEditChannelDrawer(channel)">编辑</Button>
            <!-- <Button size="sm" variant="outline" @click="openConfigSheet(channel)">查看</Button> -->
            <Button v-permission="'message:sendways:delete'" size="sm" variant="outline" class="text-red-500 border-red-300 hover:bg-red-50 
              hover:border-red-400 hover:text-red-600 hover:shadow-md
               transition-all duration-200" @click="openDeleteConfirm(channel)">删除</Button>
            <!-- <Badge :class="channel.status === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600'">
              {{ getStatusText(channel.status) }} -->
            <!-- </Badge> -->
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

    <!-- 编辑渠道Dialog -->
    <Dialog v-model:open="isEditChannelDrawerOpen">
      <DialogContent class="sm:max-w-[800px] max-h-[90vh] flex flex-col p-0 gap-0">
        <DialogHeader class="p-6 pb-2">
          <DialogTitle>编辑发信渠道</DialogTitle>
        </DialogHeader>

        <div class="flex-1 overflow-y-auto px-6 pb-6">
          <EditWays v-model:open="isEditChannelDrawerOpen" :edit-data="editChannelData" @save="handleEditChannel" />
        </div>
      </DialogContent>
    </Dialog>

    <Dialog :open="isDeleteConfirmOpen" @update:open="(value) => value ? (isDeleteConfirmOpen = true) : closeDeleteConfirm()">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除渠道</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">请输入要删除的渠道名称以确认操作</div>
          <Input
            v-model="deleteConfirmInput"
            :max-length="50"
            placeholder="请输入渠道名称"
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

    <!-- 配置详情Sheet -->
    <Sheet v-model:open="isSheetOpen">
      <SheetContent class="w-[600px] sm:w-[900px] lg:w-[1000px]">
        <SheetHeader>
          <SheetTitle>{{ selectedChannelName }} - 发信方式配置详情</SheetTitle>
        </SheetHeader>
        <div class="mt-6">
          <div class="rounded-lg p-4 bg-muted/40 dark:bg-white/5 ring-1 ring-border/50 shadow-sm">
            <pre
              class="whitespace-pre-wrap text-sm font-mono leading-relaxed text-foreground">{{ selectedConfig }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
