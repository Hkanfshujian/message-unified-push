<script setup lang="ts">
import { computed, ref } from 'vue'

const props = withDefaults(defineProps<{
  text?: string | number | null
  title?: string
  width?: string
  preview?: boolean
}>(), {
  text: '',
  title: '内容详情',
  width: '640px',
  preview: false
})

const visible = ref(false)
const displayText = computed(() => String(props.text ?? '') || '-')

const open = () => {
  if (props.preview) visible.value = true
}
</script>

<template>
  <button
    v-if="preview"
    type="button"
    class="inline-block max-w-full cursor-pointer truncate border-0 bg-transparent p-0 align-middle text-left text-foreground underline decoration-dotted underline-offset-4"
    :title="displayText"
    :aria-label="`预览${title}：${displayText}`"
    @click="open"
  >
    {{ displayText }}
  </button>
  <span v-else class="inline-block max-w-full truncate align-middle text-foreground" :title="displayText">
    {{ displayText }}
  </span>
  <el-dialog v-model="visible" :title="title" :width="width">
    <pre class="max-h-[65vh] overflow-auto whitespace-pre-wrap break-words rounded-lg bg-muted/40 p-4 text-sm leading-relaxed">{{ displayText }}</pre>
  </el-dialog>
</template>
