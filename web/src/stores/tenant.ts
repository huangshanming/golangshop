import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchPublicTenants, type Tenant } from '../api/tenant'

const STORAGE_ID = 'tenant_id'
const STORAGE_NAME = 'tenant_name'
const STORAGE_CODE = 'tenant_code'

export const useTenantStore = defineStore('tenant', () => {
  const tenantId = ref(Number(uni.getStorageSync(STORAGE_ID) || import.meta.env.VITE_TENANT_ID || 1))
  const tenantName = ref(String(uni.getStorageSync(STORAGE_NAME) || ''))
  const tenantCode = ref(String(uni.getStorageSync(STORAGE_CODE) || ''))
  const tenants = ref<Tenant[]>([])
  const loaded = ref(false)

  function persist() {
    uni.setStorageSync(STORAGE_ID, String(tenantId.value))
    uni.setStorageSync(STORAGE_NAME, tenantName.value)
    uni.setStorageSync(STORAGE_CODE, tenantCode.value)
  }

  async function loadTenants() {
    const list = await fetchPublicTenants()
    tenants.value = list
    loaded.value = true
    const current = list.find((t) => t.id === tenantId.value)
    if (current) {
      tenantName.value = current.name
      tenantCode.value = current.code
      persist()
    } else if (list.length > 0) {
      setTenant(list[0])
    }
    return list
  }

  function setTenant(t: Pick<Tenant, 'id' | 'name' | 'code'>) {
    if (t.id === tenantId.value) {
      tenantName.value = t.name
      tenantCode.value = t.code
      persist()
      return false
    }
    tenantId.value = t.id
    tenantName.value = t.name
    tenantCode.value = t.code
    persist()
    return true
  }

  function setTenantById(id: number, name = '', code = '') {
    const found = tenants.value.find((t) => t.id === id)
    if (found) {
      return setTenant(found)
    }
    if (id === tenantId.value) {
      if (name) tenantName.value = name
      if (code) tenantCode.value = code
      persist()
      return false
    }
    tenantId.value = id
    if (name) tenantName.value = name
    if (code) tenantCode.value = code
    persist()
    return true
  }

  function getTenantId() {
    return tenantId.value
  }

  return {
    tenantId,
    tenantName,
    tenantCode,
    tenants,
    loaded,
    loadTenants,
    setTenant,
    setTenantById,
    getTenantId,
    persist,
  }
})
