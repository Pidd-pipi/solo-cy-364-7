import { createRouter, createWebHistory } from 'vue-router'
import { setupGuards, AdminRoles } from './guards'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/Login.vue'),
      meta: { public: true, title: '登录' }
    },
    {
      path: '/',
      component: () => import('@/pages/Layout.vue'),
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'dashboard', component: () => import('@/pages/Dashboard.vue'), meta: { title: '库存总览' } },
        { path: 'skus', name: 'skus', component: () => import('@/pages/Skus.vue'), meta: { title: 'SKU 主数据', roles: AdminRoles } },
        { path: 'inventory', name: 'inventory', component: () => import('@/pages/Inventory.vue'), meta: { title: '门店库存' } },
        { path: 'transfers', name: 'transfers', component: () => import('@/pages/Transfers.vue'), meta: { title: '调拨管理' } },
        { path: 'records', name: 'records', component: () => import('@/pages/Records.vue'), meta: { title: '出入库与盘点' } },
        { path: 'analysis', name: 'analysis', component: () => import('@/pages/Analysis.vue'), meta: { title: '滞销分析与补货' } },
        { path: 'profile', name: 'profile', component: () => import('@/pages/Profile.vue'), meta: { title: '个人中心' } }
      ]
    }
  ]
})

setupGuards(router)
export default router
