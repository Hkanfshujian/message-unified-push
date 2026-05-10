<template>
  <Card
    class="w-full cursor-pointer border-border/60 shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all duration-[var(--motion-fast)]"
    @click="handleClick"
  >
    <CardContent class="pt-6">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm text-muted-foreground">{{ title }}</p>
          <div class="mt-1 text-3xl font-semibold leading-none tracking-tight">{{ value }}</div>
          <p class="mt-2 text-xs text-muted-foreground">
            {{ description || '较昨日' }}
            <span :class="trendClass">{{ trendText || '0%' }}</span>
          </p>
        </div>
        <div :class="iconWrapClass">
          <component :is="icon" class="h-5 w-5" />
        </div>
      </div>
      <div class="mt-3 h-8">
        <svg viewBox="0 0 120 32" class="h-full w-full">
          <defs>
            <linearGradient :id="gradientId" x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" :stop-color="sparkColor" stop-opacity="0.35" />
              <stop offset="100%" :stop-color="sparkColor" stop-opacity="0" />
            </linearGradient>
          </defs>
          <path d="M2 26 C 16 24, 20 20, 32 18 C 46 16, 58 24, 70 20 C 84 16, 92 9, 118 8" fill="none" :stroke="sparkColor" stroke-width="2.2" stroke-linecap="round" />
          <path d="M2 26 C 16 24, 20 20, 32 18 C 46 16, 58 24, 70 20 C 84 16, 92 9, 118 8 L118 32 L2 32 Z" :fill="`url(#${gradientId})`" />
        </svg>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent } from "@/components/ui/card"
import { computed, type Component } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const props = defineProps<{
  title: string
  value: string | number
  description?: string
  icon?: Component
  routePath?: string
  trendText?: string
  trendType?: 'up' | 'down' | 'flat'
  tone?: 'blue' | 'teal' | 'purple' | 'red'
}>()

const toneMap = {
  blue: {
    wrap: 'h-10 w-10 rounded-full bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300',
    spark: '#3b82f6'
  },
  teal: {
    wrap: 'h-10 w-10 rounded-full bg-teal-50 text-teal-600 dark:bg-teal-500/15 dark:text-teal-300',
    spark: '#14b8a6'
  },
  purple: {
    wrap: 'h-10 w-10 rounded-full bg-purple-50 text-purple-600 dark:bg-purple-500/15 dark:text-purple-300',
    spark: '#8b5cf6'
  },
  red: {
    wrap: 'h-10 w-10 rounded-full bg-red-50 text-red-600 dark:bg-red-500/15 dark:text-red-300',
    spark: '#ef4444'
  }
} as const

const iconWrapClass = computed(() => toneMap[props.tone || 'blue'].wrap)
const sparkColor = computed(() => toneMap[props.tone || 'blue'].spark)
const gradientId = computed(() => `spark-gradient-${props.tone || 'blue'}`)
const trendClass = computed(() => {
  if (props.trendType === 'up') return 'text-emerald-500 ml-1'
  if (props.trendType === 'down') return 'text-red-500 ml-1'
  return 'text-muted-foreground ml-1'
})

const handleClick = () => {
  if (props.routePath) {
    router.push(props.routePath)
  }
}
</script>

<script lang="ts">
export default {
  name: 'CardNum'
}
</script>
