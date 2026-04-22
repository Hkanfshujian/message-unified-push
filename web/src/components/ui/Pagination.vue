<template>
  <div class="flex items-center justify-end gap-2 w-full">
    <!-- 统计信息 + 每页条数选择 -->
    <div class="flex items-center gap-2 text-[12px] text-muted-foreground shrink-0">
      <span>共 {{ total }} 条</span>
      <div class="hidden sm:flex items-center gap-1">
        <span>每页</span>
        <select
          class="rounded-[1px] border border-input bg-transparent px-3 py-1.5 text-[12px] leading-none focus:outline-none focus:ring-1 focus:ring-primary"
          :value="pageSize"
          @change="handlePageSizeNativeChange"
        >
          <option v-for="size in pageSizeDisplayOptions" :key="size" :value="size">
            {{ size }}
          </option>
        </select>
        <span>条</span>
      </div>
    </div>
    
    <!-- 分页控件 -->
    <div
      class="flex justify-end shrink-0"
      tabindex="0"
      @keydown="handleKeydown"
    >
      <Pagination
        :total="total"
        :items-per-page="pageSize"
        :sibling-count="1"
        :show-edges="true"
        :default-page="currentPage"
        :page="currentPage"
        @update:page="handlePageChange"
      >
      <PaginationContent class="flex gap-2">
        <PaginationItem :value="currentPage - 1">
          <PaginationPrevious 
            :disabled="currentPage <= 1" 
            class="px-3 py-1.5 rounded-[1px] text-[12px] leading-none border-none bg-transparent text-foreground/80 hover:text-foreground disabled:text-muted-foreground disabled:hover:text-muted-foreground disabled:cursor-default disabled:pointer-events-none" 
            @click.prevent.stop="handlePageChange(currentPage - 1)"
          >
            上一页
          </PaginationPrevious>
        </PaginationItem>

        <template v-for="(page, index) in displayItems" :key="index">
          <PaginationItem v-if="page.type === 'page'" :value="page.value" :is-active="page.value === currentPage">
            <button
              :class="[
                'px-3 py-1.5 rounded-[1px] text-[12px] leading-none select-none transition-transform duration-150',
                page.value === currentPage
                  ? 'font-semibold text-foreground cursor-default'
                  : 'text-gray-400 hover:text-gray-500 hover:-translate-y-0.5'
              ]"
              :disabled="page.value === currentPage"
              @click="page.value !== currentPage && handlePageChange(page.value)"
            >
              {{ page.value }}
            </button>
          </PaginationItem>
          <PaginationEllipsis v-else-if="page.type === 'ellipsis'" :index="index" class="h-8 w-5 sm:h-9 sm:w-9" />
        </template>

        <PaginationItem :value="currentPage + 1">
          <PaginationNext 
            :disabled="currentPage >= totalPages" 
            class="px-3 py-1.5 rounded-[1px] text-[12px] leading-none border-none bg-transparent text-foreground/80 hover:text-foreground disabled:text-muted-foreground disabled:hover:text-muted-foreground disabled:cursor-default disabled:pointer-events-none" 
            @click.prevent.stop="handlePageChange(currentPage + 1)"
          >
            下一页
          </PaginationNext>
        </PaginationItem>
      </PaginationContent>
      </Pagination>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'

interface Props {
  total: number
  currentPage: number
  pageSize: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'page-change': [page: number]
  'page-size-change': [size: number]
}>()

const totalPages = computed(() => {
  return Math.ceil(props.total / props.pageSize)
})

// 分页项目类型
type PaginationItem = 
  | { type: 'page'; value: number }
  | { type: 'ellipsis' }

const pageSizeOptions = [10, 20, 50, 100]

const pageSizeDisplayOptions = computed(() => {
  const base = [...pageSizeOptions]
  if (!base.includes(props.pageSize)) {
    base.push(props.pageSize)
  }
  return base.sort((a, b) => a - b)
})

// 响应式显示项目（根据屏幕大小使用不同逻辑）
const displayItems = computed((): PaginationItem[] => {
  const pages: PaginationItem[] = []
  const total = totalPages.value
  const current = props.currentPage
  
  // 使用媒体查询判断是否为小屏
  const isSmallScreen = typeof window !== 'undefined' && window.innerWidth < 640
  
  if (isSmallScreen) {
    // 小屏逻辑：仅保留当前页为可点击页码，前后使用省略号占位
    if (total <= 1) {
      pages.push({ type: 'page', value: 1 })
    } else {
      if (current > 2) {
        pages.push({ type: 'ellipsis' })
      }
      pages.push({ type: 'page', value: current })
      if (current < total - 1) {
        pages.push({ type: 'ellipsis' })
      }
    }
  } else {
    // 大屏逻辑：显示更多页码
    if (total <= 7) {
      // 如果总页数小于等于7，显示所有页码
      for (let i = 1; i <= total; i++) {
        pages.push({ type: 'page', value: i })
      }
    } else {
      // 复杂分页逻辑
      if (current <= 2) {
        // 当前页在前面
        for (let i = 1; i <= 3; i++) {
          pages.push({ type: 'page', value: i })
        }
        pages.push({ type: 'ellipsis' })
        pages.push({ type: 'page', value: total })
      } else if (current >= total - 3) {
        // 当前页在后面
        pages.push({ type: 'page', value: 1 })
        pages.push({ type: 'ellipsis' })
        for (let i = total - 4; i <= total; i++) {
          pages.push({ type: 'page', value: i })
        }
      } else {
        // 当前页在中间
        pages.push({ type: 'page', value: 1 })
        pages.push({ type: 'ellipsis' })
        for (let i = current - 1; i <= current + 1; i++) {
          pages.push({ type: 'page', value: i })
        }
        pages.push({ type: 'ellipsis' })
        pages.push({ type: 'page', value: total })
      }
    }
  }
  
  return pages
})

const handlePageChange = (page: number) => {
  if (page !== props.currentPage && page >= 1 && page <= totalPages.value) {
    emit('page-change', page)
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    handlePageChange(props.currentPage - 1)
  } else if (event.key === 'ArrowRight') {
    event.preventDefault()
    handlePageChange(props.currentPage + 1)
  }
}

const handlePageSizeNativeChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  const size = Number(target.value) || props.pageSize
  if (size !== props.pageSize) {
    emit('page-size-change', size)
  }
}
</script>

<script lang="ts">
export default {
  name: 'Pagination'
}
</script>
