import { del, get, post, put } from '../utils/request'

export interface Address {
  id: number
  name: string
  phone: string
  province: string
  city: string
  district: string
  detail: string
  is_default: boolean
}

export function fetchAddresses() {
  return get<Address[]>('/api/v1/app/addresses')
}

export function createAddress(data: Omit<Address, 'id'>) {
  return post<Address>('/api/v1/app/addresses', data)
}

export function updateAddress(id: number, data: Omit<Address, 'id'>) {
  return put<null>(`/api/v1/app/addresses/${id}`, data)
}

export function deleteAddress(id: number) {
  return del<null>(`/api/v1/app/addresses/${id}`)
}

export function formatAddress(a: Address) {
  return `${a.province}${a.city}${a.district}${a.detail}`
}
