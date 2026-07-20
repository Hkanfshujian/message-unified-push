<script setup lang="ts">
import { reactive, onMounted, ref, computed, watch } from 'vue'
import { settingsApi } from '@/api/settings'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import {
  QuestionCircleOutlined,
  CheckCircleOutlined,
  CheckCircleFilled,
  RightOutlined,
  FolderOutlined
} from '@ant-design/icons-vue'
// @ts-ignore
import config from '../../../../config.js'

const state = reactive({
  title: '',
  slogan: '',
  login_title: '',
  logo: '',
  logo_storage_profile_id: '',
  pagesize: '',
  cookieExpDays: '',
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
  const rsp = await settingsApi.getStorageConfig()
  const data = rsp?.data?.data || {}
  const list = Array.isArray(data.profiles) ? data.profiles : []
  defaultStorageProfileID.value = (data.default_storage_id || '').trim()
  logoStorageProfiles.value = list
    .map((item: any) => ({
      id: String(item?.id || '').trim(),
      name: String(item?.name || '').trim(),
      provider: (String(item?.provider || '').trim().toLowerCase() === 's3' ? 's3' : 'local') as 's3' | 'local',
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
  logoBrowseDialogOpen.value = false
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
  const rsp = await settingsApi.uploadSiteLogo(formData)
  if (rsp?.data?.code !== 200) {
    notifyError(rsp?.data?.msg || '上传失败')
    return
  }
  const data = rsp?.data?.data || {}
  state.logo = normalizeUploadedLogoUrl(data.url || '')
  state.logo_storage_profile_id = data.storage_profile_id || state.logo_storage_profile_id
  notifySuccess('站点图标上传成功')
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
    notifyError(error?.message || '图标上传失败')
  } finally {
    logoUploading.value = false
  }
}

const onSelectLogoFile = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  if (file.size > 2 * 1024 * 1024) {
    notifyError('图片不能超过 2MB')
    target.value = ''
    return
  }
  try {
    await openLogoCropDialog(file)
  } catch (error: any) {
    notifyError(error?.message || '图片读取失败')
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
    const rsp = await settingsApi.clearSiteLogo(clearLogoDeleteSource.value)
    if (rsp?.data?.code !== 200) {
      notifyError(rsp?.data?.msg || '恢复默认失败')
      return
    }
    await getSiteConfig()
    notifySuccess(clearLogoDeleteSource.value ? '已恢复默认图标，并删除源文件' : '已恢复默认图标')
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
      const rsp = await settingsApi.listLocalFiles(profile.id, path)
      const data = rsp?.data?.data || {}
      logoBrowseCurrentPath.value = data.current_path || ''
      logoBrowseParentPath.value = data.parent_path || ''
      logoBrowsePrefix.value = data.prefix || ''
      logoBrowseRootPath.value = ''
      logoBrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
      logoBrowseFiles.value = Array.isArray(data.files) ? data.files : []
      return
    }
    const rsp = await settingsApi.listLocalFiles(profile.id, path)
    const data = rsp?.data?.data || {}
    logoBrowseCurrentPath.value = data.current_path || ''
    logoBrowseParentPath.value = data.parent_path || ''
    logoBrowseRootPath.value = data.root_path || ''
    logoBrowsePrefix.value = ''
    logoBrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
    logoBrowseFiles.value = Array.isArray(data.files) ? data.files : []
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || '读取存储文件失败')
  } finally {
    logoBrowseLoading.value = false
  }
}

const openLogoBrowseDialog = async () => {
  if (!state.logo_storage_profile_id) {
    notifyError('请先选择图标存储')
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

const logoBrowseThumbGridClass = computed(() => `app-file-browser-grid app-file-browser-grid--${logoBrowseThumbSize.value}`)

const logoBrowseThumbPreviewClass = computed(() => `site-logo-preview-tile app-file-browser-card-preview app-file-browser-card-preview--${logoBrowseThumbSize.value}`)

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
    notifyError('当前 S3 文件缺少 public_url，无法作为站点图标')
    return
  }
  if (!rawUrl) {
    notifyError('文件地址为空，无法选择')
    return
  }
  state.logo = rawUrl
  logoBrowseDialogOpen.value = false
  notifySuccess(`已选择图标：${file.name}`)
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
        slogan_initial_enabled: state.sloganInitialEnabled ? 'true' : 'false',
        channel_test_message: state.channel_test_message.trim(),
      },
    }
    const response = await settingsApi.set(postData.section, postData.data)
    if (response.data.code === 200) {
      const msg = response.data.msg
      notifySuccess(msg)
    }
  } catch (error) {
    notifyError('保存失败，请稍后重试')
  }
}

// 恢复默认设置
const handleSubmitReset = async () => {
  try {
    const response = await settingsApi.reset()
    if (response.data.code === 200) {
      const msg = response.data.msg
      notifySuccess(msg)
      // 重新获取设置
      await getSiteConfig()
    }
  } catch (error) {
    notifyError('恢复默认设置失败，请稍后重试')
  }
}

// 获取站点配置
const getSiteConfig = async () => {
  try {
    const params = { params: { section: 'site_config' } }
    const response = await settingsApi.get(params.params.section)
    if (response.data.code === 200) {
      const data = response.data.data
      state.title = data.title || ''
      state.logo = data.logo || ''
      state.logo_storage_profile_id = data.logo_storage_profile_id || ''
      state.slogan = data.slogan || ''
      state.login_title = data.login_title || ''
      state.pagesize = data.pagesize || ''
      state.cookieExpDays = data.cookie_exp_days || '1'
      state.sloganInitialEnabled = String(data.slogan_initial_enabled || 'false') === 'true'
      state.channel_test_message = data.channel_test_message || 'This is a test message from ops-message-unified-push.'

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
    notifyError('获取配置失败')
  }
}

onMounted(async () => {
  await loadLogoStorageProfiles()
  await getSiteConfig()
})
</script>

<script lang="ts">
export default {
  name: 'SiteSettings'
}
</script>

<template>
  <div class="site-settings-form space-y-5">
    <el-card shadow="never" class="settings-section-card">
      <div class="settings-card-heading">
        <div class="settings-card-title">基本设置</div>
        <div class="settings-card-description">统一站点展示名称、登录页文案和侧边栏品牌识别。</div>
      </div>
      <div class="space-y-4">
        <div class="app-form-field">
          <label class="app-form-label">站点标题</label>
          <el-input v-model="state.title" placeholder="请输入自定义的网站标题" />
        </div>
        <div class="app-form-field">
          <label class="app-form-label">站点标语</label>
          <el-input v-model="state.slogan" placeholder="请输入自定义的网站slogan" />
          <div class="site-settings-inline-option flex items-center justify-between gap-3 px-3 py-2">
            <div>
              <div class="settings-option-title">侧边栏折叠标识跟随标语首字母</div>
              <div class="settings-option-description">开启后优先取标语首字母；关闭后固定使用默认值 M。</div>
            </div>
            <el-switch
              :model-value="state.sloganInitialEnabled"
              @update:model-value="(val: boolean | string | number) => state.sloganInitialEnabled = val === true"
            />
          </div>
        </div>
        <div class="app-form-field">
          <label class="app-form-label">登录页标题</label>
          <el-input v-model="state.login_title" placeholder="登录页显示的标题，默认：消 息 统 一 推 送 中 台" />
        </div>
        <div class="app-form-field">
          <label class="app-form-label">站点图标</label>
          <div v-if="logoStorageProfiles.length > 0" class="space-y-1">
            <div class="flex items-center justify-between gap-2">
              <div class="settings-option-description">图标存储</div>
              <el-button
                v-if="isInlineSvgLogo"
                size="sm"
                text
                class="h-7 px-2 text-xs text-muted-foreground"
                @click="showLogoStorageAdvanced = !showLogoStorageAdvanced"
              >
                {{ showLogoStorageAdvanced ? '收起高级选项' : '高级选项' }}
              </el-button>
            </div>
            <div v-if="logoStoragePanelExpanded" class="space-y-1">
              <el-select
                v-model="state.logo_storage_profile_id"
                class="w-full"
              >
                <el-option
                  v-for="profile in logoStorageProfiles"
                  :key="profile.id"
                  :value="profile.id"
                  :label="`${profile.name}（${profile.provider === 's3' ? 'S3' : '本地'} / ${profile.id}）`"
                />
              </el-select>
              <div v-if="isInlineSvgLogo" class="settings-warning-text">
                当前为默认/SVG图标，显示效果不依赖存储类型；存储选择仅在“上传并裁剪”或“浏览”时生效。
              </div>
            </div>
            <div v-else class="settings-option-description">
              当前为默认/SVG图标，显示效果不依赖存储类型。需要上传或浏览存储文件时可展开“高级选项”。
            </div>
          </div>
          <input ref="logoInputRef" type="file" accept=".png,.jpg,.jpeg,.webp" class="hidden" @change="onSelectLogoFile">
          <div class="site-logo-action-strip">
            <el-button
              native-type="button"
              size="sm"
              type="primary"
              class="site-logo-action-button"
              :disabled="logoUploading || logoClearing || !state.logo_storage_profile_id"
              @click="logoInputRef?.click()"
            >
              {{ logoUploading ? '上传中...' : '上传并裁剪' }}
            </el-button>
            <el-button
              native-type="button"
              size="sm"
              class="site-logo-action-button"
              :disabled="logoUploading || logoClearing || !state.logo_storage_profile_id"
              @click="openLogoBrowseDialog"
            >
              浏览
            </el-button>
            <el-button
              native-type="button"
              size="sm"
              class="site-logo-action-button site-logo-action-button-danger"
              :disabled="logoUploading || logoClearing || !state.logo"
              @click="openClearLogoConfirm"
            >
              {{ logoClearing ? '恢复中...' : '恢复默认图标' }}
            </el-button>
            <button
              type="button"
              class="site-logo-delete-toggle"
              :class="clearLogoDeleteSource ? 'site-logo-delete-toggle-active' : ''"
              :title="clearLogoDeleteSource ? '恢复默认图标时同步删除存储中的源文件' : '恢复默认图标时仅清理配置，不删除存储源文件'"
              @click="clearLogoDeleteSource = !clearLogoDeleteSource"
            >
              <CheckCircleFilled v-if="clearLogoDeleteSource" class="text-[14px]" />
              <CheckCircleOutlined v-else class="text-[14px]" />
              <span>恢复时同步删除源文件</span>
            </button>
          </div>
          <div class="app-form-help">支持 jpg/png/webp，最大 2MB，上传后自动裁剪为方图。</div>
          <div v-if="state.logo" class="site-logo-current-preview">
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
              class="w-8 h-8 flex-shrink-0 rounded-xl object-cover"
            >
            <div class="text-xs text-muted-foreground break-all flex-1 line-clamp-2">
              {{ state.logo.trimStart().startsWith('<') ? '（SVG 文本，不依赖存储类型；建议上传图片替换）' : state.logo }}
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="settings-section-card">
      <div class="settings-card-heading">
        <div class="settings-card-title">系统参数</div>
        <div class="settings-card-description">配置分页和登录会话有效期。</div>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="app-form-field">
          <label class="app-form-label">分页大小</label>
          <el-input v-model="state.pagesize" placeholder="页面分页大小" />
        </div>
        <div class="app-form-field">
          <label class="app-form-label">Cookie过期天数</label>
          <el-input v-model="state.cookieExpDays" type="number" min="1" max="365" placeholder="Cookie过期天数（默认1天）" />
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="settings-section-card">
      <div class="settings-card-heading">
        <div class="settings-card-title">渠道测试默认文案</div>
      </div>
      <div class="app-form-field">
        <el-input
          v-model="state.channel_test_message"
          type="textarea"
          maxlength="2000"
          :rows="4"
          placeholder="请输入渠道测试按钮默认发送的消息正文"
        />
        <div class="settings-field-meta">
          <span>{{ state.channel_test_message.length }}/2000</span>
        </div>
      </div>
    </el-card>

    <div class="settings-form-actions">
      <div class="flex items-center space-x-2">
        <span class="settings-option-description">说明</span>
        <el-tooltip placement="top" popper-class="max-w-sm">
          <QuestionCircleOutlined class="text-[14px] text-muted-foreground hover:text-foreground transition-colors duration-[var(--motion-fast)]" />
          <template #content>
            <div class="text-sm space-y-1">
              <p>1. logo请输入svg文本，替换后登录页面，ico，导航栏logo将全部一起更换</p>
              <p>2. slogan将在登录页面展示</p>
              <p>3. Cookie过期天数设置用户登录后的有效期，修改后下次登录时生效</p>
              <p>4. 修改后在下次登录时生效，如不生效请在登录页面Ctrl+F5强制刷新</p>
            </div>
          </template>
        </el-tooltip>
      </div>
      <div class="flex space-x-2">
        <el-button size="small" @click="handleSubmitReset">恢复默认</el-button>
        <el-button size="small" type="primary" @click="handleSubmit">确定</el-button>
      </div>
    </div>
  </div>

  <!-- 清空Logo确认弹窗 -->
  <el-dialog v-model="clearLogoConfirmOpen" title="确认恢复默认站点图标" width="min(420px, calc(100dvw - 24px))" class="app-nested-dialog" append-to-body>
      <div class="space-y-2 text-sm text-foreground/80">
        <div>恢复后站点将使用默认图标。</div>
        <div v-if="clearLogoDeleteSource" class="text-red-500">同时会删除存储中的源文件，此操作不可恢复。</div>
      </div>
      <template #footer>
        <el-button :disabled="logoClearing" @click="clearLogoConfirmOpen = false">取消</el-button>
        <el-button type="primary" :disabled="logoClearing" :loading="logoClearing" @click="clearSiteLogo">确认恢复</el-button>
      </template>
  </el-dialog>

  <!-- 从存储浏览选择站点图标 -->
  <el-dialog v-model="logoBrowseDialogOpen" title="从存储选择站点图标" width="min(880px, calc(100dvw - 24px))" class="site-logo-browse-dialog app-file-browser-dialog app-nested-dialog" append-to-body>
      <div class="app-file-browser">
        <section class="app-file-browser-metadata" aria-label="当前存储信息">
          <span class="app-file-browser-metadata__icon" aria-hidden="true">
            <CloudServerOutlined />
          </span>
          <div class="app-file-browser-metadata__copy" :title="selectedLogoStorageProfile?.name || ''">
            <strong>{{ selectedLogoStorageProfile?.name || '-' }}</strong>
            <small>{{ selectedLogoStorageProfile?.provider === 's3' ? 'S3' : '本地' }} · {{ state.logo_storage_profile_id || '-' }}</small>
          </div>
        </section>
        <div class="app-file-browser-breadcrumb">
          <span class="app-file-browser-breadcrumb__label">当前目录：</span>
          <div class="app-file-browser-breadcrumb__track">
          <button
            v-for="(crumb, index) in logoBrowseBreadcrumbs"
            :key="`${crumb.path || 'logo-root'}-${index}`"
            type="button"
            class="storage-breadcrumb-button inline-flex items-center gap-1 mr-1"
            :disabled="logoBrowseLoading"
            @click="openLogoBrowseBreadcrumb(crumb.path)"
          >
            <RightOutlined v-if="index > 0" class="text-[13px]" />
            <FolderOutlined v-if="index === 0" class="text-[13px]" />
            <span class="max-w-[220px] truncate align-bottom">{{ crumb.label }}</span>
          </button>
          </div>
        </div>
        <div class="app-file-browser-toolbar">
          <el-button class="app-file-browser-toolbar__back" native-type="button" size="small" :disabled="!logoBrowseCurrentPath || logoBrowseLoading" @click="openLogoBrowseParent">返回上级</el-button>
          <el-input v-model="logoBrowseKeyword" class="app-file-browser-toolbar__search" placeholder="按名称筛选目录/图片" />
          <div class="app-segmented app-file-browser-toolbar__view" role="group" aria-label="浏览视图">
            <button type="button" class="storage-segment-button" :class="logoBrowseViewMode === 'list' ? 'storage-segment-button-active' : 'storage-segment-button-idle'" :aria-pressed="logoBrowseViewMode === 'list'" @click="logoBrowseViewMode = 'list'">列表</button>
            <button type="button" class="storage-segment-button" :class="logoBrowseViewMode === 'thumb' ? 'storage-segment-button-active' : 'storage-segment-button-idle'" :aria-pressed="logoBrowseViewMode === 'thumb'" @click="logoBrowseViewMode = 'thumb'">缩略图</button>
          </div>
          <div v-if="logoBrowseViewMode === 'thumb'" class="app-segmented app-file-browser-toolbar__size" role="group" aria-label="缩略图尺寸">
            <button v-for="size in (['sm', 'md', 'lg'] as const)" :key="size" type="button" class="storage-segment-button" :class="logoBrowseThumbSize === size ? 'storage-segment-button-active' : 'storage-segment-button-idle'" :aria-pressed="logoBrowseThumbSize === size" @click="logoBrowseThumbSize = size">{{ size === 'sm' ? '小' : size === 'md' ? '中' : '大' }}</button>
          </div>
        </div>
        <div class="storage-browser-list app-file-browser-content">
          <div v-if="logoBrowseLoading" class="app-file-browser-state" role="status">正在读取当前目录…</div>
          <template v-else-if="logoBrowseViewMode === 'list'">
            <button v-for="item in filteredLogoBrowseDirectories" :key="item.relative_path" type="button" class="storage-browser-row app-file-browser-row app-file-browser-row--directory" @click="openLogoBrowseChild(item)">
              <span class="app-file-browser-row-main"><span class="app-file-browser-folder-icon"><FolderOutlined /></span><span class="app-file-browser-row-copy"><strong>{{ item.name }}</strong><small>/{{ item.relative_path }}</small></span></span>
            </button>
            <div v-for="file in filteredLogoBrowseFiles" :key="file.object_key || file.relative_path" class="storage-browser-row app-file-browser-row">
              <div class="app-file-browser-row-main"><span class="app-file-browser-row-copy"><strong>{{ file.name }}</strong><small>{{ selectedLogoStorageProfile?.provider === 's3' ? (file.object_key || file.relative_path) : file.relative_path }}</small></span></div>
              <div class="app-file-browser-row-actions"><el-button native-type="button" size="small" type="primary" @click="applyLogoFromBrowse(file)">使用</el-button></div>
            </div>
          </template>
          <template v-else>
          <div v-if="filteredLogoBrowseDirectories.length > 0" class="app-file-browser-grid app-file-browser-grid--directories">
            <button v-for="item in filteredLogoBrowseDirectories" :key="`thumb-dir-${item.relative_path}`" type="button" class="storage-thumb-dir-card app-file-browser-directory-card" @click="openLogoBrowseChild(item)">
              <span class="app-file-browser-folder-icon"><FolderOutlined /></span><strong>{{ item.name }}</strong>
            </button>
          </div>
          <div :class="logoBrowseThumbGridClass">
            <div
              v-for="file in logoBrowseThumbFiles"
              :key="`thumb-file-${file.object_key || file.relative_path}`"
              class="storage-thumb-file-card app-file-browser-card"
            >
              <div :class="logoBrowseThumbPreviewClass">
                <img v-if="file.can_preview" :src="file.preview_url" :alt="file.name">
                <div v-else>无预览</div>
              </div>
              <div class="app-file-browser-card-copy"><strong :title="file.name">{{ file.name }}</strong><small :title="file.label">{{ file.label }}</small></div>
              <div class="app-file-browser-card-actions">
                <el-button native-type="button" size="small" type="primary" @click="applyLogoFromBrowse(file)">使用</el-button>
              </div>
            </div>
          </div>
          </template>
          <div v-if="!logoBrowseLoading && filteredLogoBrowseDirectories.length === 0 && filteredLogoBrowseFiles.length === 0" class="app-file-browser-state app-file-browser-state--empty">
            <FolderOutlined /><strong>当前路径下没有可用图片</strong><span>可返回上级目录或调整筛选条件</span>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="logoBrowseDialogOpen = false">关闭</el-button>
      </template>
  </el-dialog>

  <!-- Logo裁剪弹窗 -->
  <el-dialog :model-value="logoCropDialogOpen" title="裁剪站点图标" width="min(560px, calc(100dvw - 24px))" class="site-logo-crop-dialog app-nested-dialog" append-to-body @update:model-value="(value: boolean) => { if (!value) closeLogoCropDialog() }">
      <div class="site-logo-crop-content">
        <div class="site-logo-crop-toolbar">
          <strong :title="logoCropImageName || '待裁剪图片'">{{ logoCropImageName || '待裁剪图片' }}</strong>
          <span>输出 128 × 128 PNG</span>
          <p>拖动调整取景，使用滚轮或下方控件缩放；方框内内容将作为站点图标。</p>
        </div>
        <div
          class="site-logo-crop-frame w-[260px] h-[260px] overflow-hidden relative mx-auto touch-none select-none"
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
          <div class="absolute inset-0 pointer-events-none border border-white/60 dark:border-muted-foreground/40"></div>
          <div class="absolute top-0 bottom-0 left-1/3 w-px bg-white/45 dark:bg-muted-foreground/30 pointer-events-none"></div>
          <div class="absolute top-0 bottom-0 left-2/3 w-px bg-white/45 dark:bg-muted-foreground/30 pointer-events-none"></div>
          <div class="absolute left-0 right-0 top-1/3 h-px bg-white/45 dark:bg-muted-foreground/30 pointer-events-none"></div>
          <div class="absolute left-0 right-0 top-2/3 h-px bg-white/45 dark:bg-muted-foreground/30 pointer-events-none"></div>
        </div>
        <div class="space-y-2">
          <div class="flex items-center justify-between text-xs text-muted-foreground">
            <span>缩放（支持鼠标滚轮）</span>
            <span>{{ Math.round(logoCropScale * 100) }}%</span>
          </div>
          <div class="flex items-center gap-2">
            <el-button native-type="button" size="small" @click="zoomOutLogoCrop">-</el-button>
            <el-input
              type="range"
              :min="logoCropMinScale"
              :max="logoCropMaxScale"
              :step="0.01"
              :model-value="logoCropScale"
              @update:model-value="(value: string | number) => applyLogoCropScale(Number(value))"
            />
            <el-button native-type="button" size="small" @click="zoomInLogoCrop">+</el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button :disabled="logoUploading" @click="closeLogoCropDialog">取消</el-button>
        <el-button type="primary" :disabled="logoUploading" :loading="logoUploading" @click="confirmLogoCropAndUpload">裁剪并上传</el-button>
      </template>
  </el-dialog>
</template>

<style scoped>
.site-logo-crop-content { display: grid; gap: 12px; }
.site-logo-crop-toolbar { display: flex; align-items: baseline; flex-wrap: wrap; gap: 4px 10px; padding-bottom: 10px; border-bottom: 1px solid var(--app-overlay-border); }
.site-logo-crop-toolbar strong { max-width: 240px; overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.site-logo-crop-toolbar span { color: var(--admin-text-muted); font-size: 10px; }
.site-logo-crop-toolbar p { width: 100%; margin: 0; color: var(--admin-text-muted); font-size: 11px; line-height: 1.55; }
</style>
