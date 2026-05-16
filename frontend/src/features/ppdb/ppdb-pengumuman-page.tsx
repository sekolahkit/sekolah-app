import { useState } from 'react'
import { usePengumuman } from '@/hooks/use-ppdb'
import { Search, CheckCircle, XCircle, Clock } from 'lucide-react'

export function PpdbPengumumanPage() {
  const [inputId, setInputId] = useState('')
  const [searchId, setSearchId] = useState<number | null>(null)
  const { data, isLoading, isError } = usePengumuman(searchId)

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    const id = Number(inputId)
    if (id > 0) setSearchId(id)
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-md space-y-6 rounded-xl border border-border bg-card p-6 shadow-sm">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-card-foreground">Pengumuman PPDB</h1>
          <p className="mt-1 text-sm text-muted-foreground">Masukkan nomor pendaftaran untuk melihat hasil</p>
        </div>

        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="number"
            value={inputId}
            onChange={(e) => setInputId(e.target.value)}
            placeholder="Nomor pendaftaran"
            className="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
          <button type="submit" className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90">
            <Search className="h-4 w-4" />
          </button>
        </form>

        {isLoading && <p className="text-center text-muted-foreground">Memuat...</p>}

        {isError && searchId && (
          <div className="rounded-lg border border-border bg-muted/50 p-4 text-center">
            <XCircle className="mx-auto h-8 w-8 text-muted-foreground" />
            <p className="mt-2 text-sm text-muted-foreground">Pengumuman belum tersedia untuk nomor ini.</p>
          </div>
        )}

        {data && (
          <div className="rounded-lg border border-border p-4 text-center space-y-3">
            {data.status === 'diterima' && <CheckCircle className="mx-auto h-10 w-10 text-primary" />}
            {data.status === 'tidak_diterima' && <XCircle className="mx-auto h-10 w-10 text-destructive" />}
            {data.status === 'cadangan' && <Clock className="mx-auto h-10 w-10 text-muted-foreground" />}

            <p className="text-lg font-bold text-card-foreground">
              {data.status === 'diterima' && 'Selamat! Anda Diterima'}
              {data.status === 'tidak_diterima' && 'Maaf, Anda Tidak Diterima'}
              {data.status === 'cadangan' && 'Anda Masuk Daftar Cadangan'}
            </p>

            {data.ranking > 0 && (
              <p className="text-sm text-muted-foreground">Ranking: {data.ranking}</p>
            )}
            {data.keterangan && (
              <p className="text-sm text-muted-foreground">{data.keterangan}</p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
