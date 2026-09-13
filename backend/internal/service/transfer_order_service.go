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
	Receive(id uint) (*model.TransferOrder, error)
	Cancel(id uint) (*model.TransferOrder, error)
}

type transferOrderService struct {
	orderRepo repository.TransferOrderRepository
	invSvc    StoreInventoryService
	recSvc    StockRecordService
	db        *gorm.DB
	logger    *slog.Logger
}

// NewTransferOrderService 构造调拨单服务。
func NewTransferOrderService(orderRepo repository.TransferOrderRepository, invSvc StoreInventoryService, recSvc StockRecordService, db *gorm.DB, logger *slog.Logger) TransferOrderService {
	return &transferOrderService{orderRepo: orderRepo, invSvc: invSvc, recSvc: recSvc, db: db, logger: logger}
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

func (s *transferOrderService) Receive(id uint) (*model.TransferOrder, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("receive transfer[id=%d]: %w", id, err)
	}
	if order.Status != constants.TransferShipped {
		s.logger.Warn(constants.LogTransferStatusInvalid, "order_id", id, "status", order.Status, "action", "receive")
		return nil, fmt.Errorf("receive transfer[id=%d] status[%s]: %w", id, order.Status, util.ErrConflict)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.invSvc.EnsureTx(tx, order.ToStoreID, order.SKUID, 0); err != nil {
			return fmt.Errorf("receive transfer[id=%d]: %w", id, err)
		}
		if _, err := s.recSvc.CreateTx(tx, order.ToStoreID, order.SKUID, constants.RecordTransferIn, order.Quantity, &id); err != nil {
			return fmt.Errorf("receive transfer[id=%d] create record: %w", id, err)
		}
		if err := s.orderRepo.TransitionStatusTx(tx, id, order.Status, constants.TransferReceived); err != nil {
			return fmt.Errorf("receive transfer[id=%d] status[%s]: %w", id, order.Status, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	order.Status = constants.TransferReceived
	s.logger.Info(constants.LogTransferReceiveSuccess, "order_id", id, "to_store", order.ToStoreID, "qty", order.Quantity)
	return order, nil
}

func (s *transferOrderService) Cancel(id uint) (*model.TransferOrder, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("cancel transfer[id=%d]: %w", id, err)
	}
	if !constants.CanTransfer(order.Status, constants.TransferCancelled) {
		s.logger.Warn(constants.LogTransferStatusInvalid, "order_id", id, "status", order.Status, "action", "cancel")
		return nil, fmt.Errorf("cancel transfer[id=%d] status[%s]: %w", id, order.Status, util.ErrConflict)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.orderRepo.TransitionStatusTx(tx, id, order.Status, constants.TransferCancelled); err != nil {
			return fmt.Errorf("cancel transfer[id=%d] status[%s]: %w", id, order.Status, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	order.Status = constants.TransferCancelled
	s.logger.Info(constants.LogTransferCancelSuccess, "order_id", id)
	return order, nil
}
