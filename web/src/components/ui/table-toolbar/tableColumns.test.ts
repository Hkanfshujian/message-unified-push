import { describe, expect, it } from 'vitest'
import { createTableToolbarState, getVisibleToolbarColumns, toggleColumnVisibility } from './tableToolbar'

describe('representative table column visibility', () => {
  it('keeps required action column while optional columns are hidden', () => {
    const toolbarColumns = [
      { key: 'id', label: 'ID', required: true },
      { key: 'name', label: '名称' },
      { key: 'actions', label: '详情/状态', required: true }
    ]
    const state = toggleColumnVisibility(createTableToolbarState(toolbarColumns), toolbarColumns[1])
    const columns = getVisibleToolbarColumns([
      { prop: 'id', label: 'ID' },
      { prop: 'name', label: '名称' },
      { prop: 'actions', label: '详情/状态' }
    ], state.visibleColumns)
    expect(columns.map(column => column.prop)).toEqual(['id', 'actions'])
  })
})
