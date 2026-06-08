<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { addToCart } from '../../api/cart'
import {
  fetchProductDetail,
  matchSkuBySelection,
  parseImages,
  parseSpecValues,
  parseSpecs,
  type ProductListItem,
  type ProductSKU,
  type SpecDimension,
} from '../../api/product'
import PriceTag from '../../components/PriceTag.vue'
import { useUserStore } from '../../stores/user'

const userStore = useUserStore()
const product = ref<ProductListItem | null>(null)
const skus = ref<ProductSKU[]>([])
const specDims = ref<SpecDimension[]>([])
const selectedSpecs = ref<Record<string, string>>({})
const loading = ref(true)
const adding = ref(false)
let productId = 0

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

const allSpecsSelected = computed(() => {
  if (!useMatrixSpecs.value) return selectedSkuId.value > 0
  return specDims.value.every((d) => !!selectedSpecs.value[d.name])
})

const canAdd = computed(() => selectedSku.value && selectedSku.value.stock > 0 && !adding.value && allSpecsSelected.value)

onLoad((query) => {
  productId = Number(query?.id || 0)
  if (productId) loadDetail()
})

async function loadDetail() {
  loading.value = true
  try {
    const data = await fetchProductDetail(productId)
    product.value = data.product
    skus.value = data.skus || []
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

async function handleAddCart() {
  if (!selectedSku.value) {
    uni.showToast({ title: useMatrixSpecs.value ? '请选择完整规格' : '请选择规格', icon: 'none' })
    return
  }
  if (!userStore.token) {
    uni.navigateTo({ url: `/pages/login/index?redirect=${encodeURIComponent(`/pages/product/detail?id=${productId}`)}` })
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
</script>

<template>
  <view v-if="product" class="page">
    <swiper class="swiper" circular indicator-dots indicator-color="rgba(255,255,255,0.5)" indicator-active-color="#fff">
      <swiper-item v-for="(img, idx) in imageList" :key="idx">
        <image class="swiper-img" :src="idx === 0 && displayImage ? displayImage : img" mode="aspectFill" />
      </swiper-item>
    </swiper>

    <view class="info-card">
      <view class="price-row">
        <PriceTag :price="displayPrice" :original-price="displayOriginal" size="lg" />
        <text class="sales">已售 {{ product.sales_count || 0 }}</text>
      </view>
      <text class="title">{{ product.name }}</text>
      <text v-if="product.subtitle" class="subtitle">{{ product.subtitle }}</text>
    </view>

    <view class="section">
      <template v-if="useMatrixSpecs">
        <view v-for="dim in specDims" :key="dim.name" class="spec-dim">
          <view class="section-title">{{ dim.name }}</view>
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
        <view class="section-title">选择规格</view>
        <view class="sku-list">
          <view
            v-for="sku in skus"
            :key="sku.id"
            class="sku-chip"
            :class="{ active: selectedSkuId === sku.id, disabled: sku.stock <= 0 }"
            @click="selectSku(sku)"
          >
            {{ sku.spec }}
          </view>
        </view>
      </template>
      <text v-if="selectedSku" class="stock-tip">
        {{ selectedSku.stock > 0 ? `库存 ${selectedSku.stock} 件` : '该规格已售罄' }}
      </text>
      <text v-else class="stock-tip">请选择规格</text>
    </view>

    <view class="section">
      <view class="section-title">商品详情</view>
      <rich-text v-if="product.description" class="rich-desc" :nodes="product.description" />
      <text v-else class="desc">暂无详情</text>
    </view>

    <view class="bottom-bar">
      <button class="add-btn" :class="{ disabled: !canAdd }" :disabled="!canAdd" @click="handleAddCart">
        {{ selectedSku && selectedSku.stock <= 0 ? '已售罄' : adding ? '加入中...' : '加入购物车' }}
      </button>
    </view>
  </view>

  <view v-else-if="!loading" class="empty">商品不存在</view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}
.swiper {
  height: 750rpx;
  background: #fff;
}
.swiper-img {
  width: 100%;
  height: 100%;
}
.info-card {
  background: #fff;
  padding: 28rpx;
  margin-bottom: 16rpx;
}
.price-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.sales {
  font-size: 24rpx;
  color: #999;
}
.title {
  display: block;
  margin-top: 20rpx;
  font-size: 34rpx;
  font-weight: 700;
  color: #333;
  line-height: 1.4;
}
.subtitle {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #888;
}
.section {
  background: #fff;
  padding: 28rpx;
  margin-bottom: 16rpx;
}
.spec-dim {
  margin-bottom: 20rpx;
}
.section-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 20rpx;
}
.sku-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.sku-chip {
  padding: 14rpx 28rpx;
  font-size: 26rpx;
  color: #333;
  background: #f5f5f5;
  border-radius: 8rpx;
  border: 2rpx solid transparent;
}
.sku-chip.active {
  color: #07c160;
  background: #e8f8ef;
  border-color: #07c160;
}
.sku-chip.disabled {
  color: #ccc;
  text-decoration: line-through;
}
.stock-tip {
  display: block;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: #999;
}
.rich-desc {
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
}
.desc {
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
  white-space: pre-wrap;
}
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
}
.add-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background: #07c160;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  border-radius: 44rpx;
  border: none;
}
.add-btn.disabled {
  background: #ccc;
}
.empty {
  text-align: center;
  padding-top: 200rpx;
  color: #999;
}
</style>
