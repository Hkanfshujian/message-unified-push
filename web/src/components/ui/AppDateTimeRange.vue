<script setup lang="ts">
const modelValue = defineModel<[string, string] | []>({ required: true })

withDefaults(defineProps<{
  type?: 'daterange' | 'datetimerange'
  startPlaceholder?: string
  endPlaceholder?: string
  valueFormat?: string
  clearable?: boolean
}>(), {
  type: 'datetimerange',
  startPlaceholder: '开始时间',
  endPlaceholder: '结束时间',
  valueFormat: 'YYYY-MM-DDTHH:mm',
  clearable: true
})
</script>

<template>
  <el-date-picker
    v-model="modelValue"
    :type="type"
    :start-placeholder="startPlaceholder"
    :end-placeholder="endPlaceholder"
    :value-format="valueFormat"
    :clearable="clearable"
    range-separator="至"
    class="app-date-time-range dora-material-inset w-full"
  />
</template>
