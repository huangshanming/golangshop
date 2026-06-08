import { defineStore } from 'pinia'
import { ref } from 'vue'
import { post } from '../utils/request'

interface LoginResult {
  token: string
  user: { id: number; phone: string; nickname: string }
}

export const useUserStore = defineStore('user', () => {
  const token = ref(uni.getStorageSync('token') || '')
  const user = ref<LoginResult['user'] | null>(null)

  async function login(phone: string) {
    const data = await post<LoginResult>('/api/v1/app/auth/login', { phone })
    token.value = data.token
    user.value = data.user
    uni.setStorageSync('token', data.token)
    uni.setStorageSync('user', JSON.stringify(data.user))
    return data
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

  return { token, user, login, logout }
})
