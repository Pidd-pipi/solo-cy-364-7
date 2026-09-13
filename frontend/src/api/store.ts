import request from '@/utils/request'
import type { PageResult, Store } from '@/types'

export function listStores(params: { page?: number; page_size?: number } = {}): Promise<PageResult<Store>> {
  return request.get('/stores', { params })
}

export function listAllStores(): Promise<Store[]> {
  return request.get('/stores/all')
}

export function createStore(data: { code: string; name: string; address?: string; manager_user_id?: number }): Promise<Store> {
  return request.post('/stores', data)
}

export function updateStore(id: number, data: { code?: string; name?: string; address?: string; manager_user_id?: number }): Promise<Store> {
  return request.put(`/stores/${id}`, data)
}

export function deleteStore(id: number): Promise<void> {
  return request.delete(`/stores/${id}`)
}
