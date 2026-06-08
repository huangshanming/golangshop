import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '',
  timeout: 15000,
})

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

instance.interceptors.response.use(
  (res: AxiosResponse<ApiResponse>) => {
    const body = res.data
    if (body.code === 0) {
      return body.data as never
    }
    ElMessage.error(body.message || '请求失败')
    return Promise.reject(body)
  },
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('admin_token')
      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login'
      }
    }
    ElMessage.error(err.message || '网络错误')
    return Promise.reject(err)
  }
)

interface RequestClient {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>
  put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>
  patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>
  delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = instance as unknown as RequestClient

export default request
