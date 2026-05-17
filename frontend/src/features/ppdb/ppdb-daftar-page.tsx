import { useState } from 'react'
import { useDaftar } from '@/hooks/use-ppdb'
import { useTahunAjaranAktif } from '@/hooks/use-tahun-ajaran'
import { CheckCircle, Upload } from 'lucide-react'

export function PpdbDaftarPage() {
  const daftarMutation = useDaftar()
  const { data: taAktif } = useTahunAjaranAktif()
  const [success, setSuccess] = useState<{ id: number } | null>(null)
  const [error, setError] = useState('')
  const [foto, setFoto] = useState<File | null>(null)
  const [berkas, setBerkas] = useState<File[]>([])
  const [uploading, setUploading] = useState(false)
  const [form, setForm] = useState({
    tahun_ajaran_id: '',
    nama_lengkap: '',
    nik: '',
    tempat_lahir: '',
    tanggal_lahir: '',
    jenis_kelamin: 'L',
    agama: '',
    alamat: '',
    asal_sekolah: '',
    no_hp: '',
    email: '',
    nama_ortu: '',
    no_hp_ortu: '',
    pekerjaan_ortu: '',
  })

  const effectiveForm = {
    ...form,
    tahun_ajaran_id: form.tahun_ajaran_id || (taAktif?.id ? String(taAktif.id) : ''),
  }

  function handleChange(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setUploading(true)
    try {
      let fotoPath: string | undefined
      if (foto) {
        const { uploadFile } = await import('@/lib/upload')
        fotoPath = await uploadFile(foto, 'foto_ppdb')
      }

      const result = await daftarMutation.mutateAsync({
        ...effectiveForm,
        tahun_ajaran_id: Number(effectiveForm.tahun_ajaran_id),
        foto: fotoPath,
      })

      if (berkas.length > 0) {
        const { uploadFile } = await import('@/lib/upload')
        for (const file of berkas) {
          await uploadFile(file, 'berkas_ppdb')
        }
      }

      setSuccess({ id: result.id })
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      setError(msg || 'Pendaftaran gagal')
    } finally {
      setUploading(false)
    }
  }

  if (success) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <div className="w-full max-w-md space-y-4 rounded-xl border border-border bg-card p-6 text-center shadow-sm">
          <CheckCircle className="mx-auto h-12 w-12 text-primary" />
          <h1 className="text-xl font-bold text-card-foreground">Pendaftaran Berhasil!</h1>
          <p className="text-muted-foreground">Nomor pendaftaran Anda:</p>
          <p className="text-3xl font-bold text-primary">{success.id}</p>
          <p className="text-sm text-muted-foreground">Simpan nomor ini untuk mengecek pengumuman.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-lg space-y-6 rounded-xl border border-border bg-card p-6 shadow-sm">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-card-foreground">Pendaftaran PPDB</h1>
          <p className="mt-1 text-sm text-muted-foreground">Isi formulir di bawah untuk mendaftar</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          <FormField label="Nama Lengkap" value={form.nama_lengkap} onChange={(v) => handleChange('nama_lengkap', v)} required />
          <div className="grid grid-cols-2 gap-3">
            <FormField label="NIK" value={form.nik} onChange={(v) => handleChange('nik', v)} />
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Jenis Kelamin</label>
              <select value={form.jenis_kelamin} onChange={(e) => handleChange('jenis_kelamin', e.target.value)} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
                <option value="L">Laki-laki</option>
                <option value="P">Perempuan</option>
              </select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <FormField label="Tempat Lahir" value={form.tempat_lahir} onChange={(v) => handleChange('tempat_lahir', v)} />
            <FormField label="Tanggal Lahir" value={form.tanggal_lahir} onChange={(v) => handleChange('tanggal_lahir', v)} type="date" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <FormField label="Agama" value={form.agama} onChange={(v) => handleChange('agama', v)} />
            <FormField label="Asal Sekolah" value={form.asal_sekolah} onChange={(v) => handleChange('asal_sekolah', v)} />
          </div>
          <FormField label="Alamat" value={form.alamat} onChange={(v) => handleChange('alamat', v)} />
          <div className="grid grid-cols-2 gap-3">
            <FormField label="No HP" value={form.no_hp} onChange={(v) => handleChange('no_hp', v)} />
            <FormField label="Email" value={form.email} onChange={(v) => handleChange('email', v)} type="email" />
          </div>
          <div className="grid grid-cols-3 gap-3">
            <FormField label="Nama Orangtua" value={form.nama_ortu} onChange={(v) => handleChange('nama_ortu', v)} />
            <FormField label="No HP Orangtua" value={form.no_hp_ortu} onChange={(v) => handleChange('no_hp_ortu', v)} />
            <FormField label="Pekerjaan Orangtua" value={form.pekerjaan_ortu} onChange={(v) => handleChange('pekerjaan_ortu', v)} />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Foto (JPG/PNG)</label>
            <input
              type="file"
              accept="image/jpeg,image/png"
              onChange={(e) => setFoto(e.target.files?.[0] || null)}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground file:mr-3 file:rounded file:border-0 file:bg-muted file:px-2 file:py-1 file:text-xs file:text-muted-foreground"
            />
            {foto && <p className="text-xs text-muted-foreground">{foto.name}</p>}
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Berkas (Ijazah, Akta, KK, dll)</label>
            <div className="flex items-center gap-2">
              <label className="flex cursor-pointer items-center gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm text-muted-foreground hover:bg-muted">
                <Upload className="h-4 w-4" />
                Pilih File
                <input
                  type="file"
                  accept="image/jpeg,image/png,application/pdf"
                  multiple
                  onChange={(e) => setBerkas(Array.from(e.target.files || []))}
                  className="hidden"
                />
              </label>
              {berkas.length > 0 && <span className="text-xs text-muted-foreground">{berkas.length} file dipilih</span>}
            </div>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <button type="submit" disabled={daftarMutation.isPending || uploading} className="w-full rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50">
            {uploading ? 'Mengupload berkas...' : daftarMutation.isPending ? 'Mengirim...' : 'Daftar'}
          </button>
        </form>
      </div>
    </div>
  )
}

function FormField({ label, value, onChange, type = 'text', required = false }: { label: string; value: string; onChange: (v: string) => void; type?: string; required?: boolean }) {
  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <input type={type} value={value} onChange={(e) => onChange(e.target.value)} required={required} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
    </div>
  )
}
