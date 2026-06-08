<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  createBanner,
  deleteBanner,
  fetchBanners,
  fetchPlatformCategories,
  updateBanner,
  type HomeBanner,
  type PlatformCategory,
} from '../api/platform'
import { fetchTenants, type Tenant } from '../api/tenant'
import ImageUpload from '../components/ImageUpload.vue'

const positionFilter = ref<'home' | 'category'>('home')
const list = ref<HomeBanner[]>([])
const categories = ref<PlatformCategory[]>([])
const tenants = ref<Tenant[]>([])
const dialogVisible = ref(false)
const editing = ref<HomeBanner | null>(null)
const form = ref({
  title: '',
  image_url: '',
  position: 'home' as HomeBanner['position'],
  sort: 0,
  status: 1,
  link_type: 'category' as HomeBanner['link_type'],
  link_target: '',
})

const flatCategories = computed(() => {
  const out: { id: number; name: string }[] = []
  const walk = (nodes: PlatformCategory[], prefix = '') => {
    nodes.forEach((n) => {
      out.push({ id: n.id, name: prefix + n.name })
      if (n.children?.length) walk(n.children, prefix + n.name + ' / ')
    })
  }
  walk(categories.value)
  return out
})

const positionLabel: Record<string, string> = {
  home: '首页',
  category: '分类页',
}

async function loadList() {
  list.value = await fetchBanners(positionFilter.value)
}

async function loadOptions() {
  categories.value = await fetchPlatformCategories()
  tenants.value = await fetchTenants()
}

function onPositionChange() {
  loadList()
}

function openCreate() {
  editing.value = null
  form.value = {
    title: '',
    image_url: '',
    position: positionFilter.value,
    sort: 0,
    status: 1,
    link_type: 'category',
    link_target: '',
  }
  dialogVisible.value = true
}

function openEdit(row: HomeBanner) {
  editing.value = row
  form.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.image_url) {
    ElMessage.warning('请上传 Banner 图片')
    return
  }
  if (!form.value.link_target) {
    ElMessage.warning('请填写跳转目标')
    return
  }
  if (editing.value) {
    await updateBanner(editing.value.id, form.value)
    ElMessage.success('已更新')
  } else {
    await createBanner(form.value)
    ElMessage.success('已创建')
  }
  dialogVisible.value = false
  loadList()
}

async function handleDelete(row: HomeBanner) {
  await ElMessageBox.confirm('确定删除该 Banner？', '提示', { type: 'warning' })
  await deleteBanner(row.id)
  ElMessage.success('已删除')
  loadList()
}

const linkTypeLabel: Record<string, string> = {
  product: '商品',
  tenant: '店铺',
  category: '分类',
  url: '外链',
}

onMounted(async () => {
  await loadOptions()
  await loadList()
})
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>Banner 管理</h2>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增 Banner</el-button>
    </div>

    <el-tabs v-model="positionFilter" class="position-tabs" @tab-change="onPositionChange">
      <el-tab-pane label="首页 Banner" name="home" />
      <el-tab-pane label="分类页 Banner" name="category" />
    </el-tabs>

    <el-table :data="list" border>
      <el-table-column label="图片" width="160">
        <template #default="{ row }">
          <el-image :src="row.image_url" style="width:140px;height:56px" fit="cover" />
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="120" />
      <el-table-column label="展示位置" width="100">
        <template #default="{ row }">
          {{ positionLabel[row.position] || row.position }}
        </template>
      </el-table-column>
      <el-table-column label="跳转" min-width="160">
        <template #default="{ row }">
          {{ linkTypeLabel[row.link_type] || row.link_type }}：{{ row.link_target }}
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑 Banner' : '新增 Banner'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="展示位置" required>
          <el-radio-group v-model="form.position">
            <el-radio value="home">首页</el-radio>
            <el-radio value="category">分类页</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="图片" required>
          <ImageUpload v-model="form.image_url" />
        </el-form-item>
        <el-form-item label="跳转类型">
          <el-select v-model="form.link_type" style="width:100%">
            <el-option value="category" label="分类" />
            <el-option value="product" label="商品" />
            <el-option value="tenant" label="店铺" />
            <el-option value="url" label="外链" />
          </el-select>
        </el-form-item>
        <el-form-item label="跳转目标">
          <el-select
            v-if="form.link_type === 'category'"
            v-model="form.link_target"
            style="width:100%"
            filterable
          >
            <el-option v-for="c in flatCategories" :key="c.id" :value="String(c.id)" :label="c.name" />
          </el-select>
          <el-select
            v-else-if="form.link_type === 'tenant'"
            v-model="form.link_target"
            style="width:100%"
          >
            <el-option v-for="t in tenants" :key="t.id" :value="String(t.id)" :label="t.name" />
          </el-select>
          <el-input
            v-else
            v-model="form.link_target"
            :placeholder="form.link_type === 'product' ? '商品 ID' : 'https://...'"
          />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
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
.position-tabs { margin-bottom: 16px; }
</style>
