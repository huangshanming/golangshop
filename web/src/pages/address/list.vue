<script setup lang="ts">
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { deleteAddress, fetchAddresses, formatAddress, type Address } from '../../api/address'

const addresses = ref<Address[]>([])
const loading = ref(false)
const selectMode = ref(false)

onLoad((query) => {
  selectMode.value = query?.select === '1'
})

onShow(() => {
  loadAddresses()
})

async function loadAddresses() {
  loading.value = true
  try {
    addresses.value = await fetchAddresses()
  } catch {
    addresses.value = []
  } finally {
    loading.value = false
  }
}

function selectAddress(addr: Address) {
  if (!selectMode.value) return
  uni.setStorageSync('selected_address_id', addr.id)
  uni.navigateBack()
}

function goEdit(id?: number) {
  const url = id ? `/pages/address/edit?id=${id}` : '/pages/address/edit'
  uni.navigateTo({ url })
}

function handleDelete(addr: Address) {
  uni.showModal({
    title: '删除地址',
    content: '确定删除该地址吗？',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteAddress(addr.id)
        uni.showToast({ title: '已删除', icon: 'success' })
        loadAddresses()
      } catch {
        /* handled */
      }
    },
  })
}
</script>

<template>
  <view class="page">
    <view v-if="addresses.length > 0" class="list">
      <view
        v-for="addr in addresses"
        :key="addr.id"
        class="addr-card"
        @click="selectAddress(addr)"
      >
        <view class="addr-main">
          <view class="addr-row">
            <text class="name">{{ addr.name }}</text>
            <text class="phone">{{ addr.phone }}</text>
            <text v-if="addr.is_default" class="default-tag">默认</text>
          </view>
          <text class="detail">{{ formatAddress(addr) }}</text>
        </view>
        <view v-if="!selectMode" class="addr-actions" @click.stop>
          <text class="action" @click="goEdit(addr.id)">编辑</text>
          <text class="action delete" @click="handleDelete(addr)">删除</text>
        </view>
        <text v-else class="arrow">›</text>
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-text">暂无收货地址</text>
    </view>

    <view class="bottom-bar">
      <button class="add-btn" @click="goEdit()">+ 新增地址</button>
    </view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 140rpx;
  box-sizing: border-box;
}

/* #ifdef H5 */
.page {
  padding-bottom: calc(120rpx + var(--goshop-tab-bar-height, 50px));
}
/* #endif */
.list {
  padding: 16rpx;
}
.addr-card {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.addr-main {
  flex: 1;
  min-width: 0;
}
.addr-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
}
.name {
  font-size: 30rpx;
  font-weight: 600;
  color: #333;
}
.phone {
  font-size: 28rpx;
  color: #666;
}
.default-tag {
  font-size: 22rpx;
  color: #07c160;
  background: #e8f8ef;
  padding: 4rpx 12rpx;
  border-radius: 6rpx;
}
.detail {
  display: block;
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #666;
  line-height: 1.5;
}
.addr-actions {
  display: flex;
  gap: 24rpx;
  margin-left: 16rpx;
}
.action {
  font-size: 26rpx;
  color: #07c160;
}
.action.delete {
  color: #ff4d4f;
}
.arrow {
  font-size: 36rpx;
  color: #ccc;
  margin-left: 12rpx;
}
.empty {
  text-align: center;
  padding-top: 200rpx;
}
.empty-text {
  color: #999;
  font-size: 28rpx;
}
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -4rpx 20rpx rgba(0, 0, 0, 0.06);
  box-sizing: border-box;
}

/* #ifdef H5 */
.bottom-bar {
  bottom: var(--goshop-tab-bar-height, 50px);
  padding-bottom: 16rpx;
}
/* #endif */
.add-btn {
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
</style>
