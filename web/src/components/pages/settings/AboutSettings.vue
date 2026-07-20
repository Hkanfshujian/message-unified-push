<script setup lang="ts">
import { reactive, onMounted, computed } from 'vue'
import { settingsApi } from '@/api/settings'
import { notifyError } from '@/util/uiFeedback'

const state = reactive({
  version: '1.0.0',
  description: '一个现代化的消息推送管理平台，支持多种推送渠道和灵活的消息管理功能。',
  features: [
    '多渠道消息推送',
    '定时消息管理',
    '发信日志追踪',
    '渠道配置管理',
    '站点信息配置',
  ],
  techStack: ['Golang', 'Vue 3', 'TypeScript', 'Vite', 'Element Plus', 'UnoCSS'],
  memoryUsage: '',
  uptime: ''
})

// 获取关于页面配置
const getAboutConfig = async () => {
  try {
    const params = { params: { section: 'about' } }
    const response = await settingsApi.get(params.params.section)
    if (response.data.code === 200) {
      const data = response.data.data
      if (data.version) state.version = data.version
      if (data.memory_usage) state.memoryUsage = data.memory_usage
      if (data.uptime) state.uptime = data.uptime
    }
  } catch (error) {
    notifyError('获取关于信息失败')
  }
}

// 获取构建时间
const buildTime = computed(() => {
  try {
    return (globalThis as any).__BUILD_TIME__ || '开发模式 - 未构建'
  } catch {
    return '开发模式 - 未构建'
  }
})

onMounted(() => {
  getAboutConfig()
})
</script>

<script lang="ts">
export default {
  name: 'AboutSettings'
}
</script>

<template>
  <div class="space-y-8">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <el-card shadow="never">
        <div class="space-y-6">
          <div>
            <h3 class="text-sm font-medium text-foreground mb-2">技术栈</h3>
          <div class="flex flex-wrap gap-2">
            <el-tag
              v-for="tech in state.techStack"
              :key="tech"
              type="warning"
              effect="light"
            >
              {{ tech }}
            </el-tag>
          </div>
        </div>

        <div>
          <h3 class="text-sm font-medium text-foreground mb-2">功能特性</h3>
          <div class="flex flex-wrap gap-2">
            <el-tag
              v-for="feature in state.features"
              :key="feature"
              type="info"
              effect="plain"
            >
              {{ feature }}
            </el-tag>
          </div>
        </div>
      </div>
      </el-card>

      <el-card shadow="never">
        <h3 class="text-sm font-medium text-foreground mb-4">系统信息</h3>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="系统版本">{{ state.version }}</el-descriptions-item>
          <el-descriptions-item label="构建时间">
            {{ buildTime.includes('开发模式') ? buildTime : new Date(buildTime).toLocaleString('zh-CN') }}
          </el-descriptions-item>
          <el-descriptions-item label="内存使用">{{ state.memoryUsage || '获取中...' }}</el-descriptions-item>
          <el-descriptions-item label="运行时间">{{ state.uptime || '获取中...' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
    </div>
  </div>
</template>
