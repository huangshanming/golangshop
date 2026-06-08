import request from './request'

export interface Coupon {
  id: number
  name: string
  type: 'fixed' | 'percent'
  threshold_amount: number
  discount_value: number
  max_discount: number
  valid_from: number
  valid_to: number
  total_count: number
  received_count: number
  per_user_limit: number
  status: number
  created_at?: string
}

export interface PageResult<T> {
  list: T[]
  total: number
}

export interface CouponFormData {
  name: string
  type: 'fixed' | 'percent'
  threshold_amount: number
  discount_value: number
  max_discount: number
  valid_from: number
  valid_to: number
  total_count: number
  per_user_limit: number
  status: number
}

export function fetchCoupons(params: { page?: number; page_size?: number; keyword?: string }) {
  return request.get<PageResult<Coupon>>('/api/v1/admin/coupons', { params })
}

export function fetchCoupon(id: number) {
  return request.get<Coupon>(`/api/v1/admin/coupons/${id}`)
}

export function createCoupon(data: CouponFormData) {
  return request.post<Coupon>('/api/v1/admin/coupons', data)
}

export function updateCoupon(id: number, data: CouponFormData) {
  return request.put<Coupon>(`/api/v1/admin/coupons/${id}`, data)
}

export function updateCouponStatus(id: number, status: number) {
  return request.patch(`/api/v1/admin/coupons/${id}/status`, { status })
}

export function centToYuan(cent: number) {
  return cent / 100
}

export function yuanToCent(yuan: number) {
  return Math.round(yuan * 100)
}

export function formatCouponDiscount(row: Coupon) {
  if (row.type === 'fixed') {
    return `减 ¥${centToYuan(row.discount_value).toFixed(2)}`
  }
  const fold = (100 - row.discount_value) / 10
  return `${fold}折${row.max_discount > 0 ? `（最高 ¥${centToYuan(row.max_discount).toFixed(2)}）` : ''}`
}

export function formatThreshold(row: Coupon) {
  if (row.threshold_amount <= 0) return '无门槛'
  return `满 ¥${centToYuan(row.threshold_amount).toFixed(2)}`
}

import { formatDateTime } from '../utils/format'

export function formatValidRange(row: Coupon) {
  return `${formatDateTime(row.valid_from)} ~ ${formatDateTime(row.valid_to)}`
}

export function couponStatusLabel(status: number) {
  return status === 1 ? '启用' : '停用'
}
