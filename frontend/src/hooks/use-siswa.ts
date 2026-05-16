import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { Siswa, PaginatedResponse, ApiResponse } from '@/types'

interface SiswaListParams {
  page?: number
  limit?: number
  search?: string
  sort?: string
  order?: string
}

export function useSiswaList(params: SiswaListParams = {}) {
  return useQuery({
    queryKey: ['siswa', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<Siswa>>('/siswa', { params })
      return res.data
    },
  })
}

export function useSiswaDetail(id: number | null) {
  return useQuery({
    queryKey: ['siswa', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Siswa>>(`/siswa/${id}`)
      return res.data.data
    },
    enabled: id !== null,
  })
}

export function useCreateSiswa() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: Partial<Siswa>) => {
      const res = await api.post<ApiResponse<Siswa>>('/siswa', data)
      return res.data.data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['siswa'] }),
  })
}

export function useUpdateSiswa() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: Partial<Siswa> }) => {
      const res = await api.put<ApiResponse<Siswa>>(`/siswa/${id}`, data)
      return res.data.data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['siswa'] }),
  })
}

export function useDeleteSiswa() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/siswa/${id}`)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['siswa'] }),
  })
}
