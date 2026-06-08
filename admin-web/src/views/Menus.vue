<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createMenu, deleteMenu, fetchMenus, updateMenu, type AdminMenu } from '../api/rbac'

const list = ref<AdminMenu[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<AdminMenu | null>(null)
const form = ref({
  parent_id: 0,
  title: '',
  path: '',
  icon: '',
  sort: 0,
  scope: 'platform',
  permission: '',
  status: 1,
})

const treeData = computed(() => {
  const byID = new Map<number, AdminMenu & { children?: AdminMenu[] }>()
  list.value.forEach((m) => byID.set(m.id, { ...m, children: [] }))
  const roots: (AdminMenu & { children?: AdminMenu[] })[] = []
  byID.forEach((node) => {
    if (node.parent_id && byID.has(node.parent_id)) {
      byID.get(node.parent_id)!.children!.push(node)
    } else {
      roots.push(node)
    }
  })
  const sortNodes = (nodes: (AdminMenu & { children?: AdminMenu[] })[]) => {
    nodes.sort((a, b) => a.sort - b.sort || a.id - b.id)
    nodes.forEach((n) => n.children?.length && sortNodes(n.children))
  }
  sortNodes(roots)
  return roots
})

const parentOptions = computed(() =>
  list.value.filter((m) => !m.path).map((m) => ({ id: m.id, title: m.title }))
)

async function load() {
  loading.value = true
  try {
    list.value = await fetchMenus()
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.value = { parent_id: 0, title: '', path: '', icon: '', sort: 0, scope: 'platform', permission: '', status: 1 }
  dialogVisible.value = true
}

function openEdit(row: AdminMenu) {
  editing.value = row
  form.value = {
    parent_id: row.parent_id,
    title: row.title,
    path: row.path,
    icon: row.icon,
    sort: row.sort,
    scope: row.scope,
    permission: row.permission,
    status: row.status,
  }
  dialogVisible.value = true
}

async function submit() {
  if (editing.value) {
    await updateMenu(editing.value.id, form.value)
    ElMessage.success('已更新')
  } else {
    await createMenu(form.value)
    ElMessage.success('已创建')
  }
  dialogVisible.value = false
  load()
}

async function remove(row: AdminMenu) {
  await ElMessageBox.confirm(`确定删除菜单「${row.title}」？`, '提示')
  await deleteMenu(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h3>菜单管理</h3>
      <el-button type="primary" @click="openCreate">新建菜单</el-button>
    </div>
    <el-table
      v-loading="loading"
      :data="treeData"
      row-key="id"
      border
      default-expand-all
      :tree-props="{ children: 'children' }"
    >
      <el-table-column prop="title" label="标题" min-width="160" />
      <el-table-column prop="path" label="路径" min-width="140" />
      <el-table-column prop="scope" label="范围" width="90" />
      <el-table-column prop="permission" label="权限码" min-width="140" />
      <el-table-column prop="sort" label="排序" width="70" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑菜单' : '新建菜单'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="父级菜单">
          <el-select v-model="form.parent_id" clearable placeholder="无（顶级菜单）" style="width: 100%">
            <el-option :value="0" label="无（顶级菜单）" />
            <el-option v-for="p in parentOptions" :key="p.id" :value="p.id" :label="p.title" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="路径">
          <el-input v-model="form.path" placeholder="父级菜单可留空" />
        </el-form-item>
        <el-form-item label="范围">
          <el-select v-model="form.scope" style="width: 100%">
            <el-option label="平台" value="platform" />
            <el-option label="租户" value="tenant" />
            <el-option label="通用" value="all" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限码"><el-input v-model="form.permission" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
