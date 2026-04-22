import { defineStore } from 'pinia'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'
import { clearAuthzDataStorage, readAuthzDataFromStorage, writeAuthzDataToStorage, type RbacAuthzData } from '@/util/rbacAuthz'

export const useRbacAuthzStore = defineStore('rbacAuthz', {
  state: () => {
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
    hydrateFromStorage() {
      const authz = readAuthzDataFromStorage()
      this.userId = authz.user_id || 0
      this.username = authz.username || ''
      this.roles = authz.roles || []
      this.groups = authz.groups || []
      this.permissions = authz.permissions || []
      this.isSuperAdmin = Boolean(authz.is_super_admin)
      this.loaded = this.permissions.length > 0
    },
    setAuthzData(data: RbacAuthzData) {
      this.userId = data.user_id || 0
      this.username = data.username || ''
      this.roles = data.roles || []
      this.groups = data.groups || []
      this.permissions = data.permissions || []
      this.isSuperAdmin = Boolean(data.is_super_admin)
      this.loaded = true
      writeAuthzDataToStorage({
        user_id: this.userId,
        username: this.username,
        roles: this.roles,
        groups: this.groups,
        permissions: this.permissions,
        is_super_admin: this.isSuperAdmin
      })
    },
    clearAuthzData() {
      this.userId = 0
      this.username = ''
      this.roles = []
      this.groups = []
      this.permissions = []
      this.isSuperAdmin = false
      this.loaded = false
      clearAuthzDataStorage()
    },
    hasPermission(code: string) {
      if (!code) return true
      return this.permissions.includes(code)
    },
    hasAnyPermission(codes: string[] = []) {
      if (!Array.isArray(codes) || codes.length === 0) return true
      const permissionSet = new Set(this.permissions)
      return codes.some(code => permissionSet.has(code))
    },
    async fetchCurrentUserPermissions(options: { silent?: boolean } = {}) {
      try {
        const rsp = await request.get('/rbac/me/permissions')
        if (rsp?.status === 200 && rsp?.data?.code === 200 && rsp?.data?.data) {
          this.setAuthzData(rsp.data.data)
          return true
        }
      } catch (error) {
        if (!options.silent) {
          toast.error('获取权限信息失败')
        }
      }
      return false
    }
  }
})

