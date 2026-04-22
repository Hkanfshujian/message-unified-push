<script setup lang="ts">
import { reactive, onMounted, computed } from 'vue'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'

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
  techStack: ['Golang','Vue 3', 'TypeScript', 'Vite', 'Tailwind CSS', 'Shadcn/ui'],
  memoryUsage: '',
  uptime: ''
})

// 获取关于页面配置
const getAboutConfig = async () => {
  try {
    const params = { params: { section: 'about' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      if (data.version) state.version = data.version
      if (data.memory_usage) state.memoryUsage = data.memory_usage
      if (data.uptime) state.uptime = data.uptime
    }
  } catch (error) {
    toast.error('获取关于信息失败')
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
      <!-- 技术栈 + 功能特性 -->
      <div class="space-y-6">
        <div>
          <h3 class="text-sm font-medium text-gray-900 mb-2">技术栈</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="tech in state.techStack"
              :key="tech"
              class="inline-flex items-center rounded-full bg-[#fef3c7] text-[#d97706] px-3 py-1 text-xs font-medium"
            >
              {{ tech }}
            </span>
          </div>
        </div>

        <div>
          <h3 class="text-sm font-medium text-gray-900 mb-2">功能特性</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="feature in state.features"
              :key="feature"
              class="inline-flex items-center rounded-full bg-[#f3f4f6] text-[#4b5563] px-3 py-1 text-xs"
            >
              {{ feature }}
            </span>
          </div>
        </div>
      </div>

      <!-- 系统信息 -->
      <div class="space-y-6">
        <div>
          <h3 class="text-sm font-medium text-gray-900 mb-2">系统信息</h3>
          <dl class="space-y-2 text-sm">
            <div class="flex justify-between">
              <dt class="text-gray-500">系统版本</dt>
              <dd class="text-gray-900 font-medium">
                {{ state.version }}
              </dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-gray-500">构建时间</dt>
              <dd class="text-gray-900">
                {{ buildTime.includes('开发模式') ? buildTime : new Date(buildTime).toLocaleString('zh-CN') }}
              </dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-gray-500">内存使用</dt>
              <dd class="text-gray-900">
                {{ state.memoryUsage || '获取中...' }}
              </dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-gray-500">运行时间</dt>
              <dd class="text-gray-900">
                {{ state.uptime || '获取中...' }}
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  </div>
</template>
