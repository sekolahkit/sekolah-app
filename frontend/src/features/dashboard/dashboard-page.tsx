import { useQuery } from '@tanstack/react-query'
import api from '@/lib/api'
import { useAuth } from '@/hooks/use-auth-hook'
import { useDashboardSiswa, useDashboardOrangtua, useDashboardGuru } from '@/hooks/use-selfservice'
import type { DashboardStats } from '@/types'

export function DashboardPage() {
  const { user } = useAuth()

  if (user?.role === 'siswa') return <SiswaDashboard />
  if (user?.role === 'orangtua') return <OrangtuaDashboard />
  if (user?.role === 'guru') return <GuruDashboard />
  return <AdminDashboard />
}

function AdminDashboard() {
  const { data, isLoading } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const res = await api.get<{ data: DashboardStats }>('/dashboard/admin')
      return res.data.data
    },
  })

  if (isLoading) {
    return <div className="p-6 text-muted-foreground">Memuat dashboard...</div>
  }

  const stats = data

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
      <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard label="Siswa Aktif" value={stats?.total_siswa_aktif ?? 0} />
        <StatCard label="Pembayaran Bulan Ini" value={formatCurrency(stats?.total_pembayaran_bulan_ini ?? 0)} />
        <StatCard label="Pendaftar PPDB Baru" value={stats?.pendaftar_ppdb_baru ?? 0} />
        <StatCard label="Tagihan Jatuh Tempo" value={stats?.tagihan_jatuh_tempo ?? 0} />
        <StatCard label="Pembayaran Pending" value={stats?.pembayaran_pending ?? 0} />
        <StatCard label="Total Kelas" value={stats?.total_kelas ?? 0} />
      </div>
    </div>
  )
}

function SiswaDashboard() {
  const { data, isLoading } = useDashboardSiswa()

  if (isLoading) {
    return <div className="p-6 text-muted-foreground">Memuat dashboard...</div>
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Dashboard Siswa</h1>
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <StatCard label="Total Tagihan" value={data?.total_tagihan ?? 0} />
        <StatCard label="Belum Bayar" value={data?.tagihan_belum_bayar ?? 0} />
        <StatCard label="Total Terbayar" value={formatCurrency(data?.total_terbayar ?? 0)} />
        <StatCard label="Pembayaran Pending" value={data?.pembayaran_pending ?? 0} />
      </div>
    </div>
  )
}

function OrangtuaDashboard() {
  const { data, isLoading } = useDashboardOrangtua()

  if (isLoading) {
    return <div className="p-6 text-muted-foreground">Memuat dashboard...</div>
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Dashboard Orangtua</h1>
      <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard label="Jumlah Anak" value={data?.jumlah_anak ?? 0} />
        <StatCard label="Total Tagihan" value={data?.total_tagihan ?? 0} />
        <StatCard label="Belum Bayar" value={data?.tagihan_belum_bayar ?? 0} />
        <StatCard label="Total Terbayar" value={formatCurrency(data?.total_terbayar ?? 0)} />
        <StatCard label="Pembayaran Pending" value={data?.pembayaran_pending ?? 0} />
      </div>
    </div>
  )
}

function GuruDashboard() {
  const { data, isLoading } = useDashboardGuru()

  if (isLoading) {
    return <div className="p-6 text-muted-foreground">Memuat dashboard...</div>
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Dashboard Guru</h1>
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <StatCard label="Kelas Wali" value={data?.total_kelas ?? 0} />
        <StatCard label="Total Siswa" value={data?.total_siswa ?? 0} />
      </div>
      {data?.total_kelas === 0 && (
        <p className="mt-4 text-sm text-muted-foreground">Anda belum ditugaskan sebagai wali kelas.</p>
      )}
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string | number }) {
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
