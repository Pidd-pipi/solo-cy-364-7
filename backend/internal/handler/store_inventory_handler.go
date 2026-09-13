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

// StoreInventoryHandler 门店库存接口处理器。
type StoreInventoryHandler struct {
	invSvc service.StoreInventoryService
}

// NewStoreInventoryHandler 构造门店库存处理器。
func NewStoreInventoryHandler(invSvc service.StoreInventoryService) *StoreInventoryHandler {
	return &StoreInventoryHandler{invSvc: invSvc}
}

// Ensure 初始化门店库存（采购入库时首次建档）。
func (h *StoreInventoryHandler) Ensure(c *gin.Context) {
	var req dto.InventoryEnsureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	inv, err := h.invSvc.Ensure(req.StoreID, req.SKUID, req.Quantity)
	if err != nil {
		c.Error(fmt.Errorf("handler ensure inventory: %w", err))
		return
	}
	util.OK(c, inv)
}

// SetSafetyStock 设置安全库存阈值。
func (h *StoreInventoryHandler) SetSafetyStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的库存ID", err))
		return
	}
	var req dto.SafetyStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	inv, err := h.invSvc.SetSafetyStock(uint(id), req.SafetyStock)
	if err != nil {
		c.Error(fmt.Errorf("handler set safety stock: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgSafetyStockUpdated, inv)
}

// List 门店库存列表。
func (h *StoreInventoryHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	skuID, _ := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	invs, total, err := h.invSvc.List(q.Page, q.PageSize, uint(storeID), uint(skuID))
	if err != nil {
		c.Error(fmt.Errorf("handler list inventories: %w", err))
		return
	}
	util.OK(c, gin.H{"list": invs, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// Alerts 低库存预警。
func (h *StoreInventoryHandler) Alerts(c *gin.Context) {
	alerts, err := h.invSvc.ListAlerts()
	if err != nil {
		c.Error(fmt.Errorf("handler inventory alerts: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgInventoryLowWarning, alerts)
}

// Stats 库存总览统计。
func (h *StoreInventoryHandler) Stats(c *gin.Context) {
	stats, err := h.invSvc.Stats()
	if err != nil {
		c.Error(fmt.Errorf("handler inventory stats: %w", err))
		return
	}
	util.OK(c, stats)
}
