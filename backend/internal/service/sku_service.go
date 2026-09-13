package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// SKUService 商品主数据业务逻辑。
type SKUService interface {
	Create(code, name, spec, barcode, category, unit string) (*model.SKU, error)
	BatchImport(items []SKUImportItem) (int, error)
	Update(id uint, code, name, spec, barcode, category, unit string) (*model.SKU, error)
	Delete(id uint) error
	GetByID(id uint) (*model.SKU, error)
	List(page, pageSize int, category, keyword string) ([]model.SKU, int64, error)
}

// SKUImportItem 批量导入条目。
type SKUImportItem struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Spec     string `json:"spec"`
	Barcode  string `json:"barcode"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
}

type skuService struct {
	skuRepo repository.SKURepository
	logger  *slog.Logger
}

// NewSKUService 构造 SKU 服务。
func NewSKUService(skuRepo repository.SKURepository, logger *slog.Logger) SKUService {
	return &skuService{skuRepo: skuRepo, logger: logger}
}

func (s *skuService) Create(code, name, spec, barcode, category, unit string) (*model.SKU, error) {
	if code == "" || name == "" {
		return nil, fmt.Errorf("create sku: %w", util.ErrValidation)
	}
	sku := &model.SKU{Code: code, Name: name, Spec: spec, Barcode: barcode, Category: category, Unit: unit, Status: "active"}
	if err := s.skuRepo.Create(sku); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, fmt.Errorf("create sku[code=%s]: %w", code, util.ErrConflict)
		}
		return nil, fmt.Errorf("create sku[code=%s]: %w", code, err)
	}
	s.logger.Info(constants.LogSkuCreated, "sku_id", sku.ID, "code", code)
	return sku, nil
}

func (s *skuService) BatchImport(items []SKUImportItem) (int, error) {
	skus := make([]*model.SKU, 0, len(items))
	for _, it := range items {
		if it.Code == "" || it.Name == "" {
			return 0, fmt.Errorf("batch import sku: %w", util.ErrValidation)
		}
		skus = append(skus, &model.SKU{
			Code: it.Code, Name: it.Name, Spec: it.Spec, Barcode: it.Barcode,
			Category: it.Category, Unit: it.Unit, Status: "active",
		})
	}
	n, err := s.skuRepo.BatchCreate(skus)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return 0, fmt.Errorf("batch import sku: %w", util.ErrConflict)
		}
		s.logger.Error(constants.LogSkuBatchImportFailed, "error", err)
		return 0, fmt.Errorf("batch import sku: %w", err)
	}
	s.logger.Info(constants.LogSkuBatchImported, "count", n)
	return n, nil
}

func (s *skuService) Update(id uint, code, name, spec, barcode, category, unit string) (*model.SKU, error) {
	sku, err := s.skuRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("update sku[id=%d]: %w", id, err)
	}
	if code != "" {
		sku.Code = code
	}
	if name != "" {
		sku.Name = name
	}
	if spec != "" {
		sku.Spec = spec
	}
	if barcode != "" {
		sku.Barcode = barcode
	}
	if category != "" {
		sku.Category = category
	}
	if unit != "" {
		sku.Unit = unit
	}
	if err := s.skuRepo.Update(sku); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, fmt.Errorf("update sku[id=%d]: %w", id, util.ErrConflict)
		}
		return nil, fmt.Errorf("update sku[id=%d]: %w", id, err)
	}
	s.logger.Info(constants.LogSkuUpdated, "sku_id", id)
	return sku, nil
}

func (s *skuService) Delete(id uint) error {
	if err := s.skuRepo.Delete(id); err != nil {
		return fmt.Errorf("delete sku[id=%d]: %w", id, err)
	}
	s.logger.Info(constants.LogSkuDeleted, "sku_id", id)
	return nil
}

func (s *skuService) GetByID(id uint) (*model.SKU, error) {
	sku, err := s.skuRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get sku[id=%d]: %w", id, err)
	}
	return sku, nil
}

func (s *skuService) List(page, pageSize int, category, keyword string) ([]model.SKU, int64, error) {
	skus, total, err := s.skuRepo.List(page, pageSize, category, keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("list skus: %w", err)
	}
	s.logger.Info(constants.LogSkuListQueried, "total", total)
	return skus, total, nil
}
