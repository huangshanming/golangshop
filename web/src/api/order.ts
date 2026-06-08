import { del, get, post } from '../utils/request'

export interface OrderItem {
  id: number
  product_id: number
  sku_id: number
  product_name: string
  sku_name: string
  cover_image: string
  spec_values: string
  price: number
  original_price: number
  line_amount: number
  quantity: number
}

export interface Order {
  id: number
  order_no: string
  status: number
  review_status: number
  display_status: string
  goods_amount: number
  discount_amount: number
  freight_amount: number
  total_amount: number
  pay_amount: number
  pay_channel: string
  address_snapshot: string
  remark: string
  expire_at: number
  created_at: string
  items: OrderItem[]
}

export interface PreviewLineItem {
  sku_id: number
  product_id: number
  product_name: string
  spec: string
  cover_image: string
  price: number
  quantity: number
  line_amount: number
  stock_enough: boolean
}

export interface PreviewAddress {
  id: number
  name: string
  phone: string
  full_text: string
}

export interface PreviewResult {
  items: PreviewLineItem[]
  address: PreviewAddress | null
  goods_amount: number
  discount_amount: number
  freight_amount: number
  pay_amount: number
  user_coupon_id: number
  coupon_name: string
  preview_token: string
  balance: number
}

export interface CheckoutItem {
  sku_id: number
  quantity: number
}

export function previewOrder(data: {
  items: CheckoutItem[]
  address_id?: number
  user_coupon_id?: number
  cart_item_ids?: number[]
}) {
  return post<PreviewResult>('/api/v1/app/orders/preview', data)
}

export function createOrder(data: {
  client_order_no: string
  address_id: number
  user_coupon_id?: number
  preview_token: string
  pay_channel: 'balance' | 'wechat' | 'alipay'
  remark?: string
  cart_item_ids?: number[]
  items: CheckoutItem[]
}) {
  return post<Order>('/api/v1/app/orders', data)
}

export function payOrder(order_no: string, channel: 'balance' | 'wechat' | 'alipay') {
  return post<{ paid?: boolean; channel?: string }>('/api/v1/app/orders/pay', { order_no, channel })
}

export function mockPayOrder(order_no: string) {
  return post<{ paid: boolean }>(`/api/v1/app/orders/${order_no}/mock-pay`, {})
}

export function fetchOrders(params: { page?: number; page_size?: number; tab?: string }) {
  return get<{ list: Order[]; total: number }>('/api/v1/app/orders', params as Record<string, unknown>)
}

export function fetchOrderDetail(order_no: string) {
  return get<Order>(`/api/v1/app/orders/${order_no}`)
}

export function cancelOrder(order_no: string, reason?: string) {
  return post<null>(`/api/v1/app/orders/${order_no}/cancel`, { reason })
}

export function confirmReceive(order_no: string) {
  return post<null>(`/api/v1/app/orders/${order_no}/confirm-receive`, {})
}

export function reviewOrder(order_no: string) {
  return post<null>(`/api/v1/app/orders/${order_no}/review`, {})
}

export function createAfterSale(data: {
  order_no: string
  order_item_id?: number
  type: 'exchange' | 'return_refund' | 'refund_only'
  reason: string
  refund_amount?: number
}) {
  return post('/api/v1/app/after-sales', data)
}

export function genClientOrderNo() {
  return `C${Date.now()}${Math.floor(Math.random() * 10000)}`
}
