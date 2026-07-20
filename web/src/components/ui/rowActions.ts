export type AppRowActionKind = 'view' | 'write'
export type AppRowActionType = 'primary' | 'success' | 'warning' | 'danger' | 'info' | ''
export type AppRowActionPermission = string | string[]

export interface AppRowAction {
  key: string
  label: string
  kind: AppRowActionKind
  permission?: AppRowActionPermission
  visible?: boolean
  disabled?: boolean
  loading?: boolean
  type?: AppRowActionType
  danger?: boolean
  onClick: () => void | Promise<void>
}

export interface AppRowActionAllocation {
  direct: AppRowAction[]
  more: AppRowAction[]
}

export type AppRowActionPermissionChecker = (permission: AppRowActionPermission) => boolean

export const allocateRowActions = (
  actions: AppRowAction[],
  hasPermission: AppRowActionPermissionChecker = () => true
): AppRowActionAllocation => {
  const allowed = actions
    .filter(action => action.visible !== false && (!action.permission || hasPermission(action.permission)))
    .map((action, index) => ({ action, index }))
    .sort((left, right) => {
      const leftPriority = left.action.kind === 'write' ? 0 : 1
      const rightPriority = right.action.kind === 'write' ? 0 : 1
      return leftPriority - rightPriority || left.index - right.index
    })
    .map(item => item.action)

  return allowed.length <= 3
    ? { direct: allowed, more: [] }
    : { direct: allowed.slice(0, 3), more: allowed.slice(3) }
}
