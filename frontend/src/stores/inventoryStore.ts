import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listInventories, getInventoryStats, listInventoryAlerts, setSafetyStock } from '@/api/storeInventory'
import type { InventoryStats, StoreInventory } from '@/types'

export const useInventoryStore = defineStore('inventory', () => {
  const list = ref<StoreInventory[]>([])
  const total = ref(0)
  const stats = ref<InventoryStats | null>(null)
  const alerts = ref<StoreInventory[]>([])
  const loading = ref(false)

  async function fetchList(params: { page?: number; page_size?: number; store_id?: number } = {}) {
    loading.value = true
    try {
      const res = await listInventories(params)
      list.value = res.list
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    stats.value = await getInventoryStats()
  }

  async function fetchAlerts() {
    alerts.value = await listInventoryAlerts()
  }

  async function updateSafetyStock(id: number, safetyStock: number) {
    await setSafetyStock(id, safetyStock)
  }

  return { list, total, stats, alerts, loading, fetchList, fetchStats, fetchAlerts, updateSafetyStock }
})
