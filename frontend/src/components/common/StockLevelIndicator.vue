<template>
  <div class="stock-level">
    <el-progress :percentage="percentage" :color="barColor" :stroke-width="10" />
    <div class="stock-level-text">
      <span>当前 {{ quantity }} {{ unit }}</span>
      <span>安全线 {{ safetyStock }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ quantity: number; safetyStock: number; unit?: string }>()

const percentage = computed(() => {
  if (props.safetyStock <= 0) return Math.min(props.quantity, 100)
  return Math.min(Math.round((props.quantity / (props.safetyStock * 1.5)) * 100), 100)
})

const barColor = computed(() => {
  if (props.quantity <= 0) return '#f56c6c'
  if (props.quantity < props.safetyStock) return '#e6a23c'
  return '#67c23a'
})
</script>

<style scoped>
.stock-level { width: 100%; }
.stock-level-text { display: flex; justify-content: space-between; font-size: 12px; color: #909399; margin-top: 4px; }
</style>
