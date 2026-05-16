import { useTheme } from '@/hooks/use-theme-hook'
import { cn } from '@/lib/utils'
import type { ThemeMode } from '@/lib/theme'
import { Sun, Moon, Monitor } from 'lucide-react'

const modes: { value: ThemeMode; label: string; icon: typeof Sun }[] = [
  { value: 'light', label: 'Terang', icon: Sun },
  { value: 'dark', label: 'Gelap', icon: Moon },
  { value: 'system', label: 'Sistem', icon: Monitor },
]

export function PengaturanPage() {
  const { theme, setMode } = useTheme()

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-foreground">Pengaturan</h1>

      <div className="mt-6 max-w-lg space-y-6">
        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="font-medium text-card-foreground">Tema Tampilan</h2>
          <p className="mt-1 text-sm text-muted-foreground">Pilih mode tampilan yang nyaman untuk Anda.</p>

          <div className="mt-4 flex gap-2">
            {modes.map((m) => (
              <button
                key={m.value}
                onClick={() => setMode(m.value)}
                className={cn(
                  'flex flex-1 flex-col items-center gap-2 rounded-lg border p-3 text-sm transition-colors',
                  theme.mode === m.value
                    ? 'border-primary bg-primary/5 text-primary'
                    : 'border-border text-muted-foreground hover:border-primary/50'
                )}
              >
                <m.icon className="h-5 w-5" />
                {m.label}
              </button>
            ))}
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <h2 className="font-medium text-card-foreground">Preview</h2>
          <div className="mt-4 space-y-3">
            <div className="flex gap-2">
              <div className="h-8 w-8 rounded-md bg-primary" />
              <div className="h-8 w-8 rounded-md bg-secondary" />
              <div className="h-8 w-8 rounded-md bg-accent" />
              <div className="h-8 w-8 rounded-md bg-muted" />
              <div className="h-8 w-8 rounded-md bg-destructive" />
            </div>
            <div className="rounded-md border border-border bg-background p-3">
              <p className="text-sm text-foreground">Teks foreground</p>
              <p className="text-sm text-muted-foreground">Teks muted</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
