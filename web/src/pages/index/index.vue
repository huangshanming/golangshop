<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import { fetchCategories, fetchProducts, type Category, type ProductListItem } from '../../api/product'
import ProductCard from '../../components/ProductCard.vue'

const categories = ref<Category[]>([])
const products = ref<ProductListItem[]>([])
const activeCategory = ref(0)
const keyword = ref('')
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const pageSize = 10

async function loadCategories() {
  try {
    categories.value = await fetchCategories()
  } catch {
    categories.value = []
  }
}

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
    const data = await fetchProducts({
      page: page.value,
      page_size: pageSize,
      category_id: activeCategory.value || undefined,
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

function selectCategory(id: number) {
  activeCategory.value = id
  loadProducts(true)
}

function onSearch() {
  loadProducts(true)
}

onLoad(() => {
  loadCategories()
  loadProducts(true)
})

onPullDownRefresh(() => loadProducts(true))
onReachBottom(() => loadProducts(false))
</script>

<template>
  <view class="page">
    <view class="search-bar">
      <view class="search-inner">
        <text class="search-icon">🔍</text>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索商品"
          confirm-type="search"
          @confirm="onSearch"
        />
      </view>
    </view>

    <scroll-view scroll-x class="category-scroll" :show-scrollbar="false">
      <view class="category-list">
        <view
          class="category-item"
          :class="{ active: activeCategory === 0 }"
          @click="selectCategory(0)"
        >
          全部
        </view>
        <view
          v-for="cat in categories"
          :key="cat.id"
          class="category-item"
          :class="{ active: activeCategory === cat.id }"
          @click="selectCategory(cat.id)"
        >
          {{ cat.name }}
        </view>
      </view>
    </scroll-view>

    <view v-if="products.length > 0" class="grid">
      <view v-for="item in products" :key="item.id" class="grid-item">
        <ProductCard :item="item" />
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-icon">🛍️</text>
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
  padding-bottom: 24rpx;
}
.search-bar {
  padding: 20rpx 24rpx;
  background: #fff;
}
.search-inner {
  display: flex;
  align-items: center;
  background: #f5f5f5;
  border-radius: 40rpx;
  padding: 16rpx 24rpx;
}
.search-icon {
  font-size: 28rpx;
  margin-right: 12rpx;
}
.search-input {
  flex: 1;
  font-size: 28rpx;
}
.category-scroll {
  background: #fff;
  white-space: nowrap;
  border-bottom: 1rpx solid #eee;
}
.category-list {
  display: inline-flex;
  padding: 16rpx 24rpx;
  gap: 16rpx;
}
.category-item {
  display: inline-block;
  padding: 12rpx 28rpx;
  font-size: 26rpx;
  color: #666;
  background: #f5f5f5;
  border-radius: 32rpx;
  flex-shrink: 0;
}
.category-item.active {
  color: #fff;
  background: #07c160;
  font-weight: 600;
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
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 160rpx;
}
.empty-icon {
  font-size: 80rpx;
}
.empty-text {
  margin-top: 24rpx;
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
