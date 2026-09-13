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

// StoreService 门店业务逻辑。
type StoreService interface {
	Create(code, name, address string, managerUserID *uint) (*model.Store, error)
	Update(id uint, code, name, address string, managerUserID *uint) (*model.Store, error)
	Delete(id uint) error
	GetByID(id uint) (*model.Store, error)
	List(page, pageSize int) ([]model.Store, int64, error)
	ListAll() ([]model.Store, error)
}

type storeService struct {
	storeRepo repository.StoreRepository
	logger    *slog.Logger
}

// NewStoreService 构造门店服务。
func NewStoreService(storeRepo repository.StoreRepository, logger *slog.Logger) StoreService {
	return &storeService{storeRepo: storeRepo, logger: logger}
}

func (s *storeService) Create(code, name, address string, managerUserID *uint) (*model.Store, error) {
	if code == "" || name == "" {
		return nil, fmt.Errorf("create store: %w", util.ErrValidation)
	}
	if _, err := s.storeRepo.FindByCode(code); err == nil {
		return nil, fmt.Errorf("create store[code=%s]: %w", code, util.ErrConflict)
	}
	store := &model.Store{Code: code, Name: name, Address: address, ManagerUserID: managerUserID}
	if err := s.storeRepo.Create(store); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, fmt.Errorf("create store[code=%s]: %w", code, util.ErrConflict)
		}
		return nil, fmt.Errorf("create store[code=%s]: %w", code, err)
	}
	s.logger.Info(constants.LogStoreCreated, "store_id", store.ID, "code", code)
	return store, nil
}

func (s *storeService) Update(id uint, code, name, address string, managerUserID *uint) (*model.Store, error) {
	store, err := s.storeRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("update store[id=%d]: %w", id, err)
	}
	if code != "" {
		store.Code = code
	}
	if name != "" {
		store.Name = name
	}
	if address != "" {
		store.Address = address
	}
	if managerUserID != nil {
		store.ManagerUserID = managerUserID
	}
	if err := s.storeRepo.Update(store); err != nil {
		return nil, fmt.Errorf("update store[id=%d]: %w", id, err)
	}
	s.logger.Info(constants.LogStoreUpdated, "store_id", id)
	return store, nil
}

func (s *storeService) Delete(id uint) error {
	if err := s.storeRepo.Delete(id); err != nil {
		return fmt.Errorf("delete store[id=%d]: %w", id, err)
	}
	s.logger.Info(constants.LogStoreDeleted, "store_id", id)
	return nil
}

func (s *storeService) GetByID(id uint) (*model.Store, error) {
	store, err := s.storeRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get store[id=%d]: %w", id, err)
	}
	return store, nil
}

func (s *storeService) List(page, pageSize int) ([]model.Store, int64, error) {
	stores, total, err := s.storeRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list stores: %w", err)
	}
	s.logger.Info(constants.LogStoreListQueried, "total", total)
	return stores, total, nil
}

func (s *storeService) ListAll() ([]model.Store, error) {
	stores, err := s.storeRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("list all stores: %w", err)
	}
	return stores, nil
}
