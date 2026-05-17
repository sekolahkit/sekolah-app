import { useState } from 'react'
import { useNotifikasiList, useQueueStats, useTestSend, useRetryNotifikasi } from '@/hooks/use-notifikasi'
import { usePreferensiList, useUpsertPreferensi, useGenerateTelegramInvite } from '@/hooks/use-notifikasi-preferensi'
import { useWhatsAppStatus, useWhatsAppConnect, useWhatsAppDisconnect } from '@/hooks/use-whatsapp'
import { cn } from '@/lib/utils'
import { Send, AlertCircle, CheckCircle, Clock, X, RotateCcw, Shield, Wifi, WifiOff } from 'lucide-react'
import type { Notifikasi } from '@/hooks/use-notifikasi'
import type { NotifikasiPreferensi } from '@/hooks/use-notifikasi-preferensi'

type Tab = 'queue' | 'preferensi'

export function NotifikasiPage() {
  const [tab, setTab] = useState<Tab>('queue')

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Notifikasi</h1>
        <div className="flex items-center gap-1 rounded-lg border border-border p-0.5">
          <button onClick={() => setTab('queue')} className={cn('rounded-md px-3 py-1.5 text-sm font-medium transition-colors', tab === 'queue' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>Queue</button>
          <button onClick={() => setTab('preferensi')} className={cn('rounded-md px-3 py-1.5 text-sm font-medium transition-colors', tab === 'preferensi' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>Preferensi</button>
        </div>
      </div>

      <WhatsAppStatusCard />

      {tab === 'queue' ? <QueueTab /> : <PreferensiTab />}
    </div>
  )
}

function QueueTab() {
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
    <div className="space-y-6">
      <div className="flex items-center justify-end">
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
                  <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((n: Notifikasi) => (
                  <NotifikasiRow key={n.id} notifikasi={n} />
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

function PreferensiTab() {
  const [page, setPage] = useState(1)
  const [channelFilter, setChannelFilter] = useState('')
  const [consentFilter, setConsentFilter] = useState('')
  const [showForm, setShowForm] = useState(false)

  const { data, isLoading } = usePreferensiList({
    page,
    limit: 20,
    channel: channelFilter || undefined,
    consent_status: consentFilter || undefined,
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">Kelola persetujuan penerima notifikasi per kanal.</p>
        <button onClick={() => setShowForm(true)} className="flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">
          <Shield className="h-4 w-4" /> Tambah Preferensi
        </button>
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <select value={channelFilter} onChange={(e) => { setChannelFilter(e.target.value); setPage(1) }} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
          <option value="">Semua Kanal</option>
          <option value="email">Email</option>
          <option value="whatsapp">WhatsApp</option>
          <option value="telegram">Telegram</option>
        </select>
        <select value={consentFilter} onChange={(e) => { setConsentFilter(e.target.value); setPage(1) }} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
          <option value="">Semua Status</option>
          <option value="granted">Granted</option>
          <option value="pending">Pending</option>
          <option value="revoked">Revoked</option>
        </select>
      </div>

      {isLoading ? (
        <div className="text-center text-muted-foreground">Memuat...</div>
      ) : !data?.data?.length ? (
        <div className="text-center text-muted-foreground">Belum ada preferensi.</div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead className="border-b border-border bg-muted/50">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Kanal</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Tujuan</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Tipe</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Konsen</th>
                  <th className="px-4 py-3 text-left font-medium text-muted-foreground">Sumber</th>
                  <th className="px-4 py-3 text-right font-medium text-muted-foreground">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {data.data.map((p: NotifikasiPreferensi) => (
                  <PreferensiRow key={p.id} preferensi={p} />
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

      {showForm && <PreferensiForm onClose={() => setShowForm(false)} />}
    </div>
  )
}

function PreferensiRow({ preferensi: p }: { preferensi: NotifikasiPreferensi }) {
  const upsert = useUpsertPreferensi()
  const genInvite = useGenerateTelegramInvite()
  const [inviteLink, setInviteLink] = useState('')

  function toggleEnabled() {
    upsert.mutate({
      channel: p.channel,
      destination: p.destination,
      recipient_type: p.recipient_type,
      enabled: !p.enabled,
      consent_status: p.consent_status,
      consent_source: p.consent_source,
    })
  }

  function toggleConsent() {
    const next = p.consent_status === 'granted' ? 'revoked' : 'granted'
    upsert.mutate({
      channel: p.channel,
      destination: p.destination,
      recipient_type: p.recipient_type,
      enabled: p.enabled,
      consent_status: next,
      consent_source: 'admin',
    })
  }

  async function handleGenerateInvite() {
    try {
      const result = await genInvite.mutateAsync(p.id)
      setInviteLink(result.invite_link)
    } catch {
      setInviteLink('')
    }
  }

  function copyLink() {
    if (inviteLink) navigator.clipboard.writeText(inviteLink)
  }

  return (
    <>
    <tr className="border-b border-border last:border-0 hover:bg-muted/30">
      <td className="px-4 py-3"><ChannelBadge channel={p.channel} /></td>
      <td className="px-4 py-3 text-foreground font-mono text-xs">{p.destination}</td>
      <td className="px-4 py-3 text-xs text-muted-foreground">{p.recipient_type}</td>
      <td className="px-4 py-3">
        <button onClick={toggleEnabled} disabled={upsert.isPending} className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium cursor-pointer', p.enabled ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground')}>
          {p.enabled ? 'Aktif' : 'Nonaktif'}
        </button>
      </td>
      <td className="px-4 py-3">
        <ConsentBadge status={p.consent_status} />
      </td>
      <td className="px-4 py-3 text-xs text-muted-foreground">{p.consent_source}</td>
      <td className="px-4 py-3 text-right">
        <div className="flex items-center justify-end gap-1">
          {p.channel === 'telegram' && (
            <button onClick={handleGenerateInvite} disabled={genInvite.isPending} className="rounded bg-muted p-1.5 hover:bg-accent" title="Buat Link Undangan Telegram">
              <Send className="h-3.5 w-3.5 text-muted-foreground" />
            </button>
          )}
          <button onClick={toggleConsent} disabled={upsert.isPending} className="rounded bg-muted p-1.5 hover:bg-accent" title={p.consent_status === 'granted' ? 'Revoke' : 'Grant'}>
            <Shield className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
      </td>
    </tr>
    {inviteLink && (
      <tr>
        <td colSpan={7} className="px-4 py-2 bg-muted/20">
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Undangan:</span>
            <code className="flex-1 rounded bg-muted px-2 py-1 text-xs text-foreground break-all">{inviteLink}</code>
            <button onClick={copyLink} className="rounded bg-muted px-2 py-1 text-xs hover:bg-accent">Salin</button>
            <button onClick={() => setInviteLink('')} className="rounded bg-muted px-2 py-1 text-xs hover:bg-accent">Tutup</button>
          </div>
        </td>
      </tr>
    )}
    </>
  )
}

function PreferensiForm({ onClose }: { onClose: () => void }) {
  const upsert = useUpsertPreferensi()
  const [channel, setChannel] = useState('email')
  const [destination, setDestination] = useState('')
  const [recipientType, setRecipientType] = useState('manual')
  const [consentStatus, setConsentStatus] = useState('granted')
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await upsert.mutateAsync({
        channel,
        destination,
        recipient_type: recipientType,
        consent_status: consentStatus,
        consent_source: 'admin',
      })
      onClose()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Gagal menyimpan preferensi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 p-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-card-foreground">Tambah Preferensi</h2>
          <button onClick={onClose} className="rounded p-1 hover:bg-muted"><X className="h-5 w-5" /></button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Kanal</label>
            <select value={channel} onChange={(e) => setChannel(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              <option value="email">Email</option>
              <option value="whatsapp">WhatsApp</option>
              <option value="telegram">Telegram</option>
            </select>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Tujuan</label>
            <input type="text" value={destination} onChange={(e) => setDestination(e.target.value)} required placeholder="email@example.com / 081234567890" className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Tipe Penerima</label>
            <select value={recipientType} onChange={(e) => setRecipientType(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              <option value="manual">Manual</option>
              <option value="pengguna">Pengguna</option>
              <option value="siswa">Siswa</option>
              <option value="calon_siswa">Calon Siswa</option>
            </select>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Persetujuan</label>
            <select value={consentStatus} onChange={(e) => setConsentStatus(e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
              <option value="granted">Granted</option>
              <option value="pending">Pending</option>
              <option value="revoked">Revoked</option>
            </select>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className="rounded-md border border-border px-4 py-2 text-sm hover:bg-muted">Batal</button>
            <button type="submit" disabled={upsert.isPending} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
              {upsert.isPending ? 'Menyimpan...' : 'Simpan'}
            </button>
          </div>
        </form>
      </div>
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

function NotifikasiRow({ notifikasi: n }: { notifikasi: Notifikasi }) {
  const retryMutation = useRetryNotifikasi()

  return (
    <tr className="border-b border-border last:border-0 hover:bg-muted/30">
      <td className="px-4 py-3"><ChannelBadge channel={n.tipe} /></td>
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
      <td className="px-4 py-3 text-right">
        {n.status === 'failed' && (
          <button
            onClick={() => retryMutation.mutate(n.id)}
            disabled={retryMutation.isPending}
            className="rounded bg-muted p-1.5 hover:bg-accent" title="Retry"
          >
            <RotateCcw className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        )}
      </td>
    </tr>
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

function ChannelBadge({ channel }: { channel: string }) {
  const styles: Record<string, string> = {
    whatsapp: 'bg-primary/10 text-primary',
    telegram: 'bg-accent text-accent-foreground',
    email: 'bg-secondary text-secondary-foreground',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[channel] || 'bg-muted text-muted-foreground')}>
      {channel}
    </span>
  )
}

function ConsentBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    granted: 'bg-primary/10 text-primary',
    pending: 'bg-muted text-muted-foreground',
    revoked: 'bg-destructive/10 text-destructive',
  }
  return (
    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', styles[status] || 'bg-muted text-muted-foreground')}>
      {status}
    </span>
  )
}

function WhatsAppStatusCard() {
  const { data: waStatus } = useWhatsAppStatus()
  const connect = useWhatsAppConnect()
  const disconnect = useWhatsAppDisconnect()

  if (!waStatus) return null

  const isConnected = waStatus.connected && waStatus.status === 'connected'
  const isWaitingQR = waStatus.status === 'waiting_qr'

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {isConnected ? (
            <Wifi className="h-5 w-5 text-primary" />
          ) : (
            <WifiOff className="h-5 w-5 text-muted-foreground" />
          )}
          <div>
            <p className="text-sm font-medium text-card-foreground">
              WhatsApp: {isConnected ? 'Terhubung' : isWaitingQR ? 'Menunggu QR' : 'Terputus'}
            </p>
            {waStatus.last_error && (
              <p className="text-xs text-destructive mt-0.5">{waStatus.last_error}</p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isConnected ? (
            <button
              onClick={() => disconnect.mutate()}
              disabled={disconnect.isPending}
              className="rounded-md border border-border px-3 py-1.5 text-xs font-medium hover:bg-muted disabled:opacity-50"
            >
              Putus
            </button>
          ) : (
            <button
              onClick={() => connect.mutate()}
              disabled={connect.isPending}
              className="rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50"
            >
              {connect.isPending ? 'Menghubungkan...' : 'Hubungkan'}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
