package service

import (
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// StoreInventoryService 门店库存业务逻辑。
type StoreInventoryService interface {
	Ensure(storeID, skuID uint, quantity int) (*model.StoreInventory, error)
	EnsureTx(tx *gorm.DB, storeID, skuID uint, quantity int) (*model.StoreInventory, error)
	SetSafetyStock(id uint, safetyStock int) (*model.StoreInventory, error)
	GetByStoreAndSKU(storeID, skuID uint) (*model.StoreInventory, error)
	List(page, pageSize int, storeID, skuID uint) ([]model.StoreInventory, int64, error)
	ListAlerts() ([]model.StoreInventory, error)
	Stats() (*InventoryStats, error)
	AdjustQuantity(storeID, skuID uint, delta int) error
	AdjustQuantityTx(tx *gorm.DB, storeID, skuID uint, delta int) error
	CheckSufficient(storeID, skuID uint, qty int) error
	CheckSufficientTx(tx *gorm.DB, storeID, skuID uint, qty int) error
}

// InventoryStats 库存总览统计。
type InventoryStats struct {
	TotalSKUCount   int64 `json:"total_sku_count"`
	TotalQuantity   int64 `json:"total_quantity"`
	LowStockCount   int64 `json:"low_stock_count"`
	OutOfStockCount int64 `json:"out_of_stock_count"`
	StoreCount      int64 `json:"store_count"`
}

type storeInventoryService struct {
	invRepo repository.StoreInventoryRepository
	skuRepo repository.SKURepository
	db      *gorm.DB
	logger  *slog.Logger
}

// NewStoreInventoryService 构造门店库存服务。
func NewStoreInventoryService(invRepo repository.StoreInventoryRepository, skuRepo repository.SKURepository, db *gorm.DB, logger *slog.Logger) StoreInventoryService {
	return &storeInventoryService{invRepo: invRepo, skuRepo: skuRepo, db: db, logger: logger}
}

func (s *storeInventoryService) Ensure(storeID, skuID uint, quantity int) (*model.StoreInventory, error) {
	return s.EnsureTx(nil, storeID, skuID, quantity)
}

func (s *storeInventoryService) EnsureTx(tx *gorm.DB, storeID, skuID uint, quantity int) (*model.StoreInventory, error) {
	inv, err := s.invRepo.FindByStoreAndSKUTx(tx, storeID, skuID)
	if err == nil {
		return inv, nil
	}
	if !errors.Is(err, util.ErrNotFound) {
		return nil, fmt.Errorf("ensure inventory store[%d] sku[%d]: %w", storeID, skuID, err)
	}
	inv = &model.StoreInventory{StoreID: storeID, SKUID: skuID, Quantity: quantity, SafetyStock: 0}
	if err := s.invRepo.CreateTx(tx, inv); err != nil {
		return nil, fmt.Errorf("ensure inventory store[%d] sku[%d]: %w", storeID, skuID, err)
	}
	s.logger.Info(constants.LogInventoryUpdated, "store_id", storeID, "sku_id", skuID, "quantity", quantity)
	return inv, nil
}

func (s *storeInventoryService) SetSafetyStock(id uint, safetyStock int) (*model.StoreInventory, error) {
	inv, err := s.invRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("set safety stock inventory[id=%d]: %w", id, err)
	}
	if safetyStock < 0 {
		return nil, fmt.Errorf("set safety stock inventory[id=%d]: %w", id, util.ErrValidation)
	}
	inv.SafetyStock = safetyStock
	if err := s.invRepo.Update(inv); err != nil {
		return nil, fmt.Errorf("set safety stock inventory[id=%d]: %w", id, err)
	}
	s.logger.Info(constants.LogInventorySafetySet, "inventory_id", id, "safety_stock", safetyStock)
	return inv, nil
}

func (s *storeInventoryService) GetByStoreAndSKU(storeID, skuID uint) (*model.StoreInventory, error) {
	return s.invRepo.FindByStoreAndSKU(storeID, skuID)
}

func (s *storeInventoryService) List(page, pageSize int, storeID, skuID uint) ([]model.StoreInventory, int64, error) {
	invs, total, err := s.invRepo.List(page, pageSize, storeID, skuID)
	if err != nil {
		return nil, 0, fmt.Errorf("list inventories: %w", err)
	}
	s.logger.Info(constants.LogInventoryStatsFetched, "total", total)
	return invs, total, nil
}

func (s *storeInventoryService) ListAlerts() ([]model.StoreInventory, error) {
	alerts, err := s.invRepo.ListAlerts()
	if err != nil {
		s.logger.Warn(constants.LogStockWarnAlertFailed, "error", err)
		return nil, fmt.Errorf("list inventory alerts: %w", err)
	}
	s.logger.Info(constants.LogInventoryAlertFetched, "count", len(alerts))
	return alerts, nil
}

func (s *storeInventoryService) Stats() (*InventoryStats, error) {
	stats := &InventoryStats{}
	if err := s.db.Model(&model.StoreInventory{}).Distinct("sku_id").Count(&stats.TotalSKUCount).Error; err != nil {
		return nil, fmt.Errorf("stats total sku count: %w", err)
	}
	if err := s.db.Model(&model.StoreInventory{}).Select("COALESCE(SUM(quantity),0)").Scan(&stats.TotalQuantity).Error; err != nil {
		return nil, fmt.Errorf("stats total quantity: %w", err)
	}
	if err := s.db.Model(&model.StoreInventory{}).Where("quantity < safety_stock AND quantity > 0").Count(&stats.LowStockCount).Error; err != nil {
		return nil, fmt.Errorf("stats low stock count: %w", err)
	}
	if err := s.db.Model(&model.StoreInventory{}).Where("quantity <= 0").Count(&stats.OutOfStockCount).Error; err != nil {
		return nil, fmt.Errorf("stats out of stock count: %w", err)
	}
	if err := s.db.Model(&model.Store{}).Count(&stats.StoreCount).Error; err != nil {
		return nil, fmt.Errorf("stats store count: %w", err)
	}
	s.logger.Info(constants.LogInventoryStatsFetched, "total_sku", stats.TotalSKUCount)
	return stats, nil
}

func (s *storeInventoryService) AdjustQuantity(storeID, skuID uint, delta int) error {
	return s.AdjustQuantityTx(nil, storeID, skuID, delta)
}

func (s *storeInventoryService) AdjustQuantityTx(tx *gorm.DB, storeID, skuID uint, delta int) error {
	inv, err := s.invRepo.FindByStoreAndSKUTx(tx, storeID, skuID)
	if err != nil {
		return fmt.Errorf("adjust quantity store[%d] sku[%d]: %w", storeID, skuID, err)
	}
	if inv.Quantity+delta < 0 {
		return fmt.Errorf("adjust quantity store[%d] sku[%d]: %w", storeID, skuID, util.ErrStockNotEnough)
	}
	return s.invRepo.AdjustQuantityTx(tx, storeID, skuID, delta)
}

func (s *storeInventoryService) CheckSufficient(storeID, skuID uint, qty int) error {
	return s.CheckSufficientTx(nil, storeID, skuID, qty)
}

func (s *storeInventoryService) CheckSufficientTx(tx *gorm.DB, storeID, skuID uint, qty int) error {
	inv, err := s.invRepo.FindByStoreAndSKUTx(tx, storeID, skuID)
	if err != nil {
		return fmt.Errorf("check sufficient store[%d] sku[%d]: %w", storeID, skuID, err)
	}
	if inv.Quantity < qty {
		return fmt.Errorf("check sufficient store[%d] sku[%d] need[%d] have[%d]: %w", storeID, skuID, qty, inv.Quantity, util.ErrStockNotEnough)
	}
	return nil
}
