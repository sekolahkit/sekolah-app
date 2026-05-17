import { useState } from 'react'
import { useTheme } from '@/hooks/use-theme-hook'
import { useTahunAjaranList, useCreateTahunAjaran, useUpdateTahunAjaran, useSetTahunAjaranAktif } from '@/hooks/use-tahun-ajaran'
import { useJurusanList, useCreateJurusan, useUpdateJurusan, useDeleteJurusan } from '@/hooks/use-jurusan'
import { cn } from '@/lib/utils'
import type { ThemeMode } from '@/lib/theme'
import { Sun, Moon, Monitor, Plus, X, Check, Star, Trash2 } from 'lucide-react'
import type { TahunAjaran } from '@/hooks/use-tahun-ajaran'
import type { Jurusan } from '@/hooks/use-jurusan'

const modes: { value: ThemeMode; label: string; icon: typeof Sun }[] = [
  { value: 'light', label: 'Terang', icon: Sun },
  { value: 'dark', label: 'Gelap', icon: Moon },
  { value: 'system', label: 'Sistem', icon: Monitor },
]

export function PengaturanPage() {
  const { theme, setMode } = useTheme()

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Pengaturan</h1>

      <div className="max-w-2xl space-y-6">
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="font-medium text-card-foreground">Tema Tampilan</h2>
          <p className="mt-1 text-sm text-muted-foreground">Pilih mode tampilan yang nyaman untuk Anda.</p>
          <div className="mt-4 flex gap-2">
            {modes.map((m) => (
              <button
                key={m.value}
                onClick={() => setMode(m.value)}
                className={cn(
                  'flex flex-1 flex-col items-center gap-2 rounded-lg border p-3 text-sm transition-colors',
                  theme.mode === m.value
                    ? 'border-primary bg-primary/5 text-primary'
                    : 'border-border text-muted-foreground hover:border-primary/50'
                )}
              >
                <m.icon className="h-5 w-5" />
                {m.label}
              </button>
            ))}
          </div>
        </div>

        <TahunAjaranSection />
        <JurusanSection />
      </div>
    </div>
  )
}

function TahunAjaranSection() {
  const { data: list, isLoading } = useTahunAjaranList()
  const createMut = useCreateTahunAjaran()
  const updateMut = useUpdateTahunAjaran()
  const setAktifMut = useSetTahunAjaranAktif()
  const [showForm, setShowForm] = useState(false)
  const [editItem, setEditItem] = useState<TahunAjaran | null>(null)
  const [nama, setNama] = useState('')
  const [tglMulai, setTglMulai] = useState('')
  const [tglSelesai, setTglSelesai] = useState('')
  const [error, setError] = useState('')

  function openCreate() {
    setEditItem(null)
    setNama('')
    setTglMulai('')
    setTglSelesai('')
    setShowForm(true)
    setError('')
  }

  function openEdit(t: TahunAjaran) {
    setEditItem(t)
    setNama(t.nama)
    setTglMulai(t.tanggal_mulai)
    setTglSelesai(t.tanggal_selesai)
    setShowForm(true)
    setError('')
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (editItem) {
        await updateMut.mutateAsync({ id: editItem.id, nama, tanggal_mulai: tglMulai, tanggal_selesai: tglSelesai })
      } else {
        await createMut.mutateAsync({ nama, tanggal_mulai: tglMulai, tanggal_selesai: tglSelesai })
      }
      setShowForm(false)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal menyimpan')
    }
  }

  async function handleSetAktif(id: number) {
    await setAktifMut.mutateAsync(id)
  }

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-medium text-card-foreground">Tahun Ajaran</h2>
          <p className="mt-1 text-sm text-muted-foreground">Kelola tahun ajaran dan set yang aktif.</p>
        </div>
        <button onClick={openCreate} className="flex items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Plus className="h-4 w-4" /> Tambah
        </button>
      </div>

      {isLoading ? (
        <p className="mt-4 text-sm text-muted-foreground">Memuat...</p>
      ) : !list?.length ? (
        <p className="mt-4 text-sm text-muted-foreground">Belum ada tahun ajaran.</p>
      ) : (
        <div className="mt-4 space-y-2">
          {list.map((t) => (
            <div key={t.id} className="flex items-center justify-between rounded-md border border-border px-3 py-2">
              <div className="flex items-center gap-2">
                {t.aktif && <Check className="h-4 w-4 text-primary" />}
                <span className={cn('text-sm', t.aktif ? 'font-semibold text-foreground' : 'text-muted-foreground')}>{t.nama}</span>
                {t.tanggal_mulai && <span className="text-xs text-muted-foreground">({t.tanggal_mulai} — {t.tanggal_selesai})</span>}
              </div>
              <div className="flex items-center gap-1">
                {!t.aktif && (
                  <button onClick={() => handleSetAktif(t.id)} disabled={setAktifMut.isPending} className="rounded p-1.5 hover:bg-muted" title="Set Aktif">
                    <Star className="h-3.5 w-3.5 text-muted-foreground" />
                  </button>
                )}
                <button onClick={() => openEdit(t)} className="rounded p-1.5 hover:bg-muted text-xs text-muted-foreground">Edit</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
          <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-card-foreground">{editItem ? 'Edit' : 'Tambah'} Tahun Ajaran</h3>
              <button onClick={() => setShowForm(false)} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
            </div>
            <form onSubmit={handleSubmit} className="mt-4 space-y-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Nama</label>
                <input type="text" value={nama} onChange={(e) => setNama(e.target.value)} required placeholder="2025/2026" className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-xs font-medium text-muted-foreground">Tanggal Mulai</label>
                  <input type="date" value={tglMulai} onChange={(e) => setTglMulai(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium text-muted-foreground">Tanggal Selesai</label>
                  <input type="date" value={tglSelesai} onChange={(e) => setTglSelesai(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
                </div>
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setShowForm(false)} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
                <button type="submit" disabled={createMut.isPending || updateMut.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
                  {(createMut.isPending || updateMut.isPending) ? 'Menyimpan...' : 'Simpan'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

function JurusanSection() {
  const { data: list, isLoading } = useJurusanList()
  const createMut = useCreateJurusan()
  const updateMut = useUpdateJurusan()
  const deleteMut = useDeleteJurusan()
  const [showForm, setShowForm] = useState(false)
  const [editItem, setEditItem] = useState<Jurusan | null>(null)
  const [nama, setNama] = useState('')
  const [kode, setKode] = useState('')
  const [error, setError] = useState('')

  function openCreate() {
    setEditItem(null)
    setNama('')
    setKode('')
    setShowForm(true)
    setError('')
  }

  function openEdit(j: Jurusan) {
    setEditItem(j)
    setNama(j.nama)
    setKode(j.kode)
    setShowForm(true)
    setError('')
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (editItem) {
        await updateMut.mutateAsync({ id: editItem.id, nama, kode })
      } else {
        await createMut.mutateAsync({ nama, kode })
      }
      setShowForm(false)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal menyimpan')
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Hapus jurusan ini?')) return
    try {
      await deleteMut.mutateAsync(id)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      alert(msg || 'Gagal menghapus')
    }
  }

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-medium text-card-foreground">Jurusan</h2>
          <p className="mt-1 text-sm text-muted-foreground">Kelola jurusan untuk SMK.</p>
        </div>
        <button onClick={openCreate} className="flex items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Plus className="h-4 w-4" /> Tambah
        </button>
      </div>

      {isLoading ? (
        <p className="mt-4 text-sm text-muted-foreground">Memuat...</p>
      ) : !list?.length ? (
        <p className="mt-4 text-sm text-muted-foreground">Belum ada jurusan.</p>
      ) : (
        <div className="mt-4 space-y-2">
          {list.map((j) => (
            <div key={j.id} className="flex items-center justify-between rounded-md border border-border px-3 py-2">
              <div className="flex items-center gap-2">
                {j.kode && <span className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono text-muted-foreground">{j.kode}</span>}
                <span className="text-sm text-foreground">{j.nama}</span>
              </div>
              <div className="flex items-center gap-1">
                <button onClick={() => openEdit(j)} className="rounded p-1.5 hover:bg-muted text-xs text-muted-foreground">Edit</button>
                <button onClick={() => handleDelete(j.id)} disabled={deleteMut.isPending} className="rounded p-1.5 hover:bg-muted" title="Hapus">
                  <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
          <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-card-foreground">{editItem ? 'Edit' : 'Tambah'} Jurusan</h3>
              <button onClick={() => setShowForm(false)} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
            </div>
            <form onSubmit={handleSubmit} className="mt-4 space-y-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Nama</label>
                <input type="text" value={nama} onChange={(e) => setNama(e.target.value)} required placeholder="Teknik Informatika" className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Kode</label>
                <input type="text" value={kode} onChange={(e) => setKode(e.target.value)} placeholder="TI" className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setShowForm(false)} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
                <button type="submit" disabled={createMut.isPending || updateMut.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
                  {(createMut.isPending || updateMut.isPending) ? 'Menyimpan...' : 'Simpan'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
