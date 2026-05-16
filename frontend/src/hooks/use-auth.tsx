import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import api from '@/lib/api'
import type { User } from '@/types'

interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (kodeSekolah: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refetch: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  async function fetchUser() {
    try {
      const res = await api.get('/auth/me')
      setUser(res.data.data)
    } catch {
      setUser(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchUser()
  }, [])

  async function login(kodeSekolah: string, email: string, password: string) {
    await api.post('/auth/login', {
      kode_sekolah: kodeSekolah,
      email,
      password,
    })
    await fetchUser()
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
    } finally {
      setUser(null)
    }
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refetch: fetchUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
