<template>
  <el-tag :type="statusType" size="small">{{ statusText }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { inventoryStatusText, inventoryStatusType } from '@/utils/replenishCalculator'

const props = defineProps<{ quantity: number; safetyStock: number }>()

const statusText = computed(() => inventoryStatusText(props.quantity, props.safetyStock))
const statusType = computed(() => inventoryStatusType(props.quantity, props.safetyStock))
</script>
