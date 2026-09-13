import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listTransfers, createTransfer, confirmTransfer, shipTransfer, receiveTransfer, cancelTransfer } from '@/api/transferOrder'
import type { TransferOrder } from '@/types'
import type { TransferStatusValue } from '@/constants/transfer'

export const useTransferStore = defineStore('transfer', () => {
  const list = ref<TransferOrder[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchList(params: { page?: number; page_size?: number; store_id?: number; status?: TransferStatusValue } = {}) {
    loading.value = true
    try {
      const res = await listTransfers(params)
      list.value = res.list
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  async function create(data: { from_store_id: number; to_store_id: number; sku_id: number; quantity: number; reason?: string }) {
    await createTransfer(data)
    await fetchList()
  }

  async function confirm(id: number) {
    await confirmTransfer(id)
    await fetchList()
  }

  async function ship(id: number) {
    await shipTransfer(id)
    await fetchList()
  }

  async function receive(id: number) {
    await receiveTransfer(id)
    await fetchList()
  }

  async function cancel(id: number) {
    await cancelTransfer(id)
    await fetchList()
  }

  return { list, total, loading, fetchList, create, confirm, ship, receive, cancel }
})
