import { useState } from 'react'
import { usePendaftarList, useUpdatePendaftarStatus, useBerkasList, useVerifikasiBerkas, usePublishPengumuman } from '@/hooks/use-ppdb'
import { cn } from '@/lib/utils'
import { Search, Eye, X, FileText, Check } from 'lucide-react'
import type { Berkas } from '@/hooks/use-ppdb'

export function PpdbAdminPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [search, setSearch] = useState('')
  const [selectedId, setSelectedId] = useState<number | null>(null)

  const { data, isLoading } = usePendaftarList({ page, limit: 20, status: statusFilter || undefined, search: search || undefined })

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-bold text-foreground">PPDB - Pendaftar</h1>

      <div className="flex items-center gap-2 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input type="text" placeholder="Cari nama..." value={search} onChange={(e) => { setSearch(e.target.value); setPage(1) }} className="w-full rounded-md border border-input bg-background py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
        </div>
        <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
          <option value="">Semua Status</option>
          <option value="menunggu">Menunggu</option>
          <option value="berkas_lengkap">Berkas Lengkap</option>
          <option value="berkas_ditolak">Berkas Ditolak</option>
          <option value="diterima">Diterima</option>
          <option value="cadangan">Cadangan</option>
          <option value="tidak_diterima">Tidak Diterima</option>
          <option value="daftar_ulang">Daftar Ulang</option>
        </select>
      </div>

      {isLoading ? (
        <div className="text-center text-muted-foreground">Memuat...</div>
      ) : !data?.data?.length ? (
        <div className="text-center text-muted-foreground">Belum ada pendaftar.</div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="border-b border-border bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">No</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Nama</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">L/P</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Asal Sekolah</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((p) => (
                  <tr key={p.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-foreground">{p.id}</td>
                    <td className="px-4 py-3 text-foreground">{p.nama_lengkap}</td>
                    <td className="px-4 py-3 text-foreground">{p.jenis_kelamin}</td>
                    <td className="px-4 py-3 text-foreground">{p.asal_sekolah || '-'}</td>
                    <td className="px-4 py-3"><PpdbStatusBadge status={p.status} /></td>
                    <td className="px-4 py-3 text-right">
                      <button onClick={() => setSelectedId(p.id)} className="rounded p-1.5 hover:bg-muted" title="Detail">
                        <Eye className="h-4 w-4 text-muted-foreground" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.meta.total_pages > 1 && (
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>Halaman {data.meta.page} dari {data.meta.total_pages} ({data.meta.total} pendaftar)</span>
              <div className="flex gap-2">
                <button disabled={page <= 1} onClick={() => setPage(page - 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Sebelumnya</button>
                <button disabled={page >= data.meta.total_pages} onClick={() => setPage(page + 1)} className="rounded border border-border px-3 py-1 hover:bg-muted disabled:opacity-50">Berikutnya</button>
              </div>
            </div>
          )}
        </>
      )}

      {selectedId !== null && (
        <PendaftarDetailDialog id={selectedId} onClose={() => setSelectedId(null)} />
      )}
    </div>
  )
}

function PendaftarDetailDialog({ id, onClose }: { id: number; onClose: () => void }) {
  const { data: berkas } = useBerkasList(id)
  const updateStatus = useUpdatePendaftarStatus()
  const verifikasiBerkas = useVerifikasiBerkas()
  const publishPengumuman = usePublishPengumuman()
  const [newStatus, setNewStatus] = useState('')
  const [error, setError] = useState('')

  async function handleUpdateStatus() {
    if (!newStatus) return
    setError('')
    try {
      await updateStatus.mutateAsync({ id, status: newStatus })
      setNewStatus('')
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal update status')
    }
  }

  async function handleVerifBerkas(berkasId: number, status: string) {
    try {
      await verifikasiBerkas.mutateAsync({ id: berkasId, status })
    } catch { /* handled by query invalidation */ }
  }

  async function handlePublish() {
    setError('')
    try {
      await publishPengumuman.mutateAsync({ pendaftaran_id: id, status: 'diterima', keterangan: 'Selamat, Anda diterima!' })
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal publish pengumuman')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-foreground/20 p-4 pt-20">
      <div className="w-full max-w-lg rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">Detail Pendaftar #{id}</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-muted-foreground">Update Status</h3>
            <div className="flex gap-2">
              <select value={newStatus} onChange={(e) => setNewStatus(e.target.value)} className="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
                <option value="">Pilih status...</option>
                <option value="menunggu">Menunggu</option>
                <option value="berkas_lengkap">Berkas Lengkap</option>
                <option value="berkas_ditolak">Berkas Ditolak</option>
                <option value="diterima">Diterima</option>
                <option value="cadangan">Cadangan</option>
                <option value="tidak_diterima">Tidak Diterima</option>
                <option value="daftar_ulang">Daftar Ulang</option>
              </select>
              <button onClick={handleUpdateStatus} disabled={!newStatus || updateStatus.isPending} className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
                Update
              </button>
            </div>
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-medium text-muted-foreground">Berkas</h3>
            {!berkas?.length ? (
              <p className="text-sm text-muted-foreground">Belum ada berkas.</p>
            ) : (
              <div className="space-y-2">
                {berkas.map((b: Berkas) => (
                  <div key={b.id} className="flex items-center justify-between rounded-md border border-border p-2">
                    <div className="flex items-center gap-2">
                      <FileText className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-foreground">{b.jenis_berkas}</span>
                      <PpdbStatusBadge status={b.status} />
                    </div>
                    {b.status === 'pending' && (
                      <div className="flex gap-1">
                        <button onClick={() => handleVerifBerkas(b.id, 'diterima')} className="rounded bg-primary/10 p-1 text-primary hover:bg-primary/20">
                          <Check className="h-3.5 w-3.5" />
                        </button>
                        <button onClick={() => handleVerifBerkas(b.id, 'ditolak')} className="rounded bg-destructive/10 p-1 text-destructive hover:bg-destructive/20">
                          <X className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-medium text-muted-foreground">Pengumuman</h3>
            <button onClick={handlePublish} disabled={publishPengumuman.isPending} className="rounded-md border border-border px-3 py-2 text-sm hover:bg-muted disabled:opacity-50">
              Publish Pengumuman (Diterima)
            </button>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
      </div>
    </div>
  )
}

function PpdbStatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    menunggu: 'bg-muted text-muted-foreground',
    berkas_lengkap: 'bg-primary/10 text-primary',
    berkas_ditolak: 'bg-destructive/10 text-destructive',
    diterima: 'bg-primary/10 text-primary',
    cadangan: 'bg-accent text-accent-foreground',
    tidak_diterima: 'bg-destructive/10 text-destructive',
    daftar_ulang: 'bg-primary/10 text-primary',
    pending: 'bg-muted text-muted-foreground',
    ditolak: 'bg-destructive/10 text-destructive',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {status.replace(/_/g, ' ')}
    </span>
  )
}
