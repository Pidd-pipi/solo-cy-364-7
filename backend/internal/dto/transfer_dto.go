package dto

// TransferCreateRequest 调拨单创建请求。
type TransferCreateRequest struct {
	FromStoreID uint   `json:"from_store_id" binding:"required"`
	ToStoreID   uint   `json:"to_store_id" binding:"required"`
	SKUID       uint   `json:"sku_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
	Reason      string `json:"reason" binding:"max=255"`
}
