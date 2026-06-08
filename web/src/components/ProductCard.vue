<script setup lang="ts">
import type { ProductListItem } from '../api/product'
import PriceTag from './PriceTag.vue'

withDefaults(defineProps<{
  item: ProductListItem
  compact?: boolean
}>(), {
  compact: false,
})

function goDetail(item: ProductListItem) {
  const tenantId = item.tenant_id || 0
  uni.navigateTo({ url: `/pages/product/detail?id=${item.id}&tenant_id=${tenantId}` })
}
</script>

<template>
  <view class="card-wrap" :class="{ compact }">
    <view class="card" @click="goDetail(item)">
      <view class="cover-wrap">
        <image class="cover" :src="item.cover_image" mode="aspectFill" lazy-load />
        <view class="cover-shine" />
      </view>
      <view class="body">
        <template v-if="compact">
          <text class="name">{{ item.name }}</text>
          <PriceTag class="compact-price" :price="item.min_price" size="sm" />
        </template>
        <template v-else>
          <text v-if="item.tenant_name" class="shop-name">{{ item.tenant_name }}</text>
          <text class="name">{{ item.name }}</text>
          <text v-if="item.subtitle" class="subtitle">{{ item.subtitle }}</text>
          <view class="footer">
            <PriceTag :price="item.min_price" :original-price="item.original_price" size="sm" />
            <text class="sales">已售 {{ item.sales_count || 0 }}</text>
          </view>
        </template>
      </view>
    </view>
  </view>
</template>

<style scoped>
.card-wrap {
  position: relative;
}

.card-wrap.compact {
  filter: drop-shadow(0 10rpx 24rpx rgba(30, 40, 80, 0.12));
}

.card {
  background: linear-gradient(180deg, #ffffff 0%, #fafbff 100%);
  border-radius: 20rpx;
  overflow: hidden;
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  box-shadow:
    0 12rpx 32rpx rgba(30, 40, 80, 0.1),
    0 4rpx 12rpx rgba(30, 40, 80, 0.06),
    inset 0 2rpx 0 rgba(255, 255, 255, 0.95);
}

.cover-wrap {
  position: relative;
  overflow: hidden;
}

.cover {
  width: 100%;
  height: 340rpx;
  background: linear-gradient(135deg, #eef1f8, #e4e9f2);
  display: block;
}

.cover-shine {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 50%;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.28) 0%, transparent 100%);
  pointer-events: none;
}

.body {
  padding: 20rpx;
}

.shop-name {
  display: block;
  margin-bottom: 8rpx;
  font-size: 22rpx;
  color: #6b7cff;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.name {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #1a1f36;
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
  color: #8b93a8;
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
  color: #b0b8c8;
}

.card-wrap.compact .card {
  border-radius: 18rpx;
}

.card-wrap.compact .cover {
  height: 300rpx;
}

.card-wrap.compact .body {
  padding: 16rpx 14rpx 18rpx;
}

.card-wrap.compact .name {
  font-size: 26rpx;
  font-weight: 600;
  -webkit-line-clamp: 1;
  margin-bottom: 6rpx;
}

.card-wrap.compact .compact-price {
  margin-top: 2rpx;
}
</style>
