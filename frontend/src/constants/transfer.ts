// TransferStatus 调拨单状态枚举（与 backend/internal/constants/transfer.go 对应）
export const TransferStatus = {
  PENDING: 'pending',
  CONFIRMED: 'confirmed',
  SHIPPED: 'shipped',
  RECEIVED: 'received',
  CANCELLED: 'cancelled'
} as const

export type TransferStatusValue = (typeof TransferStatus)[keyof typeof TransferStatus]

export const TRANSFER_STATUS_OPTIONS: { label: string; value: TransferStatusValue }[] = [
  { label: '待确认', value: TransferStatus.PENDING },
  { label: '已确认', value: TransferStatus.CONFIRMED },
  { label: '已发货', value: TransferStatus.SHIPPED },
  { label: '已收货', value: TransferStatus.RECEIVED },
  { label: '已取消', value: TransferStatus.CANCELLED }
]

export const TRANSFER_STATUS_TEXT: Record<TransferStatusValue, string> = {
  [TransferStatus.PENDING]: '待确认',
  [TransferStatus.CONFIRMED]: '已确认',
  [TransferStatus.SHIPPED]: '已发货',
  [TransferStatus.RECEIVED]: '已收货',
  [TransferStatus.CANCELLED]: '已取消'
}

// 状态机：允许的流转，与后端 CanTransfer 对应
export const TRANSFER_STATUS_FLOW: Record<TransferStatusValue, TransferStatusValue[]> = {
  [TransferStatus.PENDING]: [TransferStatus.CONFIRMED, TransferStatus.CANCELLED],
  [TransferStatus.CONFIRMED]: [TransferStatus.SHIPPED, TransferStatus.CANCELLED],
  [TransferStatus.SHIPPED]: [TransferStatus.RECEIVED, TransferStatus.CANCELLED],
  [TransferStatus.RECEIVED]: [],
  [TransferStatus.CANCELLED]: []
}

export function canTransfer(from: TransferStatusValue, to: TransferStatusValue): boolean {
  return TRANSFER_STATUS_FLOW[from]?.includes(to) ?? false
}
