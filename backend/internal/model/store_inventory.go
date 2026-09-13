package model

import "time"

// StoreInventory 门店库存。
type StoreInventory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	StoreID     uint      `gorm:"index;not null" json:"store_id"`
	Store       *Store    `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	SKUID       uint      `gorm:"column:sku_id;index;not null" json:"sku_id"`
	SKU         *SKU      `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	Quantity    int       `gorm:"not null;default:0" json:"quantity"`
	SafetyStock int       `gorm:"not null;default:0" json:"safety_stock"`
	UpdatedAt   time.Time `json:"updated_at"`
}
