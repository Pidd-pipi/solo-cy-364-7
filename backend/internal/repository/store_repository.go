package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/util"
)

// StoreRepository 门店仓储。
type StoreRepository interface {
	Create(store *model.Store) error
	FindByID(id uint) (*model.Store, error)
	FindByCode(code string) (*model.Store, error)
	List(page, pageSize int) ([]model.Store, int64, error)
	ListAll() ([]model.Store, error)
	Update(store *model.Store) error
	Delete(id uint) error
}

type storeRepository struct {
	db *gorm.DB
}

// NewStoreRepository 构造门店仓储。
func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{db: db}
}

func (r *storeRepository) Create(store *model.Store) error {
	if err := r.db.Create(store).Error; err != nil {
		if isDuplicate(err) {
			return fmt.Errorf("create store: %w", ErrDuplicate)
		}
		return fmt.Errorf("create store: %w", err)
	}
	return nil
}

func (r *storeRepository) FindByID(id uint) (*model.Store, error) {
	var s model.Store
	err := r.db.Preload("Manager").First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find store by id: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find store by id: %w", err)
	}
	return &s, nil
}

func (r *storeRepository) FindByCode(code string) (*model.Store, error) {
	var s model.Store
	err := r.db.Where("code = ?", code).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find store by code: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find store by code: %w", err)
	}
	return &s, nil
}

func (r *storeRepository) List(page, pageSize int) ([]model.Store, int64, error) {
	var stores []model.Store
	var total int64
	q := r.db.Model(&model.Store{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count stores: %w", err)
	}
	if err := q.Preload("Manager").Offset((page - 1) * pageSize).Limit(pageSize).Order("id asc").Find(&stores).Error; err != nil {
		return nil, 0, fmt.Errorf("list stores: %w", err)
	}
	return stores, total, nil
}

func (r *storeRepository) ListAll() ([]model.Store, error) {
	var stores []model.Store
	if err := r.db.Order("id asc").Find(&stores).Error; err != nil {
		return nil, fmt.Errorf("list all stores: %w", err)
	}
	return stores, nil
}

func (r *storeRepository) Update(store *model.Store) error {
	if err := r.db.Save(store).Error; err != nil {
		return fmt.Errorf("update store: %w", err)
	}
	return nil
}

func (r *storeRepository) Delete(id uint) error {
	res := r.db.Delete(&model.Store{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete store: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("delete store: %w", util.ErrNotFound)
	}
	return nil
}
