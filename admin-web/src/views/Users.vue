<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adjustBalance, centToYuan, fetchBalanceLogs, fetchUsers, type AppUser, type BalanceLog } from '../api/users'
import { formatDateTime } from '../utils/format'

const bizTypeMap: Record<string, string> = {
  admin_adjust: '管理员调账',
  order_pay: '订单支付',
  order_refund: '订单退款',
  recharge: '充值',
}

const list = ref<AppUser[]>([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')
const logsVisible = ref(false)
const logs = ref<BalanceLog[]>([])
const currentUser = ref<AppUser | null>(null)
const adjustVisible = ref(false)
const adjustForm = ref({ changeYuan: 0, remark: '' })
const adjusting = ref(false)

async function loadList() {
  const data = await fetchUsers({ page: page.value, page_size: 10, keyword: keyword.value.trim() || undefined })
  list.value = data.list || []
  total.value = data.total || 0
}

async function openLogs(row: AppUser) {
  currentUser.value = row
  const data = await fetchBalanceLogs(row.id)
  logs.value = data.list || []
  logsVisible.value = true
}

function openAdjust(row: AppUser) {
  currentUser.value = row
  adjustForm.value = { changeYuan: 0, remark: '' }
  adjustVisible.value = true
}

async function submitAdjust() {
  if (!currentUser.value) return
  adjusting.value = true
  try {
    const change = Math.round(adjustForm.value.changeYuan * 100)
    await adjustBalance(currentUser.value.id, change, adjustForm.value.remark)
    ElMessage.success('调账成功')
    adjustVisible.value = false
    loadList()
  } finally {
    adjusting.value = false
  }
}

onMounted(loadList)
</script>

<template>
  <el-card shadow="never">
    <div class="toolbar">
      <h3>用户管理</h3>
      <el-input v-model="keyword" placeholder="搜索手机号/昵称" style="width: 220px" clearable @keyup.enter="loadList" />
      <el-button @click="loadList">搜索</el-button>
    </div>
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="phone" label="手机号" width="140" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column label="余额" width="120">
        <template #default="{ row }">¥{{ centToYuan(row.balance).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" @click="openAdjust(row)">调账</el-button>
          <el-button link @click="openLogs(row)">流水</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="page" :total="total" layout="total, prev, pager, next" @current-change="loadList" />
  </el-card>

  <el-dialog v-model="adjustVisible" title="余额调账" width="400px">
    <el-form label-width="80px">
      <el-form-item label="用户">{{ currentUser?.phone }}</el-form-item>
      <el-form-item label="变动(元)">
        <el-input-number v-model="adjustForm.changeYuan" :precision="2" :step="10" />
        <div class="hint">正数增加，负数扣减</div>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="adjustForm.remark" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="adjustVisible = false">取消</el-button>
      <el-button type="primary" :loading="adjusting" @click="submitAdjust">确定</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="logsVisible" title="余额流水" width="680px">
    <el-table :data="logs" size="small">
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="变动" width="100">
        <template #default="{ row }">
          <span :class="row.change_amount >= 0 ? 'plus' : 'minus'">
            {{ row.change_amount >= 0 ? '+' : '' }}{{ centToYuan(row.change_amount).toFixed(2) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="余额" width="100">
        <template #default="{ row }">¥{{ centToYuan(row.balance_after).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">{{ bizTypeMap[row.biz_type] || row.biz_type }}</template>
      </el-table-column>
      <el-table-column label="说明" min-width="180">
        <template #default="{ row }">
          {{ row.goods_summary || row.remark || row.biz_no || '-' }}
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.toolbar h3 { margin: 0; }
.hint { font-size: 12px; color: #999; margin-top: 4px; }
.plus { color: #07c160; }
.minus { color: #ff4d4f; }
</style>
