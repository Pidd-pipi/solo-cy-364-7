import { computed } from 'vue'
import { useAuthStore } from '@/stores/authStore'
import type { UserRoleValue } from '@/constants/user'

// useAuth：登录态与角色判断
export function useAuth() {
  const authStore = useAuthStore()
  const isLoggedIn = computed(() => authStore.isLoggedIn)
  const user = computed(() => authStore.user)
  const role = computed(() => authStore.role)

  function hasRole(roles: UserRoleValue[]): boolean {
    return authStore.hasRole(roles)
  }

  return { isLoggedIn, user, role, hasRole, logout: authStore.logout, login: authStore.login, fetchMe: authStore.fetchMe }
}
