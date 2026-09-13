<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">门店库存调配系统</div>
      <el-menu :default-active="route.path" router background-color="#001529" text-color="#a6adb4" active-text-color="#fff">
        <el-menu-item index="/dashboard"><el-icon><Odometer /></el-icon><span>库存总览</span></el-menu-item>
        <el-menu-item index="/skus"><el-icon><Goods /></el-icon><span>SKU 主数据</span></el-menu-item>
        <el-menu-item index="/inventory"><el-icon><Box /></el-icon><span>门店库存</span></el-menu-item>
        <el-menu-item index="/transfers"><el-icon><Switch /></el-icon><span>调拨管理</span></el-menu-item>
        <el-menu-item index="/records"><el-icon><Document /></el-icon><span>出入库与盘点</span></el-menu-item>
        <el-menu-item index="/analysis"><el-icon><DataAnalysis /></el-icon><span>滞销分析与补货</span></el-menu-item>
        <el-menu-item index="/profile"><el-icon><User /></el-icon><span>个人中心</span></el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-title">{{ route.meta.title }}</div>
        <div class="header-user">
          <span>{{ auth.user?.name || auth.user?.username }}</span>
          <el-tag size="small" type="info">{{ USER_ROLE_TEXT[auth.role as UserRoleValue] || auth.role }}</el-tag>
          <el-button link type="primary" @click="onLogout">退出</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/authStore'
import { USER_ROLE_TEXT, type UserRoleValue } from '@/constants/user'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  if (!auth.user) auth.fetchMe()
})

async function onLogout() {
  await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout { height: 100vh; }
.aside { background: #001529; }
.logo { height: 56px; line-height: 56px; text-align: center; color: #fff; font-weight: 700; font-size: 15px; }
.aside :deep(.el-menu) { border-right: none; }
.header { display: flex; justify-content: space-between; align-items: center; background: #fff; border-bottom: 1px solid #eee; }
.header-title { font-size: 16px; font-weight: 600; }
.header-user { display: flex; align-items: center; gap: 10px; }
.main { background: #f5f7fa; }
</style>
