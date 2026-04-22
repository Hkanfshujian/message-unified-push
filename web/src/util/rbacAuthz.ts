import { CONSTANT } from '@/constant'

export interface RbacAuthzData {
  user_id?: number
  username?: string
  roles: string[]
  groups: string[]
  permissions: string[]
  is_super_admin?: boolean
}

const emptyAuthzData: RbacAuthzData = {
  roles: [],
  groups: [],
  permissions: []
}

export const readAuthzDataFromStorage = (): RbacAuthzData => {
  try {
    const raw = localStorage.getItem(CONSTANT.STORE_RBAC_AUTHZ_NAME)
    if (!raw) return { ...emptyAuthzData }
    const parsed = JSON.parse(raw)
    return {
      ...emptyAuthzData,
      ...parsed,
      roles: Array.isArray(parsed.roles) ? parsed.roles : [],
      groups: Array.isArray(parsed.groups) ? parsed.groups : [],
      permissions: Array.isArray(parsed.permissions) ? parsed.permissions : []
    }
  } catch {
    return { ...emptyAuthzData }
  }
}

export const writeAuthzDataToStorage = (data: RbacAuthzData) => {
  localStorage.setItem(CONSTANT.STORE_RBAC_AUTHZ_NAME, JSON.stringify({
    ...emptyAuthzData,
    ...data,
    roles: Array.isArray(data.roles) ? data.roles : [],
    groups: Array.isArray(data.groups) ? data.groups : [],
    permissions: Array.isArray(data.permissions) ? data.permissions : []
  }))
}

export const clearAuthzDataStorage = () => {
  localStorage.removeItem(CONSTANT.STORE_RBAC_AUTHZ_NAME)
}

export const hasPermissionFromStorage = (code: string) => {
  if (!code) return true
  const authz = readAuthzDataFromStorage()
  return authz.permissions.includes(code)
}

export const hasAnyPermissionFromStorage = (codes: string[] = []) => {
  if (!Array.isArray(codes) || codes.length === 0) return true
  const authz = readAuthzDataFromStorage()
  const permissionSet = new Set(authz.permissions)
  return codes.some(code => permissionSet.has(code))
}

