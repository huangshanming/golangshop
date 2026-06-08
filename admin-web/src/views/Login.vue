<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import type { LoginCandidate } from '../api/auth'

const router = useRouter()
const auth = useAuthStore()
const form = ref({ username: 'superadmin', password: 'admin123' })
const loading = ref(false)
const pickVisible = ref(false)
const candidates = ref<LoginCandidate[]>([])
const pickTenantId = ref<number>()

async function handleLogin(tenantId?: number) {
  loading.value = true
  try {
    const result = await auth.login(form.value.username, form.value.password, tenantId)
    if (result.needTenantPick) {
      candidates.value = result.candidates || []
      pickTenantId.value = candidates.value[0]?.tenant_id
      pickVisible.value = true
      return
    }
    router.push('/')
  } finally {
    loading.value = false
  }
}

async function confirmPick() {
  if (pickTenantId.value == null) return
  pickVisible.value = false
  await handleLogin(pickTenantId.value)
}
</script>

<template>
  <div class="login-wrap">
    <el-card class="login-card">
      <h2>搜店网 管理后台</h2>
      <p class="subtitle">账号登录 · 系统自动识别所属租户</p>
      <el-form @submit.prevent="handleLogin()">
        <el-form-item label="用户名">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">
          登录
        </el-button>
      </el-form>
    </el-card>

    <el-dialog v-model="pickVisible" title="选择登录租户" width="400px">
      <p class="pick-tip">该账号存在于多个租户，请选择要进入的租户：</p>
      <el-radio-group v-model="pickTenantId" style="display: flex; flex-direction: column; gap: 12px">
        <el-radio v-for="c in candidates" :key="c.user_id" :value="c.tenant_id">
          {{ c.tenant_name }}
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="pickVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmPick">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}
.login-card {
  width: 420px;
}
h2 {
  text-align: center;
  margin-bottom: 8px;
}
.subtitle {
  text-align: center;
  color: #999;
  font-size: 13px;
  margin: 0 0 24px;
}
.pick-tip {
  margin: 0 0 16px;
  color: #666;
}
</style>
