import { useState } from 'react'
import { useLinkedSiswa, useTagihanSelf, usePembayaranSelf, useCreatePembayaranSelf } from '@/hooks/use-selfservice'
import { useRekeningAktif } from '@/hooks/use-pembayaran'
import { cn } from '@/lib/utils'
import { User, FileText, CreditCard, X } from 'lucide-react'
import type { LinkedSiswa, TagihanSelf } from '@/hooks/use-selfservice'

export function SelfServicePage() {
  const { data: linked, isLoading } = useLinkedSiswa()
  const [selectedSiswa, setSelectedSiswa] = useState<LinkedSiswa | null>(null)
  const [tab, setTab] = useState<'tagihan' | 'pembayaran'>('tagihan')

  if (isLoading) {
    return <div className="p-6 text-muted-foreground">Memuat data...</div>
  }

  if (!linked?.length) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold text-foreground">Data Siswa</h1>
        <p className="mt-4 text-muted-foreground">Belum ada data siswa yang terhubung dengan akun Anda.</p>
      </div>
    )
  }

  const activeSiswa = selectedSiswa || linked[0]

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Data Siswa</h1>

      {linked.length > 1 && (
        <div className="mt-4 flex flex-wrap gap-2">
          {linked.map((s) => (
            <button
              key={s.id}
              onClick={() => setSelectedSiswa(s)}
              className={cn(
                'flex items-center gap-2 rounded-md border px-3 py-2 text-sm transition-colors',
                activeSiswa.id === s.id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:bg-muted'
              )}
            >
              <User className="h-4 w-4" />
              {s.nama}
              <span className="text-xs text-muted-foreground">({s.hubungan})</span>
            </button>
          ))}
        </div>
      )}

      <div className="mt-4 rounded-lg border border-border bg-card p-4">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
            <User className="h-5 w-5 text-primary" />
          </div>
          <div>
            <p className="font-medium text-card-foreground">{activeSiswa.nama}</p>
            <p className="text-sm text-muted-foreground">NIS: {activeSiswa.nis} &middot; {activeSiswa.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan'} &middot; {activeSiswa.hubungan}</p>
          </div>
        </div>
      </div>

      <div className="mt-6 flex gap-1 border-b border-border">
        <button
          onClick={() => setTab('tagihan')}
          className={cn('flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors', tab === 'tagihan' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground')}
        >
          <FileText className="h-4 w-4" /> Tagihan
        </button>
        <button
          onClick={() => setTab('pembayaran')}
          className={cn('flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors', tab === 'pembayaran' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground')}
        >
          <CreditCard className="h-4 w-4" /> Pembayaran
        </button>
      </div>

      {tab === 'tagihan' && <TagihanSection siswaId={activeSiswa.id} />}
      {tab === 'pembayaran' && <PembayaranSection siswaId={activeSiswa.id} />}
    </div>
  )
}

function TagihanSection({ siswaId }: { siswaId: number }) {
  const { data, isLoading } = useTagihanSelf(siswaId)
  const [payTagihan, setPayTagihan] = useState<TagihanSelf | null>(null)

  if (isLoading) return <div className="mt-4 text-muted-foreground">Memuat tagihan...</div>
  if (!data?.length) return <div className="mt-4 text-muted-foreground">Tidak ada tagihan.</div>

  return (
    <div className="mt-4">
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          <thead className="border-b border-border bg-muted/50">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Kategori</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Nominal</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Jatuh Tempo</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
              <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {data.map((t) => (
              <tr key={t.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-4 py-3 text-foreground">{t.kategori_nama || '-'}</td>
                <td className="px-4 py-3 text-foreground">{formatCurrency(t.nominal)}</td>
                <td className="px-4 py-3 text-foreground">{t.jatuh_tempo || '-'}</td>
                <td className="px-4 py-3"><TagihanStatusBadge status={t.status} /></td>
                <td className="px-4 py-3 text-right">
                  {t.status !== 'lunas' && (
                    <button onClick={() => setPayTagihan(t)} className="rounded-md border border-border px-2 py-1 text-xs hover:bg-muted">
                      Bayar
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {payTagihan && (
        <PaymentDialog tagihan={payTagihan} onClose={() => setPayTagihan(null)} />
      )}
    </div>
  )
}

function PembayaranSection({ siswaId }: { siswaId: number }) {
  const { data, isLoading } = usePembayaranSelf(siswaId)

  if (isLoading) return <div className="mt-4 text-muted-foreground">Memuat pembayaran...</div>
  if (!data?.length) return <div className="mt-4 text-muted-foreground">Belum ada riwayat pembayaran.</div>

  return (
    <div className="mt-4 overflow-x-auto rounded-lg border border-border">
      <table className="w-full text-sm">
        <thead className="border-b border-border bg-muted/50">
          <tr>
            <th className="px-4 py-3 text-left font-medium text-muted-foreground">Tanggal</th>
            <th className="px-4 py-3 text-left font-medium text-muted-foreground">Jumlah</th>
            <th className="px-4 py-3 text-left font-medium text-muted-foreground">Metode</th>
            <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
          </tr>
        </thead>
        <tbody>
          {data.map((p) => (
            <tr key={p.id} className="border-b border-border last:border-0 hover:bg-muted/30">
              <td className="px-4 py-3 text-foreground">{p.tanggal}</td>
              <td className="px-4 py-3 text-foreground">{formatCurrency(p.jumlah)}</td>
              <td className="px-4 py-3 text-foreground">{p.metode}</td>
              <td className="px-4 py-3"><PembayaranStatusBadge status={p.status} /></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function PaymentDialog({ tagihan, onClose }: { tagihan: TagihanSelf; onClose: () => void }) {
  const createMutation = useCreatePembayaranSelf()
  const { data: rekeningList } = useRekeningAktif()
  const [form, setForm] = useState({
    jumlah: String(tagihan.nominal),
    tanggal: new Date().toISOString().split('T')[0],
    metode: 'transfer',
    rekening_sekolah_id: '',
    catatan: '',
  })
  const [error, setError] = useState('')

  function handleChange(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await createMutation.mutateAsync({
        tagihan_id: tagihan.id,
        jumlah: Number(form.jumlah),
        tanggal: form.tanggal,
        metode: form.metode,
        rekening_sekolah_id: form.rekening_sekolah_id ? Number(form.rekening_sekolah_id) : undefined,
        catatan: form.catatan,
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
        <p className="mt-2 text-sm text-muted-foreground">{tagihan.kategori_nama} — {formatCurrency(tagihan.nominal)}</p>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Jumlah</label>
            <input type="number" value={form.jumlah} onChange={(e) => handleChange('jumlah', e.target.value)} required min="1" className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Tanggal</label>
            <input type="date" value={form.tanggal} onChange={(e) => handleChange('tanggal', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Metode</label>
            <select value={form.metode} onChange={(e) => handleChange('metode', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              <option value="transfer">Transfer</option>
              <option value="cash">Cash</option>
            </select>
          </div>
          {form.metode === 'transfer' && rekeningList?.length ? (
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Rekening Tujuan</label>
              <select value={form.rekening_sekolah_id} onChange={(e) => handleChange('rekening_sekolah_id', e.target.value)} required className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
                <option value="">Pilih rekening</option>
                {rekeningList.map((r) => (
                  <option key={r.id} value={r.id}>{r.nama_bank} - {r.nomor_rekening} ({r.nama_pemilik})</option>
                ))}
              </select>
            </div>
          ) : null}
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Catatan</label>
            <input type="text" value={form.catatan} onChange={(e) => handleChange('catatan', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
            <button type="submit" disabled={createMutation.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {createMutation.isPending ? 'Mengirim...' : 'Kirim Pembayaran'}
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
  const labels: Record<string, string> = { belum_bayar: 'Belum Bayar', sebagian: 'Sebagian', lunas: 'Lunas' }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {labels[status] || status}
    </span>
  )
}

function PembayaranStatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    pending: 'bg-accent text-accent-foreground',
    verified: 'bg-primary/10 text-primary',
    rejected: 'bg-destructive/10 text-destructive',
  }
  const labels: Record<string, string> = { pending: 'Pending', verified: 'Terverifikasi', rejected: 'Ditolak' }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {labels[status] || status}
    </span>
  )
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value)
}
