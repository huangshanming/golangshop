import request from './request'

export interface ProductListItem {
  id: number
  category_id: number
  name: string
  subtitle: string
  description: string
  cover_image: string
  images: string
  specs: string
  status: number
  sales_count: number
  min_price: number
  max_price: number
  total_stock: number
  category_name: string
}

export interface SpecDimension {
  name: string
  values: string[]
}

export interface ProductSKU {
  id?: number
  spec: string
  spec_values: string
  price: number
  original_price: number
  stock: number
  image: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface ProductForm {
  category_id: number
  name: string
  subtitle: string
  description: string
  cover_image: string
  images: string
  specs: SpecDimension[]
  skus: Array<{
    spec: string
    spec_values: Record<string, string>
    price: number
    original_price: number
    stock: number
    image: string
  }>
}

export function fetchProducts(params: { page?: number; page_size?: number; keyword?: string }) {
  return request.get<PageResult<ProductListItem>>('/api/v1/admin/products', { params })
}

export function fetchProductDetail(id: number) {
  return request.get<{ product: ProductListItem; skus: ProductSKU[] }>(`/api/v1/admin/products/${id}`)
}

export function createProduct(data: ProductForm) {
  return request.post('/api/v1/admin/products', data)
}

export function updateProduct(id: number, data: ProductForm) {
  return request.put(`/api/v1/admin/products/${id}`, data)
}

export function updateProductStatus(id: number, status: number) {
  return request.patch(`/api/v1/admin/products/${id}/status`, { status })
}

export function yuanToCent(yuan: number) {
  return Math.round(yuan * 100)
}

export function centToYuan(cent: number) {
  return cent / 100
}

export function parseSpecs(raw?: string): SpecDimension[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw) as SpecDimension[]
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

export function parseSpecValues(raw?: string): Record<string, string> {
  if (!raw) return {}
  try {
    return JSON.parse(raw) as Record<string, string>
  } catch {
    return {}
  }
}

export function parseImages(raw?: string): string[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw) as string[]
    return Array.isArray(arr) ? arr.filter(Boolean) : []
  } catch {
    return []
  }
}

export function buildImagesJSON(urls: string[]): string {
  return JSON.stringify(urls.filter(Boolean))
}
