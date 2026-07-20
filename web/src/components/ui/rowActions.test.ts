import { describe, expect, it } from 'vitest'
import { allocateRowActions, type AppRowAction } from './rowActions'

const action = (key: string, kind: AppRowAction['kind'] = 'view', overrides: Partial<AppRowAction> = {}): AppRowAction => ({
  key,
  label: key,
  kind,
  onClick: () => undefined,
  ...overrides
})

const keys = (actions: AppRowAction[]) => actions.map(item => item.key)

describe('allocateRowActions', () => {
  it('directly displays the only view action', () => {
    const result = allocateRowActions([action('查看')])
    expect(keys(result.direct)).toEqual(['查看'])
    expect(result.more).toEqual([])
  })

  it.each([
    [['编辑', '查看'], 2],
    [['编辑', '删除', '查看'], 3]
  ])('directly displays %i actions without more', (labels) => {
    const actions = labels.map((label, index) => action(label, index < 2 ? 'write' : 'view'))
    const result = allocateRowActions(actions)
    expect(keys(result.direct)).toEqual(labels)
    expect(result.more).toEqual([])
  })

  it('places actions after the first three in more', () => {
    const result = allocateRowActions([action('a', 'write'), action('b', 'write'), action('c'), action('d')])
    expect(keys(result.direct)).toEqual(['a', 'b', 'c'])
    expect(keys(result.more)).toEqual(['d'])
  })

  it('prioritizes writes and preserves order within each kind', () => {
    const result = allocateRowActions([action('view-1'), action('write-1', 'write'), action('view-2'), action('write-2', 'write')])
    expect([...keys(result.direct), ...keys(result.more)]).toEqual(['write-1', 'write-2', 'view-1', 'view-2'])
  })

  it('filters permissions before allocating and leaves the view direct', () => {
    const result = allocateRowActions([
      action('编辑', 'write', { permission: 'edit' }),
      action('查看', 'view', { permission: ['view', 'other'] })
    ], permission => Array.isArray(permission) ? permission.includes('view') : permission === 'view')
    expect(keys(result.direct)).toEqual(['查看'])
    expect(result.more).toEqual([])
  })

  it('does not produce empty more after visible filtering', () => {
    const result = allocateRowActions([action('隐藏', 'write', { visible: false }), action('查看')])
    expect(keys(result.direct)).toEqual(['查看'])
    expect(result.more).toEqual([])
  })

  it('keeps disabled actions visible', () => {
    const result = allocateRowActions([action('禁用', 'write', { disabled: true })])
    expect(keys(result.direct)).toEqual(['禁用'])
    expect(result.direct[0].disabled).toBe(true)
  })
})
