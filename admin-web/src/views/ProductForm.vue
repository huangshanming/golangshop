<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchCategories, type Category } from '../api/categories'
import {
  buildImagesJSON,
  centToYuan,
  createProduct,
  fetchProductDetail,
  parseImages,
  parseSpecs,
  parseSpecValues,
  updateProduct,
  yuanToCent,
} from '../api/products'
import ImageUpload from '../components/ImageUpload.vue'
import RichEditor from '../components/RichEditor.vue'
import SpecMatrix, { type SkuMatrixRow, type SpecDimension } from '../components/SpecMatrix.vue'

const route = useRoute()
const router = useRouter()

const categories = ref<Category[]>([])
const loading = ref(false)
const submitting = ref(false)

const editingId = computed(() => {
  const id = route.params.id
  return id ? Number(id) : 0
})

const pageTitle = computed(() => (editingId.value ? '编辑商品' : '新增商品'))

const form = reactive({
  category_id: 0,
  name: '',
  subtitle: '',
  description: '',
  cover_image: '',
  gallery: [] as string[],
  specs: [] as SpecDimension[],
  skus: [] as SkuMatrixRow[],
})

function firstCategoryId(nodes: Category[]): number {
  for (const n of nodes) {
    if (n.children?.length) {
      const child = firstCategoryId(n.children)
      if (child) return child
    }
    return n.id
  }
  return 0
}

function resetForm() {
  form.category_id = firstCategoryId(categories.value)
  form.name = ''
  form.subtitle = ''
  form.description = ''
  form.cover_image = ''
  form.gallery = []
  form.specs = []
  form.skus = []
}

async function loadCategories() {
  categories.value = await fetchCategories()
}

async function loadProduct() {
  if (!editingId.value) {
    resetForm()
    return
  }
  loading.value = true
  try {
    const data = await fetchProductDetail(editingId.value)
    const p = data.product
    form.category_id = p.category_id
    form.name = p.name
    form.subtitle = p.subtitle
    form.description = p.description
    form.cover_image = p.cover_image
    const imgs = parseImages(p.images)
    form.gallery = imgs.filter((u) => u !== p.cover_image)
    form.specs = parseSpecs(p.specs)
    form.skus = (data.skus || []).map((s) => ({
      spec_values: parseSpecValues(s.spec_values),
      spec: s.spec,
      priceYuan: centToYuan(s.price),
      originalPriceYuan: centToYuan(s.original_price),
      stock: s.stock,
      image: s.image || '',
    }))
    if (form.skus.length === 0 && !form.specs.length) {
      form.specs = [{ name: '规格', values: ['默认'] }]
      form.skus = [{ spec_values: { 规格: '默认' }, spec: '默认', priceYuan: 99, originalPriceYuan: 129, stock: 100, image: '' }]
    }
  } finally {
    loading.value = false
  }
}

function buildPayload() {
  const allImages = form.cover_image
    ? [form.cover_image, ...form.gallery.filter((u) => u !== form.cover_image)]
    : form.gallery
  return {
    category_id: form.category_id,
    name: form.name,
    subtitle: form.subtitle,
    description: form.description,
    cover_image: form.cover_image,
    images: buildImagesJSON(allImages),
    specs: form.specs,
    skus: form.skus.map((s) => ({
      spec: s.spec,
      spec_values: s.spec_values,
      price: yuanToCent(s.priceYuan),
      original_price: yuanToCent(s.originalPriceYuan),
      stock: s.stock,
      image: s.image,
    })),
  }
}

async function handleSubmit() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入商品名称')
    return
  }
  if (form.skus.length === 0) {
    ElMessage.warning('请生成 SKU 矩阵')
    return
  }
  submitting.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updateProduct(editingId.value, payload)
      ElMessage.success('更新成功')
    } else {
      await createProduct(payload)
      ElMessage.success('创建成功')
    }
    router.push('/products')
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.push('/products')
}

onMounted(async () => {
  await loadCategories()
  await loadProduct()
})
</script>

<template>
  <div v-loading="loading" class="product-form-page">
    <div class="page-header">
      <el-page-header @back="goBack">
        <template #content>
          <span class="page-title">{{ pageTitle }}</span>
        </template>
      </el-page-header>
    </div>

    <el-form label-width="90px" label-position="left" class="form-body">
      <el-card shadow="never" class="section-card">
        <template #header>基础信息</template>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="商品名称" style="max-width: 480px" />
        </el-form-item>
        <el-form-item label="副标题">
          <el-input v-model="form.subtitle" placeholder="卖点简述" style="max-width: 480px" />
        </el-form-item>
        <el-form-item label="分类">
          <el-tree-select
            v-model="form.category_id"
            :data="categories"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            placeholder="选择平台分类"
            style="width: 280px"
          />
        </el-form-item>
        <el-form-item label="封面图">
          <ImageUpload v-model="form.cover_image" />
        </el-form-item>
        <el-form-item label="轮播图">
          <ImageUpload v-model="form.gallery" multiple :limit="8" />
        </el-form-item>
      </el-card>

      <el-card shadow="never" class="section-card">
        <template #header>规格与库存</template>
        <SpecMatrix v-model:specs="form.specs" v-model:skus="form.skus" />
      </el-card>

      <el-card shadow="never" class="section-card">
        <template #header>商品详情</template>
        <RichEditor v-model="form.description" />
      </el-card>
    </el-form>

    <div class="footer-bar">
      <el-button @click="goBack">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
    </div>
  </div>
</template>

<style scoped>
.product-form-page {
  max-width: 1200px;
  margin: 0 auto;
  padding-bottom: 80px;
}
.page-header {
  margin-bottom: 20px;
}
.page-title {
  font-size: 18px;
  font-weight: 600;
}
.section-card {
  margin-bottom: 16px;
}
.section-card :deep(.el-card__header) {
  font-weight: 600;
  padding: 14px 20px;
}
.footer-bar {
  position: sticky;
  bottom: 0;
  z-index: 10;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  background: #fff;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.04);
}
</style>
