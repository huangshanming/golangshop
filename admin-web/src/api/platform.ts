import request from './request'

export interface PlatformCategory {
  id: number
  parent_id: number
  name: string
  icon: string
  sort: number
  status: number
  children?: PlatformCategory[]
}

export interface HomeBanner {
  id: number
  title: string
  image_url: string
  position: 'home' | 'category'
  sort: number
  status: number
  link_type: 'product' | 'tenant' | 'category' | 'url'
  link_target: string
}

export interface SiteSettings {
  logo_url: string
  site_name: string
}

export interface FeaturedItem {
  id: number
  product_id: number
  sort: number
  status: number
  product_name: string
  tenant_name: string
  cover_image: string
  min_price: number
}

export interface SearchProductItem {
  id: number
  tenant_id: number
  name: string
  cover_image: string
  min_price: number
  tenant_name?: string
}

export function fetchPlatformCategories() {
  return request.get<PlatformCategory[]>('/api/v1/admin/platform-categories')
}

export function createPlatformCategory(data: Partial<PlatformCategory>) {
  return request.post<PlatformCategory>('/api/v1/admin/platform-categories', data)
}

export function updatePlatformCategory(id: number, data: Partial<PlatformCategory>) {
  return request.put(`/api/v1/admin/platform-categories/${id}`, data)
}

export function deletePlatformCategory(id: number) {
  return request.delete(`/api/v1/admin/platform-categories/${id}`)
}

export function fetchBanners(position?: 'home' | 'category') {
  return request.get<HomeBanner[]>('/api/v1/admin/banners', { params: position ? { position } : {} })
}

export function createBanner(data: Partial<HomeBanner>) {
  return request.post<HomeBanner>('/api/v1/admin/banners', data)
}

export function updateBanner(id: number, data: Partial<HomeBanner>) {
  return request.put(`/api/v1/admin/banners/${id}`, data)
}

export function deleteBanner(id: number) {
  return request.delete(`/api/v1/admin/banners/${id}`)
}

export function fetchSiteSettings() {
  return request.get<SiteSettings>('/api/v1/admin/site-settings')
}

export function updateSiteSettings(data: SiteSettings) {
  return request.put('/api/v1/admin/site-settings', data)
}

export function fetchFeatured() {
  return request.get<FeaturedItem[]>('/api/v1/admin/home-featured')
}

export function addFeatured(product_id: number, sort = 0) {
  return request.post('/api/v1/admin/home-featured', { product_id, sort })
}

export function deleteFeatured(id: number) {
  return request.delete(`/api/v1/admin/home-featured/${id}`)
}

export function updateFeaturedSort(items: { id: number; sort: number }[]) {
  return request.put('/api/v1/admin/home-featured/sort', { items })
}

export function searchProducts(keyword: string) {
  return request.get<SearchProductItem[]>('/api/v1/admin/products/search', { params: { keyword } })
}
