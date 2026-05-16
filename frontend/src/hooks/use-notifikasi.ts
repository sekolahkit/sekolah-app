import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse, PaginatedResponse } from '@/types'

export interface Notifikasi {
  id: number
  sekolah_id: number
  tipe: string
  penerima: string
  pesan: string
  status: string
  retry_count: number
  max_retries: number
  last_error: string
  scheduled_at: string
  sent_at: string
  created_at: string
}

export interface QueueStats {
  pending: number
  sent: number
  failed: number
}

interface NotifikasiListParams {
  page?: number
  limit?: number
  status?: string
  tipe?: string
}

export function useNotifikasiList(params: NotifikasiListParams = {}) {
  return useQuery({
    queryKey: ['notifikasi', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<Notifikasi>>('/notifikasi', { params })
      return res.data
    },
  })
}

export function useQueueStats() {
  return useQuery({
    queryKey: ['notifikasi-queue'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<QueueStats>>('/notifikasi/queue')
      return res.data.data
    },
  })
}

export function useTestSend() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { tipe: string; penerima: string; pesan: string }) => {
      const res = await api.post<ApiResponse<Notifikasi>>('/notifikasi/test', data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifikasi'] })
      qc.invalidateQueries({ queryKey: ['notifikasi-queue'] })
    },
  })
}
