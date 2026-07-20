export interface GlassStateOptions {
  focused?: boolean
  danger?: boolean
  disabled?: boolean
  loading?: boolean
  selected?: boolean
  active?: boolean
}

export type DoraStateName = keyof GlassStateOptions
export type GlassStateName = DoraStateName

const stateClassMap: Record<DoraStateName, string> = {
  focused: 'dora-state-focused',
  danger: 'dora-state-danger',
  disabled: 'dora-state-disabled',
  loading: 'dora-state-loading',
  selected: 'dora-state-selected',
  active: 'dora-state-active'
}

export const getDoraStateClass = (state: DoraStateName) => stateClassMap[state]

export const createDoraStateClasses = (options: GlassStateOptions) => {
  return Object.entries(options)
    .filter(([, enabled]) => Boolean(enabled))
    .map(([state]) => stateClassMap[state as DoraStateName])
}

export const hasInteractiveDoraState = (options: GlassStateOptions) => {
  return Boolean(options.focused || options.loading || options.selected || options.active)
}

export const getGlassStateClass = getDoraStateClass
export const createGlassStateClasses = createDoraStateClasses
export const hasInteractiveGlassState = hasInteractiveDoraState
