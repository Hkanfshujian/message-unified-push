import { defineConfig, presetUno } from 'unocss'

export default defineConfig({
  presets: [presetUno()],
  theme: {
    colors: {
      brand: 'var(--brand)',
      primary: 'var(--primary)',
      background: 'var(--background)',
      foreground: 'var(--foreground)',
      card: 'var(--card-bg)',
      sidebar: 'var(--sidebar-bg)'
    }
  },
  shortcuts: {
    'admin-shell': 'min-h-screen bg-[var(--admin-layout-bg)] text-[var(--foreground)]',
    'admin-surface': 'bg-[var(--admin-surface-bg)] border border-[var(--admin-border)] shadow-[var(--admin-shadow-sm)]',
    'page-card': 'rounded-[var(--admin-radius-lg)] bg-[var(--admin-surface-bg)] shadow-[var(--admin-shadow-sm)] border border-solid border-[var(--admin-border)]',
    'page-toolbar': 'flex flex-wrap items-center gap-3 mb-4 rounded-[var(--admin-radius-md)] bg-[var(--admin-surface-muted)] p-4',
    'page-actions': 'flex flex-wrap items-center justify-end gap-2',
    'metric-card': 'rounded-[var(--admin-radius-xl)] bg-[var(--admin-surface-bg)] shadow-[var(--admin-shadow-sm)] border border-solid border-[var(--admin-border)]',
    'dashboard-grid': 'grid gap-4 grid-cols-1 sm:grid-cols-2 xl:grid-cols-4',
    'content-grid': 'grid gap-4 grid-cols-1 xl:grid-cols-[minmax(0,7fr)_minmax(300px,3fr)]',
    'glass-surface': 'bg-[var(--glass-base-bg)] border border-[var(--glass-border)] shadow-[var(--glass-shadow-soft)] backdrop-blur-[var(--glass-blur)]',
    'glass-panel': 'bg-[var(--glass-panel-bg)] border border-[var(--glass-border)] shadow-[var(--glass-shadow-panel)] backdrop-blur-[var(--glass-blur)]',
    'glass-overlay': 'bg-[var(--glass-overlay-bg)] border border-[var(--glass-border-strong)] shadow-[var(--glass-shadow-overlay)] backdrop-blur-[var(--glass-blur-strong)]',
    'glass-inset': 'bg-[var(--glass-inset-bg)] border border-[var(--glass-inset-border)] shadow-[var(--glass-shadow-inset)] backdrop-blur-[14px]',
    'glass-control': 'rounded-[var(--glass-control-radius)] border border-[var(--glass-border)] bg-[var(--glass-panel-bg)] shadow-[var(--glass-shadow-inset)]',
    'glass-stack': 'flex flex-col gap-4 md:gap-5'
  }
})
