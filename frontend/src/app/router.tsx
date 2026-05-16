import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from '@/hooks/use-auth'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { LoginPage } from '@/features/auth/login-page'
import { SetupPage } from '@/features/setup/setup-page'
import { DashboardPage } from '@/features/dashboard/dashboard-page'
import { SiswaPage } from '@/features/siswa/siswa-page'
import { PembayaranPage } from '@/features/pembayaran/pembayaran-page'
import { PengaturanPage } from '@/features/pengaturan/pengaturan-page'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-muted-foreground">Memuat...</div>
      </div>
    )
  }

  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/setup" element={<SetupPage />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <DashboardLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="siswa" element={<SiswaPage />} />
          <Route path="pembayaran" element={<PembayaranPage />} />
          <Route path="pengaturan" element={<PengaturanPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
