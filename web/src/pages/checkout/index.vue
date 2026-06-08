<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { fetchAvailableCoupons, type CouponItem } from '../../api/coupon'
import {
  createOrder,
  genClientOrderNo,
  previewOrder,
  type CheckoutItem,
  type PreviewResult,
} from '../../api/order'
import PriceTag from '../../components/PriceTag.vue'
import { formatPrice } from '../../utils/format'

interface CheckoutPayload {
  items: CheckoutItem[]
  cart_item_ids?: number[]
}

const checkoutItems = ref<CheckoutItem[]>([])
const cartItemIds = ref<number[]>([])
const preview = ref<PreviewResult | null>(null)
const addressId = ref(0)
const addressText = ref('')
const userCouponId = ref(0)
const payChannel = ref<'balance' | 'wechat' | 'alipay'>('wechat')
const remark = ref('')
const loading = ref(false)
const submitting = ref(false)
const showCouponPicker = ref(false)
const coupons = ref<CouponItem[]>([])

const selectedCoupon = computed(() =>
  coupons.value.find((c) => c.user_coupon_id === userCouponId.value),
)

onLoad(() => {
  const payload = uni.getStorageSync('checkout_payload') as CheckoutPayload | null
  if (!payload?.items?.length) {
    uni.showToast({ title: '结算数据无效', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  checkoutItems.value = payload.items
  cartItemIds.value = payload.cart_item_ids || []
})

onShow(() => {
  const selectedId = uni.getStorageSync('selected_address_id') as number
  if (selectedId) {
    addressId.value = selectedId
    uni.removeStorageSync('selected_address_id')
  }
  if (checkoutItems.value.length) {
    loadPreview()
  }
})

async function loadPreview() {
  loading.value = true
  try {
    const data = await previewOrder({
      items: checkoutItems.value,
      address_id: addressId.value || undefined,
      user_coupon_id: userCouponId.value || undefined,
      cart_item_ids: cartItemIds.value.length ? cartItemIds.value : undefined,
    })
    preview.value = data
    if (data.address && !addressId.value) {
      addressId.value = data.address.id
      addressText.value = `${data.address.name} ${data.address.phone}\n${data.address.full_text}`
    } else if (addressId.value && data.address) {
      addressText.value = `${data.address.name} ${data.address.phone}\n${data.address.full_text}`
    }
    if (!userCouponId.value && data.user_coupon_id) {
      userCouponId.value = data.user_coupon_id
    }
    await loadCoupons(data.goods_amount)
  } catch {
    preview.value = null
  } finally {
    loading.value = false
  }
}

async function loadCoupons(goodsAmount: number) {
  try {
    coupons.value = await fetchAvailableCoupons(goodsAmount)
  } catch {
    coupons.value = []
  }
}

function goAddressList() {
  uni.navigateTo({ url: '/pages/address/list?select=1' })
}

function openCouponPicker() {
  showCouponPicker.value = true
}

function selectCoupon(coupon: CouponItem | null) {
  if (coupon && !coupon.can_use) return
  userCouponId.value = coupon?.user_coupon_id || 0
  showCouponPicker.value = false
  loadPreview()
}

async function handleSubmit() {
  if (!addressId.value) {
    uni.showToast({ title: '请选择收货地址', icon: 'none' })
    return
  }
  if (!preview.value?.preview_token) {
    uni.showToast({ title: '订单预览失败', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    const order = await createOrder({
      client_order_no: genClientOrderNo(),
      address_id: addressId.value,
      user_coupon_id: userCouponId.value || undefined,
      preview_token: preview.value.preview_token,
      pay_channel: payChannel.value,
      remark: remark.value || undefined,
      cart_item_ids: cartItemIds.value.length ? cartItemIds.value : undefined,
      items: checkoutItems.value,
    })
    uni.removeStorageSync('checkout_payload')
    if (payChannel.value === 'balance' || order.status === 1) {
      uni.redirectTo({
        url: `/pages/pay/result?success=1&order_no=${order.order_no}`,
      })
    } else {
      uni.redirectTo({
        url: `/pages/pay/index?order_no=${order.order_no}&amount=${order.pay_amount}`,
      })
    }
  } catch {
    /* handled */
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <view class="page">
    <view v-if="loading && !preview" class="loading">加载中...</view>

    <template v-else-if="preview">
      <view class="card address-card" @click="goAddressList">
        <view v-if="addressId" class="address-info">
          <view class="addr-row">
            <text class="addr-name">{{ preview.address?.name }}</text>
            <text class="addr-phone">{{ preview.address?.phone }}</text>
          </view>
          <text class="addr-detail">{{ preview.address?.full_text || addressText }}</text>
        </view>
        <view v-else class="addr-empty">
          <text class="addr-empty-text">+ 添加收货地址</text>
        </view>
        <text class="arrow">›</text>
      </view>

      <view class="card">
        <view v-for="item in preview.items" :key="item.sku_id" class="goods-item">
          <image class="goods-cover" :src="item.cover_image" mode="aspectFill" />
          <view class="goods-info">
            <text class="goods-name">{{ item.product_name }}</text>
            <text class="goods-spec">{{ item.spec }}</text>
            <view class="goods-bottom">
              <PriceTag :price="item.price" size="sm" />
              <text class="goods-qty">x{{ item.quantity }}</text>
            </view>
            <text v-if="!item.stock_enough" class="stock-warn">库存不足</text>
          </view>
        </view>
      </view>

      <view class="card row-item" @click="openCouponPicker">
        <text class="row-label">优惠券</text>
        <view class="row-value">
          <text v-if="selectedCoupon || preview.coupon_name" class="coupon-text">
            {{ selectedCoupon?.name || preview.coupon_name || '已选优惠券' }}
            <text v-if="preview.discount_amount > 0" class="discount">-¥{{ formatPrice(preview.discount_amount) }}</text>
          </text>
          <text v-else class="placeholder">{{ coupons.filter(c => c.can_use).length }}张可用</text>
          <text class="arrow">›</text>
        </view>
      </view>

      <view class="card">
        <view class="summary-row">
          <text>商品金额</text>
          <text>¥{{ formatPrice(preview.goods_amount) }}</text>
        </view>
        <view class="summary-row">
          <text>优惠</text>
          <text class="discount">-¥{{ formatPrice(preview.discount_amount) }}</text>
        </view>
        <view class="summary-row">
          <text>运费</text>
          <text>{{ preview.freight_amount > 0 ? `¥${formatPrice(preview.freight_amount)}` : '免运费' }}</text>
        </view>
        <view class="summary-row total-row">
          <text>应付金额</text>
          <text class="pay-amount">¥{{ formatPrice(preview.pay_amount) }}</text>
        </view>
      </view>

      <view class="card">
        <view class="section-title">支付方式</view>
        <radio-group @change="(e: any) => payChannel = e.detail.value">
          <label class="pay-option">
            <radio value="balance" :checked="payChannel === 'balance'" color="#07c160" />
            <text>余额支付</text>
            <text class="balance-tip">（余额 ¥{{ formatPrice(preview.balance) }}）</text>
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

      <view class="card">
        <input v-model="remark" class="remark-input" placeholder="订单备注（选填）" />
      </view>
    </template>

    <view v-if="preview" class="bottom-bar">
      <view class="bar-total">
        合计：<text class="bar-price">¥{{ formatPrice(preview.pay_amount) }}</text>
      </view>
      <button class="submit-btn" :disabled="submitting" @click="handleSubmit">
        {{ submitting ? '提交中...' : '提交订单' }}
      </button>
    </view>

    <view v-if="showCouponPicker" class="mask" @click="showCouponPicker = false">
      <view class="coupon-panel" @click.stop>
        <view class="panel-header">
          <text class="panel-title">选择优惠券</text>
          <text class="panel-close" @click="showCouponPicker = false">×</text>
        </view>
        <scroll-view scroll-y class="coupon-list">
          <view class="coupon-item none" @click="selectCoupon(null)">
            <text>不使用优惠券</text>
            <text v-if="!userCouponId" class="selected">✓</text>
          </view>
          <view
            v-for="c in coupons"
            :key="c.user_coupon_id"
            class="coupon-item"
            :class="{ disabled: !c.can_use, active: userCouponId === c.user_coupon_id }"
            @click="selectCoupon(c)"
          >
            <view class="coupon-main">
              <text class="coupon-name">{{ c.name }}</text>
              <text class="coupon-desc">
                {{ c.can_use ? `可减 ¥${formatPrice(c.discount_amount)}` : c.reason }}
              </text>
            </view>
            <text v-if="userCouponId === c.user_coupon_id" class="selected">✓</text>
          </view>
          <view v-if="coupons.length === 0" class="coupon-empty">暂无可用优惠券</view>
        </scroll-view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}
.loading {
  text-align: center;
  padding-top: 200rpx;
  color: #999;
  font-size: 28rpx;
}
.card {
  background: #fff;
  margin: 16rpx;
  border-radius: 16rpx;
  padding: 24rpx;
}
.address-card {
  display: flex;
  align-items: center;
}
.address-info {
  flex: 1;
}
.addr-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.addr-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}
.addr-phone {
  font-size: 28rpx;
  color: #666;
}
.addr-detail {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #666;
  line-height: 1.5;
}
.addr-empty-text {
  font-size: 28rpx;
  color: #07c160;
}
.arrow {
  font-size: 36rpx;
  color: #ccc;
  margin-left: 12rpx;
}
.goods-item {
  display: flex;
  padding: 16rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}
.goods-item:last-child {
  border-bottom: none;
}
.goods-cover {
  width: 140rpx;
  height: 140rpx;
  border-radius: 12rpx;
  margin-right: 20rpx;
  background: #f5f5f5;
  flex-shrink: 0;
}
.goods-info {
  flex: 1;
  min-width: 0;
}
.goods-name {
  display: block;
  font-size: 28rpx;
  color: #333;
  font-weight: 600;
}
.goods-spec {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.goods-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12rpx;
}
.goods-qty {
  font-size: 26rpx;
  color: #666;
}
.stock-warn {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #ff4d4f;
}
.row-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.row-label {
  font-size: 28rpx;
  color: #333;
}
.row-value {
  display: flex;
  align-items: center;
  font-size: 26rpx;
  color: #666;
}
.coupon-text {
  color: #ff4d4f;
}
.placeholder {
  color: #999;
}
.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 12rpx 0;
  font-size: 28rpx;
  color: #666;
}
.discount {
  color: #ff4d4f;
}
.total-row {
  border-top: 2rpx solid #f5f5f5;
  margin-top: 8rpx;
  padding-top: 20rpx;
  font-weight: 600;
  color: #333;
}
.pay-amount {
  color: #ff4d4f;
  font-size: 32rpx;
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
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #333;
}
.balance-tip {
  margin-left: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.remark-input {
  width: 100%;
  height: 72rpx;
  font-size: 28rpx;
}
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
}
.bar-total {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}
.bar-price {
  color: #ff4d4f;
  font-weight: 700;
  font-size: 36rpx;
}
.submit-btn {
  width: 260rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: #07c160;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
  border-radius: 40rpx;
  border: none;
}
.mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 100;
  display: flex;
  align-items: flex-end;
}
.coupon-panel {
  width: 100%;
  max-height: 70vh;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx 32rpx;
  border-bottom: 2rpx solid #f5f5f5;
}
.panel-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.panel-close {
  font-size: 44rpx;
  color: #999;
  line-height: 1;
}
.coupon-list {
  max-height: 60vh;
  padding: 16rpx;
}
.coupon-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx;
  margin-bottom: 12rpx;
  background: #fff5f5;
  border: 2rpx solid #ffe0e0;
  border-radius: 12rpx;
}
.coupon-item.none {
  background: #f5f5f5;
  border-color: #eee;
}
.coupon-item.disabled {
  opacity: 0.5;
}
.coupon-item.active {
  border-color: #07c160;
  background: #e8f8ef;
}
.coupon-name {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
}
.coupon-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.selected {
  color: #07c160;
  font-size: 32rpx;
  font-weight: 700;
}
.coupon-empty {
  text-align: center;
  padding: 60rpx;
  color: #999;
  font-size: 28rpx;
}
</style>
