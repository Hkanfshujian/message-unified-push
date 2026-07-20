import { describe, expect, it } from 'vitest'
import { createTableToolbarState, getTableToolbarMaterialClasses, getVisibleToolbarColumns, setFocusedView, setRefreshing, toggleColumnVisibility } from './tableToolbar'

describe('table toolbar helpers', () => {
  const toolbarColumns = [
    { key: 'id', label: 'ID', required: true },
    { key: 'name', label: '名称' },
    { key: 'log', label: '日志', visible: false }
  ]

  it('creates default state from columns', () => {
    expect(createTableToolbarState(toolbarColumns).visibleColumns).toEqual(['id', 'name'])
  })

  it('toggles optional columns but keeps required columns visible', () => {
    const state = createTableToolbarState(toolbarColumns)
    expect(toggleColumnVisibility(state, toolbarColumns[0]).visibleColumns).toEqual(['id', 'name'])
    expect(toggleColumnVisibility(state, toolbarColumns[1]).visibleColumns).toEqual(['id'])
  })

  it('filters visible table columns and updates view state', () => {
    const state = createTableToolbarState(toolbarColumns)
    expect(getVisibleToolbarColumns([{ prop: 'id', label: 'ID' }, { prop: 'log', label: '日志' }], state.visibleColumns).map(column => column.prop)).toEqual(['id'])
    expect(setFocusedView(state, true).focused).toBe(true)
    expect(setRefreshing(state, true).refreshing).toBe(true)
  })

  it('maps material state to stable Dora classes', () => {
    const state = setRefreshing(setFocusedView(createTableToolbarState(toolbarColumns), true), true)
    expect(getTableToolbarMaterialClasses(state)).toEqual([
      'dora-material-panel',
      'dora-state-focused',
      'dora-state-loading'
    ])
  })
})
