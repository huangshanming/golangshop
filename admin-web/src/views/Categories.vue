<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  createPlatformCategory,
  deletePlatformCategory,
  fetchPlatformCategories,
  updatePlatformCategory,
  type PlatformCategory,
} from '../api/platform'
import ImageUpload from '../components/ImageUpload.vue'

const list = ref<PlatformCategory[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<PlatformCategory | null>(null)
const form = ref({
  parent_id: 0,
  name: '',
  icon: '',
  sort: 0,
  status: 1,
})

const flatList = computed(() => {
  const out: PlatformCategory[] = []
  const walk = (nodes: PlatformCategory[], depth = 0) => {
    nodes.forEach((n) => {
      out.push({ ...n, name: `${'　'.repeat(depth)}${n.name}` })
      if (n.children?.length) walk(n.children, depth + 1)
    })
  }
  walk(list.value)
  return out
})

async function loadList() {
  loading.value = true
  try {
    list.value = await fetchPlatformCategories()
  } finally {
    loading.value = false
  }
}

function openCreate(parentId = 0) {
  editing.value = null
  form.value = { parent_id: parentId, name: '', icon: '', sort: 0, status: 1 }
  dialogVisible.value = true
}

function openEdit(row: PlatformCategory) {
  editing.value = row
  form.value = {
    parent_id: row.parent_id,
    name: row.name.trim(),
    icon: row.icon,
    sort: row.sort,
    status: row.status,
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.name.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  if (editing.value) {
    await updatePlatformCategory(editing.value.id, form.value)
    ElMessage.success('已更新')
  } else {
    await createPlatformCategory(form.value)
    ElMessage.success('已创建')
  }
  dialogVisible.value = false
  loadList()
}

async function handleDelete(row: PlatformCategory) {
  await ElMessageBox.confirm(`确定删除分类「${row.name}」？`, '提示', { type: 'warning' })
  await deletePlatformCategory(row.id)
  ElMessage.success('已删除')
  loadList()
}

onMounted(loadList)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>平台分类</h2>
      <el-button type="primary" :icon="Plus" @click="openCreate(0)">新增大类</el-button>
    </div>
    <el-table v-loading="loading" :data="flatList" row-key="id" border>
      <el-table-column prop="name" label="名称" min-width="180" />
      <el-table-column label="图标" width="80">
        <template #default="{ row }">
          <el-image v-if="row.icon" :src="row.icon" style="width:36px;height:36px" fit="cover" />
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openCreate(row.id)">加子类</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑分类' : '新增分类'" width="480px">
      <el-form label-width="80px">
        <el-form-item label="父分类">
          <el-select v-model="form.parent_id" style="width:100%">
            <el-option :value="0" label="无（一级分类）" />
            <el-option
              v-for="c in flatList.filter((x) => x.parent_id === 0)"
              :key="c.id"
              :value="c.id"
              :label="c.name.trim()"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="图标">
          <ImageUpload v-model="form.icon" size="small" />
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
.toolbar h2 {
  margin: 0;
  font-size: 18px;
}
</style>
