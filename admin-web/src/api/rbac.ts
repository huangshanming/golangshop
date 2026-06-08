import request from './request'

export interface AdminMenu {
  id: number
  parent_id: number
  title: string
  path: string
  icon: string
  sort: number
  scope: string
  permission: string
  status: number
}

export interface AdminRole {
  id: number
  tenant_id: number
  code: string
  name: string
  scope: string
  status: number
}

export interface AdminAccount {
  id: number
  tenant_id: number
  username: string
  nickname: string
  user_type: string
  status: number
}

export function fetchMenus() {
  return request.get<AdminMenu[]>('/api/v1/admin/menus')
}

export function createMenu(data: Partial<AdminMenu>) {
  return request.post<AdminMenu>('/api/v1/admin/menus', data)
}

export function updateMenu(id: number, data: Partial<AdminMenu>) {
  return request.put<AdminMenu>(`/api/v1/admin/menus/${id}`, data)
}

export function deleteMenu(id: number) {
  return request.delete(`/api/v1/admin/menus/${id}`)
}

export function fetchRoles(scope?: string) {
  return request.get<AdminRole[]>('/api/v1/admin/roles', { params: scope ? { scope } : {} })
}

export function createRole(data: Partial<AdminRole>) {
  return request.post<AdminRole>('/api/v1/admin/roles', data)
}

export function updateRole(id: number, data: { name?: string; status?: number }) {
  return request.put<AdminRole>(`/api/v1/admin/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete(`/api/v1/admin/roles/${id}`)
}

export function fetchRoleMenus(roleId: number) {
  return request.get<number[]>(`/api/v1/admin/roles/${roleId}/menus`)
}

export function setRoleMenus(roleId: number, menuIds: number[]) {
  return request.put(`/api/v1/admin/roles/${roleId}/menus`, { menu_ids: menuIds })
}

export function fetchAdmins(params?: { page?: number; page_size?: number; tenant_id?: number; keyword?: string }) {
  return request.get<{ list: AdminAccount[]; total: number }>('/api/v1/admin/admins', { params })
}

export function createAdmin(data: {
  tenant_id?: number
  username: string
  password: string
  nickname?: string
  user_type?: string
  role_ids?: number[]
}) {
  return request.post<AdminAccount>('/api/v1/admin/admins', data)
}

export function updateAdmin(
  id: number,
  data: { nickname?: string; password?: string; status?: number; role_ids?: number[] }
) {
  return request.put<AdminAccount>(`/api/v1/admin/admins/${id}`, data)
}

export function fetchAdminRoles(adminId: number) {
  return request.get<number[]>(`/api/v1/admin/admins/${adminId}/roles`)
}
