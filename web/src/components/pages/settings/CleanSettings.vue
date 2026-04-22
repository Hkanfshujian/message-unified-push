<script setup lang="ts">
import { reactive, onMounted, ref } from 'vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { useRouter } from 'vue-router'
import { CONSTANT } from '@/constant'

const router = useRouter()

type CleanupState = {
  section: string
  name: string
  cron: string
  cronRemark: string
  keepNum: string
  enabled: boolean
  defaultCron: string
  defaultKeepNum: string
  viewPath: string
  viewQuery?: Record<string, string>
}

const taskLogsState = reactive<CleanupState>({
  section: 'log_config',
  name: '任务日志',
  cron: '',
  cronRemark: '',
  keepNum: '1000',
  enabled: true,
  defaultCron: '1 0 * * *',
  defaultKeepNum: '1000',
  viewPath: '/logs/task',
  viewQuery: { taskid: CONSTANT.LOG_TASK_ID },
})

const consumeLogsState = reactive<CleanupState>({
  section: 'consume_log_config',
  name: '消费日志',
  cron: '',
  cronRemark: '',
  keepNum: '1000',
  enabled: false,
  defaultCron: '1 0 * * *',
  defaultKeepNum: '1000',
  viewPath: '/logs/consume',
})

const loginLogsState = reactive<CleanupState>({
  section: 'login_log_config',
  name: '登录日志',
  cron: '',
  cronRemark: '',
  keepNum: '1000',
  enabled: false,
  defaultCron: '1 0 * * *',
  defaultKeepNum: '1000',
  viewPath: '/logs/login',
})

const cleanupCards = [taskLogsState, consumeLogsState, loginLogsState]
const editDialogOpen = ref(false)
const currentEditCard = ref<CleanupState | null>(null)
const editForm = reactive({
  cron: '',
  cronRemark: '',
})

const cronQuickTemplates = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每5分钟', value: '*/5 * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天凌晨2点', value: '0 2 * * *' },
  { label: '每周一凌晨2点', value: '0 2 * * 1' },
  { label: '每月1号凌晨2点', value: '0 2 1 * *' },
]

const buildPostData = (state: CleanupState) => {
  return {
    section: state.section,
    data: {
      cron: state.cron.trim(),
      cron_remark: state.cronRemark.trim(),
      keep_num: state.keepNum.trim(),
      enabled: state.enabled ? 'true' : 'false',
    },
  }
}

const saveCleanupConfig = async (state: CleanupState, source: 'button' | 'switch' = 'button') => {
  try {
    const postData = buildPostData(state)
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      if (source === 'switch') {
        const statusText = state.enabled ? '启用' : '停用'
        toast.success(`${state.name}清理已${statusText}，配置已保存`)
      } else {
        toast.success(`${state.name}清理配置已保存并生效`)
      }
    }
  } catch (error) {
    toast.error(`${state.name}清理配置保存失败，请稍后重试`)
  }
}

const handleViewLogs = (state: CleanupState) => {
  router.push({ path: state.viewPath, query: state.viewQuery || {} })
}

const openEditDialog = (state: CleanupState) => {
  currentEditCard.value = state
  editForm.cron = state.cron || state.defaultCron
  editForm.cronRemark = state.cronRemark || ''
  editDialogOpen.value = true
}

const saveEditDialog = async () => {
  const target = currentEditCard.value
  if (!target) {
    return
  }
  target.cron = editForm.cron.trim() || target.defaultCron
  target.cronRemark = editForm.cronRemark.trim()
  await saveCleanupConfig(target, 'button')
  editDialogOpen.value = false
}

const applyCronTemplate = (cronValue: string) => {
  editForm.cron = cronValue
}

const loadCleanupConfig = async (state: CleanupState) => {
  try {
    const params = { params: { section: state.section } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data || {}
      Object.assign(state, {
        cron: data.cron || state.defaultCron,
        cronRemark: data.cron_remark || '',
        keepNum: data.keep_num || state.defaultKeepNum,
        enabled: data.enabled === 'true' || data.enabled === true,
      })
    }
  } catch (error) {
    toast.error(`获取${state.name}清理配置失败`)
  }
}

onMounted(async () => {
  await Promise.all(cleanupCards.map((card) => loadCleanupConfig(card)))
})
</script>

<script lang="ts">
export default {
  name: 'CleanSettings'
}
</script>

<template>
  <div class="space-y-5">
    <div>
      <div class="text-lg font-semibold">数据清理设置</div>
      <div class="text-sm text-muted-foreground">按日志类型分别配置自动清理策略</div>
    </div>

    <div class="overflow-hidden rounded-lg border border-slate-300/80 dark:border-slate-700 bg-white/70 dark:bg-slate-900/30">
      <Table class="data-table border-collapse">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[80px] text-center">序号</TableHead>
            <TableHead class="w-[220px] text-center">任务名称</TableHead>
            <TableHead class="text-center">cron表达式</TableHead>
            <TableHead class="text-center">表达式备注</TableHead>
            <TableHead class="w-[340px] text-center">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(card, index) in cleanupCards" :key="card.section">
            <TableCell class="font-medium">{{ index + 1 }}</TableCell>
            <TableCell class="font-medium">{{ card.name }}清理</TableCell>
            <TableCell>
              <span class="font-mono text-sm">{{ card.cron || card.defaultCron }}</span>
            </TableCell>
            <TableCell>
              <span class="text-sm text-muted-foreground">{{ card.cronRemark || '-' }}</span>
            </TableCell>
            <TableCell class="whitespace-nowrap">
              <div class="grid grid-cols-3 items-center justify-items-center w-full gap-2">
                <Button variant="outline" size="sm" @click="handleViewLogs(card)">查看日志</Button>
                <Button size="sm" @click="openEditDialog(card)">编辑</Button>
                <div class="inline-flex items-center gap-2 justify-center">
                  <Switch v-model="card.enabled" @update:model-value="() => saveCleanupConfig(card, 'switch')" />
                  <span class="text-xs" :class="card.enabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-muted-foreground'">
                    {{ card.enabled ? '启用' : '停用' }}
                  </span>
                </div>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <Dialog v-model:open="editDialogOpen">
      <DialogContent class="max-w-[720px]">
        <DialogHeader>
          <DialogTitle>编辑清理任务：{{ currentEditCard?.name || '' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <div class="text-sm font-medium">Cron表达式</div>
            <Input v-model="editForm.cron" placeholder="请输入 cron，如：0 2 * * *" />
            <div class="flex flex-wrap gap-2 pt-1">
              <Button
                v-for="item in cronQuickTemplates"
                :key="item.value"
                size="sm"
                variant="outline"
                @click="applyCronTemplate(item.value)"
              >
                {{ item.label }}
              </Button>
            </div>
          </div>
          <div class="space-y-2">
            <div class="text-sm font-medium">表达式备注</div>
            <Input v-model="editForm.cronRemark" placeholder="手动输入备注信息（例如：每日凌晨执行清理）" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="editDialogOpen = false">取消</Button>
          <Button @click="saveEditDialog">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
