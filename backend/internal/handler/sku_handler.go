package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/dto"
	"github.com/ld/storeinventory/internal/service"
	"github.com/ld/storeinventory/internal/util"
)

// SKUHandler SKU 主数据接口处理器。
type SKUHandler struct {
	skuSvc service.SKUService
}

// NewSKUHandler 构造 SKU 处理器。
func NewSKUHandler(skuSvc service.SKUService) *SKUHandler {
	return &SKUHandler{skuSvc: skuSvc}
}

// Create 新增 SKU。
func (h *SKUHandler) Create(c *gin.Context) {
	var req dto.SKUCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	sku, err := h.skuSvc.Create(req.Code, req.Name, req.Spec, req.Barcode, req.Category, req.Unit)
	if err != nil {
		c.Error(fmt.Errorf("handler create sku: %w", err))
		return
	}
	util.OK(c, sku)
}

// BatchImport 批量导入。
func (h *SKUHandler) BatchImport(c *gin.Context) {
	var req dto.SKUBatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	n, err := h.skuSvc.BatchImport(req.Items)
	if err != nil {
		c.Error(fmt.Errorf("handler batch import sku: %w", err))
		return
	}
	util.OKMessage(c, fmt.Sprintf("成功导入 %d 条 SKU", n), gin.H{"imported": n})
}

// Update 修改 SKU。
func (h *SKUHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的SKU ID", err))
		return
	}
	var req dto.SKUUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	sku, err := h.skuSvc.Update(uint(id), req.Code, req.Name, req.Spec, req.Barcode, req.Category, req.Unit)
	if err != nil {
		c.Error(fmt.Errorf("handler update sku: %w", err))
		return
	}
	util.OK(c, sku)
}

// Delete 删除 SKU。
func (h *SKUHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的SKU ID", err))
		return
	}
	if err := h.skuSvc.Delete(uint(id)); err != nil {
		c.Error(fmt.Errorf("handler delete sku: %w", err))
		return
	}
	util.OKMessage(c, "SKU 已删除", nil)
}

// List SKU 列表。
func (h *SKUHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	category := c.Query("category")
	keyword := c.Query("keyword")
	skus, total, err := h.skuSvc.List(q.Page, q.PageSize, category, keyword)
	if err != nil {
		c.Error(fmt.Errorf("handler list skus: %w", err))
		return
	}
	util.OK(c, gin.H{"list": skus, "total": total, "page": q.Page, "page_size": q.PageSize})
}
