<template>
  <div v-loading="loading">
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6"><el-card><el-statistic title="SKU 总数" :value="stats?.total_sku_count || 0" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="库存总量" :value="stats?.total_quantity || 0" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="低库存预警" :value="stats?.low_stock_count || 0" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="缺货" :value="stats?.out_of_stock_count || 0" /></el-card></el-col>
    </el-row>
    <el-card class="mt">
      <template #header>低库存预警（补货提醒）</template>
      <el-table :data="alerts" border stripe>
        <el-table-column label="门店" prop="store.name" min-width="140" />
        <el-table-column label="商品" min-width="160">
          <template #default="{ row }">{{ row.sku?.name }}（{{ row.sku?.code }}）</template>
        </el-table-column>
        <el-table-column label="当前库存" prop="quantity" width="110" />
        <el-table-column label="安全库存" prop="safety_stock" width="110" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <InventoryStatusBadge :quantity="row.quantity" :safety-stock="row.safety_stock" />
          </template>
        </el-table-column>
        <el-table-column label="建议补货" width="110">
          <template #default="{ row }">{{ calculateSuggestQty(row.quantity, row.safety_stock) }}</template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!alerts.length" description="暂无低库存预警" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useInventoryStats } from '@/hooks/useInventoryStats'
import InventoryStatusBadge from '@/components/common/InventoryStatusBadge.vue'
import { calculateSuggestQty } from '@/utils/replenishCalculator'

const { stats, alerts, loading } = useInventoryStats()
</script>

<style scoped>
.stat-row { margin-bottom: 0; }
.mt { margin-top: 16px; }
</style>
