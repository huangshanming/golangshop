<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { useTenantStore } from './stores/tenant'

onLaunch(async () => {
  const tenantStore = useTenantStore()
  try {
    await tenantStore.loadTenants()
  } catch {
    /* ignore */
  }
})
</script>

<template>
  <view class="app">
    <slot />
  </view>
</template>

<style>
.app {
  background: #f5f5f5;
}

/* H5：隐藏框架默认 TabBar，使用 Teleport 到 body 的 AppTabBar */
/* #ifdef H5 */
html,
body {
  margin: 0;
  padding: 0;
  height: 100%;
}

.uni-tabbar,
.uni-tabbar-bottom,
.uni-app--showtabbar .uni-tabbar {
  display: none !important;
  height: 0 !important;
  overflow: hidden !important;
}

.uni-app--showtabbar uni-page-wrapper,
.uni-app--showtabbar uni-page-body {
  padding-bottom: 0 !important;
}

uni-page-body {
  min-height: 100%;
}

:root {
  --goshop-tab-bar-height: calc(50px + env(safe-area-inset-bottom, 0px));
}
/* #endif */
</style>