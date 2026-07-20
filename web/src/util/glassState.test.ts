import { describe, expect, it } from 'vitest'
import { createDoraStateClasses, createGlassStateClasses, getDoraStateClass, getGlassStateClass, hasInteractiveDoraState, hasInteractiveGlassState } from './glassState'

describe('Dora state helpers', () => {
  it('creates Dora class names for enabled component states', () => {
    expect(createDoraStateClasses({ focused: true, danger: true, disabled: false })).toEqual([
      'dora-state-focused',
      'dora-state-danger'
    ])
  })

  it('returns stable single state class names', () => {
    expect(getDoraStateClass('loading')).toBe('dora-state-loading')
    expect(getDoraStateClass('selected')).toBe('dora-state-selected')
  })

  it('detects interactive visual states', () => {
    expect(hasInteractiveDoraState({ danger: true })).toBe(false)
    expect(hasInteractiveDoraState({ active: true })).toBe(true)
    expect(hasInteractiveDoraState({ loading: true })).toBe(true)
  })

  it('keeps legacy helper names as Dora-compatible wrappers', () => {
    expect(createGlassStateClasses({ focused: true })).toEqual(['dora-state-focused'])
    expect(getGlassStateClass('loading')).toBe('dora-state-loading')
    expect(hasInteractiveGlassState({ selected: true })).toBe(true)
  })
})
