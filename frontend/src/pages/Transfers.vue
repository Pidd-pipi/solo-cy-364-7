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
        <el-table-column prop="quantity" label="数量" width="80" />
        <el-table-column label="收货进度" width="180">
          <template #default="{ row }">
            <span>已收 {{ row.received_quantity }} / {{ row.quantity }}</span>
            <span v-if="progressOf(row).showRemaining" class="remaining">（剩余 {{ progressOf(row).remaining }}）</span>
            <span v-else-if="progressOf(row).showReturned" class="returned">（未收 {{ progressOf(row).remaining }} 已退回）</span>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="120" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><TransferStatusBadge :status="row.status" /></template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canConfirm(row)" link type="primary" size="small" @click="doConfirm(row)">确认</el-button>
            <el-button v-if="canShip(row)" link type="warning" size="small" @click="doShip(row)">发货</el-button>
            <el-button v-if="canReceive(row)" link type="success" size="small" @click="openReceive(row)">收货</el-button>
            <el-button v-if="row.received_quantity > 0" link type="info" size="small" @click="openReceive(row, true)">明细</el-button>
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

    <el-dialog v-model="receiveVisible" :title="receiveReadonly ? '收货明细' : '分批收货'" width="640px">
      <template v-if="receiveOrder">
        <el-descriptions :column="3" border size="small" class="mb">
          <el-descriptions-item label="调拨单号">#{{ receiveOrder.id }}</el-descriptions-item>
          <el-descriptions-item label="商品">{{ receiveOrder.sku?.name || `#${receiveOrder.sku_id}` }}</el-descriptions-item>
          <el-descriptions-item label="调入门店">{{ receiveOrder.to_store?.name || `#${receiveOrder.to_store_id}` }}</el-descriptions-item>
          <el-descriptions-item label="调拨数量">{{ receiveOrder.quantity }}</el-descriptions-item>
          <el-descriptions-item label="累计收货">{{ receiveOrder.received_quantity }}</el-descriptions-item>
          <el-descriptions-item label="剩余数量">{{ remainingOf(receiveOrder) }}</el-descriptions-item>
        </el-descriptions>

        <el-form v-if="!receiveReadonly" :model="receiveForm" label-width="90px" class="mb">
          <el-form-item label="实收数量" required>
            <el-input-number v-model="receiveForm.quantity" :min="1" :max="remainingOf(receiveOrder)" :disabled="remainingOf(receiveOrder) <= 0" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="receiveForm.remark" type="textarea" :rows="2" maxlength="255" placeholder="本次收货备注" />
          </el-form-item>
        </el-form>

        <div class="receipt-title">收货明细</div>
        <el-table :data="receipts" v-loading="receiptsLoading" border size="small" max-height="260">
          <el-table-column type="index" label="#" width="50" />
          <el-table-column prop="quantity" label="实收数量" width="100" />
          <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">{{ row.remark || '-' }}</template>
          </el-table-column>
          <el-table-column label="收货时间" width="170">
            <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
          </el-table-column>
        </el-table>
      </template>
      <template #footer>
        <el-button @click="receiveVisible = false">关闭</el-button>
        <el-button v-if="!receiveReadonly" type="primary" :loading="receiving" :disabled="!receiveOrder || remainingOf(receiveOrder) <= 0" @click="onReceive">确认收货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import TransferStatusBadge from '@/components/common/TransferStatusBadge.vue'
import { listTransfers, createTransfer, confirmTransfer, shipTransfer, receiveTransfer, listTransferReceipts, cancelTransfer } from '@/api/transferOrder'
import { listAllStores } from '@/api/store'
import { listSkus } from '@/api/sku'
import { TRANSFER_STATUS_OPTIONS, TransferStatus, canTransfer, type TransferStatusValue } from '@/constants/transfer'
import { useAuthStore } from '@/stores/authStore'
import { formatDateTime } from '@/utils/dateFormat'
import { transferProgress, transferRemaining } from '@/utils/transferProgress'
import type { Store, SKU, TransferOrder, TransferReceipt } from '@/types'

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

const receiveVisible = ref(false)
const receiveReadonly = ref(false)
const receiveOrder = ref<TransferOrder | null>(null)
const receiveForm = reactive({ quantity: 1, remark: '' })
const receipts = ref<TransferReceipt[]>([])
const receiptsLoading = ref(false)
const receiving = ref(false)

function remainingOf(row: TransferOrder) {
  return transferRemaining(row.quantity, row.received_quantity)
}

function progressOf(row: TransferOrder) {
  return transferProgress(row.status, row.quantity, row.received_quantity)
}

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
async function doCancel(row: TransferOrder) {
  const tip = row.status === TransferStatus.SHIPPED
    ? `该单已发货（已收 ${row.received_quantity} 件），取消后未收部分（${remainingOf(row)} 件）将退回调出门店，已收部分留在调入门店。确认取消？`
    : '确认取消该调拨单？'
  await ElMessageBox.confirm(tip, '取消调拨单', { type: 'warning' })
  await cancelTransfer(row.id)
  ElMessage.success('已取消')
  await load()
}

async function loadReceipts(orderId: number) {
  receiptsLoading.value = true
  try {
    const res = await listTransferReceipts(orderId)
    receipts.value = res.list
  } finally {
    receiptsLoading.value = false
  }
}

async function openReceive(row: TransferOrder, readonly = false) {
  receiveOrder.value = row
  receiveReadonly.value = readonly
  receiveForm.quantity = Math.max(1, remainingOf(row))
  receiveForm.remark = ''
  receipts.value = []
  receiveVisible.value = true
  await loadReceipts(row.id)
}

async function onReceive() {
  const order = receiveOrder.value
  if (!order || receiving.value) return
  if (!receiveForm.quantity || receiveForm.quantity < 1) {
    ElMessage.warning('请填写实收数量')
    return
  }
  if (receiveForm.quantity > remainingOf(order)) {
    ElMessage.warning('实收数量不能超过剩余数量')
    return
  }
  receiving.value = true
  try {
    const updated = await receiveTransfer(order.id, { quantity: receiveForm.quantity, remark: receiveForm.remark })
    receiveOrder.value = updated
    if (updated.status === TransferStatus.RECEIVED) {
      ElMessage.success('已收满，调拨单已收货')
    } else {
      ElMessage.success(`本次收货 ${receiveForm.quantity} 件，剩余 ${remainingOf(updated)} 件`)
    }
    receiveForm.quantity = Math.max(1, remainingOf(updated))
    receiveForm.remark = ''
    await Promise.all([load(), loadReceipts(order.id)])
  } finally {
    receiving.value = false
  }
}

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
.mb { margin-bottom: 14px; }
.remaining { color: #e6a23c; }
.returned { color: #909399; }
.receipt-title { font-weight: 600; margin-bottom: 8px; }
</style>
