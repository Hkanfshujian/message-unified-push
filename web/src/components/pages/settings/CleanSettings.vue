<script setup lang="ts">
import { reactive, onMounted, ref } from 'vue'
import { settingsApi } from '@/api/settings'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
import { useRouter } from 'vue-router'
import { CONSTANT } from '@/constant'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'

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
    const response = await settingsApi.set(postData.section, postData.data)
    if (response.data.code === 200) {
      if (source === 'switch') {
        const statusText = state.enabled ? '启用' : '停用'
        notifySuccess(`${state.name}清理已${statusText}，配置已保存`)
      } else {
        notifySuccess(`${state.name}清理配置已保存并生效`)
      }
    }
  } catch (error) {
    notifyError(`${state.name}清理配置保存失败，请稍后重试`)
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
    const response = await settingsApi.get(params.params.section)
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
    notifyError(`获取${state.name}清理配置失败`)
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

    <el-card shadow="never" class="overflow-hidden">
      <el-table :data="cleanupCards" class="w-full">
        <el-table-column label="序号" width="80" align="center">
          <template #default="{ $index }">{{ $index + 1 }}</template>
        </el-table-column>
        <el-table-column label="任务名称" width="180" align="center">
          <template #default="{ row }"><span class="font-medium">{{ row.name }}清理</span></template>
        </el-table-column>
        <el-table-column label="cron表达式" min-width="180" align="center">
          <template #default="{ row }">
              <span class="font-mono text-sm">{{ row.cron || row.defaultCron }}</span>
          </template>
        </el-table-column>
        <el-table-column label="表达式备注" min-width="180" align="center">
          <template #default="{ row }">
              <span class="text-sm text-muted-foreground">{{ row.cronRemark || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
              <AppRowActions :actions="[
                { key: 'edit', label: '编辑', kind: 'write', permission: 'system:settings:edit', onClick: () => openEditDialog(row) },
                { key: 'disable', label: '停用', kind: 'write', permission: 'system:settings:edit', visible: row.enabled, danger: true, onClick: () => { row.enabled = false; return saveCleanupConfig(row, 'switch') } },
                { key: 'enable', label: '启用', kind: 'write', permission: 'system:settings:edit', visible: !row.enabled, onClick: () => { row.enabled = true; return saveCleanupConfig(row, 'switch') } },
                { key: 'logs', label: '查看日志', kind: 'view', onClick: () => handleViewLogs(row) }
              ]" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <AppFormDrawer v-model="editDialogOpen" :title="`编辑清理任务：${currentEditCard?.name || ''}`" size="720px" :show-footer="false">
        <div class="grid gap-4">
          <section class="app-form-section border border-[var(--line-weak)] p-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 class="app-form-section-title">执行计划</h3>
                <p class="app-form-section-description">配置 {{ currentEditCard?.name || '-' }}日志的自动清理时间。</p>
              </div>
              <div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                <span>状态：{{ currentEditCard?.enabled ? '已启用' : '已停用' }}</span>
                <span>保留：{{ currentEditCard?.keepNum || '-' }} 条</span>
              </div>
            </div>
            <div class="mt-4 grid gap-4">
              <div class="app-form-field">
                <div class="app-form-label">Cron表达式</div>
                <el-input v-model="editForm.cron" placeholder="请输入 cron，如：0 2 * * *" />
                <div class="flex flex-wrap gap-2">
                  <el-button v-for="item in cronQuickTemplates" :key="item.value" size="small" @click="applyCronTemplate(item.value)">{{ item.label }}</el-button>
                </div>
              </div>
              <div class="app-form-field">
                <div class="app-form-label">表达式备注</div>
                <el-input v-model="editForm.cronRemark" placeholder="手动输入备注信息（例如：每日凌晨执行清理）" />
              </div>
            </div>
          </section>
        </div>
        <template #footer>
          <el-button @click="editDialogOpen = false">取消</el-button>
          <el-button type="primary" @click="saveEditDialog">保存</el-button>
        </template>
    </AppFormDrawer>
  </div>
</template>
