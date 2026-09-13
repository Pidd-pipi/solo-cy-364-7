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

// StockRecordHandler 出入库记录接口处理器。
type StockRecordHandler struct {
	recordSvc service.StockRecordService
}

// NewStockRecordHandler 构造出入库记录处理器。
func NewStockRecordHandler(recordSvc service.StockRecordService) *StockRecordHandler {
	return &StockRecordHandler{recordSvc: recordSvc}
}

// Create 创建出入库记录。
func (h *StockRecordHandler) Create(c *gin.Context) {
	var req dto.StockRecordCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	record, err := h.recordSvc.Create(req.StoreID, req.SKUID, req.RecordType, req.Quantity, nil)
	if err != nil {
		c.Error(fmt.Errorf("handler create stock record: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgRecordCreatedSuccess, record)
}

// List 出入库明细。
func (h *StockRecordHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	skuID, _ := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	recordType := constants.StockRecordType(c.Query("record_type"))
	records, total, err := h.recordSvc.List(q.Page, q.PageSize, uint(storeID), uint(skuID), recordType)
	if err != nil {
		c.Error(fmt.Errorf("handler list stock records: %w", err))
		return
	}
	util.OK(c, gin.H{"list": records, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// Export 出入库记录导出。
func (h *StockRecordHandler) Export(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	records, err := h.recordSvc.Export(uint(storeID))
	if err != nil {
		c.Error(fmt.Errorf("handler export stock records: %w", err))
		return
	}
	util.OK(c, records)
}

// CreateStocktake 周期盘点。
func (h *StockRecordHandler) CreateStocktake(c *gin.Context) {
	var req dto.StocktakeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	st, err := h.recordSvc.CreateStocktake(req.StoreID, req.SKUID, req.StocktakeDate, req.ActualQty, req.Remark)
	if err != nil {
		c.Error(fmt.Errorf("handler create stocktake: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgStocktakeSuccess, st)
}

// ListStocktakes 盘点记录。
func (h *StockRecordHandler) ListStocktakes(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	list, total, err := h.recordSvc.ListStocktakes(q.Page, q.PageSize, uint(storeID))
	if err != nil {
		c.Error(fmt.Errorf("handler list stocktakes: %w", err))
		return
	}
	util.OK(c, gin.H{"list": list, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// Suggestions 补货建议报表。
func (h *StockRecordHandler) Suggestions(c *gin.Context) {
	suggestions, err := h.recordSvc.ReplenishSuggestions()
	if err != nil {
		c.Error(fmt.Errorf("handler replenish suggestions: %w", err))
		return
	}
	util.OK(c, suggestions)
}
