import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchProfile, type UserProfile } from '../api/user'
import { post } from '../utils/request'
import { useTenantStore } from './tenant'

interface LoginResult {
  token: string
  user: UserProfile
  tenant?: { id: number; name: string; code: string }
}

export const useUserStore = defineStore('user', () => {
  const token = ref(uni.getStorageSync('token') || '')
  const user = ref<UserProfile | null>(null)

  async function login(phone: string) {
    const data = await post<LoginResult>('/api/v1/app/auth/login', { phone })
    token.value = data.token
    user.value = data.user
    uni.setStorageSync('token', data.token)
    uni.setStorageSync('user', JSON.stringify(data.user))
    if (data.tenant) {
      const tenantStore = useTenantStore()
      tenantStore.setTenant(data.tenant)
    }
    await refreshProfile()
    return data
  }

  async function refreshProfile() {
    if (!token.value) return
    try {
      user.value = await fetchProfile()
      uni.setStorageSync('user', JSON.stringify(user.value))
    } catch (e: unknown) {
      const err = e as { code?: number; message?: string }
      if (err?.code === 40300 || err?.message?.includes('Token') || err?.message?.includes('租户')) {
        logout()
      }
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    uni.removeStorageSync('token')
    uni.removeStorageSync('user')
  }

  function restoreUser() {
    if (!user.value) {
      try {
        const raw = uni.getStorageSync('user')
        if (raw) user.value = JSON.parse(raw as string)
      } catch {
        /* ignore */
      }
    }
  }

  restoreUser()

  return { token, user, login, logout, refreshProfile }
})
