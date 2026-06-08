<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import {
  cancelOrder,
  confirmReceive,
  createAfterSale,
  expressCompanyName,
  fetchOrderDetail,
  reviewOrder,
  type Order,
} from '../../api/order'
import PriceTag from '../../components/PriceTag.vue'
import { formatDateTime, formatPrice } from '../../utils/format'

const orderNo = ref('')
const order = ref<Order | null>(null)
const loading = ref(false)

interface AddressSnap {
  name?: string
  phone?: string
  province?: string
  city?: string
  district?: string
  detail?: string
  full_text?: string
}

const addressSnap = computed<AddressSnap | null>(() => {
  if (!order.value?.address_snapshot) return null
  try {
    return JSON.parse(order.value.address_snapshot) as AddressSnap
  } catch {
    return null
  }
})

const addressText = computed(() => {
  const a = addressSnap.value
  if (!a) return order.value?.address_snapshot || ''
  if (a.full_text) return a.full_text
  return `${a.province || ''}${a.city || ''}${a.district || ''}${a.detail || ''}`
})

const showPay = computed(() => order.value?.status === 0)
const showCancel = computed(() => order.value?.status === 0)
const showConfirm = computed(() => order.value?.status === 2)
const showReview = computed(() => order.value?.status === 3 && order.value?.review_status === 1)
const showAfterSale = computed(() => order.value && [1, 2, 3].includes(order.value.status))
const showLogistics = computed(() => order.value && order.value.status >= 2 && !!order.value.express_no)

onLoad((query) => {
  orderNo.value = String(query?.order_no || '')
})

onShow(() => {
  if (orderNo.value) loadDetail()
})

async function loadDetail() {
  loading.value = true
  try {
    order.value = await fetchOrderDetail(orderNo.value)
  } catch {
    order.value = null
  } finally {
    loading.value = false
  }
}

function goPay() {
  if (!order.value) return
  uni.navigateTo({
    url: `/pages/pay/index?order_no=${order.value.order_no}&amount=${order.value.pay_amount}`,
  })
}

async function handleCancel() {
  uni.showModal({
    title: '取消订单',
    content: '确定要取消该订单吗？',
    success: async (res) => {
      if (!res.confirm || !order.value) return
      try {
        await cancelOrder(order.value.order_no)
        uni.showToast({ title: '已取消', icon: 'success' })
        loadDetail()
      } catch {
        /* handled */
      }
    },
  })
}

async function handleConfirm() {
  if (!order.value) return
  try {
    await confirmReceive(order.value.order_no)
    uni.showToast({ title: '已确认收货', icon: 'success' })
    loadDetail()
  } catch {
    /* handled */
  }
}

async function handleReview() {
  if (!order.value) return
  try {
    await reviewOrder(order.value.order_no)
    uni.showToast({ title: '评价成功', icon: 'success' })
    loadDetail()
  } catch {
    /* handled */
  }
}

function handleAfterSale() {
  if (!order.value) return
  uni.showActionSheet({
    itemList: ['仅退款', '退货退款', '换货'],
    success: async (res) => {
      const types = ['refund_only', 'return_refund', 'exchange'] as const
      const type = types[res.tapIndex]
      try {
        await createAfterSale({
          order_no: order.value!.order_no,
          type,
          reason: '用户申请售后',
        })
        uni.showToast({ title: '售后申请已提交', icon: 'success' })
        loadDetail()
      } catch {
        /* handled */
      }
    },
  })
}

function copyExpressNo() {
  if (!order.value?.express_no) return
  uni.setClipboardData({
    data: order.value.express_no,
    success: () => uni.showToast({ title: '单号已复制', icon: 'success' }),
  })
}
</script>

<template>
  <view class="page">
    <view v-if="loading && !order" class="loading">加载中...</view>

    <template v-else-if="order">
      <view class="status-card">
        <text class="status-text">{{ order.display_status }}</text>
        <text class="order-no">订单号：{{ order.order_no }}</text>
      </view>

      <view v-if="addressSnap || order.address_snapshot" class="card">
        <view class="card-title">收货信息</view>
        <view v-if="addressSnap" class="addr-row">
          <text class="addr-name">{{ addressSnap.name }}</text>
          <text class="addr-phone">{{ addressSnap.phone }}</text>
        </view>
        <text class="addr-detail">{{ addressText }}</text>
      </view>

      <view v-if="showLogistics" class="card">
        <view class="card-title">物流信息</view>
        <view class="summary-row">
          <text>物流公司</text>
          <text>{{ expressCompanyName(order.express_company) }}</text>
        </view>
        <view class="summary-row express-row">
          <text>运单号</text>
          <view class="express-no-wrap">
            <text class="express-no">{{ order.express_no }}</text>
            <text class="copy-btn" @click="copyExpressNo">复制</text>
          </view>
        </view>
        <view v-if="order.shipped_at" class="summary-row">
          <text>发货时间</text>
          <text>{{ formatDateTime(order.shipped_at) }}</text>
        </view>
        <view v-if="order.ship_remark" class="summary-row">
          <text>发货备注</text>
          <text class="remark">{{ order.ship_remark }}</text>
        </view>
      </view>

      <view class="card">
        <view class="card-title">商品信息</view>
        <view v-for="item in order.items" :key="item.id" class="goods-item">
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
      </view>

      <view class="card">
        <view class="card-title">价格明细</view>
        <view class="summary-row">
          <text>商品金额</text>
          <text>¥{{ formatPrice(order.goods_amount) }}</text>
        </view>
        <view class="summary-row">
          <text>优惠</text>
          <text class="discount">-¥{{ formatPrice(order.discount_amount) }}</text>
        </view>
        <view class="summary-row">
          <text>运费</text>
          <text>{{ order.freight_amount > 0 ? `¥${formatPrice(order.freight_amount)}` : '免运费' }}</text>
        </view>
        <view class="summary-row total-row">
          <text>实付金额</text>
          <text class="pay-amount">¥{{ formatPrice(order.pay_amount) }}</text>
        </view>
      </view>

      <view v-if="order.remark" class="card">
        <view class="card-title">备注</view>
        <text class="remark">{{ order.remark }}</text>
      </view>

      <view class="card">
        <view class="summary-row">
          <text>下单时间</text>
          <text>{{ formatDateTime(order.created_at) }}</text>
        </view>
        <view v-if="order.pay_channel" class="summary-row">
          <text>支付方式</text>
          <text>{{ order.pay_channel === 'balance' ? '余额' : order.pay_channel === 'wechat' ? '微信' : '支付宝' }}</text>
        </view>
      </view>
    </template>

    <view v-else class="empty">订单不存在</view>

    <view v-if="order && (showPay || showCancel || showConfirm || showReview || showAfterSale)" class="bottom-bar">
      <button v-if="showAfterSale" class="action-btn outline" @click="handleAfterSale">申请售后</button>
      <button v-if="showCancel" class="action-btn outline" @click="handleCancel">取消订单</button>
      <button v-if="showConfirm" class="action-btn primary" @click="handleConfirm">确认收货</button>
      <button v-if="showReview" class="action-btn primary" @click="handleReview">去评价</button>
      <button v-if="showPay" class="action-btn primary" @click="goPay">去支付</button>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}
.loading, .empty {
  text-align: center;
  padding-top: 200rpx;
  color: #999;
  font-size: 28rpx;
}
.status-card {
  background: linear-gradient(135deg, #07c160, #06ae56);
  padding: 48rpx 32rpx;
  color: #fff;
}
.status-text {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
}
.order-no {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  opacity: 0.85;
}
.card {
  background: #fff;
  margin: 16rpx;
  border-radius: 16rpx;
  padding: 24rpx;
}
.card-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #333;
  margin-bottom: 16rpx;
}
.addr-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 8rpx;
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
  font-size: 26rpx;
  color: #666;
  line-height: 1.5;
}
.goods-item {
  display: flex;
  padding: 16rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}
.goods-item:last-child {
  border-bottom: none;
}
.cover {
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
.name {
  display: block;
  font-size: 28rpx;
  color: #333;
  font-weight: 600;
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
.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 10rpx 0;
  font-size: 28rpx;
  color: #666;
}
.discount {
  color: #ff4d4f;
}
.total-row {
  border-top: 2rpx solid #f5f5f5;
  margin-top: 8rpx;
  padding-top: 16rpx;
  font-weight: 600;
  color: #333;
}
.pay-amount {
  color: #ff4d4f;
  font-size: 32rpx;
}
.remark {
  font-size: 28rpx;
  color: #666;
}
.express-row {
  align-items: center;
}
.express-no-wrap {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.express-no {
  font-size: 28rpx;
  color: #333;
  font-weight: 600;
}
.copy-btn {
  font-size: 24rpx;
  color: #07c160;
  padding: 4rpx 16rpx;
  border: 2rpx solid #07c160;
  border-radius: 20rpx;
}
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
}
.action-btn {
  min-width: 180rpx;
  height: 72rpx;
  line-height: 72rpx;
  font-size: 28rpx;
  border-radius: 36rpx;
  margin: 0;
  padding: 0 24rpx;
}
.action-btn.primary {
  background: #07c160;
  color: #fff;
  border: none;
}
.action-btn.outline {
  background: #fff;
  color: #333;
  border: 2rpx solid #ddd;
}
</style>
