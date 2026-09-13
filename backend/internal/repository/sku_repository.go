package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/util"
)

// SKURepository 商品主数据仓储。
type SKURepository interface {
	Create(sku *model.SKU) error
	FindByID(id uint) (*model.SKU, error)
	FindByCode(code string) (*model.SKU, error)
	List(page, pageSize int, category, keyword string) ([]model.SKU, int64, error)
	BatchCreate(skus []*model.SKU) (int, error)
	Update(sku *model.SKU) error
	Delete(id uint) error
}

type skuRepository struct {
	db *gorm.DB
}

// NewSKURepository 构造 SKU 仓储。
func NewSKURepository(db *gorm.DB) SKURepository {
	return &skuRepository{db: db}
}

func (r *skuRepository) Create(sku *model.SKU) error {
	if err := r.db.Create(sku).Error; err != nil {
		if isDuplicate(err) {
			return fmt.Errorf("create sku: %w", ErrDuplicate)
		}
		return fmt.Errorf("create sku: %w", err)
	}
	return nil
}

func (r *skuRepository) FindByID(id uint) (*model.SKU, error) {
	var s model.SKU
	err := r.db.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find sku by id: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find sku by id: %w", err)
	}
	return &s, nil
}

func (r *skuRepository) FindByCode(code string) (*model.SKU, error) {
	var s model.SKU
	err := r.db.Where("code = ?", code).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find sku by code: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find sku by code: %w", err)
	}
	return &s, nil
}

func (r *skuRepository) List(page, pageSize int, category, keyword string) ([]model.SKU, int64, error) {
	var skus []model.SKU
	var total int64
	q := r.db.Model(&model.SKU{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("code LIKE ? OR name LIKE ? OR barcode LIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count skus: %w", err)
	}
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&skus).Error; err != nil {
		return nil, 0, fmt.Errorf("list skus: %w", err)
	}
	return skus, total, nil
}

func (r *skuRepository) BatchCreate(skus []*model.SKU) (int, error) {
	if len(skus) == 0 {
		return 0, nil
	}
	if err := r.db.Create(&skus).Error; err != nil {
		if isDuplicate(err) {
			return 0, fmt.Errorf("batch create sku: %w", ErrDuplicate)
		}
		return 0, fmt.Errorf("batch create sku: %w", err)
	}
	return len(skus), nil
}

func (r *skuRepository) Update(sku *model.SKU) error {
	if err := r.db.Save(sku).Error; err != nil {
		if isDuplicate(err) {
			return fmt.Errorf("update sku: %w", ErrDuplicate)
		}
		return fmt.Errorf("update sku: %w", err)
	}
	return nil
}

func (r *skuRepository) Delete(id uint) error {
	res := r.db.Delete(&model.SKU{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete sku: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("delete sku: %w", util.ErrNotFound)
	}
	return nil
}
