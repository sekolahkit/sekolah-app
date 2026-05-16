import { useState } from 'react'
import { useNotifikasiList, useQueueStats, useTestSend } from '@/hooks/use-notifikasi'
import { cn } from '@/lib/utils'
import { Send, AlertCircle, CheckCircle, Clock, X } from 'lucide-react'
import type { Notifikasi } from '@/hooks/use-notifikasi'

export function NotifikasiPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [tipeFilter, setTipeFilter] = useState('')
  const [showTestForm, setShowTestForm] = useState(false)

  const { data: stats } = useQueueStats()
  const { data, isLoading } = useNotifikasiList({
    page,
    limit: 20,
    status: statusFilter || undefined,
    tipe: tipeFilter || undefined,
  })

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Notifikasi</h1>
        <button onClick={() => setShowTestForm(true)} className="flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Send className="h-4 w-4" /> Test Kirim
        </button>
      </div>

      {stats && (
        <div className="grid grid-cols-3 gap-4">
          <StatCard icon={Clock} label="Pending" value={stats.pending} className="text-muted-foreground" />
          <StatCard icon={CheckCircle} label="Terkirim" value={stats.sent} className="text-primary" />
          <StatCard icon={AlertCircle} label="Gagal" value={stats.failed} className="text-destructive" />
        </div>
      )}

      <div className="flex items-center gap-2 flex-wrap">
        <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
          <option value="">Semua Status</option>
          <option value="pending">Pending</option>
          <option value="sent">Terkirim</option>
          <option value="failed">Gagal</option>
        </select>
        <select value={tipeFilter} onChange={(e) => { setTipeFilter(e.target.value); setPage(1) }} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
          <option value="">Semua Tipe</option>
          <option value="whatsapp">WhatsApp</option>
          <option value="telegram">Telegram</option>
          <option value="email">Email</option>
        </select>
      </div>

      {isLoading ? (
        <div className="text-center text-muted-foreground">Memuat...</div>
      ) : !data?.data?.length ? (
        <div className="text-center text-muted-foreground">Belum ada notifikasi.</div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="border-b border-border bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Tipe</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Penerima</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Pesan</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Retry</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Waktu</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((n: Notifikasi) => (
                  <tr key={n.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                    <td className="px-4 py-3"><TipeBadge tipe={n.tipe} /></td>
                    <td className="px-4 py-3 text-foreground font-mono text-xs">{n.penerima}</td>
                    <td className="px-4 py-3 text-foreground max-w-[200px] truncate">{n.pesan}</td>
                    <td className="px-4 py-3"><StatusBadge status={n.status} /></td>
                    <td className="px-4 py-3 text-foreground">
                      {n.retry_count > 0 && (
                        <span className="text-xs text-muted-foreground">{n.retry_count}/{n.max_retries}</span>
                      )}
                      {n.last_error && (
                        <p className="text-xs text-destructive truncate max-w-[150px]" title={n.last_error}>{n.last_error}</p>
                      )}
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {n.sent_at ? n.sent_at.split('T')[0] : n.created_at?.split('T')[0] || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.meta && data.meta.total_pages > 1 && (
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

      {showTestForm && <TestSendDialog onClose={() => setShowTestForm(false)} />}
    </div>
  )
}

function TestSendDialog({ onClose }: { onClose: () => void }) {
  const testSend = useTestSend()
  const [tipe, setTipe] = useState('whatsapp')
  const [penerima, setPenerima] = useState('')
  const [pesan, setPesan] = useState('')
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await testSend.mutateAsync({ tipe, penerima, pesan })
      onClose()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal mengirim notifikasi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">Test Kirim Notifikasi</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Tipe</label>
            <select value={tipe} onChange={(e) => setTipe(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              <option value="whatsapp">WhatsApp</option>
              <option value="telegram">Telegram</option>
              <option value="email">Email</option>
            </select>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Penerima</label>
            <input type="text" value={penerima} onChange={(e) => setPenerima(e.target.value)} required placeholder="081234567890 / @username / email" className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Pesan</label>
            <textarea value={pesan} onChange={(e) => setPesan(e.target.value)} required rows={3} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none" />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
            <button type="submit" disabled={testSend.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {testSend.isPending ? 'Mengirim...' : 'Kirim'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function StatCard({ icon: Icon, label, value, className }: { icon: typeof Clock; label: string; value: number; className?: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-2">
        <Icon className={cn('h-4 w-4', className)} />
        <span className="text-sm text-muted-foreground">{label}</span>
      </div>
      <p className="mt-1 text-2xl font-semibold text-card-foreground">{value}</p>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    pending: 'bg-muted text-muted-foreground',
    sent: 'bg-primary/10 text-primary',
    failed: 'bg-destructive/10 text-destructive',
  }
  const labels: Record<string, string> = {
    pending: 'Pending',
    sent: 'Terkirim',
    failed: 'Gagal',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {labels[status] || status}
    </span>
  )
}

function TipeBadge({ tipe }: { tipe: string }) {
  const styles: Record<string, string> = {
    whatsapp: 'bg-primary/10 text-primary',
    telegram: 'bg-accent text-accent-foreground',
    email: 'bg-secondary text-secondary-foreground',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[tipe] || 'bg-muted text-muted-foreground')}>
      {tipe}
    </span>
  )
}
