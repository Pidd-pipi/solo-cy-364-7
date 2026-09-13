package model

import "time"

// SKU 商品主数据。
type SKU struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Spec      string    `gorm:"size:128" json:"spec"`
	Barcode   string    `gorm:"size:64;index" json:"barcode"`
	Category  string    `gorm:"size:64;index" json:"category"`
	Unit      string    `gorm:"size:16" json:"unit"`
	Status    string    `gorm:"size:16;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
