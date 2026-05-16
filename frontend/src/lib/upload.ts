import api from '@/lib/api'

export async function uploadFile(file: File, category: string): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('category', category)

  const res = await api.post<{ data: { path: string } }>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data.data.path
}

export function getFileUrl(path: string): string {
  return `/api/v1/upload/${path}`
}
