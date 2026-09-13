<template>
  <div>
    <el-card>
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="编码/名称/条码搜索" clearable style="width: 240px" @keyup.enter="load" />
        <el-select v-model="category" placeholder="分类筛选" clearable style="width: 160px">
          <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
        </el-select>
        <el-button type="primary" @click="load">查询</el-button>
        <div class="spacer"></div>
        <el-button type="primary" @click="openCreate">新增 SKU</el-button>
        <el-button type="warning" @click="openImport">批量导入</el-button>
      </div>
      <SkuTable :skus="list" :loading="loading" @edit="openEdit" @delete="onDelete" />
      <el-pagination class="mt" layout="total, prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="load" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑 SKU' : '新增 SKU'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="编码" required><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="规格"><el-input v-model="form.spec" /></el-form-item>
        <el-form-item label="条码"><el-input v-model="form.barcode" /></el-form-item>
        <el-form-item label="分类"><el-input v-model="form.category" /></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.unit" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="批量导入 SKU" width="560px">
      <el-alert type="info" :closable="false" :title="importTip" />
      <el-input v-model="importText" type="textarea" :rows="8" class="mt" placeholder="JSON 数组" />
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="onImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import SkuTable from '@/components/common/SkuTable.vue'
import { listSkus, createSku, updateSku, deleteSku, batchImportSkus } from '@/api/sku'
import type { SKU } from '@/types'

const list = ref<SKU[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const category = ref('')
const categories = ['饮料', '方便食品', '零食', '日用品']
const loading = ref(false)
const saving = ref(false)
const importing = ref(false)
const dialogVisible = ref(false)
const importVisible = ref(false)
const editing = ref<SKU | null>(null)
const importText = ref('')
const importTip = '请粘贴 JSON 数组，例如：[{\"code\":\"SKU9001\",\"name\":\"测试商品\",\"category\":\"饮料\",\"unit\":\"瓶\"}]'
const form = reactive({ code: '', name: '', spec: '', barcode: '', category: '', unit: '' })

async function load() {
  loading.value = true
  try {
    const res = await listSkus({ page: page.value, page_size: pageSize.value, category: category.value, keyword: keyword.value })
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { code: '', name: '', spec: '', barcode: '', category: '', unit: '' })
  dialogVisible.value = true
}

function openEdit(row: SKU) {
  editing.value = row
  Object.assign(form, { code: row.code, name: row.name, spec: row.spec, barcode: row.barcode, category: row.category, unit: row.unit })
  dialogVisible.value = true
}

async function onSave() {
  if (!form.code || !form.name) {
    ElMessage.warning('编码与名称必填')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateSku(editing.value.id, form)
    } else {
      await createSku(form)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await load()
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}

async function onDelete(row: SKU) {
  await ElMessageBox.confirm(`确认删除 SKU「${row.name}」？`, '提示', { type: 'warning' })
  await deleteSku(row.id)
  ElMessage.success('删除成功')
  await load()
}

async function onImport() {
  let items: unknown
  try {
    items = JSON.parse(importText.value)
  } catch {
    ElMessage.error('JSON 格式错误')
    return
  }
  if (!Array.isArray(items) || items.length === 0) {
    ElMessage.warning('请至少提供一条数据')
    return
  }
  importing.value = true
  try {
    const res = await batchImportSkus(items as never)
    ElMessage.success(`成功导入 ${res.imported} 条`)
    importVisible.value = false
    await load()
  } catch {
    // handled
  } finally {
    importing.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 10px; margin-bottom: 14px; align-items: center; }
.spacer { flex: 1; }
.mt { margin-top: 14px; }
</style>
