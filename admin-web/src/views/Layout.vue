<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MenuTreeItem from '../components/MenuTreeItem.vue'
import { fetchPublicTenants, type Tenant } from '../api/tenant'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const enterVisible = ref(false)
const tenants = ref<Tenant[]>([])
const enterTenantId = ref<number>()

const viewLabel = computed(() => {
  if (auth.platformView) return '平台视角'
  return auth.activeTenant?.name || '租户视角'
})

onMounted(async () => {
  if (!auth.admin) {
    await auth.restore()
  }
})

function logout() {
  auth.clearSession()
  router.push('/login')
}

async function openEnterTenant() {
  try {
    tenants.value = await fetchPublicTenants()
    enterTenantId.value = tenants.value[0]?.id
    enterVisible.value = true
  } catch {
    tenants.value = []
  }
}

async function confirmEnter() {
  if (!enterTenantId.value) return
  await auth.doEnterTenant(enterTenantId.value)
  enterVisible.value = false
  router.push('/dashboard')
}

async function handleExitTenant() {
  await auth.doExitTenant()
  router.push('/dashboard')
}
</script>

<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">搜店网</div>
      <el-menu
        router
        :default-active="route.path"
        class="side-menu"
        background-color="#001529"
        text-color="rgba(255, 255, 255, 0.78)"
        active-text-color="#ffffff"
      >
        <MenuTreeItem v-for="m in auth.menus" :key="m.id" :item="m" />
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <span>管理后台</span>
          <el-tag type="success" effect="plain" class="tenant-tag">{{ viewLabel }}</el-tag>
          <span v-if="auth.admin" class="user-name">{{ auth.admin.nickname || auth.admin.username }}</span>
        </div>
        <div class="header-right">
          <template v-if="auth.isPlatformAdmin && auth.platformView">
            <el-button link type="primary" @click="openEnterTenant">进入租户</el-button>
          </template>
          <template v-if="auth.isPlatformAdmin && !auth.platformView">
            <el-button link type="primary" @click="handleExitTenant">退出租户</el-button>
          </template>
          <el-button link type="danger" @click="logout">退出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>

    <el-dialog v-model="enterVisible" title="进入租户" width="400px">
      <el-select v-model="enterTenantId" placeholder="选择租户" style="width: 100%">
        <el-option v-for="t in tenants" :key="t.id" :label="`${t.name}（${t.code}）`" :value="t.id" />
      </el-select>
      <template #footer>
        <el-button @click="enterVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmEnter">进入</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<style scoped>
.layout { min-height: 100vh; }
.aside { background: #001529; }
.side-menu {
  border-right: none;
  background: #001529;
}
.side-menu :deep(.el-menu-item),
.side-menu :deep(.el-sub-menu__title) {
  color: rgba(255, 255, 255, 0.78);
}
.side-menu :deep(.el-menu-item:hover),
.side-menu :deep(.el-sub-menu__title:hover) {
  background-color: rgba(255, 255, 255, 0.08) !important;
  color: #fff;
}
.side-menu :deep(.el-menu-item.is-active) {
  background-color: #1677ff !important;
  color: #fff;
}
.side-menu :deep(.el-sub-menu .el-menu) {
  background-color: #000c17 !important;
}
.side-menu :deep(.el-sub-menu .el-menu-item) {
  background-color: transparent;
  color: rgba(255, 255, 255, 0.72);
}
.side-menu :deep(.el-sub-menu .el-menu-item.is-active) {
  background-color: #1677ff !important;
  color: #fff;
}
.side-menu :deep(.el-sub-menu__icon-arrow) {
  color: rgba(255, 255, 255, 0.55);
}
.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  font-weight: bold;
  color: #fff;
  background: #001529;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #eee;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tenant-tag { font-weight: 500; }
.user-name { color: #666; font-size: 13px; }
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
