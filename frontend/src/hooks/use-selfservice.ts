import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse } from '@/types'

export interface LinkedSiswa {
  id: number
  sekolah_id: number
  nis: string
  nama: string
  jenis_kelamin: string
  status: string
  hubungan: string
}

export interface SiswaDetailSelf {
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
  status: string
}

export interface TagihanSelf {
  id: number
  siswa_id: number
  kategori_id: number
  kategori_nama: string
  tahun_ajaran_id: number
  semester: string
  nominal: number
  jatuh_tempo: string
  status: 'belum_bayar' | 'sebagian' | 'lunas'
  catatan: string
}

export interface PembayaranSelf {
  id: number
  tagihan_id: number
  siswa_id: number
  jumlah: number
  tanggal: string
  metode: string
  bukti_bayar: string
  rekening_sekolah_id: number
  status: 'pending' | 'verified' | 'rejected'
  catatan: string
  created_at: string
}

export interface DashboardSiswaStats {
  total_tagihan: number
  tagihan_belum_bayar: number
  total_terbayar: number
  pembayaran_pending: number
}

export interface DashboardOrangtuaStats {
  jumlah_anak: number
  total_tagihan: number
  tagihan_belum_bayar: number
  total_terbayar: number
  pembayaran_pending: number
}

export interface DashboardGuruStats {
  total_kelas: number
  total_siswa: number
}

export interface GuruKelas {
  id: number
  nama: string
  tingkat: number
  jurusan_nama: string
  tahun_ajaran_nama: string
  jumlah_siswa: number
}

export interface GuruSiswa {
  id: number
  nis: string
  nama: string
  jenis_kelamin: string
  status: string
}

export function useLinkedSiswa() {
  return useQuery({
    queryKey: ['me', 'siswa'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<LinkedSiswa[]>>('/me/siswa')
      return res.data.data
    },
  })
}

export function useSiswaDetailSelf(id: number | null) {
  return useQuery({
    queryKey: ['me', 'siswa', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<SiswaDetailSelf>>(`/me/siswa/${id}`)
      return res.data.data
    },
    enabled: id !== null,
  })
}

export function useTagihanSelf(siswaId: number | null) {
  return useQuery({
    queryKey: ['me', 'siswa', siswaId, 'tagihan'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<TagihanSelf[]>>(`/me/siswa/${siswaId}/tagihan`)
      return res.data.data
    },
    enabled: siswaId !== null,
  })
}

export function usePembayaranSelf(siswaId: number | null) {
  return useQuery({
    queryKey: ['me', 'siswa', siswaId, 'pembayaran'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<PembayaranSelf[]>>(`/me/siswa/${siswaId}/pembayaran`)
      return res.data.data
    },
    enabled: siswaId !== null,
  })
}

export function useCreatePembayaranSelf() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { tagihan_id: number; jumlah: number; tanggal: string; metode: string; bukti_bayar?: string; rekening_sekolah_id?: number; catatan?: string }) => {
      const res = await api.post('/me/pembayaran', data)
      return res.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['me'] })
    },
  })
}

export interface GatewayPaymentResult {
  provider: string
  order_id: string
  payment_url: string
  payment_gateway_id: string
  status: string
}

export function useGatewayProviders() {
  return useQuery({
    queryKey: ['me', 'payment', 'providers'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<{ providers: string[] }>>('/me/payment/providers')
      return res.data.data.providers
    },
  })
}

export function useInitiateGatewayPayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { tagihan_id: number; provider: string }) => {
      const res = await api.post<ApiResponse<GatewayPaymentResult>>(`/me/tagihan/${data.tagihan_id}/pay`, { provider: data.provider })
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['me'] })
    },
  })
}

export function useDashboardSiswa() {
  return useQuery({
    queryKey: ['dashboard', 'siswa'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<DashboardSiswaStats>>('/dashboard/siswa')
      return res.data.data
    },
  })
}

export function useDashboardOrangtua() {
  return useQuery({
    queryKey: ['dashboard', 'orangtua'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<DashboardOrangtuaStats>>('/dashboard/orangtua')
      return res.data.data
    },
  })
}

export function useDashboardGuru() {
  return useQuery({
    queryKey: ['dashboard', 'guru'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<DashboardGuruStats>>('/dashboard/guru')
      return res.data.data
    },
  })
}

export function useGuruKelas() {
  return useQuery({
    queryKey: ['guru', 'kelas'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<GuruKelas[]>>('/guru/kelas')
      return res.data.data
    },
  })
}

export function useGuruSiswaByKelas(kelasId: number | null) {
  return useQuery({
    queryKey: ['guru', 'kelas', kelasId, 'siswa'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<GuruSiswa[]>>(`/guru/kelas/${kelasId}/siswa`)
      return res.data.data
    },
    enabled: kelasId !== null,
  })
}
