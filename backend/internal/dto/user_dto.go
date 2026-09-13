package dto

import "github.com/ld/storeinventory/internal/constants"

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username string             `json:"username" binding:"required,min=3,max=64"`
	Password string             `json:"password" binding:"required,min=6,max=64"`
	Name     string             `json:"name" binding:"required,max=64"`
	Role     constants.UserRole `json:"role"`
	StoreID  *uint              `json:"store_id"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// UpdateProfileRequest 修改资料请求。
type UpdateProfileRequest struct {
	Name    string `json:"name" binding:"max=64"`
	StoreID *uint  `json:"store_id"`
}

// ListQuery 分页查询参数。
type ListQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=200"`
}

// Normalize 归一化分页参数。
func (q *ListQuery) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
}
