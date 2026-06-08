import request from './request'

export interface AdminInfo {
  id: number
  username: string
  nickname: string
  user_type: string
  tenant_id: number
  role?: string
}

export interface MenuTreeNode {
  id: number
  title: string
  path: string
  icon?: string
  sort: number
  scope: string
  permission: string
  children?: MenuTreeNode[]
}

export interface TenantBrief {
  id: number
  name: string
  code: string
}

export interface LoginCandidate {
  user_id: number
  tenant_id: number
  tenant_name: string
}

export interface LoginResult {
  token?: string
  admin?: AdminInfo
  need_tenant_pick?: boolean
  candidates?: LoginCandidate[]
}

export interface MeResult {
  admin: AdminInfo
  roles: string[]
  menus: MenuTreeNode[]
  platform_view: boolean
  active_tenant?: TenantBrief
}

export function adminLogin(username: string, password: string, tenantId?: number): Promise<LoginResult> {
  return request.post('/api/v1/admin/auth/login', {
    username,
    password,
    ...(tenantId != null ? { tenant_id: tenantId } : {}),
  })
}

export function fetchMe(): Promise<MeResult> {
  return request.get('/api/v1/admin/auth/me')
}

export function enterTenant(tenantId: number): Promise<{ token: string }> {
  return request.post('/api/v1/admin/auth/enter-tenant', { tenant_id: tenantId })
}

export function exitTenant(): Promise<{ token: string }> {
  return request.post('/api/v1/admin/auth/exit-tenant')
}
