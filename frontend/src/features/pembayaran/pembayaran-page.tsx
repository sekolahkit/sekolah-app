import { useState } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { useTagihanList, useCreateTagihan, usePembayaranList, useVerifyPembayaran, useRejectPembayaran, useCreatePembayaran, useRekeningAktif } from '@/hooks/use-pembayaran'
import { cn } from '@/lib/utils'
import { Plus, Check, X, CreditCard, Search } from 'lucide-react'
import type { Tagihan, Pembayaran, Rekening } from '@/types'

export function PembayaranPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin' || user?.role === 'operator'

  return (
    <div className="p-6 space-y-8">
      {isAdmin ? <AdminPembayaranView /> : <SiswaPembayaranView />}
    </div>
  )
}

function AdminPembayaranView() {
  const [tab, setTab] = useState<'tagihan' | 'verifikasi'>('tagihan')

  return (
    <>
      <div className="flex items-center gap-4">
        <h1 className="text-2xl font-bold text-foreground">Pembayaran</h1>
        <div className="flex rounded-md border border-border">
          <button onClick={() => setTab('tagihan')} className={cn('px-3 py-1.5 text-sm', tab === 'tagihan' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted')}>Tagihan</button>
          <button onClick={() => setTab('verifikasi')} className={cn('px-3 py-1.5 text-sm', tab === 'verifikasi' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted')}>Verifikasi</button>
        </div>
      </div>
      {tab === 'tagihan' ? <TagihanSection /> : <VerifikasiSection />}
    </>
  )
}

function TagihanSection() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [search, setSearch] = useState('')
  const [showCreate, setShowCreate] = useState(false)

  const { data, isLoading } = useTagihanList({ page, limit: 20, status: statusFilter || undefined, search: search || undefined })

  return (
    <>
      <div className="flex items-center gap-2 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input type="text" placeholder="Cari siswa..." value={search} onChange={(e) => { setSearch(e.target.value); setPage(1) }} className="w-full rounded-md border border-input bg-background py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
        </div>
        <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
          <option value="">Semua Status</option>
          <option value="belum_bayar">Belum Bayar</option>
          <option value="sebagian">Sebagian</option>
          <option value="lunas">Lunas</option>
        </select>
        <button onClick={() => setShowCreate(true)} className="flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Plus className="h-4 w-4" /> Buat Tagihan
        </button>
      </div>

      {isLoading ? (
        <div className="text-center text-muted-foreground">Memuat...</div>
      ) : !data?.data?.length ? (
        <div className="text-center text-muted-foreground">Belum ada tagihan.</div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="border-b border-border bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Siswa</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Nominal</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Jatuh Tempo</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((t) => (
                  <tr key={t.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="px-4 py-3 text-foreground">ID: {t.siswa_id}</td>
                    <td className="px-4 py-3 text-foreground font-mono">{formatCurrency(t.nominal)}</td>
                    <td className="px-4 py-3 text-foreground">{t.jatuh_tempo || '-'}</td>
                    <td className="px-4 py-3"><TagihanStatusBadge status={t.status} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {data.meta.total_pages > 1 && (
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>Halaman {data.meta.page} dari {data.meta.total_pages}</span>
              <div className="flex gap-2">
                <button disabled={page <= 1} onClick={() => setPage(page - 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Sebelumnya</button>
                <button disabled={page >= data.meta.total_pages} onClick={() => setPage(page + 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Berikutnya</button>
              </div>
            </div>
          )}
        </>
      )}

      {showCreate && <CreateTagihanDialog onClose={() => setShowCreate(false)} />}
    </>
  )
}

function VerifikasiSection() {
  const { data, isLoading } = usePembayaranList({ status: 'pending' })
  const verifyMutation = useVerifyPembayaran()
  const rejectMutation = useRejectPembayaran()
  const [error, setError] = useState('')

  async function handleVerify(id: number) {
    setError('')
    try {
      await verifyMutation.mutateAsync(id)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal verifikasi')
    }
  }

  async function handleReject(id: number) {
    setError('')
    try {
      await rejectMutation.mutateAsync(id)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal menolak')
    }
  }

  if (isLoading) return <div className="text-muted-foreground">Memuat...</div>
  if (!data?.data?.length) return <div className="text-muted-foreground">Tidak ada pembayaran pending.</div>

  return (
    <>
      {error && <p className="text-sm text-destructive">{error}</p>}
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          <thead className="border-b border-border bg-muted/50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Siswa</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Jumlah</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Metode</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Tanggal</th>
              <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {data.data.map((p: Pembayaran) => (
              <tr key={p.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-4 py-3 text-foreground">ID: {p.siswa_id}</td>
                <td className="px-4 py-3 text-foreground font-mono">{formatCurrency(p.jumlah)}</td>
                <td className="px-4 py-3 text-foreground">{p.metode}</td>
                <td className="px-4 py-3 text-foreground">{p.tanggal?.split('T')[0]}</td>
                <td className="px-4 py-3 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <button onClick={() => handleVerify(p.id)} disabled={verifyMutation.isPending} className="rounded bg-primary/10 p-1.5 text-primary hover:bg-primary/20" title="Verifikasi">
                      <Check className="h-4 w-4" />
                    </button>
                    <button onClick={() => handleReject(p.id)} disabled={rejectMutation.isPending} className="rounded bg-destructive/10 p-1.5 text-destructive hover:bg-destructive/20" title="Tolak">
                      <X className="h-4 w-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}

function SiswaPembayaranView() {
  const { data: rekening } = useRekeningAktif()
  const { data: tagihan, isLoading } = useTagihanList({ status: 'belum_bayar' })
  const [showBayar, setShowBayar] = useState<Tagihan | null>(null)

  if (isLoading) return <div className="p-6 text-muted-foreground">Memuat tagihan...</div>

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground">Tagihan Saya</h1>

      {rekening && rekening.length > 0 && (
        <div className="mt-4 rounded-lg border border-border bg-card p-4">
          <h3 className="text-sm font-medium text-card-foreground">Rekening Tujuan Transfer</h3>
          <div className="mt-2 space-y-2">
            {rekening.map((r: Rekening) => (
              <div key={r.id} className="flex items-center gap-3 text-sm">
                <CreditCard className="h-4 w-4 text-muted-foreground" />
                <span className="text-foreground font-medium">{r.nama_bank}</span>
                <span className="font-mono text-foreground">{r.nomor_rekening}</span>
                <span className="text-muted-foreground">a.n. {r.nama_pemilik}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {!tagihan?.data?.length ? (
        <div className="mt-6 text-muted-foreground">Tidak ada tagihan yang belum dibayar.</div>
      ) : (
        <div className="mt-4 space-y-3">
          {tagihan.data.map((t) => (
            <div key={t.id} className="flex items-center justify-between rounded-lg border border-border bg-card p-4">
              <div>
                <p className="font-medium text-card-foreground">{formatCurrency(t.nominal)}</p>
                <p className="text-sm text-muted-foreground">Jatuh tempo: {t.jatuh_tempo || '-'}</p>
              </div>
              <button onClick={() => setShowBayar(t)} className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">Bayar</button>
            </div>
          ))}
        </div>
      )}

      {showBayar && <BayarDialog tagihan={showBayar} rekening={rekening || []} onClose={() => setShowBayar(null)} />}
    </div>
  )
}

function BayarDialog({ tagihan, rekening, onClose }: { tagihan: Tagihan; rekening: Rekening[]; onClose: () => void }) {
  const createPembayaran = useCreatePembayaran()
  const [metode, setMetode] = useState('transfer')
  const [rekeningId, setRekeningId] = useState(rekening[0]?.id || 0)
  const [jumlah, setJumlah] = useState(tagihan.nominal.toString())
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await createPembayaran.mutateAsync({
        tagihan_id: tagihan.id,
        siswa_id: tagihan.siswa_id,
        jumlah: Number(jumlah),
        tanggal: new Date().toISOString().split('T')[0],
        metode,
        rekening_sekolah_id: metode === 'transfer' ? rekeningId : undefined,
      })
      onClose()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal mengirim pembayaran')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">Bayar Tagihan</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <p className="mt-2 text-sm text-muted-foreground">Nominal tagihan: {formatCurrency(tagihan.nominal)}</p>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Jumlah Bayar</label>
            <input type="number" value={jumlah} onChange={(e) => setJumlah(e.target.value)} required min={1} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Metode</label>
            <select value={metode} onChange={(e) => setMetode(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              <option value="transfer">Transfer Bank</option>
            </select>
          </div>

          {metode === 'transfer' && rekening.length > 0 && (
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Rekening Tujuan</label>
              <select value={rekeningId} onChange={(e) => setRekeningId(Number(e.target.value))} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
                {rekening.map((r) => (
                  <option key={r.id} value={r.id}>{r.nama_bank} - {r.nomor_rekening} (a.n. {r.nama_pemilik})</option>
                ))}
              </select>
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
            <button type="submit" disabled={createPembayaran.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {createPembayaran.isPending ? 'Mengirim...' : 'Kirim Pembayaran'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function CreateTagihanDialog({ onClose }: { onClose: () => void }) {
  const createMutation = useCreateTagihan()
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    siswa_id: '',
    kategori_id: '',
    tahun_ajaran_id: '',
    nominal: '',
    jatuh_tempo: '',
    semester: '',
    catatan: '',
  })

  function handleChange(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await createMutation.mutateAsync({
        siswa_id: Number(form.siswa_id),
        kategori_id: Number(form.kategori_id),
        tahun_ajaran_id: Number(form.tahun_ajaran_id),
        nominal: Number(form.nominal),
        jatuh_tempo: form.jatuh_tempo,
        semester: form.semester,
        catatan: form.catatan,
      })
      onClose()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal membuat tagihan')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">Buat Tagihan</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Siswa ID</label>
              <input type="number" value={form.siswa_id} onChange={(e) => handleChange('siswa_id', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Kategori ID</label>
              <input type="number" value={form.kategori_id} onChange={(e) => handleChange('kategori_id', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Tahun Ajaran ID</label>
              <input type="number" value={form.tahun_ajaran_id} onChange={(e) => handleChange('tahun_ajaran_id', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Nominal</label>
              <input type="number" value={form.nominal} onChange={(e) => handleChange('nominal', e.target.value)} required min={1} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Jatuh Tempo</label>
              <input type="date" value={form.jatuh_tempo} onChange={(e) => handleChange('jatuh_tempo', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Semester</label>
              <select value={form.semester} onChange={(e) => handleChange('semester', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
                <option value="">-</option>
                <option value="Ganjil">Ganjil</option>
                <option value="Genap">Genap</option>
              </select>
            </div>
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Catatan</label>
            <input type="text" value={form.catatan} onChange={(e) => handleChange('catatan', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
            <button type="submit" disabled={createMutation.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {createMutation.isPending ? 'Membuat...' : 'Buat Tagihan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function TagihanStatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    belum_bayar: 'bg-destructive/10 text-destructive',
    sebagian: 'bg-accent text-accent-foreground',
    lunas: 'bg-primary/10 text-primary',
  }
  const labels: Record<string, string> = {
    belum_bayar: 'Belum Bayar',
    sebagian: 'Sebagian',
    lunas: 'Lunas',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {labels[status] || status}
    </span>
  )
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value)
}
