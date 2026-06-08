<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  addFeatured,
  deleteFeatured,
  fetchFeatured,
  searchProducts,
  updateFeaturedSort,
  type FeaturedItem,
  type SearchProductItem,
} from '../api/platform'
import { centToYuan } from '../utils/format'

const list = ref<FeaturedItem[]>([])
const dialogVisible = ref(false)
const keyword = ref('')
const searchResults = ref<SearchProductItem[]>([])
const searching = ref(false)

async function loadList() {
  list.value = await fetchFeatured()
}

async function handleSearch() {
  if (!keyword.value.trim()) return
  searching.value = true
  try {
    searchResults.value = await searchProducts(keyword.value.trim())
  } finally {
    searching.value = false
  }
}

async function handleAdd(row: SearchProductItem) {
  await addFeatured(row.id, list.value.length + 1)
  ElMessage.success('已添加')
  dialogVisible.value = false
  loadList()
}

async function handleDelete(row: FeaturedItem) {
  await ElMessageBox.confirm(`移除推荐「${row.product_name}」？`, '提示', { type: 'warning' })
  await deleteFeatured(row.id)
  ElMessage.success('已移除')
  loadList()
}

async function moveUp(index: number) {
  if (index <= 0) return
  const items = [...list.value]
  ;[items[index - 1], items[index]] = [items[index], items[index - 1]]
  await updateFeaturedSort(items.map((it, i) => ({ id: it.id, sort: i + 1 })))
  loadList()
}

async function moveDown(index: number) {
  if (index >= list.value.length - 1) return
  const items = [...list.value]
  ;[items[index + 1], items[index]] = [items[index], items[index + 1]]
  await updateFeaturedSort(items.map((it, i) => ({ id: it.id, sort: i + 1 })))
  loadList()
}

function openAdd() {
  keyword.value = ''
  searchResults.value = []
  dialogVisible.value = true
}

onMounted(loadList)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>首页推荐商品</h2>
      <el-button type="primary" :icon="Plus" @click="openAdd">添加商品</el-button>
    </div>
    <el-table :data="list" border>
      <el-table-column label="商品" min-width="200">
        <template #default="{ row }">
          <div class="prod-cell">
            <el-image :src="row.cover_image" style="width:48px;height:48px" fit="cover" />
            <div>
              <div>{{ row.product_name }}</div>
              <div class="sub">{{ row.tenant_name }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="价格" width="100">
        <template #default="{ row }">¥{{ centToYuan(row.min_price).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row, $index }">
          <el-button link @click="moveUp($index)">上移</el-button>
          <el-button link @click="moveDown($index)">下移</el-button>
          <el-button link type="danger" @click="handleDelete(row)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="搜索并添加商品" width="640px">
      <div class="search-row">
        <el-input v-model="keyword" placeholder="商品名称" @keyup.enter="handleSearch" />
        <el-button type="primary" :loading="searching" @click="handleSearch">搜索</el-button>
      </div>
      <el-table :data="searchResults" max-height="360" style="margin-top:12px">
        <el-table-column prop="name" label="商品" min-width="180" />
        <el-table-column prop="tenant_name" label="店铺" width="120" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleAdd(row)">添加</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.toolbar h2 { margin: 0; font-size: 18px; }
.prod-cell { display: flex; align-items: center; gap: 10px; }
.sub { font-size: 12px; color: #999; }
.search-row { display: flex; gap: 8px; }
</style>
