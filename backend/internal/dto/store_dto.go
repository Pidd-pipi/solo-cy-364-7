package dto

// StoreCreateRequest 门店创建请求。
type StoreCreateRequest struct {
	Code          string `json:"code" binding:"required,max=32"`
	Name          string `json:"name" binding:"required,max=128"`
	Address       string `json:"address" binding:"max=255"`
	ManagerUserID *uint  `json:"manager_user_id"`
}

// StoreUpdateRequest 门店更新请求。
type StoreUpdateRequest struct {
	Code          string `json:"code" binding:"max=32"`
	Name          string `json:"name" binding:"max=128"`
	Address       string `json:"address" binding:"max=255"`
	ManagerUserID *uint  `json:"manager_user_id"`
}
