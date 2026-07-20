<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { request } from '@/api/api'
import { settingsApi } from '@/api/settings'
import { notifyError, notifySuccess } from '@/util/uiFeedback'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import AppFormDrawer from '@/components/ui/AppFormDrawer.vue'
import AppRowActions from '@/components/ui/AppRowActions.vue'
import {
  CloudServerOutlined,
  CopyOutlined,
  FileOutlined,
  FolderOutlined,
  FolderOpenOutlined,
  LinkOutlined,
  RightOutlined,
  StarFilled,
  StarOutlined,
  UploadOutlined,
  ArrowLeftOutlined
} from '@ant-design/icons-vue'

const loading = ref(false)
const saving = ref(false)
const editorOpen = ref(false)
const editingProfileId = ref('')
const defaultStorageID = ref('')
const deleteConfirmOpen = ref(false)
const deleteConfirmInput = ref('')
const deleteTarget = ref<StorageProfile | null>(null)
const localDirDialogOpen = ref(false)
const localDirLoading = ref(false)
const localDirCreating = ref(false)
const localDirCurrentPath = ref('')
const localDirParentPath = ref('')
const localDirItems = ref<LocalDirItem[]>([])
const localDirNewFolderName = ref('')
const localUploadTestOpen = ref(false)
const localUploadTestProfile = ref<StorageProfile | null>(null)
const localUploadTestFile = ref<File | null>(null)
const localUploadTesting = ref(false)
const localUploadFileInputRef = ref<HTMLInputElement | null>(null)
const localUploadDeleteAfter = ref(false)
const localUploadLastLocation = ref('')
const localUploadAutoCloseCountdown = ref(0)
const localUploadAutoCloseTimer = ref<number | null>(null)
const s3BrowseDialogOpen = ref(false)
const s3BrowseLoading = ref(false)
const s3BrowseProfile = ref<StorageProfile | null>(null)
const s3BrowseCurrentPath = ref('')
const s3BrowseParentPath = ref('')
const s3BrowsePrefix = ref('')
const s3BrowseDirectories = ref<S3DirItem[]>([])
const s3BrowseFiles = ref<S3FileItem[]>([])
const localBrowseDialogOpen = ref(false)
const localBrowseLoading = ref(false)
const localBrowseProfile = ref<StorageProfile | null>(null)
const localBrowseCurrentPath = ref('')
const localBrowseParentPath = ref('')
const localBrowseRootPath = ref('')
const localBrowseDirectories = ref<LocalBrowseDirItem[]>([])
const localBrowseFiles = ref<LocalBrowseFileItem[]>([])
const browseKeyword = ref('')
const browseViewMode = ref<'list' | 'thumb'>('list')
const browseThumbSize = ref<'sm' | 'md' | 'lg'>('md')
const filePreviewDialogOpen = ref(false)
const filePreviewUrl = ref('')
const filePreviewName = ref('')
const filePreviewPath = ref('')
const filePreviewIsImage = ref(false)
const filePreviewKey = ref('')
const fileDeleteConfirmOpen = ref(false)
const fileDeleting = ref(false)
const fileDeleteTargetName = ref('')
const fileDeleteTargetPath = ref('')
const fileDeleteTargetObjectKey = ref('')

type StorageProfile = {
  id: string
  name: string
  provider: 'local' | 's3'
  enabled: boolean
  upload_file_prefix: string
  local_sub_path: string
  s3_endpoint: string
  s3_region: string
  s3_bucket: string
  s3_access_key: string
  s3_secret_key: string
  s3_use_ssl: boolean
  s3_public_base_url: string
  s3_proxy_public_read: boolean
  s3_object_key_prefix: string
}

type LocalDirItem = {
  name: string
  relative_path: string
}

type LocalDirBreadcrumb = {
  label: string
  path: string
}

type S3DirItem = {
  name: string
  relative_path: string
}

type S3FileItem = {
  name: string
  relative_path: string
  object_key: string
  size: number
  last_modified: string
  public_url: string
}

type S3Breadcrumb = {
  label: string
  path: string
}

type LocalBrowseDirItem = {
  name: string
  relative_path: string
}

type LocalBrowseFileItem = {
  name: string
  relative_path: string
  size: number
  last_modified: string
  public_url: string
}

type LocalBrowseBreadcrumb = {
  label: string
  path: string
}

type BrowseDirItem = {
  name: string
  relative_path: string
}

type BrowseFileItem = {
  name: string
  relative_path: string
  object_key?: string
  size?: number
  public_url?: string
}

const profiles = ref<StorageProfile[]>([])

const isEightDigitStorageId = (id: string) => /^\d{8}$/.test((id || '').trim())

const normalizeProfilesFromApi = (input: StorageProfile[], defaultId: string) => {
  const exists = new Set<string>()
  const normalized = input.map((item, index) => {
    let nextId = (item.id || '').trim()
    if (!isEightDigitStorageId(nextId) || exists.has(nextId)) {
      let seed = Date.now() + index * 37
      nextId = String(Math.floor(seed % 90000000) + 10000000)
      while (exists.has(nextId)) {
        seed += 97
        nextId = String(Math.floor(seed % 90000000) + 10000000)
      }
    }
    exists.add(nextId)
    return {
      ...item,
      id: nextId,
      upload_file_prefix: (item.upload_file_prefix || 'upload').trim() || 'upload',
      local_sub_path: (item.local_sub_path || 'uploads').trim() || 'uploads'
    }
  })
  let nextDefaultId = (defaultId || '').trim()
  if (!isEightDigitStorageId(nextDefaultId) || !exists.has(nextDefaultId)) {
    nextDefaultId = normalized[0]?.id || ''
  }
  return { normalized, nextDefaultId }
}

const editor = reactive<StorageProfile>({
  id: '',
  name: '',
  provider: 'local',
  enabled: true,
  upload_file_prefix: 'upload',
  local_sub_path: 'uploads',
  s3_endpoint: '',
  s3_region: '',
  s3_bucket: '',
  s3_access_key: '',
  s3_secret_key: '',
  s3_use_ssl: true,
  s3_public_base_url: '',
  s3_proxy_public_read: true,
  s3_object_key_prefix: ''
})

const resetEditor = () => {
  editor.id = ''
  editor.name = ''
  editor.provider = 'local'
  editor.enabled = true
  editor.upload_file_prefix = 'upload'
  editor.local_sub_path = 'uploads'
  editor.s3_endpoint = ''
  editor.s3_region = ''
  editor.s3_bucket = ''
  editor.s3_access_key = ''
  editor.s3_secret_key = ''
  editor.s3_use_ssl = true
  editor.s3_public_base_url = ''
  editor.s3_proxy_public_read = true
  editor.s3_object_key_prefix = ''
}

const loadConfig = async () => {
  loading.value = true
  try {
    const rsp = await settingsApi.getStorageConfig()
    const data = rsp?.data?.data || {}
    const list = Array.isArray(data.profiles) ? data.profiles : []
    const { normalized, nextDefaultId } = normalizeProfilesFromApi(list, data.default_storage_id || '')
    profiles.value = normalized
    defaultStorageID.value = nextDefaultId
  } finally {
    loading.value = false
  }
}

const persistConfig = async (nextProfiles: StorageProfile[], nextDefaultStorageID: string, successText: string) => {
  if (nextProfiles.some(item => !isEightDigitStorageId(item.id))) {
    notifyError('存储ID必须为8位数字')
    return false
  }
  if (!isEightDigitStorageId(nextDefaultStorageID)) {
    notifyError('默认存储ID必须为8位数字')
    return false
  }
  saving.value = true
  try {
    const rsp = await settingsApi.saveStorageConfig({
      default_storage_id: nextDefaultStorageID,
      profiles: nextProfiles
    })
    if (rsp?.data?.code === 200) {
      profiles.value = nextProfiles
      defaultStorageID.value = nextDefaultStorageID
      notifySuccess(successText)
      return true
    }
    notifyError(rsp?.data?.msg || '存储配置保存失败')
    return false
  } finally {
    saving.value = false
  }
}

const openLocalUploadTest = (profile: StorageProfile) => {
  if (localUploadAutoCloseTimer.value !== null) {
    window.clearInterval(localUploadAutoCloseTimer.value)
    localUploadAutoCloseTimer.value = null
  }
  localUploadAutoCloseCountdown.value = 0
  localUploadTestProfile.value = profile
  localUploadTestFile.value = null
  localUploadDeleteAfter.value = false
  localUploadLastLocation.value = ''
  localUploadTestOpen.value = true
}

const closeLocalUploadTest = () => {
  if (localUploadAutoCloseTimer.value !== null) {
    window.clearInterval(localUploadAutoCloseTimer.value)
    localUploadAutoCloseTimer.value = null
  }
  localUploadAutoCloseCountdown.value = 0
  localUploadTestOpen.value = false
  localUploadTestProfile.value = null
  localUploadTestFile.value = null
  localUploadDeleteAfter.value = false
  localUploadLastLocation.value = ''
  if (localUploadFileInputRef.value) {
    localUploadFileInputRef.value.value = ''
  }
}

const handleLocalUploadFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  localUploadTestFile.value = target?.files?.[0] || null
}

const localUploadSelectedFileName = computed(() => localUploadTestFile.value?.name || '')

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(2)} MB`
}

const localUploadSelectedFileSize = computed(() => {
  if (!localUploadTestFile.value) return ''
  return formatFileSize(localUploadTestFile.value.size)
})

const clearLocalUploadFile = () => {
  localUploadTestFile.value = null
  if (localUploadFileInputRef.value) {
    localUploadFileInputRef.value.value = ''
  }
}

const joinPublicUrlAndObjectKey = (baseUrl: string, objectKey: string) => {
  const base = (baseUrl || '').trim().replace(/\/+$/, '')
  const key = (objectKey || '').trim().replace(/^\/+/, '')
  if (!base || !key) return ''
  return `${base}/${key}`
}

const submitLocalUploadTest = async () => {
  if (!localUploadTestProfile.value) {
    notifyError('未找到本地存储配置')
    return
  }
  if (!localUploadTestFile.value) {
    notifyError('请选择要上传的测试文件')
    return
  }
  localUploadTesting.value = true
  try {
    const formData = new FormData()
    formData.append('profile_id', localUploadTestProfile.value.id)
    formData.append('file', localUploadTestFile.value)
    formData.append('delete_after_upload', localUploadDeleteAfter.value ? '1' : '0')
    let rsp = await settingsApi.uploadStorageFile(formData)
    if ((rsp as any)?.status === 404 && localUploadTestProfile.value.provider === 'local') {
      rsp = await settingsApi.testLocalUpload(formData)
    }
    if (rsp?.data?.code === 200) {
      const objectKey = rsp?.data?.data?.object_key
      const publicUrl = rsp?.data?.data?.public_url || joinPublicUrlAndObjectKey(localUploadTestProfile.value.s3_public_base_url, objectKey)
      const deleted = rsp?.data?.data?.deleted === 'true'
      const suffix = deleted ? '（已立即删除）' : ''
      const location = publicUrl || objectKey || ''
      localUploadLastLocation.value = location
      notifySuccess(location ? `${rsp?.data?.msg || '文件上传成功'}：${location}${suffix}` : `${rsp?.data?.msg || '文件上传成功'}${suffix}`)
      clearLocalUploadFile()
      if (localUploadAutoCloseTimer.value !== null) {
        window.clearInterval(localUploadAutoCloseTimer.value)
        localUploadAutoCloseTimer.value = null
      }
      localUploadAutoCloseCountdown.value = 5
      localUploadAutoCloseTimer.value = window.setInterval(() => {
        if (localUploadAutoCloseCountdown.value <= 1) {
          if (localUploadAutoCloseTimer.value !== null) {
            window.clearInterval(localUploadAutoCloseTimer.value)
            localUploadAutoCloseTimer.value = null
          }
          localUploadAutoCloseCountdown.value = 0
          closeLocalUploadTest()
          return
        }
        localUploadAutoCloseCountdown.value -= 1
      }, 1000)
      return
    }
    notifyError(rsp?.data?.msg || '文件上传失败')
  } catch (error: any) {
    if (error?.response?.status === 404) {
      notifyError('上传接口不存在，请重启后端服务后重试')
      return
    }
    notifyError(error?.response?.data?.msg || '文件上传失败')
  } finally {
    localUploadTesting.value = false
  }
}

const handleTestClick = async (profile: StorageProfile) => {
  openLocalUploadTest(profile)
}

const openS3BrowseDialog = async (profile: StorageProfile) => {
  if (profile.provider !== 's3') {
    notifyError('仅支持浏览 S3 存储')
    return
  }
  s3BrowseProfile.value = profile
  s3BrowseDialogOpen.value = true
  browseKeyword.value = ''
  browseViewMode.value = 'list'
  browseThumbSize.value = 'md'
  await loadS3Objects('')
}

const loadS3Objects = async (path: string) => {
  if (!s3BrowseProfile.value) return
  s3BrowseLoading.value = true
  try {
    const rsp = await settingsApi.listS3Objects(s3BrowseProfile.value.id, path)
    const data = rsp?.data?.data || {}
    s3BrowseCurrentPath.value = data.current_path || ''
    s3BrowseParentPath.value = data.parent_path || ''
    s3BrowsePrefix.value = data.prefix || ''
    s3BrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
    s3BrowseFiles.value = Array.isArray(data.files) ? data.files : []
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || '读取 S3 对象失败')
  } finally {
    s3BrowseLoading.value = false
  }
}

const openS3ChildDirectory = async (item: S3DirItem) => {
  await loadS3Objects(item.relative_path || '')
}

const openS3ParentDirectory = async () => {
  await loadS3Objects(s3BrowseParentPath.value || '')
}

const s3BrowseBreadcrumbs = computed<S3Breadcrumb[]>(() => {
  const rootLabel = s3BrowsePrefix.value ? `${s3BrowsePrefix.value}` : '根目录'
  const items: S3Breadcrumb[] = [{ label: rootLabel, path: '' }]
  const current = (s3BrowseCurrentPath.value || '').trim()
  if (!current) return items
  const parts = current.split('/').filter(Boolean)
  let acc = ''
  for (const part of parts) {
    acc = acc ? `${acc}/${part}` : part
    items.push({ label: part, path: acc })
  }
  return items
})

const openS3Breadcrumb = async (path: string) => {
  await loadS3Objects(path || '')
}

const previewS3File = (file: S3FileItem) => {
  const url = normalizeS3PreviewUrl(file.public_url || '', file.object_key || '')
  if (!url) {
    notifyError('当前存储未配置 Public URL，无法预览')
    return
  }
  applyPreviewPayload({
    key: file.object_key || file.relative_path || file.name || url,
    url,
    name: file.name || '文件预览',
    path: file.object_key || file.relative_path || '',
    isImage: isImageLikeFile(file.name || file.object_key || file.relative_path || '')
  })
}

const getApiOrigin = () => {
  const baseURL = String((request as any)?.defaults?.baseURL || '').trim()
  if (baseURL) {
    try {
      const url = new URL(baseURL, window.location.origin)
      return `${url.origin}${url.pathname.replace(/\/+$/, '')}`
    } catch (_error) {
      return baseURL.replace(/\/+$/, '')
    }
  }
  return window.location.origin
}

const normalizeS3PreviewUrl = (rawUrl: string, objectKey: string) => {
  const raw = (rawUrl || '').trim()
  if (raw && /^https?:\/\//i.test(raw)) return raw
  const profileBase = (s3BrowseProfile.value?.s3_public_base_url || '').trim().replace(/\/+$/, '')
  const key = (objectKey || '').trim().replace(/^\/+/, '')
  if (profileBase && key) return `${profileBase}/${key}`
  if (raw.startsWith('/public/')) return `${getApiOrigin()}${raw}`
  if (raw && profileBase) return `${profileBase}/${raw.replace(/^\/+/, '')}`
  if (raw.startsWith('/')) return `${getApiOrigin()}${raw}`
  return raw
}

const handleBrowseClick = async (profile: StorageProfile) => {
  if (profile.provider === 's3') {
    await openS3BrowseDialog(profile)
    return
  }
  await openLocalBrowseDialog(profile)
}

const openLocalBrowseDialog = async (profile: StorageProfile) => {
  if (profile.provider !== 'local') {
    notifyError('仅支持浏览本地存储')
    return
  }
  localBrowseProfile.value = profile
  localBrowseDialogOpen.value = true
  browseKeyword.value = ''
  browseViewMode.value = 'list'
  browseThumbSize.value = 'md'
  await loadLocalBrowseFiles('')
}

const loadLocalBrowseFiles = async (path: string) => {
  if (!localBrowseProfile.value) return
  localBrowseLoading.value = true
  try {
    const rsp = await settingsApi.listLocalFiles(localBrowseProfile.value.id, path)
    const data = rsp?.data?.data || {}
    localBrowseCurrentPath.value = data.current_path || ''
    localBrowseParentPath.value = data.parent_path || ''
    localBrowseRootPath.value = data.root_path || ''
    localBrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
    localBrowseFiles.value = Array.isArray(data.files) ? data.files : []
  } catch (error: any) {
    notifyError(error?.response?.data?.msg || '读取本地文件失败')
  } finally {
    localBrowseLoading.value = false
  }
}

const openLocalBrowseChild = async (item: LocalBrowseDirItem) => {
  await loadLocalBrowseFiles(item.relative_path || '')
}

const openLocalBrowseParent = async () => {
  await loadLocalBrowseFiles(localBrowseParentPath.value || '')
}

const localBrowseBreadcrumbs = computed<LocalBrowseBreadcrumb[]>(() => {
  const rootLabel = localBrowseRootPath.value ? localBrowseRootPath.value : 'uploads'
  const items: LocalBrowseBreadcrumb[] = [{ label: rootLabel, path: '' }]
  const current = (localBrowseCurrentPath.value || '').trim()
  if (!current) return items
  const parts = current.split('/').filter(Boolean)
  let acc = ''
  for (const part of parts) {
    acc = acc ? `${acc}/${part}` : part
    items.push({ label: part, path: acc })
  }
  return items
})

const openLocalBrowseBreadcrumb = async (path: string) => {
  await loadLocalBrowseFiles(path || '')
}

const resolveLocalPublicUrl = (url: string) => {
  const raw = (url || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw)) return raw
  const base = `${getApiOrigin()}`.replace(/\/+$/, '')
  const path = raw.startsWith('/') ? raw : `/${raw}`
  return `${base}${path}`
}

const previewLocalFile = (file: LocalBrowseFileItem) => {
  const url = resolveLocalPublicUrl(file.public_url)
  if (!url) {
    notifyError('文件地址为空，无法预览')
    return
  }
  applyPreviewPayload({
    key: file.relative_path || file.name || url,
    url,
    name: file.name || '文件预览',
    path: file.relative_path || '',
    isImage: isImageLikeFile(file.name || file.relative_path || '')
  })
}

const openPreviewInNewTab = () => {
  if (!filePreviewUrl.value) return
  window.open(filePreviewUrl.value, '_blank')
}

const handlePreviewKeydown = (event: KeyboardEvent) => {
  if (!filePreviewDialogOpen.value) return
  const target = event.target as HTMLElement | null
  const tag = target?.tagName?.toLowerCase() || ''
  if (tag === 'input' || tag === 'textarea') return
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    movePreviewBy(-1)
    return
  }
  if (event.key === 'ArrowRight') {
    event.preventDefault()
    movePreviewBy(1)
  }
}

const browseMode = computed<'s3' | 'local' | ''>(() => {
  if (s3BrowseDialogOpen.value) return 's3'
  if (localBrowseDialogOpen.value) return 'local'
  return ''
})

const browseDialogOpen = computed<boolean>({
  get: () => s3BrowseDialogOpen.value || localBrowseDialogOpen.value,
  set: (value) => {
    if (!value) {
      s3BrowseDialogOpen.value = false
      localBrowseDialogOpen.value = false
      filePreviewDialogOpen.value = false
    }
  }
})

const browseDialogTitle = computed(() => (browseMode.value === 's3' ? 'S3 文件浏览' : '本地文件浏览'))
const browseProfileName = computed(() => (browseMode.value === 's3' ? s3BrowseProfile.value?.name : localBrowseProfile.value?.name) || '')
const browseProfileId = computed(() => (browseMode.value === 's3' ? s3BrowseProfile.value?.id : localBrowseProfile.value?.id) || '')
const browseLoading = computed(() => (browseMode.value === 's3' ? s3BrowseLoading.value : localBrowseLoading.value))
const browseCurrentPath = computed(() => (browseMode.value === 's3' ? s3BrowseCurrentPath.value : localBrowseCurrentPath.value))
const browseDirectories = computed<BrowseDirItem[]>(() => (browseMode.value === 's3' ? s3BrowseDirectories.value : localBrowseDirectories.value))
const browseFiles = computed<BrowseFileItem[]>(() => (browseMode.value === 's3' ? s3BrowseFiles.value : localBrowseFiles.value))
const browseBreadcrumbs = computed<{ label: string, path: string }[]>(() => (browseMode.value === 's3' ? s3BrowseBreadcrumbs.value : localBrowseBreadcrumbs.value))
const normalizedBrowseKeyword = computed(() => (browseKeyword.value || '').trim().toLowerCase())
const filteredBrowseDirectories = computed<BrowseDirItem[]>(() => {
  if (!normalizedBrowseKeyword.value) return browseDirectories.value
  return browseDirectories.value.filter(item => (item.name || '').toLowerCase().includes(normalizedBrowseKeyword.value))
})
const filteredBrowseFiles = computed<BrowseFileItem[]>(() => {
  if (!normalizedBrowseKeyword.value) return browseFiles.value
  return browseFiles.value.filter(item => {
    const byName = (item.name || '').toLowerCase().includes(normalizedBrowseKeyword.value)
    const byPath = (item.relative_path || '').toLowerCase().includes(normalizedBrowseKeyword.value)
    const byKey = (item.object_key || '').toLowerCase().includes(normalizedBrowseKeyword.value)
    return byName || byPath || byKey
  })
})

const isImageLikeFile = (nameOrPath: string) => {
  const value = (nameOrPath || '').toLowerCase()
  return /\.(png|jpe?g|webp|gif|bmp|svg)$/i.test(value)
}

const browseThumbFiles = computed(() =>
  filteredBrowseFiles.value.map((file) => {
    const label = file.object_key || file.relative_path || file.name
    const previewUrl = browseMode.value === 's3'
      ? normalizeS3PreviewUrl(file.public_url || '', file.object_key || '')
      : resolveLocalPublicUrl(file.public_url || '')
    const imageLike = isImageLikeFile(file.name || label)
    return {
      ...file,
      label,
      preview_url: previewUrl,
      is_image: imageLike,
      can_preview: Boolean(previewUrl)
    }
  })
)

const previewFileEntries = computed(() =>
  filteredBrowseFiles.value
    .map((file) => {
      const key = file.object_key || file.relative_path || file.name
      const url = browseMode.value === 's3'
        ? normalizeS3PreviewUrl(file.public_url || '', file.object_key || '')
        : resolveLocalPublicUrl(file.public_url || '')
      return {
        key,
        url,
        name: file.name || '文件预览',
        path: file.object_key || file.relative_path || '',
        isImage: isImageLikeFile(file.name || file.object_key || file.relative_path || '')
      }
    })
    .filter(item => item.key && item.url)
)

const filePreviewCurrentIndex = computed(() =>
  previewFileEntries.value.findIndex(item => item.key === filePreviewKey.value)
)

const filePreviewHasPrev = computed(() => filePreviewCurrentIndex.value > 0)
const filePreviewHasNext = computed(() =>
  filePreviewCurrentIndex.value >= 0 && filePreviewCurrentIndex.value < previewFileEntries.value.length - 1
)

const filePreviewProgressText = computed(() => {
  if (filePreviewCurrentIndex.value < 0 || previewFileEntries.value.length === 0) return ''
  return `${filePreviewCurrentIndex.value + 1} / ${previewFileEntries.value.length}`
})

const applyPreviewPayload = (payload: { key: string, url: string, name: string, path: string, isImage: boolean }) => {
  filePreviewKey.value = payload.key
  filePreviewUrl.value = payload.url
  filePreviewName.value = payload.name
  filePreviewPath.value = payload.path
  filePreviewIsImage.value = payload.isImage
  filePreviewDialogOpen.value = true
}

const movePreviewBy = (step: number) => {
  if (previewFileEntries.value.length === 0 || filePreviewCurrentIndex.value < 0) return
  const nextIndex = filePreviewCurrentIndex.value + step
  if (nextIndex < 0 || nextIndex >= previewFileEntries.value.length) return
  applyPreviewPayload(previewFileEntries.value[nextIndex])
}

const browseThumbGridClass = computed(() => `app-file-browser-grid app-file-browser-grid--${browseThumbSize.value}`)

const browseDirectoryGridClass = computed(() => `app-file-browser-grid app-file-browser-grid--directories app-file-browser-grid--${browseThumbSize.value}`)

const hasBrowseContent = computed(() => filteredBrowseDirectories.value.length > 0 || filteredBrowseFiles.value.length > 0)

const browseThumbPreviewClass = computed(() => `storage-preview-tile app-file-browser-card-preview app-file-browser-card-preview--${browseThumbSize.value}`)

const openBrowseParent = async () => {
  if (browseMode.value === 's3') {
    await openS3ParentDirectory()
    return
  }
  await openLocalBrowseParent()
}

const openBrowseChild = async (item: BrowseDirItem) => {
  if (browseMode.value === 's3') {
    await openS3ChildDirectory(item)
    return
  }
  await openLocalBrowseChild(item)
}

const openBrowseBreadcrumb = async (path: string) => {
  if (browseMode.value === 's3') {
    await openS3Breadcrumb(path)
    return
  }
  await openLocalBrowseBreadcrumb(path)
}

const previewBrowseFile = (file: BrowseFileItem) => {
  if (browseMode.value === 's3') {
    previewS3File(file as S3FileItem)
    return
  }
  previewLocalFile(file as LocalBrowseFileItem)
}

const askDeleteBrowseFile = (file: BrowseFileItem) => {
  fileDeleteTargetName.value = file.name || ''
  fileDeleteTargetPath.value = file.relative_path || ''
  fileDeleteTargetObjectKey.value = file.object_key || ''
  fileDeleteConfirmOpen.value = true
}

const confirmDeleteBrowseFile = async () => {
  if (fileDeleting.value) return
  fileDeleting.value = true
  try {
    const profileID = browseMode.value === 's3' ? s3BrowseProfile.value?.id : localBrowseProfile.value?.id
    if (!profileID) {
      notifyError('未找到存储配置')
      return
    }
    const payload: Record<string, string> = {
      profile_id: profileID
    }
    if (browseMode.value === 's3') {
      payload.object_key = fileDeleteTargetObjectKey.value
      payload.relative_path = fileDeleteTargetPath.value
    } else {
      payload.relative_path = fileDeleteTargetPath.value
    }
    const rsp = await settingsApi.deleteStorageFile(payload)
    if (rsp?.data?.code !== 200) {
      notifyError(rsp?.data?.msg || '删除失败')
      return
    }
    if (filePreviewPath.value === fileDeleteTargetObjectKey.value || filePreviewPath.value === fileDeleteTargetPath.value) {
      filePreviewDialogOpen.value = false
    }
    fileDeleteConfirmOpen.value = false
    notifySuccess('删除成功')
    if (browseMode.value === 's3') {
      await loadS3Objects(s3BrowseCurrentPath.value || '')
    } else {
      await loadLocalBrowseFiles(localBrowseCurrentPath.value || '')
    }
  } finally {
    fileDeleting.value = false
  }
}

const openCreateDialog = () => {
  editingProfileId.value = ''
  resetEditor()
  editorOpen.value = true
}

const openEditDialog = (profile: StorageProfile) => {
  editingProfileId.value = profile.id
  Object.assign(editor, profile)
  editorOpen.value = true
}

const removeProfile = async (profileID: string) => {
  const next = profiles.value.filter(item => item.id !== profileID)
  if (next.length === 0) {
    notifyError('至少保留一个存储配置')
    return false
  }
  const nextDefaultStorageID = defaultStorageID.value === profileID ? next[0].id : defaultStorageID.value
  const ok = await persistConfig(next, nextDefaultStorageID, '删除存储成功')
  if (ok && editingProfileId.value === profileID) {
    editingProfileId.value = ''
  }
  return ok
}

const openDeleteConfirm = (profile: StorageProfile) => {
  deleteTarget.value = profile
  deleteConfirmInput.value = ''
  deleteConfirmOpen.value = true
}

const closeDeleteConfirm = () => {
  deleteConfirmOpen.value = false
  deleteConfirmInput.value = ''
  deleteTarget.value = null
}

const isDeleteMatch = computed(() => {
  const target = deleteTarget.value?.name || ''
  return deleteConfirmInput.value.trim().toLowerCase() === target.trim().toLowerCase() && target.length > 0
})

const showDeleteError = computed(() => {
  return deleteConfirmInput.value.length > 0 && !isDeleteMatch.value
})

const confirmDelete = async () => {
  if (!deleteTarget.value || !isDeleteMatch.value) return
  const ok = await removeProfile(deleteTarget.value.id)
  if (ok) {
    closeDeleteConfirm()
  }
}

const generateStorageId = () => {
  const exists = new Set(profiles.value.map(item => item.id))
  for (let i = 0; i < 50; i += 1) {
    const candidate = String(Math.floor(Math.random() * 90000000) + 10000000)
    if (!exists.has(candidate)) {
      return candidate
    }
  }
  return String(Date.now()).slice(-8)
}

const copyText = async (text: string, label: string) => {
  const value = (text || '').trim()
  if (!value) {
    notifyError(`${label}为空，无法复制`)
    return
  }
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(value)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = value
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    notifySuccess(`${label}已复制`)
  } catch (error) {
    notifyError(`${label}复制失败`)
  }
}

const normalizeLocalSubPath = (value: string) => {
  const normalized = (value || '').split('\\').join('/').trim().replace(/^\/+|\/+$/g, '')
  if (!normalized) return ''
  const parts = normalized.split('/').map((item: string) => item.trim()).filter(Boolean)
  const validParts = parts.filter((item: string) => item !== '.' && item !== '..')
  return validParts.join('/')
}

const hasInvalidLocalPathSegment = computed(() => {
  const raw = (editor.local_sub_path || '').split('\\').join('/')
  return raw.split('/').some((item: string) => item.trim() === '..')
})

const normalizedLocalSubPath = computed(() => {
  return normalizeLocalSubPath(editor.local_sub_path) || 'uploads'
})

const selectedLocalDirPath = computed(() => normalizedLocalSubPath.value)
const isCurrentDirSelected = computed(() => selectedLocalDirPath.value === (localDirCurrentPath.value || ''))

const openLocalDirDialog = async () => {
  localDirDialogOpen.value = true
  localDirNewFolderName.value = ''
  await loadLocalDirectories(normalizeLocalSubPath(editor.local_sub_path))
}

const loadLocalDirectories = async (path: string) => {
  localDirLoading.value = true
  try {
    const rsp = await settingsApi.listLocalDirectories({ path })
    const data = rsp?.data?.data || {}
    localDirCurrentPath.value = data.current_path || ''
    localDirParentPath.value = data.parent_path || ''
    localDirItems.value = Array.isArray(data.directories) ? data.directories : []
  } finally {
    localDirLoading.value = false
  }
}

const openChildDirectory = async (item: LocalDirItem) => {
  await loadLocalDirectories(item.relative_path || '')
}

const openParentDirectory = async () => {
  await loadLocalDirectories(localDirParentPath.value || '')
}

const localDirBreadcrumbs = computed<LocalDirBreadcrumb[]>(() => {
  const items: LocalDirBreadcrumb[] = [{ label: 'data', path: '' }]
  const current = (localDirCurrentPath.value || '').trim()
  if (!current) {
    return items
  }
  const parts = current.split('/').filter(Boolean)
  let acc = ''
  for (const part of parts) {
    acc = acc ? `${acc}/${part}` : part
    items.push({ label: part, path: acc })
  }
  return items
})

const openBreadcrumbDirectory = async (path: string) => {
  await loadLocalDirectories(path || '')
}

const createLocalDirectory = async () => {
  const folderName = (localDirNewFolderName.value || '').trim()
  if (!folderName) {
    notifyError('请输入目录名称')
    return
  }
  localDirCreating.value = true
  try {
    const rsp = await settingsApi.createLocalDirectory({
      path: localDirCurrentPath.value || '',
      name: folderName
    })
    if (rsp?.data?.code === 200) {
      notifySuccess('目录创建成功')
      localDirNewFolderName.value = ''
      await loadLocalDirectories(localDirCurrentPath.value || '')
      return
    }
    notifyError(rsp?.data?.msg || '目录创建失败')
  } finally {
    localDirCreating.value = false
  }
}

const chooseCurrentDirectory = () => {
  if (!localDirCurrentPath.value) {
    notifyError('请先选择子目录')
    return
  }
  editor.local_sub_path = localDirCurrentPath.value
  localDirDialogOpen.value = false
}

const handleLocalDirDialogKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Enter' && event.ctrlKey) {
    event.preventDefault()
    if (!localDirLoading.value && !!localDirCurrentPath.value && !isCurrentDirSelected.value) {
      chooseCurrentDirectory()
    }
  }
}

const applyEditor = async () => {
  if (!editor.name.trim()) {
    notifyError('请输入存储名称')
    return
  }
  if (editor.provider === 's3') {
    if (!editor.s3_endpoint.trim() || !editor.s3_bucket.trim() || !editor.s3_object_key_prefix.trim() || !editor.s3_access_key.trim() || !editor.s3_secret_key.trim()) {
      notifyError('Please fill Endpoint, Bucket, Object Prefix, Access Key, Secret Key')
      return
    }
  } else if (!editor.local_sub_path.trim()) {
    notifyError('请输入本地路径')
    return
  } else if (hasInvalidLocalPathSegment.value) {
    notifyError('本地路径不能包含 ..')
    return
  }
  const payload: StorageProfile = {
    ...editor,
    id: editingProfileId.value ? editor.id.trim() : generateStorageId(),
    name: editor.name.trim(),
    upload_file_prefix: editor.upload_file_prefix.trim() || 'upload',
    local_sub_path: normalizedLocalSubPath.value,
    s3_endpoint: editor.s3_endpoint.trim(),
    s3_region: editor.s3_region.trim(),
    s3_bucket: editor.s3_bucket.trim(),
    s3_access_key: editor.s3_access_key.trim(),
    s3_secret_key: editor.s3_secret_key.trim(),
    s3_public_base_url: editor.s3_public_base_url.trim(),
    s3_object_key_prefix: editor.s3_object_key_prefix.trim()
  }
  const nextProfiles = [...profiles.value]
  const existsIndex = nextProfiles.findIndex(item => item.id === payload.id)
  if (existsIndex >= 0) {
    nextProfiles.splice(existsIndex, 1, payload)
  } else {
    nextProfiles.push(payload)
  }
  let nextDefaultStorageID = defaultStorageID.value
  if (!nextDefaultStorageID) {
    nextDefaultStorageID = payload.id
  }
  const ok = await persistConfig(nextProfiles, nextDefaultStorageID, '存储配置已保存')
  if (ok) {
    if (!editingProfileId.value) {
      editingProfileId.value = payload.id
    }
    editorOpen.value = false
  }
}

const setDefaultStorage = async (profileID: string) => {
  if (defaultStorageID.value === profileID) return
  await persistConfig([...profiles.value], profileID, '默认存储已更新')
}

onMounted(async () => {
  await loadConfig()
  window.addEventListener('keydown', handlePreviewKeydown)
})

onBeforeUnmount(() => {
  if (localUploadAutoCloseTimer.value !== null) {
    window.clearInterval(localUploadAutoCloseTimer.value)
    localUploadAutoCloseTimer.value = null
  }
  window.removeEventListener('keydown', handlePreviewKeydown)
})
</script>

<template>
  <div class="space-y-3">
    <div class="storage-toolbar flex items-center justify-between gap-3">
      <div class="text-sm text-muted-foreground">支持配置多个存储实例，并可设置默认存储供未指定模块兜底使用</div>
      <el-button type="primary" @click="openCreateDialog">新建存储</el-button>
    </div>

    <el-table
      :data="profiles"
      row-key="id"
      class="app-data-table storage-config-table w-full"
      stripe
    >
      <el-table-column prop="id" label="存储ID" width="120">
        <template #default="{ row }">
          <div class="min-w-0 flex items-center gap-2">
            <span class="text-xs text-muted-foreground font-mono truncate" :title="row.id">{{ row.id }}</span>
            <button type="button" class="storage-icon-action" title="复制存储ID" aria-label="复制存储ID" @click="copyText(row.id, '存储ID')">
              <CopyOutlined class="text-[14px]" />
            </button>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="存储名称" min-width="160">
        <template #default="{ row }">
          <div class="flex items-center gap-2 min-w-0">
            <span class="text-sm font-medium truncate" :title="row.name">{{ row.name }}</span>
            <button type="button" class="storage-icon-action shrink-0" title="复制存储名称" aria-label="复制存储名称" @click="copyText(row.name, '存储名称')">
              <CopyOutlined class="text-[14px]" />
            </button>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="provider" label="类型" width="90" align="center">
        <template #default="{ row }">{{ row.provider === 's3' ? 'S3' : '本地' }}</template>
      </el-table-column>
      <el-table-column label="路径" width="220">
        <template #default="{ row }">
          <span v-if="row.provider === 's3'" class="inline-flex items-center gap-2 max-w-full">
            <span class="storage-path-text" :title="`${row.s3_bucket}/${row.s3_object_key_prefix}`">{{ row.s3_bucket }}/{{ row.s3_object_key_prefix }}</span>
            <button type="button" class="storage-icon-action storage-icon-action--path shrink-0" title="复制路径" aria-label="复制路径" @click="copyText(`${row.s3_bucket}/${row.s3_object_key_prefix}`, '路径')">
              <LinkOutlined class="text-[14px]" />
            </button>
          </span>
          <span v-else class="inline-flex items-center gap-2 max-w-full">
            <span class="storage-path-text" :title="`./data/${row.local_sub_path || 'uploads'}`">./data/{{ row.local_sub_path || 'uploads' }}</span>
            <button type="button" class="storage-icon-action storage-icon-action--path shrink-0" title="复制路径" aria-label="复制路径" @click="copyText(`./data/${row.local_sub_path || 'uploads'}`, '路径')">
              <LinkOutlined class="text-[14px]" />
            </button>
          </span>
        </template>
      </el-table-column>
      <el-table-column label="默认" width="88" align="center">
        <template #default="{ row }">
          <button
            type="button"
            class="storage-default-action inline-flex items-center justify-center h-8 w-8 rounded-full border transition-all"
            :class="defaultStorageID === row.id ? 'storage-default-action-active cursor-default pointer-events-none' : 'storage-default-action-idle'"
            :title="defaultStorageID === row.id ? '默认存储' : '设为默认存储'"
            :disabled="saving || loading"
            @click="setDefaultStorage(row.id)"
          >
            <StarFilled v-if="defaultStorageID === row.id" class="text-[15px]" />
            <StarOutlined v-else class="text-[15px]" />
          </button>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" align="center" fixed="right">
        <template #default="{ row }">
          <AppRowActions :actions="[
            { key: 'edit', label: '编辑', kind: 'write', permission: 'system:settings:edit', disabled: saving || loading, onClick: () => openEditDialog(row) },
            { key: 'delete', label: '删除', kind: 'write', permission: 'system:settings:edit', danger: true, disabled: saving || loading, onClick: () => openDeleteConfirm(row) },
            { key: 'upload', label: localUploadTesting ? '上传中' : '上传', kind: 'write', permission: 'system:settings:edit', disabled: localUploadTesting || saving || loading, loading: localUploadTesting, onClick: () => handleTestClick(row) },
            { key: 'browse', label: '浏览', kind: 'view', permission: 'system:settings:view', disabled: saving || loading, onClick: () => handleBrowseClick(row) }
          ]" />
        </template>
      </el-table-column>
    </el-table>

    <AppFormDrawer
      v-model="editorOpen"
      :title="editingProfileId ? '编辑存储配置' : '新建存储配置'"
      size="min(760px, 96vw)"
      confirm-text="保存"
      :loading="saving"
      @confirm="applyEditor"
    >
        <div class="storage-profile-dialog-body">
          <section class="storage-profile-section">
            <div class="settings-card-heading">
              <div class="settings-card-title">存储配置</div>
              <div class="settings-card-description">填写基本信息，并选择 Provider 配置对应的路径或连接凭据。</div>
            </div>
            <div class="storage-inline-fields">
              <div class="app-form-field storage-profile-inline-field">
                <label class="app-form-label">存储名称</label>
                <Input v-model="editor.name" placeholder="请输入存储名称" />
              </div>
              <div class="app-form-field storage-profile-inline-field">
                <label class="app-form-label">上传文件名前缀</label>
                <Input v-model="editor.upload_file_prefix" placeholder="默认 upload" />
              </div>
            </div>
            <div class="storage-profile-provider-panel">
              <div class="settings-card-heading storage-provider-heading"><div class="settings-card-title">Provider / 连接配置</div><div class="settings-card-description">选择存储类型并填写对应配置。</div></div>
              <div class="storage-profile-tabs" role="tablist" aria-label="存储类型">
                <button
                  id="storage-provider-local-tab"
                  type="button"
                  role="tab"
                  class="storage-profile-tab"
                  :class="editor.provider === 'local' ? 'storage-profile-tab-active' : ''"
                  :aria-selected="editor.provider === 'local'"
                  aria-controls="storage-provider-local-panel"
                  @click="editor.provider = 'local'"
                >
                  本地存储
                </button>
                <button
                  id="storage-provider-s3-tab"
                  type="button"
                  role="tab"
                  class="storage-profile-tab"
                  :class="editor.provider === 's3' ? 'storage-profile-tab-active' : ''"
                  :aria-selected="editor.provider === 's3'"
                  aria-controls="storage-provider-s3-panel"
                  @click="editor.provider = 's3'"
                >
                  S3 存储
                </button>
              </div>

            <div
              v-if="editor.provider === 'local'"
              id="storage-provider-local-panel"
              class="storage-profile-subsection"
              role="tabpanel"
              aria-labelledby="storage-provider-local-tab"
            >
            <div class="app-form-field">
              <label class="app-form-label">本地路径</label>
              <div class="storage-local-path-control">
                <div class="storage-prefix-chip">./data/</div>
                <Input v-model="editor.local_sub_path" placeholder="例如 uploads/oidc" />
                <el-button
                  type="button"
                  text
                  title="浏览目录"
                  aria-label="浏览本地目录"
                  class="storage-folder-button"
                  @click="openLocalDirDialog"
                >
                  <FolderOpenOutlined class="text-[16px]" />
                </el-button>
              </div>
              <div class="app-form-help">路径会拼接在 ./data/ 下，保存时自动创建不存在的目录。</div>
              <div v-if="hasInvalidLocalPathSegment" class="app-form-error">路径不能包含 ..，保存时会拦截</div>
            </div>
          </div>

            <div
              v-if="editor.provider === 's3'"
              id="storage-provider-s3-panel"
              class="storage-profile-subsection"
              role="tabpanel"
              aria-labelledby="storage-provider-s3-tab"
            >
            <div class="storage-s3-fields">
              <div class="app-form-field storage-s3-inline-field">
                <label class="app-form-label">服务地址</label>
                <Input v-model="editor.s3_endpoint" placeholder="请输入 S3 服务地址" />
              </div>
              <div class="app-form-field storage-s3-inline-field">
                <label class="app-form-label">区域</label>
                <Input v-model="editor.s3_region" placeholder="请输入区域，可选" />
              </div>
              <div class="app-form-field storage-s3-inline-field">
                <label class="app-form-label">存储桶</label>
                <Input v-model="editor.s3_bucket" placeholder="请输入存储桶名称" />
              </div>
              <div class="app-form-field storage-s3-inline-field">
                <label class="app-form-label">对象前缀</label>
                <Input v-model="editor.s3_object_key_prefix" placeholder="请输入对象路径前缀" />
              </div>
              <div class="app-form-field storage-s3-inline-field">
                <label class="app-form-label">访问密钥</label>
                <Input v-model="editor.s3_access_key" placeholder="请输入 Access Key" />
              </div>
              <div class="app-form-field storage-s3-inline-field">
                <label class="app-form-label">密钥</label>
                <Input v-model="editor.s3_secret_key" placeholder="请输入 Secret Key" />
              </div>
              <div class="app-form-field storage-s3-inline-field storage-s3-inline-field-wide">
                <label class="app-form-label">公开地址</label>
                <Input v-model="editor.s3_public_base_url" placeholder="请输入公开访问地址，可选" />
              </div>
            </div>
            <div class="storage-s3-switches">
              <div class="app-form-field storage-switch-field">
                <label class="app-form-label" for="storage-s3-use-ssl">使用 HTTPS</label>
                <div class="storage-switch-control">
                  <el-switch id="storage-s3-use-ssl" v-model="editor.s3_use_ssl" />
                  <span>{{ editor.s3_use_ssl ? '已开启' : '已关闭' }}</span>
                </div>
              </div>
              <div class="app-form-field storage-switch-field">
                <label class="app-form-label" for="storage-s3-proxy-public-read">代理公开读取</label>
                <div class="storage-switch-control">
                  <el-switch id="storage-s3-proxy-public-read" v-model="editor.s3_proxy_public_read" />
                  <span>{{ editor.s3_proxy_public_read ? '已开启' : '已关闭' }}</span>
                </div>
              </div>
            </div>
          </div>
          </div>
          </section>
        </div>
    </AppFormDrawer>

    <Dialog :open="deleteConfirmOpen" @update:open="(value) => value ? (deleteConfirmOpen = true) : closeDeleteConfirm()">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除存储配置</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-muted-foreground">
            请输入要删除的存储名称
            <span v-if="deleteTarget?.name" class="text-red-500 font-semibold mx-1">{{ deleteTarget.name }}</span>
            以确认操作
          </div>
          <Input
            v-model="deleteConfirmInput"
            :max-length="100"
            placeholder="请输入存储名称"
          />
          <div v-if="showDeleteError" class="text-xs text-red-500">名称不匹配，请重新输入</div>
        </div>
        <DialogFooter class="flex justify-end gap-2 mt-4">
          <Button type="button" variant="outline" @click="closeDeleteConfirm">取消</Button>
          <Button type="button" :disabled="!isDeleteMatch" @click="confirmDelete">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="browseDialogOpen" @update:open="(value) => { browseDialogOpen = value }">
      <DialogContent class="app-file-browser-dialog">
        <DialogHeader>
          <DialogTitle>{{ browseDialogTitle }}</DialogTitle>
        </DialogHeader>
        <div class="app-file-browser">
          <section class="app-file-browser-metadata" aria-label="当前存储信息">
            <span class="app-file-browser-metadata__icon" aria-hidden="true">
              <CloudServerOutlined />
            </span>
            <div class="app-file-browser-metadata__copy" :title="browseProfileName || ''">
              <strong>{{ browseProfileName || '-' }}</strong>
              <small>{{ browseMode === 's3' ? 'S3' : '本地' }} · {{ browseProfileId || '-' }}</small>
            </div>
          </section>
          <div class="app-file-browser-controls">
            <div class="app-file-browser-controls__navigation app-file-browser-navigation-compact">
              <button class="storage-local-dir-back" type="button" :disabled="!browseCurrentPath || browseLoading" @click="openBrowseParent">
                <ArrowLeftOutlined />
                <span>上一级</span>
              </button>
              <nav class="app-file-browser-breadcrumb" aria-label="当前路径">
                <div class="app-file-browser-breadcrumb__track">
                  <button
                    v-for="(crumb, index) in browseBreadcrumbs"
                    :key="`${crumb.path || 'browse-root'}-${index}`"
                    type="button"
                    class="storage-breadcrumb-button"
                    :disabled="browseLoading"
                    @click="openBrowseBreadcrumb(crumb.path)"
                  >
                    <RightOutlined v-if="index > 0" />
                    <FolderOutlined v-if="index === 0" />
                    <span>{{ crumb.label }}</span>
                  </button>
                </div>
              </nav>
            </div>
            <Input v-model="browseKeyword" class="app-file-browser-toolbar__search" placeholder="按名称筛选文件/目录" />
          </div>
          <div class="storage-browser-list app-file-browser-content">
            <div class="app-file-browser-content-toolbar">
              <div class="app-file-browser-content-toolbar__actions">
                <div class="app-segmented app-file-browser-toolbar__view" role="group" aria-label="浏览视图">
                  <button type="button" class="storage-segment-button" :class="browseViewMode === 'list' ? 'storage-segment-button-active' : 'storage-segment-button-idle'" :aria-pressed="browseViewMode === 'list'" @click="browseViewMode = 'list'">列表</button>
                  <button type="button" class="storage-segment-button" :class="browseViewMode === 'thumb' ? 'storage-segment-button-active' : 'storage-segment-button-idle'" :aria-pressed="browseViewMode === 'thumb'" @click="browseViewMode = 'thumb'">缩略图</button>
                </div>
                <div v-if="browseViewMode === 'thumb' && !browseLoading && hasBrowseContent" class="app-segmented app-file-browser-toolbar__size" role="group" aria-label="缩略图密度">
                  <button v-for="size in (['sm', 'md', 'lg'] as const)" :key="size" type="button" class="storage-segment-button" :class="browseThumbSize === size ? 'storage-segment-button-active' : 'storage-segment-button-idle'" :aria-pressed="browseThumbSize === size" @click="browseThumbSize = size">{{ size === 'sm' ? '紧凑' : size === 'md' ? '标准' : '宽松' }}</button>
                </div>
              </div>
            </div>
            <div v-if="browseLoading" class="app-file-browser-state" role="status">正在读取当前目录…</div>
            <template v-else-if="browseViewMode === 'list'">
              <button v-for="item in filteredBrowseDirectories" :key="item.relative_path" type="button" class="storage-browser-row app-file-browser-row app-file-browser-row--directory" @click="openBrowseChild(item)">
                <span class="app-file-browser-row-main">
                  <span class="app-file-browser-folder-icon"><FolderOutlined /></span>
                  <span class="app-file-browser-row-copy"><strong>{{ item.name }}</strong><small v-if="item.relative_path !== item.name">{{ item.relative_path }}</small></span>
                </span>
              </button>
              <div v-for="file in filteredBrowseFiles" :key="file.object_key || file.relative_path" class="storage-browser-row app-file-browser-row">
                <div class="app-file-browser-row-main">
                  <span class="app-file-browser-row-copy"><strong>{{ file.name }}</strong><small :title="browseMode === 's3' ? (file.object_key || file.relative_path) : file.relative_path">{{ browseMode === 's3' ? (file.object_key || file.relative_path) : file.relative_path }}</small></span>
                </div>
                <div class="app-file-browser-row-actions">
                  <span v-if="typeof file.size === 'number'" class="app-file-browser-file-size">{{ formatFileSize(file.size || 0) }}</span>
                  <Button type="button" variant="outline" size="sm" :disabled="browseMode === 's3' && !file.public_url" @click="previewBrowseFile(file)">预览</Button>
                  <Button type="button" variant="outline" size="sm" @click="copyText(file.public_url || file.object_key || file.relative_path, file.public_url ? '文件链接' : (browseMode === 's3' ? '对象键' : '相对路径'))">复制</Button>
                  <Button type="button" variant="destructive" size="sm" @click="askDeleteBrowseFile(file)">删除</Button>
                </div>
              </div>
            </template>
            <template v-else>
              <div v-if="filteredBrowseDirectories.length" :class="browseDirectoryGridClass">
                <button v-for="item in filteredBrowseDirectories" :key="`thumb-dir-${item.relative_path}`" type="button" class="storage-thumb-dir-card app-file-browser-directory-card" @click="openBrowseChild(item)">
                  <span class="app-file-browser-folder-icon"><FolderOutlined /></span><strong>{{ item.name }}</strong>
                </button>
              </div>
              <div :class="browseThumbGridClass">
                <div v-for="file in browseThumbFiles" :key="`thumb-file-${file.object_key || file.relative_path}`" class="storage-thumb-file-card app-file-browser-card">
                  <button type="button" :class="browseThumbPreviewClass" :disabled="!file.can_preview" @click="previewBrowseFile(file)">
                    <img v-if="file.is_image && file.can_preview" :src="file.preview_url" :alt="file.name"><FolderOpenOutlined v-else />
                  </button>
                  <div class="app-file-browser-card-copy"><strong :title="file.name">{{ file.name }}</strong><small :title="file.label">{{ file.label }}</small></div>
                  <div class="app-file-browser-card-actions">
                    <Button type="button" variant="outline" size="sm" :disabled="browseMode === 's3' && !file.public_url" @click="previewBrowseFile(file)">预览</Button>
                    <Button type="button" variant="outline" size="sm" @click="copyText(file.public_url || file.object_key || file.relative_path, file.public_url ? '文件链接' : (browseMode === 's3' ? '对象键' : '相对路径'))">复制</Button>
                    <Button type="button" variant="destructive" size="sm" @click="askDeleteBrowseFile(file)">删除</Button>
                  </div>
                </div>
              </div>
            </template>
            <div v-if="!browseLoading && filteredBrowseDirectories.length === 0 && filteredBrowseFiles.length === 0" class="app-file-browser-state app-file-browser-state--empty">
              <FolderOpenOutlined /><strong>当前路径下没有内容</strong><span>可返回上级目录或调整筛选条件</span>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <Dialog :open="filePreviewDialogOpen" @update:open="(value) => { filePreviewDialogOpen = value }">
      <DialogContent class="app-nested-dialog app-file-preview-dialog !w-[min(92vw,760px)] !max-w-[760px]">
        <DialogHeader>
          <DialogTitle>文件预览</DialogTitle>
        </DialogHeader>
        <div class="app-file-preview-body">
          <div class="text-sm font-medium truncate" :title="filePreviewName">{{ filePreviewName }}</div>
          <div class="text-xs text-muted-foreground truncate" :title="filePreviewPath">{{ filePreviewPath }}</div>
          <div class="storage-preview-frame overflow-hidden">
            <img
              v-if="filePreviewIsImage"
              :src="filePreviewUrl"
              class="w-full h-full object-contain"
              :alt="filePreviewName"
            >
            <iframe
              v-else
              :src="filePreviewUrl"
              class="w-full h-full border-0"
            />
          </div>
        </div>
        <DialogFooter class="app-file-preview-footer">
          <div class="app-file-preview-footer__navigation">
            <Button type="button" variant="outline" :disabled="!filePreviewHasPrev" @click="movePreviewBy(-1)">上一张</Button>
            <span v-if="filePreviewProgressText">{{ filePreviewProgressText }}</span>
            <Button type="button" variant="outline" :disabled="!filePreviewHasNext" @click="movePreviewBy(1)">下一张</Button>
          </div>
          <div class="app-file-preview-footer__actions">
            <Button type="button" variant="outline" @click="copyText(filePreviewUrl, '预览地址')">复制地址</Button>
            <Button type="button" variant="outline" @click="openPreviewInNewTab">新标签打开</Button>
            <Button type="button" @click="filePreviewDialogOpen = false">关闭</Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="fileDeleteConfirmOpen" @update:open="(value) => { fileDeleteConfirmOpen = value }">
      <DialogContent class="app-nested-dialog max-w-[480px]">
        <DialogHeader>
          <DialogTitle>确认删除文件</DialogTitle>
        </DialogHeader>
        <div class="space-y-2 text-sm">
          <div class="text-foreground">该操作不可恢复，确认要删除以下文件吗？</div>
          <div class="storage-danger-summary px-3 py-2">
            <div class="font-medium truncate" :title="fileDeleteTargetName">{{ fileDeleteTargetName }}</div>
            <div class="text-xs text-muted-foreground truncate" :title="fileDeleteTargetObjectKey || fileDeleteTargetPath">{{ fileDeleteTargetObjectKey || fileDeleteTargetPath }}</div>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" :disabled="fileDeleting" @click="fileDeleteConfirmOpen = false">取消</Button>
          <Button type="button" variant="destructive" :disabled="fileDeleting" @click="confirmDeleteBrowseFile">
            {{ fileDeleting ? '删除中...' : '确认删除' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="localUploadTestOpen" @update:open="(value) => value ? (localUploadTestOpen = true) : closeLocalUploadTest()">
      <DialogContent class="storage-upload-dialog">
        <DialogHeader>
          <DialogTitle>存储上传</DialogTitle>
        </DialogHeader>
        <div class="storage-upload-panel">
          <section class="storage-upload-metadata" aria-label="当前存储信息">
            <span class="storage-upload-metadata__icon" aria-hidden="true">
              <CloudServerOutlined />
            </span>
            <div class="storage-upload-metadata__copy" :title="localUploadTestProfile?.name">
              <strong>{{ localUploadTestProfile?.name }}</strong>
              <small>{{ localUploadTestProfile?.provider === 's3' ? 'S3' : '本地' }} · {{ localUploadTestProfile?.id }}</small>
            </div>
          </section>

          <input
            ref="localUploadFileInputRef"
            type="file"
            class="hidden"
            tabindex="-1"
            aria-hidden="true"
            @change="handleLocalUploadFileChange"
          />
          <div class="storage-upload-controls" :class="{ 'storage-upload-controls--delete-active': localUploadDeleteAfter }">
            <button
              type="button"
              class="storage-upload-option"
              :class="{ 'storage-upload-option--active': localUploadDeleteAfter }"
              :aria-pressed="localUploadDeleteAfter"
              :title="localUploadDeleteAfter ? '上传成功后将立即删除该文件' : '上传成功后保留该文件'"
              @click="localUploadDeleteAfter = !localUploadDeleteAfter"
            >
              <span class="storage-upload-option__indicator" aria-hidden="true"></span>
              <span><strong>上传后立即删除</strong><small>用于仅验证写入能力的测试；开启后，上传成功的文件不会保留。</small></span>
            </button>
            <Button class="storage-upload-file-button" type="button" variant="outline" size="sm" :disabled="localUploadTesting" :aria-label="localUploadTestFile ? '重新选择上传文件' : '选择上传文件'" @click="localUploadFileInputRef?.click()">
              <FolderOpenOutlined aria-hidden="true" />
              {{ localUploadTestFile ? '重新选择' : '选择文件' }}
            </Button>
          </div>

          <div v-if="localUploadTestFile" class="storage-upload-selected-file" aria-live="polite">
            <FileOutlined aria-hidden="true" />
            <strong :title="localUploadSelectedFileName">{{ localUploadSelectedFileName }}</strong>
            <small>{{ localUploadSelectedFileSize }}</small>
          </div>

          <section v-if="localUploadLastLocation" class="storage-upload-result" aria-live="polite">
            <div class="storage-upload-result__copy">
              <strong>上传成功地址</strong>
              <span :title="localUploadLastLocation">{{ localUploadLastLocation }}</span>
            </div>
            <Button type="button" variant="outline" size="sm" :disabled="localUploadTesting" @click="copyText(localUploadLastLocation, '上传地址')">
              <CopyOutlined aria-hidden="true" />
              复制地址
            </Button>
          </section>
        </div>
        <DialogFooter class="storage-upload-footer">
          <span v-if="localUploadAutoCloseCountdown > 0" class="storage-upload-footer__status" role="status">
            上传成功，{{ localUploadAutoCloseCountdown }} 秒后自动关闭
          </span>
          <div class="storage-upload-footer__actions">
            <Button type="button" class="storage-upload-footer__cancel" variant="outline" :disabled="localUploadTesting" @click="closeLocalUploadTest">取消</Button>
            <el-button
              type="primary"
              class="storage-upload-footer__submit"
              :loading="localUploadTesting"
              :disabled="localUploadTesting"
              @click="submitLocalUploadTest"
            >
              <UploadOutlined v-if="!localUploadTesting" aria-hidden="true" />
              {{ localUploadTesting ? '上传中...' : '上传' }}
            </el-button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <el-dialog
      v-model="localDirDialogOpen"
      title="选择本地目录"
      width="min(560px, calc(100dvw - 24px))"
      class="app-nested-dialog storage-local-dir-dialog"
      append-to-body
      @keydown="handleLocalDirDialogKeydown"
    >
      <div class="storage-local-dir-shell">
      <section class="storage-local-dir-parent-summary">
        <span><CloudServerOutlined /></span>
        <div><strong>{{ editor.name || '当前存储配置' }}</strong><small>本地存储 · ./data/{{ selectedLocalDirPath }}</small></div>
      </section>
      <div class="storage-local-dir-description">浏览 ./data 下的目录；选择后仅覆盖当前存储的本地目录路径。</div>
      <div class="storage-local-dir-drawer app-file-browser">
          <div class="app-file-browser-controls">
            <div class="storage-local-dir-navigation app-file-browser-controls__navigation">
              <button class="storage-local-dir-back" type="button" :disabled="!localDirCurrentPath || localDirLoading" @click="openParentDirectory">
                <ArrowLeftOutlined />
                <span>上一级</span>
              </button>
              <nav class="app-file-browser-breadcrumb" aria-label="当前目录">
                <div class="app-file-browser-breadcrumb__track">
                  <button
                    v-for="(crumb, index) in localDirBreadcrumbs"
                    :key="crumb.path || 'data-root'"
                    type="button"
                    class="storage-breadcrumb-button"
                    :class="index === localDirBreadcrumbs.length - 1 && isCurrentDirSelected ? 'storage-breadcrumb-button-selected' : ''"
                    :disabled="localDirLoading"
                    @click="openBreadcrumbDirectory(crumb.path)"
                  >
                    <RightOutlined v-if="index > 0" />
                    <FolderOutlined v-if="index === 0" />
                    <span>{{ crumb.label }}</span>
                  </button>
                </div>
              </nav>
            </div>
            <div class="storage-local-dir-action-row">
              <div class="storage-local-dir-create">
                <Input
                  v-model="localDirNewFolderName"
                  placeholder="新建子目录名称"
                  :disabled="localDirLoading || localDirCreating"
                  @keyup.enter="createLocalDirectory"
                />
                <el-button :disabled="localDirLoading || localDirCreating" @click="createLocalDirectory">
                  {{ localDirCreating ? '创建中...' : '新建目录' }}
                </el-button>
              </div>
            </div>
            <div class="storage-local-dir-tip">Enter 创建目录，Ctrl + Enter 选择当前目录</div>
          </div>
          <div class="storage-browser-list storage-local-dir-list app-file-browser-content">
            <div v-if="localDirLoading" class="app-file-browser-state" role="status">正在读取当前目录…</div>
            <template v-else>
              <button
                v-for="item in localDirItems"
                :key="item.relative_path"
                type="button"
                class="storage-browser-row storage-local-dir-row app-file-browser-row"
                :class="item.relative_path === selectedLocalDirPath ? 'storage-browser-row-selected' : ''"
                @click="openChildDirectory(item)"
              >
                <span class="app-file-browser-row-main">
                  <span class="app-file-browser-folder-icon"><FolderOutlined /></span>
                  <span class="app-file-browser-row-copy">
                    <strong>{{ item.name }}</strong>
                  </span>
                </span>
              </button>
              <div v-if="localDirItems.length === 0" class="app-file-browser-state app-file-browser-state--empty">
                <FolderOpenOutlined />
                <strong>当前目录下没有子目录</strong>
                <span>可返回上级目录或新建子目录</span>
              </div>
            </template>
          </div>
      </div>
      </div>
      <template #footer>
        <div class="storage-local-dir-footer"><span>当前目录：./data/{{ localDirCurrentPath || '-' }}</span><div><el-button @click="localDirDialogOpen = false">取消</el-button><el-button class="storage-local-dir-select" type="primary" :disabled="localDirLoading || !localDirCurrentPath || isCurrentDirSelected" @click="chooseCurrentDirectory">{{ isCurrentDirSelected ? '已选择' : '选择当前目录' }}</el-button></div></div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.storage-profile-dialog-body { display: grid; gap: 12px; }
.storage-profile-section { padding: 14px; border: 1px solid var(--app-overlay-border); border-radius: 9px; background: var(--app-overlay-surface); }
.storage-profile-section .storage-profile-provider-panel { overflow: visible; border: 0; border-radius: 0; background: transparent; }
.storage-profile-section .storage-profile-tabs { padding-inline: 0; background: transparent; }
.storage-provider-heading { margin-top: 4px; }
.storage-local-dir-parent-summary > span { display: inline-flex; flex: none; align-items: center; justify-content: center; width: 38px; height: 38px; border-radius: 9px; background: color-mix(in srgb, var(--brand-500) 10%, transparent); color: var(--brand-700); }
.storage-local-dir-shell { display: flex; flex-direction: column; gap: 10px; min-height: 0; }
.storage-local-dir-parent-summary { display: flex; align-items: center; gap: 12px; padding: 12px 14px; border: 1px solid var(--app-overlay-border); border-radius: 8px; background: color-mix(in srgb, var(--brand-50) 30%, var(--app-overlay-surface)); }
.storage-local-dir-parent-summary div { display: grid; min-width: 0; }
.storage-local-dir-parent-summary strong { font-size: 13px; }
.storage-local-dir-parent-summary small { overflow: hidden; color: var(--admin-text-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.storage-local-dir-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; width: 100%; }
.storage-local-dir-footer > span { overflow: hidden; color: var(--admin-text-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.storage-local-dir-footer > div { display: flex; flex: none; gap: 8px; }
@media (max-width: 760px) { .storage-inline-fields, .storage-s3-fields { grid-template-columns: 1fr; } .storage-local-dir-footer { align-items: flex-start; flex-direction: column; } .storage-local-dir-footer > div { width: 100%; } .storage-local-dir-footer :deep(.el-button) { flex: 1; } }
</style>
