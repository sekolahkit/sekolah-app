import { createContext } from 'react'
import type { ThemeContextValue } from '@/hooks/use-theme'

export const ThemeContext = createContext<ThemeContextValue | undefined>(undefined)
