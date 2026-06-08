import request from './request'

export interface OrderItem {
  product_name: string
  sku_name: string
  price: number
  quantity: number
  cover_image?: string
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
  address_snapshot: string
  express_company: string
  express_no: string
  ship_remark: string
  shipped_at?: number
  created_at: string
  items: OrderItem[]
}

export interface ShipOrderPayload {
  express_company: string
  express_no: string
  ship_remark?: string
}

export const expressCompanies = [
  { code: 'SF', name: '顺丰速运' },
  { code: 'ZTO', name: '中通快递' },
  { code: 'YTO', name: '圆通速递' },
  { code: 'YD', name: '韵达快递' },
  { code: 'STO', name: '申通快递' },
  { code: 'JTSD', name: '极兔速递' },
  { code: 'EMS', name: '邮政 EMS' },
  { code: 'JD', name: '京东物流' },
  { code: 'DBL', name: '德邦快递' },
  { code: 'OTHER', name: '其他快递' },
]

export function expressCompanyName(code: string) {
  return expressCompanies.find((c) => c.code === code)?.name || code || '-'
}

export function fetchAdminOrders(params: { page?: number; page_size?: number; status?: number }) {
  return request.get<{ list: Order[]; total: number }>('/api/v1/admin/orders', { params })
}

export function shipOrder(orderNo: string, data: ShipOrderPayload) {
  return request.post(`/api/v1/admin/orders/${orderNo}/ship`, data)
}

export function centToYuan(c: number) {
  return (c / 100).toFixed(2)
}

export const statusMap: Record<number, string> = {
  0: '待付款', 1: '待发货', 2: '待收货', 3: '已完成', 4: '已取消', 5: '已关闭',
}

export interface AddressSnap {
  name?: string
  phone?: string
  province?: string
  city?: string
  district?: string
  detail?: string
}

export function parseAddressSnapshot(raw: string): AddressSnap | null {
  if (!raw) return null
  try {
    return JSON.parse(raw) as AddressSnap
  } catch {
    return null
  }
}

export function formatAddress(snap: AddressSnap | null) {
  if (!snap) return '-'
  return `${snap.province || ''}${snap.city || ''}${snap.district || ''}${snap.detail || ''}`
}
