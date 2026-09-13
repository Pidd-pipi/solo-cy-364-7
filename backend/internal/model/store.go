package model

import "time"

// Store 门店。
type Store struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Code          string    `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name          string    `gorm:"size:128;not null" json:"name"`
	Address       string    `gorm:"size:255" json:"address"`
	ManagerUserID *uint     `gorm:"index" json:"manager_user_id"`
	Manager       *User     `gorm:"foreignKey:ManagerUserID" json:"manager,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
