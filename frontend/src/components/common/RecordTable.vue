<template>
  <el-table :data="records" v-loading="loading" border stripe>
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="门店" min-width="130">
      <template #default="{ row }">{{ row.store?.name || `#${row.store_id}` }}</template>
    </el-table-column>
    <el-table-column label="商品" min-width="150">
      <template #default="{ row }">{{ row.sku?.name || `#${row.sku_id}` }}</template>
    </el-table-column>
    <el-table-column label="类型" width="100">
      <template #default="{ row }">
        <el-tag :type="STOCK_RECORD_TYPE_TAG[row.record_type] || 'info'" size="small">
          {{ STOCK_RECORD_TYPE_TEXT[row.record_type] || row.record_type }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="quantity" label="数量" width="90" />
    <el-table-column label="时间" width="160">
      <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { StockRecord } from '@/types'
import { STOCK_RECORD_TYPE_TAG, STOCK_RECORD_TYPE_TEXT } from '@/constants/stockRecord'
import { formatDateTime } from '@/utils/dateFormat'

defineProps<{ records: StockRecord[]; loading?: boolean }>()
</script>
