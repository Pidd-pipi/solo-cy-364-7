<template>
  <div>
    <el-card>
      <div class="toolbar">
        <el-select v-model="storeId" placeholder="按门店筛选" clearable style="width: 220px" @change="load">
          <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button type="primary" @click="load">查询</el-button>
        <div class="spacer"></div>
        <el-button type="primary" @click="openEnsure">初始化库存</el-button>
      </div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column label="门店" min-width="150">
          <template #default="{ row }">{{ row.store?.name || `#${row.store_id}` }}</template>
        </el-table-column>
        <el-table-column label="商品" min-width="160">
          <template #default="{ row }">{{ row.sku?.name || `#${row.sku_id}` }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="当前库存" width="110" />
        <el-table-column prop="safety_stock" label="安全库存" width="110" />
        <el-table-column label="库存状态" width="120">
          <template #default="{ row }">
            <InventoryStatusBadge :quantity="row.quantity" :safety-stock="row.safety_stock" />
          </template>
        </el-table-column>
        <el-table-column label="库存水位" min-width="180">
          <template #default="{ row }">
            <StockLevelIndicator :quantity="row.quantity" :safety-stock="row.safety_stock" :unit="row.sku?.unit" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openSafety(row)">设置安全库存</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="mt" layout="total, prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="load" />
    </el-card>

    <el-dialog v-model="safetyVisible" title="设置安全库存" width="400px">
      <el-form label-width="100px">
        <el-form-item label="安全库存">
          <el-input-number v-model="safetyStock" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="safetyVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSaveSafety">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="ensureVisible" title="初始化门店库存" width="440px">
      <el-form label-width="90px">
        <el-form-item label="门店" required>
          <el-select v-model="ensureForm.store_id" style="width: 100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品" required>
          <el-select v-model="ensureForm.sku_id" style="width: 100%">
            <el-option v-for="s in skus" :key="s.id" :label="`${s.name}（${s.code}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量" required>
          <el-input-number v-model="ensureForm.quantity" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ensureVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onEnsure">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import InventoryStatusBadge from '@/components/common/InventoryStatusBadge.vue'
import StockLevelIndicator from '@/components/common/StockLevelIndicator.vue'
import { listInventories, setSafetyStock, ensureInventory } from '@/api/storeInventory'
import { listAllStores } from '@/api/store'
import { listSkus } from '@/api/sku'
import type { Store, SKU, StoreInventory } from '@/types'

const list = ref<StoreInventory[]>([])
const stores = ref<Store[]>([])
const skus = ref<SKU[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const storeId = ref<number>()
const loading = ref(false)
const saving = ref(false)
const safetyVisible = ref(false)
const ensureVisible = ref(false)
const safetyStock = ref(0)
const currentInv = ref<StoreInventory | null>(null)
const ensureForm = reactive({ store_id: 0, sku_id: 0, quantity: 0 })

async function load() {
  loading.value = true
  try {
    const res = await listInventories({ page: page.value, page_size: pageSize.value, store_id: storeId.value })
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function openSafety(row: StoreInventory) {
  currentInv.value = row
  safetyStock.value = row.safety_stock
  safetyVisible.value = true
}

async function onSaveSafety() {
  if (!currentInv.value) return
  saving.value = true
  try {
    await setSafetyStock(currentInv.value.id, safetyStock.value)
    ElMessage.success('安全库存已更新')
    safetyVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

function openEnsure() {
  ensureForm.store_id = stores.value[0]?.id || 0
  ensureForm.sku_id = skus.value[0]?.id || 0
  ensureForm.quantity = 0
  ensureVisible.value = true
}

async function onEnsure() {
  saving.value = true
  try {
    await ensureInventory(ensureForm)
    ElMessage.success('库存已初始化')
    ensureVisible.value = false
    await load()
  } finally {
    saving.value = false
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
</style>
