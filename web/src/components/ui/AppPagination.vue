<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const currentPage = defineModel<number>('currentPage', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const props = withDefaults(defineProps<{
  total: number
  pageSizes?: number[]
  layout?: string
  background?: boolean
  compact?: boolean
}>(), {
  pageSizes: () => [10, 20, 50, 100],
  layout: 'total, sizes, prev, pager, next, jumper',
  background: true,
  compact: false
})

const emit = defineEmits<{
  change: [payload: { page: number; pageSize: number }]
}>()

const jumpPage = ref(String(currentPage.value || 1))

const totalPages = computed(() => Math.max(Math.ceil(props.total / pageSize.value), 1))
const pageSizeOptions = computed(() => {
  const options = [...props.pageSizes]
  if (!options.includes(pageSize.value)) options.push(pageSize.value)
  return options.sort((a, b) => a - b)
})

const canPrev = computed(() => currentPage.value > 1)
const canNext = computed(() => currentPage.value < totalPages.value)

const displayPages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value
  const current = Math.min(Math.max(currentPage.value, 1), total)
  const start = Math.max(1, Math.min(current - 1, total - 2))
  const end = Math.min(total, start + 2)
  for (let page = start; page <= end; page += 1) pages.push(page)
  return pages
})

const emitChange = () => {
  emit('change', { page: currentPage.value, pageSize: pageSize.value })
}

const handleCurrentChange = (value: number) => {
  currentPage.value = Math.min(Math.max(value, 1), totalPages.value)
  jumpPage.value = String(currentPage.value)
  emitChange()
}

const handleSizeChange = (value: number) => {
  pageSize.value = value
  currentPage.value = 1
  jumpPage.value = '1'
  emitChange()
}

const handlePageSizeChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  handleSizeChange(Number(target.value) || pageSize.value)
}

const handleJump = () => {
  const page = Number(jumpPage.value)
  if (!Number.isFinite(page)) {
    jumpPage.value = String(currentPage.value)
    return
  }
  handleCurrentChange(page)
}

watch(currentPage, value => {
  jumpPage.value = String(value || 1)
})
</script>

<template>
  <div class="app-pagination-wrap flex justify-end py-4" :class="{ 'is-compact': compact }">
    <div class="app-pagination-bar">
      <div class="app-pagination-summary">共 {{ total }} 条</div>
      <select v-if="!compact" class="app-pagination-size" :value="pageSize" @change="handlePageSizeChange">
        <option v-for="size in pageSizeOptions" :key="size" :value="size">{{ size }}条/页</option>
      </select>
      <div class="app-pagination-pages" role="navigation" aria-label="分页">
        <button class="app-pagination-button" type="button" :disabled="!canPrev" aria-label="上一页" @click="handleCurrentChange(currentPage - 1)">
          ‹
        </button>
        <template v-if="compact">
          <span class="app-pagination-current" aria-live="polite">{{ currentPage }} / {{ totalPages }}</span>
        </template>
        <template v-else>
          <button
            v-for="page in displayPages"
            :key="page"
            class="app-pagination-button app-pagination-number"
            :class="page === currentPage ? 'is-active' : ''"
            type="button"
            :aria-current="page === currentPage ? 'page' : undefined"
            @click="handleCurrentChange(page)"
          >
            {{ page }}
          </button>
        </template>
        <button class="app-pagination-button" type="button" :disabled="!canNext" aria-label="下一页" @click="handleCurrentChange(currentPage + 1)">
          ›
        </button>
      </div>
      <div v-if="!compact" class="app-pagination-jumper">
        <span>前往</span>
        <input v-model="jumpPage" class="app-pagination-input" type="number" min="1" :max="totalPages" @keyup.enter="handleJump" @blur="handleJump">
        <span>页</span>
      </div>
    </div>
  </div>
</template>
