package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/model"
)

// TransferReceiptRepository 调拨收货明细仓储。
type TransferReceiptRepository interface {
	CreateTx(tx *gorm.DB, receipt *model.TransferReceipt) error
	ListByOrderID(orderID uint) ([]model.TransferReceipt, error)
	SumByOrderIDTx(tx *gorm.DB, orderID uint) (int, error)
}

type transferReceiptRepository struct {
	db *gorm.DB
}

// NewTransferReceiptRepository 构造调拨收货明细仓储。
func NewTransferReceiptRepository(db *gorm.DB) TransferReceiptRepository {
	return &transferReceiptRepository{db: db}
}

func (r *transferReceiptRepository) CreateTx(tx *gorm.DB, receipt *model.TransferReceipt) error {
	if err := dbOrTx(r.db, tx).Create(receipt).Error; err != nil {
		return fmt.Errorf("create transfer receipt: %w", err)
	}
	return nil
}

func (r *transferReceiptRepository) ListByOrderID(orderID uint) ([]model.TransferReceipt, error) {
	var receipts []model.TransferReceipt
	if err := r.db.Where("transfer_order_id = ?", orderID).Order("id asc").Find(&receipts).Error; err != nil {
		return nil, fmt.Errorf("list transfer receipts by order[%d]: %w", orderID, err)
	}
	return receipts, nil
}

func (r *transferReceiptRepository) SumByOrderIDTx(tx *gorm.DB, orderID uint) (int, error) {
	var sum int64
	if err := dbOrTx(r.db, tx).Model(&model.TransferReceipt{}).
		Where("transfer_order_id = ?", orderID).
		Select("COALESCE(SUM(quantity),0)").Scan(&sum).Error; err != nil {
		return 0, fmt.Errorf("sum transfer receipts by order[%d]: %w", orderID, err)
	}
	return int(sum), nil
}
