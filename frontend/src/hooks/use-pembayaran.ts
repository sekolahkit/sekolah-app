import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { Tagihan, Pembayaran, Rekening, PaginatedResponse, ApiResponse } from '@/types'

interface TagihanListParams {
  page?: number
  limit?: number
  status?: string
  siswa_id?: number
  search?: string
}

interface PembayaranListParams {
  page?: number
  limit?: number
  status?: string
}

export function useTagihanList(params: TagihanListParams = {}) {
  return useQuery({
    queryKey: ['tagihan', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<Tagihan>>('/tagihan', { params })
      return res.data
    },
  })
}

export function useTagihanDetail(id: number | null) {
  return useQuery({
    queryKey: ['tagihan', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Tagihan>>(`/tagihan/${id}`)
      return res.data.data
    },
    enabled: id !== null,
  })
}

export function useCreateTagihan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Tagihan>) => {
      const res = await api.post<ApiResponse<Tagihan>>('/tagihan', data)
      return res.data.data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tagihan'] }),
  })
}

export function useBulkCreateTagihan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { siswa_ids: number[]; kategori_id: number; tahun_ajaran_id: number; nominal: number; jatuh_tempo: string; semester?: string; catatan?: string }) => {
      const res = await api.post('/tagihan/bulk', data)
      return res.data.data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tagihan'] }),
  })
}

export function usePembayaranList(params: PembayaranListParams = {}) {
  return useQuery({
    queryKey: ['pembayaran', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<Pembayaran>>('/pembayaran', { params })
      return res.data
    },
  })
}

export function useCreatePembayaran() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { tagihan_id: number; siswa_id: number; jumlah: number; tanggal: string; metode: string; rekening_sekolah_id?: number; catatan?: string }) => {
      const res = await api.post<ApiResponse<Pembayaran>>('/pembayaran', data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pembayaran'] })
      qc.invalidateQueries({ queryKey: ['tagihan'] })
    },
  })
}

export function useVerifyPembayaran() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      await api.put(`/pembayaran/${id}/verify`)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pembayaran'] })
      qc.invalidateQueries({ queryKey: ['tagihan'] })
    },
  })
}

export function useRejectPembayaran() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      await api.put(`/pembayaran/${id}/reject`)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pembayaran'] })
      qc.invalidateQueries({ queryKey: ['tagihan'] })
    },
  })
}

export function useRekeningAktif() {
  return useQuery({
    queryKey: ['rekening-aktif'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Rekening[]>>('/rekening-sekolah/aktif')
      return res.data.data
    },
  })
}
