import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse } from '@/types'

export interface TahunAjaran {
  id: number
  sekolah_id: number
  nama: string
  aktif: boolean
  tanggal_mulai: string
  tanggal_selesai: string
  created_at: string
}

export function useTahunAjaranList() {
  return useQuery({
    queryKey: ['tahun-ajaran'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<TahunAjaran[]>>('/tahun-ajaran')
      return res.data.data
    },
  })
}

export function useTahunAjaranAktif() {
  return useQuery({
    queryKey: ['tahun-ajaran-aktif'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<TahunAjaran | null>>('/tahun-ajaran/aktif')
      return res.data.data
    },
  })
}

export function useCreateTahunAjaran() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { nama: string; tanggal_mulai?: string; tanggal_selesai?: string }) => {
      const res = await api.post<ApiResponse<TahunAjaran>>('/tahun-ajaran', data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tahun-ajaran'] })
    },
  })
}

export function useUpdateTahunAjaran() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: { id: number; nama: string; tanggal_mulai?: string; tanggal_selesai?: string }) => {
      const res = await api.put<ApiResponse<TahunAjaran>>(`/tahun-ajaran/${id}`, data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tahun-ajaran'] })
    },
  })
}

export function useSetTahunAjaranAktif() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      const res = await api.put<ApiResponse<{ message: string }>>(`/tahun-ajaran/${id}/aktif`)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tahun-ajaran'] })
      qc.invalidateQueries({ queryKey: ['tahun-ajaran-aktif'] })
    },
  })
}
