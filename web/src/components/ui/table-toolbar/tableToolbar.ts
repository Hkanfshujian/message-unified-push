import type { TableToolbarColumn, TableToolbarState } from './types'

export const createTableToolbarState = (columns: TableToolbarColumn[]): TableToolbarState => ({
  focused: false,
  refreshing: false,
  visibleColumns: columns.filter(column => column.required || column.visible !== false).map(column => column.key)
})

export const getVisibleToolbarColumns = <T extends { prop?: string; label: string }>(columns: T[], visibleColumns: string[]) => {
  const visible = new Set(visibleColumns)
  return columns.filter(column => !column.prop || visible.has(column.prop) || visible.has(column.label))
}

export const toggleColumnVisibility = (state: TableToolbarState, column: TableToolbarColumn) => {
  if (column.required) return state
  const visible = new Set(state.visibleColumns)
  if (visible.has(column.key)) visible.delete(column.key)
  else visible.add(column.key)
  return { ...state, visibleColumns: Array.from(visible) }
}

export const setFocusedView = (state: TableToolbarState, focused: boolean) => ({ ...state, focused })

export const setRefreshing = (state: TableToolbarState, refreshing: boolean) => ({ ...state, refreshing })

export const getTableToolbarMaterialClasses = (state: Pick<TableToolbarState, 'focused' | 'refreshing'>) => [
  'dora-material-panel',
  state.focused ? 'dora-state-focused' : '',
  state.refreshing ? 'dora-state-loading' : ''
].filter(Boolean)
