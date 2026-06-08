<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { fetchMarketplaceProducts, type ProductListItem } from '../../api/product'
import ProductCard from '../../components/ProductCard.vue'

const products = ref<ProductListItem[]>([])
const categoryId = ref(0)
const categoryName = ref('')
const keyword = ref('')
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const pageSize = 10

async function loadProducts(reset = false) {
  if (loading.value) return
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
      category_id: categoryId.value || undefined,
      keyword: keyword.value.trim() || undefined,
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

onLoad((query) => {
  categoryId.value = Number(query?.category_id || 0)
  categoryName.value = query?.category_name ? decodeURIComponent(String(query.category_name)) : ''
  keyword.value = query?.keyword ? decodeURIComponent(String(query.keyword)) : ''
  const title = categoryName.value || (keyword.value ? `搜索：${keyword.value}` : '商品列表')
  uni.setNavigationBarTitle({ title })
  loadProducts(true)
})

onPullDownRefresh(() => loadProducts(true))
onReachBottom(() => loadProducts(false))
</script>

<template>
  <view class="page">
    <view v-if="products.length" class="grid">
      <view v-for="item in products" :key="`${item.tenant_id}-${item.id}`" class="grid-item">
        <ProductCard :item="item" />
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-text">暂无商品</text>
    </view>

    <view v-if="loading" class="loading-tip">加载中...</view>
    <view v-else-if="finished && products.length > 0" class="loading-tip">— 没有更多了 —</view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16rpx 12rpx 24rpx;
}
.grid {
  display: flex;
  flex-wrap: wrap;
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
