import { useState } from 'react'
import { useUserList, useCreateUser, useUpdateUser, useDeactivateUser, useResetUserPassword } from '@/hooks/use-users'
import { cn } from '@/lib/utils'
import { Plus, Search, Edit2, UserX, X, KeyRound } from 'lucide-react'
import type { UserDetail } from '@/types'

const ROLES = ['admin', 'operator', 'guru', 'siswa', 'orangtua'] as const

export function UserManagementPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editingUser, setEditingUser] = useState<UserDetail | null>(null)
  const [deactivateConfirm, setDeactivateConfirm] = useState<number | null>(null)
  const [resetPasswordUser, setResetPasswordUser] = useState<UserDetail | null>(null)

  const { data, isLoading } = useUserList({ page, limit: 20, search: search || undefined, role: roleFilter || undefined })
  const deactivateMutation = useDeactivateUser()

  function handleEdit(u: UserDetail) {
    setEditingUser(u)
    setShowForm(true)
  }

  function handleCreate() {
    setEditingUser(null)
    setShowForm(true)
  }

  function handleDeactivate(id: number) {
    deactivateMutation.mutate(id, { onSuccess: () => setDeactivateConfirm(null) })
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-bold text-foreground">Manajemen Pengguna</h1>
        <button onClick={handleCreate} className="flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Plus className="h-4 w-4" /> Tambah Pengguna
        </button>
      </div>

      <div className="mt-4 flex items-center gap-2 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Cari nama atau email..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full rounded-md border border-input bg-background py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
        <select
          value={roleFilter}
          onChange={(e) => { setRoleFilter(e.target.value); setPage(1) }}
          className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground"
        >
          <option value="">Semua Role</option>
          {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
        </select>
      </div>

      {isLoading ? (
        <div className="mt-8 text-center text-muted-foreground">Memuat data...</div>
      ) : !data?.data?.length ? (
        <div className="mt-8 text-center text-muted-foreground">Belum ada data pengguna.</div>
      ) : (
        <>
          <div className="mt-4 overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="border-b border-border bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Nama</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Email</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Role</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((u) => (
                  <tr key={u.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="px-4 py-3 text-foreground">{u.nama}</td>
                    <td className="px-4 py-3 text-foreground">{u.email}</td>
                    <td className="px-4 py-3">
                      <RoleBadge role={u.role} />
                    </td>
                    <td className="px-4 py-3">
                      <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', u.aktif ? 'bg-primary/10 text-primary' : 'bg-destructive/10 text-destructive')}>
                        {u.aktif ? 'Aktif' : 'Nonaktif'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button onClick={() => handleEdit(u)} className="rounded p-1.5 hover:bg-muted" title="Edit">
                          <Edit2 className="h-4 w-4 text-muted-foreground" />
                        </button>
                        <button onClick={() => setResetPasswordUser(u)} className="rounded p-1.5 hover:bg-muted" title="Reset Password">
                          <KeyRound className="h-4 w-4 text-muted-foreground" />
                        </button>
                        {u.aktif && (
                          <button onClick={() => setDeactivateConfirm(u.id)} className="rounded p-1.5 hover:bg-destructive/10" title="Nonaktifkan">
                            <UserX className="h-4 w-4 text-destructive" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.meta.total_pages > 1 && (
            <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
              <span>Halaman {data.meta.page} dari {data.meta.total_pages} ({data.meta.total} pengguna)</span>
              <div className="flex gap-2">
                <button disabled={page <= 1} onClick={() => setPage(page - 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Sebelumnya</button>
                <button disabled={page >= data.meta.total_pages} onClick={() => setPage(page + 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Berikutnya</button>
              </div>
            </div>
          )}
        </>
      )}

      {showForm && (
        <UserFormDialog user={editingUser} onClose={() => setShowForm(false)} />
      )}

      {deactivateConfirm !== null && (
        <ConfirmDialog
          message="Yakin ingin menonaktifkan pengguna ini?"
          confirmLabel="Nonaktifkan"
          onConfirm={() => handleDeactivate(deactivateConfirm)}
          onCancel={() => setDeactivateConfirm(null)}
          loading={deactivateMutation.isPending}
        />
      )}

      {resetPasswordUser && (
        <ResetPasswordDialog user={resetPasswordUser} onClose={() => setResetPasswordUser(null)} />
      )}
    </div>
  )
}

function RoleBadge({ role }: { role: string }) {
  const styles: Record<string, string> = {
    admin: 'bg-primary/10 text-primary',
    operator: 'bg-accent text-accent-foreground',
    guru: 'bg-muted text-muted-foreground',
    siswa: 'bg-muted text-muted-foreground',
    orangtua: 'bg-muted text-muted-foreground',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[role] || 'bg-muted text-muted-foreground')}>
      {role}
    </span>
  )
}

function UserFormDialog({ user, onClose }: { user: UserDetail | null; onClose: () => void }) {
  const createMutation = useCreateUser()
  const updateMutation = useUpdateUser()
  const [error, setError] = useState('')

  const [form, setForm] = useState({
    nama: user?.nama || '',
    email: user?.email || '',
    role: user?.role || 'operator',
    no_hp: user?.no_hp || '',
    aktif: user?.aktif ?? true,
    password: '',
  })

  function handleChange(field: string, value: string | boolean) {
    setForm((f) => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (user) {
        await updateMutation.mutateAsync({
          id: user.id,
          data: { nama: form.nama, email: form.email, role: form.role, no_hp: form.no_hp || undefined, aktif: form.aktif },
        })
      } else {
        await createMutation.mutateAsync({
          nama: form.nama,
          email: form.email,
          password: form.password,
          role: form.role,
          no_hp: form.no_hp || undefined,
        })
      }
      onClose()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal menyimpan data')
    }
  }

  const loading = createMutation.isPending || updateMutation.isPending

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">{user ? 'Edit Pengguna' : 'Tambah Pengguna'}</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <FormField label="Nama" value={form.nama} onChange={(v) => handleChange('nama', v)} required />
          <FormField label="Email" value={form.email} onChange={(v) => handleChange('email', v)} type="email" required />
          {!user && (
            <FormField label="Password" value={form.password} onChange={(v) => handleChange('password', v)} type="password" required />
          )}
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Role</label>
            <select value={form.role} onChange={(e) => handleChange('role', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
            </select>
          </div>
          <FormField label="No HP" value={form.no_hp} onChange={(v) => handleChange('no_hp', v)} />
          {user && (
            <div className="flex items-center gap-2">
              <input type="checkbox" id="aktif" checked={form.aktif} onChange={(e) => handleChange('aktif', e.target.checked)} className="rounded border-input" />
              <label htmlFor="aktif" className="text-sm text-foreground">Aktif</label>
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
            <button type="submit" disabled={loading} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {loading ? 'Menyimpan...' : 'Simpan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function ResetPasswordDialog({ user, onClose }: { user: UserDetail; onClose: () => void }) {
  const resetMutation = useResetUserPassword()
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await resetMutation.mutateAsync({ id: user.id, password })
      setSuccess(true)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal reset password')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">Reset Password</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>
        <p className="mt-2 text-sm text-muted-foreground">Reset password untuk <strong>{user.nama}</strong></p>

        {success ? (
          <div className="mt-4">
            <p className="text-sm text-primary">Password berhasil direset.</p>
            <div className="mt-4 flex justify-end">
              <button onClick={onClose} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">Tutup</button>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="mt-4 space-y-3">
            <FormField label="Password Baru" value={password} onChange={setPassword} type="password" required />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
              <button type="submit" disabled={resetMutation.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
                {resetMutation.isPending ? 'Mereset...' : 'Reset'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

function ConfirmDialog({ message, confirmLabel, onConfirm, onCancel, loading }: { message: string; confirmLabel: string; onConfirm: () => void; onCancel: () => void; loading: boolean }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-6 shadow-lg">
        <p className="text-sm text-card-foreground">{message}</p>
        <div className="mt-4 flex justify-end gap-2">
          <button onClick={onCancel} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
          <button onClick={onConfirm} disabled={loading} className="rounded-md bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:opacity-90 disabled:opacity-50">
            {loading ? 'Memproses...' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}

function FormField({ label, value, onChange, type = 'text', required = false }: { label: string; value: string; onChange: (v: string) => void; type?: string; required?: boolean }) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <input type={type} value={value} onChange={(e) => onChange(e.target.value)} required={required} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
    </div>
  )
}
