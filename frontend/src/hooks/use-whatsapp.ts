import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse } from '@/types'

export interface WhatsAppStatus {
  status: string
  connected: boolean
  logged_in: boolean
  last_error: string
}

export interface WhatsAppQR {
  status: string
  qr: string | null
}

export function useWhatsAppStatus() {
  return useQuery({
    queryKey: ['whatsapp-status'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<WhatsAppStatus>>('/whatsapp/status')
      return res.data.data
    },
    refetchInterval: 5000,
  })
}

export function useWhatsAppConnect() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const res = await api.post<ApiResponse<{ message: string }>>('/whatsapp/connect')
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['whatsapp-status'] })
    },
  })
}

export function useWhatsAppDisconnect() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const res = await api.post<ApiResponse<{ message: string }>>('/whatsapp/disconnect')
      return res.data.data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['whatsapp-status'] })
    },
  })
}

export function useWhatsAppQR() {
  return useQuery({
    queryKey: ['whatsapp-qr'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<WhatsAppQR>>('/whatsapp/qr')
      return res.data.data
    },
    refetchInterval: 3000,
    enabled: false,
  })
}
