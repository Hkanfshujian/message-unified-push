<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ChevronDown } from 'lucide-vue-next'

interface Props {
  modelValue: string
  placeholder?: string
}

interface CalendarCell {
  year: number
  month: number
  day: number
  inCurrentMonth: boolean
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '选择日期时间'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change'): void
}>()

const rootRef = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const panelYear = ref(0)
const panelMonth = ref(0)

const draft = reactive({
  year: 0,
  month: 0,
  day: 0,
  hour: 0,
  minute: 0
})

const weekDays = ['一', '二', '三', '四', '五', '六', '日']
const hours = Array.from({ length: 24 }, (_, index) => index)
const minutes = Array.from({ length: 60 }, (_, index) => index)

const pad = (value: number) => String(value).padStart(2, '0')

const isValidDate = (date: Date) => !Number.isNaN(date.getTime())

const parseDateTime = (value: string): Date | null => {
  const raw = value?.trim()
  if (!raw) return null

  const normalized = raw.replace(' ', 'T')
  const matched = normalized.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?$/)

  if (matched) {
    const [, year, month, day, hour, minute, second = '0'] = matched
    const parsed = new Date(
      Number(year),
      Number(month) - 1,
      Number(day),
      Number(hour),
      Number(minute),
      Number(second)
    )
    if (isValidDate(parsed)) return parsed
  }

  const fallback = new Date(normalized)
  return isValidDate(fallback) ? fallback : null
}

const applyDateToDraft = (date: Date) => {
  draft.year = date.getFullYear()
  draft.month = date.getMonth()
  draft.day = date.getDate()
  draft.hour = date.getHours()
  draft.minute = date.getMinutes()
}

const initializeFromModel = (modelValue: string) => {
  const parsed = parseDateTime(modelValue) ?? new Date()
  applyDateToDraft(parsed)
  panelYear.value = parsed.getFullYear()
  panelMonth.value = parsed.getMonth()
}

watch(
  () => props.modelValue,
  (modelValue) => {
    if (isOpen.value || !modelValue) return
    const parsed = parseDateTime(modelValue)
    if (!parsed) return
    applyDateToDraft(parsed)
    panelYear.value = parsed.getFullYear()
    panelMonth.value = parsed.getMonth()
  },
  { immediate: true }
)

const monthTitle = computed(() => `${panelYear.value}-${pad(panelMonth.value + 1)}`)

const displayValue = computed(() => {
  const parsed = parseDateTime(props.modelValue)
  if (!parsed) return ''
  return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())} ${pad(parsed.getHours())}:${pad(parsed.getMinutes())}`
})

const calendarCells = computed<CalendarCell[]>(() => {
  const firstDay = new Date(panelYear.value, panelMonth.value, 1).getDay()
  const startOffset = (firstDay + 6) % 7
  const daysInCurrentMonth = new Date(panelYear.value, panelMonth.value + 1, 0).getDate()
  const daysInPrevMonth = new Date(panelYear.value, panelMonth.value, 0).getDate()

  const cells: CalendarCell[] = []

  for (let i = startOffset; i > 0; i--) {
    cells.push({
      year: new Date(panelYear.value, panelMonth.value - 1, 1).getFullYear(),
      month: new Date(panelYear.value, panelMonth.value - 1, 1).getMonth(),
      day: daysInPrevMonth - i + 1,
      inCurrentMonth: false
    })
  }

  for (let day = 1; day <= daysInCurrentMonth; day++) {
    cells.push({
      year: panelYear.value,
      month: panelMonth.value,
      day,
      inCurrentMonth: true
    })
  }

  let trailingDay = 1
  while (cells.length < 42) {
    const nextMonth = new Date(panelYear.value, panelMonth.value + 1, 1)
    cells.push({
      year: nextMonth.getFullYear(),
      month: nextMonth.getMonth(),
      day: trailingDay++,
      inCurrentMonth: false
    })
  }

  return cells
})

const isSelectedDay = (cell: CalendarCell) => {
  return draft.year === cell.year && draft.month === cell.month && draft.day === cell.day
}

const previousMonth = () => {
  const previous = new Date(panelYear.value, panelMonth.value - 1, 1)
  panelYear.value = previous.getFullYear()
  panelMonth.value = previous.getMonth()
}

const nextMonth = () => {
  const next = new Date(panelYear.value, panelMonth.value + 1, 1)
  panelYear.value = next.getFullYear()
  panelMonth.value = next.getMonth()
}

const selectDay = (cell: CalendarCell) => {
  draft.year = cell.year
  draft.month = cell.month
  draft.day = cell.day
  panelYear.value = cell.year
  panelMonth.value = cell.month
}

const selectNow = () => {
  const now = new Date()
  applyDateToDraft(now)
  panelYear.value = now.getFullYear()
  panelMonth.value = now.getMonth()
}

const buildModelValue = () =>
  `${draft.year}-${pad(draft.month + 1)}-${pad(draft.day)}T${pad(draft.hour)}:${pad(draft.minute)}`

const confirmSelection = () => {
  emit('update:modelValue', buildModelValue())
  emit('change')
  isOpen.value = false
}

const closePanel = () => {
  isOpen.value = false
}

const openPanel = () => {
  initializeFromModel(props.modelValue)
  isOpen.value = true
}

const togglePanel = () => {
  if (isOpen.value) {
    closePanel()
    return
  }
  openPanel()
}

const handleDocumentMousedown = (event: MouseEvent) => {
  if (!isOpen.value) return
  const target = event.target as Node | null
  if (target && rootRef.value?.contains(target)) return
  closePanel()
}

const handleDocumentKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    closePanel()
  }
}

onMounted(() => {
  document.addEventListener('mousedown', handleDocumentMousedown)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleDocumentMousedown)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
</script>

<template>
  <div ref="rootRef" class="datetime-picker-root">
    <button type="button" class="datetime-picker-trigger" @click="togglePanel">
      <span :class="['datetime-picker-text', { placeholder: !displayValue }]">
        {{ displayValue || placeholder }}
      </span>
      <span class="datetime-picker-arrow" :class="{ open: isOpen }">
        <ChevronDown class="datetime-picker-arrow-icon" />
      </span>
    </button>

    <div v-if="isOpen" class="datetime-panel">
      <div class="panel-header">
        <button type="button" class="header-btn" @click="previousMonth">‹</button>
        <span class="month-title">{{ monthTitle }}</span>
        <button type="button" class="header-btn" @click="nextMonth">›</button>
      </div>

      <div class="panel-body">
        <div class="calendar-block">
          <div class="week-row">
            <span v-for="day in weekDays" :key="day" class="week-cell">{{ day }}</span>
          </div>
          <div class="date-grid">
            <button
              v-for="cell in calendarCells"
              :key="`${cell.year}-${cell.month}-${cell.day}`"
              type="button"
              :class="[
                'date-cell',
                { 'out-month': !cell.inCurrentMonth },
                { selected: isSelectedDay(cell) }
              ]"
              @click="selectDay(cell)"
            >
              {{ cell.day }}
            </button>
          </div>
        </div>

        <div class="time-block">
          <div class="time-column">
            <div class="time-title">时</div>
            <div class="time-list">
              <button
                v-for="hour in hours"
                :key="`hour-${hour}`"
                type="button"
                :class="['time-item', { selected: draft.hour === hour }]"
                @click="draft.hour = hour"
              >
                {{ pad(hour) }}
              </button>
            </div>
          </div>

          <div class="time-column">
            <div class="time-title">分</div>
            <div class="time-list">
              <button
                v-for="minute in minutes"
                :key="`minute-${minute}`"
                type="button"
                :class="['time-item', { selected: draft.minute === minute }]"
                @click="draft.minute = minute"
              >
                {{ pad(minute) }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="panel-footer">
        <button type="button" class="now-btn" @click="selectNow">此刻</button>
        <div class="footer-actions">
          <button type="button" class="cancel-btn" @click="closePanel">取消</button>
          <button type="button" class="confirm-btn" @click="confirmSelection">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.datetime-picker-root {
  position: relative;
  width: 200px;
}

.datetime-picker-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 36px;
  padding: 8px 12px;
  font-size: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background-color: #fff;
  color: #1e293b;
  outline: none;
  cursor: pointer;
  transition: all 0.2s;
}

.datetime-picker-trigger:hover {
  border-color: #94a3b8;
}

.datetime-picker-trigger:focus-visible {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgba(22, 119, 255, 0.18);
}

.datetime-picker-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.datetime-picker-text.placeholder {
  color: #94a3b8;
}

.datetime-picker-arrow {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  margin-left: 8px;
  color: #64748b;
  flex-shrink: 0;
  transition: transform 0.18s ease, color 0.18s ease;
}

.datetime-picker-arrow-icon {
  width: 14px;
  height: 14px;
  stroke-width: 2.25;
}

.datetime-picker-arrow.open {
  transform: rotate(180deg);
  color: #334155;
}

.datetime-panel {
  position: absolute;
  z-index: 60;
  top: calc(100% + 6px);
  left: 0;
  width: min(392px, calc(100vw - 20px));
  border: 1px solid #e2e8f0;
  border-radius: 7px;
  background: #fff;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.1);
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 7px 8px;
  border-bottom: 1px solid #f1f5f9;
}

.month-title {
  font-size: 12px;
  font-weight: 600;
  color: #0f172a;
}

.header-btn {
  width: 20px;
  height: 20px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  background: #fff;
  color: #334155;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;
}

.header-btn:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.panel-body {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px;
  padding: 8px;
}

.calendar-block {
  min-width: 194px;
}

.week-row {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  margin-bottom: 3px;
}

.week-cell {
  text-align: center;
  font-size: 10px;
  color: #64748b;
  line-height: 20px;
}

.date-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 3px;
}

.date-cell {
  width: 100%;
  height: 24px;
  border: 1px solid transparent;
  border-radius: 4px;
  background: #fff;
  color: #1e293b;
  font-size: 11px;
  cursor: pointer;
}

.date-cell:hover {
  background: #f1f5f9;
}

.date-cell.out-month {
  color: #94a3b8;
}

.date-cell.selected {
  border-color: #1677ff;
  background: #1677ff;
  color: #fff;
}

.time-block {
  display: flex;
  gap: 5px;
}

.time-column {
  width: 42px;
}

.time-title {
  height: 18px;
  font-size: 10px;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
}

.time-list {
  height: 154px;
  padding: 3px 2px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  overflow-y: auto;
  background: #f8fafc;
}

.time-item {
  width: 100%;
  height: 22px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #334155;
  font-size: 12px;
  cursor: pointer;
}

.time-item:hover {
  background: #e2e8f0;
}

.time-item.selected {
  background: #1677ff;
  color: #fff;
}

.panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid #f1f5f9;
  padding: 7px 8px;
}

.now-btn {
  border: 0;
  background: transparent;
  color: #1677ff;
  font-size: 11px;
  cursor: pointer;
}

.footer-actions {
  display: flex;
  gap: 5px;
}

.cancel-btn,
.confirm-btn {
  min-width: 42px;
  height: 24px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
}

.cancel-btn {
  border: 1px solid #dbe3ef;
  color: #334155;
  background: #fff;
}

.cancel-btn:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.confirm-btn {
  border: 1px solid #1677ff;
  background: #1677ff;
  color: #fff;
}

.confirm-btn:hover {
  background: #0958d9;
  border-color: #0958d9;
}

@media (max-width: 760px) {
  .datetime-panel {
    left: auto;
    right: 0;
  }

  .panel-body {
    grid-template-columns: 1fr;
  }

  .time-block {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
