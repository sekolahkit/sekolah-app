import { useState } from 'react'
import { useGuruKelas, useGuruSiswaByKelas } from '@/hooks/use-selfservice'
import { cn } from '@/lib/utils'
import { Users, ChevronRight } from 'lucide-react'
import type { GuruKelas } from '@/hooks/use-selfservice'

export function GuruKelasPage() {
  const { data: kelasList, isLoading } = useGuruKelas()
  const [selectedKelas, setSelectedKelas] = useState<GuruKelas | null>(null)

  if (isLoading) {
    return <div className="p-6 text-muted-foreground">Memuat data...</div>
  }

  if (!kelasList?.length) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold text-foreground">Kelas Saya</h1>
        <p className="mt-4 text-muted-foreground">Anda belum ditugaskan sebagai wali kelas.</p>
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Kelas Saya</h1>

      <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {kelasList.map((k) => (
          <button
            key={k.id}
            onClick={() => setSelectedKelas(k)}
            className={cn(
              'flex items-center justify-between rounded-lg border p-4 text-left transition-colors',
              selectedKelas?.id === k.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'
            )}
          >
            <div>
              <p className="font-medium text-foreground">{k.nama}</p>
              <p className="text-sm text-muted-foreground">
                Tingkat {k.tingkat} {k.jurusan_nama ? `· ${k.jurusan_nama}` : ''} · {k.tahun_ajaran_nama}
              </p>
              <p className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                <Users className="h-3 w-3" /> {k.jumlah_siswa} siswa
              </p>
            </div>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          </button>
        ))}
      </div>

      {selectedKelas && <KelasSiswaSection kelas={selectedKelas} />}
    </div>
  )
}

function KelasSiswaSection({ kelas }: { kelas: GuruKelas }) {
  const { data, isLoading } = useGuruSiswaByKelas(kelas.id)

  return (
    <div className="mt-6">
      <h2 className="text-lg font-semibold text-foreground">Siswa — {kelas.nama}</h2>

      {isLoading ? (
        <div className="mt-4 text-muted-foreground">Memuat siswa...</div>
      ) : !data?.length ? (
        <div className="mt-4 text-muted-foreground">Belum ada siswa di kelas ini.</div>
      ) : (
        <div className="mt-4 overflow-x-auto rounded-lg border border-border">
          <table className="w-full text-sm">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">NIS</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Nama</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">L/P</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
              </tr>
            </thead>
            <tbody>
              {data.map((s) => (
                <tr key={s.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-mono text-foreground">{s.nis}</td>
                  <td className="px-4 py-3 text-foreground">{s.nama}</td>
                  <td className="px-4 py-3 text-foreground">{s.jenis_kelamin}</td>
                  <td className="px-4 py-3">
                    <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', s.status === 'aktif' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground')}>
                      {s.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
