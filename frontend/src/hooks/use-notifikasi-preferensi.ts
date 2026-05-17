import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse, PaginatedResponse } from '@/types'

export interface NotifikasiPreferensi {
  id: number
  sekolah_id: number
  pengguna_id: number | null
  siswa_id: number | null
  recipient_type: string
  channel: string
  destination: string
  enabled: boolean
  consent_status: string
  consent_source: string
  consent_at: string
  revoked_at: string
  created_at: string
  updated_at: string
}

interface PreferensiListParams {
  page?: number
  limit?: number
  channel?: string
  consent_status?: string
  enabled?: string
}

export function usePreferensiList(params: PreferensiListParams = {}) {
  return useQuery({
    queryKey: ['notifikasi-preferensi', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<NotifikasiPreferensi>>('/notifikasi/preferensi', { params })
      return res.data
    },
  })
}

export function useUpsertPreferensi() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: {
      recipient_type?: string
      channel: string
      destination: string
      enabled?: boolean
      consent_status: string
      consent_source?: string
    }) => {
      const res = await api.put<ApiResponse<NotifikasiPreferensi>>('/notifikasi/preferensi', data)
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifikasi-preferensi'] })
    },
  })
}

export function useGenerateTelegramInvite() {
  return useMutation({
    mutationFn: async (preferenceId: number) => {
      const res = await api.post<ApiResponse<{ invite_link: string; expires_in: string }>>(
        `/notifikasi/preferensi/${preferenceId}/telegram-invite`
      )
      return res.data.data
    },
  })
}
