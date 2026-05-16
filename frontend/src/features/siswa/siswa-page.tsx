import { useState } from 'react'
import { useSiswaList, useCreateSiswa, useUpdateSiswa, useDeleteSiswa } from '@/hooks/use-siswa'
import { useAuth } from '@/hooks/use-auth'
import { cn } from '@/lib/utils'
import { Plus, Search, Edit2, Trash2, X } from 'lucide-react'
import type { Siswa } from '@/types'

export function SiswaPage() {
  const { user } = useAuth()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editingSiswa, setEditingSiswa] = useState<Siswa | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<number | null>(null)

  const { data, isLoading } = useSiswaList({ page, limit: 20, search: search || undefined })
  const deleteMutation = useDeleteSiswa()

  function handleEdit(s: Siswa) {
    setEditingSiswa(s)
    setShowForm(true)
  }

  function handleCreate() {
    setEditingSiswa(null)
    setShowForm(true)
  }

  function handleDelete(id: number) {
    deleteMutation.mutate(id, { onSuccess: () => setDeleteConfirm(null) })
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Data Siswa</h1>
        <button onClick={handleCreate} className="flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Plus className="h-4 w-4" /> Tambah Siswa
        </button>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Cari nama atau NIS..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full rounded-md border border-input bg-background py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="mt-8 text-center text-muted-foreground">Memuat data...</div>
      ) : !data?.data?.length ? (
        <div className="mt-8 text-center text-muted-foreground">Belum ada data siswa.</div>
      ) : (
        <>
          <div className="mt-4 overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="border-b border-border bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">NIS</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Nama</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">L/P</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((s) => (
                  <tr key={s.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-foreground">{s.nis}</td>
                    <td className="px-4 py-3 text-foreground">{s.nama}</td>
                    <td className="px-4 py-3 text-foreground">{s.jenis_kelamin}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={s.status} />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button onClick={() => handleEdit(s)} className="rounded p-1.5 hover:bg-muted" title="Edit">
                          <Edit2 className="h-4 w-4 text-muted-foreground" />
                        </button>
                        {user?.role === 'admin' && (
                          <button onClick={() => setDeleteConfirm(s.id)} className="rounded p-1.5 hover:bg-destructive/10" title="Hapus">
                            <Trash2 className="h-4 w-4 text-destructive" />
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
              <span>Halaman {data.meta.page} dari {data.meta.total_pages} ({data.meta.total} siswa)</span>
              <div className="flex gap-2">
                <button disabled={page <= 1} onClick={() => setPage(page - 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Sebelumnya</button>
                <button disabled={page >= data.meta.total_pages} onClick={() => setPage(page + 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Berikutnya</button>
              </div>
            </div>
          )}
        </>
      )}

      {showForm && (
        <SiswaFormDialog siswa={editingSiswa} onClose={() => setShowForm(false)} />
      )}

      {deleteConfirm !== null && (
        <ConfirmDialog
          message="Yakin ingin menghapus siswa ini? Data yang sudah dihapus tidak bisa dikembalikan."
          onConfirm={() => handleDelete(deleteConfirm)}
          onCancel={() => setDeleteConfirm(null)}
          loading={deleteMutation.isPending}
        />
      )}
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    aktif: 'bg-primary/10 text-primary',
    lulus: 'bg-muted text-muted-foreground',
    pindah: 'bg-accent text-accent-foreground',
    keluar: 'bg-destructive/10 text-destructive',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {status}
    </span>
  )
}

function SiswaFormDialog({ siswa, onClose }: { siswa: Siswa | null; onClose: () => void }) {
  const createMutation = useCreateSiswa()
  const updateMutation = useUpdateSiswa()
  const [error, setError] = useState('')

  const [form, setForm] = useState({
    nis: siswa?.nis || '',
    nama: siswa?.nama || '',
    jenis_kelamin: siswa?.jenis_kelamin || 'L',
    tempat_lahir: siswa?.tempat_lahir || '',
    tanggal_lahir: siswa?.tanggal_lahir || '',
    agama: siswa?.agama || '',
    alamat: siswa?.alamat || '',
    no_hp: siswa?.no_hp || '',
    email: siswa?.email || '',
    nama_ortu: siswa?.nama_ortu || '',
    no_hp_ortu: siswa?.no_hp_ortu || '',
  })

  function handleChange(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (siswa) {
        await updateMutation.mutateAsync({ id: siswa.id, data: form })
      } else {
        await createMutation.mutateAsync(form)
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
      <div className="w-full max-w-lg rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">{siswa ? 'Edit Siswa' : 'Tambah Siswa'}</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <FormField label="NIS" value={form.nis} onChange={(v) => handleChange('nis', v)} required />
            <FormField label="Nama" value={form.nama} onChange={(v) => handleChange('nama', v)} required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Jenis Kelamin</label>
              <select value={form.jenis_kelamin} onChange={(e) => handleChange('jenis_kelamin', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
                <option value="L">Laki-laki</option>
                <option value="P">Perempuan</option>
              </select>
            </div>
            <FormField label="Agama" value={form.agama} onChange={(v) => handleChange('agama', v)} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <FormField label="Tempat Lahir" value={form.tempat_lahir} onChange={(v) => handleChange('tempat_lahir', v)} />
            <FormField label="Tanggal Lahir" value={form.tanggal_lahir} onChange={(v) => handleChange('tanggal_lahir', v)} type="date" />
          </div>
          <FormField label="Alamat" value={form.alamat} onChange={(v) => handleChange('alamat', v)} />
          <div className="grid grid-cols-2 gap-3">
            <FormField label="No HP" value={form.no_hp} onChange={(v) => handleChange('no_hp', v)} />
            <FormField label="Email" value={form.email} onChange={(v) => handleChange('email', v)} type="email" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <FormField label="Nama Orangtua" value={form.nama_ortu} onChange={(v) => handleChange('nama_ortu', v)} />
            <FormField label="No HP Orangtua" value={form.no_hp_ortu} onChange={(v) => handleChange('no_hp_ortu', v)} />
          </div>

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

function FormField({ label, value, onChange, type = 'text', required = false }: { label: string; value: string; onChange: (v: string) => void; type?: string; required?: boolean }) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <input type={type} value={value} onChange={(e) => onChange(e.target.value)} required={required} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
    </div>
  )
}

function ConfirmDialog({ message, onConfirm, onCancel, loading }: { message: string; onConfirm: () => void; onCancel: () => void; loading: boolean }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-6 shadow-lg">
        <p className="text-sm text-card-foreground">{message}</p>
        <div className="mt-4 flex justify-end gap-2">
          <button onClick={onCancel} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
          <button onClick={onConfirm} disabled={loading} className="rounded-md bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:opacity-90 disabled:opacity-50">
            {loading ? 'Menghapus...' : 'Hapus'}
          </button>
        </div>
      </div>
    </div>
  )
}
