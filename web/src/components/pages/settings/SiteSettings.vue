<script setup lang="ts">
import { reactive, onMounted, ref, watch, computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { HelpCircleIcon, CheckIcon, CheckCircle2, Circle, ChevronRight, Folder } from 'lucide-vue-next'
import { THEMES, applyTheme, getStoredTheme } from '@/util/theme'
// @ts-ignore
import config from '../../../../config.js'

const currentThemeColor = ref(getStoredTheme())
const customColor = ref('#1890ff')
const customColorInput = ref('#1890ff')

interface SidebarPreset {
  color: string
  builtin: boolean
}

const BUILTIN_SIDEBAR_PRESETS: SidebarPreset[] = [
  { color: '#0b3c51', builtin: true }
]

const sidebarPresets = ref<SidebarPreset[]>([])

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

const state = reactive({
  title: '',
  slogan: '',
  login_title: '',
  logo: '',
  logo_storage_profile_id: '',
  pagesize: '',
  cookieExpDays: '',
  sidebarBg: '#0b3c51',
  theme_color: getStoredTheme(),
  sloganInitialEnabled: false,
  channel_test_message: 'This is a test message from message-platform.',
  section: 'site_config',
})

// ===== 站点 Logo 上传裁剪 =====
const logoInputRef = ref<HTMLInputElement | null>(null)
const logoUploading = ref(false)
const logoClearing = ref(false)
const clearLogoConfirmOpen = ref(false)
const clearLogoDeleteSource = ref(false)
const logoCropDialogOpen = ref(false)
const logoCropImageUrl = ref('')
const logoCropImageElement = ref<HTMLImageElement | null>(null)
const logoCropImageName = ref('')
const logoCropScale = ref(1)
const logoCropMinScale = ref(1)
const logoCropMaxScale = ref(6)
const logoCropOffsetX = ref(0)
const logoCropOffsetY = ref(0)
const logoCropDragging = ref(false)
const logoCropDragStartX = ref(0)
const logoCropDragStartY = ref(0)
const logoCropDragInitOffsetX = ref(0)
const logoCropDragInitOffsetY = ref(0)
const logoCropViewportSize = 260
const logoStorageProfiles = ref<Array<{ id: string, name: string, provider: 'local' | 's3', s3_public_base_url: string }>>([])
const defaultStorageProfileID = ref('')
const logoBrowseDialogOpen = ref(false)
const logoBrowseLoading = ref(false)
const logoBrowseCurrentPath = ref('')
const logoBrowseParentPath = ref('')
const logoBrowseRootPath = ref('')
const logoBrowsePrefix = ref('')
const logoBrowseKeyword = ref('')
const logoBrowseViewMode = ref<'list' | 'thumb'>('thumb')
const logoBrowseThumbSize = ref<'sm' | 'md' | 'lg'>('md')
const logoBrowseDirectories = ref<Array<{ name: string, relative_path: string }>>([])
const logoBrowseFiles = ref<Array<{ name: string, relative_path: string, public_url?: string, object_key?: string, size?: number }>>([])

const selectedLogoStorageProfile = computed(() =>
  logoStorageProfiles.value.find(item => item.id === state.logo_storage_profile_id)
)

const isInlineSvgLogo = computed(() => state.logo.trimStart().startsWith('<'))
const showLogoStorageAdvanced = ref(false)
const logoStoragePanelExpanded = computed(() => !isInlineSvgLogo.value || showLogoStorageAdvanced.value)

watch(isInlineSvgLogo, (value) => {
  if (value) {
    showLogoStorageAdvanced.value = false
  }
})

const loadLogoStorageProfiles = async () => {
  const rsp = await request.get('/system/storage-config')
  const data = rsp?.data?.data || {}
  const list = Array.isArray(data.profiles) ? data.profiles : []
  defaultStorageProfileID.value = (data.default_storage_id || '').trim()
  logoStorageProfiles.value = list
    .map((item: any) => ({
      id: String(item?.id || '').trim(),
      name: String(item?.name || '').trim(),
      provider: String(item?.provider || '').trim().toLowerCase() === 's3' ? 's3' : 'local',
      s3_public_base_url: String(item?.s3_public_base_url || '').trim()
    }))
    .filter((item: any) => item.id)
  // 确保选择有效的存储配置
  if (logoStorageProfiles.value.length > 0) {
    const exists = logoStorageProfiles.value.some(item => item.id === state.logo_storage_profile_id)
    if (!exists) {
      const byDefault = logoStorageProfiles.value.find(item => item.id === defaultStorageProfileID.value)
      state.logo_storage_profile_id = byDefault?.id || logoStorageProfiles.value[0].id
    }
  }
}

const resolveLogoUrl = (url: string) => {
  if (!url) return ''
  if (/^https?:\/\//i.test(url) || url.startsWith('data:')) return url
  const base = `${config.apiUrl}`.replace(/\/+$/, '')
  const normalized = url.trim()
  const path = normalized.startsWith('/public/') || normalized.startsWith('/uploads/') || normalized.startsWith('/storage/')
    ? normalized
    : `/public/storage/local/${normalized.replace(/^\/+/, '')}`
  return `${base}${path}`
}

const normalizeUploadedLogoUrl = (url: string) => {
  const raw = (url || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw) || raw.startsWith('data:')) return raw
  const profile = selectedLogoStorageProfile.value
  if (profile?.provider === 's3' && profile.s3_public_base_url) {
    const match = raw.match(/\/public\/storage\/oidc-icons\/[^/]+\/(.+)$/)
    if (match?.[1]) {
      return `${profile.s3_public_base_url.replace(/\/+$/, '')}/${match[1].replace(/^\/+/, '')}`
    }
  }
  return raw
}

const cleanupLogoCropImage = () => {
  if (logoCropImageUrl.value) URL.revokeObjectURL(logoCropImageUrl.value)
  logoCropImageUrl.value = ''
  logoCropImageElement.value = null
  logoCropImageName.value = ''
  logoCropScale.value = 1
  logoCropOffsetX.value = 0
  logoCropOffsetY.value = 0
}

const getLogoCropDisplayWidth = () => (logoCropImageElement.value ? logoCropImageElement.value.width * logoCropScale.value : 0)
const getLogoCropDisplayHeight = () => (logoCropImageElement.value ? logoCropImageElement.value.height * logoCropScale.value : 0)

const clampLogoCropOffset = (offsetX: number, offsetY: number) => {
  const displayWidth = getLogoCropDisplayWidth()
  const displayHeight = getLogoCropDisplayHeight()
  if (displayWidth <= 0 || displayHeight <= 0) return { offsetX: 0, offsetY: 0 }
  const minX = logoCropViewportSize - displayWidth
  const minY = logoCropViewportSize - displayHeight
  return {
    offsetX: Math.min(0, Math.max(minX, offsetX)),
    offsetY: Math.min(0, Math.max(minY, offsetY))
  }
}

const applyLogoCropScale = (newScale: number) => {
  if (!logoCropImageElement.value) return
  const nextScale = Math.max(logoCropMinScale.value, Math.min(logoCropMaxScale.value, newScale))
  const currentCenterX = (logoCropViewportSize / 2 - logoCropOffsetX.value) / logoCropScale.value
  const currentCenterY = (logoCropViewportSize / 2 - logoCropOffsetY.value) / logoCropScale.value
  logoCropScale.value = nextScale
  const nextOffsetX = logoCropViewportSize / 2 - currentCenterX * nextScale
  const nextOffsetY = logoCropViewportSize / 2 - currentCenterY * nextScale
  const clamped = clampLogoCropOffset(nextOffsetX, nextOffsetY)
  logoCropOffsetX.value = clamped.offsetX
  logoCropOffsetY.value = clamped.offsetY
}

const zoomInLogoCrop = () => applyLogoCropScale(logoCropScale.value + 0.08)
const zoomOutLogoCrop = () => applyLogoCropScale(logoCropScale.value - 0.08)

const openLogoCropDialog = (file: File): Promise<void> => {
  return new Promise((resolve, reject) => {
    const objectUrl = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      logoCropImageElement.value = img
      logoCropImageUrl.value = objectUrl
      logoCropImageName.value = file.name
      const minScale = Math.max(logoCropViewportSize / img.width, logoCropViewportSize / img.height)
      logoCropMinScale.value = minScale
      logoCropScale.value = minScale
      const displayWidth = img.width * minScale
      const displayHeight = img.height * minScale
      logoCropOffsetX.value = (logoCropViewportSize - displayWidth) / 2
      logoCropOffsetY.value = (logoCropViewportSize - displayHeight) / 2
      logoCropDialogOpen.value = true
      resolve()
    }
    img.onerror = () => {
      URL.revokeObjectURL(objectUrl)
      reject(new Error('图片读取失败'))
    }
    img.src = objectUrl
  })
}

const buildLogoCropBlob = (): Promise<Blob> => {
  return new Promise((resolve, reject) => {
    const image = logoCropImageElement.value
    if (!image) { reject(new Error('未选择图片')); return }
    const outputSize = 128
    const canvas = document.createElement('canvas')
    canvas.width = outputSize
    canvas.height = outputSize
    const ctx = canvas.getContext('2d')
    if (!ctx) { reject(new Error('图像处理失败')); return }
    const srcX = -logoCropOffsetX.value / logoCropScale.value
    const srcY = -logoCropOffsetY.value / logoCropScale.value
    const srcSize = logoCropViewportSize / logoCropScale.value
    ctx.clearRect(0, 0, outputSize, outputSize)
    ctx.drawImage(image, srcX, srcY, srcSize, srcSize, 0, 0, outputSize, outputSize)
    canvas.toBlob((blob) => {
      if (!blob) { reject(new Error('图像处理失败')); return }
      resolve(blob)
    }, 'image/png')
  })
}

const uploadLogoCroppedBlob = async (blob: Blob) => {
  const formData = new FormData()
  formData.append('file', new File([blob], 'site-logo.png', { type: 'image/png' }))
  formData.append('storage_profile_id', state.logo_storage_profile_id)
  const rsp = await request.post('/system/site-logo/upload', formData)
  if (rsp?.data?.code !== 200) {
    toast.error(rsp?.data?.msg || '上传失败')
    return
  }
  const data = rsp?.data?.data || {}
  state.logo = normalizeUploadedLogoUrl(data.url || '')
  state.logo_storage_profile_id = data.storage_profile_id || state.logo_storage_profile_id
  toast.success('站点图标上传成功')
}

const onLogoCropPointerDown = (event: PointerEvent) => {
  if (!logoCropImageElement.value) return
  logoCropDragging.value = true
  logoCropDragStartX.value = event.clientX
  logoCropDragStartY.value = event.clientY
  logoCropDragInitOffsetX.value = logoCropOffsetX.value
  logoCropDragInitOffsetY.value = logoCropOffsetY.value
}

const onLogoCropPointerMove = (event: PointerEvent) => {
  if (!logoCropDragging.value) return
  const diffX = event.clientX - logoCropDragStartX.value
  const diffY = event.clientY - logoCropDragStartY.value
  const clamped = clampLogoCropOffset(logoCropDragInitOffsetX.value + diffX, logoCropDragInitOffsetY.value + diffY)
  logoCropOffsetX.value = clamped.offsetX
  logoCropOffsetY.value = clamped.offsetY
}

const onLogoCropWheel = (event: WheelEvent) => {
  event.preventDefault()
  const delta = event.deltaY > 0 ? -0.05 : 0.05
  applyLogoCropScale(logoCropScale.value + delta)
}

const stopLogoCropDragging = () => { logoCropDragging.value = false }

const closeLogoCropDialog = () => {
  logoCropDialogOpen.value = false
  stopLogoCropDragging()
  cleanupLogoCropImage()
}

const confirmLogoCropAndUpload = async () => {
  logoUploading.value = true
  try {
    const blob = await buildLogoCropBlob()
    await uploadLogoCroppedBlob(blob)
    closeLogoCropDialog()
  } catch (error: any) {
    toast.error(error?.message || '图标上传失败')
  } finally {
    logoUploading.value = false
  }
}

const onSelectLogoFile = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  if (file.size > 2 * 1024 * 1024) {
    toast.error('图片不能超过 2MB')
    target.value = ''
    return
  }
  try {
    await openLogoCropDialog(file)
  } catch (error: any) {
    toast.error(error?.message || '图片读取失败')
  }
  target.value = ''
}

const openClearLogoConfirm = () => {
  if (!state.logo || logoUploading.value || logoClearing.value) return
  clearLogoConfirmOpen.value = true
}

const clearSiteLogo = async () => {
  logoClearing.value = true
  try {
    const rsp = await request.post('/system/site-logo/clear', {
      delete_source: clearLogoDeleteSource.value
    })
    if (rsp?.data?.code !== 200) {
      toast.error(rsp?.data?.msg || '恢复默认失败')
      return
    }
    await getSiteConfig()
    toast.success(clearLogoDeleteSource.value ? '已恢复默认图标，并删除源文件' : '已恢复默认图标')
    clearLogoConfirmOpen.value = false
  } finally {
    logoClearing.value = false
  }
}

const isImageLikeFile = (name: string) => /\.(png|jpe?g|webp|gif|svg)$/i.test((name || '').trim())

const loadLogoBrowseFiles = async (path: string) => {
  const profile = selectedLogoStorageProfile.value
  if (!profile?.id) return
  logoBrowseLoading.value = true
  try {
    if (profile.provider === 's3') {
      const rsp = await request.get('/system/storage-config/s3-objects', {
        params: { profile_id: profile.id, path }
      })
      const data = rsp?.data?.data || {}
      logoBrowseCurrentPath.value = data.current_path || ''
      logoBrowseParentPath.value = data.parent_path || ''
      logoBrowsePrefix.value = data.prefix || ''
      logoBrowseRootPath.value = ''
      logoBrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
      logoBrowseFiles.value = Array.isArray(data.files) ? data.files : []
      return
    }
    const rsp = await request.get('/system/storage-config/local-files', {
      params: { profile_id: profile.id, path }
    })
    const data = rsp?.data?.data || {}
    logoBrowseCurrentPath.value = data.current_path || ''
    logoBrowseParentPath.value = data.parent_path || ''
    logoBrowseRootPath.value = data.root_path || ''
    logoBrowsePrefix.value = ''
    logoBrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
    logoBrowseFiles.value = Array.isArray(data.files) ? data.files : []
  } catch (error: any) {
    toast.error(error?.response?.data?.msg || '读取存储文件失败')
  } finally {
    logoBrowseLoading.value = false
  }
}

const openLogoBrowseDialog = async () => {
  if (!state.logo_storage_profile_id) {
    toast.error('请先选择图标存储')
    return
  }
  logoBrowseDialogOpen.value = true
  logoBrowseKeyword.value = ''
  logoBrowseViewMode.value = 'thumb'
  await loadLogoBrowseFiles('')
}

const openLogoBrowseChild = async (item: { relative_path: string }) => {
  await loadLogoBrowseFiles(item.relative_path || '')
}

const openLogoBrowseParent = async () => {
  await loadLogoBrowseFiles(logoBrowseParentPath.value || '')
}

const logoBrowseBreadcrumbs = computed(() => {
  const profile = selectedLogoStorageProfile.value
  const rootLabel = profile?.provider === 's3'
    ? (logoBrowsePrefix.value || '根目录')
    : (logoBrowseRootPath.value || 'uploads')
  const items: Array<{ label: string, path: string }> = [{ label: rootLabel, path: '' }]
  const current = (logoBrowseCurrentPath.value || '').trim()
  if (!current) return items
  const parts = current.split('/').filter(Boolean)
  let acc = ''
  for (const part of parts) {
    acc = acc ? `${acc}/${part}` : part
    items.push({ label: part, path: acc })
  }
  return items
})

const openLogoBrowseBreadcrumb = async (path: string) => {
  await loadLogoBrowseFiles(path || '')
}

const filteredLogoBrowseDirectories = computed(() => {
  const keyword = logoBrowseKeyword.value.trim().toLowerCase()
  if (!keyword) return logoBrowseDirectories.value
  return logoBrowseDirectories.value.filter(item => (item.name || '').toLowerCase().includes(keyword))
})

const filteredLogoBrowseFiles = computed(() => {
  const keyword = logoBrowseKeyword.value.trim().toLowerCase()
  const images = logoBrowseFiles.value.filter(item => isImageLikeFile(item.name || item.relative_path || item.object_key || ''))
  if (!keyword) return images
  return images.filter(item => (item.name || '').toLowerCase().includes(keyword))
})

const logoBrowseThumbGridClass = computed(() => {
  if (logoBrowseThumbSize.value === 'sm') return 'grid grid-cols-3 md:grid-cols-5 lg:grid-cols-6 gap-2'
  if (logoBrowseThumbSize.value === 'lg') return 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
  return 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3'
})

const logoBrowseThumbPreviewClass = computed(() => {
  if (logoBrowseThumbSize.value === 'sm') return 'w-full h-20 rounded border bg-slate-50 dark:bg-slate-900 overflow-hidden flex items-center justify-center'
  if (logoBrowseThumbSize.value === 'lg') return 'w-full h-36 rounded border bg-slate-50 dark:bg-slate-900 overflow-hidden flex items-center justify-center'
  return 'w-full h-28 rounded border bg-slate-50 dark:bg-slate-900 overflow-hidden flex items-center justify-center'
})

const logoBrowseThumbFiles = computed(() =>
  filteredLogoBrowseFiles.value.map((file) => {
    const label = file.object_key || file.relative_path || file.name
    const previewUrl = resolveLogoUrl(file.public_url || '')
    return {
      ...file,
      label,
      preview_url: previewUrl,
      can_preview: Boolean(previewUrl)
    }
  })
)

const applyLogoFromBrowse = (file: { name: string, public_url?: string, object_key?: string, relative_path: string }) => {
  const profile = selectedLogoStorageProfile.value
  const rawUrl = (file.public_url || '').trim()
  if (profile?.provider === 's3' && !rawUrl) {
    toast.error('当前 S3 文件缺少 public_url，无法作为站点图标')
    return
  }
  if (!rawUrl) {
    toast.error('文件地址为空，无法选择')
    return
  }
  state.logo = rawUrl
  logoBrowseDialogOpen.value = false
  toast.success(`已选择图标：${file.name}`)
}

const applySidebarBg = (value: string) => {
  const color = (value || '#0b3c51').trim() || '#0b3c51'
  document.documentElement.style.setProperty('--sidebar-bg', color)
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
  }
}

const addSidebarPreset = () => {
  const color = (state.sidebarBg || '').trim()
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

// 提交设置
const handleSubmit = async () => {
  try {
    const postData = {
      section: state.section,
      data: {
        title: state.title.trim(),
        slogan: state.slogan.trim(),
        login_title: state.login_title.trim(),
        logo: state.logo.trim(),
        logo_storage_profile_id: state.logo_storage_profile_id.trim(),
        pagesize: state.pagesize.toString(),
        cookie_exp_days: state.cookieExpDays.toString(),
        sidebar_bg: state.sidebarBg,
        theme_color: state.theme_color,
        slogan_initial_enabled: state.sloganInitialEnabled ? 'true' : 'false',
        channel_test_message: state.channel_test_message.trim(),
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      const msg = response.data.msg
      toast.success(msg)
    }
  } catch (error) {
    toast.error('保存失败，请稍后重试')
  }
}

// 恢复默认设置
const handleSubmitReset = async () => {
  try {
    const response = await request.post('/settings/reset', {})
    if (response.data.code === 200) {
      const msg = response.data.msg
      toast.success(msg)
      // 重新获取设置
      await getSiteConfig()
    }
  } catch (error) {
    toast.error('恢复默认设置失败，请稍后重试')
  }
}

// 获取站点配置
const getSiteConfig = async () => {
  try {
    const params = { params: { section: 'site_config' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      state.title = data.title || ''
      state.logo = data.logo || ''
      state.logo_storage_profile_id = data.logo_storage_profile_id || ''
      state.slogan = data.slogan || ''
      state.login_title = data.login_title || ''
      state.pagesize = data.pagesize || ''
      state.cookieExpDays = data.cookie_exp_days || '1'
      state.sidebarBg = data.sidebar_bg || '#0b3c51'
      state.theme_color = data.theme_color || getStoredTheme()
      state.sloganInitialEnabled = String(data.slogan_initial_enabled || 'false') === 'true'
      state.channel_test_message = data.channel_test_message || 'This is a test message from ops-message-unified-push.'

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
      applySidebarBg(state.sidebarBg)

      // 确保存储配置选择有效
      if (logoStorageProfiles.value.length > 0) {
        const exists = logoStorageProfiles.value.some(item => item.id === state.logo_storage_profile_id)
        if (!exists) {
          const byDefault = logoStorageProfiles.value.find(item => item.id === defaultStorageProfileID.value)
          state.logo_storage_profile_id = byDefault?.id || logoStorageProfiles.value[0].id
        }
      }

      LocalStieConfigUtils.updateLocalConfig(data)
    }
  } catch (error) {
    toast.error('获取配置失败')
  }
}

onMounted(async () => {
  loadSidebarPresets()
  await loadLogoStorageProfiles()
  await getSiteConfig()
})

watch(
  () => state.sidebarBg,
  (val) => {
    applySidebarBg(val)
  }
)
</script>

<script lang="ts">
export default {
  name: 'SiteSettings'
}
</script>

<template>
  <div class="space-y-5">
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">基本设置</div>
      <div class="space-y-4">
        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">站点标题</label>
          <Input v-model="state.title" placeholder="请输入自定义的网站标题" />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">站点标语</label>
          <Input v-model="state.slogan" placeholder="请输入自定义的网站slogan" />
          <div class="flex items-center justify-between rounded-md border border-dashed border-slate-300/80 dark:border-slate-700 px-3 py-2">
            <div>
              <div class="text-sm text-gray-700 dark:text-slate-200">菜单侧边栏收起状态左上角显示字母跟随标语首字母</div>
              <div class="text-xs text-muted-foreground">开启后优先取标语首字母；关闭后固定使用默认值 M</div>
            </div>
            <Switch
              :model-value="state.sloganInitialEnabled"
              @update:model-value="(val) => state.sloganInitialEnabled = val === true"
            />
          </div>
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">登录页标题</label>
          <Input v-model="state.login_title" placeholder="登录页显示的标题，默认：消 息 统 一 推 送 中 台" />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">站点图标</label>
          <!-- 存储选择 -->
          <div v-if="logoStorageProfiles.length > 0" class="space-y-1">
            <div class="flex items-center justify-between gap-2">
              <div class="text-xs text-muted-foreground">图标存储</div>
              <Button
                v-if="isInlineSvgLogo"
                type="button"
                variant="ghost"
                size="sm"
                class="h-7 px-2 text-xs text-muted-foreground"
                @click="showLogoStorageAdvanced = !showLogoStorageAdvanced"
              >
                {{ showLogoStorageAdvanced ? '收起高级选项' : '高级选项' }}
              </Button>
            </div>
            <div v-if="logoStoragePanelExpanded" class="space-y-1">
              <select
                v-model="state.logo_storage_profile_id"
                class="w-full h-9 px-3 rounded-md border border-input bg-background text-sm"
              >
                <option v-for="profile in logoStorageProfiles" :key="profile.id" :value="profile.id">
                  {{ profile.name }}（{{ profile.provider === 's3' ? 'S3' : '本地' }} / {{ profile.id }}）
                </option>
              </select>
              <div v-if="isInlineSvgLogo" class="text-xs text-amber-600">
                当前为默认/SVG图标，显示效果不依赖存储类型；存储选择仅在“上传并裁剪”或“浏览”时生效。
              </div>
            </div>
            <div v-else class="text-xs text-muted-foreground">
              当前为默认/SVG图标，显示效果不依赖存储类型。需要上传或浏览存储文件时可展开“高级选项”。
            </div>
          </div>
          <!-- 上传操作 -->
          <input ref="logoInputRef" type="file" accept=".png,.jpg,.jpeg,.webp" class="hidden" @change="onSelectLogoFile">
          <div class="flex items-center gap-2 flex-wrap">
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="logoUploading || logoClearing || !state.logo_storage_profile_id"
              @click="logoInputRef?.click()"
            >
              {{ logoUploading ? '上传中...' : '上传并裁剪' }}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="logoUploading || logoClearing || !state.logo_storage_profile_id"
              @click="openLogoBrowseDialog"
            >
              浏览
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="logoUploading || logoClearing || !state.logo"
              @click="openClearLogoConfirm"
            >
              {{ logoClearing ? '恢复中...' : '恢复默认图标' }}
            </Button>
            <button
              type="button"
              class="inline-flex items-center gap-1.5 text-xs text-slate-700 dark:text-slate-200"
              :title="clearLogoDeleteSource ? '恢复默认图标时同步删除存储中的源文件' : '恢复默认图标时仅清理配置，不删除存储源文件'"
              @click="clearLogoDeleteSource = !clearLogoDeleteSource"
            >
              <CheckCircle2 v-if="clearLogoDeleteSource" class="w-3.5 h-3.5 text-brand-600" />
              <Circle v-else class="w-3.5 h-3.5 text-slate-400" />
              <span>恢复时同步删除源文件</span>
            </button>
          </div>
          <div class="text-xs text-muted-foreground">支持 jpg/png/webp，最大 2MB，上传后自动裁剪为方图</div>
          <!-- 已上传预览 -->
          <div v-if="state.logo" class="flex items-center gap-3 rounded border p-2 bg-muted/30">
            <!-- 旧数据：SVG 文本，直接渲染 -->
            <div
              v-if="state.logo.trimStart().startsWith('<')"
              class="w-8 h-8 flex-shrink-0 flex items-center justify-center overflow-hidden"
              v-html="state.logo"
            />
            <!-- 新数据：图片 URL -->
            <img
              v-else
              :src="resolveLogoUrl(state.logo)"
              alt="site-logo"
              class="w-8 h-8 flex-shrink-0 rounded object-cover"
            >
            <div class="text-xs text-muted-foreground break-all flex-1 line-clamp-2">
              {{ state.logo.trimStart().startsWith('<') ? '（SVG 文本，不依赖存储类型；建议上传图片替换）' : state.logo }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">个性设置</div>
      <div class="space-y-4">
        <div class="space-y-3">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">主题色</label>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <button v-for="t in THEMES" :key="t.key" @click="changeTheme(t.key)"
              class="group relative flex items-center gap-2.5 p-2.5 rounded-lg border transition-all duration-200"
              :class="[
                currentThemeColor === t.key
                  ? 'border-brand bg-brand/5 shadow-sm ring-1 ring-brand/20'
                  : 'border-border hover:border-brand/40 hover:bg-muted/30'
              ]">
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

        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">侧边栏背景色</label>
          <div v-if="sidebarPresets.length" class="flex flex-wrap gap-2">
            <button
              v-for="preset in sidebarPresets"
              :key="preset.color"
              type="button"
              class="flex items-center gap-2 px-2 py-1 rounded-md border text-xs"
              :class="state.sidebarBg === preset.color ? 'border-brand bg-brand/5' : 'border-border hover:border-brand/40'"
              @click="state.sidebarBg = preset.color"
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
              v-model="state.sidebarBg"
              class="w-8 h-8 rounded border border-border bg-transparent cursor-pointer"
            />
            <Input
              v-model="state.sidebarBg"
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
    </div>

    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">系统参数</div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">分页大小</label>
          <Input v-model="state.pagesize" placeholder="页面分页大小" />
        </div>
        <div class="space-y-2">
          <label class="text-sm font-medium text-gray-700 dark:text-slate-200">Cookie过期天数</label>
          <Input v-model="state.cookieExpDays" type="number" min="1" max="365" placeholder="Cookie过期天数（默认1天）" />
        </div>
      </div>
    </div>

    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">渠道测试默认文案</div>
      <div class="space-y-2">
        <label class="text-sm font-medium text-gray-700 dark:text-slate-200">测试消息正文</label>
        <Textarea
          v-model="state.channel_test_message"
          :max-length="2000"
          rows="4"
          placeholder="请输入渠道测试按钮默认发送的消息正文"
        />
        <div class="flex items-center justify-between text-xs text-muted-foreground">
          <span>用于“渠道管理-新增/编辑渠道”右下角测试按钮的默认消息内容。</span>
          <span>{{ state.channel_test_message.length }}/2000</span>
        </div>
      </div>
    </div>

    <div class="flex items-center justify-between pt-2 border-t border-slate-200/80 dark:border-slate-700/80">
      <div class="flex items-center space-x-2">
        <span class="text-sm text-gray-600 dark:text-slate-300">说明</span>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <HelpCircleIcon class="w-4 h-4 text-gray-400 hover:text-gray-600 dark:hover:text-slate-200" />
            </TooltipTrigger>
            <TooltipContent class="max-w-sm">
              <div class="text-sm space-y-1">
                <p>1. logo请输入svg文本，替换后登录页面，ico，导航栏logo将全部一起更换</p>
                <p>2. slogan将在登录页面展示</p>
                <p>3. Cookie过期天数设置用户登录后的有效期，修改后下次登录时生效</p>
                <p>4. 修改后在下次登录时生效，如不生效请在登录页面Ctrl+F5强制刷新</p>
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
      <div class="flex space-x-2">
        <Button variant="outline" size="sm" @click="handleSubmitReset">恢复默认</Button>
        <Button size="sm" @click="handleSubmit">确定</Button>
      </div>
    </div>
  </div>

  <!-- 清空Logo确认弹窗 -->
  <Dialog :open="clearLogoConfirmOpen" @update:open="(value) => clearLogoConfirmOpen = value">
    <DialogContent class="max-w-[420px]">
      <DialogHeader>
        <DialogTitle>确认恢复默认站点图标</DialogTitle>
      </DialogHeader>
      <div class="space-y-2 text-sm text-slate-700 dark:text-slate-200">
        <div>恢复后站点将使用默认图标。</div>
        <div v-if="clearLogoDeleteSource" class="text-red-500">同时会删除存储中的源文件，此操作不可恢复。</div>
      </div>
      <DialogFooter>
        <Button type="button" variant="outline" :disabled="logoClearing" @click="clearLogoConfirmOpen = false">取消</Button>
        <Button type="button" :disabled="logoClearing" @click="clearSiteLogo">{{ logoClearing ? '恢复中...' : '确认恢复' }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- 从存储浏览选择站点图标 -->
  <Dialog :open="logoBrowseDialogOpen" @update:open="(value) => logoBrowseDialogOpen = value">
    <DialogContent class="w-[min(855px,98vw)] !max-w-[98vw] sm:!max-w-[98vw] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader class="flex-shrink-0 border-b border-border/60 pb-3">
        <DialogTitle>从存储选择站点图标</DialogTitle>
      </DialogHeader>
      <div class="flex-1 overflow-y-auto mt-4 space-y-3">
        <div class="text-sm text-muted-foreground">
          当前存储：{{ selectedLogoStorageProfile?.name || '-' }}（{{ selectedLogoStorageProfile?.provider === 's3' ? 'S3' : '本地' }} / {{ state.logo_storage_profile_id || '-' }}）
        </div>
        <div class="min-w-0 overflow-x-auto whitespace-nowrap pb-1 text-sm text-muted-foreground">
          <button
            v-for="(crumb, index) in logoBrowseBreadcrumbs"
            :key="`${crumb.path || 'logo-root'}-${index}`"
            type="button"
            class="inline-flex items-center gap-1 hover:text-brand-600 mr-1"
            :disabled="logoBrowseLoading"
            @click="openLogoBrowseBreadcrumb(crumb.path)"
          >
            <ChevronRight v-if="index > 0" class="w-3.5 h-3.5" />
            <Folder v-if="index === 0" class="w-3.5 h-3.5" />
            <span class="max-w-[220px] truncate align-bottom">{{ crumb.label }}</span>
          </button>
        </div>
        <div class="grid grid-cols-[auto_minmax(240px,1fr)_auto_auto] items-center gap-2">
          <Button type="button" variant="outline" size="sm" :disabled="!logoBrowseCurrentPath || logoBrowseLoading" @click="openLogoBrowseParent">
            返回上级
          </Button>
          <Input v-model="logoBrowseKeyword" class="h-8 w-full" placeholder="按名称筛选目录/图片" />
          <div class="inline-flex items-center gap-1 rounded border p-1">
            <button
              type="button"
              class="px-2 py-1 text-xs rounded"
              :class="logoBrowseViewMode === 'list' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
              @click="logoBrowseViewMode = 'list'"
            >
              列表
            </button>
            <button
              type="button"
              class="px-2 py-1 text-xs rounded"
              :class="logoBrowseViewMode === 'thumb' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
              @click="logoBrowseViewMode = 'thumb'"
            >
              缩略图
            </button>
          </div>
          <div
            class="inline-flex w-[132px] items-center gap-1 rounded border p-1 transition-opacity"
            :class="logoBrowseViewMode === 'thumb' ? 'opacity-100' : 'opacity-0 pointer-events-none'"
          >
            <button
              type="button"
              class="px-2 py-1 text-xs rounded"
              :class="logoBrowseThumbSize === 'sm' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
              @click="logoBrowseThumbSize = 'sm'"
            >
              小
            </button>
            <button
              type="button"
              class="px-2 py-1 text-xs rounded"
              :class="logoBrowseThumbSize === 'md' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
              @click="logoBrowseThumbSize = 'md'"
            >
              中
            </button>
            <button
              type="button"
              class="px-2 py-1 text-xs rounded"
              :class="logoBrowseThumbSize === 'lg' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
              @click="logoBrowseThumbSize = 'lg'"
            >
              大
            </button>
          </div>
        </div>
        <div v-if="logoBrowseViewMode === 'list'" class="space-y-3">
          <div class="rounded-md border">
            <div class="px-3 py-2 border-b text-xs text-muted-foreground">目录（{{ filteredLogoBrowseDirectories.length }}）</div>
            <div class="max-h-40 overflow-y-auto divide-y">
              <button
                v-for="item in filteredLogoBrowseDirectories"
                :key="item.relative_path"
                type="button"
                class="w-full text-left px-3 py-2 hover:bg-muted/50"
                @click="openLogoBrowseChild(item)"
              >
                <div class="text-sm flex items-center gap-2"><Folder class="w-4 h-4" />{{ item.name }}</div>
                <div class="text-xs text-muted-foreground">/{{ item.relative_path }}</div>
              </button>
              <div v-if="!filteredLogoBrowseDirectories.length" class="px-3 py-3 text-xs text-muted-foreground">暂无目录</div>
            </div>
          </div>
          <div class="rounded-md border">
            <div class="px-3 py-2 border-b text-xs text-muted-foreground">图片文件（{{ filteredLogoBrowseFiles.length }}）</div>
            <div class="max-h-56 overflow-y-auto divide-y">
              <div
                v-for="file in filteredLogoBrowseFiles"
                :key="file.object_key || file.relative_path"
                class="px-3 py-2 flex items-center justify-between gap-3"
              >
                <div class="min-w-0">
                  <div class="text-sm truncate">{{ file.name }}</div>
                  <div class="text-xs text-muted-foreground truncate">
                    {{ selectedLogoStorageProfile?.provider === 's3' ? (file.object_key || file.relative_path) : file.relative_path }}
                  </div>
                </div>
                <Button type="button" size="sm" @click="applyLogoFromBrowse(file)">使用</Button>
              </div>
              <div v-if="!filteredLogoBrowseFiles.length" class="px-3 py-3 text-xs text-muted-foreground">当前目录暂无可用图片文件</div>
            </div>
          </div>
        </div>
        <div v-else class="rounded border max-h-[56vh] overflow-y-auto p-3 space-y-3">
          <div v-if="filteredLogoBrowseDirectories.length > 0" class="grid grid-cols-2 md:grid-cols-3 gap-2">
            <button
              v-for="item in filteredLogoBrowseDirectories"
              :key="`thumb-dir-${item.relative_path}`"
              type="button"
              class="rounded border px-3 py-2 text-left hover:bg-slate-50/80 dark:hover:bg-slate-900/40"
              @click="openLogoBrowseChild(item)"
            >
              <div class="inline-flex h-8 w-8 items-center justify-center rounded border border-brand-200 bg-brand-50 text-brand-600 dark:border-brand-700/60 dark:bg-brand-900/30 dark:text-brand-300">
                <Folder class="w-4 h-4" />
              </div>
              <div class="mt-1 text-sm truncate font-medium">{{ item.name }}</div>
            </button>
          </div>
          <div :class="logoBrowseThumbGridClass">
            <div
              v-for="file in logoBrowseThumbFiles"
              :key="`thumb-file-${file.object_key || file.relative_path}`"
              class="rounded border p-2 space-y-2"
            >
              <div :class="logoBrowseThumbPreviewClass">
                <img
                  v-if="file.can_preview"
                  :src="file.preview_url"
                  class="w-full h-full object-cover"
                  :alt="file.name"
                >
                <div v-else class="text-xs text-muted-foreground">无预览</div>
              </div>
              <div class="text-xs">
                <div class="truncate font-medium" :title="file.name">{{ file.name }}</div>
                <div class="text-muted-foreground truncate" :title="file.label">{{ file.label }}</div>
              </div>
              <div class="flex items-center justify-end">
                <Button type="button" size="sm" class="h-7 px-2" @click="applyLogoFromBrowse(file)">使用</Button>
              </div>
            </div>
          </div>
          <div v-if="!logoBrowseLoading && filteredLogoBrowseDirectories.length === 0 && logoBrowseThumbFiles.length === 0" class="px-3 py-6 text-sm text-muted-foreground text-center">
            当前路径下暂无内容
          </div>
        </div>
      </div>
      <DialogFooter class="flex-shrink-0 border-t border-border/60 pt-3">
        <Button type="button" variant="outline" @click="logoBrowseDialogOpen = false">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- Logo裁剪弹窗 -->
  <Dialog :open="logoCropDialogOpen" @update:open="(value) => { if (!value) closeLogoCropDialog() }">
    <DialogContent class="max-w-[560px]">
      <DialogHeader>
        <DialogTitle>裁剪站点图标</DialogTitle>
      </DialogHeader>
      <div class="space-y-3">
        <div class="text-xs text-muted-foreground">{{ logoCropImageName }}</div>
        <div
          class="w-[260px] h-[260px] border rounded-md overflow-hidden relative bg-slate-100 dark:bg-slate-900 mx-auto touch-none select-none"
          @pointerdown="onLogoCropPointerDown"
          @pointermove="onLogoCropPointerMove"
          @pointerup="stopLogoCropDragging"
          @pointerleave="stopLogoCropDragging"
          @wheel="onLogoCropWheel"
        >
          <img
            v-if="logoCropImageUrl"
            :src="logoCropImageUrl"
            alt="crop-source"
            class="absolute top-0 left-0 max-w-none"
            :style="{
              width: `${logoCropImageElement ? logoCropImageElement.width * logoCropScale : 0}px`,
              height: `${logoCropImageElement ? logoCropImageElement.height * logoCropScale : 0}px`,
              transform: `translate(${logoCropOffsetX}px, ${logoCropOffsetY}px)`
            }"
          >
          <div class="absolute inset-0 pointer-events-none border border-white/60 dark:border-slate-200/50"></div>
          <div class="absolute top-0 bottom-0 left-1/3 w-px bg-white/45 dark:bg-slate-200/35 pointer-events-none"></div>
          <div class="absolute top-0 bottom-0 left-2/3 w-px bg-white/45 dark:bg-slate-200/35 pointer-events-none"></div>
          <div class="absolute left-0 right-0 top-1/3 h-px bg-white/45 dark:bg-slate-200/35 pointer-events-none"></div>
          <div class="absolute left-0 right-0 top-2/3 h-px bg-white/45 dark:bg-slate-200/35 pointer-events-none"></div>
        </div>
        <div class="space-y-2">
          <div class="flex items-center justify-between text-xs text-muted-foreground">
            <span>缩放（支持鼠标滚轮）</span>
            <span>{{ Math.round(logoCropScale * 100) }}%</span>
          </div>
          <div class="flex items-center gap-2">
            <Button type="button" variant="outline" size="sm" @click="zoomOutLogoCrop">-</Button>
            <Input
              type="range"
              :min="logoCropMinScale"
              :max="logoCropMaxScale"
              :step="0.01"
              :model-value="logoCropScale"
              @update:model-value="(value) => applyLogoCropScale(Number(value))"
            />
            <Button type="button" variant="outline" size="sm" @click="zoomInLogoCrop">+</Button>
          </div>
        </div>
      </div>
      <DialogFooter>
        <Button type="button" variant="outline" :disabled="logoUploading" @click="closeLogoCropDialog">取消</Button>
        <Button type="button" :disabled="logoUploading" @click="confirmLogoCropAndUpload">{{ logoUploading ? '上传中...' : '裁剪并上传' }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
