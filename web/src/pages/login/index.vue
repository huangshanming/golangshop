<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../stores/user'
import { useTenantStore } from '../../stores/tenant'
import { formatPrice } from '../../utils/format'
import { syncTabBarSelected } from '../../utils/tabbar'
import PageTabBar from '../../components/PageTabBar.vue'

const userStore = useUserStore()
const tenantStore = useTenantStore()
const phone = ref('')
const redirect = ref('')

onLoad((query) => {
  redirect.value = query?.redirect ? decodeURIComponent(String(query.redirect)) : ''
  const tenantId = Number(query?.tenant_id || 0)
  if (tenantId) {
    tenantStore.setTenantById(tenantId)
  }
})

onShow(async () => {
  syncTabBarSelected()
  if (userStore.token) {
    await userStore.refreshProfile()
  }
})

async function handleLogin() {
  if (!phone.value.trim()) {
    uni.showToast({ title: '请输入手机号', icon: 'none' })
    return
  }
  try {
    await userStore.login(phone.value.trim())
    uni.showToast({ title: '登录成功', icon: 'success' })
    if (redirect.value) {
      setTimeout(() => {
        const tabPages = ['/pages/index/index', '/pages/category/index', '/pages/login/index']
        if (tabPages.some((p) => redirect.value.startsWith(p.split('?')[0]))) {
          uni.switchTab({ url: redirect.value.split('?')[0] })
        } else {
          uni.redirectTo({ url: redirect.value })
        }
      }, 500)
    }
  } catch {
    /* handled */
  }
}

function handleLogout() {
  userStore.logout()
  uni.showToast({ title: '已退出', icon: 'none' })
}

function goPage(url: string) {
  if (!userStore.token) {
    uni.navigateTo({ url: `/pages/login/index?redirect=${encodeURIComponent(url)}` })
    return
  }
  uni.navigateTo({ url })
}
</script>

<template>
  <view class="page">
    <view v-if="userStore.token" class="profile">
      <view class="avatar">👤</view>
      <view class="info">
        <text class="nickname">{{ userStore.user?.nickname || '用户' }}</text>
        <text class="phone">{{ userStore.user?.phone || '' }}</text>
      </view>
      <view class="balance-card" @click="goPage('/pages/user/balance-logs')">
        <text class="balance-label">账户余额</text>
        <text class="balance-value">¥{{ formatPrice(userStore.user?.balance ?? 0) }}</text>
        <text class="balance-link">查看明细 ›</text>
      </view>
      <button class="logout-btn" @click="handleLogout">退出登录</button>
    </view>

    <view v-if="userStore.token" class="menu">
      <view class="menu-item" @click="goPage('/pages/user/profile')">
        <text class="menu-icon">👤</text>
        <text class="menu-text">个人信息</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @click="goPage('/pages/user/balance-logs')">
        <text class="menu-icon">💰</text>
        <text class="menu-text">余额明细</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @click="goPage('/pages/cart/index')">
        <text class="menu-icon">🛒</text>
        <text class="menu-text">购物车</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @click="goPage('/pages/order/list')">
        <text class="menu-icon">📦</text>
        <text class="menu-text">我的订单</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @click="goPage('/pages/address/list')">
        <text class="menu-icon">📍</text>
        <text class="menu-text">收货地址</text>
        <text class="menu-arrow">›</text>
      </view>
    </view>

    <view v-else class="login-form">
      <view class="welcome">
        <text class="welcome-title">欢迎来到搜店网</text>
        <text class="welcome-desc">登录后可加购、查看购物车</text>
      </view>
      <view class="form-card">
        <text class="label">手机号</text>
        <input v-model="phone" class="input" type="number" maxlength="11" placeholder="请输入手机号" />
        <button class="login-btn" @click="handleLogin">登录</button>
      </view>
    </view>
  </view>
  <PageTabBar />
</template>

<style scoped>
.page {
  background: #f5f5f5;
  box-sizing: border-box;
}

/* #ifdef H5 */
.page {
  padding-bottom: calc(50px + env(safe-area-inset-bottom, 0px));
}
/* #endif */

.menu {
  margin: 16rpx;
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.profile {
  background: linear-gradient(135deg, #07c160, #06ae56);
  padding: 48rpx 32rpx 56rpx;
  text-align: center;
  box-sizing: border-box;
}
.avatar {
  width: 120rpx;
  height: 120rpx;
  line-height: 120rpx;
  text-align: center;
  font-size: 60rpx;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  margin: 0 auto;
}
.info {
  margin-top: 24rpx;
}
.nickname {
  display: block;
  font-size: 36rpx;
  font-weight: 600;
  color: #fff;
}
.phone {
  display: block;
  margin-top: 8rpx;
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.85);
}
.balance-card {
  margin-top: 28rpx;
  padding: 20rpx 24rpx;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 16rpx;
  box-sizing: border-box;
  text-align: left;
}
.balance-label {
  display: block;
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.85);
}
.balance-value {
  display: block;
  margin-top: 8rpx;
  font-size: 40rpx;
  font-weight: 700;
  color: #fff;
}
.balance-link {
  display: block;
  margin-top: 10rpx;
  text-align: right;
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.75);
}
.logout-btn {
  margin: 40rpx auto 0;
  width: 280rpx;
  height: 72rpx;
  line-height: 72rpx;
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
  font-size: 28rpx;
  border-radius: 36rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.6);
  display: block;
}
.login-form {
  padding: 40rpx;
}
.welcome {
  padding: 60rpx 0 40rpx;
  text-align: center;
}
.welcome-title {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
  color: #333;
}
.welcome-desc {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #999;
}
.form-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 40rpx;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.06);
}
.label {
  display: block;
  font-size: 28rpx;
  color: #666;
  margin-bottom: 16rpx;
}
.input {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  background: #f5f5f5;
  border-radius: 12rpx;
  font-size: 30rpx;
  box-sizing: border-box;
}
.login-btn {
  margin-top: 40rpx;
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
.menu-item {
  display: flex;
  align-items: center;
  padding: 32rpx 24rpx;
  border-bottom: 2rpx solid #f5f5f5;
}
.menu-item:last-child {
  border-bottom: none;
}
.menu-icon {
  font-size: 36rpx;
  margin-right: 16rpx;
}
.menu-text {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}
.menu-arrow {
  font-size: 32rpx;
  color: #ccc;
}
</style>
