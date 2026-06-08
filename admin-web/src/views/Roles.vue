<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createRole,
  deleteRole,
  fetchMenus,
  fetchRoleMenus,
  fetchRoles,
  setRoleMenus,
  updateRole,
  type AdminMenu,
  type AdminRole,
} from '../api/rbac'

const list = ref<AdminRole[]>([])
const allMenus = ref<AdminMenu[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const menuDialogVisible = ref(false)
const editing = ref<AdminRole | null>(null)
const currentRole = ref<AdminRole | null>(null)
const selectedMenuIds = ref<number[]>([])
const form = ref({ code: '', name: '', scope: 'platform', tenant_id: 0 })

async function load() {
  loading.value = true
  try {
    list.value = await fetchRoles()
    allMenus.value = await fetchMenus()
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.value = { code: '', name: '', scope: 'platform', tenant_id: 0 }
  dialogVisible.value = true
}

async function submit() {
  if (editing.value) {
    await updateRole(editing.value.id, { name: form.value.name })
    ElMessage.success('已更新')
  } else {
    await createRole(form.value)
    ElMessage.success('已创建')
  }
  dialogVisible.value = false
  load()
}

async function remove(row: AdminRole) {
  await ElMessageBox.confirm(`确定删除角色「${row.name}」？`, '提示')
  await deleteRole(row.id)
  ElMessage.success('已删除')
  load()
}

async function openMenus(row: AdminRole) {
  currentRole.value = row
  selectedMenuIds.value = await fetchRoleMenus(row.id)
  menuDialogVisible.value = true
}

async function saveMenus() {
  if (!currentRole.value) return
  await setRoleMenus(currentRole.value.id, selectedMenuIds.value)
  ElMessage.success('菜单权限已保存')
  menuDialogVisible.value = false
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h3>角色管理</h3>
      <el-button type="primary" @click="openCreate">新建角色</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border>
      <el-table-column prop="code" label="编码" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="scope" label="范围" width="100" />
      <el-table-column prop="tenant_id" label="租户ID" width="90" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" @click="openMenus(row)">菜单授权</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="新建角色" width="420px">
      <el-form label-width="80px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="范围">
          <el-select v-model="form.scope" style="width: 100%">
            <el-option label="平台" value="platform" />
            <el-option label="租户" value="tenant" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="menuDialogVisible" title="菜单授权" width="480px">
      <el-checkbox-group v-model="selectedMenuIds">
        <div v-for="m in allMenus" :key="m.id" class="menu-check">
          <el-checkbox :label="m.id">{{ m.title }} ({{ m.path }})</el-checkbox>
        </div>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="menuDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveMenus">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.menu-check { margin-bottom: 8px; }
</style>
