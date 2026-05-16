export type ThemeMode = 'light' | 'dark' | 'system'

export interface ThemeConfig {
  mode: ThemeMode
  primaryColor?: string
  accentColor?: string
  radius?: string
  density?: 'comfortable' | 'compact'
  logo?: string
  namaSekolah?: string
}

const STORAGE_KEY = 'sekolah-theme'

export function getStoredTheme(): ThemeConfig {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) return JSON.parse(stored)
  } catch {
    void 0
  }
  return { mode: 'system' }
}

export function setStoredTheme(config: ThemeConfig) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(config))
}

export function applyThemeMode(mode: ThemeMode) {
  const root = document.documentElement
  root.classList.remove('light', 'dark')

  if (mode === 'system') {
    const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    root.classList.add(systemDark ? 'dark' : 'light')
  } else {
    root.classList.add(mode)
  }
}

export function applyThemeConfig(config: ThemeConfig) {
  applyThemeMode(config.mode)

  const root = document.documentElement
  if (config.primaryColor) {
    root.style.setProperty('--primary', config.primaryColor)
  }
  if (config.accentColor) {
    root.style.setProperty('--accent', config.accentColor)
  }
  if (config.radius) {
    root.style.setProperty('--radius', config.radius)
  }
}
