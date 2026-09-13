package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/dto"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/service"
	"github.com/ld/storeinventory/internal/util"
)

// TransferOrderHandler 调拨单接口处理器。
type TransferOrderHandler struct {
	transferSvc service.TransferOrderService
}

// NewTransferOrderHandler 构造调拨单处理器。
func NewTransferOrderHandler(transferSvc service.TransferOrderService) *TransferOrderHandler {
	return &TransferOrderHandler{transferSvc: transferSvc}
}

// Create 发起调拨申请。
func (h *TransferOrderHandler) Create(c *gin.Context) {
	var req dto.TransferCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	order, err := h.transferSvc.Create(req.FromStoreID, req.ToStoreID, req.SKUID, req.Quantity, req.Reason)
	if err != nil {
		c.Error(fmt.Errorf("handler create transfer: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgTransferCreateSuccess, order)
}

// List 调拨单列表。
func (h *TransferOrderHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	status := constants.TransferStatus(c.Query("status"))
	orders, total, err := h.transferSvc.List(q.Page, q.PageSize, uint(storeID), status)
	if err != nil {
		c.Error(fmt.Errorf("handler list transfers: %w", err))
		return
	}
	util.OK(c, gin.H{"list": orders, "total": total, "page": q.Page, "page_size": q.PageSize})
}

func (h *TransferOrderHandler) transition(c *gin.Context, action func(uint) (*model.TransferOrder, error), successMsg string) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的调拨单ID", err))
		return
	}
	order, err := action(uint(id))
	if err != nil {
		c.Error(fmt.Errorf("handler transfer transition: %w", err))
		return
	}
	util.OKMessage(c, successMsg, order)
}

// Confirm 审批确认。
func (h *TransferOrderHandler) Confirm(c *gin.Context) {
	h.transition(c, h.transferSvc.Confirm, constants.MsgTransferConfirmSuccess)
}

// Ship 发货。
func (h *TransferOrderHandler) Ship(c *gin.Context) {
	h.transition(c, h.transferSvc.Ship, constants.MsgTransferShipSuccess)
}

// Receive 确认收货。
func (h *TransferOrderHandler) Receive(c *gin.Context) {
	h.transition(c, h.transferSvc.Receive, constants.MsgTransferReceiveSuccess)
}

// Cancel 取消。
func (h *TransferOrderHandler) Cancel(c *gin.Context) {
	h.transition(c, h.transferSvc.Cancel, constants.MsgTransferCancelSuccess)
}
