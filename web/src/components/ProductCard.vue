<script setup lang="ts">
import type { ProductListItem } from '../api/product'
import PriceTag from './PriceTag.vue'

defineProps<{
  item: ProductListItem
}>()

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/product/detail?id=${id}` })
}
</script>

<template>
  <view class="card" @click="goDetail(item.id)">
    <image class="cover" :src="item.cover_image" mode="aspectFill" lazy-load />
    <view class="body">
      <text class="name">{{ item.name }}</text>
      <text v-if="item.subtitle" class="subtitle">{{ item.subtitle }}</text>
      <view class="footer">
        <PriceTag :price="item.min_price" :original-price="item.original_price" size="sm" />
        <text class="sales">已售 {{ item.sales_count || 0 }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.card {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.06);
}
.cover {
  width: 100%;
  height: 340rpx;
  background: #f0f0f0;
  display: block;
}
.body {
  padding: 20rpx;
}
.name {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.subtitle {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16rpx;
}
.sales {
  font-size: 22rpx;
  color: #bbb;
}
</style>
