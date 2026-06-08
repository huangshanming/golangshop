export interface Tenant {
  id: number
  name: string
  code: string
  status: number
}

const BASE_URL = import.meta.env.VITE_API_BASE || ''

export async function fetchPublicTenants(): Promise<Tenant[]> {
  const res = await fetch(`${BASE_URL}/api/v1/public/tenants`)
  const body = await res.json()
  if (body.code !== 0) {
    throw new Error(body.message || '获取店铺列表失败')
  }
  return body.data || []
}
