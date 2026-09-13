// 调拨单收货进度展示逻辑（纯函数，便于测试）。
// 规则：仅已发货在途单显示剩余待收；已取消且发生过收货的单据显示未收已退回；
// 收满（已收货）与未进入收货阶段的单据不显示剩余。
import { TransferStatus, type TransferStatusValue } from '../constants/transfer'

export interface TransferProgress {
  received: number
  quantity: number
  remaining: number
  // 剩余待收：仅已发货未收满时展示
  showRemaining: boolean
  // 未收已退回：仅已取消且已有收货记录时展示
  showReturned: boolean
}

export function transferRemaining(quantity: number, receivedQuantity: number): number {
  return Math.max(0, quantity - (receivedQuantity || 0))
}

export function transferProgress(status: TransferStatusValue, quantity: number, receivedQuantity: number): TransferProgress {
  const received = receivedQuantity || 0
  const remaining = transferRemaining(quantity, received)
  return {
    received,
    quantity,
    remaining,
    showRemaining: status === TransferStatus.SHIPPED && remaining > 0,
    showReturned: status === TransferStatus.CANCELLED && received > 0 && remaining > 0,
  }
}
