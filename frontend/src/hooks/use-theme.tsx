import { useEffect, useState, type ReactNode } from 'react'
import { type ThemeConfig, type ThemeMode, getStoredTheme, setStoredTheme, applyThemeConfig } from '@/lib/theme'
import { ThemeContext } from '@/hooks/theme-context'

export interface ThemeContextValue {
  theme: ThemeConfig
  setTheme: (config: Partial<ThemeConfig>) => void
  setMode: (mode: ThemeMode) => void
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<ThemeConfig>(getStoredTheme)

  useEffect(() => {
    applyThemeConfig(theme)
  }, [theme])

  useEffect(() => {
    if (theme.mode === 'system') {
      const mq = window.matchMedia('(prefers-color-scheme: dark)')
      const handler = () => applyThemeConfig(theme)
      mq.addEventListener('change', handler)
      return () => mq.removeEventListener('change', handler)
    }
  }, [theme])

  function setTheme(config: Partial<ThemeConfig>) {
    const next = { ...theme, ...config }
    setThemeState(next)
    setStoredTheme(next)
  }

  function setMode(mode: ThemeMode) {
    setTheme({ mode })
  }

  return (
    <ThemeContext.Provider value={{ theme, setTheme, setMode }}>
      {children}
    </ThemeContext.Provider>
  )
}
