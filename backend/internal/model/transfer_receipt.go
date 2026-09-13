package model

import (
	"time"
)

// TransferReceipt 调拨收货明细：每次分批收货一条记录，关联调拨单。
type TransferReceipt struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TransferOrderID uint           `gorm:"index;not null" json:"transfer_order_id"`
	TransferOrder   *TransferOrder `gorm:"foreignKey:TransferOrderID" json:"transfer_order,omitempty"`
	Quantity        int            `gorm:"not null" json:"quantity"`
	Remark          string         `gorm:"size:255" json:"remark"`
	CreatedAt       time.Time      `json:"created_at"`
}
