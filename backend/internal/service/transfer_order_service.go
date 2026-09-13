package service

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// TransferOrderService 调拨单业务逻辑。
type TransferOrderService interface {
	Create(fromStoreID, toStoreID, skuID uint, quantity int, reason string) (*model.TransferOrder, error)
	List(page, pageSize int, storeID uint, status constants.TransferStatus) ([]model.TransferOrder, int64, error)
	Confirm(id uint) (*model.TransferOrder, error)
	Ship(id uint) (*model.TransferOrder, error)
	Receive(id uint, quantity int, remark string) (*model.TransferOrder, error)
	ListReceipts(orderID uint) ([]model.TransferReceipt, error)
	Cancel(id uint) (*model.TransferOrder, error)
}

type transferOrderService struct {
	orderRepo   repository.TransferOrderRepository
	receiptRepo repository.TransferReceiptRepository
	invSvc      StoreInventoryService
	recSvc      StockRecordService
	db          *gorm.DB
	logger      *slog.Logger
}

// NewTransferOrderService 构造调拨单服务。
func NewTransferOrderService(orderRepo repository.TransferOrderRepository, receiptRepo repository.TransferReceiptRepository, invSvc StoreInventoryService, recSvc StockRecordService, db *gorm.DB, logger *slog.Logger) TransferOrderService {
	return &transferOrderService{orderRepo: orderRepo, receiptRepo: receiptRepo, invSvc: invSvc, recSvc: recSvc, db: db, logger: logger}
}

func (s *transferOrderService) Create(fromStoreID, toStoreID, skuID uint, quantity int, reason string) (*model.TransferOrder, error) {
	if fromStoreID == toStoreID {
		return nil, fmt.Errorf("create transfer from[%d] to[%d] same store: %w", fromStoreID, toStoreID, util.ErrValidation)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("create transfer quantity[%d]: %w", quantity, util.ErrValidation)
	}
	if err := s.invSvc.CheckSufficient(fromStoreID, skuID, quantity); err != nil {
		return nil, fmt.Errorf("create transfer from[%d] sku[%d]: %w", fromStoreID, skuID, err)
	}
	order := &model.TransferOrder{
		FromStoreID: fromStoreID, ToStoreID: toStoreID, SKUID: skuID,
		Quantity: quantity, Reason: reason, Status: constants.TransferPending,
	}
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}
	s.logger.Info(constants.LogTransferCreateSuccess, "order_id", order.ID, "from", fromStoreID, "to", toStoreID, "sku", skuID, "qty", quantity)
	return order, nil
}

func (s *transferOrderService) List(page, pageSize int, storeID uint, status constants.TransferStatus) ([]model.TransferOrder, int64, error) {
	orders, total, err := s.orderRepo.List(page, pageSize, storeID, status)
	if err != nil {
		return nil, 0, fmt.Errorf("list transfers: %w", err)
	}
	s.logger.Info(constants.LogTransferListQueried, "total", total)
	return orders, total, nil
}

func (s *transferOrderService) Confirm(id uint) (*model.TransferOrder, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("confirm transfer[id=%d]: %w", id, err)
	}
	if !constants.CanTransfer(order.Status, constants.TransferConfirmed) {
		s.logger.Warn(constants.LogTransferStatusInvalid, "order_id", id, "status", order.Status, "action", "confirm")
		return nil, fmt.Errorf("confirm transfer[id=%d] status[%s]: %w", id, order.Status, util.ErrConflict)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.orderRepo.TransitionStatusTx(tx, id, order.Status, constants.TransferConfirmed); err != nil {
			return fmt.Errorf("confirm transfer[id=%d] status[%s]: %w", id, order.Status, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	order.Status = constants.TransferConfirmed
	s.logger.Info(constants.LogTransferConfirmSuccess, "order_id", id)
	return order, nil
}

func (s *transferOrderService) Ship(id uint) (*model.TransferOrder, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("ship transfer[id=%d]: %w", id, err)
	}
	if !constants.CanTransfer(order.Status, constants.TransferShipped) {
		s.logger.Warn(constants.LogTransferStatusInvalid, "order_id", id, "status", order.Status, "action", "ship")
		return nil, fmt.Errorf("ship transfer[id=%d] status[%s]: %w", id, order.Status, util.ErrConflict)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.invSvc.CheckSufficientTx(tx, order.FromStoreID, order.SKUID, order.Quantity); err != nil {
			return fmt.Errorf("ship transfer[id=%d]: %w", id, err)
		}
		if err := s.invSvc.AdjustQuantityTx(tx, order.FromStoreID, order.SKUID, -order.Quantity); err != nil {
			return fmt.Errorf("ship transfer[id=%d]: %w", id, err)
		}
		if err := s.orderRepo.TransitionStatusTx(tx, id, order.Status, constants.TransferShipped); err != nil {
			return fmt.Errorf("ship transfer[id=%d] status[%s]: %w", id, order.Status, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	order.Status = constants.TransferShipped
	s.logger.Info(constants.LogTransferShipSuccess, "order_id", id, "from_store", order.FromStoreID, "sku", order.SKUID, "qty", order.Quantity)
	return order, nil
}

// Receive 分批收货：每次按本次实收数量增加调入门店库存、生成调拨入库记录与收货明细；
// 累计达到调拨数量时状态流转为已收货，未收满保持已发货；累计不允许超过调拨数量。
func (s *transferOrderService) Receive(id uint, quantity int, remark string) (*model.TransferOrder, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("receive transfer[id=%d] quantity[%d]: %w", id, quantity, util.ErrValidation)
	}
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("receive transfer[id=%d]: %w", id, err)
	}
	if order.Status != constants.TransferShipped {
		s.logger.Warn(constants.LogTransferStatusInvalid, "order_id", id, "status", order.Status, "action", "receive")
		return nil, fmt.Errorf("receive transfer[id=%d] status[%s]: %w", id, order.Status, util.ErrConflict)
	}
	if remaining := order.Quantity - order.ReceivedQuantity; quantity > remaining {
		s.logger.Warn(constants.LogTransferReceiveExceed, "order_id", id, "qty", quantity, "remaining", remaining)
		return nil, fmt.Errorf("receive transfer[id=%d] qty[%d] exceed remaining[%d]: %w", id, quantity, remaining, util.ErrReceiveExceed)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 事务内重读，基于最新累计值校验，防止并发收货超收
		fresh, err := s.orderRepo.FindByIDTx(tx, id)
		if err != nil {
			return fmt.Errorf("receive transfer[id=%d]: %w", id, err)
		}
		if fresh.Status != constants.TransferShipped {
			return fmt.Errorf("receive transfer[id=%d] status[%s]: %w", id, fresh.Status, util.ErrConflict)
		}
		if fresh.ReceivedQuantity+quantity > fresh.Quantity {
			return fmt.Errorf("receive transfer[id=%d] qty[%d] exceed remaining[%d]: %w", id, quantity, fresh.Quantity-fresh.ReceivedQuantity, util.ErrReceiveExceed)
		}
		// 原子累计已收数量并置状态（防重复提交/并发超收的最后防线），失败则整体回滚
		if err := s.orderRepo.AddReceivedQuantityTx(tx, id, quantity); err != nil {
			return fmt.Errorf("receive transfer[id=%d] add received qty[%d]: %w", id, quantity, err)
		}
		if _, err := s.invSvc.EnsureTx(tx, fresh.ToStoreID, fresh.SKUID, 0); err != nil {
			return fmt.Errorf("receive transfer[id=%d]: %w", id, err)
		}
		if _, err := s.recSvc.CreateTx(tx, fresh.ToStoreID, fresh.SKUID, constants.RecordTransferIn, quantity, &id); err != nil {
			return fmt.Errorf("receive transfer[id=%d] create record: %w", id, err)
		}
		receipt := &model.TransferReceipt{TransferOrderID: id, Quantity: quantity, Remark: remark}
		if err := s.receiptRepo.CreateTx(tx, receipt); err != nil {
			return fmt.Errorf("receive transfer[id=%d] create receipt: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("receive transfer[id=%d] reload: %w", id, err)
	}
	s.logger.Info(constants.LogTransferReceiveSuccess, "order_id", id, "to_store", updated.ToStoreID, "qty", quantity, "received", updated.ReceivedQuantity, "total", updated.Quantity, "status", updated.Status)
	return updated, nil
}

// ListReceipts 查询调拨单的收货明细（按收货时间升序）。
func (s *transferOrderService) ListReceipts(orderID uint) ([]model.TransferReceipt, error) {
	if _, err := s.orderRepo.FindByID(orderID); err != nil {
		return nil, fmt.Errorf("list receipts transfer[id=%d]: %w", orderID, err)
	}
	receipts, err := s.receiptRepo.ListByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("list receipts transfer[id=%d]: %w", orderID, err)
	}
	s.logger.Info(constants.LogTransferReceiptListed, "order_id", orderID, "count", len(receipts))
	return receipts, nil
}

// Cancel 取消调拨单。已发货未收满的单据取消时，未收部分退回调出门店并生成调拨入库记录留痕，
// 已收部分留在调入门店（收货明细保留）；收满（已收货）或已取消的单据不允许取消。
func (s *transferOrderService) Cancel(id uint) (*model.TransferOrder, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("cancel transfer[id=%d]: %w", id, err)
	}
	if !constants.CanTransfer(order.Status, constants.TransferCancelled) {
		s.logger.Warn(constants.LogTransferStatusInvalid, "order_id", id, "status", order.Status, "action", "cancel")
		return nil, fmt.Errorf("cancel transfer[id=%d] status[%s]: %w", id, order.Status, util.ErrConflict)
	}
	returnQty := 0
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 事务内重读，基于最新状态与累计收货计算退回数量
		fresh, err := s.orderRepo.FindByIDTx(tx, id)
		if err != nil {
			return fmt.Errorf("cancel transfer[id=%d]: %w", id, err)
		}
		if !constants.CanTransfer(fresh.Status, constants.TransferCancelled) {
			return fmt.Errorf("cancel transfer[id=%d] status[%s]: %w", id, fresh.Status, util.ErrConflict)
		}
		// 先认领状态（条件更新），防止并发收货/重复取消
		if err := s.orderRepo.TransitionStatusTx(tx, id, fresh.Status, constants.TransferCancelled); err != nil {
			return fmt.Errorf("cancel transfer[id=%d] status[%s]: %w", id, fresh.Status, err)
		}
		// 已发货取消：未收部分退回调出门店并留痕；已收部分留在调入门店不动
		if fresh.Status == constants.TransferShipped {
			returnQty = fresh.Quantity - fresh.ReceivedQuantity
			if returnQty > 0 {
				if _, err := s.recSvc.CreateTx(tx, fresh.FromStoreID, fresh.SKUID, constants.RecordTransferIn, returnQty, &id); err != nil {
					return fmt.Errorf("cancel transfer[id=%d] return qty[%d]: %w", id, returnQty, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("cancel transfer[id=%d] reload: %w", id, err)
	}
	s.logger.Info(constants.LogTransferCancelSuccess, "order_id", id, "from_status", order.Status, "returned_qty", returnQty)
	return updated, nil
}
