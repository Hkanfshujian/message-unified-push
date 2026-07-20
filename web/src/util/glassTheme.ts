export type GlassThemeMode = 'light' | 'dark'

export type GlassMaterialLevel = 'canvas' | 'base' | 'panel' | 'inset' | 'overlay' | 'active' | 'danger'
export type DoraThemeMode = GlassThemeMode
export type DoraMaterialLevel = GlassMaterialLevel

export interface GlassMaterialTokens {
  background: string
  border: string
  shadow: string
  highlight: string
  blur: string
}

export interface GlassThemeTokens {
  mode: GlassThemeMode
  materials: Record<GlassMaterialLevel, GlassMaterialTokens>
  fallbackClass: string
  reducedMotionClass: string
}

const lightMaterials: Record<GlassMaterialLevel, GlassMaterialTokens> = {
  canvas: {
    background: '#f7fafc',
    border: 'rgba(239, 243, 248, 0.96)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  },
  base: {
    background: '#ffffff',
    border: 'rgba(239, 243, 248, 0.96)',
    shadow: '0 1px 2px rgba(0, 21, 41, 0.08)',
    highlight: 'transparent',
    blur: '0px'
  },
  panel: {
    background: '#ffffff',
    border: 'rgba(239, 243, 248, 0.96)',
    shadow: '0 1px 2px rgba(0, 21, 41, 0.08)',
    highlight: 'transparent',
    blur: '0px'
  },
  inset: {
    background: '#f7fafc',
    border: 'rgba(239, 243, 248, 0.96)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  },
  overlay: {
    background: '#ffffff',
    border: 'rgba(239, 243, 248, 0.96)',
    shadow: '0 8px 24px rgba(0, 21, 41, 0.10)',
    highlight: 'transparent',
    blur: '0px'
  },
  active: {
    background: 'rgba(37, 99, 235, 0.10)',
    border: 'rgba(37, 99, 235, 0.24)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  },
  danger: {
    background: 'rgba(239, 68, 68, 0.10)',
    border: 'rgba(239, 68, 68, 0.24)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  }
}

const darkMaterials: Record<GlassMaterialLevel, GlassMaterialTokens> = {
  canvas: {
    background: '#121212',
    border: 'rgba(255, 255, 255, 0.08)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  },
  base: {
    background: '#1c1c1c',
    border: 'rgba(255, 255, 255, 0.08)',
    shadow: '0 1px 2px rgba(0, 0, 0, 0.28)',
    highlight: 'transparent',
    blur: '0px'
  },
  panel: {
    background: '#1c1c1c',
    border: 'rgba(255, 255, 255, 0.08)',
    shadow: '0 1px 2px rgba(0, 0, 0, 0.28)',
    highlight: 'transparent',
    blur: '0px'
  },
  inset: {
    background: '#121212',
    border: 'rgba(255, 255, 255, 0.08)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  },
  overlay: {
    background: '#1c1c1c',
    border: 'rgba(255, 255, 255, 0.08)',
    shadow: '0 8px 24px rgba(0, 0, 0, 0.34)',
    highlight: 'transparent',
    blur: '0px'
  },
  active: {
    background: 'rgba(59, 130, 246, 0.16)',
    border: 'rgba(59, 130, 246, 0.28)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  },
  danger: {
    background: 'rgba(239, 68, 68, 0.16)',
    border: 'rgba(239, 68, 68, 0.28)',
    shadow: 'none',
    highlight: 'transparent',
    blur: '0px'
  }
}

export const createDoraThemeTokens = (mode: DoraThemeMode): GlassThemeTokens => ({
  mode,
  materials: mode === 'dark' ? darkMaterials : lightMaterials,
  fallbackClass: 'dora-no-backdrop-filter',
  reducedMotionClass: 'dora-reduced-motion'
})

export const getDoraMaterialClass = (level: DoraMaterialLevel) => `dora-material-${level}`

export const getDoraThemeClass = (mode: DoraThemeMode) => `dora-theme-${mode}`

export const createGlassThemeTokens = createDoraThemeTokens
export const getGlassMaterialClass = getDoraMaterialClass
export const getGlassThemeClass = getDoraThemeClass
