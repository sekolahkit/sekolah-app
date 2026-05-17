import { useQuery } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse } from '@/types'

export interface RekapPembayaranItem {
  tanggal: string
  metode: string
  total_transaksi: number
  total_nominal: number
}

export interface RekapPPDB {
  total_pendaftar: number
  menunggu: number
  berkas_lengkap: number
  diterima: number
  tidak_diterima: number
  cadangan: number
  daftar_ulang: number
}

export interface RekapSiswa {
  total: number
  aktif: number
  lulus: number
  pindah: number
  keluar: number
  laki_laki: number
  perempuan: number
}

export function useRekapPembayaran(tanggalMulai: string, tanggalSelesai: string, tahunAjaranId?: number) {
  return useQuery({
    queryKey: ['laporan-pembayaran', tanggalMulai, tanggalSelesai, tahunAjaranId],
    queryFn: async () => {
      const res = await api.get<ApiResponse<RekapPembayaranItem[]>>('/laporan/pembayaran', {
        params: { tanggal_mulai: tanggalMulai, tanggal_selesai: tanggalSelesai, tahun_ajaran_id: tahunAjaranId || undefined },
      })
      return res.data.data
    },
    enabled: !!tanggalMulai && !!tanggalSelesai,
  })
}

export function useRekapPPDB(tahunAjaranId: number) {
  return useQuery({
    queryKey: ['laporan-ppdb', tahunAjaranId],
    queryFn: async () => {
      const res = await api.get<ApiResponse<RekapPPDB>>('/laporan/ppdb', {
        params: { tahun_ajaran_id: tahunAjaranId },
      })
      return res.data.data
    },
    enabled: tahunAjaranId > 0,
  })
}

export function useRekapSiswa(tahunAjaranId?: number) {
  return useQuery({
    queryKey: ['laporan-siswa', tahunAjaranId],
    queryFn: async () => {
      const res = await api.get<ApiResponse<RekapSiswa>>('/laporan/siswa', {
        params: { tahun_ajaran_id: tahunAjaranId || undefined },
      })
      return res.data.data
    },
  })
}
