import request from '@/utils/request'
import type { InventoryStats, PageResult, StoreInventory } from '@/types'

export function listInventories(params: { page?: number; page_size?: number; store_id?: number; sku_id?: number } = {}): Promise<PageResult<StoreInventory>> {
  return request.get('/inventories', { params })
}

export function getInventoryStats(): Promise<InventoryStats> {
  return request.get('/inventories/stats')
}

export function listInventoryAlerts(): Promise<StoreInventory[]> {
  return request.get('/inventories/alerts')
}

export function setSafetyStock(id: number, safetyStock: number): Promise<StoreInventory> {
  return request.put(`/inventories/${id}/safety-stock`, { safety_stock: safetyStock })
}

export function ensureInventory(data: { store_id: number; sku_id: number; quantity: number }): Promise<StoreInventory> {
  return request.post('/inventories/ensure', data)
}
