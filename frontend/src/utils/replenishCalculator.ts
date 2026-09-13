// 补货建议数量与滞销分析计算
export function calculateSuggestQty(quantity: number, safetyStock: number): number {
  if (quantity >= safetyStock) return 0
  const target = Math.max(Math.floor((safetyStock * 3) / 2), safetyStock)
  return Math.max(target - quantity, 0)
}

export function isSlowMoving(quantity: number, monthlySales: number): boolean {
  if (quantity <= 0) return false
  return monthlySales * 10 < quantity
}

export function inventoryStatusText(quantity: number, safetyStock: number): string {
  if (quantity <= 0) return '缺货'
  if (quantity < safetyStock) return '低库存'
  return '正常'
}

export function inventoryStatusType(quantity: number, safetyStock: number): 'success' | 'warning' | 'danger' {
  if (quantity <= 0) return 'danger'
  if (quantity < safetyStock) return 'warning'
  return 'success'
}
