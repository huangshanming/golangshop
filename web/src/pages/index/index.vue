<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import {
  fetchBanners,
  fetchFeaturedProducts,
  fetchSiteSettings,
  fetchTopCategories,
  type HomeBanner,
  type PlatformCategory,
  type SiteSettings,
} from '../../api/home'
import type { ProductListItem } from '../../api/product'
import ProductCard from '../../components/ProductCard.vue'
import { syncTabBarSelected } from '../../utils/tabbar'
import PageTabBar from '../../components/PageTabBar.vue'
import { catGradient } from '../../utils/theme'

const settings = ref<SiteSettings | null>(null)
const banners = ref<HomeBanner[]>([])
const categories = ref<PlatformCategory[]>([])
const products = ref<ProductListItem[]>([])
const keyword = ref('')
const loading = ref(false)

const showBento = computed(() => products.value.length >= 3)
const promoMain = computed(() => (showBento.value ? products.value[0] : null))
const promoSub = computed(() => (showBento.value ? products.value.slice(1, 3) : []))
const feedProducts = computed(() => (showBento.value ? products.value.slice(3) : products.value))

const displaySiteName = computed(() => normalizeSiteName(settings.value?.site_name))

function normalizeSiteName(name?: string) {
  if (!name || name === 'GoShop') return '搜店网'
  return name
}

async function loadHome() {
  loading.value = true
  try {
    const [site, bannerList, cats, featured] = await Promise.all([
      fetchSiteSettings(),
      fetchBanners('home'),
      fetchTopCategories(8),
      fetchFeaturedProducts(),
    ])
    settings.value = site
    banners.value = bannerList || []
    categories.value = (cats || []).slice(0, 8)
    products.value = featured || []
    const siteName = normalizeSiteName(site?.site_name)
    uni.setNavigationBarTitle({ title: siteName })
  } catch {
    products.value = []
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

function onSearch() {
  const kw = keyword.value.trim()
  uni.navigateTo({
    url: `/pages/product/list?keyword=${encodeURIComponent(kw)}`,
  })
}

function goCart() {
  uni.navigateTo({ url: '/pages/cart/index' })
}

function goMore() {
  uni.navigateTo({ url: '/pages/product/list' })
}

function goCategory(cat: PlatformCategory) {
  uni.navigateTo({
    url: `/pages/product/list?category_id=${cat.id}&category_name=${encodeURIComponent(cat.name)}`,
  })
}

function goProduct(item: ProductListItem) {
  uni.navigateTo({
    url: `/pages/product/detail?id=${item.id}&tenant_id=${item.tenant_id || 0}`,
  })
}

function handleBanner(banner: HomeBanner) {
  const target = banner.link_target
  switch (banner.link_type) {
    case 'category':
      uni.navigateTo({ url: `/pages/product/list?category_id=${target}` })
      break
    case 'product':
      uni.navigateTo({ url: `/pages/product/detail?id=${target}` })
      break
    case 'tenant':
      uni.navigateTo({ url: `/pages/shop/index?tenant_id=${target}` })
      break
    case 'url':
      // #ifdef H5
      window.open(target, '_blank')
      // #endif
      // #ifndef H5
      uni.setClipboardData({ data: target })
      uni.showToast({ title: '链接已复制', icon: 'none' })
      // #endif
      break
  }
}

onLoad(() => loadHome())
onShow(() => syncTabBarSelected())
onPullDownRefresh(() => loadHome())
</script>

<template>
  <view class="page">
    <view class="page-bg">
      <view class="blob blob-a" />
      <view class="blob blob-b" />
      <view class="blob blob-c" />
    </view>

    <view class="page-content">
      <view class="top-bar">
        <view class="search-glass">
          <text class="search-icon">🔍</text>
          <input
            v-model="keyword"
            class="search-input"
            placeholder="搜索好物、店铺..."
            confirm-type="search"
            @confirm="onSearch"
          />
        </view>
        <view class="cart-orb" @click="goCart">
          <text class="cart-icon">🛒</text>
        </view>
      </view>

      <view class="banner-stage">
        <swiper
          v-if="banners.length"
          class="banner-swiper"
          circular
          autoplay
          indicator-dots
          indicator-color="rgba(255,255,255,0.4)"
          indicator-active-color="#fff"
        >
          <swiper-item v-for="b in banners" :key="b.id">
            <view class="banner-card" @click="handleBanner(b)">
              <image class="banner-img" :src="b.image_url" mode="aspectFill" />
              <view class="banner-shine" />
            </view>
          </swiper-item>
        </swiper>
        <view v-else class="banner-placeholder">
          <view class="banner-placeholder-inner">
            <text class="banner-placeholder-text">{{ displaySiteName }}</text>
            <text class="banner-placeholder-sub">发现精选好物</text>
          </view>
        </view>
      </view>

      <view v-if="categories.length" class="panel category-panel">
        <view class="panel-head">
          <view class="panel-accent" />
          <text class="panel-title">热门分类</text>
        </view>
        <view class="category-grid">
          <view
            v-for="(cat, index) in categories.slice(0, 8)"
            :key="cat.id"
            class="category-item"
            @click="goCategory(cat)"
          >
            <view class="cat-orb" :style="{ background: catGradient(index) }">
              <view class="cat-orb-shine" />
              <image v-if="cat.icon" class="cat-icon-img" :src="cat.icon" mode="aspectFill" />
              <text v-else class="cat-icon-text">{{ cat.name.slice(0, 1) }}</text>
            </view>
            <text class="cat-name">{{ cat.name }}</text>
          </view>
        </view>
      </view>

      <view v-if="promoMain" class="promo-section">
        <view class="panel-head">
          <view class="panel-accent accent-pink" />
          <text class="panel-title">今日精选</text>
        </view>
        <view class="promo-bento">
          <view class="promo-main" @click="goProduct(promoMain)">
            <image class="promo-bg" :src="promoMain.cover_image" mode="aspectFill" />
            <view class="promo-mask" />
            <view class="promo-shine" />
            <view class="promo-content">
              <text class="promo-tag">精选</text>
              <text class="promo-title">{{ promoMain.name }}</text>
              <text class="promo-sub">{{ promoMain.tenant_name || '好物推荐' }}</text>
            </view>
          </view>
          <view v-if="promoSub.length" class="promo-side">
            <view
              v-for="item in promoSub"
              :key="item.id"
              class="promo-sub-card"
              @click="goProduct(item)"
            >
              <image class="promo-sub-bg" :src="item.cover_image" mode="aspectFill" />
              <view class="promo-sub-mask" />
              <view class="promo-sub-content">
                <text class="promo-sub-tag">热卖</text>
                <text class="promo-sub-title">{{ item.name }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view v-if="feedProducts.length" class="panel product-panel">
        <view class="section-head">
          <view class="section-head-left">
            <view class="panel-accent accent-blue" />
            <text class="section-title">{{ showBento ? '品类狂欢' : '精选好物' }}</text>
          </view>
          <view class="section-more-btn" @click="goMore">
            <text class="section-more">更多</text>
            <text class="section-more-arrow">›</text>
          </view>
        </view>
        <view class="product-grid">
          <view v-for="item in feedProducts" :key="`${item.tenant_id}-${item.id}`" class="grid-item">
            <ProductCard :item="item" compact />
          </view>
        </view>
      </view>

      <view v-else-if="!loading && !promoMain" class="empty">
        <text class="empty-icon">✨</text>
        <text class="empty-text">暂无推荐商品</text>
      </view>

      <view v-if="loading" class="loading-tip">加载中...</view>
    </view>
  </view>
  <PageTabBar />
</template>

<style scoped>
.page {
  position: relative;
  min-height: 100vh;
  background: linear-gradient(165deg, #eef1fb 0%, #f5f0ee 35%, #edf4f8 70%, #e8edf8 100%);
  overflow: hidden;
}

/* #ifdef H5 */
.page-content {
  padding-bottom: calc(50px + env(safe-area-inset-bottom, 0px));
}
/* #endif */

.page-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(60rpx);
}

.blob-a {
  width: 480rpx;
  height: 480rpx;
  top: -120rpx;
  right: -100rpx;
  background: radial-gradient(circle, rgba(107, 124, 255, 0.35) 0%, transparent 70%);
}

.blob-b {
  width: 400rpx;
  height: 400rpx;
  top: 320rpx;
  left: -140rpx;
  background: radial-gradient(circle, rgba(255, 140, 66, 0.28) 0%, transparent 70%);
}

.blob-c {
  width: 360rpx;
  height: 360rpx;
  bottom: 200rpx;
  right: -80rpx;
  background: radial-gradient(circle, rgba(61, 190, 184, 0.25) 0%, transparent 70%);
}

.page-content {
  position: relative;
  z-index: 1;
  padding: 0 24rpx;
}

.top-bar {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 20rpx 0 24rpx;
}

.search-glass {
  flex: 1;
  display: flex;
  align-items: center;
  height: 80rpx;
  padding: 0 28rpx;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(20rpx);
  -webkit-backdrop-filter: blur(20rpx);
  border-radius: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.95);
  box-shadow:
    0 8rpx 32rpx rgba(30, 40, 80, 0.08),
    0 2rpx 8rpx rgba(30, 40, 80, 0.04),
    inset 0 2rpx 6rpx rgba(255, 255, 255, 0.9),
    inset 0 -2rpx 6rpx rgba(30, 40, 80, 0.03);
}

.search-icon {
  font-size: 28rpx;
  margin-right: 14rpx;
  opacity: 0.5;
}

.search-input {
  flex: 1;
  font-size: 28rpx;
  color: #1a1f36;
}

.cart-orb {
  width: 80rpx;
  height: 80rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 24rpx;
  background: linear-gradient(145deg, #ffffff, #f0f2f8);
  border: 2rpx solid rgba(255, 255, 255, 0.95);
  box-shadow:
    0 8rpx 24rpx rgba(30, 40, 80, 0.1),
    inset 0 3rpx 6rpx rgba(255, 255, 255, 0.95),
    inset 0 -4rpx 8rpx rgba(30, 40, 80, 0.06);
}

.cart-icon {
  font-size: 38rpx;
}

.banner-stage {
  margin-bottom: 8rpx;
  filter: drop-shadow(0 16rpx 40rpx rgba(30, 40, 80, 0.18));
}

.banner-swiper,
.banner-placeholder {
  height: 300rpx;
  border-radius: 28rpx;
  overflow: hidden;
}

.banner-card {
  position: relative;
  width: 100%;
  height: 300rpx;
  border-radius: 28rpx;
  overflow: hidden;
  border: 3rpx solid rgba(255, 255, 255, 0.6);
  box-shadow:
    inset 0 2rpx 8rpx rgba(255, 255, 255, 0.3),
    inset 0 -4rpx 12rpx rgba(0, 0, 0, 0.15);
}

.banner-img {
  width: 100%;
  height: 100%;
}

.banner-shine {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 45%;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.22) 0%, transparent 100%);
  pointer-events: none;
}

.banner-placeholder {
  background: linear-gradient(135deg, #6b7cff 0%, #9b6bff 50%, #ff8c42 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 3rpx solid rgba(255, 255, 255, 0.4);
}

.banner-placeholder-inner {
  text-align: center;
}

.banner-placeholder-text {
  display: block;
  color: #fff;
  font-size: 44rpx;
  font-weight: 800;
  letter-spacing: 2rpx;
  text-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.2);
}

.banner-placeholder-sub {
  display: block;
  margin-top: 8rpx;
  color: rgba(255, 255, 255, 0.85);
  font-size: 24rpx;
}

.panel {
  margin-top: 28rpx;
  background: rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(16rpx);
  -webkit-backdrop-filter: blur(16rpx);
  border-radius: 28rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.95);
  box-shadow:
    0 12rpx 40rpx rgba(30, 40, 80, 0.08),
    0 4rpx 12rpx rgba(30, 40, 80, 0.04),
    inset 0 2rpx 0 rgba(255, 255, 255, 0.95);
  overflow: hidden;
}

.panel-head {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 24rpx 24rpx 8rpx;
}

.panel-accent {
  width: 8rpx;
  height: 32rpx;
  border-radius: 4rpx;
  background: linear-gradient(180deg, #6b7cff, #9b6bff);
  box-shadow: 0 4rpx 12rpx rgba(107, 124, 255, 0.4);
}

.panel-accent.accent-pink {
  background: linear-gradient(180deg, #ff8caa, #e07a9f);
  box-shadow: 0 4rpx 12rpx rgba(224, 122, 159, 0.4);
}

.panel-accent.accent-blue {
  background: linear-gradient(180deg, #6bb8ff, #6b7cff);
  box-shadow: 0 4rpx 12rpx rgba(107, 184, 255, 0.4);
}

.panel-title {
  font-size: 32rpx;
  font-weight: 800;
  color: #1a1f36;
  letter-spacing: 1rpx;
}

.category-panel {
  padding-bottom: 16rpx;
}

.category-grid {
  display: flex;
  flex-wrap: wrap;
  padding: 8rpx 12rpx 12rpx;
}

.category-item {
  width: 25%;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24rpx;
}

.cat-orb {
  position: relative;
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 3rpx solid rgba(255, 255, 255, 0.55);
  box-shadow:
    0 10rpx 24rpx rgba(30, 40, 80, 0.18),
    0 4rpx 10rpx rgba(30, 40, 80, 0.1),
    inset 0 4rpx 8rpx rgba(255, 255, 255, 0.45),
    inset 0 -6rpx 12rpx rgba(0, 0, 0, 0.12);
  transform: translateY(-2rpx);
}

.cat-orb-shine {
  position: absolute;
  top: 8rpx;
  left: 12rpx;
  width: 36rpx;
  height: 20rpx;
  background: rgba(255, 255, 255, 0.55);
  border-radius: 50%;
  filter: blur(4rpx);
  transform: rotate(-25deg);
  pointer-events: none;
}

.cat-icon-img {
  width: 52rpx;
  height: 52rpx;
  border-radius: 50%;
  position: relative;
  z-index: 1;
}

.cat-icon-text {
  font-size: 36rpx;
  color: #fff;
  font-weight: 800;
  position: relative;
  z-index: 1;
  text-shadow: 0 2rpx 6rpx rgba(0, 0, 0, 0.2);
}

.cat-name {
  margin-top: 14rpx;
  font-size: 24rpx;
  color: #3d4563;
  font-weight: 600;
  text-align: center;
}

.promo-section {
  margin-top: 28rpx;
}

.promo-bento {
  display: flex;
  gap: 16rpx;
  min-height: 340rpx;
  padding: 0 20rpx 20rpx;
}

.promo-main {
  position: relative;
  flex: 1;
  border-radius: 24rpx;
  overflow: hidden;
  background: #e8ecf4;
  border: 2rpx solid rgba(255, 255, 255, 0.7);
  box-shadow:
    0 16rpx 40rpx rgba(30, 40, 80, 0.15),
    0 6rpx 16rpx rgba(30, 40, 80, 0.08),
    inset 0 2rpx 4rpx rgba(255, 255, 255, 0.4);
  transform: perspective(800rpx) rotateY(-2deg);
}

.promo-bg {
  width: 100%;
  height: 100%;
  min-height: 340rpx;
}

.promo-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 30%, rgba(20, 30, 60, 0.55) 100%);
}

.promo-shine {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 40%;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.15) 0%, transparent 100%);
  pointer-events: none;
}

.promo-content {
  position: absolute;
  left: 24rpx;
  right: 24rpx;
  bottom: 24rpx;
}

.promo-tag {
  display: inline-block;
  padding: 6rpx 18rpx;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 20rpx;
  font-size: 22rpx;
  color: #6b7cff;
  font-weight: 700;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
}

.promo-title {
  display: block;
  margin-top: 14rpx;
  font-size: 30rpx;
  font-weight: 800;
  color: #fff;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.3);
}

.promo-sub {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.88);
}

.promo-side {
  width: 228rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.promo-sub-card {
  position: relative;
  flex: 1;
  min-height: 162rpx;
  border-radius: 22rpx;
  overflow: hidden;
  background: #e8ecf4;
  border: 2rpx solid rgba(255, 255, 255, 0.6);
  box-shadow:
    0 10rpx 28rpx rgba(30, 40, 80, 0.12),
    inset 0 2rpx 4rpx rgba(255, 255, 255, 0.35);
  transform: perspective(800rpx) rotateY(3deg);
}

.promo-sub-bg {
  width: 100%;
  height: 100%;
}

.promo-sub-mask {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 25%, rgba(20, 30, 60, 0.5) 100%);
}

.promo-sub-content {
  position: absolute;
  left: 16rpx;
  right: 16rpx;
  bottom: 16rpx;
}

.promo-sub-tag {
  display: inline-block;
  padding: 4rpx 14rpx;
  background: linear-gradient(135deg, #ffb3c6, #e07a9f);
  border-radius: 12rpx;
  font-size: 20rpx;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 4rpx 10rpx rgba(224, 122, 159, 0.35);
}

.promo-sub-title {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  font-weight: 700;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-shadow: 0 2rpx 6rpx rgba(0, 0, 0, 0.25);
}

.product-panel {
  padding: 24rpx 20rpx 16rpx;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
}

.section-head-left {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 800;
  color: #1a1f36;
}

.section-more-btn {
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 10rpx 20rpx;
  background: linear-gradient(145deg, #f8f9fc, #eef1f8);
  border-radius: 24rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  box-shadow:
    0 4rpx 12rpx rgba(30, 40, 80, 0.06),
    inset 0 2rpx 4rpx rgba(255, 255, 255, 0.9);
}

.section-more {
  font-size: 24rpx;
  color: #6b7cff;
  font-weight: 600;
}

.section-more-arrow {
  font-size: 28rpx;
  color: #6b7cff;
  font-weight: 700;
}

.product-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 20rpx;
}

.grid-item {
  width: calc(50% - 10rpx);
}

.empty {
  text-align: center;
  padding: 120rpx 0;
}

.empty-icon {
  display: block;
  font-size: 64rpx;
  margin-bottom: 16rpx;
}

.empty-text {
  color: #8b93a8;
  font-size: 28rpx;
}

.loading-tip {
  text-align: center;
  padding: 32rpx;
  color: #b0b8c8;
  font-size: 24rpx;
}
</style>
