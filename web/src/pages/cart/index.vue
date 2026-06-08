<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { checkAllCart, fetchCart, removeCartItem, updateCartItem, type CartItem } from '../../api/cart'
import PriceTag from '../../components/PriceTag.vue'
import { formatPrice } from '../../utils/format'
import { useUserStore } from '../../stores/user'

const userStore = useUserStore()
const items = ref<CartItem[]>([])
const loading = ref(false)

const checkedItems = computed(() => items.value.filter((i) => i.checked))
const allChecked = computed(() => items.value.length > 0 && items.value.every((i) => i.checked))
const totalAmount = computed(() =>
  checkedItems.value.reduce((sum, item) => sum + item.price * item.quantity, 0),
)

onShow(() => {
  if (userStore.token) {
    loadCart()
  } else {
    items.value = []
  }
})

async function loadCart() {
  loading.value = true
  try {
    items.value = await fetchCart()
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
}

async function toggleItem(item: CartItem) {
  const checked = !item.checked
  try {
    await updateCartItem(item.id, { checked })
    item.checked = checked
  } catch {
    /* handled */
  }
}

async function toggleAll() {
  const checked = !allChecked.value
  try {
    await checkAllCart(checked)
    items.value.forEach((i) => {
      i.checked = checked
    })
  } catch {
    /* handled */
  }
}

async function changeQty(item: CartItem, delta: number) {
  const quantity = item.quantity + delta
  if (quantity < 1) return
  if (quantity > item.stock) {
    uni.showToast({ title: '库存不足', icon: 'none' })
    return
  }
  try {
    await updateCartItem(item.id, { quantity })
    item.quantity = quantity
  } catch {
    /* handled */
  }
}

async function handleRemove(id: number) {
  try {
    await removeCartItem(id)
    items.value = items.value.filter((i) => i.id !== id)
    uni.showToast({ title: '已删除', icon: 'success' })
  } catch {
    /* handled */
  }
}

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index?redirect=' + encodeURIComponent('/pages/cart/index') })
}

function checkout() {
  if (checkedItems.value.length === 0) {
    uni.showToast({ title: '请选择商品', icon: 'none' })
    return
  }
  const payload = {
    items: checkedItems.value.map((i) => ({ sku_id: i.sku_id, quantity: i.quantity })),
    cart_item_ids: checkedItems.value.map((i) => i.id),
  }
  uni.setStorageSync('checkout_payload', payload)
  uni.navigateTo({ url: '/pages/checkout/index' })
}
</script>

<template>
  <view class="page">
    <view v-if="!userStore.token" class="guest">
      <text class="guest-icon">🛒</text>
      <text class="guest-text">登录后查看购物车</text>
      <button class="login-btn" @click="goLogin">去登录</button>
    </view>

    <view v-else-if="items.length > 0" class="list">
      <view v-for="item in items" :key="item.id" class="cart-item">
        <view class="check" :class="{ checked: item.checked }" @click="toggleItem(item)">
          <text v-if="item.checked" class="check-icon">✓</text>
        </view>
        <image class="cover" :src="item.cover_image" mode="aspectFill" />
        <view class="item-info">
          <text class="name">{{ item.product_name }}</text>
          <text class="spec">{{ item.spec }}</text>
          <view class="price-row">
            <PriceTag :price="item.price" size="sm" />
            <view class="qty-ctrl">
              <text class="qty-btn" :class="{ disabled: item.quantity <= 1 }" @click="changeQty(item, -1)">−</text>
              <text class="qty-num">{{ item.quantity }}</text>
              <text class="qty-btn" :class="{ disabled: item.quantity >= item.stock }" @click="changeQty(item, 1)">+</text>
            </view>
          </view>
        </view>
        <text class="remove" @click="handleRemove(item.id)">删除</text>
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-icon">🛒</text>
      <text class="empty-text">购物车是空的</text>
    </view>

    <view v-if="userStore.token && items.length > 0" class="bottom-bar">
      <view class="check-all" @click="toggleAll">
        <view class="check" :class="{ checked: allChecked }">
          <text v-if="allChecked" class="check-icon">✓</text>
        </view>
        <text class="check-label">全选</text>
      </view>
      <view class="total">
        合计：<text class="total-price">¥{{ formatPrice(totalAmount) }}</text>
      </view>
      <button class="checkout-btn" @click="checkout">去结算({{ checkedItems.length }})</button>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
}
.guest, .empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 160rpx;
}
.guest-icon, .empty-icon {
  font-size: 80rpx;
}
.guest-text, .empty-text {
  margin-top: 24rpx;
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
.cart-item {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.check {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  border: 2rpx solid #ccc;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: 16rpx;
}
.check.checked {
  background: #07c160;
  border-color: #07c160;
}
.check-icon {
  color: #fff;
  font-size: 24rpx;
  font-weight: 700;
}
.cover {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  flex-shrink: 0;
  margin-right: 20rpx;
  background: #f5f5f5;
}
.item-info {
  flex: 1;
  min-width: 0;
}
.name {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
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
  margin-top: 16rpx;
}
.qty-ctrl {
  display: flex;
  align-items: center;
  border: 2rpx solid #eee;
  border-radius: 8rpx;
}
.qty-btn {
  width: 52rpx;
  height: 52rpx;
  line-height: 52rpx;
  text-align: center;
  font-size: 32rpx;
  color: #333;
}
.qty-btn.disabled {
  color: #ccc;
}
.qty-num {
  min-width: 56rpx;
  text-align: center;
  font-size: 26rpx;
  color: #333;
  border-left: 2rpx solid #eee;
  border-right: 2rpx solid #eee;
}
.remove {
  font-size: 26rpx;
  color: #ff4d4f;
  padding: 16rpx 0 16rpx 12rpx;
  flex-shrink: 0;
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
.check-all {
  display: flex;
  align-items: center;
  margin-right: 16rpx;
}
.check-label {
  font-size: 26rpx;
  color: #333;
  margin-left: 8rpx;
}
.total {
  flex: 1;
  font-size: 26rpx;
  color: #333;
  text-align: right;
  margin-right: 16rpx;
}
.total-price {
  color: #ff4d4f;
  font-weight: 700;
  font-size: 34rpx;
}
.checkout-btn {
  width: 220rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: #07c160;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
  border-radius: 40rpx;
  border: none;
  padding: 0;
}
</style>
