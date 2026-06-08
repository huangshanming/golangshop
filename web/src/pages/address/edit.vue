<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { createAddress, fetchAddresses, updateAddress, type Address } from '../../api/address'

const addressId = ref(0)
const form = ref({
  name: '',
  phone: '',
  province: '',
  city: '',
  district: '',
  detail: '',
  is_default: false,
})
const submitting = ref(false)

onLoad(async (query) => {
  addressId.value = Number(query?.id || 0)
  if (addressId.value) {
    try {
      const list = await fetchAddresses()
      const addr = list.find((a) => a.id === addressId.value)
      if (addr) {
        form.value = {
          name: addr.name,
          phone: addr.phone,
          province: addr.province,
          city: addr.city,
          district: addr.district,
          detail: addr.detail,
          is_default: addr.is_default,
        }
      }
    } catch {
      /* handled */
    }
  }
})

function validate() {
  const f = form.value
  if (!f.name.trim()) {
    uni.showToast({ title: '请输入收货人', icon: 'none' })
    return false
  }
  if (!/^1\d{10}$/.test(f.phone.trim())) {
    uni.showToast({ title: '请输入正确手机号', icon: 'none' })
    return false
  }
  if (!f.province.trim() || !f.city.trim() || !f.district.trim()) {
    uni.showToast({ title: '请填写省市区', icon: 'none' })
    return false
  }
  if (!f.detail.trim()) {
    uni.showToast({ title: '请输入详细地址', icon: 'none' })
    return false
  }
  return true
}

async function handleSubmit() {
  if (!validate()) return
  submitting.value = true
  const data: Omit<Address, 'id'> = {
    name: form.value.name.trim(),
    phone: form.value.phone.trim(),
    province: form.value.province.trim(),
    city: form.value.city.trim(),
    district: form.value.district.trim(),
    detail: form.value.detail.trim(),
    is_default: form.value.is_default,
  }
  try {
    if (addressId.value) {
      await updateAddress(addressId.value, data)
      uni.showToast({ title: '保存成功', icon: 'success' })
    } else {
      await createAddress(data)
      uni.showToast({ title: '添加成功', icon: 'success' })
    }
    setTimeout(() => uni.navigateBack(), 500)
  } catch {
    /* handled */
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <view class="page">
    <view class="form-card">
      <view class="field">
        <text class="label">收货人</text>
        <input v-model="form.name" class="input" placeholder="请输入收货人姓名" />
      </view>
      <view class="field">
        <text class="label">手机号</text>
        <input v-model="form.phone" class="input" type="number" maxlength="11" placeholder="请输入手机号" />
      </view>
      <view class="field">
        <text class="label">省份</text>
        <input v-model="form.province" class="input" placeholder="如：广东省" />
      </view>
      <view class="field">
        <text class="label">城市</text>
        <input v-model="form.city" class="input" placeholder="如：深圳市" />
      </view>
      <view class="field">
        <text class="label">区县</text>
        <input v-model="form.district" class="input" placeholder="如：南山区" />
      </view>
      <view class="field">
        <text class="label">详细地址</text>
        <input v-model="form.detail" class="input" placeholder="街道、门牌号等" />
      </view>
      <view class="field switch-field">
        <text class="label">设为默认地址</text>
        <switch :checked="form.is_default" color="#07c160" @change="(e: any) => form.is_default = e.detail.value" />
      </view>
    </view>

    <view class="bottom-bar">
      <button class="submit-btn" :disabled="submitting" @click="handleSubmit">
        {{ submitting ? '保存中...' : '保存' }}
      </button>
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
.form-card {
  margin: 16rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 8rpx 24rpx;
}
.field {
  display: flex;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}
.field:last-child {
  border-bottom: none;
}
.switch-field {
  justify-content: space-between;
}
.label {
  width: 180rpx;
  font-size: 28rpx;
  color: #333;
  flex-shrink: 0;
}
.input {
  flex: 1;
  font-size: 28rpx;
  color: #333;
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
.submit-btn {
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
