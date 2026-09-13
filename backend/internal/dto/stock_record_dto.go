package dto

import "github.com/ld/storeinventory/internal/constants"

// StockRecordCreateRequest 出入库记录创建请求。
type StockRecordCreateRequest struct {
	StoreID    uint                      `json:"store_id" binding:"required"`
	SKUID      uint                      `json:"sku_id" binding:"required"`
	RecordType constants.StockRecordType `json:"record_type" binding:"required"`
	Quantity   int                       `json:"quantity" binding:"required,min=1"`
}

// StocktakeCreateRequest 盘点创建请求。
type StocktakeCreateRequest struct {
	StoreID       uint   `json:"store_id" binding:"required"`
	SKUID         uint   `json:"sku_id" binding:"required"`
	StocktakeDate string `json:"stocktake_date" binding:"required"`
	ActualQty     int    `json:"actual_qty" binding:"min=0"`
	Remark        string `json:"remark" binding:"max=255"`
}
