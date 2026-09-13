import request from '@/utils/request'
import type { PageResult, SKU } from '@/types'

export function listSkus(params: { page?: number; page_size?: number; category?: string; keyword?: string } = {}): Promise<PageResult<SKU>> {
  return request.get('/skus', { params })
}

export function createSku(data: { code: string; name: string; spec?: string; barcode?: string; category?: string; unit?: string }): Promise<SKU> {
  return request.post('/skus', data)
}

export function updateSku(id: number, data: { code?: string; name?: string; spec?: string; barcode?: string; category?: string; unit?: string }): Promise<SKU> {
  return request.put(`/skus/${id}`, data)
}

export function deleteSku(id: number): Promise<void> {
  return request.delete(`/skus/${id}`)
}

export function batchImportSkus(items: { code: string; name: string; spec?: string; barcode?: string; category?: string; unit?: string }[]): Promise<{ imported: number }> {
  return request.post('/skus/batch-import', { items })
}
