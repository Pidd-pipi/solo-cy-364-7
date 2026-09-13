<template>
  <div>
    <el-tabs v-model="tab">
      <el-tab-pane label="出入库明细" name="records">
        <el-card>
          <div class="toolbar">
            <el-select v-model="storeId" placeholder="按门店筛选" clearable style="width: 200px" @change="loadRecords">
              <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
            </el-select>
            <el-select v-model="recordType" placeholder="类型筛选" clearable style="width: 150px" @change="loadRecords">
              <el-option v-for="t in STOCK_RECORD_TYPE_OPTIONS" :key="t.value" :label="t.label" :value="t.value" />
            </el-select>
            <div class="spacer"></div>
            <el-button type="primary" @click="openCreateRecord">新增出入库</el-button>
            <el-button type="success" @click="onExport">导出</el-button>
          </div>
          <RecordTable :records="records" :loading="recordsLoading" />
          <el-pagination class="mt" layout="total, prev, pager, next" :total="recordsTotal" :page-size="pageSize" v-model:current-page="page" @current-change="loadRecords" />
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="周期盘点" name="stocktakes">
        <el-card>
          <div class="toolbar">
            <div class="spacer"></div>
            <el-button type="primary" @click="openStocktake">新增盘点</el-button>
          </div>
          <el-table :data="stocktakes" v-loading="stLoading" border stripe>
            <el-table-column label="门店" min-width="140">
              <template #default="{ row }">{{ row.store?.name || `#${row.store_id}` }}</template>
            </el-table-column>
            <el-table-column label="商品" min-width="150">
              <template #default="{ row }">{{ row.sku?.name || `#${row.sku_id}` }}</template>
            </el-table-column>
            <el-table-column prop="stocktake_date" label="盘点日期" width="120" />
            <el-table-column prop="system_qty" label="系统数量" width="100" />
            <el-table-column prop="actual_qty" label="实盘数量" width="100" />
            <el-table-column label="差异" width="100">
              <template #default="{ row }">
                <el-tag :type="row.difference > 0 ? 'success' : row.difference < 0 ? 'danger' : 'info'" size="small">
                  {{ row.difference > 0 ? '+' : '' }}{{ row.difference }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="120" />
          </el-table>
          <el-pagination class="mt" layout="total, prev, pager, next" :total="stTotal" :page-size="pageSize" v-model:current-page="stPage" @current-change="loadStocktakes" />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="recordVisible" title="新增出入库记录" width="440px">
      <el-form :model="recordForm" label-width="90px">
        <el-form-item label="门店" required>
          <el-select v-model="recordForm.store_id" style="width: 100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品" required>
          <el-select v-model="recordForm.sku_id" style="width: 100%">
            <el-option v-for="s in skus" :key="s.id" :label="`${s.name}（${s.code}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="recordForm.record_type" style="width: 100%">
            <el-option v-for="t in STOCK_RECORD_TYPE_OPTIONS" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量" required>
          <el-input-number v-model="recordForm.quantity" :min="1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recordVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onCreateRecord">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="stVisible" title="新增周期盘点" width="440px">
      <el-form :model="stForm" label-width="90px">
        <el-form-item label="门店" required>
          <el-select v-model="stForm.store_id" style="width: 100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品" required>
          <el-select v-model="stForm.sku_id" style="width: 100%">
            <el-option v-for="s in skus" :key="s.id" :label="`${s.name}（${s.code}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="盘点日期" required>
          <el-date-picker v-model="stForm.stocktake_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="实盘数量" required>
          <el-input-number v-model="stForm.actual_qty" :min="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="stForm.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onCreateStocktake">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import RecordTable from '@/components/common/RecordTable.vue'
import { listStockRecords, createStockRecord, exportStockRecords, createStocktake, listStocktakes } from '@/api/stockRecord'
import { listAllStores } from '@/api/store'
import { listSkus } from '@/api/sku'
import { STOCK_RECORD_TYPE_OPTIONS, type StockRecordTypeValue } from '@/constants/stockRecord'
import type { Store, SKU, StockRecord, Stocktake } from '@/types'

const tab = ref('records')
const stores = ref<Store[]>([])
const skus = ref<SKU[]>([])
const records = ref<StockRecord[]>([])
const recordsTotal = ref(0)
const recordsLoading = ref(false)
const stocktakes = ref<Stocktake[]>([])
const stTotal = ref(0)
const stLoading = ref(false)
const page = ref(1)
const stPage = ref(1)
const pageSize = ref(10)
const storeId = ref<number>()
const recordType = ref<StockRecordTypeValue | ''>('')
const recordVisible = ref(false)
const stVisible = ref(false)
const saving = ref(false)
const recordForm = reactive({ store_id: 0, sku_id: 0, record_type: 'purchase' as StockRecordTypeValue, quantity: 1 })
const stForm = reactive({ store_id: 0, sku_id: 0, stocktake_date: '', actual_qty: 0, remark: '' })

async function loadRecords() {
  recordsLoading.value = true
  try {
    const res = await listStockRecords({ page: page.value, page_size: pageSize.value, store_id: storeId.value, record_type: recordType.value || undefined })
    records.value = res.list
    recordsTotal.value = res.total
  } finally {
    recordsLoading.value = false
  }
}

async function loadStocktakes() {
  stLoading.value = true
  try {
    const res = await listStocktakes({ page: stPage.value, page_size: pageSize.value })
    stocktakes.value = res.list
    stTotal.value = res.total
  } finally {
    stLoading.value = false
  }
}

function openCreateRecord() {
  recordForm.store_id = stores.value[0]?.id || 0
  recordForm.sku_id = skus.value[0]?.id || 0
  recordForm.record_type = 'purchase'
  recordForm.quantity = 1
  recordVisible.value = true
}

async function onCreateRecord() {
  saving.value = true
  try {
    await createStockRecord(recordForm)
    ElMessage.success('记录已创建')
    recordVisible.value = false
    await loadRecords()
  } finally {
    saving.value = false
  }
}

function openStocktake() {
  stForm.store_id = stores.value[0]?.id || 0
  stForm.sku_id = skus.value[0]?.id || 0
  stForm.stocktake_date = new Date().toISOString().slice(0, 10)
  stForm.actual_qty = 0
  stForm.remark = ''
  stVisible.value = true
}

async function onCreateStocktake() {
  saving.value = true
  try {
    await createStocktake(stForm)
    ElMessage.success('盘点完成，盘盈盘亏已计算')
    stVisible.value = false
    await loadStocktakes()
  } finally {
    saving.value = false
  }
}

async function onExport() {
  const data = await exportStockRecords(storeId.value)
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `stock-records-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success(`已导出 ${data.length} 条记录`)
}

onMounted(async () => {
  stores.value = await listAllStores()
  const res = await listSkus({ page: 1, page_size: 200 })
  skus.value = res.list
  await loadRecords()
  await loadStocktakes()
})
</script>

<style scoped>
.toolbar { display: flex; gap: 10px; margin-bottom: 14px; align-items: center; }
.spacer { flex: 1; }
.mt { margin-top: 14px; }
</style>
