package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/util"
)

// TransferOrderRepository 调拨单仓储。
type TransferOrderRepository interface {
	Create(order *model.TransferOrder) error
	FindByID(id uint) (*model.TransferOrder, error)
	List(page, pageSize int, storeID uint, status constants.TransferStatus) ([]model.TransferOrder, int64, error)
	Update(order *model.TransferOrder) error
	UpdateTx(tx *gorm.DB, order *model.TransferOrder) error
	TransitionStatusTx(tx *gorm.DB, id uint, from, to constants.TransferStatus) error
}

type transferOrderRepository struct {
	db *gorm.DB
}

// NewTransferOrderRepository 构造调拨单仓储。
func NewTransferOrderRepository(db *gorm.DB) TransferOrderRepository {
	return &transferOrderRepository{db: db}
}

func (r *transferOrderRepository) Create(order *model.TransferOrder) error {
	if err := r.db.Create(order).Error; err != nil {
		return fmt.Errorf("create transfer order: %w", err)
	}
	return nil
}

func (r *transferOrderRepository) FindByID(id uint) (*model.TransferOrder, error) {
	var o model.TransferOrder
	err := r.db.Preload("FromStore").Preload("ToStore").Preload("SKU").First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find transfer order by id: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find transfer order by id: %w", err)
	}
	return &o, nil
}

func (r *transferOrderRepository) List(page, pageSize int, storeID uint, status constants.TransferStatus) ([]model.TransferOrder, int64, error) {
	var orders []model.TransferOrder
	var total int64
	q := r.db.Model(&model.TransferOrder{})
	if storeID > 0 {
		q = q.Where("from_store_id = ? OR to_store_id = ?", storeID, storeID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count transfer orders: %w", err)
	}
	if err := q.Preload("FromStore").Preload("ToStore").Preload("SKU").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("list transfer orders: %w", err)
	}
	return orders, total, nil
}

func (r *transferOrderRepository) Update(order *model.TransferOrder) error {
	return r.UpdateTx(nil, order)
}

func (r *transferOrderRepository) UpdateTx(tx *gorm.DB, order *model.TransferOrder) error {
	if err := dbOrTx(r.db, tx).Save(order).Error; err != nil {
		return fmt.Errorf("update transfer order: %w", err)
	}
	return nil
}

func (r *transferOrderRepository) TransitionStatusTx(tx *gorm.DB, id uint, from, to constants.TransferStatus) error {
	res := dbOrTx(r.db, tx).Model(&model.TransferOrder{}).
		Where("id = ? AND status = ?", id, from).
		Update("status", to)
	if res.Error != nil {
		return fmt.Errorf("transition transfer order status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("transition transfer order status: %w", util.ErrConflict)
	}
	return nil
}
