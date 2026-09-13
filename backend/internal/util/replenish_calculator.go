package util

// ReplenishSuggestion 补货建议项。
type ReplenishSuggestion struct {
	SkuID       uint   `json:"sku_id"`
	SkuCode     string `json:"sku_code"`
	SkuName     string `json:"sku_name"`
	StoreID     uint   `json:"store_id"`
	StoreName   string `json:"store_name"`
	Quantity    int    `json:"quantity"`
	SafetyStock int    `json:"safety_stock"`
	SuggestQty  int    `json:"suggest_qty"`
	SlowMoving  bool   `json:"slow_moving"`
	Reason      string `json:"reason"`
}

// CalculateSuggestQty 补货建议数量：建议补至安全库存的 1.5 倍，至少补齐缺口。
func CalculateSuggestQty(quantity, safetyStock int) int {
	if quantity >= safetyStock {
		return 0
	}
	target := safetyStock * 3 / 2
	if target < safetyStock {
		target = safetyStock
	}
	return target - quantity
}

// IsSlowMoving 滞销判定：库存充足但销量极少（参考月销量低于库存的 10%）。
func IsSlowMoving(quantity, monthlySales int) bool {
	if quantity <= 0 {
		return false
	}
	return monthlySales*10 < quantity
}
