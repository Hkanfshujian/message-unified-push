<script setup lang="ts">
import type { UploadFile, UploadFiles, UploadRawFile } from 'element-plus'

withDefaults(defineProps<{
  accept?: string
  buttonText?: string
  tip?: string
  disabled?: boolean
  emptyText?: string
}>(), {
  accept: '',
  buttonText: '选择文件',
  tip: '',
  disabled: false,
  emptyText: '拖拽文件到此处，或点击按钮选择文件'
})

const emit = defineEmits<{
  select: [file: UploadRawFile]
  change: [file: UploadFile, files: UploadFiles]
}>()

const beforeUpload = (file: UploadRawFile) => {
  emit('select', file)
  return false
}

const handleChange = (file: UploadFile, files: UploadFiles) => {
  emit('change', file, files)
}
</script>

<template>
  <el-upload
    class="app-upload dora-material-inset"
    drag
    :accept="accept"
    :disabled="disabled"
    :auto-upload="false"
    :before-upload="beforeUpload"
    @change="handleChange"
  >
    <div class="app-upload-empty" :class="{ 'dora-state-disabled': disabled }">
      <div class="app-upload-icon">↑</div>
      <div class="app-upload-title">{{ emptyText }}</div>
      <el-button type="primary" :disabled="disabled">{{ buttonText }}</el-button>
    </div>
    <template v-if="tip" #tip>
      <div class="app-upload-tip">{{ tip }}</div>
    </template>
  </el-upload>
</template>
