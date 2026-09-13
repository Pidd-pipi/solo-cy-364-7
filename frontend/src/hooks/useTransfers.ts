import { onMounted, ref } from 'vue'
import { listTransfers, confirmTransfer, shipTransfer, receiveTransfer, cancelTransfer } from '@/api/transferOrder'
import type { TransferOrder } from '@/types'
import type { TransferStatusValue } from '@/constants/transfer'

// useTransfers：调拨单列表与审批状态管理
export function useTransfers() {
  const list = ref<TransferOrder[]>([])
  const total = ref(0)
  const loading = ref(false)
  const page = ref(1)
  const pageSize = ref(10)
  const status = ref<TransferStatusValue | ''>('')

  async function load() {
    loading.value = true
    try {
      const params: { page: number; page_size: number; status?: TransferStatusValue } = { page: page.value, page_size: pageSize.value }
      if (status.value) params.status = status.value
      const res = await listTransfers(params)
      list.value = res.list
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  async function confirm(id: number) { await confirmTransfer(id); await load() }
  async function ship(id: number) { await shipTransfer(id); await load() }
  async function receive(id: number) { await receiveTransfer(id); await load() }
  async function cancel(id: number) { await cancelTransfer(id); await load() }

  onMounted(load)

  return { list, total, loading, page, pageSize, status, load, confirm, ship, receive, cancel }
}
