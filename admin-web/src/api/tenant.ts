import request from './request'

export interface Tenant {
  id: number
  name: string
  code: string
  status: number
  created_at?: string
}

const BASE_URL = import.meta.env.VITE_API_BASE || ''

export async function fetchPublicTenants(): Promise<Tenant[]> {
  const res = await fetch(`${BASE_URL}/api/v1/public/tenants`)
  const body = await res.json()
  if (body.code !== 0) throw new Error(body.message || '获取租户列表失败')
  return body.data || []
}

export function fetchTenants() {
  return request.get<Tenant[]>('/api/v1/admin/tenants')
}

export function fetchCurrentTenant() {
  return request.get<Tenant>('/api/v1/admin/tenants/current')
}

export function createTenant(data: { name: string; code: string }) {
  return request.post<Tenant>('/api/v1/admin/tenants', data)
}
