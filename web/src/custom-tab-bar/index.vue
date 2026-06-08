<script setup lang="ts">
import { ref } from 'vue'
import AppTabBar from '../components/AppTabBar.vue'

const tabBarRef = ref<InstanceType<typeof AppTabBar> | null>(null)

function updateSelected() {
  tabBarRef.value?.updateSelected()
}

defineExpose({ updateSelected })
</script>

<template>
  <!-- 小程序端由框架渲染 custom-tab-bar -->
  <!-- #ifndef H5 -->
  <view class="custom-tab-bar-root">
    <AppTabBar ref="tabBarRef" />
  </view>
  <!-- #endif -->
</template>

<style scoped>
.custom-tab-bar-root {
  width: 100%;
  height: calc(120rpx + env(safe-area-inset-bottom));
  overflow: visible;
  background: #fff;
}
</style>
