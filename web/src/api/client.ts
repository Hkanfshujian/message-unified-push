import axios, { type AxiosError, type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import config from '../../config.js'
import { CONSTANT } from '@/constant'
import { clearAuthzDataStorage } from '@/util/rbacAuthz'
import type { ApiResponse, RequestMeta } from '@/types/common'

declare module 'axios' {
  interface InternalAxiosRequestConfig {
    meta?: RequestMeta
  }

  interface AxiosRequestConfig {
    meta?: RequestMeta
  }
}

const tokenErrorCodes = [20001, 20002, 20003, 20004, 20005]

const normalizePathPrefix = (value: unknown) => {
  const prefix = String(value || '').trim()
  if (!prefix) return ''
  if (!prefix.startsWith('/') || prefix.startsWith('//') || prefix.includes('\\')) {
    throw new Error('Invalid API path prefix')
  }
  return prefix.replace(/\/+$/, '')
}

const getAllowedApiOrigins = () => {
  const origins = new Set([window.location.origin])
  const configuredOrigins = String(import.meta.env.VITE_API_ALLOWED_ORIGINS || '').split(',')
  for (const value of configuredOrigins) {
    const candidate = value.trim()
    if (!candidate) continue
    try {
      origins.add(new URL(candidate).origin)
    } catch {
      // Invalid allowlist entries are ignored instead of weakening the boundary.
    }
  }
  return origins
}

const resolveApiBaseURL = () => {
  const prefix = normalizePathPrefix(config.pathPrefix)
  const configuredBase = String(config.apiUrl || '').trim()
  if (!configuredBase) return prefix || '/'

  const parsed = new URL(configuredBase, window.location.origin)
  if (!['http:', 'https:'].includes(parsed.protocol) || parsed.username || parsed.password) {
    throw new Error('Invalid API base URL')
  }

  const isDevelopmentLoopback = import.meta.env.DEV && ['127.0.0.1', 'localhost', '::1'].includes(parsed.hostname)
  if (!getAllowedApiOrigins().has(parsed.origin) && !isDevelopmentLoopback) {
    throw new Error(`API origin is not allowed: ${parsed.origin}`)
  }

  const basePath = parsed.pathname.replace(/\/+$/, '')
  parsed.pathname = `${basePath}${prefix}` || '/'
  parsed.search = ''
  parsed.hash = ''
  return parsed.toString().replace(/\/$/, '')
}

const isNoAuthUrl = (url = '') => CONSTANT.NO_AUTH_URL.some((item: string) => url === item || url.startsWith(`${item}?`))

export const request: AxiosInstance = axios.create({
  baseURL: resolveApiBaseURL(),
  timeout: 50000,
  withCredentials: true
})

const clearClientAuthState = () => {
  localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
  localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME)
  localStorage.removeItem(CONSTANT.STORE_OPEN_TABS_NAME || '__message_nest_open_tabs_v1')
  clearAuthzDataStorage()
}

let isLoggingOut = false

export const logout = async (reason = '', redirect = '') => {
  if (isLoggingOut) return
  isLoggingOut = true

  const authSource = localStorage.getItem(CONSTANT.STORE_AUTH_SOURCE_NAME)

  if (authSource === 'casdoor') {
    try {
      const redirectUri = encodeURIComponent(`${window.location.origin}/#/login?casdoor_logout=1`)
      const response = await request.post<ApiResponse<{ logout_url?: string }>>(`/auth/casdoor/logout?redirect_uri=${redirectUri}`)
      const logoutUrl = response.data?.data?.logout_url
      if (logoutUrl) {
        clearClientAuthState()
        window.location.href = logoutUrl
        return
      }
    } catch {
      clearClientAuthState()
    }
  }

  clearClientAuthState()
  const query: Record<string, string> = {}
  if (reason) query.reason = reason
  if (redirect.startsWith('/') && !redirect.startsWith('//') && redirect !== '/login') query.redirect = redirect
  await router.push({ path: '/login', query })
  isLoggingOut = false
}

export const handleException = (error: AxiosError<ApiResponse>) => {
	if (!error.response) {
		ElMessage.error('网络请求失败，请确认后端服务已启动')
		return
	}

	const message = error.response.data?.msg || '请求处理失败，请稍后重试'
  ElMessage.error(message)
}

request.interceptors.request.use((requestConfig: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || ''
  const originalUrl = requestConfig.url || ''

  if (!isNoAuthUrl(originalUrl) && !originalUrl.startsWith('/api/')) {
    requestConfig.url = `/api/v1${originalUrl}`
  }

  if (token && token.trim() !== '' && !isNoAuthUrl(originalUrl)) {
    requestConfig.headers.set('m-token', token)
  }

  return requestConfig
})

request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    if (response.data && response.data.code !== 200) {
      if (tokenErrorCodes.includes(response.data.code)) {
        void logout('expired', router.currentRoute.value.fullPath)
        return Promise.reject(response)
      }

      if (!response.config?.meta?.silentBizToast) {
        ElMessage.error(response.data.msg || '接口逻辑错误')
      }
    }

    return response
  },
  (error: AxiosError<ApiResponse>) => {
    if (error.response?.status === 401 || (error.response?.data?.code && tokenErrorCodes.includes(error.response.data.code))) {
      void logout('expired', router.currentRoute.value.fullPath)
    } else if (!error.config?.meta?.silentErrorToast) {
      handleException(error)
    }

    return Promise.reject(error)
  }
)
