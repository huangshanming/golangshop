import request from './request'

export interface Category {
  id: number
  name: string
}

export function fetchCategories() {
  return request.get<Category[]>('/api/v1/admin/categories')
}
