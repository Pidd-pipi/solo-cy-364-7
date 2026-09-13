import { onMounted, ref } from 'vue'
import { getInventoryStats, listInventoryAlerts } from '@/api/storeInventory'
import type { InventoryStats, StoreInventory } from '@/types'

// useInventoryStats：多门店库存汇总与预警统计
export function useInventoryStats() {
  const stats = ref<InventoryStats | null>(null)
  const alerts = ref<StoreInventory[]>([])
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      const [s, a] = await Promise.all([getInventoryStats(), listInventoryAlerts()])
      stats.value = s
      alerts.value = a
    } finally {
      loading.value = false
    }
  }

  onMounted(load)

  return { stats, alerts, loading, reload: load }
}
