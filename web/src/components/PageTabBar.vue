<script setup lang="ts">
/** H5：Teleport 到 body，避免 uni-page 容器导致 fixed 失效 */
import AppTabBar from './AppTabBar.vue'
</script>

<template>
  <!-- #ifdef H5 -->
  <Teleport to="body">
    <AppTabBar fixed />
  </Teleport>
  <!-- #endif -->
</template>
