import { request } from '@/api/api'

export const rbacApi = {
  getRoles: (params: any) => request.get('/rbac/roles', { params }),
  addRole: (data: any) => request.post('/rbac/roles', data),
  editRole: (data: any) => request.post('/rbac/roles/edit', data),
  deleteRole: (data: any) => request.post('/rbac/roles/delete', data),
  getRolePermissionIDs: (roleId: number) => request.get('/rbac/roles/permissions', { params: { role_id: roleId } }),
  assignRolePermissions: (data: any) => request.post('/rbac/roles/assign-permissions', data),

  getGroups: (params: any) => request.get('/rbac/groups', { params }),
  addGroup: (data: any) => request.post('/rbac/groups', data),
  editGroup: (data: any) => request.post('/rbac/groups/edit', data),
  deleteGroup: (data: any) => request.post('/rbac/groups/delete', data),
  getGroupRoleIDs: (groupId: number) => request.get('/rbac/groups/roles', { params: { group_id: groupId } }),
  getGroupMemberIDs: (groupId: number) => request.get('/rbac/groups/members', { params: { group_id: groupId } }),
  assignGroupRoles: (data: any) => request.post('/rbac/groups/assign-roles', data),
  assignGroupMembers: (data: any) => request.post('/rbac/groups/assign-members', data),

  getPermissions: (params: any) => request.get('/rbac/permissions', { params }),
  addPermission: (data: any) => request.post('/rbac/permissions', data),
  editPermission: (data: any) => request.post('/rbac/permissions/edit', data),

  getUsers: (params: any) => request.get('/rbac/users', { params }),
  getManageUsers: (params: any) => request.get('/rbac/users/manage', { params }),
  addManageUser: (data: any) => request.post('/rbac/users/manage', data),
  editManageUser: (data: any) => request.post('/rbac/users/manage/edit', data),
  deleteManageUser: (data: any) => request.post('/rbac/users/manage/delete', data),
  getUserRoleIDs: (userId: number) => request.get('/rbac/users/role-ids', { params: { user_id: userId } }),
  getUserGroupIDs: (userId: number) => request.get('/rbac/users/group-ids', { params: { user_id: userId } }),
  assignUserRoles: (data: any) => request.post('/rbac/users/assign-roles', data),
  assignUserGroups: (data: any) => request.post('/rbac/users/assign-groups', data),

  getOIDCMetrics: (params: any) => request.get('/oidc/metrics', { params }),
  getOIDCAudits: (params: any) => request.get('/oidc/audits', { params }),
  getOIDCConflicts: (params: any) => request.get('/oidc/conflicts', { params }),
  approveOIDCConflict: (data: any) => request.post('/oidc/conflicts/approve', data),
  rejectOIDCConflict: (data: any) => request.post('/oidc/conflicts/reject', data),
  getOIDCIdentities: (params: any) => request.get('/oidc/identities', { params }),
  unbindOIDCIdentity: (data: any) => request.post('/oidc/identities/unbind', data),
  getOIDCAlertConfig: () => request.get('/oidc/alert-config'),
  updateOIDCAlertConfig: (data: any) => request.post('/oidc/alert-config', data),
}
