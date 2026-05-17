import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'

export interface BackupInfo {
  id: string
  filename: string
  size: number
  created_at: string
  checksum: string
}

export interface RestoreRequest {
  confirm: string
  backup_id: string
}

export function useBackupList() {
  return useQuery<BackupInfo[]>({
    queryKey: ['backup'],
    queryFn: async () => {
      const res = await api.get('/backup')
      return res.data.data
    },
  })
}

export function useCreateBackup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const res = await api.post('/backup')
      return res.data.data as BackupInfo
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backup'] })
    },
  })
}

export function useRestoreBackup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ backupId, confirm }: { backupId: string; confirm: string }) => {
      const res = await api.post(`/backup/restore/${backupId}`, {
        confirm,
        backup_id: backupId,
      })
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backup'] })
    },
  })
}
