<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onLoad, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import {
  fetchBanners,
  fetchCategoryTree,
  type CategoryTreeItem,
  type HomeBanner,
} from '../../api/home'
import { syncTabBarSelected } from '../../utils/tabbar'
import PageTabBar from '../../components/PageTabBar.vue'
import { catGradient } from '../../utils/theme'

const tree = ref<CategoryTreeItem[]>([])
const banners = ref<HomeBanner[]>([])
const activeIndex = ref(0)
const keyword = ref('')
const loading = ref(false)

const activeParent = computed(() => tree.value[activeIndex.value] || null)

const rightItems = computed(() => {
  const parent = activeParent.value
  if (!parent) return []
  if (parent.children?.length) return parent.children
  return [parent]
})

async function loadData() {
  loading.value = true
  try {
    const [treeData, bannerList] = await Promise.all([
      fetchCategoryTree(),
      fetchBanners('category'),
    ])
    tree.value = treeData || []
    banners.value = bannerList || []
    if (activeIndex.value >= tree.value.length) {
      activeIndex.value = 0
    }
  } catch {
    tree.value = []
    banners.value = []
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

function selectParent(index: number) {
  activeIndex.value = index
}

function onSearch() {
  const kw = keyword.value.trim()
  uni.navigateTo({
    url: `/pages/product/list?keyword=${encodeURIComponent(kw)}`,
  })
}

function goCategory(cat: CategoryTreeItem) {
  uni.navigateTo({
    url: `/pages/product/list?category_id=${cat.id}&category_name=${encodeURIComponent(cat.name)}`,
  })
}

function handleBanner(banner: HomeBanner) {
  const target = banner.link_target
  switch (banner.link_type) {
    case 'category':
      uni.navigateTo({ url: `/pages/product/list?category_id=${target}` })
      break
    case 'product':
      uni.navigateTo({ url: `/pages/product/detail?id=${target}` })
      break
    case 'tenant':
      uni.navigateTo({ url: `/pages/shop/index?tenant_id=${target}` })
      break
    case 'url':
      // #ifdef H5
      window.open(target, '_blank')
      // #endif
      // #ifndef H5
      uni.setClipboardData({ data: target })
      uni.showToast({ title: '链接已复制', icon: 'none' })
      // #endif
      break
  }
}

watch(tree, (val) => {
  if (activeIndex.value >= val.length) activeIndex.value = 0
})

onLoad(() => loadData())
onShow(() => {
  uni.setNavigationBarTitle({ title: '商品分类' })
  syncTabBarSelected()
})
onPullDownRefresh(() => loadData())
</script>

<template>
  <view class="page">
    <view class="page-bg">
      <view class="blob blob-a" />
      <view class="blob blob-b" />
    </view>

    <view class="header">
      <view class="search-row">
        <view class="search-glass">
          <text class="search-icon">🔍</text>
          <input
            v-model="keyword"
            class="search-input"
            placeholder="输入名称搜索商品"
            confirm-type="search"
            @confirm="onSearch"
          />
        </view>
        <view class="search-btn" @click="onSearch">
          <text class="search-btn-text">搜索</text>
        </view>
      </view>

      <view v-if="banners.length" class="banner-stage">
        <swiper
          class="banner-swiper"
          circular
          autoplay
          indicator-dots
          indicator-color="rgba(255,255,255,0.4)"
          indicator-active-color="#fff"
        >
          <swiper-item v-for="b in banners" :key="b.id">
            <view class="banner-card" @click="handleBanner(b)">
              <image class="banner-img" :src="b.image_url" mode="aspectFill" />
              <view class="banner-shine" />
            </view>
          </swiper-item>
        </swiper>
      </view>
    </view>

    <view v-if="loading && !tree.length" class="loading">
      <text class="loading-text">加载中...</text>
    </view>

    <view v-else-if="tree.length" class="body">
      <scroll-view class="sidebar" scroll-y :show-scrollbar="false">
        <view
          v-for="(parent, index) in tree"
          :key="parent.id"
          class="sidebar-item"
          :class="{ active: index === activeIndex }"
          @click="selectParent(index)"
        >
          <view v-if="index === activeIndex" class="sidebar-pill">
            <view class="sidebar-indicator" />
            <text class="sidebar-text">{{ parent.name }}</text>
          </view>
          <text v-else class="sidebar-text">{{ parent.name }}</text>
        </view>
      </scroll-view>

      <scroll-view class="content" scroll-y :show-scrollbar="false">
        <view v-if="activeParent" class="content-head">
          <view class="content-head-left">
            <view class="head-accent" />
            <text class="content-title">{{ activeParent.name }}</text>
          </view>
          <view class="content-all-btn" @click="goCategory(activeParent)">
            <text class="content-all">查看全部</text>
            <text class="content-all-arrow">›</text>
          </view>
        </view>

        <view
          v-for="(item, index) in rightItems"
          :key="item.id"
          class="list-card"
          @click="goCategory(item)"
        >
          <view class="item-icon" :style="{ background: catGradient(index) }">
            <view class="item-icon-shine" />
            <image v-if="item.icon" class="item-icon-img" :src="item.icon" mode="aspectFill" />
            <text v-else class="item-icon-text">{{ item.name.slice(0, 1) }}</text>
          </view>
          <view class="item-info">
            <text class="item-name">{{ item.name }}</text>
            <text class="item-desc">探索 {{ item.name }} 精选好物</text>
          </view>
          <view class="item-go">
            <text class="item-arrow">›</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <view v-else-if="!loading" class="empty">
      <text class="empty-icon">📂</text>
      <text class="empty-text">暂无分类</text>
    </view>
  </view>
  <PageTabBar />
</template>

<style scoped>
.page {
  position: relative;
  display: flex;
  flex-direction: column;
  background: linear-gradient(165deg, #eef1fb 0%, #f5f0ee 40%, #edf4f8 100%);
  box-sizing: border-box;
  overflow: hidden;
}

/* #ifdef H5 */
.page {
  height: 100vh;
  min-height: 100vh;
}
/* #endif */

/* #ifndef H5 */
.page {
  height: 100%;
}
/* #endif */

.page-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(60rpx);
}

.blob-a {
  width: 420rpx;
  height: 420rpx;
  top: -80rpx;
  left: -100rpx;
  background: radial-gradient(circle, rgba(107, 124, 255, 0.3) 0%, transparent 70%);
}

.blob-b {
  width: 360rpx;
  height: 360rpx;
  top: 200rpx;
  right: -80rpx;
  background: radial-gradient(circle, rgba(61, 190, 184, 0.25) 0%, transparent 70%);
}

.header {
  position: relative;
  z-index: 2;
  flex-shrink: 0;
  padding: 20rpx 24rpx 16rpx;
}

.search-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.search-glass {
  flex: 1;
  display: flex;
  align-items: center;
  height: 80rpx;
  padding: 0 28rpx;
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(20rpx);
  -webkit-backdrop-filter: blur(20rpx);
  border-radius: 40rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.95);
  box-shadow:
    0 8rpx 28rpx rgba(30, 40, 80, 0.08),
    inset 0 2rpx 6rpx rgba(255, 255, 255, 0.9),
    inset 0 -2rpx 6rpx rgba(30, 40, 80, 0.03);
}

.search-icon {
  font-size: 28rpx;
  margin-right: 12rpx;
  opacity: 0.5;
}

.search-input {
  flex: 1;
  font-size: 28rpx;
  height: 80rpx;
  color: #1a1f36;
}

.search-btn {
  flex-shrink: 0;
  height: 80rpx;
  padding: 0 32rpx;
  display: flex;
  align-items: center;
  background: linear-gradient(145deg, #7b8cff 0%, #6b7cff 45%, #5a68e6 100%);
  border-radius: 24rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.35);
  box-shadow:
    0 10rpx 28rpx rgba(107, 124, 255, 0.4),
    inset 0 3rpx 6rpx rgba(255, 255, 255, 0.35),
    inset 0 -4rpx 8rpx rgba(50, 60, 160, 0.2);
}

.search-btn-text {
  color: #fff;
  font-size: 28rpx;
  font-weight: 700;
  text-shadow: 0 2rpx 4rpx rgba(40, 50, 140, 0.3);
}

.banner-stage {
  filter: drop-shadow(0 14rpx 36rpx rgba(30, 40, 80, 0.16));
}

.banner-swiper {
  width: 100%;
  height: 220rpx;
  border-radius: 24rpx;
  overflow: hidden;
}

.banner-card {
  position: relative;
  width: 100%;
  height: 220rpx;
  border-radius: 24rpx;
  overflow: hidden;
  border: 3rpx solid rgba(255, 255, 255, 0.55);
  box-shadow: inset 0 2rpx 8rpx rgba(255, 255, 255, 0.25);
}

.banner-img {
  width: 100%;
  height: 100%;
}

.banner-shine {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 45%;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.2) 0%, transparent 100%);
  pointer-events: none;
}

.body {
  position: relative;
  z-index: 1;
  flex: 1;
  display: flex;
  min-height: 0;
  height: 0;
  margin: 0 16rpx 16rpx;
  border-radius: 28rpx;
  overflow: hidden;
  border: 2rpx solid rgba(255, 255, 255, 0.85);
  box-shadow:
    0 16rpx 48rpx rgba(30, 40, 80, 0.1),
    inset 0 2rpx 0 rgba(255, 255, 255, 0.9);
}

.sidebar {
  width: 188rpx;
  flex-shrink: 0;
  background: rgba(240, 243, 250, 0.85);
  height: 100%;
}

.sidebar-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100rpx;
  padding: 20rpx 10rpx;
  box-sizing: border-box;
}

.sidebar-pill {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-height: 72rpx;
  padding: 12rpx 8rpx 12rpx 16rpx;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 0 20rpx 20rpx 0;
  margin-right: -2rpx;
  box-shadow:
    4rpx 4rpx 16rpx rgba(30, 40, 80, 0.08),
    inset 0 2rpx 4rpx rgba(255, 255, 255, 0.95);
}

.sidebar-indicator {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 6rpx;
  height: 40rpx;
  background: linear-gradient(180deg, #6b7cff, #9b6bff);
  border-radius: 0 4rpx 4rpx 0;
  box-shadow: 0 4rpx 12rpx rgba(107, 124, 255, 0.45);
}

.sidebar-text {
  font-size: 26rpx;
  color: #7a8499;
  text-align: center;
  line-height: 1.4;
  font-weight: 500;
}

.sidebar-pill .sidebar-text {
  color: #1a1f36;
  font-weight: 800;
}

.content {
  flex: 1;
  background: rgba(255, 255, 255, 0.88);
  height: 100%;
  padding: 0 20rpx 16rpx;
  box-sizing: border-box;
}

.content-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 4rpx 16rpx;
  border-bottom: 2rpx solid rgba(240, 243, 250, 0.9);
  margin-bottom: 12rpx;
}

.content-head-left {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.head-accent {
  width: 8rpx;
  height: 32rpx;
  border-radius: 4rpx;
  background: linear-gradient(180deg, #3dbeb8, #6b7cff);
  box-shadow: 0 4rpx 12rpx rgba(61, 190, 184, 0.4);
}

.content-title {
  font-size: 32rpx;
  font-weight: 800;
  color: #1a1f36;
}

.content-all-btn {
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: 8rpx 18rpx;
  background: linear-gradient(145deg, #f8f9fc, #eef1f8);
  border-radius: 20rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.9);
  box-shadow: 0 4rpx 12rpx rgba(30, 40, 80, 0.06);
}

.content-all {
  font-size: 24rpx;
  color: #6b7cff;
  font-weight: 600;
}

.content-all-arrow {
  font-size: 26rpx;
  color: #6b7cff;
  font-weight: 700;
}

.list-card {
  display: flex;
  align-items: center;
  padding: 20rpx;
  margin-bottom: 16rpx;
  background: linear-gradient(145deg, #ffffff, #f8faff);
  border-radius: 22rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.95);
  box-shadow:
    0 8rpx 24rpx rgba(30, 40, 80, 0.07),
    0 2rpx 8rpx rgba(30, 40, 80, 0.04),
    inset 0 2rpx 0 rgba(255, 255, 255, 0.95);
}

.item-icon {
  position: relative;
  width: 96rpx;
  height: 96rpx;
  border-radius: 22rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: 20rpx;
  border: 2rpx solid rgba(255, 255, 255, 0.5);
  box-shadow:
    0 8rpx 20rpx rgba(30, 40, 80, 0.15),
    inset 0 3rpx 6rpx rgba(255, 255, 255, 0.4),
    inset 0 -4rpx 8rpx rgba(0, 0, 0, 0.1);
}

.item-icon-shine {
  position: absolute;
  top: 10rpx;
  left: 14rpx;
  width: 28rpx;
  height: 14rpx;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 50%;
  filter: blur(3rpx);
  transform: rotate(-20deg);
  pointer-events: none;
}

.item-icon-img {
  width: 60rpx;
  height: 60rpx;
  border-radius: 14rpx;
  position: relative;
  z-index: 1;
}

.item-icon-text {
  color: #fff;
  font-size: 36rpx;
  font-weight: 800;
  position: relative;
  z-index: 1;
  text-shadow: 0 2rpx 6rpx rgba(0, 0, 0, 0.2);
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: #1a1f36;
  margin-bottom: 8rpx;
}

.item-desc {
  display: block;
  font-size: 24rpx;
  color: #8b93a8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-go {
  width: 48rpx;
  height: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(145deg, #f0f2f8, #e8ecf4);
  border-radius: 50%;
  box-shadow:
    0 4rpx 12rpx rgba(30, 40, 80, 0.08),
    inset 0 2rpx 4rpx rgba(255, 255, 255, 0.9);
  margin-left: 8rpx;
}

.item-arrow {
  font-size: 32rpx;
  color: #6b7cff;
  font-weight: 700;
  line-height: 1;
}

.loading,
.empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
}

.loading-text,
.empty-text {
  color: #8b93a8;
  font-size: 28rpx;
}

.empty-icon {
  font-size: 64rpx;
}
</style>
