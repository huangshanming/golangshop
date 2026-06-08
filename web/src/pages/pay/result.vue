<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const success = ref(false)
const orderNo = ref('')

onLoad((query) => {
  success.value = query?.success === '1'
  orderNo.value = String(query?.order_no || '')
})

function goOrders() {
  uni.redirectTo({ url: '/pages/order/list' })
}

function goHome() {
  uni.switchTab({ url: '/pages/index/index' })
}

function goDetail() {
  if (orderNo.value) {
    uni.redirectTo({ url: `/pages/order/detail?order_no=${orderNo.value}` })
  }
}

function goPay() {
  if (orderNo.value) {
    uni.redirectTo({ url: `/pages/pay/index?order_no=${orderNo.value}` })
  }
}
</script>

<template>
  <view class="page">
    <view class="result-card">
      <text class="icon" :class="{ fail: !success }">{{ success ? '✓' : '✕' }}</text>
      <text class="title">{{ success ? '支付成功' : '支付失败' }}</text>
      <text class="desc">
        {{ success ? '您的订单已支付成功，我们会尽快为您发货' : '支付未完成，请重试或选择其他支付方式' }}
      </text>
      <text v-if="orderNo" class="order-no">订单号：{{ orderNo }}</text>
    </view>

    <view class="actions">
      <button v-if="success" class="primary-btn" @click="goDetail">查看订单</button>
      <button v-else class="primary-btn" @click="goPay">重新支付</button>
      <button class="secondary-btn" @click="goOrders">我的订单</button>
      <button class="secondary-btn" @click="goHome">返回首页</button>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 40rpx 32rpx;
}
.result-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 80rpx 40rpx;
  text-align: center;
}
.icon {
  display: block;
  width: 120rpx;
  height: 120rpx;
  line-height: 120rpx;
  margin: 0 auto;
  font-size: 60rpx;
  font-weight: 700;
  color: #fff;
  background: #07c160;
  border-radius: 50%;
}
.icon.fail {
  background: #ff4d4f;
}
.title {
  display: block;
  margin-top: 32rpx;
  font-size: 40rpx;
  font-weight: 700;
  color: #333;
}
.desc {
  display: block;
  margin-top: 16rpx;
  font-size: 26rpx;
  color: #999;
  line-height: 1.6;
}
.order-no {
  display: block;
  margin-top: 24rpx;
  font-size: 24rpx;
  color: #666;
}
.actions {
  margin-top: 48rpx;
}
.primary-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background: #07c160;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
  border-radius: 44rpx;
  border: none;
  margin-bottom: 24rpx;
}
.secondary-btn {
  width: 100%;
  height: 80rpx;
  line-height: 80rpx;
  background: #fff;
  color: #333;
  font-size: 28rpx;
  border-radius: 40rpx;
  border: 2rpx solid #eee;
  margin-bottom: 16rpx;
}
</style>
