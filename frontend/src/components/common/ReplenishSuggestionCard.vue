<template>
  <el-card shadow="hover" class="suggestion-card">
    <div class="suggestion-head">
      <span class="sku-name">{{ suggestion.sku_name }}</span>
      <el-tag v-if="suggestion.slow_moving" type="warning" size="small">滞销</el-tag>
      <el-tag v-else type="danger" size="small">低库存</el-tag>
    </div>
    <div class="suggestion-row"><span>门店</span><b>{{ suggestion.store_name }}</b></div>
    <div class="suggestion-row"><span>现有库存</span><b>{{ suggestion.quantity }}</b></div>
    <div class="suggestion-row"><span>安全库存</span><b>{{ suggestion.safety_stock }}</b></div>
    <div class="suggestion-row"><span>建议补货</span><b class="suggest-qty">{{ suggestion.suggest_qty }}</b></div>
    <div class="suggestion-reason">{{ suggestion.reason }}</div>
  </el-card>
</template>

<script setup lang="ts">
import type { ReplenishSuggestion } from '@/types'
defineProps<{ suggestion: ReplenishSuggestion }>()
</script>

<style scoped>
.suggestion-card { margin-bottom: 12px; }
.suggestion-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.sku-name { font-weight: 600; }
.suggestion-row { display: flex; justify-content: space-between; font-size: 13px; padding: 2px 0; color: #606266; }
.suggest-qty { color: #f56c6c; font-size: 15px; }
.suggestion-reason { margin-top: 6px; font-size: 12px; color: #909399; }
</style>
