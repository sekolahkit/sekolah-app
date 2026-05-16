import { useState, useRef, useEffect } from 'react'
import { useSiswaList } from '@/hooks/use-siswa'
import { cn } from '@/lib/utils'
import { Search, X, Check } from 'lucide-react'
import type { Siswa } from '@/types'

interface SiswaPickerProps {
  value: number | null
  onChange: (siswa: Siswa | null) => void
  placeholder?: string
}

export function SiswaPicker({ value, onChange, placeholder = 'Pilih siswa...' }: SiswaPickerProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  const { data, isLoading } = useSiswaList({ search: search || undefined, limit: 10 })
  const selected = data?.data?.find((s) => s.id === value)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          'flex w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm',
          value ? 'text-foreground' : 'text-muted-foreground'
        )}
      >
        {selected ? `${selected.nis} - ${selected.nama}` : placeholder}
        {value ? (
          <X className="h-4 w-4 text-muted-foreground" onClick={(e) => { e.stopPropagation(); onChange(null) }} />
        ) : (
          <Search className="h-4 w-4 text-muted-foreground" />
        )}
      </button>

      {open && (
        <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg">
          <div className="p-2">
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari nama/NIS..."
              autoFocus
              className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <div className="max-h-48 overflow-y-auto">
            {isLoading ? (
              <div className="px-3 py-2 text-sm text-muted-foreground">Memuat...</div>
            ) : !data?.data?.length ? (
              <div className="px-3 py-2 text-sm text-muted-foreground">Tidak ditemukan</div>
            ) : (
              data.data.map((s) => (
                <button
                  key={s.id}
                  type="button"
                  onClick={() => { onChange(s); setOpen(false); setSearch('') }}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-muted',
                    s.id === value && 'bg-muted'
                  )}
                >
                  {s.id === value && <Check className="h-3.5 w-3.5 text-primary" />}
                  <span className="font-mono text-xs text-muted-foreground">{s.nis}</span>
                  <span className="text-foreground">{s.nama}</span>
                  <span className="ml-auto text-xs text-muted-foreground">{s.status}</span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}

interface SiswaMultiPickerProps {
  value: Siswa[]
  onChange: (siswa: Siswa[]) => void
  placeholder?: string
}

export function SiswaMultiPicker({ value, onChange, placeholder = 'Pilih siswa...' }: SiswaMultiPickerProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  const { data, isLoading } = useSiswaList({ search: search || undefined, limit: 20 })
  const selectedIds = new Set(value.map((s) => s.id))

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  function toggle(s: Siswa) {
    if (selectedIds.has(s.id)) {
      onChange(value.filter((v) => v.id !== s.id))
    } else {
      onChange([...value, s])
    }
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm text-muted-foreground"
      >
        {value.length > 0 ? `${value.length} siswa dipilih` : placeholder}
        <Search className="h-4 w-4" />
      </button>

      {value.length > 0 && (
        <div className="mt-1 flex flex-wrap gap-1">
          {value.map((s) => (
            <span key={s.id} className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              {s.nis} - {s.nama}
              <X className="h-3 w-3 cursor-pointer" onClick={() => onChange(value.filter((v) => v.id !== s.id))} />
            </span>
          ))}
        </div>
      )}

      {open && (
        <div className="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-lg">
          <div className="p-2">
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari nama/NIS..."
              autoFocus
              className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <div className="max-h-48 overflow-y-auto">
            {isLoading ? (
              <div className="px-3 py-2 text-sm text-muted-foreground">Memuat...</div>
            ) : !data?.data?.length ? (
              <div className="px-3 py-2 text-sm text-muted-foreground">Tidak ditemukan</div>
            ) : (
              data.data.map((s) => (
                <button
                  key={s.id}
                  type="button"
                  onClick={() => toggle(s)}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-muted',
                    selectedIds.has(s.id) && 'bg-muted'
                  )}
                >
                  <div className={cn('flex h-4 w-4 items-center justify-center rounded border', selectedIds.has(s.id) ? 'border-primary bg-primary' : 'border-input')}>
                    {selectedIds.has(s.id) && <Check className="h-3 w-3 text-primary-foreground" />}
                  </div>
                  <span className="font-mono text-xs text-muted-foreground">{s.nis}</span>
                  <span className="text-foreground">{s.nama}</span>
                  <span className="ml-auto text-xs text-muted-foreground">{s.status}</span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
