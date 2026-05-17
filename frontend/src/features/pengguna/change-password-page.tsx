import { useState } from 'react'
import { useChangePassword } from '@/hooks/use-users'
import { useAuth } from '@/hooks/use-auth-hook'
import { useNavigate } from 'react-router-dom'

export function ChangePasswordPage() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const changePassword = useChangePassword()
  const [form, setForm] = useState({ current_password: '', new_password: '', confirm_password: '' })
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  function handleChange(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (form.new_password.length < 8) {
      setError('Password baru minimal 8 karakter')
      return
    }
    if (form.new_password !== form.confirm_password) {
      setError('Konfirmasi password tidak cocok')
      return
    }

    try {
      await changePassword.mutateAsync({ current_password: form.current_password, new_password: form.new_password })
      setSuccess(true)
      setTimeout(async () => {
        await logout()
        navigate('/login')
      }, 2000)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal mengubah password')
    }
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Ubah Password</h1>
      <div className="mt-6 max-w-md">
        {success ? (
          <div className="rounded-lg border border-primary/20 bg-primary/5 p-4">
            <p className="text-sm text-primary">Password berhasil diubah. Anda akan diarahkan ke halaman login...</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Password Saat Ini</label>
              <input type="password" value={form.current_password} onChange={(e) => handleChange('current_password', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Password Baru</label>
              <input type="password" value={form.new_password} onChange={(e) => handleChange('new_password', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Konfirmasi Password Baru</label>
              <input type="password" value={form.confirm_password} onChange={(e) => handleChange('confirm_password', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>

            {error && <p className="text-sm text-destructive">{error}</p>}

            <button type="submit" disabled={changePassword.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {changePassword.isPending ? 'Menyimpan...' : 'Ubah Password'}
            </button>
          </form>
        )}
      </div>
    </div>
  )
}
