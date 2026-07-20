import type { LayoutViewportClass } from '@/types/app'

export const MOBILE_MAX_WIDTH = 760
export const COMPACT_DESKTOP_MAX_WIDTH = 990
export const SIDEBAR_EXPANDED_WIDTH = 220
export const SIDEBAR_COLLAPSED_WIDTH = 64
export const DORA_HEADER_HEIGHT = 56
export const DORA_TABBAR_HEIGHT = 48
export const DORA_FOOTER_HEIGHT = 32

export const classifyViewport = (width: number): LayoutViewportClass => {
  if (width <= MOBILE_MAX_WIDTH) return 'mobile'
  if (width <= COMPACT_DESKTOP_MAX_WIDTH) return 'compact-desktop'
  return 'wide-desktop'
}

export const isMobileViewport = (width: number) => classifyViewport(width) === 'mobile'

export const getSidebarWidth = (collapsed: boolean) => collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_EXPANDED_WIDTH

export const getContentOffset = (viewportClass: LayoutViewportClass, collapsed: boolean) => {
  if (viewportClass === 'mobile') return 0
  return getSidebarWidth(collapsed)
}

export const getLayoutMode = (width: number) => {
  const viewportClass = classifyViewport(width)
  if (viewportClass === 'mobile') return 'temporary'
  if (viewportClass === 'compact-desktop') return 'collapsed'
  return 'expanded'
}

export const getShellTopOffset = (showTabBar: boolean) => DORA_HEADER_HEIGHT + (showTabBar ? DORA_TABBAR_HEIGHT : 0)

export const getDoraShellMetrics = (width: number, showTabBar = true) => {
  const viewportClass = classifyViewport(width)
  return {
    viewportClass,
    headerHeight: DORA_HEADER_HEIGHT,
    tabbarHeight: showTabBar ? DORA_TABBAR_HEIGHT : 0,
    footerHeight: DORA_FOOTER_HEIGHT,
    topOffset: getShellTopOffset(showTabBar),
    pagePadding: viewportClass === 'mobile' ? 12 : 16,
    contentGap: viewportClass === 'mobile' ? 8 : 12,
    shouldWrapHeaderActions: viewportClass === 'mobile',
    shouldStackPageToolbar: viewportClass === 'mobile'
  }
}

export const getResponsiveGlassShellMetrics = (width: number, showTabBar = true) => {
  const metrics = getDoraShellMetrics(width, showTabBar)
  return { ...metrics, mainCardPadding: metrics.pagePadding }
}

export const getDoraLayoutState = (width: number, mobileSidebarVisible = false) => {
  const viewportClass = classifyViewport(width)
  const mode = getLayoutMode(width)
  return {
    viewportClass,
    mode,
    sidebarMode: mode === 'temporary' && mobileSidebarVisible ? 'overlay' : 'inline',
    contentMode: viewportClass === 'mobile' ? 'stacked' : 'contained'
  }
}

export const getGlassLayoutState = (width: number, mobileSidebarVisible = false) => {
  const state = getDoraLayoutState(width, mobileSidebarVisible)
  return {
    viewportClass: state.viewportClass,
    mode: state.mode,
    sidebarMaterial: state.sidebarMode === 'overlay' ? 'overlay' : 'base',
    contentMaterial: state.contentMode === 'stacked' ? 'panel' : 'base'
  }
}
