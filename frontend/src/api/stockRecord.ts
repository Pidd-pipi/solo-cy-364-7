import request from '@/utils/request'
import type { PageResult, ReplenishSuggestion, StockRecord, Stocktake } from '@/types'
import type { StockRecordTypeValue } from '@/constants/stockRecord'

export function listStockRecords(params: { page?: number; page_size?: number; store_id?: number; sku_id?: number; record_type?: StockRecordTypeValue } = {}): Promise<PageResult<StockRecord>> {
  return request.get('/records', { params })
}

export function createStockRecord(data: { store_id: number; sku_id: number; record_type: StockRecordTypeValue; quantity: number }): Promise<StockRecord> {
  return request.post('/records', data)
}

export function exportStockRecords(storeId?: number): Promise<StockRecord[]> {
  return request.get('/records/export', { params: { store_id: storeId } })
}

export function createStocktake(data: { store_id: number; sku_id: number; stocktake_date: string; actual_qty: number; remark?: string }): Promise<Stocktake> {
  return request.post('/records/stocktakes', data)
}

export function listStocktakes(params: { page?: number; page_size?: number; store_id?: number } = {}): Promise<PageResult<Stocktake>> {
  return request.get('/records/stocktakes', { params })
}

export function getReplenishSuggestions(): Promise<ReplenishSuggestion[]> {
  return request.get('/analysis/suggestions')
}
