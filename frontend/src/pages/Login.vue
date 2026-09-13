<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="login-title">连锁门店库存调配系统</h2>
      <el-form :model="form" @keyup.enter="onSubmit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" show-password :prefix-icon="Lock" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-btn" :loading="loading" @click="onSubmit">登录</el-button>
        </el-form-item>
      </el-form>
      <div class="login-tip">种子账号：admin/admin123、hquser/hq123456、manager1/sm123456</div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/authStore'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch {
    // 错误提示由拦截器处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page { height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #1f3b73 0%, #2f6f9f 100%); }
.login-card { width: 380px; }
.login-title { text-align: center; margin-bottom: 24px; }
.login-btn { width: 100%; }
.login-tip { text-align: center; color: #909399; font-size: 12px; margin-top: 8px; }
</style>
