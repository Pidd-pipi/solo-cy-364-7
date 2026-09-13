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

// StoreHandler 门店接口处理器。
type StoreHandler struct {
	storeSvc service.StoreService
}

// NewStoreHandler 构造门店处理器。
func NewStoreHandler(storeSvc service.StoreService) *StoreHandler {
	return &StoreHandler{storeSvc: storeSvc}
}

// Create 门店建档。
func (h *StoreHandler) Create(c *gin.Context) {
	var req dto.StoreCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	store, err := h.storeSvc.Create(req.Code, req.Name, req.Address, req.ManagerUserID)
	if err != nil {
		c.Error(fmt.Errorf("handler create store: %w", err))
		return
	}
	util.OK(c, store)
}

// Update 更新门店。
func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的门店ID", err))
		return
	}
	var req dto.StoreUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	store, err := h.storeSvc.Update(uint(id), req.Code, req.Name, req.Address, req.ManagerUserID)
	if err != nil {
		c.Error(fmt.Errorf("handler update store: %w", err))
		return
	}
	util.OK(c, store)
}

// Delete 删除门店。
func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的门店ID", err))
		return
	}
	if err := h.storeSvc.Delete(uint(id)); err != nil {
		c.Error(fmt.Errorf("handler delete store: %w", err))
		return
	}
	util.OKMessage(c, "门店已删除", nil)
}

// Get 门店详情。
func (h *StoreHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的门店ID", err))
		return
	}
	store, err := h.storeSvc.GetByID(uint(id))
	if err != nil {
		c.Error(fmt.Errorf("handler get store: %w", err))
		return
	}
	util.OK(c, store)
}

// List 门店列表。
func (h *StoreHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	stores, total, err := h.storeSvc.List(q.Page, q.PageSize)
	if err != nil {
		c.Error(fmt.Errorf("handler list stores: %w", err))
		return
	}
	util.OK(c, gin.H{"list": stores, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// ListAll 全量门店（下拉选项使用）。
func (h *StoreHandler) ListAll(c *gin.Context) {
	stores, err := h.storeSvc.ListAll()
	if err != nil {
		c.Error(fmt.Errorf("handler list all stores: %w", err))
		return
	}
	util.OK(c, stores)
}
