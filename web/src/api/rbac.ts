import { request } from '@/api/api'
import type { PageQuery } from '@/types/common'

type IdPayload = { id?: number; role_id?: number; group_id?: number; user_id?: number }
type AssignmentPayload = Record<string, number | number[] | string | string[] | undefined>
type RbacPayload = Record<string, unknown>

export const rbacApi = {
  getRoles: (params: PageQuery) => request.get('/rbac/roles', { params }),
  addRole: (data: RbacPayload) => request.post('/rbac/roles', data),
  editRole: (data: RbacPayload) => request.post('/rbac/roles/edit', data),
  deleteRole: (data: IdPayload) => request.post('/rbac/roles/delete', data),
  getRolePermissionIDs: (roleId: number) => request.get('/rbac/roles/permissions', { params: { role_id: roleId } }),
  assignRolePermissions: (data: AssignmentPayload) => request.post('/rbac/roles/assign-permissions', data),

  getGroups: (params: PageQuery) => request.get('/rbac/groups', { params }),
  addGroup: (data: RbacPayload) => request.post('/rbac/groups', data),
  editGroup: (data: RbacPayload) => request.post('/rbac/groups/edit', data),
  deleteGroup: (data: IdPayload) => request.post('/rbac/groups/delete', data),
  getGroupRoleIDs: (groupId: number) => request.get('/rbac/groups/roles', { params: { group_id: groupId } }),
  getGroupMemberIDs: (groupId: number) => request.get('/rbac/groups/members', { params: { group_id: groupId } }),
  assignGroupRoles: (data: AssignmentPayload) => request.post('/rbac/groups/assign-roles', data),
  assignGroupMembers: (data: AssignmentPayload) => request.post('/rbac/groups/assign-members', data),

  getPermissions: (params: PageQuery) => request.get('/rbac/permissions', { params }),
  addPermission: (data: RbacPayload) => request.post('/rbac/permissions', data),
  editPermission: (data: RbacPayload) => request.post('/rbac/permissions/edit', data),

  getUsers: (params: PageQuery) => request.get('/rbac/users', { params }),
  getManageUsers: (params: PageQuery) => request.get('/rbac/users/manage', { params }),
  addManageUser: (data: RbacPayload) => request.post('/rbac/users/manage', data),
  editManageUser: (data: RbacPayload) => request.post('/rbac/users/manage/edit', data),
  deleteManageUser: (data: IdPayload) => request.post('/rbac/users/manage/delete', data),
  getUserRoleIDs: (userId: number) => request.get('/rbac/users/role-ids', { params: { user_id: userId } }),
  getUserGroupIDs: (userId: number) => request.get('/rbac/users/group-ids', { params: { user_id: userId } }),
  assignUserRoles: (data: AssignmentPayload) => request.post('/rbac/users/assign-roles', data),
  assignUserGroups: (data: AssignmentPayload) => request.post('/rbac/users/assign-groups', data),

  getOIDCMetrics: (params: PageQuery) => request.get('/oidc/metrics', { params }),
  getOIDCAudits: (params: PageQuery) => request.get('/oidc/audits', { params }),
  getOIDCConflicts: (params: PageQuery) => request.get('/oidc/conflicts', { params }),
  approveOIDCConflict: (data: RbacPayload) => request.post('/oidc/conflicts/approve', data),
  rejectOIDCConflict: (data: RbacPayload) => request.post('/oidc/conflicts/reject', data),
  getOIDCIdentities: (params: PageQuery) => request.get('/oidc/identities', { params }),
  unbindOIDCIdentity: (data: RbacPayload) => request.post('/oidc/identities/unbind', data),
  getOIDCAlertConfig: () => request.get('/oidc/alert-config'),
  updateOIDCAlertConfig: (data: RbacPayload) => request.post('/oidc/alert-config', data),
}
