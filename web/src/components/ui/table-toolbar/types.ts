export interface TableToolbarColumn {
  key: string
  label: string
  required?: boolean
  visible?: boolean
}

export interface TableToolbarState {
  focused: boolean
  refreshing: boolean
  visibleColumns: string[]
}
