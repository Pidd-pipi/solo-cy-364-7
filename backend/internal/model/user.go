package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
)

// User 用户/员工：总部维护 SKU 主数据，店长管理所属 Store 的库存与调拨。
type User struct {
	ID           uint               `gorm:"primaryKey" json:"id"`
	Username     string             `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string             `gorm:"size:255;not null" json:"-"`
	Name         string             `gorm:"size:64" json:"name"`
	Role         constants.UserRole `gorm:"size:32;index;not null" json:"role"`
	StoreID      *uint              `gorm:"index" json:"store_id"`
	Store        *Store             `gorm:"foreignKey:StoreID" json:"store,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// BeforeCreate 校验角色合法性。
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if !u.Role.Valid() {
		return constants.ErrInvalidUserRole
	}
	return nil
}
