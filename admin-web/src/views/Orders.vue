<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  centToYuan,
  expressCompanies,
  expressCompanyName,
  fetchAdminOrders,
  formatAddress,
  parseAddressSnapshot,
  shipOrder,
  statusMap,
  type Order,
  type ShipOrderPayload,
} from '../api/orders'
import { formatDateTime } from '../utils/format'

const list = ref<Order[]>([])
const total = ref(0)
const page = ref(1)
const statusFilter = ref<number | undefined>(1)
const shipVisible = ref(false)
const shipping = ref(false)
const currentOrder = ref<Order | null>(null)
const shipForm = ref<ShipOrderPayload>({
  express_company: 'SF',
  express_no: '',
  ship_remark: '',
})

const currentAddress = computed(() => {
  if (!currentOrder.value) return null
  return parseAddressSnapshot(currentOrder.value.address_snapshot)
})

async function loadList() {
  const data = await fetchAdminOrders({
    page: page.value,
    page_size: 10,
    status: statusFilter.value,
  })
  list.value = data.list || []
  total.value = data.total || 0
}

function openShip(row: Order) {
  currentOrder.value = row
  shipForm.value = { express_company: 'SF', express_no: '', ship_remark: '' }
  shipVisible.value = true
}

async function submitShip() {
  if (!currentOrder.value) return
  const no = shipForm.value.express_no.trim()
  if (!no) {
    ElMessage.warning('请填写物流单号')
    return
  }
  if (no.length < 6) {
    ElMessage.warning('物流单号至少 6 位')
    return
  }
  shipping.value = true
  try {
    await shipOrder(currentOrder.value.order_no, {
      express_company: shipForm.value.express_company,
      express_no: no,
      ship_remark: shipForm.value.ship_remark?.trim() || undefined,
    })
    ElMessage.success('发货成功')
    shipVisible.value = false
    loadList()
  } finally {
    shipping.value = false
  }
}

onMounted(loadList)
</script>

<template>
  <el-card shadow="never">
    <div class="toolbar">
      <h3>订单管理</h3>
      <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 140px" @change="loadList">
        <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
      </el-select>
      <el-button @click="loadList">刷新</el-button>
    </div>
    <el-table :data="list" stripe>
      <el-table-column prop="order_no" label="订单号" min-width="150" />
      <el-table-column label="下单时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="收货人" width="140">
        <template #default="{ row }">
          <template v-if="parseAddressSnapshot(row.address_snapshot)">
            <div>{{ parseAddressSnapshot(row.address_snapshot)?.name }}</div>
            <div class="sub">{{ parseAddressSnapshot(row.address_snapshot)?.phone }}</div>
          </template>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'warning' : row.status === 2 ? 'primary' : 'info'" size="small">
            {{ statusMap[row.status] || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="实付" width="90">
        <template #default="{ row }">¥{{ centToYuan(row.pay_amount || row.total_amount) }}</template>
      </el-table-column>
      <el-table-column label="商品" min-width="180">
        <template #default="{ row }">
          <div v-for="(it, i) in row.items" :key="i" class="goods-line">{{ it.product_name }} x{{ it.quantity }}</div>
        </template>
      </el-table-column>
      <el-table-column label="物流信息" min-width="180">
        <template #default="{ row }">
          <template v-if="row.express_no">
            <div>{{ expressCompanyName(row.express_company) }}</div>
            <div class="sub">{{ row.express_no }}</div>
          </template>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 1" link type="primary" @click="openShip(row)">发货</el-button>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-model:current-page="page" :total="total" layout="total, prev, pager, next" @current-change="loadList" />
  </el-card>

  <el-dialog v-model="shipVisible" title="订单发货" width="560px" destroy-on-close>
    <template v-if="currentOrder">
      <el-descriptions :column="1" border size="small" class="ship-desc">
        <el-descriptions-item label="订单号">{{ currentOrder.order_no }}</el-descriptions-item>
        <el-descriptions-item label="实付金额">¥{{ centToYuan(currentOrder.pay_amount || currentOrder.total_amount) }}</el-descriptions-item>
        <el-descriptions-item label="收货人">
          {{ currentAddress?.name || '-' }} {{ currentAddress?.phone || '' }}
        </el-descriptions-item>
        <el-descriptions-item label="收货地址">{{ formatAddress(currentAddress) }}</el-descriptions-item>
        <el-descriptions-item label="商品">
          <div v-for="(it, i) in currentOrder.items" :key="i">{{ it.product_name }}（{{ it.sku_name || '默认' }}）x{{ it.quantity }}</div>
        </el-descriptions-item>
      </el-descriptions>

      <el-form label-width="90px" class="ship-form">
        <el-form-item label="物流公司" required>
          <el-select v-model="shipForm.express_company" placeholder="请选择物流公司" style="width: 100%">
            <el-option v-for="c in expressCompanies" :key="c.code" :label="c.name" :value="c.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号" required>
          <el-input v-model="shipForm.express_no" placeholder="请输入快递单号" maxlength="64" clearable />
        </el-form-item>
        <el-form-item label="发货备注">
          <el-input v-model="shipForm.ship_remark" type="textarea" :rows="2" placeholder="选填，如分批发货说明" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>
    </template>
    <template #footer>
      <el-button @click="shipVisible = false">取消</el-button>
      <el-button type="primary" :loading="shipping" @click="submitShip">确认发货</el-button>
    </template>
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
.sub {
  font-size: 12px;
  color: #999;
}
.goods-line {
  font-size: 13px;
  line-height: 1.6;
}
.ship-desc {
  margin-bottom: 20px;
}
.ship-form {
  margin-top: 8px;
}
</style>
