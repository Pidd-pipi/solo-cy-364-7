package dto

import "github.com/ld/storeinventory/internal/service"

// SKUCreateRequest SKU 创建请求。
type SKUCreateRequest struct {
	Code     string `json:"code" binding:"required,max=64"`
	Name     string `json:"name" binding:"required,max=128"`
	Spec     string `json:"spec" binding:"max=128"`
	Barcode  string `json:"barcode" binding:"max=64"`
	Category string `json:"category" binding:"max=64"`
	Unit     string `json:"unit" binding:"max=16"`
}

// SKUUpdateRequest SKU 更新请求。
type SKUUpdateRequest struct {
	Code     string `json:"code" binding:"max=64"`
	Name     string `json:"name" binding:"max=128"`
	Spec     string `json:"spec" binding:"max=128"`
	Barcode  string `json:"barcode" binding:"max=64"`
	Category string `json:"category" binding:"max=64"`
	Unit     string `json:"unit" binding:"max=16"`
}

// SKUBatchImportRequest 批量导入请求。
type SKUBatchImportRequest struct {
	Items []service.SKUImportItem `json:"items" binding:"required,min=1,max=500"`
}
