export interface User {
  id: number
  sekolah_id: number
  email: string
  nama: string
  role: 'admin' | 'operator' | 'guru' | 'siswa' | 'orangtua'
}

export interface Sekolah {
  id: number
  nama: string
  kode: string
  alamat: string
  telepon: string
  email: string
  logo: string
  website: string
}

export interface Siswa {
  id: number
  sekolah_id: number
  nis: string
  nama: string
  jenis_kelamin: string
  tempat_lahir: string
  tanggal_lahir: string
  agama: string
  alamat: string
  no_hp: string
  email: string
  nama_ortu: string
  no_hp_ortu: string
  email_ortu: string
  tahun_ajaran_masuk: number
  status: string
  created_at: string
  updated_at: string
}

export interface Tagihan {
  id: number
  sekolah_id: number
  siswa_id: number
  kategori_id: number
  tahun_ajaran_id: number
  semester: string
  nominal: number
  jatuh_tempo: string
  status: 'belum_bayar' | 'sebagian' | 'lunas'
  catatan: string
  created_at: string
  updated_at: string
}

export interface Pembayaran {
  id: number
  tagihan_id: number
  siswa_id: number
  jumlah: number
  tanggal: string
  metode: string
  provider: string
  bukti_bayar: string
  rekening_sekolah_id: number
  status: 'pending' | 'verified' | 'rejected'
  catatan: string
  verified_by: number
  verified_at: string
  created_at: string
}

export interface Rekening {
  id: number
  sekolah_id: number
  nama_bank: string
  nomor_rekening: string
  nama_pemilik: string
  cabang: string
  aktif: boolean
  urutan: number
  catatan: string
}

export interface DashboardStats {
  total_siswa_aktif: number
  total_pembayaran_bulan_ini: number
  pendaftar_ppdb_baru: number
  tagihan_jatuh_tempo: number
  pembayaran_pending: number
  total_kelas: number
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

export interface ApiResponse<T> {
  data: T
}

export interface ApiError {
  error: {
    code: string
    message: string
    details?: unknown
  }
}
