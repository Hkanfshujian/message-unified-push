import { defineStore } from 'pinia'
import { request } from '@/api/client'
import { clearAuthzDataStorage, readAuthzDataFromStorage, writeAuthzDataToStorage } from '@/util/rbacAuthz'
import type { RbacPermissionState } from '@/types/app'
import type { ApiResponse } from '@/types/common'

type RbacStoreState = RbacPermissionState

export const useRbacStore = defineStore('rbac', {
  state: (): RbacStoreState => {
    const authz = readAuthzDataFromStorage()
    return {
      userId: authz.user_id || 0,
      username: authz.username || '',
      roles: authz.roles || [],
      groups: authz.groups || [],
      permissions: authz.permissions || [],
      isSuperAdmin: Boolean(authz.is_super_admin),
      loaded: (authz.permissions || []).length > 0
    }
  },
  actions: {
    hasPermission(this: RbacStoreState, code?: string) {
      if (!code || this.isSuperAdmin) return true
      return this.permissions.includes(code)
    },
    hasAnyPermission(this: RbacStoreState, codes: string[] = []) {
      if (this.isSuperAdmin || codes.length === 0) return true
      const permissionSet = new Set(this.permissions)
      return codes.some(code => permissionSet.has(code))
    },
    clear(this: RbacStoreState) {
      this.userId = 0
      this.username = ''
      this.roles = []
      this.groups = []
      this.permissions = []
      this.isSuperAdmin = false
      this.loaded = false
      clearAuthzDataStorage()
    },
    async loadCurrentUserPermissions(this: RbacStoreState) {
      const response = await request.get<ApiResponse<RbacPermissionState & { user_id?: number; is_super_admin?: boolean }>>('/rbac/me/permissions')
      const data = response.data.data
      this.userId = data.user_id || data.userId || 0
      this.username = data.username || ''
      this.roles = data.roles || []
      this.groups = data.groups || []
      this.permissions = data.permissions || []
      this.isSuperAdmin = Boolean(data.is_super_admin || data.isSuperAdmin)
      this.loaded = true
      writeAuthzDataToStorage({
        user_id: this.userId,
        username: this.username,
        roles: this.roles,
        groups: this.groups,
        permissions: this.permissions,
        is_super_admin: this.isSuperAdmin
      })
      return true
    }
  }
})
