<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { CheckIcon } from 'lucide-vue-next'
import { THEMES, applyTheme, getStoredTheme } from '@/util/theme'

const modeOptions = [
  { label: '跟随系统', value: 'system' },
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' }
]

interface SidebarPreset {
  color: string
  builtin: boolean
}

const BUILTIN_SIDEBAR_PRESETS: SidebarPreset[] = [
  { color: '#0b3c51', builtin: true }
]

const currentThemeColor = ref(getStoredTheme())
const customColor = ref('#1890ff')
const customColorInput = ref('#1890ff')
const sidebarPresets = ref<SidebarPreset[]>([])
const saving = ref(false)

const state = reactive({
  theme_color: 'blue',
  theme_mode: 'system',
  sidebar_bg: '#0b3c51'
})

const applyThemeMode = (mode: string) => {
  const value = mode === 'light' || mode === 'dark' || mode === 'system' ? mode : 'system'
  const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  const effective = value === 'system' ? (systemDark ? 'dark' : 'light') : value
  if (effective === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  localStorage.setItem('themePreference', value)
}

const applySidebarBg = (value: string) => {
  const color = (value || '#0b3c51').trim() || '#0b3c51'
  document.documentElement.style.setProperty('--sidebar-bg', color)
}

const changeTheme = (themeKey: string) => {
  currentThemeColor.value = themeKey
  state.theme_color = themeKey
  applyTheme(themeKey)
}

const applyCustomTheme = (raw: string) => {
  const value = raw.trim() || '#1890ff'
  customColorInput.value = value
  if (value.startsWith('#') && value.length >= 4) {
    customColor.value = value
  }
  const key = `custom:${value}`
  currentThemeColor.value = key
  state.theme_color = key
  applyTheme(key)
}

const changeCustomTheme = (event: Event) => {
  const value = (event.target as HTMLInputElement).value || '#1890ff'
  applyCustomTheme(value)
}

const applyCustomThemeFromInput = () => {
  applyCustomTheme(customColorInput.value)
}

const loadSidebarPresets = () => {
  let userColors: string[] = []
  try {
    const stored = localStorage.getItem('sidebarBgPresets')
    if (stored) {
      userColors = JSON.parse(stored) || []
    }
  } catch {
    userColors = []
  }

  sidebarPresets.value = [
    ...BUILTIN_SIDEBAR_PRESETS,
    ...userColors
      .filter(c => !BUILTIN_SIDEBAR_PRESETS.some(b => b.color === c))
      .map(color => ({ color, builtin: false })),
  ]
}

const saveSidebarPresets = () => {
  const userColors = sidebarPresets.value.filter(p => !p.builtin).map(p => p.color)
  try {
    localStorage.setItem('sidebarBgPresets', JSON.stringify(userColors))
  } catch {
    // ignore
  }
}

const addSidebarPreset = () => {
  const color = (state.sidebar_bg || '').trim()
  if (!color) return
  if (sidebarPresets.value.some(p => p.color === color)) return
  sidebarPresets.value.push({ color, builtin: false })
  saveSidebarPresets()
}

const removeSidebarPreset = (preset: SidebarPreset) => {
  if (preset.builtin) return
  sidebarPresets.value = sidebarPresets.value.filter(p => p.color !== preset.color)
  saveSidebarPresets()
}

const load = async () => {
  const rsp = await request.get('/profile/theme')
  const data = rsp?.data?.data || {}
  state.theme_color = data.theme_color || 'blue'
  state.theme_mode = data.theme_mode || 'system'
  state.sidebar_bg = data.sidebar_bg || '#0b3c51'
  
  currentThemeColor.value = state.theme_color
  if (state.theme_color.startsWith('custom:')) {
    const raw = state.theme_color.split(':')[1]
    customColorInput.value = raw || '#1890ff'
    customColor.value = raw && raw.startsWith('#') ? raw : '#1890ff'
  } else {
    customColorInput.value = '#1890ff'
    customColor.value = '#1890ff'
  }
  
  applyTheme(state.theme_color)
  applyThemeMode(state.theme_mode)
  applySidebarBg(state.sidebar_bg)
}

const save = async () => {
  saving.value = true
  try {
    await request.post('/profile/theme', {
      theme_color: state.theme_color,
      theme_mode: state.theme_mode,
      sidebar_bg: state.sidebar_bg
    })
    applyTheme(state.theme_color)
    applyThemeMode(state.theme_mode)
    applySidebarBg(state.sidebar_bg)
    toast.success('个性设置已保存')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  loadSidebarPresets()
  await load()
})

watch(
  () => state.sidebar_bg,
  (val) => {
    applySidebarBg(val)
  }
)
</script>

<template>
  <div class="space-y-5">
    <!-- 主题颜色 -->
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">主题颜色</div>
      <div class="space-y-3">
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
          <button
            v-for="t in THEMES"
            :key="t.key"
            @click="changeTheme(t.key)"
            class="group relative flex items-center gap-2.5 p-2.5 rounded-lg border transition-all duration-200"
            :class="[
              currentThemeColor === t.key
                ? 'border-brand bg-brand/5 shadow-sm ring-1 ring-brand/20'
                : 'border-border hover:border-brand/40 hover:bg-muted/30'
            ]"
          >
            <div class="w-4 h-4 rounded-full shadow-inner border border-white/20 flex-shrink-0"
              :style="{ backgroundColor: t.light }"></div>
            <span class="text-xs font-medium truncate"
              :class="currentThemeColor === t.key ? 'text-brand' : 'text-foreground/80'">{{ t.name }}</span>
            <div v-if="currentThemeColor === t.key"
              class="absolute -top-1 -right-1 w-3.5 h-3.5 rounded-full bg-brand text-white flex items-center justify-center shadow-sm">
              <CheckIcon class="w-2 h-2" />
            </div>
          </button>
          <button
            class="group relative flex items-center gap-2.5 p-2.5 rounded-lg border transition-all duration-200"
            :class="[
              currentThemeColor.startsWith('custom:')
                ? 'border-brand bg-brand/5 shadow-sm ring-1 ring-brand/20'
                : 'border-border hover:border-brand/40 hover:bg-muted/30'
            ]"
          >
            <input
              type="color"
              :value="customColor"
              @input.stop="changeCustomTheme"
              class="w-6 h-6 rounded border border-border bg-transparent cursor-pointer"
            />
            <Input
              v-model="customColorInput"
              placeholder="#1890ff 或 rgb(24,144,255)"
              class="h-8 text-xs"
              @keyup.enter="applyCustomThemeFromInput"
              @blur="applyCustomThemeFromInput"
            />
          </button>
        </div>
      </div>
    </div>

    <!-- 主题模式 -->
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">主题模式</div>
      <div class="flex items-center gap-2 flex-wrap">
        <button
          v-for="mode in modeOptions"
          :key="mode.value"
          type="button"
          class="px-3 py-1.5 text-sm rounded border transition-colors"
          :class="state.theme_mode === mode.value ? 'border-brand bg-brand/5 dark:bg-slate-800' : 'border-border hover:border-brand/40'"
          @click="state.theme_mode = mode.value"
        >
          {{ mode.label }}
        </button>
      </div>
    </div>

    <!-- 侧边栏背景色 -->
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">侧边栏背景色</div>
      <div class="space-y-2">
        <div v-if="sidebarPresets.length" class="flex flex-wrap gap-2">
          <button
            v-for="preset in sidebarPresets"
            :key="preset.color"
            type="button"
            class="flex items-center gap-2 px-2 py-1 rounded-md border text-xs"
            :class="state.sidebar_bg === preset.color ? 'border-brand bg-brand/5' : 'border-border hover:border-brand/40'"
            @click="state.sidebar_bg = preset.color"
          >
            <span
              class="w-4 h-4 rounded border border-border"
              :style="{ backgroundColor: preset.color }"
            />
            <span class="font-mono">{{ preset.color }}</span>
            <button
              v-if="!preset.builtin"
              type="button"
              class="ml-1 text-[10px] text-muted-foreground hover:text-red-500"
              @click.stop="removeSidebarPreset(preset)"
            >
              ×
            </button>
          </button>
        </div>
        <div class="flex items-center gap-2 mt-2">
          <input
            type="color"
            v-model="state.sidebar_bg"
            class="w-8 h-8 rounded border border-border bg-transparent cursor-pointer"
          />
          <Input
            v-model="state.sidebar_bg"
            placeholder="#0b3c51 或 rgb(11,60,81)"
            class="h-8 text-xs max-w-[220px]"
          />
          <Button
            type="button"
            size="sm"
            variant="outline"
            class="h-8 px-2 text-xs"
            @click="addSidebarPreset"
          >
            加入固定方案
          </Button>
        </div>
      </div>
    </div>

    <Button :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存个性设置' }}</Button>
  </div>
</template>

