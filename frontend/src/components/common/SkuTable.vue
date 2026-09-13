<template>
  <el-table :data="skus" v-loading="loading" border stripe>
    <el-table-column prop="code" label="SKU 编码" width="120" />
    <el-table-column prop="name" label="名称" min-width="160" />
    <el-table-column prop="spec" label="规格" width="140" />
    <el-table-column prop="barcode" label="条码" width="150" />
    <el-table-column prop="category" label="分类" width="110" />
    <el-table-column prop="unit" label="单位" width="80" />
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
          {{ row.status === 'active' ? '启用' : '停用' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="140" fixed="right">
      <template #default="{ row }">
        <el-button link type="primary" size="small" @click="$emit('edit', row)">编辑</el-button>
        <el-button link type="danger" size="small" @click="$emit('delete', row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { SKU } from '@/types'

defineProps<{ skus: SKU[]; loading?: boolean }>()
defineEmits<{ (e: 'edit', row: SKU): void; (e: 'delete', row: SKU): void }>()
</script>
