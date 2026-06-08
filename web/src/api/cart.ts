import { del, get, patch, post } from '../utils/request'

export interface CartItem {
  id: number
  product_id: number
  sku_id: number
  product_name: string
  spec: string
  cover_image: string
  price: number
  stock: number
  quantity: number
  checked: boolean
}

export function fetchCart() {
  return get<CartItem[]>('/api/v1/app/cart')
}

export function addToCart(data: { product_id: number; sku_id: number; quantity?: number }) {
  return post<null>('/api/v1/app/cart', data)
}

export function updateCartItem(id: number, data: { quantity?: number; checked?: boolean }) {
  return patch<null>(`/api/v1/app/cart/${id}`, data)
}

export function checkAllCart(checked: boolean) {
  return post<null>('/api/v1/app/cart/check-all', { checked })
}

export function removeCartItem(id: number) {
  return del<null>(`/api/v1/app/cart/${id}`)
}
