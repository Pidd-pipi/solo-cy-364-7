import type { UserRoleValue } from '@/constants/user'
import type { TransferStatusValue } from '@/constants/transfer'
import type { StockRecordTypeValue } from '@/constants/stockRecord'

export interface User {
  id: number
  username: string
  name: string
  role: UserRoleValue
  store_id: number | null
  store?: Store | null
  created_at: string
}

export interface Store {
  id: number
  code: string
  name: string
  address: string
  manager_user_id: number | null
  manager?: User | null
  created_at: string
}

export interface SKU {
  id: number
  code: string
  name: string
  spec: string
  barcode: string
  category: string
  unit: string
  status: string
  created_at: string
}

export interface StoreInventory {
  id: number
  store_id: number
  store?: Store
  sku_id: number
  sku?: SKU
  quantity: number
  safety_stock: number
  updated_at: string
}

export interface TransferOrder {
  id: number
  from_store_id: number
  from_store?: Store
  to_store_id: number
  to_store?: Store
  sku_id: number
  sku?: SKU
  quantity: number
  reason: string
  status: TransferStatusValue
  created_at: string
}

export interface StockRecord {
  id: number
  store_id: number
  store?: Store
  sku_id: number
  sku?: SKU
  record_type: StockRecordTypeValue
  quantity: number
  related_order_id: number | null
  created_at: string
}

export interface Stocktake {
  id: number
  store_id: number
  store?: Store
  sku_id: number
  sku?: SKU
  stocktake_date: string
  system_qty: number
  actual_qty: number
  difference: number
  remark: string
  created_at: string
}

export interface InventoryStats {
  total_sku_count: number
  total_quantity: number
  low_stock_count: number
  out_of_stock_count: number
  store_count: number
}

export interface ReplenishSuggestion {
  sku_id: number
  sku_code: string
  sku_name: string
  store_id: number
  store_name: string
  quantity: number
  safety_stock: number
  suggest_qty: number
  slow_moving: boolean
  reason: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
