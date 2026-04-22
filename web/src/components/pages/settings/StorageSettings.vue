<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Copy, Folder, FolderOpen, ChevronRight, CheckCircle2, Circle } from 'lucide-vue-next'

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
const browseThumbSize = ref<'sm' | 'md' | 'lg'>('sm')
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
    const rsp = await request.get('/system/storage-config')
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
    toast.error('存储ID必须为8位数字')
    return false
  }
  if (!isEightDigitStorageId(nextDefaultStorageID)) {
    toast.error('默认存储ID必须为8位数字')
    return false
  }
  saving.value = true
  try {
    const rsp = await request.post('/system/storage-config', {
      default_storage_id: nextDefaultStorageID,
      profiles: nextProfiles
    })
    if (rsp?.data?.code === 200) {
      profiles.value = nextProfiles
      defaultStorageID.value = nextDefaultStorageID
      toast.success(successText)
      return true
    }
    toast.error(rsp?.data?.msg || '存储配置保存失败')
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
    toast.error('未找到本地存储配置')
    return
  }
  if (!localUploadTestFile.value) {
    toast.error('请选择要上传的测试文件')
    return
  }
  localUploadTesting.value = true
  try {
    const formData = new FormData()
    formData.append('profile_id', localUploadTestProfile.value.id)
    formData.append('file', localUploadTestFile.value)
    formData.append('delete_after_upload', localUploadDeleteAfter.value ? '1' : '0')
    let rsp = await request.post('/system/storage-config/upload-file', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    if ((rsp as any)?.status === 404 && localUploadTestProfile.value.provider === 'local') {
      rsp = await request.post('/system/storage-config/test-local-upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
    }
    if (rsp?.data?.code === 200) {
      const objectKey = rsp?.data?.data?.object_key
      const publicUrl = rsp?.data?.data?.public_url || joinPublicUrlAndObjectKey(localUploadTestProfile.value.s3_public_base_url, objectKey)
      const deleted = rsp?.data?.data?.deleted === 'true'
      const suffix = deleted ? '（已立即删除）' : ''
      const location = publicUrl || objectKey || ''
      localUploadLastLocation.value = location
      toast.success(location ? `${rsp?.data?.msg || '文件上传成功'}：${location}${suffix}` : `${rsp?.data?.msg || '文件上传成功'}${suffix}`)
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
    toast.error(rsp?.data?.msg || '文件上传失败')
  } catch (error: any) {
    if (error?.response?.status === 404) {
      toast.error('上传接口不存在，请重启后端服务后重试')
      return
    }
    toast.error(error?.response?.data?.msg || '文件上传失败')
  } finally {
    localUploadTesting.value = false
  }
}

const handleTestClick = async (profile: StorageProfile) => {
  openLocalUploadTest(profile)
}

const openS3BrowseDialog = async (profile: StorageProfile) => {
  if (profile.provider !== 's3') {
    toast.error('仅支持浏览 S3 存储')
    return
  }
  s3BrowseProfile.value = profile
  s3BrowseDialogOpen.value = true
  browseKeyword.value = ''
  browseViewMode.value = 'list'
  browseThumbSize.value = 'sm'
  await loadS3Objects('')
}

const loadS3Objects = async (path: string) => {
  if (!s3BrowseProfile.value) return
  s3BrowseLoading.value = true
  try {
    const rsp = await request.get('/system/storage-config/s3-objects', {
      params: {
        profile_id: s3BrowseProfile.value.id,
        path
      }
    })
    const data = rsp?.data?.data || {}
    s3BrowseCurrentPath.value = data.current_path || ''
    s3BrowseParentPath.value = data.parent_path || ''
    s3BrowsePrefix.value = data.prefix || ''
    s3BrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
    s3BrowseFiles.value = Array.isArray(data.files) ? data.files : []
  } catch (error: any) {
    toast.error(error?.response?.data?.msg || '读取 S3 对象失败')
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
    toast.error('当前存储未配置 Public URL，无法预览')
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
    toast.error('仅支持浏览本地存储')
    return
  }
  localBrowseProfile.value = profile
  localBrowseDialogOpen.value = true
  browseKeyword.value = ''
  browseViewMode.value = 'list'
  browseThumbSize.value = 'sm'
  await loadLocalBrowseFiles('')
}

const loadLocalBrowseFiles = async (path: string) => {
  if (!localBrowseProfile.value) return
  localBrowseLoading.value = true
  try {
    const rsp = await request.get('/system/storage-config/local-files', {
      params: {
        profile_id: localBrowseProfile.value.id,
        path
      }
    })
    const data = rsp?.data?.data || {}
    localBrowseCurrentPath.value = data.current_path || ''
    localBrowseParentPath.value = data.parent_path || ''
    localBrowseRootPath.value = data.root_path || ''
    localBrowseDirectories.value = Array.isArray(data.directories) ? data.directories : []
    localBrowseFiles.value = Array.isArray(data.files) ? data.files : []
  } catch (error: any) {
    toast.error(error?.response?.data?.msg || '读取本地文件失败')
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
    toast.error('文件地址为空，无法预览')
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
const browseRootLabel = computed(() => (browseMode.value === 's3' ? '对象前缀：' : '根目录：'))
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

const browseThumbGridClass = computed(() => {
  if (browseThumbSize.value === 'sm') return 'grid grid-cols-3 md:grid-cols-5 lg:grid-cols-6 gap-2'
  if (browseThumbSize.value === 'lg') return 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
  return 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3'
})

const browseThumbPreviewClass = computed(() => {
  if (browseThumbSize.value === 'sm') return 'w-full h-20 rounded border bg-slate-50 dark:bg-slate-900 overflow-hidden flex items-center justify-center'
  if (browseThumbSize.value === 'lg') return 'w-full h-36 rounded border bg-slate-50 dark:bg-slate-900 overflow-hidden flex items-center justify-center'
  return 'w-full h-28 rounded border bg-slate-50 dark:bg-slate-900 overflow-hidden flex items-center justify-center'
})
const browseDialogWidthClass = computed(() => {
  if (browseViewMode.value === 'thumb') return '!w-[56vw] !max-w-[56vw] !sm:max-w-[792px]'
  return '!w-[55vw] !max-w-[55vw] !sm:max-w-[672px]'
})

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
      toast.error('未找到存储配置')
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
    const rsp = await request.post('/system/storage-config/delete-file', payload)
    if (rsp?.data?.code !== 200) {
      toast.error(rsp?.data?.msg || '删除失败')
      return
    }
    if (filePreviewPath.value === fileDeleteTargetObjectKey.value || filePreviewPath.value === fileDeleteTargetPath.value) {
      filePreviewDialogOpen.value = false
    }
    fileDeleteConfirmOpen.value = false
    toast.success('删除成功')
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
    toast.error('至少保留一个存储配置')
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
    toast.error(`${label}为空，无法复制`)
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
    toast.success(`${label}已复制`)
  } catch (error) {
    toast.error(`${label}复制失败`)
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

const localPathPreview = computed(() => `./data/${normalizedLocalSubPath.value}`)
const localPathWillCreate = computed(() => !hasInvalidLocalPathSegment.value && normalizedLocalSubPath.value.length > 0)
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
    const rsp = await request.get('/system/storage-config/local-directories', {
      params: { path }
    })
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
    toast.error('请输入目录名称')
    return
  }
  localDirCreating.value = true
  try {
    const rsp = await request.post('/system/storage-config/local-directories', {
      path: localDirCurrentPath.value || '',
      name: folderName
    })
    if (rsp?.data?.code === 200) {
      toast.success('目录创建成功')
      localDirNewFolderName.value = ''
      await loadLocalDirectories(localDirCurrentPath.value || '')
      return
    }
    toast.error(rsp?.data?.msg || '目录创建失败')
  } finally {
    localDirCreating.value = false
  }
}

const chooseCurrentDirectory = () => {
  if (!localDirCurrentPath.value) {
    toast.error('请先选择子目录')
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
    toast.error('请输入存储名称')
    return
  }
  if (editor.provider === 's3') {
    if (!editor.s3_endpoint.trim() || !editor.s3_bucket.trim() || !editor.s3_object_key_prefix.trim() || !editor.s3_access_key.trim() || !editor.s3_secret_key.trim()) {
      toast.error('Please fill Endpoint, Bucket, Object Prefix, Access Key, Secret Key')
      return
    }
  } else if (!editor.local_sub_path.trim()) {
    toast.error('请输入本地路径')
    return
  } else if (hasInvalidLocalPathSegment.value) {
    toast.error('本地路径不能包含 ..')
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
    <div class="flex items-center justify-between">
      <div class="text-sm text-muted-foreground">支持配置多个存储实例，并可设置默认存储供未指定模块兜底使用</div>
      <Button type="button" variant="outline" @click="openCreateDialog">新建存储</Button>
    </div>

    <div class="rounded border border-slate-300 dark:border-slate-600 overflow-x-auto">
      <div class="min-w-[1160px] divide-y">
        <div class="grid grid-cols-[140px_220px_100px_minmax(220px,1fr)_120px_280px] py-2 text-xs text-muted-foreground divide-x divide-slate-300/80 dark:divide-slate-600/80 bg-slate-50/60 dark:bg-slate-900/40 [&>div]:px-3">
          <div class="text-center">存储ID</div>
          <div class="text-center">存储名称</div>
          <div class="text-center">类型</div>
          <div class="text-center">路径</div>
          <div class="text-center">默认</div>
          <div class="text-center">操作</div>
        </div>
        <div
          v-for="profile in profiles"
          :key="profile.id"
          class="grid grid-cols-[140px_220px_100px_minmax(220px,1fr)_120px_280px] py-3 items-center divide-x divide-slate-300/70 dark:divide-slate-600/70 [&>div]:px-3"
        >
          <div class="min-w-0 flex items-center gap-2">
            <span class="text-xs text-muted-foreground font-mono truncate" :title="profile.id">{{ profile.id }}</span>
            <button
              type="button"
              class="inline-flex h-6 w-6 items-center justify-center rounded border border-transparent text-muted-foreground hover:text-brand-600 hover:border-brand-200"
              title="复制存储ID"
              aria-label="复制存储ID"
              @click="copyText(profile.id, '存储ID')"
            >
              <Copy class="h-3.5 w-3.5" />
            </button>
          </div>
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <div class="text-sm font-medium truncate" :title="profile.name">{{ profile.name }}</div>
              <button
                type="button"
                class="inline-flex h-6 w-6 items-center justify-center rounded border border-transparent text-muted-foreground hover:text-brand-600 hover:border-brand-200 shrink-0"
                title="复制存储名称"
                aria-label="复制存储名称"
                @click="copyText(profile.name, '存储名称')"
              >
                <Copy class="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
          <div class="text-sm">{{ profile.provider === 's3' ? 'S3' : '本地' }}</div>
          <div class="text-sm text-muted-foreground min-w-0">
            <span
              v-if="profile.provider === 's3'"
              class="inline-flex items-center gap-2 max-w-full"
            >
              <span
                class="block truncate"
                :title="`${profile.s3_bucket}/${profile.s3_object_key_prefix}`"
              >
                {{ profile.s3_bucket }}/{{ profile.s3_object_key_prefix }}
              </span>
              <button
                type="button"
                class="inline-flex h-6 w-6 items-center justify-center rounded border border-transparent text-muted-foreground hover:text-brand-600 hover:border-brand-200 shrink-0"
                title="复制路径前缀"
                aria-label="复制路径前缀"
                @click="copyText(`${profile.s3_bucket}/${profile.s3_object_key_prefix}`, '路径前缀')"
              >
                <Copy class="h-3.5 w-3.5" />
              </button>
            </span>
            <span v-else class="inline-flex items-center gap-2 max-w-full">
              <span class="block truncate" :title="`./data/${profile.local_sub_path || 'uploads'}`">./data/{{ profile.local_sub_path || 'uploads' }}</span>
              <button
                type="button"
                class="inline-flex h-6 w-6 items-center justify-center rounded border border-transparent text-muted-foreground hover:text-brand-600 hover:border-brand-200 shrink-0"
                title="复制路径前缀"
                aria-label="复制路径前缀"
                @click="copyText(`./data/${profile.local_sub_path || 'uploads'}`, '路径前缀')"
              >
                <Copy class="h-3.5 w-3.5" />
              </button>
            </span>
          </div>
          <div class="flex items-center justify-center">
            <button
              type="button"
              class="inline-flex items-center justify-center h-8 w-8 rounded-full border transition-colors"
              :class="defaultStorageID === profile.id
                ? 'border-emerald-300 bg-emerald-100 text-emerald-700 dark:border-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300 cursor-default pointer-events-none'
                : 'border-slate-300 bg-white text-slate-600 hover:border-emerald-300 hover:text-emerald-700 dark:border-slate-600 dark:bg-slate-900 dark:text-slate-300 dark:hover:border-emerald-700 dark:hover:text-emerald-300'"
              :title="defaultStorageID === profile.id ? '默认存储' : '设为默认存储'"
              :disabled="saving || loading"
              @click="setDefaultStorage(profile.id)"
            >
              <CheckCircle2 v-if="defaultStorageID === profile.id" class="w-3.5 h-3.5" />
              <Circle v-else class="w-3.5 h-3.5" />
            </button>
          </div>
          <div class="flex items-center justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="border-slate-200 bg-slate-50 text-slate-600 hover:bg-slate-100 hover:border-slate-300 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            :disabled="saving || loading"
            @click="handleBrowseClick(profile)"
          >
            浏览
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="border-slate-200 bg-slate-50 text-slate-600 hover:bg-slate-100 hover:border-slate-300 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            :disabled="localUploadTesting || saving || loading"
            @click="handleTestClick(profile)"
          >
            {{ localUploadTesting ? '上传中...' : '上传' }}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="border-slate-200 bg-slate-50 text-slate-600 hover:bg-slate-100 hover:border-slate-300 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            :disabled="saving || loading"
            @click="openEditDialog(profile)"
          >
            编辑
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="border-red-200 bg-red-50 text-red-600 hover:bg-red-100 hover:border-red-300"
            :disabled="saving || loading"
            @click="openDeleteConfirm(profile)"
          >
            删除
          </Button>
          </div>
        </div>
      </div>
    </div>

    <Dialog :open="editorOpen" @update:open="(value) => { editorOpen = value }">
      <DialogContent class="max-w-[760px]">
        <DialogHeader>
          <DialogTitle>{{ editingProfileId ? '编辑存储配置' : '新建存储配置' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <Input v-model="editor.name" placeholder="存储名称" />
          </div>
          <div class="flex items-center gap-3">
            <div class="text-sm whitespace-nowrap">上传文件名前缀</div>
            <Input v-model="editor.upload_file_prefix" class="w-[240px]" placeholder="默认 upload" />
          </div>
          <div class="space-y-2">
            <div class="text-sm">存储类型</div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="px-3 py-1.5 text-sm rounded border transition-colors"
                :class="editor.provider === 'local' ? 'border-brand-500 bg-brand-50 dark:bg-slate-800' : 'border-gray-200 dark:border-slate-700'"
                @click="editor.provider = 'local'"
              >
                本地
              </button>
              <button
                type="button"
                class="px-3 py-1.5 text-sm rounded border transition-colors"
                :class="editor.provider === 's3' ? 'border-brand-500 bg-brand-50 dark:bg-slate-800' : 'border-gray-200 dark:border-slate-700'"
                @click="editor.provider = 's3'"
              >
                S3
              </button>
            </div>
          </div>

          <div v-if="editor.provider === 'local'" class="space-y-3">
            <div class="space-y-2">
              <div class="text-sm">本地路径</div>
              <div class="flex items-center gap-2">
                <div class="px-3 py-2 text-sm rounded border bg-slate-50 text-slate-600 dark:bg-slate-900 dark:text-slate-300">./data/</div>
                <Input v-model="editor.local_sub_path" placeholder="例如 uploads/oidc" />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  title="浏览目录"
                  class="border-brand-200 bg-brand-50 text-brand-600 hover:bg-brand-100 hover:border-brand-300 dark:border-brand-700/60 dark:bg-brand-900/30 dark:text-brand-300 dark:hover:bg-brand-900/40"
                  @click="openLocalDirDialog"
                >
                  <FolderOpen class="w-4 h-4" />
                </Button>
              </div>
              <div class="flex items-center gap-2">
                <div class="text-xs text-muted-foreground">最终目录：{{ localPathPreview }}</div>
                <span
                  v-if="localPathWillCreate"
                  class="px-1.5 py-0.5 text-[10px] rounded border border-emerald-200 bg-emerald-50 text-emerald-600"
                >
                  目录将自动创建
                </span>
              </div>
              <div v-if="hasInvalidLocalPathSegment" class="text-xs text-red-500">路径不能包含 ..，保存时会拦截</div>
            </div>
          </div>

          <div v-if="editor.provider === 's3'" class="space-y-3">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
              <Input v-model="editor.s3_endpoint" placeholder="Endpoint" />
              <Input v-model="editor.s3_region" placeholder="Region" />
              <Input v-model="editor.s3_bucket" placeholder="Bucket" />
              <Input v-model="editor.s3_object_key_prefix" placeholder="Object Prefix" />
              <Input v-model="editor.s3_access_key" placeholder="Access Key" />
              <Input v-model="editor.s3_secret_key" placeholder="Secret Key" />
              <Input v-model="editor.s3_public_base_url" placeholder="Public URL (Optional)" />
            </div>
            <div class="space-y-2">
              <div class="text-sm">Use HTTPS</div>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  class="px-3 py-1.5 text-sm rounded border transition-colors"
                  :class="editor.s3_use_ssl ? 'border-brand-500 bg-brand-50 dark:bg-slate-800' : 'border-gray-200 dark:border-slate-700'"
                  @click="editor.s3_use_ssl = true"
                >
                  开启
                </button>
                <button
                  type="button"
                  class="px-3 py-1.5 text-sm rounded border transition-colors"
                  :class="!editor.s3_use_ssl ? 'border-brand-500 bg-brand-50 dark:bg-slate-800' : 'border-gray-200 dark:border-slate-700'"
                  @click="editor.s3_use_ssl = false"
                >
                  关闭
                </button>
              </div>
            </div>
            <div class="space-y-2">
              <div class="text-sm">Proxy Public Read via Backend</div>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  class="px-3 py-1.5 text-sm rounded border transition-colors"
                  :class="editor.s3_proxy_public_read ? 'border-brand-500 bg-brand-50 dark:bg-slate-800' : 'border-gray-200 dark:border-slate-700'"
                  @click="editor.s3_proxy_public_read = true"
                >
                  开启
                </button>
                <button
                  type="button"
                  class="px-3 py-1.5 text-sm rounded border transition-colors"
                  :class="!editor.s3_proxy_public_read ? 'border-brand-500 bg-brand-50 dark:bg-slate-800' : 'border-gray-200 dark:border-slate-700'"
                  @click="editor.s3_proxy_public_read = false"
                >
                  关闭
                </button>
              </div>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="editorOpen = false">取消</Button>
          <Button type="button" :disabled="saving" @click="applyEditor">{{ saving ? '保存中...' : '保存并生效' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="deleteConfirmOpen" @update:open="(value) => value ? (deleteConfirmOpen = true) : closeDeleteConfirm()">
      <DialogContent class="w-[420px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>确认删除存储配置</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm text-gray-600">
            请输入要删除的存储名称
            <span v-if="deleteTarget?.name" class="text-red-500 font-semibold mx-1">{{ deleteTarget.name }}</span>
            以确认操作
          </div>
          <Input
            v-model="deleteConfirmInput"
            :max-length="100"
            placeholder="请输入存储名称"
            class="confirm-delete-input"
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
      <DialogContent :class="browseDialogWidthClass">
        <DialogHeader>
          <DialogTitle>{{ browseDialogTitle }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <div class="text-sm text-muted-foreground">
            当前存储：{{ browseProfileName }}（{{ browseMode === 's3' ? 'S3' : '本地' }} / {{ browseProfileId }}）
          </div>
          <div class="grid grid-cols-[auto_minmax(0,1fr)] items-center gap-2 text-sm text-muted-foreground">
            <span>{{ browseRootLabel }}</span>
            <div class="min-w-0 overflow-x-auto whitespace-nowrap pb-1">
              <button
                v-for="(crumb, index) in browseBreadcrumbs"
                :key="`${crumb.path || 'browse-root'}-${index}`"
                type="button"
                class="inline-flex items-center gap-1 hover:text-brand-600 mr-1"
                :disabled="browseLoading"
                @click="openBrowseBreadcrumb(crumb.path)"
              >
                <ChevronRight v-if="index > 0" class="w-3.5 h-3.5" />
                <Folder v-if="index === 0" class="w-3.5 h-3.5" />
                <span class="max-w-[240px] truncate align-bottom">{{ crumb.label }}</span>
              </button>
            </div>
          </div>
          <div class="grid grid-cols-[auto_minmax(240px,1fr)_auto_auto] items-center gap-2">
            <Button type="button" variant="outline" size="sm" :disabled="!browseCurrentPath || browseLoading" @click="openBrowseParent">
              返回上级
            </Button>
            <Input v-model="browseKeyword" class="h-8 w-full min-w-[240px]" placeholder="按名称筛选文件/目录" />
            <div class="inline-flex items-center gap-1 rounded border p-1">
              <button
                type="button"
                class="px-2 py-1 text-xs rounded"
                :class="browseViewMode === 'list' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
                @click="browseViewMode = 'list'"
              >
                列表
              </button>
              <button
                type="button"
                class="px-2 py-1 text-xs rounded"
                :class="browseViewMode === 'thumb' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
                @click="browseViewMode = 'thumb'"
              >
                缩略图
              </button>
            </div>
            <div
              class="inline-flex w-[132px] items-center gap-1 rounded border p-1 transition-opacity"
              :class="browseViewMode === 'thumb' ? 'opacity-100' : 'opacity-0 pointer-events-none'"
            >
              <button
                type="button"
                class="px-2 py-1 text-xs rounded"
                :class="browseThumbSize === 'sm' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
                @click="browseThumbSize = 'sm'"
              >
                小
              </button>
              <button
                type="button"
                class="px-2 py-1 text-xs rounded"
                :class="browseThumbSize === 'md' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
                @click="browseThumbSize = 'md'"
              >
                中
              </button>
              <button
                type="button"
                class="px-2 py-1 text-xs rounded"
                :class="browseThumbSize === 'lg' ? 'bg-brand-50 text-brand-600 dark:bg-slate-800 dark:text-brand-300' : 'text-muted-foreground'"
                @click="browseThumbSize = 'lg'"
              >
                大
              </button>
            </div>
          </div>
          <div v-if="browseViewMode === 'list'" class="rounded border max-h-[360px] overflow-y-auto divide-y">
            <div
              v-for="item in filteredBrowseDirectories"
              :key="item.relative_path"
              class="flex items-center gap-3 px-3 py-2 hover:bg-slate-50/80 dark:hover:bg-slate-900/40 cursor-pointer"
              @click="openBrowseChild(item)"
            >
              <div class="inline-flex h-7 w-7 items-center justify-center rounded border border-brand-200 bg-brand-50 text-brand-600 dark:border-brand-700/60 dark:bg-brand-900/30 dark:text-brand-300 shrink-0">
                <Folder class="w-4 h-4" />
              </div>
              <div class="text-sm min-w-0">
                <div class="truncate font-medium">{{ item.name }}</div>
                <div class="text-[11px] text-muted-foreground truncate">{{ item.relative_path }}</div>
              </div>
            </div>
            <div
              v-for="file in filteredBrowseFiles"
              :key="file.object_key || file.relative_path"
              class="flex items-center justify-between gap-3 px-3 py-2 hover:bg-slate-50/80 dark:hover:bg-slate-900/40"
            >
              <div class="min-w-0 text-sm">
                <div class="truncate font-medium">{{ file.name }}</div>
                <div
                  class="text-[11px] text-muted-foreground truncate"
                  :title="browseMode === 's3' ? (file.object_key || file.relative_path) : file.relative_path"
                >
                  {{ browseMode === 's3' ? (file.object_key || file.relative_path) : file.relative_path }}
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <span v-if="typeof file.size === 'number'" class="text-[11px] text-muted-foreground">{{ formatFileSize(file.size || 0) }}</span>
                <Button type="button" variant="outline" size="sm" :disabled="browseMode === 's3' && !file.public_url" @click="previewBrowseFile(file)">预览</Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  @click="copyText(file.public_url || file.object_key || file.relative_path, file.public_url ? '文件链接' : (browseMode === 's3' ? '对象键' : '相对路径'))"
                >
                  复制
                </Button>
                <Button type="button" variant="outline" size="sm" class="text-red-600 border-red-200 hover:bg-red-50" @click="askDeleteBrowseFile(file)">删除</Button>
              </div>
            </div>
            <div v-if="!browseLoading && filteredBrowseDirectories.length === 0 && filteredBrowseFiles.length === 0" class="px-3 py-6 text-sm text-muted-foreground text-center">
              当前路径下暂无内容
            </div>
          </div>
          <div v-else class="rounded border max-h-[360px] overflow-y-auto p-3 space-y-3">
            <div v-if="filteredBrowseDirectories.length > 0" class="grid grid-cols-2 md:grid-cols-3 gap-2">
              <button
                v-for="item in filteredBrowseDirectories"
                :key="`thumb-dir-${item.relative_path}`"
                type="button"
                class="rounded border px-3 py-2 text-left hover:bg-slate-50/80 dark:hover:bg-slate-900/40"
                @click="openBrowseChild(item)"
              >
                <div class="inline-flex h-8 w-8 items-center justify-center rounded border border-brand-200 bg-brand-50 text-brand-600 dark:border-brand-700/60 dark:bg-brand-900/30 dark:text-brand-300">
                  <Folder class="w-4 h-4" />
                </div>
                <div class="mt-1 text-sm truncate font-medium">{{ item.name }}</div>
              </button>
            </div>
            <div :class="browseThumbGridClass">
              <div
                v-for="file in browseThumbFiles"
                :key="`thumb-file-${file.object_key || file.relative_path}`"
                class="rounded border p-2 space-y-2"
              >
                <button
                  type="button"
                  :class="browseThumbPreviewClass"
                  :disabled="!file.can_preview"
                  @click="previewBrowseFile(file)"
                >
                  <img
                    v-if="file.is_image && file.can_preview"
                    :src="file.preview_url"
                    class="w-full h-full object-cover"
                    :alt="file.name"
                  >
                  <FolderOpen v-else class="w-7 h-7 text-slate-400" />
                </button>
                <div class="text-xs">
                  <div class="truncate font-medium" :title="file.name">{{ file.name }}</div>
                  <div class="text-muted-foreground truncate" :title="file.label">{{ file.label }}</div>
                </div>
                <div class="flex items-center justify-between gap-1">
                  <Button type="button" variant="outline" size="sm" class="h-7 px-2" :disabled="browseMode === 's3' && !file.public_url" @click="previewBrowseFile(file)">预览</Button>
                  <Button type="button" variant="outline" size="sm" class="h-7 px-2" @click="copyText(file.public_url || file.object_key || file.relative_path, file.public_url ? '文件链接' : (browseMode === 's3' ? '对象键' : '相对路径'))">复制</Button>
                  <Button type="button" variant="outline" size="sm" class="h-7 px-2 text-red-600 border-red-200 hover:bg-red-50" @click="askDeleteBrowseFile(file)">删除</Button>
                </div>
              </div>
            </div>
            <div v-if="!browseLoading && filteredBrowseDirectories.length === 0 && browseThumbFiles.length === 0" class="px-3 py-6 text-sm text-muted-foreground text-center">
              当前路径下暂无内容
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="browseDialogOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="filePreviewDialogOpen" @update:open="(value) => { filePreviewDialogOpen = value }">
      <DialogContent class="!w-[92vw] !max-w-[600px]">
        <DialogHeader>
          <DialogTitle>文件预览</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <div class="text-sm font-medium truncate" :title="filePreviewName">{{ filePreviewName }}</div>
          <div class="text-xs text-muted-foreground truncate" :title="filePreviewPath">{{ filePreviewPath }}</div>
          <div class="rounded border bg-slate-50 dark:bg-slate-900 h-[68vh] overflow-hidden">
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
        <DialogFooter class="flex-row items-center justify-end gap-2">
          <Button type="button" variant="outline" :disabled="!filePreviewHasPrev" @click="movePreviewBy(-1)">上一张</Button>
          <Button type="button" variant="outline" :disabled="!filePreviewHasNext" @click="movePreviewBy(1)">下一张</Button>
          <span v-if="filePreviewProgressText" class="text-xs text-muted-foreground px-1 min-w-[44px] text-center whitespace-nowrap">{{ filePreviewProgressText }}</span>
          <Button type="button" variant="outline" @click="copyText(filePreviewUrl, '预览地址')">复制地址</Button>
          <Button type="button" variant="outline" @click="openPreviewInNewTab">新标签打开</Button>
          <Button type="button" @click="filePreviewDialogOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="fileDeleteConfirmOpen" @update:open="(value) => { fileDeleteConfirmOpen = value }">
      <DialogContent class="max-w-[480px]">
        <DialogHeader>
          <DialogTitle>确认删除文件</DialogTitle>
        </DialogHeader>
        <div class="space-y-2 text-sm">
          <div class="text-slate-700 dark:text-slate-200">该操作不可恢复，确认要删除以下文件吗？</div>
          <div class="rounded border bg-slate-50 dark:bg-slate-900 px-3 py-2">
            <div class="font-medium truncate" :title="fileDeleteTargetName">{{ fileDeleteTargetName }}</div>
            <div class="text-xs text-muted-foreground truncate" :title="fileDeleteTargetObjectKey || fileDeleteTargetPath">{{ fileDeleteTargetObjectKey || fileDeleteTargetPath }}</div>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" :disabled="fileDeleting" @click="fileDeleteConfirmOpen = false">取消</Button>
          <Button type="button" :disabled="fileDeleting" class="bg-red-600 hover:bg-red-700 text-white" @click="confirmDeleteBrowseFile">
            {{ fileDeleting ? '删除中...' : '确认删除' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="localUploadTestOpen" @update:open="(value) => value ? (localUploadTestOpen = true) : closeLocalUploadTest()">
      <DialogContent class="max-w-[520px]">
        <DialogHeader>
          <DialogTitle>存储上传</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <div class="text-sm text-muted-foreground">
            当前存储：{{ localUploadTestProfile?.name }}（{{ localUploadTestProfile?.provider === 's3' ? 'S3' : '本地' }} / {{ localUploadTestProfile?.id }}）
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 text-sm text-slate-700 dark:text-slate-200"
              :title="localUploadDeleteAfter ? '上传成功后将立即删除该文件' : '上传成功后保留该文件'"
              @click="localUploadDeleteAfter = !localUploadDeleteAfter"
            >
              <CheckCircle2 v-if="localUploadDeleteAfter" class="w-4 h-4 text-brand-600" />
              <Circle v-else class="w-4 h-4 text-slate-400" />
              <span>上传后立即删除</span>
            </button>
          </div>
          <input
            ref="localUploadFileInputRef"
            type="file"
            class="w-full text-sm file:mr-3 file:px-3 file:py-1.5 file:rounded file:border file:bg-slate-50 file:text-slate-700"
            @change="handleLocalUploadFileChange"
          />
          <div v-if="localUploadTestFile" class="flex items-center justify-between gap-2">
            <div class="text-xs text-muted-foreground">
              已选文件：{{ localUploadSelectedFileName }}（{{ localUploadSelectedFileSize }}）
            </div>
            <Button type="button" variant="outline" size="sm" :disabled="localUploadTesting" @click="clearLocalUploadFile">
              重新选择
            </Button>
          </div>
          <div v-if="localUploadLastLocation" class="flex items-start gap-2 min-w-0">
            <div class="text-xs text-muted-foreground min-w-0 flex-1 break-all leading-5" :title="localUploadLastLocation">上传地址：{{ localUploadLastLocation }}</div>
            <Button type="button" variant="outline" size="sm" class="shrink-0" :disabled="localUploadTesting" @click="copyText(localUploadLastLocation, '上传地址')">
              复制地址
            </Button>
          </div>
          <div class="text-xs text-muted-foreground">可在此手动上传文件，上传成功表示当前存储可正常写入</div>
        </div>
        <DialogFooter>
          <span v-if="localUploadAutoCloseCountdown > 0" class="mr-auto text-xs text-muted-foreground">
            上传成功，{{ localUploadAutoCloseCountdown }} 秒后自动关闭
          </span>
          <Button type="button" variant="outline" :disabled="localUploadTesting" @click="closeLocalUploadTest">取消</Button>
          <Button type="button" :disabled="localUploadTesting" @click="submitLocalUploadTest">
            {{ localUploadTesting ? '上传中...' : '上传' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="localDirDialogOpen" @update:open="(value) => { localDirDialogOpen = value }">
      <DialogContent class="max-w-[560px]" tabindex="0" @keydown="handleLocalDirDialogKeydown">
        <DialogHeader>
          <DialogTitle>选择本地目录</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <div class="flex flex-wrap items-center gap-1 text-sm text-muted-foreground">
            <span>当前目录：</span>
            <button
              v-for="(crumb, index) in localDirBreadcrumbs"
              :key="crumb.path || 'data-root'"
              type="button"
              class="inline-flex items-center gap-1 hover:text-brand-600"
              :disabled="localDirLoading"
              @click="openBreadcrumbDirectory(crumb.path)"
            >
              <ChevronRight v-if="index > 0" class="w-3.5 h-3.5" />
              <Folder v-if="index === 0" class="w-3.5 h-3.5" />
              <span>{{ crumb.label }}</span>
            </button>
            <span
              v-if="isCurrentDirSelected"
              class="ml-1 inline-flex items-center gap-1 px-1.5 py-0.5 text-[10px] rounded border border-blue-200 bg-blue-50 text-brand-600"
            >
              <CheckCircle2 class="w-3 h-3" />
              已选中
            </span>
          </div>
          <div class="flex items-center gap-2">
            <Button type="button" variant="outline" size="sm" :disabled="!localDirCurrentPath || localDirLoading" @click="openParentDirectory">
              返回上级
            </Button>
            <Button type="button" size="sm" :disabled="localDirLoading || !localDirCurrentPath || isCurrentDirSelected" @click="chooseCurrentDirectory">
              {{ isCurrentDirSelected ? '已选择当前目录' : '选择当前目录' }}
            </Button>
          </div>
          <div class="flex items-center gap-2">
            <Input
              v-model="localDirNewFolderName"
              placeholder="新建子目录名称"
              :disabled="localDirLoading || localDirCreating"
              @keyup.enter="createLocalDirectory"
            />
            <Button type="button" variant="outline" :disabled="localDirLoading || localDirCreating" @click="createLocalDirectory">
              {{ localDirCreating ? '创建中...' : '新建目录' }}
            </Button>
          </div>
          <div class="text-[11px] text-muted-foreground">快捷键：在“新建子目录名称”输入框按 Enter 创建目录，按 Ctrl + Enter 选择当前目录</div>
          <div class="rounded border max-h-[320px] overflow-y-auto divide-y">
            <div
              v-for="item in localDirItems"
              :key="item.relative_path"
              class="flex items-center gap-3 px-3 py-2 hover:bg-slate-50/80 dark:hover:bg-slate-900/40 cursor-pointer"
              :class="item.relative_path === selectedLocalDirPath ? 'bg-brand-50/70 dark:bg-brand-900/20' : ''"
              @click="openChildDirectory(item)"
            >
              <div class="flex items-center gap-2 min-w-0 text-sm hover:text-brand-600 text-left">
                <div class="inline-flex h-7 w-7 items-center justify-center rounded border border-brand-200 bg-brand-50 text-brand-600 dark:border-brand-700/60 dark:bg-brand-900/30 dark:text-brand-300 shrink-0">
                  <Folder class="w-4 h-4" />
                </div>
                <div class="min-w-0">
                  <div class="truncate font-medium">{{ item.name }}</div>
                  <div class="text-[11px] text-muted-foreground truncate">./data/{{ item.relative_path }}</div>
                </div>
              </div>
            </div>
            <div v-if="!localDirLoading && localDirItems.length === 0" class="px-3 py-6 text-sm text-muted-foreground text-center">
              当前目录下没有子目录
            </div>
            <div v-if="localDirLoading" class="px-3 py-6 text-sm text-muted-foreground text-center">读取中...</div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
