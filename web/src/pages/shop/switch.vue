<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { useTenantStore } from '../../stores/tenant'
import { useUserStore } from '../../stores/user'
import type { Tenant } from '../../api/tenant'

const tenantStore = useTenantStore()
const userStore = useUserStore()

onShow(async () => {
  if (!tenantStore.loaded) {
    try {
      await tenantStore.loadTenants()
    } catch {
      /* ignore */
    }
  }
})

function selectTenant(t: Tenant) {
  if (t.id === tenantStore.tenantId) {
    uni.navigateBack()
    return
  }
  userStore.logout()
  tenantStore.setTenant(t)
  uni.showToast({ title: `已进入 ${t.name}`, icon: 'success' })
  setTimeout(() => {
    uni.switchTab({ url: '/pages/index/index' })
  }, 400)
}
</script>

<template>
  <view class="page">
    <view class="tip">选择要进入的店铺，不同店铺的商品与订单相互独立</view>
    <view
      v-for="t in tenantStore.tenants"
      :key="t.id"
      class="tenant-item"
      :class="{ active: t.id === tenantStore.tenantId }"
      @click="selectTenant(t)"
    >
      <view class="tenant-main">
        <text class="tenant-name">{{ t.name }}</text>
        <text class="tenant-code">编码：{{ t.code }}</text>
      </view>
      <text v-if="t.id === tenantStore.tenantId" class="check">当前</text>
      <text v-else class="arrow">›</text>
    </view>
    <view v-if="tenantStore.tenants.length === 0" class="empty">暂无可用店铺</view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
}
.tip {
  font-size: 24rpx;
  color: #999;
  margin-bottom: 24rpx;
  line-height: 1.6;
}
.tenant-item {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx 28rpx;
  margin-bottom: 16rpx;
  border: 2rpx solid transparent;
}
.tenant-item.active {
  border-color: #07c160;
  background: #f6ffed;
}
.tenant-main {
  flex: 1;
}
.tenant-name {
  display: block;
  font-size: 32rpx;
  font-weight: 600;
  color: #333;
}
.tenant-code {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #999;
}
.check {
  font-size: 24rpx;
  color: #07c160;
  font-weight: 600;
}
.arrow {
  font-size: 36rpx;
  color: #ccc;
}
.empty {
  text-align: center;
  padding: 120rpx 0;
  color: #999;
}
</style>
