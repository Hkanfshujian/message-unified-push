import { ElLoading, ElMessage, ElMessageBox, type LoadingOptions } from 'element-plus'

export const notifySuccess = (message: string) => ElMessage.success(message)
export const notifyError = (message: string) => ElMessage.error(message)
export const notifyWarning = (message: string) => ElMessage.warning(message)
export const notifyInfo = (message: string) => ElMessage.info(message)

export const confirmAction = async (message: string, title = '确认操作') => {
  await ElMessageBox.confirm(message, title, {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning'
  })
}

export const showLoading = (options: LoadingOptions = {}) => ElLoading.service({
  lock: true,
  text: '加载中...',
  background: 'rgba(255, 255, 255, 0.7)',
  ...options
})

export const downloadBlob = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export const getEmptyText = (loading: boolean, text = '当前没有可展示的数据') => loading ? '加载中...' : text
