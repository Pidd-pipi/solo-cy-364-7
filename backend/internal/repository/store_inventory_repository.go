package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/util"
)

// StoreInventoryRepository 门店库存仓储。
type StoreInventoryRepository interface {
	Create(inv *model.StoreInventory) error
	CreateTx(tx *gorm.DB, inv *model.StoreInventory) error
	FindByStoreAndSKU(storeID, skuID uint) (*model.StoreInventory, error)
	FindByStoreAndSKUTx(tx *gorm.DB, storeID, skuID uint) (*model.StoreInventory, error)
	FindByID(id uint) (*model.StoreInventory, error)
	List(page, pageSize int, storeID, skuID uint) ([]model.StoreInventory, int64, error)
	ListAlerts() ([]model.StoreInventory, error)
	Update(inv *model.StoreInventory) error
	AdjustQuantity(storeID, skuID uint, delta int) error
	AdjustQuantityTx(tx *gorm.DB, storeID, skuID uint, delta int) error
}

type storeInventoryRepository struct {
	db *gorm.DB
}

// NewStoreInventoryRepository 构造门店库存仓储。
func NewStoreInventoryRepository(db *gorm.DB) StoreInventoryRepository {
	return &storeInventoryRepository{db: db}
}

func (r *storeInventoryRepository) Create(inv *model.StoreInventory) error {
	return r.CreateTx(nil, inv)
}

func (r *storeInventoryRepository) CreateTx(tx *gorm.DB, inv *model.StoreInventory) error {
	if err := dbOrTx(r.db, tx).Create(inv).Error; err != nil {
		return fmt.Errorf("create inventory: %w", err)
	}
	return nil
}

func (r *storeInventoryRepository) FindByStoreAndSKU(storeID, skuID uint) (*model.StoreInventory, error) {
	return r.FindByStoreAndSKUTx(nil, storeID, skuID)
}

func (r *storeInventoryRepository) FindByStoreAndSKUTx(tx *gorm.DB, storeID, skuID uint) (*model.StoreInventory, error) {
	var inv model.StoreInventory
	err := dbOrTx(r.db, tx).Preload("Store").Preload("SKU").Where("store_id = ? AND sku_id = ?", storeID, skuID).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find inventory: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find inventory: %w", err)
	}
	return &inv, nil
}

func (r *storeInventoryRepository) FindByID(id uint) (*model.StoreInventory, error) {
	var inv model.StoreInventory
	err := r.db.Preload("Store").Preload("SKU").First(&inv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find inventory by id: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find inventory by id: %w", err)
	}
	return &inv, nil
}

func (r *storeInventoryRepository) List(page, pageSize int, storeID, skuID uint) ([]model.StoreInventory, int64, error) {
	var invs []model.StoreInventory
	var total int64
	q := r.db.Model(&model.StoreInventory{})
	if storeID > 0 {
		q = q.Where("store_id = ?", storeID)
	}
	if skuID > 0 {
		q = q.Where("sku_id = ?", skuID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count inventories: %w", err)
	}
	if err := q.Preload("Store").Preload("SKU").Offset((page - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&invs).Error; err != nil {
		return nil, 0, fmt.Errorf("list inventories: %w", err)
	}
	return invs, total, nil
}

func (r *storeInventoryRepository) ListAlerts() ([]model.StoreInventory, error) {
	var invs []model.StoreInventory
	err := r.db.Preload("Store").Preload("SKU").
		Where("quantity < safety_stock").
		Order("safety_stock - quantity desc").
		Find(&invs).Error
	if err != nil {
		return nil, fmt.Errorf("list inventory alerts: %w", err)
	}
	return invs, nil
}

func (r *storeInventoryRepository) Update(inv *model.StoreInventory) error {
	if err := r.db.Save(inv).Error; err != nil {
		return fmt.Errorf("update inventory: %w", err)
	}
	return nil
}

func (r *storeInventoryRepository) AdjustQuantity(storeID, skuID uint, delta int) error {
	return r.AdjustQuantityTx(nil, storeID, skuID, delta)
}

func (r *storeInventoryRepository) AdjustQuantityTx(tx *gorm.DB, storeID, skuID uint, delta int) error {
	q := dbOrTx(r.db, tx).Model(&model.StoreInventory{}).
		Where("store_id = ? AND sku_id = ?", storeID, skuID)
	if delta < 0 {
		q = q.Where("quantity + ? >= 0", delta)
	}
	res := q.UpdateColumn("quantity", gorm.Expr("quantity + ?", delta))
	if res.Error != nil {
		return fmt.Errorf("adjust inventory quantity: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("adjust inventory quantity: %w", util.ErrNotFound)
	}
	return nil
}
