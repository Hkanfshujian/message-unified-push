<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const loading = ref(false)
const saving = ref(false)

const state = reactive({
  section: 'mq_status_policy',
  enabled: 'false',
  interval_seconds: '300',
  log_level: 'info'
})

const enabledBool = computed({
  get: () => state.enabled === 'true',
  set: (val: boolean) => {
    state.enabled = val ? 'true' : 'false'
  }
})

const loadConfig = async () => {
  loading.value = true
  try {
    const rsp = await request.get('/settings/getsetting', {
      params: { section: state.section }
    })
    const data = rsp?.data?.data || {}
    state.enabled = data.enabled || 'false'
    state.interval_seconds = data.interval_seconds || '300'
    state.log_level = (data.log_level || 'info').toLowerCase()
  } catch (error) {
    toast.error('获取消息队列状态策略失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  if (!/^\d+$/.test(state.interval_seconds)) {
    toast.warning('自动更新频率必须是整数秒')
    return
  }
  const seconds = Number(state.interval_seconds)
  if (seconds < 10 || seconds > 86400) {
    toast.warning('自动更新频率范围为 10 ~ 86400 秒')
    return
  }

  saving.value = true
  try {
    const rsp = await request.post('/settings/set', {
      section: state.section,
      data: {
        enabled: state.enabled,
        interval_seconds: String(seconds),
        log_level: state.log_level
      }
    })
    if (rsp?.data?.code === 200) {
      toast.success('策略保存成功')
      return
    }
    toast.error(rsp?.data?.msg || '策略保存失败')
  } catch (error: any) {
    toast.error(error?.response?.data?.msg || '策略保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<template>
  <div class="space-y-5">
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-4 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">消息队列状态更新策略</div>

      <div class="flex items-center justify-between">
        <div class="space-y-1">
          <div class="text-sm font-medium">自动更新</div>
          <div class="text-xs text-muted-foreground">
            关闭表示手动更新（仅点击测试时更新状态）；打开表示按频率自动更新
          </div>
        </div>
        <Switch v-model="enabledBool" :disabled="loading || saving" />
      </div>

      <div v-if="enabledBool" class="space-y-2">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-200">自动更新频率（秒）</label>
        <Input
          v-model="state.interval_seconds"
          type="number"
          min="10"
          max="86400"
          step="1"
          placeholder="例如：300"
          :disabled="loading || saving"
        />
        <p class="text-xs text-muted-foreground">建议范围：60~600 秒</p>
      </div>

      <div class="space-y-2">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-200">日志级别</label>
        <Select v-model="state.log_level" :disabled="loading || saving">
          <SelectTrigger>
            <SelectValue placeholder="选择日志级别" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="debug">debug</SelectItem>
            <SelectItem value="info">info</SelectItem>
            <SelectItem value="warn">warn</SelectItem>
            <SelectItem value="error">error</SelectItem>
          </SelectContent>
        </Select>
        <p class="text-xs text-muted-foreground">
          推荐生产环境使用 warn 或 error，可明显减少终端日志噪音
        </p>
        <p v-if="state.log_level === 'debug'" class="text-xs text-amber-600 dark:text-amber-400">
          已选择 debug，日志量会明显增加（建议仅在排查问题时临时开启）
        </p>
      </div>

      <div class="flex justify-end">
        <Button :disabled="loading || saving" @click="saveConfig">
          {{ saving ? '保存中...' : '保存策略' }}
        </Button>
      </div>
    </div>
  </div>
</template>
