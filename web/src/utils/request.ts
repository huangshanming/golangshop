const BASE_URL = import.meta.env.VITE_API_BASE || ''
const TENANT_ID = import.meta.env.VITE_TENANT_ID || '1'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

export function request<T>(options: UniApp.RequestOptions): Promise<T> {
  const token = uni.getStorageSync('token') as string
  return new Promise((resolve, reject) => {
    uni.request({
      ...options,
      url: `${BASE_URL}${options.url}`,
      header: {
        'Content-Type': 'application/json',
        'X-Tenant-Id': TENANT_ID,
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.header,
      },
      success: (res) => {
        const body = res.data as ApiResponse<T>
        if (body.code === 0) {
          resolve(body.data as T)
        } else {
          uni.showToast({ title: body.message || '请求失败', icon: 'none' })
          reject(body)
        }
      },
      fail: reject,
    })
  })
}

export function post<T>(url: string, data?: unknown) {
  return request<T>({ url, method: 'POST', data })
}

export function get<T>(url: string, data?: Record<string, unknown>) {
  return request<T>({ url, method: 'GET', data })
}

export function del<T>(url: string) {
  return request<T>({ url, method: 'DELETE' })
}

export function put<T>(url: string, data?: unknown) {
  return request<T>({ url, method: 'PUT', data })
}

export function patch<T>(url: string, data?: unknown) {
  return request<T>({ url, method: 'PATCH', data })
}
