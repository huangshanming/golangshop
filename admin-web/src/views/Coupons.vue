<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  centToYuan,
  couponStatusLabel,
  createCoupon,
  fetchCoupons,
  formatCouponDiscount,
  formatThreshold,
  formatValidRange,
  updateCoupon,
  updateCouponStatus,
  yuanToCent,
  type Coupon,
  type CouponFormData,
} from '../api/coupons'

const list = ref<Coupon[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref(0)

const form = reactive({
  name: '',
  type: 'fixed' as 'fixed' | 'percent',
  threshold_yuan: 0,
  discount_yuan: 10,
  discount_percent: 90,
  max_discount_yuan: 0,
  valid_range: [] as Date[],
  total_count: 0,
  per_user_limit: 1,
  status: 1,
})

function resetForm() {
  editingId.value = 0
  form.name = ''
  form.type = 'fixed'
  form.threshold_yuan = 99
  form.discount_yuan = 10
  form.discount_percent = 90
  form.max_discount_yuan = 50
  const now = new Date()
  form.valid_range = [now, new Date(now.getTime() + 365 * 86400000)]
  form.total_count = 1000
  form.per_user_limit = 3
  form.status = 1
}

async function loadList() {
  const data = await fetchCoupons({
    page: page.value,
    page_size: pageSize.value,
    keyword: keyword.value.trim() || undefined,
  })
  list.value = data.list || []
  total.value = data.total || 0
}

function openCreate() {
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: Coupon) {
  editingId.value = row.id
  form.name = row.name
  form.type = row.type
  form.threshold_yuan = centToYuan(row.threshold_amount)
  if (row.type === 'fixed') {
    form.discount_yuan = centToYuan(row.discount_value)
  } else {
    form.discount_percent = 100 - row.discount_value
    form.max_discount_yuan = centToYuan(row.max_discount)
  }
  form.valid_range = [new Date(row.valid_from * 1000), new Date(row.valid_to * 1000)]
  form.total_count = row.total_count
  form.per_user_limit = row.per_user_limit
  form.status = row.status
  dialogVisible.value = true
}

function buildPayload(): CouponFormData {
  const [from, to] = form.valid_range
  return {
    name: form.name.trim(),
    type: form.type,
    threshold_amount: yuanToCent(form.threshold_yuan),
    discount_value: form.type === 'fixed' ? yuanToCent(form.discount_yuan) : 100 - form.discount_percent,
    max_discount: form.type === 'percent' ? yuanToCent(form.max_discount_yuan) : 0,
    valid_from: Math.floor(from.getTime() / 1000),
    valid_to: Math.floor(to.getTime() / 1000),
    total_count: form.total_count,
    per_user_limit: form.per_user_limit,
    status: form.status,
  }
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入优惠券名称')
    return
  }
  if (!form.valid_range || form.valid_range.length !== 2) {
    ElMessage.warning('请选择有效期')
    return
  }
  saving.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updateCoupon(editingId.value, payload)
      ElMessage.success('更新成功')
    } else {
      await createCoupon(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadList()
  } finally {
    saving.value = false
  }
}

async function toggleStatus(row: Coupon) {
  const next = row.status === 1 ? 0 : 1
  const label = couponStatusLabel(next)
  await ElMessageBox.confirm(`确定${label}该优惠券吗？`, '提示')
  await updateCouponStatus(row.id, next)
  ElMessage.success(`${label}成功`)
  loadList()
}

function onPageChange(p: number) {
  page.value = p
  loadList()
}

onMounted(loadList)
</script>

<template>
  <el-card shadow="never">
    <div class="toolbar">
      <div class="toolbar-left">
        <h3>优惠券管理</h3>
        <el-input
          v-model="keyword"
          placeholder="搜索优惠券名称"
          clearable
          style="width: 220px; margin-left: 16px"
          @keyup.enter="loadList"
          @clear="loadList"
        />
        <el-button style="margin-left: 8px" @click="loadList">搜索</el-button>
      </div>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增优惠券</el-button>
    </div>

    <el-table :data="list" stripe style="width: 100%">
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="row.type === 'fixed' ? 'success' : 'warning'">
            {{ row.type === 'fixed' ? '满减券' : '折扣券' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="门槛" width="120">
        <template #default="{ row }">{{ formatThreshold(row) }}</template>
      </el-table-column>
      <el-table-column label="优惠" width="160">
        <template #default="{ row }">{{ formatCouponDiscount(row) }}</template>
      </el-table-column>
      <el-table-column label="领取/总量" width="110">
        <template #default="{ row }">
          {{ row.received_count }} / {{ row.total_count === 0 ? '不限' : row.total_count }}
        </template>
      </el-table-column>
      <el-table-column prop="per_user_limit" label="每人限领" width="90" />
      <el-table-column label="有效期" min-width="200">
        <template #default="{ row }">
          <span class="valid-text">{{ formatValidRange(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ couponStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link :type="row.status === 1 ? 'warning' : 'success'" @click="toggleStatus(row)">
            {{ row.status === 1 ? '停用' : '启用' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="onPageChange"
      />
    </div>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑优惠券' : '新增优惠券'" width="560px">
      <el-form label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="如：满99减10" maxlength="32" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-radio-group v-model="form.type">
            <el-radio value="fixed">满减券</el-radio>
            <el-radio value="percent">折扣券</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="使用门槛">
          <el-input-number v-model="form.threshold_yuan" :min="0" :precision="2" :step="1" />
          <span class="form-tip">元（0 表示无门槛）</span>
        </el-form-item>
        <el-form-item v-if="form.type === 'fixed'" label="减免金额">
          <el-input-number v-model="form.discount_yuan" :min="0.01" :precision="2" :step="1" />
          <span class="form-tip">元</span>
        </el-form-item>
        <template v-else>
          <el-form-item label="折扣">
            <el-input-number v-model="form.discount_percent" :min="1" :max="99" :step="1" />
            <span class="form-tip">折（如 90 表示 9 折）</span>
          </el-form-item>
          <el-form-item label="最高减免">
            <el-input-number v-model="form.max_discount_yuan" :min="0" :precision="2" :step="1" />
            <span class="form-tip">元（0 表示不限）</span>
          </el-form-item>
        </template>
        <el-form-item label="有效期" required>
          <el-date-picker
            v-model="form.valid_range"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="发放总量">
          <el-input-number v-model="form.total_count" :min="0" :step="100" />
          <span class="form-tip">0 表示不限量</span>
        </el-form-item>
        <el-form-item label="每人限领">
          <el-input-number v-model="form.per_user_limit" :min="1" :max="99" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar-left {
  display: flex;
  align-items: center;
}
.toolbar-left h3 {
  margin: 0;
}
.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.form-tip {
  margin-left: 8px;
  color: #999;
  font-size: 12px;
}
.valid-text {
  font-size: 12px;
  color: #666;
}
</style>
