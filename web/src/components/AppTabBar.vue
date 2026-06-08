<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'

withDefaults(defineProps<{
  fixed?: boolean
}>(), {
  fixed: false,
})

const selected = ref(0)

const tabs = [
  { path: '/pages/index/index', text: '首页', icon: '🏠' },
  { path: '/pages/category/index', text: '分类' },
  { path: '/pages/login/index', text: '我的', icon: '👤' },
]

function updateSelected() {
  const pages = getCurrentPages()
  const cur = pages[pages.length - 1]
  const route = cur ? `/${cur.route}` : ''
  const idx = tabs.findIndex((t) => route === t.path)
  selected.value = idx >= 0 ? idx : 0
}

function switchTab(index: number) {
  if (index === selected.value) return
  uni.switchTab({ url: tabs[index].path })
}

onShow(updateSelected)

defineExpose({ updateSelected })
</script>

<template>
  <view class="tab-bar" :class="{ 'is-fixed': fixed }">
    <view class="tab-bar-panel">
      <view v-if="!fixed" class="tab-bar-top">
        <view class="cat-orb" :class="{ active: selected === 1 }" @click="switchTab(1)">
          <view class="cat-orb-inner">
            <text class="cat-badge-icon">▦</text>
          </view>
        </view>
      </view>

      <view class="tab-bar-row">
        <view class="tab-item" @click="switchTab(0)">
          <view class="tab-icon-box">
            <view class="tab-icon-wrap" :class="{ active: selected === 0 }">
              <text class="tab-emoji">{{ tabs[0].icon }}</text>
            </view>
          </view>
          <text class="tab-text" :class="{ active: selected === 0 }">{{ tabs[0].text }}</text>
        </view>

        <view class="tab-item tab-item-center" @click="switchTab(1)">
          <view v-if="fixed" class="cat-orb-inline" :class="{ active: selected === 1 }">
            <text class="cat-badge-icon">▦</text>
          </view>
          <view v-else class="tab-icon-box" />
          <text class="tab-text" :class="{ active: selected === 1 }">{{ tabs[1].text }}</text>
        </view>

        <view class="tab-item" @click="switchTab(2)">
          <view class="tab-icon-box">
            <view class="tab-icon-wrap" :class="{ active: selected === 2 }">
              <text class="tab-emoji">{{ tabs[2].icon }}</text>
            </view>
          </view>
          <text class="tab-text" :class="{ active: selected === 2 }">{{ tabs[2].text }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.tab-bar {
  width: 100%;
  height: 100%;
}

.tab-bar.is-fixed {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99999;
  width: 100%;
  height: calc(50px + env(safe-area-inset-bottom, 0px));
  box-sizing: border-box;
  padding-bottom: env(safe-area-inset-bottom, 0px);
  background: #fff;
  border-top: 1px solid #eef0f5;
  box-shadow: 0 -2px 12px rgba(30, 40, 80, 0.08);
}

.tab-bar-panel {
  width: 100%;
  height: 100%;
  background: #fff;
  border-top: 2rpx solid rgba(240, 243, 250, 0.95);
  border-radius: 28rpx 28rpx 0 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: visible;
}

.is-fixed .tab-bar-panel {
  border-radius: 0;
  border-top: none;
  box-shadow: none;
}

.tab-bar-top {
  height: 36rpx;
  display: flex;
  justify-content: center;
  flex-shrink: 0;
  overflow: visible;
}

.cat-orb {
  width: 80rpx;
  height: 80rpx;
  margin-bottom: -40rpx;
  flex-shrink: 0;
  opacity: 0.9;
  z-index: 2;
}

.cat-orb.active {
  opacity: 1;
}

.cat-orb-inner {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: linear-gradient(145deg, #6ee8e2 0%, #3dbeb8 45%, #2a9d96 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 3rpx solid #fff;
  box-shadow: 0 8rpx 20rpx rgba(61, 190, 184, 0.35);
}

.cat-orb-inline {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(145deg, #6ee8e2 0%, #3dbeb8 45%, #2a9d96 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #fff;
  box-shadow: 0 4px 12px rgba(61, 190, 184, 0.35);
  opacity: 0.88;
}

.cat-orb-inline.active {
  opacity: 1;
}

.cat-badge-icon {
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  line-height: 1;
}

.is-fixed .cat-badge-icon {
  font-size: 14px;
}

.tab-bar-row {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 0 8px 2px;
  box-sizing: border-box;
  min-height: 0;
}

.is-fixed .tab-bar-row {
  height: 50px;
  padding: 0 8px;
}

.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.tab-icon-box {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.tab-icon-wrap {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  box-sizing: border-box;
}

.tab-icon-wrap.active {
  background: rgba(107, 124, 255, 0.12);
}

.tab-emoji {
  font-size: 20px;
  line-height: 24px;
  height: 24px;
  opacity: 0.45;
}

.tab-icon-wrap.active .tab-emoji {
  opacity: 1;
}

.tab-text {
  height: 14px;
  line-height: 14px;
  font-size: 11px;
  color: #9aa3b5;
  font-weight: 500;
}

.tab-text.active {
  color: #6b7cff;
  font-weight: 700;
}

/* 小程序端尺寸 */
.tab-bar:not(.is-fixed) .tab-icon-box {
  width: 48rpx;
  height: 48rpx;
}

.tab-bar:not(.is-fixed) .tab-icon-wrap {
  width: 48rpx;
  height: 48rpx;
  border-radius: 14rpx;
  border: 2rpx solid transparent;
}

.tab-bar:not(.is-fixed) .tab-icon-wrap.active {
  border-color: rgba(107, 124, 255, 0.08);
}

.tab-bar:not(.is-fixed) .tab-emoji {
  font-size: 34rpx;
  line-height: 48rpx;
  height: 48rpx;
}

.tab-bar:not(.is-fixed) .tab-text {
  margin-top: 4rpx;
  height: 28rpx;
  line-height: 28rpx;
  font-size: 22rpx;
}

.tab-bar:not(.is-fixed) .cat-badge-icon {
  font-size: 28rpx;
}

.tab-bar:not(.is-fixed) .tab-bar-row {
  padding: 0 16rpx 8rpx;
  min-height: 84rpx;
}
</style>
