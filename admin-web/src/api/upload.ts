import request from './request'

export interface UploadResult {
  url: string
  path: string
}

export function uploadImage(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request.post<UploadResult>('/api/v1/admin/upload', form)
}
