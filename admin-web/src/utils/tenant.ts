const ID_KEY = 'admin_tenant_id'
const NAME_KEY = 'admin_tenant_name'

export function getAdminTenantId(): string {
  return localStorage.getItem(ID_KEY) || import.meta.env.VITE_TENANT_ID || '1'
}

export function getAdminTenantName(): string {
  return localStorage.getItem(NAME_KEY) || ''
}

export function setAdminTenant(id: number, name: string) {
  localStorage.setItem(ID_KEY, String(id))
  localStorage.setItem(NAME_KEY, name)
}

export function clearAdminTenant() {
  localStorage.removeItem(ID_KEY)
  localStorage.removeItem(NAME_KEY)
}
