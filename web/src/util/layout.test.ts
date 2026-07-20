import { describe, expect, it } from 'vitest'
import { classifyViewport, getContentOffset, getDoraLayoutState, getDoraShellMetrics, getLayoutMode, getResponsiveGlassShellMetrics, getShellTopOffset, getSidebarWidth, isMobileViewport } from './layout'

describe('layout viewport helpers', () => {
  it('classifies responsive breakpoints', () => {
    expect(classifyViewport(760)).toBe('mobile')
    expect(classifyViewport(761)).toBe('compact-desktop')
    expect(classifyViewport(990)).toBe('compact-desktop')
    expect(classifyViewport(991)).toBe('wide-desktop')
  })

  it('detects mobile viewport', () => {
    expect(isMobileViewport(480)).toBe(true)
    expect(isMobileViewport(900)).toBe(false)
  })

  it('resolves sidebar width and content offset', () => {
    expect(getSidebarWidth(true)).toBe(64)
    expect(getSidebarWidth(false)).toBe(220)
    expect(getContentOffset('mobile', false)).toBe(0)
    expect(getContentOffset('wide-desktop', true)).toBe(64)
  })

  it('resolves shell mode and top offset', () => {
    expect(getLayoutMode(480)).toBe('temporary')
    expect(getLayoutMode(900)).toBe('collapsed')
    expect(getLayoutMode(1200)).toBe('expanded')
    expect(getShellTopOffset(true)).toBe(104)
    expect(getShellTopOffset(false)).toBe(56)
  })

  it('resolves Dora layout state by viewport', () => {
    expect(getDoraLayoutState(640, true)).toMatchObject({
      viewportClass: 'mobile',
      mode: 'temporary',
      sidebarMode: 'overlay',
      contentMode: 'stacked'
    })
    expect(getDoraLayoutState(1200)).toMatchObject({
      viewportClass: 'wide-desktop',
      mode: 'expanded',
      sidebarMode: 'inline',
      contentMode: 'contained'
    })
  })

  it('resolves responsive Dora shell spacing and wrapping metrics', () => {
    expect(getDoraShellMetrics(1280)).toMatchObject({
      viewportClass: 'wide-desktop',
      pagePadding: 16,
      contentGap: 12,
      shouldWrapHeaderActions: false,
      shouldStackPageToolbar: false
    })
    expect(getDoraShellMetrics(900)).toMatchObject({
      viewportClass: 'compact-desktop',
      pagePadding: 16,
      contentGap: 12,
      shouldWrapHeaderActions: false,
      shouldStackPageToolbar: false
    })
    expect(getResponsiveGlassShellMetrics(640, false)).toMatchObject({
      viewportClass: 'mobile',
      topOffset: 56,
      pagePadding: 12,
      mainCardPadding: 12,
      shouldWrapHeaderActions: true,
      shouldStackPageToolbar: true
    })
  })
})
