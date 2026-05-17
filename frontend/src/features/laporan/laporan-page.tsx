import { useState } from 'react'
import { useRekapPembayaran, useRekapPPDB, useRekapSiswa } from '@/hooks/use-laporan'
import { useTahunAjaranList, useTahunAjaranAktif } from '@/hooks/use-tahun-ajaran'
import { cn } from '@/lib/utils'
import { BarChart3, Users, GraduationCap, Download, Loader2 } from 'lucide-react'
import { downloadExport } from '@/lib/export'

export function LaporanPage() {
  const [tab, setTab] = useState<'pembayaran' | 'ppdb' | 'siswa'>('pembayaran')

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-2xl font-bold text-foreground">Laporan</h1>

      <div className="flex rounded-md border border-border">
        <TabButton active={tab === 'pembayaran'} onClick={() => setTab('pembayaran')} icon={BarChart3} label="Pembayaran" />
        <TabButton active={tab === 'ppdb'} onClick={() => setTab('ppdb')} icon={GraduationCap} label="PPDB" />
        <TabButton active={tab === 'siswa'} onClick={() => setTab('siswa')} icon={Users} label="Siswa" />
      </div>

      {tab === 'pembayaran' && <LaporanPembayaran />}
      {tab === 'ppdb' && <LaporanPPDB />}
      {tab === 'siswa' && <LaporanSiswa />}
    </div>
  )
}

function TabButton({ active, onClick, icon: Icon, label }: { active: boolean; onClick: () => void; icon: typeof BarChart3; label: string }) {
  return (
    <button onClick={onClick} className={cn('flex items-center gap-2 px-4 py-2 text-sm', active ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted')}>
      <Icon className="h-4 w-4" />
      {label}
    </button>
  )
}

function LaporanPembayaran() {
  const today = new Date()
  const firstDay = new Date(today.getFullYear(), today.getMonth(), 1).toISOString().split('T')[0]
  const lastDay = today.toISOString().split('T')[0]

  const [tanggalMulai, setTanggalMulai] = useState(firstDay)
  const [tanggalSelesai, setTanggalSelesai] = useState(lastDay)
  const [tahunAjaranId, setTahunAjaranId] = useState<number | null>(null)
  const [exporting, setExporting] = useState(false)

  const { data: taList } = useTahunAjaranList()
  const { data: taAktif } = useTahunAjaranAktif()

  const effectiveTA = tahunAjaranId ?? taAktif?.id ?? 0
  const { data, isLoading } = useRekapPembayaran(tanggalMulai, tanggalSelesai, effectiveTA)

  const totalNominal = data?.reduce((sum, item) => sum + item.total_nominal, 0) ?? 0
  const totalTransaksi = data?.reduce((sum, item) => sum + item.total_transaksi, 0) ?? 0

  async function handleExport() {
    setExporting(true)
    try {
      let url = `/laporan/pembayaran/export?tanggal_mulai=${tanggalMulai}&tanggal_selesai=${tanggalSelesai}`
      if (effectiveTA) url += `&tahun_ajaran_id=${effectiveTA}`
      await downloadExport(url, 'laporan-pembayaran.xlsx')
    } catch { /* endpoint may not exist */ }
    setExporting(false)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 flex-wrap">
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Dari</label>
          <input type="date" value={tanggalMulai} onChange={(e) => setTanggalMulai(e.target.value)} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground" />
        </div>
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Sampai</label>
          <input type="date" value={tanggalSelesai} onChange={(e) => setTanggalSelesai(e.target.value)} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground" />
        </div>
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Tahun Ajaran</label>
          <select value={effectiveTA || ''} onChange={(e) => setTahunAjaranId(Number(e.target.value) || null)} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
            <option value="">Semua Tahun</option>
            {taList?.map((ta) => (
              <option key={ta.id} value={ta.id}>{ta.nama}{ta.aktif ? ' (Aktif)' : ''}</option>
            ))}
          </select>
        </div>
        <button onClick={handleExport} disabled={exporting} className="mt-auto flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm hover:bg-muted disabled:opacity-50">
          {exporting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
          Ekspor
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <StatCard label="Total Transaksi" value={totalTransaksi.toString()} />
        <StatCard label="Total Nominal" value={formatCurrency(totalNominal)} />
      </div>

      {isLoading ? (
        <div className="text-muted-foreground">Memuat...</div>
      ) : !data?.length ? (
        <div className="text-muted-foreground">Tidak ada data untuk periode ini.</div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="w-full text-sm">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Tanggal</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Metode</th>
                <th className="px-4 py-3 text-right font-medium text-muted-foreground">Transaksi</th>
                <th className="px-4 py-3 text-right font-medium text-muted-foreground">Nominal</th>
              </tr>
            </thead>
            <tbody>
              {data.map((item, i) => (
                <tr key={i} className="border-b border-border last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 text-foreground">{item.tanggal}</td>
                  <td className="px-4 py-3 text-foreground">{item.metode}</td>
                  <td className="px-4 py-3 text-right text-foreground">{item.total_transaksi}</td>
                  <td className="px-4 py-3 text-right font-mono text-foreground">{formatCurrency(item.total_nominal)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function LaporanPPDB() {
  const { data: taList } = useTahunAjaranList()
  const { data: taAktif } = useTahunAjaranAktif()
  const [tahunAjaranId, setTahunAjaranId] = useState<number | null>(null)
  const [exporting, setExporting] = useState(false)

  const activeId = tahunAjaranId ?? taAktif?.id ?? 0
  const { data, isLoading } = useRekapPPDB(activeId)

  async function handleExport() {
    if (!activeId) return
    setExporting(true)
    try {
      await downloadExport(`/laporan/ppdb/export?tahun_ajaran_id=${activeId}`, 'laporan-ppdb.xlsx')
    } catch { /* endpoint may not exist */ }
    setExporting(false)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 flex-wrap">
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Tahun Ajaran</label>
          <select value={activeId || ''} onChange={(e) => setTahunAjaranId(Number(e.target.value) || null)} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
            <option value="">Pilih tahun ajaran</option>
            {taList?.map((ta) => (
              <option key={ta.id} value={ta.id}>{ta.nama}{ta.aktif ? ' (Aktif)' : ''}</option>
            ))}
          </select>
        </div>
        <button onClick={handleExport} disabled={exporting} className="mt-auto flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm hover:bg-muted disabled:opacity-50">
          {exporting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
          Ekspor
        </button>
      </div>

      {isLoading ? (
        <div className="text-muted-foreground">Memuat...</div>
      ) : !data ? (
        <div className="text-muted-foreground">Tidak ada data.</div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <StatCard label="Total Pendaftar" value={data.total_pendaftar.toString()} />
          <StatCard label="Menunggu" value={data.menunggu.toString()} />
          <StatCard label="Diterima" value={data.diterima.toString()} />
          <StatCard label="Tidak Diterima" value={data.tidak_diterima.toString()} />
          <StatCard label="Cadangan" value={data.cadangan.toString()} />
          <StatCard label="Berkas Lengkap" value={data.berkas_lengkap.toString()} />
          <StatCard label="Daftar Ulang" value={data.daftar_ulang.toString()} />
        </div>
      )}
    </div>
  )
}

function LaporanSiswa() {
  const [tahunAjaranId, setTahunAjaranId] = useState<number | null>(null)
  const [exporting, setExporting] = useState(false)

  const { data: taList } = useTahunAjaranList()
  const { data: taAktif } = useTahunAjaranAktif()

  const effectiveTA = tahunAjaranId ?? taAktif?.id ?? 0
  const { data, isLoading } = useRekapSiswa(effectiveTA || undefined)

  async function handleExport() {
    setExporting(true)
    try {
      let url = '/laporan/siswa/export'
      if (effectiveTA) url += `?tahun_ajaran_id=${effectiveTA}`
      await downloadExport(url, 'laporan-siswa.xlsx')
    } catch { /* endpoint may not exist */ }
    setExporting(false)
  }

  if (isLoading) return <div className="text-muted-foreground">Memuat...</div>
  if (!data) return <div className="text-muted-foreground">Tidak ada data.</div>

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 flex-wrap">
        <div className="space-y-1">
          <label className="text-xs font-medium text-muted-foreground">Tahun Ajaran</label>
          <select value={effectiveTA || ''} onChange={(e) => setTahunAjaranId(Number(e.target.value) || null)} className="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
            <option value="">Semua Tahun</option>
            {taList?.map((ta) => (
              <option key={ta.id} value={ta.id}>{ta.nama}{ta.aktif ? ' (Aktif)' : ''}</option>
            ))}
          </select>
        </div>
        <button onClick={handleExport} disabled={exporting} className="mt-auto flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm hover:bg-muted disabled:opacity-50">
          {exporting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
          Ekspor
        </button>
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <StatCard label="Total Siswa" value={data.total.toString()} />
        <StatCard label="Aktif" value={data.aktif.toString()} />
        <StatCard label="Lulus" value={data.lulus.toString()} />
        <StatCard label="Pindah" value={data.pindah.toString()} />
        <StatCard label="Keluar" value={data.keluar.toString()} />
        <StatCard label="Laki-laki" value={data.laki_laki.toString()} />
        <StatCard label="Perempuan" value={data.perempuan.toString()} />
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-semibold text-card-foreground">{value}</p>
    </div>
  )
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(value)
}
