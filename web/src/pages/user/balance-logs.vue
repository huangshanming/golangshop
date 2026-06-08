<script setup lang="ts">
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { bizTypeMap, fetchBalanceLogs, type BalanceLog } from '../../api/user'
import { formatDateTime, formatPrice } from '../../utils/format'

const list = ref<BalanceLog[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

onShow(() => {
  page.value = 1
  list.value = []
  finished.value = false
  loadList()
})

onReachBottom(() => {
  if (!finished.value && !loading.value) loadList()
})

async function loadList() {
  loading.value = true
  try {
    const data = await fetchBalanceLogs({ page: page.value, page_size: 20 })
    const rows = data.list || []
    total.value = data.total || 0
    if (page.value === 1) list.value = rows
    else list.value.push(...rows)
    if (list.value.length >= total.value || rows.length === 0) finished.value = true
    else page.value += 1
  } finally {
    loading.value = false
  }
}

function formatBiz(type: string) {
  return bizTypeMap[type] || type
}

function logDesc(item: BalanceLog) {
  if (item.goods_summary) return item.goods_summary
  return item.remark || item.biz_no || ''
}

function canGoOrder(item: BalanceLog) {
  return !!item.order_no
}

function goOrderDetail(item: BalanceLog) {
  if (!item.order_no) return
  uni.navigateTo({ url: `/pages/order/detail?order_no=${item.order_no}` })
}
</script>

<template>
  <view class="page">
    <view v-if="list.length === 0 && !loading" class="empty">暂无流水记录</view>
    <view
      v-for="item in list"
      :key="item.id"
      class="log-item"
      :class="{ clickable: canGoOrder(item) }"
      @click="goOrderDetail(item)"
    >
      <view class="log-main">
        <text class="log-title">{{ formatBiz(item.biz_type) }}</text>
        <text class="log-remark">{{ logDesc(item) }}</text>
        <text v-if="canGoOrder(item)" class="log-link">查看订单 ›</text>
        <text class="log-time">{{ formatDateTime(item.created_at) }}</text>
      </view>
      <view class="log-right">
        <text class="log-amount" :class="item.change_amount >= 0 ? 'plus' : 'minus'">
          {{ item.change_amount >= 0 ? '+' : '' }}{{ formatPrice(item.change_amount) }}
        </text>
        <text class="log-balance">余额 ¥{{ formatPrice(item.balance_after) }}</text>
      </view>
    </view>
    <view v-if="loading" class="loading-tip">加载中...</view>
    <view v-else-if="finished && list.length" class="loading-tip">没有更多了</view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16rpx;
}
.empty {
  text-align: center;
  padding: 120rpx 0;
  color: #999;
  font-size: 28rpx;
}
.log-item {
  display: flex;
  background: #fff;
  border-radius: 16rpx;
  padding: 28rpx 24rpx;
  margin-bottom: 16rpx;
}
.log-item.clickable:active {
  background: #fafafa;
}
.log-main {
  flex: 1;
  min-width: 0;
}
.log-title {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}
.log-remark {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.log-link {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #07c160;
}
.log-time {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #ccc;
}
.log-right {
  text-align: right;
  margin-left: 16rpx;
}
.log-amount {
  display: block;
  font-size: 32rpx;
  font-weight: 700;
}
.log-amount.plus {
  color: #07c160;
}
.log-amount.minus {
  color: #ff4d4f;
}
.log-balance {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #999;
}
.loading-tip {
  text-align: center;
  padding: 24rpx;
  color: #999;
  font-size: 24rpx;
}
</style>
