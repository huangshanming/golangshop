<script setup lang="ts">
import { computed } from 'vue'
import { formatPriceYuan } from '../utils/format'

const props = withDefaults(defineProps<{
  price: number
  originalPrice?: number
  size?: 'sm' | 'md' | 'lg'
}>(), {
  originalPrice: 0,
  size: 'md',
})

const priceParts = computed(() => formatPriceYuan(props.price))
const showOriginal = computed(() => props.originalPrice > props.price)
</script>

<template>
  <view class="price-tag" :class="size">
    <text class="symbol">¥</text>
    <text class="integer">{{ priceParts.integer }}</text>
    <text class="decimal">.{{ priceParts.decimal }}</text>
    <text v-if="showOriginal" class="original">¥{{ (originalPrice / 100).toFixed(2) }}</text>
  </view>
</template>

<style scoped>
.price-tag {
  display: inline-flex;
  align-items: baseline;
  color: #ff4d4f;
  font-weight: 600;
}
.price-tag.sm { font-size: 28rpx; }
.price-tag.md { font-size: 32rpx; }
.price-tag.lg { font-size: 44rpx; }
.symbol { font-size: 0.75em; margin-right: 2rpx; }
.decimal { font-size: 0.75em; }
.original {
  margin-left: 12rpx;
  color: #999;
  font-size: 24rpx;
  font-weight: 400;
  text-decoration: line-through;
}
</style>
