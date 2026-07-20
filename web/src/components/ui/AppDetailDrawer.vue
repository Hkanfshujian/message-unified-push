<script setup lang="ts">
const visible = defineModel<boolean>({ required: true })

withDefaults(defineProps<{
  title: string
  size?: string | number
  direction?: 'rtl' | 'ltr' | 'ttb' | 'btt'
  bodyMode?: 'scroll' | 'managed'
  density?: 'default' | 'compact'
}>(), {
  size: '520px',
  direction: 'rtl',
  bodyMode: 'scroll',
  density: 'default'
})
</script>

<template>
  <el-drawer
    v-model="visible"
    :size="size"
    :direction="direction"
    class="app-detail-drawer dora-material-overlay"
    :class="[`app-drawer-body-${bodyMode}`, `app-drawer-density-${density}`]"
    destroy-on-close
    append-to-body
  >
    <template #header>
      <div class="app-detail-drawer-header">
        <div class="min-w-0">
          <h2 class="app-detail-drawer-title">{{ title }}</h2>
        </div>
      </div>
    </template>
    <div class="app-detail-drawer-body dora-material-inset">
      <slot />
    </div>
    <template v-if="$slots.footer" #footer>
      <div class="app-detail-drawer-footer dora-material-panel">
        <slot name="footer" />
      </div>
    </template>
  </el-drawer>
</template>
