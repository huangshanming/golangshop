<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { updateProfile } from '../../api/user'
import { useUserStore } from '../../stores/user'

const userStore = useUserStore()
const nickname = ref('')
const saving = ref(false)

onShow(async () => {
  await userStore.refreshProfile()
  nickname.value = userStore.user?.nickname || ''
})

async function handleSave() {
  const name = nickname.value.trim()
  if (!name) {
    uni.showToast({ title: '请输入昵称', icon: 'none' })
    return
  }
  saving.value = true
  try {
    await updateProfile({ nickname: name })
    await userStore.refreshProfile()
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 500)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <view class="page">
    <view class="card">
      <view class="row">
        <text class="label">手机号</text>
        <text class="value">{{ userStore.user?.phone || '-' }}</text>
      </view>
      <view class="row">
        <text class="label">昵称</text>
        <input v-model="nickname" class="input" placeholder="请输入昵称" maxlength="20" />
      </view>
    </view>
    <button class="save-btn" :loading="saving" @click="handleSave">保存</button>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 24rpx;
}
.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 0 24rpx;
}
.row {
  display: flex;
  align-items: center;
  padding: 32rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}
.row:last-child {
  border-bottom: none;
}
.label {
  width: 140rpx;
  font-size: 28rpx;
  color: #666;
}
.value {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}
.input {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}
.save-btn {
  margin-top: 40rpx;
  height: 88rpx;
  line-height: 88rpx;
  background: #07c160;
  color: #fff;
  font-size: 30rpx;
  border-radius: 44rpx;
  border: none;
}
</style>
