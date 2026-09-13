import request from '@/utils/request'
import type { PageResult, TransferOrder } from '@/types'
import type { TransferStatusValue } from '@/constants/transfer'

export function listTransfers(params: { page?: number; page_size?: number; store_id?: number; status?: TransferStatusValue } = {}): Promise<PageResult<TransferOrder>> {
  return request.get('/transfers', { params })
}

export function createTransfer(data: { from_store_id: number; to_store_id: number; sku_id: number; quantity: number; reason?: string }): Promise<TransferOrder> {
  return request.post('/transfers', data)
}

export function confirmTransfer(id: number): Promise<TransferOrder> {
  return request.put(`/transfers/${id}/confirm`)
}

export function shipTransfer(id: number): Promise<TransferOrder> {
  return request.put(`/transfers/${id}/ship`)
}

export function receiveTransfer(id: number): Promise<TransferOrder> {
  return request.put(`/transfers/${id}/receive`)
}

export function cancelTransfer(id: number): Promise<TransferOrder> {
  return request.put(`/transfers/${id}/cancel`)
}
