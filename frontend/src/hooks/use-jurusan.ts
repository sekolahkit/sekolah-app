import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse } from '@/types'

export interface Jurusan {
  id: number
  sekolah_id: number
  nama: string
  kode: string
  created_at: string
}

export function useJurusanList() {
  return useQuery({
    queryKey: ['jurusan'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Jurusan[]>>('/jurusan')
      return res.data.data
    },
  })
}

export function useCreateJurusan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { nama: string; kode?: string }) => {
      const res = await api.post<ApiResponse<Jurusan>>('/jurusan', data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['jurusan'] })
    },
  })
}

export function useUpdateJurusan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, ...data }: { id: number; nama: string; kode?: string }) => {
      const res = await api.put<ApiResponse<Jurusan>>(`/jurusan/${id}`, data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['jurusan'] })
    },
  })
}

export function useDeleteJurusan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/jurusan/${id}`)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['jurusan'] })
    },
  })
}
