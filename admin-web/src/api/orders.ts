import request from './request'

export interface OrderItem {
  product_name: string
  sku_name: string
  price: number
  quantity: number
}

export interface Order {
  id: number
  order_no: string
  user_id: number
  status: number
  goods_amount: number
  discount_amount: number
  freight_amount: number
  total_amount: number
  pay_amount: number
  pay_channel: string
  created_at: string
  items: OrderItem[]
}

export function fetchAdminOrders(params: { page?: number; page_size?: number; status?: number }) {
  return request.get<{ list: Order[]; total: number }>('/api/v1/admin/orders', { params })
}

export function shipOrder(orderNo: string) {
  return request.post(`/api/v1/admin/orders/${orderNo}/ship`)
}

export function centToYuan(c: number) {
  return (c / 100).toFixed(2)
}

export const statusMap: Record<number, string> = {
  0: '待付款', 1: '待发货', 2: '待收货', 3: '已完成', 4: '已取消', 5: '已关闭',
}
