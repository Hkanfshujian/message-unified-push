<script setup lang="ts">
import { useRouter } from 'vue-router'
import { HomeOutlined, ArrowLeftOutlined } from '@ant-design/icons-vue'
import { zhCN } from '@/locales/zh-CN'

const router = useRouter()
const messages = zhCN.notFound

// 返回首页
const goHome = () => {
  router.push('/')
}

// 返回上一页
const goBack = () => {
  router.go(-1)
}
</script>

<template>
  <div class="min-h-screen bg-background flex items-center justify-center px-4">
    <el-card class="w-full max-w-md text-center">
      <div class="pb-4">
        <div class="mx-auto mb-4 w-24 h-24 bg-red-100 dark:bg-red-400/20 border border-border rounded-full flex items-center justify-center">
          <span class="text-4xl font-bold text-red-600 dark:text-red-400">{{ messages.code }}</span>
        </div>
        <h1 class="text-2xl font-bold text-foreground">{{ messages.title }}</h1>
        <p class="text-muted-foreground">
          {{ messages.description }}
        </p>
      </div>
      <div class="space-y-4">
        <p class="text-sm text-muted-foreground">
          {{ messages.guidance }}
        </p>
        <div class="flex flex-col sm:flex-row gap-3 justify-center">
          <el-button type="primary" @click="goHome" class="flex items-center gap-2">
            <HomeOutlined class="text-[16px]" />
            {{ messages.home }}
          </el-button>
          <el-button @click="goBack" class="flex items-center gap-2">
            <ArrowLeftOutlined class="text-[16px]" />
            {{ messages.back }}
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
/* 可以添加一些动画效果 */
.card-enter-active,
.card-leave-active {
  transition: all 0.3s ease;
}

.card-enter-from,
.card-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
