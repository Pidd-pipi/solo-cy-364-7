import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getMe, login as loginApi, register as registerApi } from '@/api/user'
import type { User } from '@/types'
import { UserRole, type UserRoleValue } from '@/constants/user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const role = computed<UserRoleValue | ''>(() => user.value?.role || '')

  async function login(username: string, password: string) {
    const res = await loginApi({ username, password })
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
  }

  async function register(data: { username: string; password: string; name: string; role: string }) {
    await registerApi(data)
  }

  async function fetchMe() {
    if (!token.value) return
    user.value = await getMe()
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  function hasRole(roles: UserRoleValue[]): boolean {
    return role.value !== '' && (roles as string[]).includes(role.value)
  }

  return { token, user, isLoggedIn, role, login, register, fetchMe, logout, hasRole }
})

export function isAdminRole(role: string | undefined): boolean {
  return role === UserRole.ADMIN || role === UserRole.HQ
}
