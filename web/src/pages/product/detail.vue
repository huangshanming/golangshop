<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { addToCart } from '../../api/cart'
import {
  fetchProductDetailPublic,
  matchSkuBySelection,
  parseImages,
  parseSpecValues,
  parseSpecs,
  type ProductListItem,
  type ProductSKU,
  type SpecDimension,
} from '../../api/product'
import PriceTag from '../../components/PriceTag.vue'
import { formatPriceYuan } from '../../utils/format'
import { useUserStore } from '../../stores/user'
import { useTenantStore } from '../../stores/tenant'

const userStore = useUserStore()
const tenantStore = useTenantStore()
const product = ref<ProductListItem | null>(null)
const skus = ref<ProductSKU[]>([])
const specDims = ref<SpecDimension[]>([])
const selectedSpecs = ref<Record<string, string>>({})
const loading = ref(true)
const adding = ref(false)
const currentImage = ref(0)
let productId = 0
let tenantId = 0

function ensureTradeTenant() {
  if (!tenantId) return
  const changed = tenantStore.setTenantById(
    tenantId,
    product.value?.tenant_name || tenantStore.tenantName,
  )
  if (changed && userStore.token) {
    userStore.logout()
  }
}

const imageList = computed(() => {
  if (!product.value) return []
  return parseImages(product.value.images, product.value.cover_image)
})

const useMatrixSpecs = computed(() => specDims.value.length > 0)

const selectedSku = computed(() => {
  if (useMatrixSpecs.value) {
    return matchSkuBySelection(skus.value, selectedSpecs.value)
  }
  return skus.value.find((s) => s.id === selectedSkuId.value) || null
})

const selectedSkuId = ref(0)

const displayPrice = computed(() => selectedSku.value?.price ?? product.value?.min_price ?? 0)
const displayOriginal = computed(() => selectedSku.value?.original_price ?? product.value?.original_price ?? 0)

const displayImage = computed(() => {
  if (selectedSku.value?.image) return selectedSku.value.image
  return product.value?.cover_image || imageList.value[0] || ''
})

const swiperImages = computed(() => {
  const imgs = [...imageList.value]
  if (displayImage.value && imgs[0] !== displayImage.value) {
    imgs[0] = displayImage.value
  }
  return imgs.length ? imgs : [displayImage.value || '']
})

const allSpecsSelected = computed(() => {
  if (!useMatrixSpecs.value) return selectedSkuId.value > 0
  return specDims.value.every((d) => !!selectedSpecs.value[d.name])
})

const canAdd = computed(() => selectedSku.value && selectedSku.value.stock > 0 && !adding.value && allSpecsSelected.value)

const discountPercent = computed(() => {
  if (!displayOriginal.value || displayOriginal.value <= displayPrice.value) return 0
  return Math.round((1 - displayPrice.value / displayOriginal.value) * 100)
})

const saveAmount = computed(() => {
  if (!displayOriginal.value || displayOriginal.value <= displayPrice.value) return 0
  return displayOriginal.value - displayPrice.value
})

const priceParts = computed(() => formatPriceYuan(displayPrice.value))

onLoad((query) => {
  productId = Number(query?.id || 0)
  tenantId = Number(query?.tenant_id || 0)
  if (productId) loadDetail()
})

async function loadDetail() {
  loading.value = true
  try {
    const data = await fetchProductDetailPublic(productId)
    product.value = data.product
    skus.value = data.skus || []
    if (!tenantId && data.product.tenant_id) {
      tenantId = data.product.tenant_id
    }
    ensureTradeTenant()
    specDims.value = parseSpecs(data.product.specs)
    selectedSpecs.value = {}
    selectedSkuId.value = 0
    if (specDims.value.length === 0 && skus.value.length === 1) {
      selectedSkuId.value = skus.value[0].id
    }
  } catch {
    uni.showToast({ title: '商品不存在', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function onSwiperChange(e: { detail: { current: number } }) {
  currentImage.value = e.detail.current
}

function selectSpecValue(dimName: string, value: string) {
  selectedSpecs.value = { ...selectedSpecs.value, [dimName]: value }
}

function isSpecDisabled(dimName: string, value: string): boolean {
  const trial = { ...selectedSpecs.value, [dimName]: value }
  return !skus.value.some((sku) => {
    const sv = parseSpecValues(sku.spec_values)
    if (Object.keys(sv).length === 0) return sku.spec === value && sku.stock > 0
    return Object.entries(trial).every(([k, v]) => sv[k] === v) && sku.stock > 0
  })
}

function selectSku(sku: ProductSKU) {
  if (sku.stock <= 0) return
  selectedSkuId.value = sku.id
}

watch(selectedSku, (sku) => {
  if (sku && !useMatrixSpecs.value) selectedSkuId.value = sku.id
})

function goHome() {
  uni.switchTab({ url: '/pages/index/index' })
}

function goCart() {
  uni.navigateTo({ url: '/pages/cart/index' })
}

async function handleAddCart() {
  if (!selectedSku.value) {
    uni.showToast({ title: useMatrixSpecs.value ? '请选择完整规格' : '请选择规格', icon: 'none' })
    return
  }
  ensureTradeTenant()
  if (!userStore.token) {
    uni.navigateTo({
      url: `/pages/login/index?tenant_id=${tenantId}&redirect=${encodeURIComponent(`/pages/product/detail?id=${productId}&tenant_id=${tenantId}`)}`,
    })
    return
  }
  adding.value = true
  try {
    await addToCart({
      product_id: productId,
      sku_id: selectedSku.value.id,
      quantity: 1,
    })
    uni.showToast({ title: '已加入购物车', icon: 'success' })
  } catch {
    /* handled */
  } finally {
    adding.value = false
  }
}

function handleBuyNow() {
  if (!selectedSku.value) {
    uni.showToast({ title: useMatrixSpecs.value ? '请选择完整规格' : '请选择规格', icon: 'none' })
    return
  }
  if (selectedSku.value.stock <= 0) {
    uni.showToast({ title: '该规格已售罄', icon: 'none' })
    return
  }
  ensureTradeTenant()
  if (!userStore.token) {
    uni.navigateTo({
      url: `/pages/login/index?tenant_id=${tenantId}&redirect=${encodeURIComponent(`/pages/product/detail?id=${productId}&tenant_id=${tenantId}`)}`,
    })
    return
  }
  uni.setStorageSync('checkout_payload', {
    items: [{ sku_id: selectedSku.value.id, quantity: 1 }],
    cart_item_ids: [],
  })
  uni.navigateTo({ url: '/pages/checkout/index' })
}
</script>

<template>
  <view v-if="product" class="page">
    <!-- 主图轮播 -->
    <view class="hero">
      <swiper
        class="swiper"
        circular
        :current="currentImage"
        @change="onSwiperChange"
      >
        <swiper-item v-for="(img, idx) in swiperImages" :key="idx">
          <image class="swiper-img" :src="img" mode="aspectFill" />
        </swiper-item>
      </swiper>
      <view class="hero-gradient" />
      <view v-if="swiperImages.length > 1" class="hero-indicator">
        <text>{{ currentImage + 1 }}/{{ swiperImages.length }}</text>
      </view>
      <view v-if="discountPercent > 0" class="hero-badge">-{{ discountPercent }}%</view>
    </view>

    <!-- 价格信息卡 -->
    <view class="price-card">
      <view class="price-main">
        <view class="price-left">
          <view class="price-row">
            <text class="price-symbol">¥</text>
            <text class="price-int">{{ priceParts.integer }}</text>
            <text class="price-dec">.{{ priceParts.decimal }}</text>
            <text v-if="displayOriginal > displayPrice" class="price-origin">¥{{ (displayOriginal / 100).toFixed(2) }}</text>
          </view>
          <view v-if="saveAmount > 0" class="save-tag">立省 ¥{{ (saveAmount / 100).toFixed(2) }}</view>
        </view>
        <view class="price-right">
          <text class="sales-num">{{ product.sales_count || 0 }}</text>
          <text class="sales-label">已售</text>
        </view>
      </view>
      <text v-if="product.tenant_name" class="shop-name">{{ product.tenant_name }}</text>
      <text class="product-name">{{ product.name }}</text>
      <text v-if="product.subtitle" class="product-sub">{{ product.subtitle }}</text>
    </view>

    <!-- 服务保障 -->
    <view class="service-bar">
      <view class="service-item">
        <text class="service-icon">✓</text>
        <text>正品保障</text>
      </view>
      <view class="service-item">
        <text class="service-icon">✓</text>
        <text>满99包邮</text>
      </view>
      <view class="service-item">
        <text class="service-icon">✓</text>
        <text>7天退换</text>
      </view>
      <view class="service-item">
        <text class="service-icon">✓</text>
        <text>极速发货</text>
      </view>
    </view>

    <!-- 规格选择 -->
    <view class="spec-card">
      <view class="card-header">
        <text class="card-title">选择规格</text>
        <text v-if="selectedSku" class="stock-badge" :class="{ out: selectedSku.stock <= 0 }">
          {{ selectedSku.stock > 0 ? `库存 ${selectedSku.stock}` : '已售罄' }}
        </text>
        <text v-else class="stock-badge warn">请选择</text>
      </view>

      <template v-if="useMatrixSpecs">
        <view v-for="dim in specDims" :key="dim.name" class="spec-dim">
          <text class="dim-label">{{ dim.name }}</text>
          <view class="sku-list">
            <view
              v-for="val in dim.values"
              :key="val"
              class="sku-chip"
              :class="{
                active: selectedSpecs[dim.name] === val,
                disabled: isSpecDisabled(dim.name, val),
              }"
              @click="!isSpecDisabled(dim.name, val) && selectSpecValue(dim.name, val)"
            >
              {{ val }}
            </view>
          </view>
        </view>
      </template>
      <template v-else>
        <view class="sku-list">
          <view
            v-for="sku in skus"
            :key="sku.id"
            class="sku-chip lg"
            :class="{ active: selectedSkuId === sku.id, disabled: sku.stock <= 0 }"
            @click="selectSku(sku)"
          >
            <text>{{ sku.spec }}</text>
            <text v-if="sku.stock <= 0" class="sold-out">售罄</text>
          </view>
        </view>
      </template>
    </view>

    <!-- 商品详情 -->
    <view class="detail-card">
      <view class="detail-header">
        <view class="detail-line" />
        <text class="detail-title">商品详情</text>
        <view class="detail-line" />
      </view>
      <rich-text v-if="product.description" class="rich-desc" :nodes="product.description" />
      <view v-else class="desc-empty">
        <text class="desc-empty-icon">📦</text>
        <text>暂无详情描述</text>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="bottom-bar">
      <view class="bar-icons">
        <view class="icon-btn" @click="goHome">
          <text class="icon-emoji">🏠</text>
          <text class="icon-label">首页</text>
        </view>
        <view class="icon-btn" @click="goCart">
          <text class="icon-emoji">🛒</text>
          <text class="icon-label">购物车</text>
        </view>
      </view>
      <view class="bar-actions">
        <button class="cart-btn" :class="{ disabled: !canAdd }" :disabled="!canAdd" @click="handleAddCart">
          {{ selectedSku && selectedSku.stock <= 0 ? '已售罄' : adding ? '加入中...' : '加入购物车' }}
        </button>
        <button class="buy-btn" :class="{ disabled: !canAdd }" :disabled="!canAdd" @click="handleBuyNow">
          {{ selectedSku && selectedSku.stock <= 0 ? '已售罄' : '立即购买' }}
        </button>
      </view>
    </view>
  </view>

  <view v-else-if="loading" class="loading-page">
    <view class="loading-spinner" />
    <text>加载中...</text>
  </view>
  <view v-else class="empty">商品不存在</view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
}

/* Hero */
.hero {
  position: relative;
  background: #fff;
}
.swiper {
  height: 750rpx;
}
.swiper-img {
  width: 100%;
  height: 100%;
}
.hero-gradient {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 120rpx;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.15));
  pointer-events: none;
}
.hero-indicator {
  position: absolute;
  right: 24rpx;
  bottom: 24rpx;
  padding: 6rpx 20rpx;
  background: rgba(0, 0, 0, 0.45);
  border-radius: 20rpx;
  font-size: 22rpx;
  color: #fff;
}
.hero-badge {
  position: absolute;
  left: 24rpx;
  top: 24rpx;
  padding: 8rpx 20rpx;
  background: linear-gradient(135deg, #ff6b6b, #ff4d4f);
  border-radius: 8rpx 20rpx 20rpx 8rpx;
  font-size: 26rpx;
  font-weight: 700;
  color: #fff;
  box-shadow: 0 4rpx 16rpx rgba(255, 77, 79, 0.4);
}

/* Price card */
.price-card {
  margin: -40rpx 24rpx 0;
  padding: 32rpx;
  background: #fff;
  border-radius: 24rpx;
  box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.08);
  position: relative;
  z-index: 2;
}
.price-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}
.price-row {
  display: flex;
  align-items: baseline;
}
.price-symbol {
  font-size: 32rpx;
  font-weight: 700;
  color: #ff4d4f;
}
.price-int {
  font-size: 56rpx;
  font-weight: 800;
  color: #ff4d4f;
  line-height: 1;
}
.price-dec {
  font-size: 32rpx;
  font-weight: 700;
  color: #ff4d4f;
}
.price-origin {
  margin-left: 16rpx;
  font-size: 26rpx;
  color: #bbb;
  text-decoration: line-through;
}
.save-tag {
  display: inline-block;
  margin-top: 12rpx;
  padding: 4rpx 16rpx;
  background: linear-gradient(90deg, #fff0f0, #ffe8e8);
  border-radius: 6rpx;
  font-size: 22rpx;
  color: #ff4d4f;
  font-weight: 600;
}
.price-right {
  text-align: center;
  padding: 8rpx 20rpx;
  background: #f7f8fa;
  border-radius: 12rpx;
}
.sales-num {
  display: block;
  font-size: 32rpx;
  font-weight: 700;
  color: #333;
}
.sales-label {
  display: block;
  font-size: 20rpx;
  color: #999;
  margin-top: 2rpx;
}
.shop-name {
  display: block;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #07c160;
}
.product-name {
  display: block;
  margin-top: 12rpx;
  font-size: 34rpx;
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1.5;
}
.product-sub {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #888;
  line-height: 1.5;
}

/* Service bar */
.service-bar {
  display: flex;
  margin: 20rpx 24rpx 0;
  padding: 24rpx 16rpx;
  background: #fff;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
}
.service-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 22rpx;
  color: #666;
}
.service-icon {
  width: 36rpx;
  height: 36rpx;
  line-height: 36rpx;
  text-align: center;
  background: #e8f8ef;
  color: #07c160;
  border-radius: 50%;
  font-size: 20rpx;
  margin-bottom: 8rpx;
}

/* Spec card */
.spec-card {
  margin: 20rpx 24rpx 0;
  padding: 28rpx;
  background: #fff;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
}
.card-header {
  display: flex;
  align-items: center;
  margin-bottom: 24rpx;
}
.card-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #1a1a1a;
}
.stock-badge {
  margin-left: auto;
  padding: 4rpx 16rpx;
  background: #e8f8ef;
  color: #07c160;
  font-size: 22rpx;
  border-radius: 20rpx;
  font-weight: 500;
}
.stock-badge.warn {
  background: #fff7e6;
  color: #fa8c16;
}
.stock-badge.out {
  background: #fff1f0;
  color: #ff4d4f;
}
.spec-dim {
  margin-bottom: 24rpx;
}
.spec-dim:last-child {
  margin-bottom: 0;
}
.dim-label {
  display: block;
  font-size: 26rpx;
  color: #666;
  margin-bottom: 16rpx;
}
.sku-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.sku-chip {
  position: relative;
  padding: 16rpx 32rpx;
  font-size: 26rpx;
  color: #333;
  background: #f7f8fa;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  transition: all 0.2s;
}
.sku-chip.lg {
  min-width: 120rpx;
  text-align: center;
}
.sku-chip.active {
  color: #07c160;
  background: linear-gradient(135deg, #e8f8ef, #d4f5e0);
  border-color: #07c160;
  font-weight: 600;
  box-shadow: 0 4rpx 12rpx rgba(7, 193, 96, 0.2);
}
.sku-chip.disabled {
  color: #ccc;
  background: #fafafa;
  text-decoration: line-through;
}
.sold-out {
  display: block;
  font-size: 20rpx;
  color: #ff4d4f;
  margin-top: 4rpx;
}

/* Detail */
.detail-card {
  margin: 20rpx 24rpx 0;
  padding: 32rpx 28rpx;
  background: #fff;
  border-radius: 20rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
}
.detail-header {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 32rpx;
}
.detail-line {
  flex: 1;
  height: 2rpx;
  background: linear-gradient(90deg, transparent, #e8e8e8, transparent);
}
.detail-title {
  padding: 0 24rpx;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}
.rich-desc {
  font-size: 28rpx;
  color: #555;
  line-height: 1.9;
}
.desc-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 0;
  color: #ccc;
  font-size: 26rpx;
}
.desc-empty-icon {
  font-size: 64rpx;
  margin-bottom: 16rpx;
}

/* Bottom bar */
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  padding: 12rpx 16rpx;
  padding-bottom: calc(12rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  box-shadow: 0 -4rpx 24rpx rgba(0, 0, 0, 0.08);
  z-index: 100;
}
.bar-icons {
  display: flex;
  gap: 8rpx;
  margin-right: 16rpx;
}
.icon-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 16rpx;
}
.icon-emoji {
  font-size: 40rpx;
  line-height: 1;
}
.icon-label {
  font-size: 20rpx;
  color: #666;
  margin-top: 4rpx;
}
.bar-actions {
  flex: 1;
  display: flex;
  gap: 12rpx;
}
.cart-btn,
.buy-btn {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  font-size: 28rpx;
  font-weight: 600;
  border-radius: 40rpx;
  border: none;
  padding: 0;
}
.cart-btn {
  background: linear-gradient(135deg, #fff, #f0fff4);
  color: #07c160;
  border: 2rpx solid #07c160;
}
.buy-btn {
  background: linear-gradient(135deg, #07c160, #06ae56);
  color: #fff;
  box-shadow: 0 6rpx 20rpx rgba(7, 193, 96, 0.35);
}
.cart-btn.disabled,
.buy-btn.disabled {
  background: #e8e8e8;
  color: #bbb;
  border-color: #e8e8e8;
  box-shadow: none;
}

.loading-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  color: #999;
  font-size: 28rpx;
}
.loading-spinner {
  width: 60rpx;
  height: 60rpx;
  border: 4rpx solid #e8e8e8;
  border-top-color: #07c160;
  border-radius: 50%;
  margin-bottom: 24rpx;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.empty {
  text-align: center;
  padding-top: 200rpx;
  color: #999;
}
</style>
