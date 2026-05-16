import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type { ApiResponse, PaginatedResponse } from '@/types'

export interface Pendaftaran {
  id: number
  sekolah_id: number
  tahun_ajaran_id: number
  nama_lengkap: string
  nik: string
  tempat_lahir: string
  tanggal_lahir: string
  jenis_kelamin: string
  agama: string
  alamat: string
  asal_sekolah: string
  no_hp: string
  email: string
  nama_ortu: string
  no_hp_ortu: string
  pekerjaan_ortu: string
  foto: string
  status: string
  skor: number
  ranking: number
  catatan: string
  created_at: string
  updated_at: string
}

export interface Berkas {
  id: number
  pendaftaran_id: number
  jenis_berkas: string
  file_path: string
  status: string
  catatan: string
  created_at: string
}

export interface Pengumuman {
  id: number
  pendaftaran_id: number
  status: string
  ranking: number
  keterangan: string
  tanggal_pengumuman: string
  created_at: string
}

interface PendaftarListParams {
  page?: number
  limit?: number
  status?: string
  tahun_ajaran_id?: number
  search?: string
}

export function usePendaftarList(params: PendaftarListParams = {}) {
  return useQuery({
    queryKey: ['ppdb-pendaftar', params],
    queryFn: async () => {
      const res = await api.get<PaginatedResponse<Pendaftaran>>('/ppdb/pendaftar', { params })
      return res.data
    },
  })
}

export function usePendaftarDetail(id: number | null) {
  return useQuery({
    queryKey: ['ppdb-pendaftar', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Pendaftaran>>(`/ppdb/pendaftar/${id}`)
      return res.data.data
    },
    enabled: id !== null,
  })
}

export function useUpdatePendaftarStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, status, catatan }: { id: number; status: string; catatan?: string }) => {
      await api.put(`/ppdb/pendaftar/${id}`, { status, catatan })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ppdb-pendaftar'] }),
  })
}

export function useBerkasList(pendaftaranId: number | null) {
  return useQuery({
    queryKey: ['ppdb-berkas', pendaftaranId],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Berkas[]>>(`/ppdb/pendaftar/${pendaftaranId}/berkas`)
      return res.data.data
    },
    enabled: pendaftaranId !== null,
  })
}

export function useVerifikasiBerkas() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, status, catatan }: { id: number; status: string; catatan?: string }) => {
      await api.put(`/ppdb/berkas/${id}`, { status, catatan })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ppdb-berkas'] }),
  })
}

export function useInputUjian() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { pendaftaran_id: number; nama_ujian: string; nilai: number; keterangan?: string }) => {
      await api.post('/ppdb/ujian', data)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ppdb-pendaftar'] }),
  })
}

export function usePublishPengumuman() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { pendaftaran_id: number; status: string; ranking?: number; keterangan?: string }) => {
      await api.post('/ppdb/pengumuman', data)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ppdb-pendaftar'] }),
  })
}

export function useDaftar() {
  return useMutation({
    mutationFn: async (data: Partial<Pendaftaran> & { tahun_ajaran_id: number }) => {
      const res = await api.post<ApiResponse<Pendaftaran>>('/ppdb/daftar', data)
      return res.data.data
    },
  })
}

export function usePengumuman(id: number | null) {
  return useQuery({
    queryKey: ['ppdb-pengumuman', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Pengumuman>>(`/ppdb/pengumuman/${id}`)
      return res.data.data
    },
    enabled: id !== null && id > 0,
  })
}
