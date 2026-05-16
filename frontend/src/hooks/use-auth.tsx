import { useEffect, useState, type ReactNode } from 'react'
import api from '@/lib/api'
import type { User } from '@/types'
import { AuthContext } from '@/hooks/auth-context'

export interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (kodeSekolah: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refetch: () => Promise<void>
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    api.get('/auth/me')
      .then((res) => { if (!cancelled) setUser(res.data.data) })
      .catch(() => { if (!cancelled) setUser(null) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  async function fetchUser() {
    try {
      const res = await api.get('/auth/me')
      setUser(res.data.data)
    } catch {
      setUser(null)
    }
  }

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
