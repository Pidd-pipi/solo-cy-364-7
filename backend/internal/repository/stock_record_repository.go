package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
)

// StockRecordRepository 出入库记录仓储。
type StockRecordRepository interface {
	Create(record *model.StockRecord) error
	CreateTx(tx *gorm.DB, record *model.StockRecord) error
	List(page, pageSize int, storeID, skuID uint, recordType constants.StockRecordType) ([]model.StockRecord, int64, error)
	CreateStocktake(st *model.Stocktake) error
	CreateStocktakeTx(tx *gorm.DB, st *model.Stocktake) error
	ListStocktakes(page, pageSize int, storeID uint) ([]model.Stocktake, int64, error)
	MonthlySalesBySKU(storeID, skuID uint) (int, error)
}

type stockRecordRepository struct {
	db *gorm.DB
}

// NewStockRecordRepository 构造出入库记录仓储。
func NewStockRecordRepository(db *gorm.DB) StockRecordRepository {
	return &stockRecordRepository{db: db}
}

func (r *stockRecordRepository) Create(record *model.StockRecord) error {
	return r.CreateTx(nil, record)
}

func (r *stockRecordRepository) CreateTx(tx *gorm.DB, record *model.StockRecord) error {
	if err := dbOrTx(r.db, tx).Create(record).Error; err != nil {
		return fmt.Errorf("create stock record: %w", err)
	}
	return nil
}

func (r *stockRecordRepository) List(page, pageSize int, storeID, skuID uint, recordType constants.StockRecordType) ([]model.StockRecord, int64, error) {
	var records []model.StockRecord
	var total int64
	q := r.db.Model(&model.StockRecord{})
	if storeID > 0 {
		q = q.Where("store_id = ?", storeID)
	}
	if skuID > 0 {
		q = q.Where("sku_id = ?", skuID)
	}
	if recordType != "" {
		q = q.Where("record_type = ?", recordType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count stock records: %w", err)
	}
	if err := q.Preload("Store").Preload("SKU").Offset((page - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list stock records: %w", err)
	}
	return records, total, nil
}

func (r *stockRecordRepository) CreateStocktake(st *model.Stocktake) error {
	return r.CreateStocktakeTx(nil, st)
}

func (r *stockRecordRepository) CreateStocktakeTx(tx *gorm.DB, st *model.Stocktake) error {
	if err := dbOrTx(r.db, tx).Create(st).Error; err != nil {
		return fmt.Errorf("create stocktake: %w", err)
	}
	return nil
}

func (r *stockRecordRepository) ListStocktakes(page, pageSize int, storeID uint) ([]model.Stocktake, int64, error) {
	var list []model.Stocktake
	var total int64
	q := r.db.Model(&model.Stocktake{})
	if storeID > 0 {
		q = q.Where("store_id = ?", storeID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count stocktakes: %w", err)
	}
	if err := q.Preload("Store").Preload("SKU").Offset((page - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("list stocktakes: %w", err)
	}
	return list, total, nil
}

// MonthlySalesBySKU 统计某门店某 SKU 近 30 天销售出库数量。
func (r *stockRecordRepository) MonthlySalesBySKU(storeID, skuID uint) (int, error) {
	var total int64
	err := r.db.Model(&model.StockRecord{}).
		Where("store_id = ? AND sku_id = ? AND record_type = ? AND created_at >= NOW() - INTERVAL '30 days'", storeID, skuID, constants.RecordSale).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("monthly sales by sku: %w", err)
	}
	return int(total), nil
}
