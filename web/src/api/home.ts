import { get } from '../utils/request'
import type { ProductListItem } from './product'

export interface SiteSettings {
  logo_url: string
  site_name: string
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

export interface PlatformCategory {
  id: number
  parent_id: number
  name: string
  icon: string
  sort: number
  status: number
}

export function fetchSiteSettings() {
  return get<SiteSettings>('/api/v1/public/site-settings')
}

export function fetchBanners(position: 'home' | 'category' = 'home') {
  return get<HomeBanner[]>('/api/v1/public/banners', { position })
}

export function fetchTopCategories(limit = 8) {
  return get<PlatformCategory[]>('/api/v1/public/categories', { parent_id: 0, limit })
}

export interface CategoryTreeItem extends PlatformCategory {
  children?: CategoryTreeItem[]
}

export function fetchCategoryTree() {
  return get<CategoryTreeItem[]>('/api/v1/public/categories/tree')
}

export function fetchFeaturedProducts() {
  return get<ProductListItem[]>('/api/v1/public/products/featured')
}
