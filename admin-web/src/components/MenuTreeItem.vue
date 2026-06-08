<script setup lang="ts">
import type { MenuTreeNode } from '../api/auth'
import MenuTreeItem from './MenuTreeItem.vue'

defineProps<{
  item: MenuTreeNode
}>()
</script>

<template>
  <el-sub-menu v-if="item.children?.length" :index="`group-${item.id}`">
    <template #title>{{ item.title }}</template>
    <MenuTreeItem v-for="child in item.children" :key="child.id" :item="child" />
  </el-sub-menu>
  <el-menu-item v-else-if="item.path" :index="item.path">
    {{ item.title }}
  </el-menu-item>
</template>
