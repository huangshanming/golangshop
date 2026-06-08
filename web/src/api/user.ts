import { get, put } from '../utils/request'

export interface UserProfile {
  id: number
  phone: string
  nickname: string
  avatar: string
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

export function fetchProfile() {
  return get<UserProfile>('/api/v1/app/user/me')
}

export function updateProfile(data: { nickname: string }) {
  return put<UserProfile>('/api/v1/app/user/profile', data)
}

export function fetchBalanceLogs(params?: { page?: number; page_size?: number }) {
  return get<{ list: BalanceLog[]; total: number }>('/api/v1/app/user/balance-logs', params as Record<string, unknown>)
}

export const bizTypeMap: Record<string, string> = {
  admin_adjust: '管理员调账',
  order_pay: '订单支付',
  order_refund: '订单退款',
  recharge: '充值',
}
