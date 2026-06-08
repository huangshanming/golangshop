import request from './request'

export interface AppUser {
  id: number
  phone: string
  nickname: string
  balance: number
  status: number
}

export interface BalanceLog {
  id: number
  change_amount: number
  balance_after: number
  biz_type: string
  biz_no: string
  remark: string
  created_at: string
  order_no?: string
  goods_summary?: string
}

export interface PageResult<T> {
  list: T[]
  total: number
}

export function fetchUsers(params: { page?: number; page_size?: number; keyword?: string }) {
  return request.get<PageResult<AppUser>>('/api/v1/admin/users', { params })
}

export function adjustBalance(userId: number, change_amount: number, remark: string) {
  return request.post(`/api/v1/admin/users/${userId}/balance`, { change_amount, remark })
}

export function fetchBalanceLogs(userId: number, page = 1) {
  return request.get<PageResult<BalanceLog>>(`/api/v1/admin/users/${userId}/balance-logs`, { params: { page } })
}

export function centToYuan(cent: number) {
  return cent / 100
}
