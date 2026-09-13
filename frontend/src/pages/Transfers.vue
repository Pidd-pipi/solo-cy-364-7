<template>
  <div>
    <el-card>
      <div class="toolbar">
        <el-select v-model="statusFilter" placeholder="状态筛选" clearable style="width: 160px" @change="load">
          <el-option v-for="s in TRANSFER_STATUS_OPTIONS" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-button type="primary" @click="load">查询</el-button>
        <div class="spacer"></div>
        <el-button type="primary" @click="openCreate">发起调拨</el-button>
      </div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="调出门店" min-width="140">
          <template #default="{ row }">{{ row.from_store?.name || `#${row.from_store_id}` }}</template>
        </el-table-column>
        <el-table-column label="调入门店" min-width="140">
          <template #default="{ row }">{{ row.to_store?.name || `#${row.to_store_id}` }}</template>
        </el-table-column>
        <el-table-column label="商品" min-width="150">
          <template #default="{ row }">{{ row.sku?.name || `#${row.sku_id}` }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="90" />
        <el-table-column prop="reason" label="原因" min-width="120" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><TransferStatusBadge :status="row.status" /></template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canConfirm(row)" link type="primary" size="small" @click="doConfirm(row)">确认</el-button>
            <el-button v-if="canShip(row)" link type="warning" size="small" @click="doShip(row)">发货</el-button>
            <el-button v-if="canReceive(row)" link type="success" size="small" @click="doReceive(row)">收货</el-button>
            <el-button v-if="canCancel(row)" link type="danger" size="small" @click="doCancel(row)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="mt" layout="total, prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="load" />
    </el-card>

    <el-dialog v-model="createVisible" title="发起调拨申请" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="调出门店" required>
          <el-select v-model="form.from_store_id" style="width: 100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="调入门店" required>
          <el-select v-model="form.to_store_id" style="width: 100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品" required>
          <el-select v-model="form.sku_id" style="width: 100%">
            <el-option v-for="s in skus" :key="s.id" :label="`${s.name}（${s.code}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量" required>
          <el-input-number v-model="form.quantity" :min="1" />
        </el-form-item>
        <el-form-item label="原因">
          <el-input v-model="form.reason" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onCreate">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import TransferStatusBadge from '@/components/common/TransferStatusBadge.vue'
import { listTransfers, createTransfer, confirmTransfer, shipTransfer, receiveTransfer, cancelTransfer } from '@/api/transferOrder'
import { listAllStores } from '@/api/store'
import { listSkus } from '@/api/sku'
import { TRANSFER_STATUS_OPTIONS, TransferStatus, canTransfer, type TransferStatusValue } from '@/constants/transfer'
import { useAuthStore } from '@/stores/authStore'
import type { Store, SKU, TransferOrder } from '@/types'

const list = ref<TransferOrder[]>([])
const stores = ref<Store[]>([])
const skus = ref<SKU[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const statusFilter = ref<TransferStatusValue | ''>('')
const loading = ref(false)
const saving = ref(false)
const createVisible = ref(false)
const auth = useAuthStore()
const form = reactive({ from_store_id: 0, to_store_id: 0, sku_id: 0, quantity: 1, reason: '' })

async function load() {
  loading.value = true
  try {
    const params: { page: number; page_size: number; status?: TransferStatusValue } = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.status = statusFilter.value
    const res = await listTransfers(params)
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.from_store_id = stores.value[0]?.id || 0
  form.to_store_id = stores.value[1]?.id || 0
  form.sku_id = skus.value[0]?.id || 0
  form.quantity = 1
  form.reason = ''
  createVisible.value = true
}

async function onCreate() {
  saving.value = true
  try {
    await createTransfer(form)
    ElMessage.success('调拨申请已提交')
    createVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

function canConfirm(row: TransferOrder) {
  return (auth.role === 'admin' || auth.role === 'hq') && canTransfer(row.status, TransferStatus.CONFIRMED)
}
function canShip(row: TransferOrder) {
  return canTransfer(row.status, TransferStatus.SHIPPED)
}
function canReceive(row: TransferOrder) {
  return canTransfer(row.status, TransferStatus.RECEIVED)
}
function canCancel(row: TransferOrder) {
  return canTransfer(row.status, TransferStatus.CANCELLED)
}

async function doConfirm(row: TransferOrder) { await confirmTransfer(row.id); ElMessage.success('已确认'); await load() }
async function doShip(row: TransferOrder) { await shipTransfer(row.id); ElMessage.success('已发货'); await load() }
async function doReceive(row: TransferOrder) { await receiveTransfer(row.id); ElMessage.success('已收货'); await load() }
async function doCancel(row: TransferOrder) { await cancelTransfer(row.id); ElMessage.success('已取消'); await load() }

onMounted(async () => {
  stores.value = await listAllStores()
  const res = await listSkus({ page: 1, page_size: 200 })
  skus.value = res.list
  await load()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 10px; margin-bottom: 14px; align-items: center; }
.spacer { flex: 1; }
.mt { margin-top: 14px; }
</style>
