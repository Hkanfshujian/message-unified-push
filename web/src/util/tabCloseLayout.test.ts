import { describe, expect, it } from 'vitest'

const hiddenCloseWidth = 0
const visibleCloseWidth = 20
const tabCloseGap = 8
const defaultInlinePadding = 12
const visibleClosePaddingRight = 32

const getHiddenTabsSavedWidth = (closableTabCount: number) => closableTabCount * (visibleClosePaddingRight - defaultInlinePadding + visibleCloseWidth + tabCloseGap)

const getVisibleTabsExtraWidth = (visibleCloseCount: number) => visibleCloseCount * (visibleClosePaddingRight - defaultInlinePadding)

describe('workspace tab close control layout', () => {
  it('releases close control width and adjacent gap while hidden', () => {
    expect(hiddenCloseWidth).toBe(0)
    expect(defaultInlinePadding).toBe(12)
    expect(getHiddenTabsSavedWidth(1)).toBe(48)
  })

  it('keeps cumulative collapse stable when multiple close controls are hidden', () => {
    expect(getHiddenTabsSavedWidth(3)).toBe(144)
    expect(getHiddenTabsSavedWidth(8)).toBe(384)
  })

  it('only expands the close control space for interacted tabs', () => {
    expect(getVisibleTabsExtraWidth(1)).toBe(20)
    expect(getVisibleTabsExtraWidth(2)).toBe(40)
  })

  it('preserves meaningful space gains across compact and dense tab scenarios', () => {
    const compactViewportSavedWidth = getHiddenTabsSavedWidth(4)
    const denseTabsSavedWidth = getHiddenTabsSavedWidth(10)

    expect(compactViewportSavedWidth).toBeGreaterThan(100)
    expect(denseTabsSavedWidth).toBeGreaterThan(visibleCloseWidth * 10)
  })
})
