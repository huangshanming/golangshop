<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { mockPayOrder, payOrder } from '../../api/order'
import { formatPrice } from '../../utils/format'

const orderNo = ref('')
const amount = ref(0)
const payChannel = ref<'balance' | 'wechat' | 'alipay'>('wechat')
const paying = ref(false)

onLoad((query) => {
  orderNo.value = String(query?.order_no || '')
  amount.value = Number(query?.amount || 0)
  if (!orderNo.value) {
    uni.showToast({ title: '订单号无效', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
  }
})

async function handlePay() {
  if (!orderNo.value) return
  paying.value = true
  try {
    const res = await payOrder(orderNo.value, payChannel.value)
    if (res.paid || payChannel.value === 'balance') {
      goResult(true)
    } else {
      uni.showToast({ title: '请在第三方完成支付', icon: 'none' })
    }
  } catch {
    goResult(false)
  } finally {
    paying.value = false
  }
}

async function handleMockPay() {
  if (!orderNo.value) return
  paying.value = true
  try {
    const res = await mockPayOrder(orderNo.value)
    goResult(res.paid)
  } catch {
    goResult(false)
  } finally {
    paying.value = false
  }
}

function goResult(success: boolean) {
  uni.redirectTo({
    url: `/pages/pay/result?success=${success ? 1 : 0}&order_no=${orderNo.value}`,
  })
}
</script>

<template>
  <view class="page">
    <view class="amount-card">
      <text class="amount-label">支付金额</text>
      <text class="amount-value">¥{{ formatPrice(amount) }}</text>
      <text class="order-no">订单号：{{ orderNo }}</text>
    </view>

    <view class="card">
      <view class="section-title">选择支付方式</view>
      <radio-group @change="(e: any) => payChannel = e.detail.value">
        <label class="pay-option">
          <radio value="balance" :checked="payChannel === 'balance'" color="#07c160" />
          <text>余额支付</text>
        </label>
        <label class="pay-option">
          <radio value="wechat" :checked="payChannel === 'wechat'" color="#07c160" />
          <text>微信支付</text>
        </label>
        <label class="pay-option">
          <radio value="alipay" :checked="payChannel === 'alipay'" color="#07c160" />
          <text>支付宝</text>
        </label>
      </radio-group>
    </view>

    <view class="actions">
      <button class="pay-btn" :disabled="paying" @click="handlePay">
        {{ paying ? '支付中...' : '确认支付' }}
      </button>
      <button class="mock-btn" :disabled="paying" @click="handleMockPay">模拟支付（开发）</button>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16rpx;
}
.amount-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 48rpx 32rpx;
  text-align: center;
  margin-bottom: 16rpx;
}
.amount-label {
  display: block;
  font-size: 28rpx;
  color: #999;
}
.amount-value {
  display: block;
  margin-top: 16rpx;
  font-size: 56rpx;
  font-weight: 700;
  color: #ff4d4f;
}
.order-no {
  display: block;
  margin-top: 20rpx;
  font-size: 24rpx;
  color: #999;
}
.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.section-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 16rpx;
}
.pay-option {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  font-size: 28rpx;
  color: #333;
  border-bottom: 2rpx solid #f5f5f5;
}
.pay-option:last-child {
  border-bottom: none;
}
.actions {
  padding: 40rpx 16rpx;
}
.pay-btn {
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
.mock-btn {
  margin-top: 24rpx;
  width: 100%;
  height: 80rpx;
  line-height: 80rpx;
  background: #fff;
  color: #666;
  font-size: 28rpx;
  border-radius: 40rpx;
  border: 2rpx solid #ddd;
}
</style>
