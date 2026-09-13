package dto

// SafetyStockRequest 安全库存设置请求。
type SafetyStockRequest struct {
	SafetyStock int `json:"safety_stock" binding:"min=0"`
}

// InventoryEnsureRequest 初始化门店库存请求。
type InventoryEnsureRequest struct {
	StoreID  uint `json:"store_id" binding:"required"`
	SKUID    uint `json:"sku_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"min=0"`
}
