package model

import "time"

// Stocktake 周期盘点记录：系统库存与实际库存对比，自动计算盘盈盘亏。
type Stocktake struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StoreID       uint      `gorm:"index;not null" json:"store_id"`
	Store         *Store    `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	SKUID         uint      `gorm:"column:sku_id;index;not null" json:"sku_id"`
	SKU           *SKU      `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	StocktakeDate string    `gorm:"size:16;index" json:"stocktake_date"`
	SystemQty     int       `gorm:"not null" json:"system_qty"`
	ActualQty     int       `gorm:"not null" json:"actual_qty"`
	Difference    int       `gorm:"not null" json:"difference"`
	Remark        string    `gorm:"size:255" json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
}
