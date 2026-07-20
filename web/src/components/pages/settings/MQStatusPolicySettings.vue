<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { settingsApi } from '@/api/settings'
import { notifyError, notifySuccess, notifyWarning } from '@/util/uiFeedback'

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
    const rsp = await settingsApi.get(state.section)
    const data = rsp?.data?.data || {}
    state.enabled = data.enabled || 'false'
    state.interval_seconds = data.interval_seconds || '300'
    state.log_level = (data.log_level || 'info').toLowerCase()
  } catch (error) {
    notifyError('获取消息队列状态策略失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  if (!/^\d+$/.test(state.interval_seconds)) {
    notifyWarning('自动更新频率必须是整数秒')
    return
  }
  const seconds = Number(state.interval_seconds)
  if (seconds < 10 || seconds > 86400) {
    notifyWarning('自动更新频率范围为 10 ~ 86400 秒')
    return
  }

  saving.value = true
  try {
    const rsp = await settingsApi.set(state.section, {
      enabled: state.enabled,
      interval_seconds: String(seconds),
      log_level: state.log_level
    })
    if (rsp?.data?.code === 200) {
      notifySuccess('策略保存成功')
      return
    }
    notifyError(rsp?.data?.msg || '策略保存失败')
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || '策略保存失败')
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
    <el-card v-loading="loading" shadow="never" class="settings-section-card">
      <div class="text-sm font-semibold">消息队列状态更新策略</div>

      <div class="flex items-center justify-between mt-4">
        <div class="space-y-1">
          <div class="text-sm font-medium">自动更新</div>
          <div class="text-xs text-muted-foreground">
            关闭表示手动更新（仅点击测试时更新状态）；打开表示按频率自动更新
          </div>
        </div>
        <el-switch v-model="enabledBool" :disabled="loading || saving" />
      </div>

      <div v-if="enabledBool" class="space-y-2">
        <label class="text-sm font-medium text-foreground">自动更新频率（秒）</label>
        <el-input
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
        <label class="text-sm font-medium text-foreground">日志级别</label>
        <el-select v-model="state.log_level" class="w-full" :disabled="loading || saving" placeholder="选择日志级别">
          <el-option label="debug" value="debug" />
          <el-option label="info" value="info" />
          <el-option label="warn" value="warn" />
          <el-option label="error" value="error" />
        </el-select>
        <p class="text-xs text-muted-foreground">
          推荐生产环境使用 warn 或 error，可明显减少终端日志噪音
        </p>
        <p v-if="state.log_level === 'debug'" class="text-xs text-amber-600 dark:text-amber-400">
          已选择 debug，日志量会明显增加（建议仅在排查问题时临时开启）
        </p>
      </div>

      <div class="flex justify-end">
        <el-button type="primary" :disabled="loading || saving" :loading="saving" @click="saveConfig">
          保存策略
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
:deep(.settings-section-card > .el-card__body) {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
