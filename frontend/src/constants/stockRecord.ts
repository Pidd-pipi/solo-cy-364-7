// StockRecordType 出入库类型枚举（与 backend/internal/constants/stock_record.go 对应）
export const StockRecordType = {
  PURCHASE: 'purchase',
  TRANSFER_IN: 'transfer_in',
  SALE: 'sale',
  LOSS: 'loss'
} as const

export type StockRecordTypeValue = (typeof StockRecordType)[keyof typeof StockRecordType]

export const STOCK_RECORD_TYPE_OPTIONS: { label: string; value: StockRecordTypeValue }[] = [
  { label: '采购入库', value: StockRecordType.PURCHASE },
  { label: '调拨入库', value: StockRecordType.TRANSFER_IN },
  { label: '销售出库', value: StockRecordType.SALE },
  { label: '损耗出库', value: StockRecordType.LOSS }
]

export const STOCK_RECORD_TYPE_TEXT: Record<StockRecordTypeValue, string> = {
  [StockRecordType.PURCHASE]: '采购入库',
  [StockRecordType.TRANSFER_IN]: '调拨入库',
  [StockRecordType.SALE]: '销售出库',
  [StockRecordType.LOSS]: '损耗出库'
}

export const STOCK_RECORD_TYPE_TAG: Record<StockRecordTypeValue, string> = {
  [StockRecordType.PURCHASE]: 'success',
  [StockRecordType.TRANSFER_IN]: 'primary',
  [StockRecordType.SALE]: 'warning',
  [StockRecordType.LOSS]: 'danger'
}
