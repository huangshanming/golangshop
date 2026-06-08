import { get, post } from '../utils/request'

export interface CouponItem {
  user_coupon_id: number
  coupon_id: number
  name: string
  type: string
  threshold_amount: number
  discount_value: number
  max_discount: number
  valid_to: number
  can_use: boolean
  reason: string
  discount_amount: number
}

export function fetchAvailableCoupons(goods_amount: number) {
  return get<CouponItem[]>('/api/v1/app/coupons/available', { goods_amount })
}

export function fetchMyCoupons() {
  return get<CouponItem[]>('/api/v1/app/coupons/mine')
}

export function receiveCoupon(couponId: number) {
  return post<null>(`/api/v1/app/coupons/${couponId}/receive`, {})
}
