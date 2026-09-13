package service

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// transferTestDBSeq 为每个测试环境分配独立内存库，保证用例互不干扰、可重复/并行运行。
var transferTestDBSeq atomic.Int64

// transferTestEnv 基于内存 SQLite 的真实事务环境，覆盖分批收货全链路。
type transferTestEnv struct {
	db          *gorm.DB
	transferSvc TransferOrderService
	invSvc      StoreInventoryService
	recordSvc   StockRecordService
	orderRepo   repository.TransferOrderRepository
	fromStoreID uint
	toStoreID   uint
	skuID       uint
}

func newTransferTestEnv(t *testing.T, fromStock int) *transferTestEnv {
	t.Helper()
	dsn := fmt.Sprintf("file:transfer_svc_test_%d?mode=memory&cache=shared", transferTestDBSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	// 单连接串行化事务，并发用例行为确定
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
	if err := db.Create(from).Error; err != nil {
		t.Fatalf("seed from store: %v", err)
	}
	if err := db.Create(to).Error; err != nil {
		t.Fatalf("seed to store: %v", err)
	}
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("seed sku: %v", err)
	}
	if err := db.Create(&model.StoreInventory{StoreID: from.ID, SKUID: sku.ID, Quantity: fromStock}).Error; err != nil {
		t.Fatalf("seed inventory: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	invRepo := repository.NewStoreInventoryRepository(db)
	skuRepo := repository.NewSKURepository(db)
	storeRepo := repository.NewStoreRepository(db)
	orderRepo := repository.NewTransferOrderRepository(db)
	receiptRepo := repository.NewTransferReceiptRepository(db)
	recordRepo := repository.NewStockRecordRepository(db)
	invSvc := NewStoreInventoryService(invRepo, skuRepo, db, logger)
	recordSvc := NewStockRecordService(recordRepo, invRepo, storeRepo, skuRepo, invSvc, db, logger)
	transferSvc := NewTransferOrderService(orderRepo, receiptRepo, invSvc, recordSvc, db, logger)

	return &transferTestEnv{
		db: db, transferSvc: transferSvc, invSvc: invSvc, recordSvc: recordSvc, orderRepo: orderRepo,
		fromStoreID: from.ID, toStoreID: to.ID, skuID: sku.ID,
	}
}

// newShippedOrder 创建 qty 的调拨单并推进到已发货。
func (e *transferTestEnv) newShippedOrder(t *testing.T, qty int) *model.TransferOrder {
	t.Helper()
	order, err := e.transferSvc.Create(e.fromStoreID, e.toStoreID, e.skuID, qty, "测试调拨")
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	if _, err := e.transferSvc.Confirm(order.ID); err != nil {
		t.Fatalf("confirm transfer: %v", err)
	}
	if _, err := e.transferSvc.Ship(order.ID); err != nil {
		t.Fatalf("ship transfer: %v", err)
	}
	return order
}

func (e *transferTestEnv) invQty(t *testing.T, storeID uint) int {
	t.Helper()
	inv, err := e.invSvc.GetByStoreAndSKU(storeID, e.skuID)
	if err != nil {
		t.Fatalf("get inventory store[%d]: %v", storeID, err)
	}
	return inv.Quantity
}

func (e *transferTestEnv) transferInRecords(t *testing.T) []model.StockRecord {
	t.Helper()
	return e.recordsAt(t, e.toStoreID)
}

func (e *transferTestEnv) recordsAt(t *testing.T, storeID uint) []model.StockRecord {
	t.Helper()
	records, _, err := e.recordSvc.List(1, 100, storeID, e.skuID, constants.RecordTransferIn)
	if err != nil {
		t.Fatalf("list transfer in records store[%d]: %v", storeID, err)
	}
	return records
}

func (e *transferTestEnv) reload(t *testing.T, id uint) *model.TransferOrder {
	t.Helper()
	order, err := e.orderRepo.FindByID(id)
	if err != nil {
		t.Fatalf("reload order[%d]: %v", id, err)
	}
	return order
}

// TestTransferReceivePartialBatches 分批收货：未收满保持已发货，收满流转已收货，
// 每次按本次数量增加调入库存并生成调拨入库记录与收货明细。
func TestTransferReceivePartialBatches(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 10)
	if got := env.invQty(t, env.fromStoreID); got != 90 {
		t.Fatalf("from store qty=%d, want 90", got)
	}

	// 第一批收 4：保持已发货
	o1, err := env.transferSvc.Receive(order.ID, 4, "第一批")
	if err != nil {
		t.Fatalf("receive batch 1: %v", err)
	}
	if o1.Status != constants.TransferShipped {
		t.Fatalf("status=%s, want shipped", o1.Status)
	}
	if o1.ReceivedQuantity != 4 {
		t.Fatalf("received_quantity=%d, want 4", o1.ReceivedQuantity)
	}
	if got := env.invQty(t, env.toStoreID); got != 4 {
		t.Fatalf("to store qty=%d, want 4", got)
	}
	receipts, err := env.transferSvc.ListReceipts(order.ID)
	if err != nil {
		t.Fatalf("list receipts: %v", err)
	}
	if len(receipts) != 1 || receipts[0].Quantity != 4 || receipts[0].Remark != "第一批" {
		t.Fatalf("receipts=%+v, want 1 receipt qty=4 remark=第一批", receipts)
	}
	records := env.transferInRecords(t)
	if len(records) != 1 || records[0].Quantity != 4 || records[0].RelatedOrderID == nil || *records[0].RelatedOrderID != order.ID {
		t.Fatalf("records=%+v, want 1 transfer_in record qty=4 related=%d", records, order.ID)
	}

	// 第二批收 6：累计 10 收满，流转已收货
	o2, err := env.transferSvc.Receive(order.ID, 6, "第二批")
	if err != nil {
		t.Fatalf("receive batch 2: %v", err)
	}
	if o2.Status != constants.TransferReceived {
		t.Fatalf("status=%s, want received", o2.Status)
	}
	if o2.ReceivedQuantity != 10 {
		t.Fatalf("received_quantity=%d, want 10", o2.ReceivedQuantity)
	}
	if got := env.invQty(t, env.toStoreID); got != 10 {
		t.Fatalf("to store qty=%d, want 10", got)
	}
	if got := env.invQty(t, env.fromStoreID); got != 90 {
		t.Fatalf("from store qty=%d, want 90", got)
	}
	receipts, err = env.transferSvc.ListReceipts(order.ID)
	if err != nil {
		t.Fatalf("list receipts: %v", err)
	}
	if len(receipts) != 2 || receipts[1].Quantity != 6 || receipts[1].Remark != "第二批" {
		t.Fatalf("receipts=%+v, want 2 receipts, last qty=6 remark=第二批", receipts)
	}
	records = env.transferInRecords(t)
	if len(records) != 2 {
		t.Fatalf("records len=%d, want 2", len(records))
	}
	sum := 0
	for _, r := range records {
		sum += r.Quantity
	}
	if sum != 10 {
		t.Fatalf("records qty sum=%d, want 10", sum)
	}
}

// TestTransferReceiveExceed 超量收货：本次数量超过剩余数量必须失败且状态不变。
func TestTransferReceiveExceed(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 5)

	if _, err := env.transferSvc.Receive(order.ID, 6, ""); !errors.Is(err, util.ErrReceiveExceed) {
		t.Fatalf("receive 6/5 err=%v, want ErrReceiveExceed", err)
	}
	o := env.reload(t, order.ID)
	if o.ReceivedQuantity != 0 || o.Status != constants.TransferShipped {
		t.Fatalf("after exceed: received=%d status=%s, want 0/shipped", o.ReceivedQuantity, o.Status)
	}
	if _, err := env.invSvc.GetByStoreAndSKU(env.toStoreID, env.skuID); !errors.Is(err, util.ErrNotFound) {
		t.Fatalf("to store inventory should not exist, err=%v", err)
	}
	if receipts, _ := env.transferSvc.ListReceipts(order.ID); len(receipts) != 0 {
		t.Fatalf("receipts len=%d, want 0", len(receipts))
	}

	// 部分收货后再超量：剩余 2 收 3 失败，已收保持 3
	if _, err := env.transferSvc.Receive(order.ID, 3, "第一批"); err != nil {
		t.Fatalf("receive 3: %v", err)
	}
	if _, err := env.transferSvc.Receive(order.ID, 3, ""); !errors.Is(err, util.ErrReceiveExceed) {
		t.Fatalf("receive 3 remaining 2 err=%v, want ErrReceiveExceed", err)
	}
	o = env.reload(t, order.ID)
	if o.ReceivedQuantity != 3 || o.Status != constants.TransferShipped {
		t.Fatalf("after partial exceed: received=%d status=%s, want 3/shipped", o.ReceivedQuantity, o.Status)
	}
	if got := env.invQty(t, env.toStoreID); got != 3 {
		t.Fatalf("to store qty=%d, want 3", got)
	}
	if receipts, _ := env.transferSvc.ListReceipts(order.ID); len(receipts) != 1 {
		t.Fatalf("receipts len=%d, want 1", len(receipts))
	}
}

// TestTransferReceiveInvalidStatus 非法状态收货：未发货或已结束单据不能收货，
// 且失败的收货不改变库存、累计收货与收货明细。
func TestTransferReceiveInvalidStatus(t *testing.T) {
	env := newTransferTestEnv(t, 100)

	// 待确认
	pending, err := env.transferSvc.Create(env.fromStoreID, env.toStoreID, env.skuID, 5, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := env.transferSvc.Receive(pending.ID, 1, ""); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("receive pending err=%v, want ErrConflict", err)
	}

	// 已确认未发货
	confirmed, err := env.transferSvc.Create(env.fromStoreID, env.toStoreID, env.skuID, 5, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := env.transferSvc.Confirm(confirmed.ID); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if _, err := env.transferSvc.Receive(confirmed.ID, 1, ""); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("receive confirmed err=%v, want ErrConflict", err)
	}

	// 已取消
	cancelled, err := env.transferSvc.Create(env.fromStoreID, env.toStoreID, env.skuID, 5, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := env.transferSvc.Cancel(cancelled.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := env.transferSvc.Receive(cancelled.ID, 1, ""); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("receive cancelled err=%v, want ErrConflict", err)
	}

	// 已收货（收满后再次收货）
	full := env.newShippedOrder(t, 5)
	if _, err := env.transferSvc.Receive(full.ID, 5, ""); err != nil {
		t.Fatalf("receive full: %v", err)
	}
	if _, err := env.transferSvc.Receive(full.ID, 1, ""); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("receive received err=%v, want ErrConflict", err)
	}

	// 数量参数非法
	shipped := env.newShippedOrder(t, 5)
	if _, err := env.transferSvc.Receive(shipped.ID, 0, ""); !errors.Is(err, util.ErrValidation) {
		t.Fatalf("receive qty=0 err=%v, want ErrValidation", err)
	}
	if _, err := env.transferSvc.Receive(shipped.ID, -2, ""); !errors.Is(err, util.ErrValidation) {
		t.Fatalf("receive qty=-2 err=%v, want ErrValidation", err)
	}

	// 全部失败收货之后：只有 full 单的成功收货生效，其余单据库存/累计/明细均不变
	if got := env.invQty(t, env.fromStoreID); got != 90 {
		t.Fatalf("from store qty=%d, want 90 (only two shipped orders deducted)", got)
	}
	if got := env.invQty(t, env.toStoreID); got != 5 {
		t.Fatalf("to store qty=%d, want 5 (only full order received)", got)
	}
	wantReceived := map[uint]int{pending.ID: 0, confirmed.ID: 0, cancelled.ID: 0, full.ID: 5, shipped.ID: 0}
	for id, want := range wantReceived {
		o := env.reload(t, id)
		if o.ReceivedQuantity != want {
			t.Fatalf("order[%d] received_quantity=%d, want %d", id, o.ReceivedQuantity, want)
		}
	}
	for _, id := range []uint{pending.ID, confirmed.ID, cancelled.ID, shipped.ID} {
		if receipts, _ := env.transferSvc.ListReceipts(id); len(receipts) != 0 {
			t.Fatalf("order[%d] receipts len=%d, want 0", id, len(receipts))
		}
	}
	if receipts, _ := env.transferSvc.ListReceipts(full.ID); len(receipts) != 1 {
		t.Fatalf("full order receipts len=%d, want 1", len(receipts))
	}
	if records := env.recordsAt(t, env.toStoreID); len(records) != 1 || records[0].Quantity != 5 {
		t.Fatalf("to store records=%+v, want 1 record qty=5", records)
	}
}

// TestTransferReceiveDuplicateSubmission 重复提交：同一请求重复到达只允许生效一次。
func TestTransferReceiveDuplicateSubmission(t *testing.T) {
	env := newTransferTestEnv(t, 100)

	// 一次性收满后重复提交
	order := env.newShippedOrder(t, 8)
	if _, err := env.transferSvc.Receive(order.ID, 8, "整单收货"); err != nil {
		t.Fatalf("receive 8: %v", err)
	}
	if _, err := env.transferSvc.Receive(order.ID, 8, "重复提交"); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("duplicate receive err=%v, want ErrConflict", err)
	}
	if got := env.invQty(t, env.toStoreID); got != 8 {
		t.Fatalf("to store qty=%d, want 8 (no double count)", got)
	}
	if receipts, _ := env.transferSvc.ListReceipts(order.ID); len(receipts) != 1 {
		t.Fatalf("receipts len=%d, want 1", len(receipts))
	}
	if records := env.transferInRecords(t); len(records) != 1 {
		t.Fatalf("records len=%d, want 1", len(records))
	}

	// 部分收货后重复提交同批次：剩余不足，按超量拒绝
	order2 := env.newShippedOrder(t, 10)
	if _, err := env.transferSvc.Receive(order2.ID, 6, "第一批"); err != nil {
		t.Fatalf("receive 6: %v", err)
	}
	if _, err := env.transferSvc.Receive(order2.ID, 6, "重复提交"); !errors.Is(err, util.ErrReceiveExceed) {
		t.Fatalf("duplicate partial receive err=%v, want ErrReceiveExceed", err)
	}
	o := env.reload(t, order2.ID)
	if o.ReceivedQuantity != 6 {
		t.Fatalf("received_quantity=%d, want 6", o.ReceivedQuantity)
	}
	if got := env.invQty(t, env.toStoreID); got != 8+6 {
		t.Fatalf("to store qty=%d, want 14", got)
	}
	if receipts, _ := env.transferSvc.ListReceipts(order2.ID); len(receipts) != 1 {
		t.Fatalf("receipts len=%d, want 1", len(receipts))
	}
}

// TestTransferReceiveConcurrent 并发收货：两个并发批次只有一个能成功，库存与累计不多加。
func TestTransferReceiveConcurrent(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 10)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = env.transferSvc.Receive(order.ID, 6, "并发批次")
		}(i)
	}
	wg.Wait()

	success := 0
	for _, err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("concurrent receive success=%d, want exactly 1 (errs=%v)", success, errs)
	}
	o := env.reload(t, order.ID)
	if o.ReceivedQuantity != 6 || o.Status != constants.TransferShipped {
		t.Fatalf("received=%d status=%s, want 6/shipped", o.ReceivedQuantity, o.Status)
	}
	if got := env.invQty(t, env.toStoreID); got != 6 {
		t.Fatalf("to store qty=%d, want 6", got)
	}
	if receipts, _ := env.transferSvc.ListReceipts(order.ID); len(receipts) != 1 {
		t.Fatalf("receipts len=%d, want 1", len(receipts))
	}
	if records := env.transferInRecords(t); len(records) != 1 {
		t.Fatalf("records len=%d, want 1", len(records))
	}
}

// TestTransferReceiveInventoryAccumulation 库存累计：调入库存等于各批次之和，记录逐批对应。
func TestTransferReceiveInventoryAccumulation(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 9)

	batches := []int{2, 3, 4}
	for i, qty := range batches {
		if _, err := env.transferSvc.Receive(order.ID, qty, "批次"); err != nil {
			t.Fatalf("receive batch %d: %v", i, err)
		}
	}
	o := env.reload(t, order.ID)
	if o.Status != constants.TransferReceived || o.ReceivedQuantity != 9 {
		t.Fatalf("status=%s received=%d, want received/9", o.Status, o.ReceivedQuantity)
	}
	if got := env.invQty(t, env.toStoreID); got != 9 {
		t.Fatalf("to store qty=%d, want 9", got)
	}
	records := env.transferInRecords(t)
	if len(records) != len(batches) {
		t.Fatalf("records len=%d, want %d", len(records), len(batches))
	}
	sum := 0
	for _, r := range records {
		if r.RecordType != constants.RecordTransferIn {
			t.Fatalf("record type=%s, want transfer_in", r.RecordType)
		}
		sum += r.Quantity
	}
	if sum != 9 {
		t.Fatalf("records sum=%d, want 9", sum)
	}
	receipts, _ := env.transferSvc.ListReceipts(order.ID)
	if len(receipts) != len(batches) {
		t.Fatalf("receipts len=%d, want %d", len(receipts), len(batches))
	}
	for i, b := range batches {
		if receipts[i].Quantity != b {
			t.Fatalf("receipt[%d] qty=%d, want %d", i, receipts[i].Quantity, b)
		}
	}
}

// TestTransferCancelPartialReceived 部分收货后取消：未收部分退回调出门店并留痕，
// 已收部分留在调入门店，状态、库存与收货明细保持一致。
func TestTransferCancelPartialReceived(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 10)
	if _, err := env.transferSvc.Receive(order.ID, 4, "第一批"); err != nil {
		t.Fatalf("receive 4: %v", err)
	}

	cancelled, err := env.transferSvc.Cancel(order.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != constants.TransferCancelled {
		t.Fatalf("status=%s, want cancelled", cancelled.Status)
	}
	// 已收 4 件留在调入门店，累计收货不变，明细保留
	if cancelled.ReceivedQuantity != 4 {
		t.Fatalf("received_quantity=%d, want 4", cancelled.ReceivedQuantity)
	}
	if got := env.invQty(t, env.toStoreID); got != 4 {
		t.Fatalf("to store qty=%d, want 4", got)
	}
	// 未收 6 件退回调出门店：100 - 10(发货) + 6(退回) = 96
	if got := env.invQty(t, env.fromStoreID); got != 96 {
		t.Fatalf("from store qty=%d, want 96", got)
	}
	receipts, err := env.transferSvc.ListReceipts(order.ID)
	if err != nil {
		t.Fatalf("list receipts: %v", err)
	}
	if len(receipts) != 1 || receipts[0].Quantity != 4 {
		t.Fatalf("receipts=%+v, want 1 receipt qty=4", receipts)
	}
	// 调入门店 1 条收货记录（4），调出门店 1 条退回记录（6），均关联本单
	toRecords := env.recordsAt(t, env.toStoreID)
	if len(toRecords) != 1 || toRecords[0].Quantity != 4 {
		t.Fatalf("to store records=%+v, want 1 record qty=4", toRecords)
	}
	fromRecords := env.recordsAt(t, env.fromStoreID)
	if len(fromRecords) != 1 || fromRecords[0].Quantity != 6 || fromRecords[0].RelatedOrderID == nil || *fromRecords[0].RelatedOrderID != order.ID {
		t.Fatalf("from store records=%+v, want 1 return record qty=6 related=%d", fromRecords, order.ID)
	}
}

// TestTransferCancelFullyReceived 收满后不能取消。
func TestTransferCancelFullyReceived(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 5)
	if _, err := env.transferSvc.Receive(order.ID, 5, "整单收货"); err != nil {
		t.Fatalf("receive 5: %v", err)
	}
	if _, err := env.transferSvc.Cancel(order.ID); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("cancel received err=%v, want ErrConflict", err)
	}
	o := env.reload(t, order.ID)
	if o.Status != constants.TransferReceived || o.ReceivedQuantity != 5 {
		t.Fatalf("status=%s received=%d, want received/5", o.Status, o.ReceivedQuantity)
	}
	if got := env.invQty(t, env.fromStoreID); got != 95 {
		t.Fatalf("from store qty=%d, want 95 (no return)", got)
	}
	if got := env.invQty(t, env.toStoreID); got != 5 {
		t.Fatalf("to store qty=%d, want 5", got)
	}
	if records := env.recordsAt(t, env.fromStoreID); len(records) != 0 {
		t.Fatalf("from store records len=%d, want 0 (no return record)", len(records))
	}
}

// TestTransferCancelShippedNoReceipts 已发货未收货取消：全部数量退回调出门店。
func TestTransferCancelShippedNoReceipts(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 7)
	if got := env.invQty(t, env.fromStoreID); got != 93 {
		t.Fatalf("from store qty=%d, want 93 after ship", got)
	}
	if _, err := env.transferSvc.Cancel(order.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if got := env.invQty(t, env.fromStoreID); got != 100 {
		t.Fatalf("from store qty=%d, want 100 (full return)", got)
	}
	if _, err := env.invSvc.GetByStoreAndSKU(env.toStoreID, env.skuID); !errors.Is(err, util.ErrNotFound) {
		t.Fatalf("to store inventory should not exist, err=%v", err)
	}
	fromRecords := env.recordsAt(t, env.fromStoreID)
	if len(fromRecords) != 1 || fromRecords[0].Quantity != 7 {
		t.Fatalf("from store records=%+v, want 1 return record qty=7", fromRecords)
	}
	// 重复取消 → 冲突
	if _, err := env.transferSvc.Cancel(order.ID); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("duplicate cancel err=%v, want ErrConflict", err)
	}
	if got := env.invQty(t, env.fromStoreID); got != 100 {
		t.Fatalf("from store qty=%d after duplicate cancel, want 100", got)
	}
}

// TestTransferCancelBeforeShip 未发货取消：只流转状态，不动库存。
func TestTransferCancelBeforeShip(t *testing.T) {
	env := newTransferTestEnv(t, 100)

	pending, err := env.transferSvc.Create(env.fromStoreID, env.toStoreID, env.skuID, 5, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := env.transferSvc.Cancel(pending.ID); err != nil {
		t.Fatalf("cancel pending: %v", err)
	}
	confirmed, err := env.transferSvc.Create(env.fromStoreID, env.toStoreID, env.skuID, 5, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := env.transferSvc.Confirm(confirmed.ID); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	cancelled, err := env.transferSvc.Cancel(confirmed.ID)
	if err != nil {
		t.Fatalf("cancel confirmed: %v", err)
	}
	if cancelled.Status != constants.TransferCancelled {
		t.Fatalf("status=%s, want cancelled", cancelled.Status)
	}
	if got := env.invQty(t, env.fromStoreID); got != 100 {
		t.Fatalf("from store qty=%d, want 100 (untouched)", got)
	}
	if records := env.recordsAt(t, env.fromStoreID); len(records) != 0 {
		t.Fatalf("from store records len=%d, want 0", len(records))
	}
	// 已取消不能收货
	if _, err := env.transferSvc.Receive(pending.ID, 1, ""); !errors.Is(err, util.ErrConflict) {
		t.Fatalf("receive cancelled err=%v, want ErrConflict", err)
	}
}

// TestTransferReceiveVsCancelConcurrent 收货与取消并发：只有一个能成功，终态一致。
func TestTransferReceiveVsCancelConcurrent(t *testing.T) {
	env := newTransferTestEnv(t, 100)
	order := env.newShippedOrder(t, 10)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); _, errs[0] = env.transferSvc.Receive(order.ID, 10, "并发收货") }()
	go func() { defer wg.Done(); _, errs[1] = env.transferSvc.Cancel(order.ID) }()
	wg.Wait()

	success := 0
	for _, err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("receive vs cancel success=%d, want exactly 1 (errs=%v)", success, errs)
	}
	o := env.reload(t, order.ID)
	switch o.Status {
	case constants.TransferReceived:
		// 收货先成功：调入 10，调出 90，无退回
		if got := env.invQty(t, env.toStoreID); got != 10 {
			t.Fatalf("to store qty=%d, want 10", got)
		}
		if got := env.invQty(t, env.fromStoreID); got != 90 {
			t.Fatalf("from store qty=%d, want 90", got)
		}
	case constants.TransferCancelled:
		// 取消先成功：10 件全部退回，调入无库存
		if got := env.invQty(t, env.fromStoreID); got != 100 {
			t.Fatalf("from store qty=%d, want 100", got)
		}
		if _, err := env.invSvc.GetByStoreAndSKU(env.toStoreID, env.skuID); !errors.Is(err, util.ErrNotFound) {
			t.Fatalf("to store inventory should not exist, err=%v", err)
		}
	default:
		t.Fatalf("status=%s, want received or cancelled", o.Status)
	}
}
