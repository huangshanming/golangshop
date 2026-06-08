import request from './request'

export interface Category {
  id: number
  parent_id?: number
  name: string
  icon?: string
  sort?: number
  status?: number
  children?: Category[]
}

export function fetchCategories() {
  return request.get<Category[]>('/api/v1/admin/categories')
}
