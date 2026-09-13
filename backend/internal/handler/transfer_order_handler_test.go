package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/middleware"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/service"
)

// transferAPITestDBSeq 为每个测试分配独立内存库，保证可重复运行。
var transferAPITestDBSeq atomic.Int64

// newTransferAPITestEnv 用内存 SQLite 装配真实 service 栈 + gin 引擎，做接口级冒烟。
func newTransferAPITestEnv(t *testing.T) (*gin.Engine, service.TransferOrderService, uint) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:transfer_api_test_%d?mode=memory&cache=shared", transferAPITestDBSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(
		&model.User{}, &model.Store{}, &model.SKU{}, &model.StoreInventory{},
		&model.TransferOrder{}, &model.TransferReceipt{}, &model.StockRecord{}, &model.Stocktake{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	from := &model.Store{Code: "S001", Name: "调出门店"}
	to := &model.Store{Code: "S002", Name: "调入门店"}
	sku := &model.SKU{Code: "SKU001", Name: "测试商品"}
	for _, v := range []interface{}{from, to, sku} {
		if err := db.Create(v).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	if err := db.Create(&model.StoreInventory{StoreID: from.ID, SKUID: sku.ID, Quantity: 100}).Error; err != nil {
		t.Fatalf("seed inventory: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	invRepo := repository.NewStoreInventoryRepository(db)
	skuRepo := repository.NewSKURepository(db)
	storeRepo := repository.NewStoreRepository(db)
	orderRepo := repository.NewTransferOrderRepository(db)
	receiptRepo := repository.NewTransferReceiptRepository(db)
	recordRepo := repository.NewStockRecordRepository(db)
	invSvc := service.NewStoreInventoryService(invRepo, skuRepo, db, logger)
	recordSvc := service.NewStockRecordService(recordRepo, invRepo, storeRepo, skuRepo, invSvc, db, logger)
	transferSvc := service.NewTransferOrderService(orderRepo, receiptRepo, invSvc, recordSvc, db, logger)

	h := NewTransferOrderHandler(transferSvc)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.PUT("/api/v1/transfers/:id/receive", h.Receive)
	r.GET("/api/v1/transfers/:id/receipts", h.ListReceipts)

	order, err := transferSvc.Create(from.ID, to.ID, sku.ID, 10, "接口测试")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := transferSvc.Confirm(order.ID); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if _, err := transferSvc.Ship(order.ID); err != nil {
		t.Fatalf("ship: %v", err)
	}
	return r, transferSvc, order.ID
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) (int, map[string]interface{}) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response %q: %v", w.Body.String(), err)
	}
	return w.Code, resp
}

// TestReceiveAPI 分批收货接口：正常批次返回累计进度，响应文案与当前状态一致，超量返回业务错误文案。
func TestReceiveAPI(t *testing.T) {
	r, _, orderID := newTransferAPITestEnv(t)

	code, resp := doJSON(t, r, http.MethodPut, "/api/v1/transfers/"+itoa(orderID)+"/receive", map[string]interface{}{
		"quantity": 4, "remark": "第一批",
	})
	if code != http.StatusOK || resp["code"].(float64) != 0 {
		t.Fatalf("receive status=%d resp=%v", code, resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["received_quantity"].(float64) != 4 || data["status"].(string) != "shipped" {
		t.Fatalf("data=%v, want received_quantity=4 status=shipped", data)
	}
	// 未收满：文案应为部分收货，而非"调拨单已收货"
	if resp["message"].(string) != "本次收货成功，调拨单未收满" {
		t.Fatalf("partial receive message=%v", resp["message"])
	}

	// 超量：剩余 6 收 7 → 400 + 超量文案
	code, resp = doJSON(t, r, http.MethodPut, "/api/v1/transfers/"+itoa(orderID)+"/receive", map[string]interface{}{
		"quantity": 7,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("exceed status=%d, want 400 (resp=%v)", code, resp)
	}
	if resp["message"].(string) != "累计收货数量不能超过调拨数量" {
		t.Fatalf("exceed message=%v", resp["message"])
	}

	// 参数非法：缺 quantity → 422
	code, _ = doJSON(t, r, http.MethodPut, "/api/v1/transfers/"+itoa(orderID)+"/receive", map[string]interface{}{})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid body status=%d, want 422", code)
	}

	// 收满：文案应为已收货
	code, resp = doJSON(t, r, http.MethodPut, "/api/v1/transfers/"+itoa(orderID)+"/receive", map[string]interface{}{
		"quantity": 6, "remark": "第二批",
	})
	if code != http.StatusOK || resp["code"].(float64) != 0 {
		t.Fatalf("receive batch 2 status=%d resp=%v", code, resp)
	}
	if resp["message"].(string) != "调拨单已收货" {
		t.Fatalf("full receive message=%v", resp["message"])
	}
	data = resp["data"].(map[string]interface{})
	if data["status"].(string) != "received" {
		t.Fatalf("status=%v, want received", data["status"])
	}

	// 收满后明细可查
	code, resp = doJSON(t, r, http.MethodGet, "/api/v1/transfers/"+itoa(orderID)+"/receipts", nil)
	if code != http.StatusOK {
		t.Fatalf("receipts status=%d", code)
	}
	list := resp["data"].(map[string]interface{})["list"].([]interface{})
	if len(list) != 2 {
		t.Fatalf("receipts len=%d, want 2", len(list))
	}
	first := list[0].(map[string]interface{})
	if first["quantity"].(float64) != 4 || first["remark"].(string) != "第一批" {
		t.Fatalf("receipt[0]=%v, want qty=4 remark=第一批", first)
	}
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
