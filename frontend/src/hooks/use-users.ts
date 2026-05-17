import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { UserDetail, PaginatedResponse, ApiResponse } from '@/types'

interface UserListParams {
  page?: number
  limit?: number
  search?: string
  role?: string
  aktif?: string
}

export function useUserList(params: UserListParams = {}) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<UserDetail>>('/users', { params })
      return res.data
    },
  })
}

export function useUserDetail(id: number | null) {
  return useQuery({
    queryKey: ['users', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<UserDetail>>(`/users/${id}`)
      return res.data.data
    },
    enabled: id !== null,
  })
}

export function useCreateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { nama: string; email: string; password: string; role: string; no_hp?: string }) => {
      const res = await api.post<ApiResponse<UserDetail>>('/users', data)
      return res.data.data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useUpdateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: { nama: string; email: string; role: string; no_hp?: string; aktif: boolean } }) => {
      const res = await api.put<ApiResponse<UserDetail>>(`/users/${id}`, data)
      return res.data.data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useDeactivateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/users/${id}`)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useResetUserPassword() {
  return useMutation({
    mutationFn: async ({ id, password }: { id: number; password: string }) => {
      await api.post(`/users/${id}/reset-password`, { password })
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: { current_password: string; new_password: string }) => {
      await api.put('/auth/password', data)
    },
  })
}
