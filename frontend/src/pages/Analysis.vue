<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card>
          <template #header>补货建议（低库存）</template>
          <div v-loading="loading">
            <ReplenishSuggestionCard v-for="s in suggestions" :key="`${s.store_id}-${s.sku_id}`" :suggestion="s" />
            <el-empty v-if="!suggestions.length" description="暂无补货建议" />
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>滞销商品分析</template>
          <el-table :data="slowMoving" v-loading="loading" border stripe>
            <el-table-column prop="sku_name" label="商品" min-width="140" />
            <el-table-column prop="store_name" label="门店" min-width="130" />
            <el-table-column prop="quantity" label="库存" width="80" />
            <el-table-column prop="safety_stock" label="安全库存" width="90" />
            <el-table-column prop="reason" label="判定" min-width="140" />
          </el-table>
          <el-empty v-if="!slowMoving.length" description="暂无滞销商品" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import ReplenishSuggestionCard from '@/components/common/ReplenishSuggestionCard.vue'
import { getReplenishSuggestions } from '@/api/stockRecord'
import type { ReplenishSuggestion } from '@/types'

const suggestions = ref<ReplenishSuggestion[]>([])
const loading = ref(false)

const slowMoving = computed(() => suggestions.value.filter((s) => s.slow_moving))

onMounted(async () => {
  loading.value = true
  try {
    suggestions.value = await getReplenishSuggestions()
  } finally {
    loading.value = false
  }
})
</script>
