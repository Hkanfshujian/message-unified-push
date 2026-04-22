
export interface Theme {
    name: string;
    key: string;
    light: string;
    dark: string;
    oklch_light: string;
    oklch_dark: string;
}

export const THEMES: Theme[] = [
    {
        name: '雅致蓝',
        key: 'blue',
        light: '#3b82f6',
        dark: '#60a5fa',
        oklch_light: 'oklch(0.55 0.18 260)',
        oklch_dark: 'oklch(0.72 0.16 260)'
    },
    {
        name: '沉静翠',
        key: 'green',
        light: '#059669',
        dark: '#34d399',
        oklch_light: 'oklch(0.58 0.14 160)',
        oklch_dark: 'oklch(0.75 0.12 160)'
    },
    {
        name: '极简紫',
        key: 'purple',
        light: '#7c3aed',
        dark: '#a78bfa',
        oklch_light: 'oklch(0.52 0.22 295)',
        oklch_dark: 'oklch(0.70 0.18 295)'
    },
    {
        name: '陶韵橙',
        key: 'orange',
        light: '#ea580c',
        dark: '#fb923c',
        oklch_light: 'oklch(0.58 0.18 45)',
        oklch_dark: 'oklch(0.74 0.16 45)'
    },
    {
        name: '霜灰蓝',
        key: 'slate',
        light: '#475569',
        dark: '#94a3b8',
        oklch_light: 'oklch(0.45 0.05 255)',
        oklch_dark: 'oklch(0.70 0.04 255)'
    },
    {
        name: '极简青',
        key: 'teal',
        light: '#0d9488',
        dark: '#2dd4bf',
        oklch_light: 'oklch(0.55 0.12 190)',
        oklch_dark: 'oklch(0.75 0.10 190)'
    },
    {
        name: '胭脂红',
        key: 'rose',
        light: '#e11d48',
        dark: '#fb7185',
        oklch_light: 'oklch(0.50 0.20 15)',
        oklch_dark: 'oklch(0.70 0.18 15)'
    }
]

// 生成品牌色的不同透明度层级
const generateBrandScale = (hexColor: string): Record<string, string> => {
    // 移除 # 号
    const hex = hexColor.replace('#', '')
    const r = parseInt(hex.substring(0, 2), 16)
    const g = parseInt(hex.substring(2, 4), 16)
    const b = parseInt(hex.substring(4, 6), 16)
    
    return {
        '50': `rgba(${r}, ${g}, ${b}, 0.05)`,
        '100': `rgba(${r}, ${g}, ${b}, 0.1)`,
        '200': `rgba(${r}, ${g}, ${b}, 0.2)`,
        '300': `rgba(${r}, ${g}, ${b}, 0.3)`,
        '400': `rgba(${r}, ${g}, ${b}, 0.4)`,
        '500': `rgba(${r}, ${g}, ${b}, 0.6)`,
        '600': `rgba(${r}, ${g}, ${b}, 0.7)`,
        '700': `rgba(${r}, ${g}, ${b}, 0.8)`,
        '800': `rgba(${r}, ${g}, ${b}, 0.9)`,
        '900': hexColor,
    }
}

export const applyTheme = (themeKey: string) => {
    const root = document.documentElement

    let brandLight: string
    let brandDark: string
    let tabbarLight: string
    let tabbarDark: string
    let brandScaleLight: Record<string, string>

    if (themeKey.startsWith('custom:')) {
        const hex = themeKey.split(':')[1] || '#1890ff'
        brandLight = hex
        brandDark = hex
        tabbarLight = '#ffffff'
        tabbarDark = '#0f172a'
        brandScaleLight = generateBrandScale(hex)
    } else {
        const theme = THEMES.find(t => t.key === themeKey) || THEMES[0]
        brandLight = theme.oklch_light
        brandDark = theme.oklch_dark
        tabbarLight = theme.key === 'slate' ? '#0f172a' : '#ffffff'
        tabbarDark = theme.key === 'slate' ? '#020617' : '#0f172a'
        brandScaleLight = generateBrandScale(theme.light)
    }

    root.style.setProperty('--brand-light', brandLight)
    root.style.setProperty('--brand-dark', brandDark)
    
    // 设置品牌色层级变量
    Object.entries(brandScaleLight).forEach(([key, value]) => {
        root.style.setProperty(`--brand-${key}`, value)
    })

    let styleEl = document.getElementById('dynamic-theme-style') as HTMLStyleElement | null
    if (!styleEl) {
        styleEl = document.createElement('style')
        styleEl.id = 'dynamic-theme-style'
        document.head.appendChild(styleEl)
    }

    styleEl.innerHTML = `
    :root {
      --brand: ${brandLight};
      --primary: ${brandLight};
      --tabbar-bg: ${tabbarLight};
    }
    .dark {
      --brand: ${brandDark};
      --primary: ${brandDark};
      --tabbar-bg: ${tabbarDark};
    }
  `

    localStorage.setItem('themeColor', themeKey)
}

export const getStoredTheme = (): string => {
    return localStorage.getItem('themeColor') || 'blue'
}
