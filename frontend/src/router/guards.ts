import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import { UserRole, type UserRoleValue } from '@/constants/user'

// 路由守卫：登录态 + 角色权限
export function setupGuards(router: Router) {
  router.beforeEach(async (to) => {
    const auth = useAuthStore()
    if (to.meta.public) {
      if (to.name === 'login' && auth.isLoggedIn) return { name: 'dashboard' }
      return true
    }
    if (!auth.isLoggedIn) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    if (!auth.user) {
      try {
        await auth.fetchMe()
      } catch {
        auth.logout()
        return { name: 'login' }
      }
    }
    const roles = to.meta.roles as UserRoleValue[] | undefined
    if (roles && roles.length > 0 && !auth.hasRole(roles)) {
      return { name: 'dashboard' }
    }
    return true
  })
}

export const AdminRoles: UserRoleValue[] = [UserRole.ADMIN, UserRole.HQ]
export const AllRoles: UserRoleValue[] = [UserRole.ADMIN, UserRole.HQ, UserRole.STORE_MANAGER]
