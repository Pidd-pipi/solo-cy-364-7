<template>
  <div>
    <el-card style="max-width: 640px">
      <template #header>个人资料</template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户名">{{ user?.username }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ user?.name }}</el-descriptions-item>
        <el-descriptions-item label="角色">
          <el-tag size="small">{{ USER_ROLE_TEXT[(user?.role || '') as UserRoleValue] || user?.role }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="所属门店">{{ user?.store?.name || '未绑定' }}</el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ formatDateTime(user?.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <el-divider />
      <el-form :model="form" label-width="80px" style="max-width: 400px">
        <el-form-item label="姓名">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="门店">
          <el-select v-model="form.store_id" clearable placeholder="选择门店" style="width: 100%">
            <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="onSave">保存修改</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/authStore'
import { useUserStore } from '@/stores/userStore'
import { listAllStores } from '@/api/store'
import { USER_ROLE_TEXT, type UserRoleValue } from '@/constants/user'
import { formatDateTime } from '@/utils/dateFormat'
import type { Store } from '@/types'

const auth = useAuthStore()
const userStore = useUserStore()
const stores = ref<Store[]>([])
const saving = ref(false)
const user = ref(auth.user)
const form = reactive({ name: '', store_id: undefined as number | undefined })

onMounted(async () => {
  await auth.fetchMe()
  user.value = auth.user
  form.name = user.value?.name || ''
  form.store_id = user.value?.store_id || undefined
  stores.value = await listAllStores()
})

async function onSave() {
  saving.value = true
  try {
    const updated = await userStore.saveProfile({ name: form.name, store_id: form.store_id })
    user.value = updated
    auth.user = updated
    ElMessage.success('资料已更新')
  } finally {
    saving.value = false
  }
}
</script>
