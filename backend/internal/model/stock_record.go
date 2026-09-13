package model

import (
	"time"

	"github.com/ld/storeinventory/internal/constants"
)

// StockRecord 出入库记录。
type StockRecord struct {
	ID             uint                      `gorm:"primaryKey" json:"id"`
	StoreID        uint                      `gorm:"index;not null" json:"store_id"`
	Store          *Store                    `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	SKUID          uint                      `gorm:"column:sku_id;index;not null" json:"sku_id"`
	SKU            *SKU                      `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	RecordType     constants.StockRecordType `gorm:"size:16;index;not null" json:"record_type"`
	Quantity       int                       `gorm:"not null" json:"quantity"`
	RelatedOrderID *uint                     `gorm:"index" json:"related_order_id"`
	CreatedAt      time.Time                 `json:"created_at"`
}
