import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/use-auth-hook'
import { useTheme } from '@/hooks/use-theme-hook'
import { cn } from '@/lib/utils'
import { LayoutDashboard, Users, CreditCard, Settings, LogOut, Moon, Sun, Monitor, GraduationCap, BarChart3, Bell } from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard', roles: ['admin', 'operator', 'guru', 'siswa', 'orangtua'] },
  { to: '/siswa', icon: Users, label: 'Siswa', roles: ['admin', 'operator'] },
  { to: '/pembayaran', icon: CreditCard, label: 'Pembayaran', roles: ['admin', 'operator', 'siswa', 'orangtua'] },
  { to: '/ppdb', icon: GraduationCap, label: 'PPDB', roles: ['admin', 'operator'] },
  { to: '/laporan', icon: BarChart3, label: 'Laporan', roles: ['admin', 'operator'] },
  { to: '/notifikasi', icon: Bell, label: 'Notifikasi', roles: ['admin'] },
  { to: '/pengaturan', icon: Settings, label: 'Pengaturan', roles: ['admin'] },
]

export function DashboardLayout() {
  const { user, logout } = useAuth()
  const { theme, setMode } = useTheme()
  const navigate = useNavigate()

  const filteredNav = navItems.filter((item) =>
    user ? item.roles.includes(user.role) : false
  )

  async function handleLogout() {
    await logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen">
      <aside className="flex w-64 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <div className="flex h-14 items-center border-b border-sidebar-border px-4">
          <span className="text-lg font-semibold text-foreground">SekolahApp</span>
        </div>

        <nav className="flex-1 space-y-1 p-2">
          {filteredNav.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors',
                  isActive
                    ? 'bg-sidebar-accent text-foreground font-medium'
                    : 'text-sidebar-foreground hover:bg-sidebar-accent'
                )
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-sidebar-border p-2">
          <div className="flex items-center gap-1 rounded-md px-2 py-1">
            <button
              onClick={() => setMode('light')}
              className={cn('rounded p-1.5', theme.mode === 'light' && 'bg-sidebar-accent')}
            >
              <Sun className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => setMode('dark')}
              className={cn('rounded p-1.5', theme.mode === 'dark' && 'bg-sidebar-accent')}
            >
              <Moon className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => setMode('system')}
              className={cn('rounded p-1.5', theme.mode === 'system' && 'bg-sidebar-accent')}
            >
              <Monitor className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="mt-2 flex items-center justify-between rounded-md px-3 py-2">
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{user?.nama}</p>
              <p className="truncate text-xs text-muted-foreground">{user?.role}</p>
            </div>
            <button onClick={handleLogout} className="rounded p-1.5 hover:bg-sidebar-accent">
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
