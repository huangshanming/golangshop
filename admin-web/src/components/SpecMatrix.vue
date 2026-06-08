<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import ImageUpload from './ImageUpload.vue'

export interface SpecDimension {
  name: string
  values: string[]
}

export interface SkuMatrixRow {
  spec_values: Record<string, string>
  spec: string
  priceYuan: number
  originalPriceYuan: number
  stock: number
  image: string
}

const props = defineProps<{
  specs: SpecDimension[]
  skus: SkuMatrixRow[]
}>()

const emit = defineEmits<{
  'update:specs': [SpecDimension[]]
  'update:skus': [SkuMatrixRow[]]
}>()

const valueInputs = ref<Record<number, string>>({})

watch(
  () => props.specs,
  (specs) => {
    specs.forEach((_, i) => {
      if (valueInputs.value[i] === undefined) valueInputs.value[i] = ''
    })
  },
  { immediate: true, deep: true },
)

function emitSpecs(next: SpecDimension[]) {
  emit('update:specs', next)
}

function emitSkus(next: SkuMatrixRow[]) {
  emit('update:skus', next)
}

function addDimension() {
  emitSpecs([...props.specs, { name: '', values: [] }])
}

async function removeDimension(index: number) {
  await ElMessageBox.confirm('删除规格维度将清空 SKU 矩阵，需重新生成', '提示')
  const next = props.specs.filter((_, i) => i !== index)
  emitSpecs(next)
  emitSkus([])
}

function addValue(dimIndex: number) {
  const val = (valueInputs.value[dimIndex] || '').trim()
  if (!val) return
  const next = props.specs.map((d, i) => {
    if (i !== dimIndex) return d
    if (d.values.includes(val)) return d
    return { ...d, values: [...d.values, val] }
  })
  valueInputs.value[dimIndex] = ''
  emitSpecs(next)
}

function removeValue(dimIndex: number, valIndex: number) {
  const next = props.specs.map((d, i) => {
    if (i !== dimIndex) return d
    return { ...d, values: d.values.filter((_, vi) => vi !== valIndex) }
  })
  emitSpecs(next)
  emitSkus([])
}

function cartesian(dims: SpecDimension[]): Record<string, string>[] {
  if (dims.length === 0) return [{}]
  let combos: Record<string, string>[] = [{}]
  for (const dim of dims) {
    const name = dim.name.trim()
    if (!name || dim.values.length === 0) continue
    const next: Record<string, string>[] = []
    for (const c of combos) {
      for (const v of dim.values) {
        next.push({ ...c, [name]: v })
      }
    }
    combos = next
  }
  return combos
}

function formatSummary(specValues: Record<string, string>) {
  const parts = props.specs
    .map((d) => specValues[d.name.trim()])
    .filter(Boolean)
  return parts.join(' / ')
}

function generateMatrix() {
  for (const d of props.specs) {
    if (!d.name.trim()) {
      ElMessage.warning('请填写规格维度名称')
      return
    }
    if (d.values.length === 0) {
      ElMessage.warning(`规格「${d.name}」至少添加一个值`)
      return
    }
  }
  const combos = cartesian(props.specs)
  if (combos.length === 0) {
    ElMessage.warning('请先定义规格维度')
    return
  }
  const oldMap = new Map<string, SkuMatrixRow>()
  for (const row of props.skus) {
    const key = props.specs.map((d) => `${d.name}=${row.spec_values[d.name] || ''}`).join('|')
    oldMap.set(key, row)
  }
  const rows: SkuMatrixRow[] = combos.map((specValues) => {
    const key = props.specs.map((d) => `${d.name}=${specValues[d.name] || ''}`).join('|')
    const old = oldMap.get(key)
    return {
      spec_values: specValues,
      spec: formatSummary(specValues),
      priceYuan: old?.priceYuan ?? 99,
      originalPriceYuan: old?.originalPriceYuan ?? 129,
      stock: old?.stock ?? 100,
      image: old?.image ?? '',
    }
  })
  emitSkus(rows)
  ElMessage.success(`已生成 ${rows.length} 个 SKU`)
}

function batchFill(field: 'priceYuan' | 'originalPriceYuan' | 'stock', val: number) {
  emitSkus(props.skus.map((r) => ({ ...r, [field]: val })))
}
</script>

<template>
  <div class="spec-matrix">
    <div v-for="(dim, di) in specs" :key="di" class="dim-block">
      <div class="dim-head">
        <el-input v-model="dim.name" placeholder="规格名，如 颜色" style="width: 160px" />
        <el-button link type="danger" @click="removeDimension(di)">删除维度</el-button>
      </div>
      <div class="values">
        <el-tag
          v-for="(v, vi) in dim.values"
          :key="v"
          closable
          class="val-tag"
          @close="removeValue(di, vi)"
        >
          {{ v }}
        </el-tag>
        <el-input
          v-model="valueInputs[di]"
          placeholder="输入后回车添加"
          style="width: 140px"
          size="small"
          @keyup.enter="addValue(di)"
        />
        <el-button size="small" @click="addValue(di)">添加</el-button>
      </div>
    </div>
    <el-button type="primary" link :icon="Plus" @click="addDimension">添加规格维度</el-button>

    <div class="matrix-actions">
      <el-button type="primary" @click="generateMatrix">生成 SKU 矩阵</el-button>
      <el-button @click="batchFill('priceYuan', 99)">批量售价 99</el-button>
      <el-button @click="batchFill('stock', 100)">批量库存 100</el-button>
    </div>

    <el-table v-if="skus.length" :data="skus" border size="small" class="sku-table">
      <el-table-column label="规格组合" min-width="160">
        <template #default="{ row }">{{ row.spec }}</template>
      </el-table-column>
      <el-table-column label="售价(元)" width="120">
        <template #default="{ row }">
          <el-input-number v-model="row.priceYuan" :min="0.01" :precision="2" size="small" />
        </template>
      </el-table-column>
      <el-table-column label="原价(元)" width="120">
        <template #default="{ row }">
          <el-input-number v-model="row.originalPriceYuan" :min="0" :precision="2" size="small" />
        </template>
      </el-table-column>
      <el-table-column label="库存" width="100">
        <template #default="{ row }">
          <el-input-number v-model="row.stock" :min="0" size="small" />
        </template>
      </el-table-column>
      <el-table-column label="SKU图" width="110">
        <template #default="{ row }">
          <ImageUpload v-model="row.image" size="small" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.spec-matrix {
  width: 100%;
}
.dim-block {
  background: #fafafa;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}
.dim-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.values {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.val-tag {
  margin-right: 4px;
}
.matrix-actions {
  margin: 16px 0;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.sku-table {
  margin-top: 8px;
}
</style>
