<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { fetchPublicTenants } from '../../api/tenant'
import { fetchMarketplaceProducts, type ProductListItem } from '../../api/product'
import ProductCard from '../../components/ProductCard.vue'

const tenantId = ref(0)
const tenantName = ref('')
const products = ref<ProductListItem[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const pageSize = 10

async function loadTenantName() {
  try {
    const tenants = await fetchPublicTenants()
    const t = tenants.find((x) => x.id === tenantId.value)
    if (t) {
      tenantName.value = t.name
      uni.setNavigationBarTitle({ title: t.name })
    }
  } catch {
    /* ignore */
  }
}

async function loadProducts(reset = false) {
  if (loading.value || !tenantId.value) return
  if (reset) {
    page.value = 1
    finished.value = false
    products.value = []
  }
  if (finished.value) return

  loading.value = true
  try {
    const data = await fetchMarketplaceProducts({
      page: page.value,
      page_size: pageSize,
      tenant_id: tenantId.value,
    })
    const list = data.list || []
    total.value = data.total || 0
    products.value = reset ? list : [...products.value, ...list]
    if (products.value.length >= total.value || list.length < pageSize) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch {
    if (reset) products.value = []
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

onLoad(async (query) => {
  tenantId.value = Number(query?.tenant_id || 0)
  await loadTenantName()
  loadProducts(true)
})

onPullDownRefresh(() => loadProducts(true))
onReachBottom(() => loadProducts(false))
</script>

<template>
  <view class="page">
    <view v-if="tenantName" class="shop-header">
      <text class="shop-name">{{ tenantName }}</text>
      <text class="shop-desc">店铺商品</text>
    </view>

    <view v-if="products.length" class="grid">
      <view v-for="item in products" :key="item.id" class="grid-item">
        <ProductCard :item="item" />
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-text">该店铺暂无商品</text>
    </view>

    <view v-if="loading" class="loading-tip">加载中...</view>
    <view v-else-if="finished && products.length > 0" class="loading-tip">— 没有更多了 —</view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 24rpx;
}
.shop-header {
  background: linear-gradient(135deg, #07c160, #06ae56);
  padding: 32rpx 24rpx;
  color: #fff;
}
.shop-name {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
}
.shop-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  opacity: 0.85;
}
.grid {
  display: flex;
  flex-wrap: wrap;
  padding: 16rpx 12rpx;
  gap: 16rpx;
}
.grid-item {
  width: calc(50% - 8rpx);
}
.empty {
  text-align: center;
  padding-top: 160rpx;
}
.empty-text {
  color: #999;
  font-size: 28rpx;
}
.loading-tip {
  text-align: center;
  padding: 24rpx;
  color: #bbb;
  font-size: 24rpx;
}
</style>
