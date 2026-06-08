<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createAdmin, fetchAdminRoles, fetchAdmins, fetchRoles, updateAdmin, type AdminAccount, type AdminRole } from '../api/rbac'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const list = ref<AdminAccount[]>([])
const roles = ref<AdminRole[]>([])
const total = ref(0)
const loading = ref(false)
const page = ref(1)
const dialogVisible = ref(false)
const form = ref({
  username: '',
  password: '',
  nickname: '',
  tenant_id: 0,
  user_type: 'tenant',
  role_ids: [] as number[],
})

async function load() {
  loading.value = true
  try {
    const data = await fetchAdmins({ page: page.value, page_size: 20 })
    list.value = data.list || []
    total.value = data.total || 0
    roles.value = await fetchRoles()
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = {
    username: '',
    password: '',
    nickname: '',
    tenant_id: auth.platformView ? 0 : auth.activeTenant?.id || 0,
    user_type: auth.platformView ? 'platform' : 'tenant',
    role_ids: [],
  }
  dialogVisible.value = true
}

async function submit() {
  await createAdmin(form.value)
  ElMessage.success('账号已创建')
  dialogVisible.value = false
  load()
}

async function toggleStatus(row: AdminAccount) {
  const status = row.status === 1 ? 0 : 1
  await updateAdmin(row.id, { status })
  ElMessage.success('状态已更新')
  load()
}

async function showRoles(row: AdminAccount) {
  const ids = await fetchAdminRoles(row.id)
  ElMessage.info(`角色 ID: ${ids.join(', ') || '无'}`)
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h3>后台账号</h3>
      <el-button type="primary" @click="openCreate">新建账号</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="user_type" label="类型" width="100" />
      <el-table-column prop="tenant_id" label="租户ID" width="90" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button link type="primary" @click="showRoles(row)">角色</el-button>
          <el-button link @click="toggleStatus(row)">{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" title="新建账号" width="480px">
      <el-form label-width="90px">
        <el-form-item label="用户名"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item v-if="auth.platformView" label="类型">
          <el-select v-model="form.user_type" style="width: 100%">
            <el-option label="平台" value="platform" />
            <el-option label="租户" value="tenant" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="auth.platformView && form.user_type === 'tenant'" label="租户ID">
          <el-input-number v-model="form.tenant_id" :min="1" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role_ids" multiple style="width: 100%">
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
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
