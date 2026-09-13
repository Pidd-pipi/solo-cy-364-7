<template>
  <el-tag :type="tagType" size="small">{{ TRANSFER_STATUS_TEXT[status] || status }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { TRANSFER_STATUS_TEXT, TransferStatus, type TransferStatusValue } from '@/constants/transfer'

const props = defineProps<{ status: TransferStatusValue }>()

const tagType = computed(() => {
  switch (props.status) {
    case TransferStatus.PENDING: return 'info'
    case TransferStatus.CONFIRMED: return 'primary'
    case TransferStatus.SHIPPED: return 'warning'
    case TransferStatus.RECEIVED: return 'success'
    case TransferStatus.CANCELLED: return 'danger'
    default: return 'info'
  }
})
</script>
