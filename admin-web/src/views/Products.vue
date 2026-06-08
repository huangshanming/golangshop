<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  centToYuan,
  fetchProducts,
  updateProductStatus,
  type ProductListItem,
} from '../api/products'

const router = useRouter()

const list = ref<ProductListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')

async function loadList() {
  const data = await fetchProducts({
    page: page.value,
    page_size: pageSize.value,
    keyword: keyword.value.trim() || undefined,
  })
  list.value = data.list || []
  total.value = data.total || 0
}

function formatPriceRange(row: ProductListItem) {
  if (row.min_price === row.max_price || row.max_price === 0) {
    return `¥${centToYuan(row.min_price).toFixed(2)}`
  }
  return `¥${centToYuan(row.min_price).toFixed(2)} ~ ¥${centToYuan(row.max_price).toFixed(2)}`
}

function openCreate() {
  router.push('/products/new')
}

function openEdit(row: ProductListItem) {
  router.push(`/products/${row.id}/edit`)
}

async function toggleStatus(row: ProductListItem) {
  const next = row.status === 1 ? 0 : 1
  const label = next === 1 ? '上架' : '下架'
  await ElMessageBox.confirm(`确定${label}该商品吗？`, '提示')
  await updateProductStatus(row.id, next)
  ElMessage.success(`${label}成功`)
  loadList()
}

function onPageChange(p: number) {
  page.value = p
  loadList()
}

onMounted(loadList)
</script>

<template>
  <el-card shadow="never">
    <div class="toolbar">
      <div class="toolbar-left">
        <h3>商品管理</h3>
        <el-input
          v-model="keyword"
          placeholder="搜索商品"
          clearable
          style="width: 220px; margin-left: 16px"
          @keyup.enter="loadList"
          @clear="loadList"
        />
        <el-button style="margin-left: 8px" @click="loadList">搜索</el-button>
      </div>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增商品</el-button>
    </div>

    <el-table :data="list" stripe style="width: 100%">
      <el-table-column label="封面" width="90">
        <template #default="{ row }">
          <el-image v-if="row.cover_image" :src="row.cover_image" fit="cover" style="width: 56px; height: 56px; border-radius: 6px" />
          <span v-else class="no-img">无图</span>
        </template>
      </el-table-column>
      <el-table-column label="名称" min-width="200">
        <template #default="{ row }">
          <div class="name-cell">
            <span class="name">{{ row.name }}</span>
            <span v-if="row.subtitle" class="subtitle">{{ row.subtitle }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="分类" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.category_name" size="small">{{ row.category_name }}</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="价格" width="160">
        <template #default="{ row }">{{ formatPriceRange(row) }}</template>
      </el-table-column>
      <el-table-column prop="total_stock" label="库存" width="80" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '上架' : '下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link :type="row.status === 1 ? 'warning' : 'success'" @click="toggleStatus(row)">
            {{ row.status === 1 ? '下架' : '上架' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="onPageChange"
      />
    </div>
  </el-card>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar-left {
  display: flex;
  align-items: center;
}
.toolbar-left h3 {
  margin: 0;
}
.name-cell {
  display: flex;
  flex-direction: column;
}
.name {
  font-weight: 500;
}
.subtitle {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
.no-img {
  color: #ccc;
  font-size: 12px;
}
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
