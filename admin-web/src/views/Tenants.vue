<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createTenant, fetchTenants, type Tenant } from '../api/tenant'
import { formatDateTime } from '../utils/format'

const list = ref<Tenant[]>([])
const dialogVisible = ref(false)
const saving = ref(false)
const form = ref({ name: '', code: '' })

async function loadList() {
  list.value = await fetchTenants()
}

function openCreate() {
  form.value = { name: '', code: '' }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.value.name.trim() || !form.value.code.trim()) {
    ElMessage.warning('请填写租户名称和编码')
    return
  }
  saving.value = true
  try {
    await createTenant({ name: form.value.name.trim(), code: form.value.code.trim().toLowerCase() })
    ElMessage.success('创建成功')
    dialogVisible.value = false
    loadList()
  } finally {
    saving.value = false
  }
}

onMounted(loadList)
</script>

<template>
  <el-card shadow="never">
    <div class="toolbar">
      <div>
        <h3>租户管理</h3>
        <p class="hint">平台管理员可在此创建与管理租户；各租户数据相互隔离</p>
      </div>
      <el-button type="primary" @click="openCreate">新建租户</el-button>
    </div>
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="租户名称" min-width="160" />
      <el-table-column prop="code" label="编码" width="140" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialogVisible" title="新建租户" width="480px">
    <el-form label-width="90px">
      <el-form-item label="租户名称" required>
        <el-input v-model="form.name" placeholder="如：华东旗舰店" maxlength="32" />
      </el-form-item>
      <el-form-item label="租户编码" required>
        <el-input v-model="form.code" placeholder="小写英文，如 east_store" maxlength="32" />
        <div class="form-tip">用于 API 标识，创建后不可随意修改</div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">创建</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}
.toolbar h3 { margin: 0 0 4px; }
.hint { margin: 0; font-size: 13px; color: #999; }
.form-tip { margin-top: 4px; font-size: 12px; color: #999; }
</style>
