import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  adminLogin,
  enterTenant,
  exitTenant,
  fetchMe,
  type AdminInfo,
  type LoginCandidate,
  type MenuTreeNode,
  type TenantBrief,
} from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const admin = ref<AdminInfo | null>(null)
  const roles = ref<string[]>([])
  const menus = ref<MenuTreeNode[]>([])
  const platformView = ref(true)
  const activeTenant = ref<TenantBrief | null>(null)

  const isPlatformAdmin = computed(() => admin.value?.user_type === 'platform')
  const menuPaths = computed(() => {
    const paths: string[] = []
    const walk = (nodes: MenuTreeNode[]) => {
      for (const n of nodes) {
        if (n.path) paths.push(n.path)
        if (n.children?.length) walk(n.children)
      }
    }
    walk(menus.value)
    return paths
  })

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('admin_token', t)
  }

  function clearSession() {
    token.value = ''
    admin.value = null
    roles.value = []
    menus.value = []
    platformView.value = true
    activeTenant.value = null
    localStorage.removeItem('admin_token')
  }

  async function applyMe() {
    const me = await fetchMe()
    admin.value = me.admin
    roles.value = me.roles || []
    menus.value = me.menus || []
    platformView.value = me.platform_view
    activeTenant.value = me.active_tenant || null
    return me
  }

  async function login(username: string, password: string, tenantId?: number) {
    const data = await adminLogin(username, password, tenantId)
    if (data.need_tenant_pick) {
      return { needTenantPick: true, candidates: data.candidates as LoginCandidate[] }
    }
    setToken(data.token!)
    await applyMe()
    return { needTenantPick: false }
  }

  async function doEnterTenant(tenantId: number) {
    const data = await enterTenant(tenantId)
    setToken(data.token)
    await applyMe()
  }

  async function doExitTenant() {
    const data = await exitTenant()
    setToken(data.token)
    await applyMe()
  }

  async function restore() {
    if (!token.value) return
    try {
      await applyMe()
    } catch {
      clearSession()
    }
  }

  function hasPath(path: string) {
    return menuPaths.value.includes(path)
  }

  return {
    token,
    admin,
    roles,
    menus,
    platformView,
    activeTenant,
    isPlatformAdmin,
    menuPaths,
    login,
    doEnterTenant,
    doExitTenant,
    restore,
    clearSession,
    hasPath,
    applyMe,
  }
})
