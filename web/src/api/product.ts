import { get } from '../utils/request'

export interface Category {
  id: number
  name: string
  sort: number
}

export interface SpecDimension {
  name: string
  values: string[]
}

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
  original_price: number
  total_stock: number
  category_name: string
}

export interface ProductSKU {
  id: number
  product_id: number
  spec: string
  spec_values: string
  price: number
  original_price: number
  stock: number
  image: string
}

export interface ProductDetail {
  product: ProductListItem
  skus: ProductSKU[]
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export function fetchCategories() {
  return get<Category[]>('/api/v1/app/categories')
}

export function fetchProducts(params: {
  page?: number
  page_size?: number
  category_id?: number
  keyword?: string
}) {
  return get<PageResult<ProductListItem>>('/api/v1/app/products', params as Record<string, unknown>)
}

export function fetchProductDetail(id: number | string) {
  return get<ProductDetail>(`/api/v1/app/products/${id}`)
}

export function parseImages(images?: string, cover?: string): string[] {
  if (images) {
    try {
      const arr = JSON.parse(images) as string[]
      if (Array.isArray(arr) && arr.length > 0) return arr
    } catch {
      /* ignore */
    }
  }
  return cover ? [cover] : []
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

export function matchSkuBySelection(skus: ProductSKU[], selected: Record<string, string>): ProductSKU | null {
  return skus.find((sku) => {
    const values = parseSpecValues(sku.spec_values)
    if (Object.keys(values).length === 0) return false
    return Object.entries(selected).every(([k, v]) => values[k] === v)
  }) || null
}
