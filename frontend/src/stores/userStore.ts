import { defineStore } from 'pinia'
import { ref } from 'vue'
import { updateProfile } from '@/api/user'
import type { User } from '@/types'

// 用户资料 store：与 authStore 分工，负责用户资料维护。
export const useUserStore = defineStore('user', () => {
  const profile = ref<User | null>(null)

  function setProfile(user: User | null) {
    profile.value = user
  }

  async function saveProfile(payload: { name?: string; store_id?: number }): Promise<User> {
    const updated = await updateProfile(payload)
    profile.value = updated
    return updated
  }

  return { profile, setProfile, saveProfile }
})
