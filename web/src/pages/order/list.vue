<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { fetchOrders, type Order } from '../../api/order'
import PriceTag from '../../components/PriceTag.vue'
import { formatDateTime, formatPrice } from '../../utils/format'
import { useUserStore } from '../../stores/user'

const userStore = useUserStore()

const tabs = [
  { key: '', label: '全部' },
  { key: 'pending_pay', label: '待付款' },
  { key: 'paid', label: '待发货' },
  { key: 'shipped', label: '待收货' },
  { key: 'review', label: '待评价' },
]

const activeTab = ref('')
const orders = ref<Order[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)

onShow(() => {
  if (userStore.token) {
    loadOrders(true)
  } else {
    orders.value = []
  }
})

async function loadOrders(reset = false) {
  if (reset) {
    page.value = 1
    orders.value = []
  }
  loading.value = true
  try {
    const data = await fetchOrders({
      page: page.value,
      page_size: 20,
      tab: activeTab.value || undefined,
    })
    orders.value = reset ? data.list : [...orders.value, ...data.list]
    total.value = data.total
  } catch {
    if (reset) orders.value = []
  } finally {
    loading.value = false
  }
}

function switchTab(key: string) {
  if (activeTab.value === key) return
  activeTab.value = key
  loadOrders(true)
}

function loadMore() {
  if (loading.value || orders.value.length >= total.value) return
  page.value += 1
  loadOrders(false)
}

function goDetail(orderNo: string) {
  uni.navigateTo({ url: `/pages/order/detail?order_no=${orderNo}` })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index?redirect=' + encodeURIComponent('/pages/order/list') })
}
</script>

<template>
  <view class="page">
    <view class="tabs">
      <view
        v-for="tab in tabs"
        :key="tab.key"
        class="tab"
        :class="{ active: activeTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
      </view>
    </view>

    <view v-if="!userStore.token" class="guest">
      <text class="guest-text">登录后查看订单</text>
      <button class="login-btn" @click="goLogin">去登录</button>
    </view>

    <view v-else-if="orders.length > 0" class="list">
      <view v-for="order in orders" :key="order.id" class="order-card" @click="goDetail(order.order_no)">
        <view class="order-header">
          <view class="order-meta">
            <text class="order-no">{{ order.order_no }}</text>
            <text class="order-time">{{ formatDateTime(order.created_at) }}</text>
          </view>
          <text class="status">{{ order.display_status }}</text>
        </view>
        <view v-for="item in order.items" :key="item.id" class="goods-row">
          <image class="cover" :src="item.cover_image" mode="aspectFill" />
          <view class="goods-info">
            <text class="name">{{ item.product_name }}</text>
            <text class="spec">{{ item.sku_name || item.spec_values }}</text>
            <view class="price-row">
              <PriceTag :price="item.price" size="sm" />
              <text class="qty">x{{ item.quantity }}</text>
            </view>
          </view>
        </view>
        <view class="order-footer">
          <text class="total">共 {{ order.items.length }} 件，实付 ¥{{ formatPrice(order.pay_amount) }}</text>
        </view>
      </view>
      <view v-if="orders.length < total" class="load-more" @click="loadMore">
        {{ loading ? '加载中...' : '加载更多' }}
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-text">暂无订单</text>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
}
.tabs {
  display: flex;
  background: #fff;
  padding: 0 8rpx;
  position: sticky;
  top: 0;
  z-index: 10;
}
.tab {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  font-size: 26rpx;
  color: #666;
  position: relative;
}
.tab.active {
  color: #07c160;
  font-weight: 600;
}
.tab.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 4rpx;
  background: #07c160;
  border-radius: 2rpx;
}
.guest, .empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 160rpx;
}
.guest-text, .empty-text {
  color: #999;
  font-size: 28rpx;
}
.login-btn {
  margin-top: 40rpx;
  width: 280rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: #07c160;
  color: #fff;
  font-size: 28rpx;
  border-radius: 40rpx;
  border: none;
}
.list {
  padding: 16rpx;
}
.order-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16rpx;
  border-bottom: 2rpx solid #f5f5f5;
}
.order-no {
  display: block;
  font-size: 24rpx;
  color: #999;
}
.order-time {
  display: block;
  margin-top: 4rpx;
  font-size: 22rpx;
  color: #ccc;
}
.order-meta {
  flex: 1;
  min-width: 0;
}
.status {
  font-size: 26rpx;
  color: #07c160;
  font-weight: 600;
}
.goods-row {
  display: flex;
  padding: 16rpx 0;
}
.cover {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  margin-right: 16rpx;
  background: #f5f5f5;
  flex-shrink: 0;
}
.goods-info {
  flex: 1;
  min-width: 0;
}
.name {
  display: block;
  font-size: 28rpx;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.spec {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.price-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12rpx;
}
.qty {
  font-size: 26rpx;
  color: #666;
}
.order-footer {
  text-align: right;
  padding-top: 16rpx;
  border-top: 2rpx solid #f5f5f5;
}
.total {
  font-size: 26rpx;
  color: #333;
}
.load-more {
  text-align: center;
  padding: 24rpx;
  font-size: 26rpx;
  color: #999;
}
</style>
