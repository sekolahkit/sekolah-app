import { createContext } from 'react'
import type { AuthContextValue } from '@/hooks/use-auth'

export const AuthContext = createContext<AuthContextValue | undefined>(undefined)
